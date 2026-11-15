import { basename, join } from "node:path";
import { composeUp } from "../generated-actions.js";
import { isNode } from "../node.js";
import { render } from "../render.js";
import type { ConnectOptions } from "../generated-options.js";
import type { DockerNode } from "../docker-node.js";
import { memoKey, readMemo, writeMemo } from "./cache.js";
import { ContainerKitError, isCKUnavailableError } from "./error.js";
import {
	clientFromRoot,
	openSession,
	sessionFromTransport,
	type LiveEvent,
	type Session,
	type SessionTransport,
} from "./session.js";

const rootBrand = Symbol.for("containerkit.live.root");

export type RootOptions = {
	fromEnv?: boolean;
	connect?: ConnectOptions;
	onError?: (err: unknown) => void;
	bin?: string;
	args?: string[];
	cacheDir?: string | false;
};

export type LiveRoot = {
	readonly [rootBrand]: true;
	readonly options: RootOptions;
	readonly render: () => DockerNode | null | undefined;
	done: boolean;
	running: boolean;
};

export type Running = {
	session: Session;
	close(): Promise<void>;
};

type Slot = {
	state?: unknown;
	stateReady?: boolean;
	memoDeps?: readonly unknown[];
	memoKey?: string;
	memoValue?: unknown;
	memoReady?: boolean;
	inflight?: boolean;
	effect?: () => void | (() => void);
	effectDeps?: readonly unknown[];
	ranDeps?: readonly unknown[];
	cleanup?: () => void;
	effectDirty?: boolean;
};

type Runtime = {
	slots: Slot[];
	index: number;
	session: Session;
	stopped: boolean;
	onError?: (err: unknown) => void;
	cacheDir?: string;
	lastYAML: string;
	pending: Promise<void>[];
	failSoon?: unknown;
	scheduled: boolean;
	flushing: boolean;
	kick: () => void;
};

let active: Runtime | undefined;
let openRoot: LiveRoot | undefined;

function current(): Runtime {
	if (!active) {
		throw new Error("containerkit/live: hooks belong to the docker.ts root");
	}
	return active;
}

function sameDeps(left: readonly unknown[], right: readonly unknown[]): boolean {
	if (left.length !== right.length) {
		return false;
	}
	for (let i = 0; i < left.length; i++) {
		if (!Object.is(left[i], right[i])) {
			return false;
		}
	}
	return true;
}

function isPromise(value: unknown): value is Promise<unknown> {
	return typeof value === "object" && value !== null && typeof (value as Promise<unknown>).then === "function";
}

export function useState<T>(init: T | (() => T)): [T, (value: T | ((prev: T) => T)) => void] {
	const rt = current();
	const slot = slotAt(rt);
	if (!slot.stateReady) {
		slot.state = typeof init === "function" ? (init as () => T)() : init;
		slot.stateReady = true;
	}
	const set = (value: T | ((prev: T) => T)) => {
		const prev = slot.state as T;
		const next = typeof value === "function" ? (value as (prev: T) => T)(prev) : value;
		if (Object.is(next, prev) || rt.stopped) {
			return;
		}
		slot.state = next;
		rt.kick();
	};
	return [slot.state as T, set];
}

export function useContainerState<T>(init: T | (() => T)): [T, (value: T | ((prev: T) => T)) => void] {
	return useState(init);
}

export function useSession(): Session {
	return current().session;
}

export function useMemo<T>(factory: () => T | Promise<T>, deps: readonly unknown[]): T | undefined {
	const rt = current();
	const slot = slotAt(rt);
	const key = memoKey(factory, deps);
	if (cached(slot, key, deps)) {
		return slot.memoReady ? (slot.memoValue as T) : undefined;
	}
	if (key && rt.cacheDir) {
		const stored = readMemo(rt.cacheDir, key);
		if (stored) {
			remember(slot, key, deps, stored.value);
			return stored.value as T;
		}
	}
	slot.memoKey = key;
	slot.memoDeps = deps.slice();
	slot.memoReady = false;
	slot.inflight = true;
	let result: T | Promise<T>;
	try {
		result = factory();
	} catch (err) {
		forget(slot);
		rt.failSoon = err;
		return undefined;
	}
	if (isPromise(result)) {
		const task = result.then(
			(value) => {
				remember(slot, key, deps, value);
				if (key && rt.cacheDir) {
					writeMemo(rt.cacheDir, key, value);
				}
			},
			(err: unknown) => {
				forget(slot);
				rt.failSoon = err;
			},
		);
		rt.pending.push(task.then(() => undefined));
		return undefined;
	}
	remember(slot, key, deps, result);
	if (key && rt.cacheDir) {
		writeMemo(rt.cacheDir, key, result);
	}
	return result;
}

function cached(slot: Slot, key: string | undefined, deps: readonly unknown[]): boolean {
	if (key) {
		return slot.memoKey === key && (slot.memoReady === true || slot.inflight === true);
	}
	return Boolean(slot.memoDeps && sameDeps(slot.memoDeps, deps) && (slot.memoReady === true || slot.inflight === true));
}

function remember(slot: Slot, key: string | undefined, deps: readonly unknown[], value: unknown): void {
	slot.memoKey = key;
	slot.memoDeps = deps.slice();
	slot.memoValue = value;
	slot.memoReady = true;
	slot.inflight = false;
}

function forget(slot: Slot): void {
	slot.inflight = false;
	slot.memoReady = false;
	slot.memoKey = undefined;
	slot.memoDeps = undefined;
	slot.memoValue = undefined;
}

export function useResource<T>(factory: () => T | Promise<T>, deps: readonly unknown[]): T | undefined {
	return useMemo(factory, deps);
}

export function useEffect(effect: () => void | (() => void), deps: readonly unknown[]): void {
	const slot = slotAt(current());
	slot.effect = effect;
	slot.effectDeps = deps.slice();
	slot.effectDirty = !slot.ranDeps || !sameDeps(slot.ranDeps, deps);
}

function slotAt(rt: Runtime): Slot {
	const index = rt.index++;
	let slot = rt.slots[index];
	if (!slot) {
		slot = {};
		rt.slots[index] = slot;
	}
	return slot;
}

function isUnit(kind: string): boolean {
	return kind === "compose" || kind === "service" || kind === "job" || kind === "container";
}

async function report(rt: Runtime, err: unknown): Promise<void> {
	rt.onError?.(err);
	const event: LiveEvent = {
		type: "event",
		kind: "error",
		message: err instanceof Error ? err.message : String(err),
		class: err instanceof ContainerKitError ? err.class : undefined,
	};
	emitError(rt, event);
	if (isCKUnavailableError(err)) {
		rt.stopped = true;
		await rt.session.close();
	}
}

let errorEmit: ((event: LiveEvent) => void) | undefined;

function emitError(rt: Runtime, event: LiveEvent): void {
	void rt;
	errorEmit?.(event);
}

function runEffects(rt: Runtime): void {
	for (const slot of rt.slots) {
		if (!slot.effect || !slot.effectDirty || !slot.effectDeps) {
			continue;
		}
		slot.cleanup?.();
		slot.cleanup = undefined;
		try {
			const cleanup = slot.effect();
			if (typeof cleanup === "function") {
				slot.cleanup = cleanup;
			}
		} catch (err) {
			rt.failSoon = err;
		}
		slot.effectDirty = false;
		slot.ranDeps = slot.effectDeps;
	}
}

async function commit(rt: Runtime, renderRoot: () => DockerNode | null | undefined): Promise<void> {
	for (;;) {
		rt.pending = [];
		rt.failSoon = undefined;
		rt.index = 0;
		active = rt;
		let node: DockerNode | null | undefined;
		try {
			node = renderRoot();
			if (rt.index !== rt.slots.length) {
				throw new Error("containerkit/live: hook order changed between renders");
			}
		} catch (err) {
			active = undefined;
			await report(rt, err);
			return;
		}
		active = undefined;
		if (rt.failSoon && rt.pending.length === 0) {
			const err = rt.failSoon;
			rt.failSoon = undefined;
			await report(rt, err);
			return;
		}
		if (rt.pending.length > 0) {
			await Promise.all(rt.pending);
			if (rt.failSoon) {
				const err = rt.failSoon;
				rt.failSoon = undefined;
				await report(rt, err);
				return;
			}
			continue;
		}
		if (node == null || !isNode(node) || !isUnit(node.$$kind)) {
			runEffects(rt);
			if (rt.failSoon) {
				const err = rt.failSoon;
				rt.failSoon = undefined;
				await report(rt, err);
			}
			return;
		}
		let yaml: string;
		try {
			yaml = render(node);
		} catch (err) {
			await report(rt, err);
			return;
		}
		if (yaml !== rt.lastYAML) {
			rt.lastYAML = yaml;
			try {
				await rt.session.call(composeUp(rt.session.client, node, { detach: true }));
			} catch (err) {
				await report(rt, err);
				return;
			}
		}
		runEffects(rt);
		if (rt.failSoon) {
			const err = rt.failSoon;
			rt.failSoon = undefined;
			await report(rt, err);
		}
		return;
	}
}

export function root(options: RootOptions, renderRoot: () => DockerNode | null | undefined): LiveRoot {
	if (openRoot && !openRoot.done) {
		throw new Error("containerkit/live: a second root() is not allowed");
	}
	const handle: LiveRoot = {
		[rootBrand]: true,
		options,
		render: renderRoot,
		done: false,
		running: false,
	};
	openRoot = handle;
	if (isEntryDocker()) {
		queueMicrotask(() => {
			void runRoot(handle).catch((err: unknown) => {
				console.error(err);
				process.exitCode = 1;
			});
		});
	}
	return handle;
}

export function isLiveRoot(value: unknown): value is LiveRoot {
	return typeof value === "object" && value !== null && rootBrand in value;
}

function cacheDirectory(options: RootOptions): string | undefined {
	if (options.cacheDir === false) {
		return undefined;
	}
	if (typeof options.cacheDir === "string") {
		return options.cacheDir;
	}
	if (process.env.CONTAINERKIT_MEMO === "0") {
		return undefined;
	}
	return join(process.cwd(), ".containerkit", "memo");
}

function isEntryDocker(): boolean {
	const entry = process.argv[1];
	if (!entry) {
		return false;
	}
	const base = basename(entry);
	return base === "docker.ts" || base === "docker.tsx" || base === "docker.js";
}

export async function runRoot(live: LiveRoot, transport?: SessionTransport): Promise<Running> {
	if (!isLiveRoot(live)) {
		throw new Error("containerkit/live: docker.ts must default-export root()");
	}
	if (live.running) {
		throw new Error("containerkit/live: root is already running");
	}
	live.running = true;
	const wire = transport ?? openSession({ bin: live.options.bin, args: live.options.args });
	const client = clientFromRoot(live.options);
	const session = sessionFromTransport(client, wire);
	const listeners = new Set<(event: LiveEvent) => void>();
	const off = session.on("error", (event) => {
		for (const fn of listeners) {
			fn(event);
		}
	});
	const rt: Runtime = {
		slots: [],
		index: 0,
		session,
		stopped: false,
		onError: live.options.onError,
		cacheDir: cacheDirectory(live.options),
		lastYAML: "",
		pending: [],
		scheduled: false,
		flushing: false,
		kick: () => undefined,
	};
	errorEmit = (event) => {
		for (const fn of listeners) {
			fn(event);
		}
	};
	const flush = async () => {
		if (rt.flushing || rt.stopped) {
			return;
		}
		rt.flushing = true;
		try {
			for (let i = 0; i < 50 && !rt.stopped; i++) {
				rt.scheduled = false;
				await commit(rt, live.render);
				if (!rt.scheduled) {
					break;
				}
			}
		} finally {
			rt.flushing = false;
			if (rt.scheduled && !rt.stopped) {
				rt.kick();
			}
		}
	};
	rt.kick = () => {
		rt.scheduled = true;
		if (!rt.flushing) {
			void flush();
		}
	};
	const running: Running = {
		session: {
			...session,
			on(kind, handler) {
				if (kind === "error") {
					listeners.add(handler);
					return () => {
						listeners.delete(handler);
					};
				}
				return session.on(kind, handler);
			},
		},
		async close() {
			rt.stopped = true;
			for (const slot of rt.slots) {
				slot.cleanup?.();
				slot.cleanup = undefined;
			}
			off();
			errorEmit = undefined;
			await session.close();
			live.done = true;
			live.running = false;
			if (openRoot === live) {
				openRoot = undefined;
			}
		},
	};
	await flush();
	return running;
}
