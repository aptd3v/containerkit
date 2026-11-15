import { spawn, type ChildProcessWithoutNullStreams } from "node:child_process";
import { createInterface } from "node:readline";
import { connect, type Client, type PlanDocument } from "../plan.js";
import { commandByName, type SessionCommands } from "./commands.js";
import { ContainerKitError, errorFromFrame, type ErrorFrame } from "./error.js";

export type LiveEvent = {
	type: "event";
	kind: "log" | "status" | "error";
	source?: string;
	stream?: string;
	line?: string;
	message?: string;
	class?: string;
};

export type Session = SessionCommands & {
	readonly client: Client;
	call(plan: PlanDocument): Promise<unknown>;
	on(kind: "log" | "status" | "error", handler: (event: LiveEvent) => void): () => void;
	close(): Promise<void>;
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

export function openSession(options?: { bin?: string; args?: string[] }): SessionTransport {
	const bin = options?.bin ?? process.env.CONTAINERKIT_BIN ?? "containerkit";
	const args = [...(options?.args ?? []), "session"];
	const child = spawn(bin, args, { stdio: ["pipe", "pipe", "pipe"] }) as ChildProcessWithoutNullStreams;
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
			failAll(new Error(`containerkit/live: session needs containerkit at ${child.spawnfile}`));
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

export function sessionFromTransport(client: Client, transport: SessionTransport): Session {
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
	const call = (plan: PlanDocument) => transport.call(plan);
	const core = {
		client,
		call,
		on(kind: "log" | "status" | "error", handler: (event: LiveEvent) => void) {
			let set = listeners.get(kind);
			if (!set) {
				set = new Set();
				listeners.set(kind, set);
			}
			set.add(handler);
			return () => {
				set.delete(handler);
			};
		},
		close: () => transport.close(),
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
