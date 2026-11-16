import { spawn, type ChildProcessWithoutNullStreams } from "node:child_process";
import { createInterface } from "node:readline";
import { connect, type Client, type PlanDocument } from "../plan.js";
import { commandByName, type SessionCommands } from "./commands.js";
import { ContainerKitError, errorFromFrame, type ErrorClass, type ErrorFrame } from "./error.js";

type LiveEventBase = {
	type: "event";
	/** working, done, warn, error, or info. */
	status?: LiveStatus;
	/** Engine channel: pull, push, build, compose, engine, or log. */
	group?: LiveGroup;
	source?: string;
	/** Image or project this event belongs to, when source is a step or layer. */
	parent?: string;
	stream?: string;
	line?: string;
	/** Operation text, such as a build step or "Downloading". */
	text?: string;
	/** Extra text, such as a cache hit, a log line, or an error. */
	detail?: string;
	/**
	 * The row the Go logger prints: kind, glyph, clipped id, and operation.
	 * ANSI color is included when this process's stdout is a terminal.
	 */
	display?: string;
	/** Bytes transferred for a pull or build progress event. */
	current?: number;
	/** Total bytes for a pull or build progress event. */
	total?: number;
};

/** working, done, warn, error, or info. */
export type LiveStatus = "working" | "done" | "warn" | "error" | "info";

/** Engine channel on a session event. */
export type LiveGroup = "pull" | "push" | "build" | "compose" | "engine" | "log";

export type LogLiveEvent = LiveEventBase & {
	kind: "log";
};

export type StatusLiveEvent = LiveEventBase & {
	kind: "status";
};

export type ErrorLiveEvent = LiveEventBase & {
	kind: "error";
	message?: string;
	class?: ErrorClass;
};

/**
 * One event from the session.
 *
 * @remarks
 * Switch on `kind`. A `log` or `status` event has `line`. An `error` event has `message` and `class`.
 * `group` is the engine channel: `pull`, `push`, `build`, `compose`, `engine`, or `log`.
 *
 * @see {@link eventText}
 */
export type LiveEvent = LogLiveEvent | StatusLiveEvent | ErrorLiveEvent;

export type LiveEventMap = {
	log: LogLiveEvent;
	status: StatusLiveEvent;
	error: ErrorLiveEvent;
};

const genericStreams = new Set(["stdout", "stderr", "progress", "step"]);

/**
 * A log or status event.
 *
 * @remarks
 * This is the event `useLiveLog` receives. Error events are a separate kind and carry `message` and `class`.
 */
export type SessionLogEvent = LiveEventMap["log"] | LiveEventMap["status"];

/**
 * Text of a session event.
 *
 * @param event - Log, status, or error event
 * @returns `message` for an error, otherwise `line`, with surrounding space removed. Empty when neither is set.
 *
 * @example
 * ```ts
 * session.on("log", (event) => {
 *   console.log(eventText(event));
 * });
 * ```
 */
export function eventText(event: LiveEvent): string {
	const message = event.kind === "error" ? event.message : undefined;
	return (event.line ?? message ?? "").trim();
}

/** One stdout line for a session log or status event. Empty when there is nothing to show. */
export function formatProgress(event: LiveEvent): string {
	const display = event.display ?? "";
	if (display.trim() !== "") {
		return display;
	}
	let text = eventText(event);
	const who = (event.source ?? "").trim();
	const stream = (event.stream ?? "").trim();
	if (!text && stream && !genericStreams.has(stream)) {
		text = stream;
	}
	let progress = "";
	if (typeof event.total === "number" && event.total > 0) {
		const current = typeof event.current === "number" ? event.current : 0;
		progress = ` ${formatBytes(current)}/${formatBytes(event.total)}`;
	}
	if (who && text) {
		return who === text ? `${who}${progress}` : `${who} ${text}${progress}`;
	}
	const head = who || text;
	return head ? `${head}${progress}` : progress.trim();
}

function formatBytes(n: number): string {
	if (n < 1024) {
		return `${n}B`;
	}
	if (n < 1024 * 1024) {
		return `${(n / 1024).toFixed(1)}kB`;
	}
	if (n < 1024 * 1024 * 1024) {
		return `${(n / (1024 * 1024)).toFixed(1)}MB`;
	}
	return `${(n / (1024 * 1024 * 1024)).toFixed(1)}GB`;
}

/** Settings for {@link Session.terminal}. */
type TerminalOptions = {
	/** Remote argv. The default is `/bin/sh`. */
	command?: readonly string[];
};

export type Session = SessionCommands & {
	readonly client: Client;
	/**
	 * Run a plan. The result is whatever the last action returns.
	 *
	 * @template T - Result of the last action. The default is `unknown`.
	 * @param plan - Actions produced by a generated function such as `composeUp`
	 * @returns The last action's JSON value
	 * @throws {@link ContainerKitError} when the session or the daemon rejects the plan
	 */
	call<T = unknown>(plan: PlanDocument): Promise<T>;
	/**
	 * Subscribe to one kind of session event.
	 *
	 * @template K - `log`, `status`, or `error`. The handler receives only that kind.
	 * @param kind - Which events to receive
	 * @param handler - Called for each matching event
	 * @returns A function that removes the subscription
	 */
	on<K extends keyof LiveEventMap>(kind: K, handler: (event: LiveEventMap[K]) => void): () => void;
	close(): Promise<void>;
	/**
	 * Attach the local terminal to a shell in a running container.
	 * Copies this process's stdin and stdout and blocks until the shell exits.
	 * When stdin is a terminal, the session is raw and the exec is resized with the window.
	 */
	terminal(container: string, options?: TerminalOptions): Promise<void>;
};

export type SessionTransport = {
	call(plan: PlanDocument): Promise<unknown>;
	on(kind: string, handler: (event: LiveEvent) => void): () => void;
	close(): Promise<void>;
};

type ResultFrame = {
	id?: number;
	type: "result";
	ok?: boolean;
	value?: unknown;
};

type WireError = ErrorFrame & { id?: number };

/** Env for the session process. Color follows this process's stdout. NO_COLOR wins. */
export function sessionChildEnv(stdoutIsTTY: boolean, env: NodeJS.ProcessEnv = process.env): NodeJS.ProcessEnv {
	const next: NodeJS.ProcessEnv = { ...env };
	if (env.NO_COLOR) {
		delete next.CONTAINERKIT_COLOR;
		return next;
	}
	if (stdoutIsTTY) {
		next.CONTAINERKIT_COLOR = "1";
	}
	return next;
}

export function openSession(options?: { bin?: string; args?: string[] }): SessionTransport {
	const bin = options?.bin ?? process.env.CONTAINERKIT_BIN ?? "containerkit";
	const args = [...(options?.args ?? []), "session"];
	const child = spawn(bin, args, {
		stdio: ["pipe", "pipe", "pipe"],
		env: sessionChildEnv(process.stdout.isTTY),
		// A new process group, so Ctrl+C reaches this process and not the session.
		// The session has to stay up long enough to take the project down.
		detached: true,
	}) as ChildProcessWithoutNullStreams;
	child.stderr.on("data", (chunk: Buffer) => {
		process.stderr.write(chunk);
	});
	return bindChild(child);
}

function bindChild(child: ChildProcessWithoutNullStreams): SessionTransport {
	const pending = new Map<number, { resolve: (value: unknown) => void; reject: (err: unknown) => void }>();
	const listeners = new Set<(event: LiveEvent) => void>();
	let next = 1;
	let closed = false;
	let spawnError: Error | undefined;

	const failAll = (err: Error) => {
		spawnError = err;
		for (const wait of pending.values()) {
			wait.reject(err);
		}
		pending.clear();
	};

	child.on("error", (err: Error) => {
		const code = (err as Error & { code?: string }).code;
		if (code === "ENOENT") {
			failAll(missingBinary(child.spawnfile));
			return;
		}
		failAll(err);
	});

	const rl = createInterface({ input: child.stdout });
	rl.on("line", (line) => {
		let msg: { type?: string; id?: number };
		try {
			msg = JSON.parse(line) as { type?: string; id?: number };
		} catch {
			return;
		}
		if (msg.type === "event") {
			const event = msg as LiveEvent;
			for (const fn of listeners) {
				fn(event);
			}
			return;
		}
		if (msg.id === undefined) {
			return;
		}
		const wait = pending.get(msg.id);
		if (!wait) {
			return;
		}
		pending.delete(msg.id);
		if (msg.type === "error") {
			wait.reject(errorFromFrame(msg as WireError));
			return;
		}
		const result = msg as ResultFrame;
		wait.resolve(result.value);
	});

	child.on("close", () => {
		if (!closed) {
			failAll(new ContainerKitError({
				type: "error",
				class: "unavailable",
				message: "session closed",
			}));
		}
	});

	return {
		call(plan) {
			if (spawnError) {
				return Promise.reject(spawnError);
			}
			const id = next++;
			return new Promise((resolve, reject) => {
				pending.set(id, { resolve, reject });
				child.stdin.write(JSON.stringify({ id, type: "exec", plan }) + "\n", (err) => {
					if (err) {
						pending.delete(id);
						reject(err);
					}
				});
			});
		},
		on(_kind, handler) {
			listeners.add(handler);
			return () => {
				listeners.delete(handler);
			};
		},
		close() {
			if (closed) {
				return Promise.resolve();
			}
			closed = true;
			return new Promise((resolve) => {
				if (child.exitCode !== null || child.signalCode !== null) {
					resolve();
					return;
				}
				child.once("close", () => resolve());
				child.stdin.end();
			});
		},
	};
}

const missingBinaryPrefix = "containerkit/live: cannot find containerkit at ";

/** True when `err` means the containerkit binary was not on `PATH` or at `CONTAINERKIT_BIN`. */
export function isMissingBinaryError(err: unknown): err is Error {
	return err instanceof Error && err.message.startsWith(missingBinaryPrefix);
}

function missingBinary(bin: string): Error {
	return new Error(`${missingBinaryPrefix}${bin} (set CONTAINERKIT_BIN or PATH)`);
}

function attachTerminal(bin: string, container: string, command?: readonly string[]): Promise<void> {
	if (container === "") {
		return Promise.reject(new Error("containerkit/live: terminal needs a container"));
	}
	const args = ["terminal", container];
	if (command && command.length > 0) {
		args.push(...command);
	}
	return new Promise((resolve, reject) => {
		const child = spawn(bin, args, { stdio: "inherit" });
		child.on("error", (err: NodeJS.ErrnoException) => {
			if (err.code === "ENOENT") {
				reject(missingBinary(bin));
				return;
			}
			reject(err);
		});
		child.on("exit", (status, signal) => {
			if (signal) {
				reject(new Error(`containerkit terminal exited on ${signal}`));
				return;
			}
			if (status !== 0) {
				reject(new Error(`containerkit terminal exited ${status ?? 1}`));
				return;
			}
			resolve();
		});
	});
}

export function sessionFromTransport(client: Client, transport: SessionTransport, options?: { bin?: string }): Session {
	const listeners = new Map<string, Set<(event: LiveEvent) => void>>();
	transport.on("event", (event) => {
		const set = listeners.get(event.kind);
		if (!set) {
			return;
		}
		for (const fn of set) {
			fn(event);
		}
	});
	const call = <T = unknown>(plan: PlanDocument) => transport.call(plan) as Promise<T>;
	const core = {
		client,
		call,
		on<K extends keyof LiveEventMap>(kind: K, handler: (event: LiveEventMap[K]) => void) {
			let set = listeners.get(kind);
			if (!set) {
				set = new Set();
				listeners.set(kind, set);
			}
			const wrapped = handler as (event: LiveEvent) => void;
			set.add(wrapped);
			return () => {
				set.delete(wrapped);
			};
		},
		close: () => transport.close(),
		terminal(container: string, terminalOptions?: TerminalOptions) {
			const bin = options?.bin ?? process.env.CONTAINERKIT_BIN ?? "containerkit";
			return attachTerminal(bin, container, terminalOptions?.command);
		},
	};
	const bound = new Map<string, (...args: unknown[]) => Promise<unknown>>();
	return new Proxy(core, {
		get(target, prop, receiver) {
			if (typeof prop !== "string" || prop in target) {
				return Reflect.get(target, prop, receiver);
			}
			const command = commandByName[prop];
			if (!command) {
				return undefined;
			}
			let fn = bound.get(prop);
			if (!fn) {
				fn = (...args: unknown[]) => call(command(client, ...(args as never[])));
				bound.set(prop, fn);
			}
			return fn;
		},
	}) as Session;
}

export function clientFromRoot(options: { fromEnv?: boolean; connect?: Client["options"] }): Client {
	return connect({ ...(options.connect ?? {}), ...(options.fromEnv ? { fromEnv: true } : {}) });
}
