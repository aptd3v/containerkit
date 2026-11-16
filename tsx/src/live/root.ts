import { basename, join } from "node:path";
import { composeDown, composeUp } from "../generated-actions.js";
import { isNode } from "../node.js";
import { projectRoot, render } from "../render.js";
import type { ConnectOptions, DownOptions } from "../generated-options.js";
import type { DockerNode } from "../docker-node.js";
import { memoKey, readMemo, writeMemo } from "./cache.js";
import { ContainerKitError, isCKUnavailableError } from "./error.js";
import { startProgress } from "./progress.js";
import {
	clientFromRoot,
	isMissingBinaryError,
	openSession,
	sessionFromTransport,
	type ErrorLiveEvent,
	type LiveEvent,
	type LiveEventMap,
	type Session,
	type SessionTransport,
} from "./session.js";

const rootBrand = Symbol.for("containerkit.live.root");

/** Settings for {@link root}. The session stays open for the life of the process. */
export type RootOptions = {
	/** Read the Docker host, TLS, and context from the environment. */
	fromEnv?: boolean;
	/** Extra client settings, combined with `fromEnv`. */
	connect?: ConnectOptions;
	/**
	 * Called when the callback or an effect throws.
	 * The session stays open so a later update can retry, unless Docker cannot be reached.
	 */
	onError?: (err: unknown) => void;
	/**
	 * `"down"` takes the last applied project down on SIGINT and SIGTERM, then closes the session.
	 * When this file is the program, the process exits. Omit it and the containers stay up.
	 */
	onExit?: "down";
	/** Runs after a successful apply. With {@link RootOptions.wait}, runs once the containers are up. */
	onUp?: () => void;
	/**
	 * Wait until containers are running, and healthy when a healthcheck exists, before `onUp`.
	 * The apply stays detached, so the session does not follow logs.
	 */
	wait?: boolean;
	/** Path to the containerkit binary. Defaults to `CONTAINERKIT_BIN`, then `containerkit`. */
	bin?: string;
	/** Extra arguments passed to the binary. */
	args?: string[];
	/**
	 * Directory for {@link useMemo} results. The default is `.containerkit/memo` in the working directory.
	 * `false`, or `CONTAINERKIT_MEMO=0`, keeps the result in this process only.
	 */
	cacheDir?: string | false;
};

/**
 * Render callback. It has no parameters, so it runs again when state changes.
 * Return a `Compose`, `Service`, `Job`, or `Container` to apply it. Return `null` to apply nothing.
 * A fragment must hold exactly one of those. Several roots is an error.
 * The same YAML as last time applies nothing, and those containers keep running.
 */
export type RenderRoot = () => DockerNode | null | undefined;

/**
 * Once callback. An `async` function, or a function that takes the session, runs once and does not render again.
 * Await commands, then return the tree to apply. A throw exits the process when this file is the program.
 */
export type OnceRoot = (session: Session) => DockerNode | null | undefined | Promise<DockerNode | null | undefined>;

/** Handle returned by {@link root}. Pass it to {@link runRoot}. `start()` requires the default export to be one. */
export type LiveRoot = {
	/** Brand that marks a value returned by {@link root}. */
	readonly [rootBrand]: true;
	/** Settings passed to {@link root}. */
	readonly options: RootOptions;
	/** Render callback or once callback. */
	readonly render: RenderRoot | OnceRoot;
	/** True after the session has closed. */
	done: boolean;
	/** True while {@link runRoot} is in progress. */
	running: boolean;
	/** True when the once callback failed. {@link runIfMain} exits the process. */
	failed: boolean;
};

/** Open session from {@link runRoot}. */
export type Running = {
	/** Session bound to this root. */
	session: Session;
	/**
	 * Close the session.
	 * This does not take the project down. {@link RootOptions.onExit} does that on SIGINT and SIGTERM.
	 */
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
	onExit?: "down";
	onUp?: () => void;
	wait?: boolean;
	cacheDir?: string;
	lastYAML: string;
	project?: DockerNode;
	/** Down options from {@link useShutdown}. Read when a signal arrives. */
	downOptions?: DownOptions;
	/** Tree to take down when {@link useShutdown} was given one and nothing has been applied yet. */
	shutdownProject?: DockerNode;
	inflight: Promise<void>;
	closed: boolean;
	pending: Promise<void>[];
	failSoon?: unknown;
	scheduled: boolean;
	flushing: boolean;
	kick: () => void;
	/** Set when the containerkit binary cannot be found. The process should exit. */
	missingBinary?: boolean;
};

let active: Runtime | undefined;
let openRoot: LiveRoot | undefined;
const runtimes = new WeakMap<Session, Runtime>();

/**
 * Remember the down options for this session.
 * A passed project is used only when the root has not applied one yet.
 */
export function rememberShutdown(session: Session, project: DockerNode | undefined, options: DownOptions | undefined): void {
	const rt = runtimes.get(session);
	if (!rt) {
		throw new Error("containerkit/live: useShutdown must be called from a root callback");
	}
	rt.downOptions = options;
	rt.shutdownProject = project;
}

/** Clear a shutdown recorded by {@link rememberShutdown}. A finished session is left alone. */
export function forgetShutdown(session: Session): void {
	const rt = runtimes.get(session);
	if (!rt) {
		return;
	}
	rt.downOptions = undefined;
	rt.shutdownProject = undefined;
}

/** Project and down options a signal should use. The applied project wins over one passed to the hook. */
export function shutdownTarget(session: Session): { project?: DockerNode; options?: DownOptions; onExit?: "down" } {
	const rt = runtimes.get(session);
	if (!rt) {
		throw new Error("containerkit/live: useShutdown must be called from a root callback");
	}
	return { project: rt.project ?? rt.shutdownProject, options: rt.downOptions, onExit: rt.onExit };
}

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

/**
 * Store a value across renders of the render callback.
 *
 * @remarks
 * Setting the value runs the callback again. The same value, compared with `Object.is`, does not.
 * Call hooks in the same order on every render. An async once callback runs once and does not render again.
 *
 * @template T - Stored value
 * @param init - Initial value, or a function that returns it. The function runs on the first render only.
 * @returns The current value and a setter. The setter accepts the next value or a function of the previous one.
 * @throws Error when called outside a render callback
 *
 * @example
 * ```tsx
 * function App() {
 *   const [ready, setReady] = useState(false);
 *   return <Service name="web" image="nginx:alpine" />;
 * }
 * ```
 *
 * @see {@link useMemo}
 */
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

/**
 * {@link useState}. Kept for call sites that already use this name.
 *
 * @template T - Stored value
 * @param init - Initial value, or a function that returns it
 * @returns The current value and a setter
 * @see {@link useState}
 */
export function useContainerState<T>(init: T | (() => T)): [T, (value: T | ((prev: T) => T)) => void] {
	return useState(init);
}

/**
 * Session for the callback that is running.
 *
 * @returns The open session
 * @throws Error when called outside a root callback, including after the first `await` of an async root
 *
 * @see {@link root}
 */
export function useSession(): Session {
	return current().session;
}

/**
 * Run `factory` when `deps` change, and reuse the result otherwise.
 *
 * @remarks
 * A promise returns `undefined` on that render. The callback runs again when it settles.
 * The result is written under `.containerkit/memo` unless {@link RootOptions.cacheDir} is `false`.
 * Do not memoize a Compose, Service, Job, or Container node. The disk cache stores JSON, and the next process would read that JSON back instead of a node.
 * An async once callback does not render again, so a promised result stays `undefined`.
 *
 * @template T - Value `factory` returns
 * @param factory - Computes the value. It may return a promise.
 * @param deps - Runs `factory` again when one of these changes, compared with `Object.is`
 * @returns The stored value, or `undefined` while a promise is still running
 * @throws Error when called outside a render callback. A throw from `factory` fails the root.
 *
 * @example
 * ```tsx
 * function App() {
 *   const image = useMemo(() => process.env.IMAGE ?? "nginx:alpine", []);
 *   return <Service name="web" image={image} />;
 * }
 * ```
 *
 * @see {@link useEffect}
 */
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

/**
 * {@link useMemo}. Kept for call sites that already use this name.
 *
 * @template T - Value `factory` returns
 * @param factory - Computes the value. It may return a promise.
 * @param deps - Runs `factory` again when one of these changes
 * @returns The stored value, or `undefined` while a promise is still running
 * @see {@link useMemo}
 */
export function useResource<T>(factory: () => T | Promise<T>, deps: readonly unknown[]): T | undefined {
	return useMemo(factory, deps);
}

/**
 * Run `effect` after a render that has nothing left to wait for.
 *
 * @remarks
 * A function returned from `effect` runs before the next run, and when the session closes.
 * An async once callback does not render again, so an effect recorded there never runs.
 * Call {@link useShutdown} directly in that callback instead of from an effect.
 *
 * @param effect - Work to run after the render. Return a function to undo it.
 * @param deps - Runs again when one of these changes, compared with `Object.is`
 * @throws Error when called outside a render callback
 *
 * @example
 * ```tsx
 * function App() {
 *   const session = useSession();
 *   useEffect(() => session.on("log", (event) => console.log(event.line)), [session]);
 *   return <Service name="web" image="nginx:alpine" />;
 * }
 * ```
 *
 * @see {@link useMemo}
 */
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

function isOnce(fn: RenderRoot | OnceRoot): fn is OnceRoot {
	return fn.length > 0 || fn.constructor.name === "AsyncFunction";
}

async function runOnce(rt: Runtime, fn: OnceRoot): Promise<"failed" | "ok"> {
	rt.index = 0;
	active = rt;
	let result: ReturnType<OnceRoot>;
	try {
		result = fn(rt.session);
	} catch (err) {
		active = undefined;
		await report(rt, err);
		return "failed";
	}
	active = undefined;
	let node: DockerNode | null | undefined;
	try {
		node = isPromise(result) ? ((await result) as DockerNode | null | undefined) : result;
	} catch (err) {
		await report(rt, err);
		return "failed";
	}
	if (rt.stopped || node == null || !isNode(node)) {
		return "ok";
	}
	let unit: DockerNode;
	try {
		unit = projectRoot(node) as DockerNode;
	} catch (err) {
		await report(rt, err);
		return "failed";
	}
	let yaml: string;
	try {
		yaml = render(unit);
	} catch (err) {
		await report(rt, err);
		return "failed";
	}
	rt.lastYAML = yaml;
	return applyProject(rt, unit);
}

async function applyProject(rt: Runtime, node: DockerNode): Promise<"failed" | "ok"> {
	const options = rt.wait ? { detach: true as const, wait: true as const } : { detach: true as const };
	try {
		await rt.session.call(composeUp(rt.session.client, node, options));
	} catch (err) {
		await report(rt, err);
		return "failed";
	}
	rt.project = node;
	if (rt.stopped) {
		return "ok";
	}
	try {
		rt.onUp?.();
	} catch (err) {
		await report(rt, err);
		return "failed";
	}
	return "ok";
}

async function report(rt: Runtime, err: unknown): Promise<void> {
	const message = err instanceof Error ? err.message : String(err);
	console.error(message);
	if (isMissingBinaryError(err)) {
		rt.missingBinary = true;
	}
	rt.onError?.(err);
	const event: ErrorLiveEvent = {
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

let errorEmit: ((event: ErrorLiveEvent) => void) | undefined;

function emitError(rt: Runtime, event: ErrorLiveEvent): void {
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
		if (node != null && isNode(node) && !isUnit(node.$$kind)) {
			try {
				node = projectRoot(node) as DockerNode;
			} catch (err) {
				await report(rt, err);
				return;
			}
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
			if ((await applyProject(rt, node)) === "failed") {
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

/**
 * Open a live session around a render callback or a once callback.
 *
 * @remarks
 * When the file you ran is `docker.ts`, `docker.tsx`, or `docker.js`, the session starts on its own.
 * Any other file must pass the handle to {@link runRoot} or {@link runIfMain}.
 * A render callback has no parameters and is not `async`. It runs again when state changes.
 * An `async` function, or a function that takes the session, runs once.
 *
 * @param options - How the session connects, and what happens on exit
 * @param renderRoot - Render callback, or a once callback that receives the session
 * @returns A handle. `start()` requires this as the default export of `docker.ts`.
 * @throws Error when `onExit` is any value other than `"down"`, or when another root is still open
 *
 * @example
 * ```tsx
 * export default root({ fromEnv: true, onExit: "down" }, async function App(session) {
 *   await session.composeUp(compose, { detach: true });
 *   return compose;
 * });
 * ```
 *
 * @see {@link useSession}
 */
export function root(options: RootOptions, renderRoot: RenderRoot | OnceRoot): LiveRoot {
	if (options.onExit !== undefined && options.onExit !== "down") {
		throw new Error('containerkit/live: onExit must be "down"');
	}
	if (openRoot && !openRoot.done) {
		throw new Error("containerkit/live: a second root() is not allowed");
	}
	const handle: LiveRoot = {
		[rootBrand]: true,
		options,
		render: renderRoot,
		done: false,
		running: false,
		failed: false,
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

/** True when `value` is the handle returned by {@link root}. */
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

/**
 * Start the session for a {@link root} handle.
 *
 * @remarks
 * {@link root} calls this itself when the entry file is `docker.ts`, `docker.tsx`, or `docker.js`.
 *
 * @param live - Handle returned by {@link root}
 * @param transport - Replaces the containerkit binary. Tests pass a fake session here.
 * @returns The open session and a `close` function. `close` does not take the project down.
 * @throws Error when `live` was not returned by {@link root}, or when that root is already running
 */
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
	const session = sessionFromTransport(client, wire, { bin: live.options.bin });
	const listeners = new Set<(event: ErrorLiveEvent) => void>();
	const progress = startProgress({ tty: process.stdout.isTTY === true });
	const printProgress = (event: LiveEvent) => {
		progress.push(event);
	};
	const stopWatching = [
		session.on("error", (event) => {
			for (const fn of listeners) {
				fn(event);
			}
		}),
		session.on("log", printProgress),
		session.on("status", printProgress),
	];
	const rt: Runtime = {
		slots: [],
		index: 0,
		session,
		stopped: false,
		onError: live.options.onError,
		onExit: live.options.onExit,
		onUp: live.options.onUp,
		wait: live.options.wait,
		cacheDir: cacheDirectory(live.options),
		lastYAML: "",
		inflight: Promise.resolve(),
		closed: false,
		pending: [],
		scheduled: false,
		flushing: false,
		kick: () => undefined,
	};
	runtimes.set(session, rt);
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
		let finished: () => void = () => undefined;
		rt.inflight = new Promise<void>((resolve) => {
			finished = resolve;
		});
		try {
			for (let i = 0; i < 50 && !rt.stopped; i++) {
				rt.scheduled = false;
				await commit(rt, live.render as RenderRoot);
				if (!rt.scheduled) {
					break;
				}
			}
		} finally {
			rt.flushing = false;
			if (rt.scheduled && !rt.stopped) {
				rt.kick();
			}
			finished();
		}
	};
	const once = isOnce(live.render);
	rt.kick = () => {
		if (once) {
			return;
		}
		rt.scheduled = true;
		if (!rt.flushing) {
			void flush();
		}
	};
	const exitProcess = rt.onExit === "down" && isEntryDocker();
	let shuttingDown: Promise<void> | undefined;
	let signaled = false;
	const onSignal = () => {
		signaled = true;
		void shutdown();
	};
	async function finish(): Promise<void> {
		if (rt.closed) {
			return;
		}
		rt.closed = true;
		rt.stopped = true;
		progress.stop();
		process.off("SIGINT", onSignal);
		process.off("SIGTERM", onSignal);
		for (const slot of rt.slots) {
			slot.cleanup?.();
			slot.cleanup = undefined;
		}
		for (const stop of stopWatching) {
			stop();
		}
		errorEmit = undefined;
		await session.close();
		runtimes.delete(session);
		live.done = true;
		live.running = false;
		if (openRoot === live) {
			openRoot = undefined;
		}
	}
	function shutdown(): Promise<void> {
		if (!shuttingDown) {
			shuttingDown = (async () => {
				rt.stopped = true;
				await rt.inflight;
				const project = rt.project ?? rt.shutdownProject;
				if (project && !rt.closed) {
					try {
						await session.call(composeDown(session.client, project, rt.downOptions));
					} catch (err) {
						rt.onError?.(err);
					}
				}
				await finish();
				if (exitProcess) {
					process.exit(0);
				}
			})();
		}
		return shuttingDown;
	}
	if (rt.onExit === "down") {
		process.on("SIGINT", onSignal);
		process.on("SIGTERM", onSignal);
	}
	const running: Running = {
		session: {
			...session,
			on<K extends keyof LiveEventMap>(kind: K, handler: (event: LiveEventMap[K]) => void) {
				if (kind === "error") {
					const errorHandler = handler as (event: ErrorLiveEvent) => void;
					listeners.add(errorHandler);
					return () => {
						listeners.delete(errorHandler);
					};
				}
				return session.on(kind, handler);
			},
		},
		close: () => finish(),
	};
	if (once) {
		let finished: () => void = () => undefined;
		rt.inflight = new Promise<void>((resolve) => {
			finished = resolve;
		});
		let outcome: "failed" | "ok" = "ok";
		try {
			outcome = await runOnce(rt, live.render);
		} finally {
			finished();
		}
		if (outcome === "failed") {
			live.failed = true;
		}
		if (outcome === "failed" && !signaled && (isEntryDocker() || rt.missingBinary)) {
			await finish();
			process.exit(1);
		}
		return running;
	}
	await flush();
	return running;
}
