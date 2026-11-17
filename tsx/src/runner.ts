import { spawn } from "node:child_process";
import type { LiveEvent } from "./live/session.js";
import type { Invoke, PlanDocument, RunProgress } from "./plan.js";

export type BinaryInvokeOptions = {
	bin?: string;
	args?: string[];
};

export function binaryInvoke(opts: string | BinaryInvokeOptions = {}): Invoke {
	const bin = typeof opts === "string" ? opts : (opts.bin ?? process.env.CONTAINERKIT_BIN ?? "containerkit");
	const extra = typeof opts === "string" ? [] : (opts.args ?? []);
	return async (op, payload) => {
		const cmd = op === "exec" || op === "validate" ? op : "exec";
		const text = await spawnBin(bin, [...extra, cmd, "-"], JSON.stringify(payload) + "\n");
		const trimmed = text.trim();
		if (!trimmed) {
			return undefined;
		}
		try {
			return JSON.parse(trimmed) as unknown;
		} catch {
			return trimmed;
		}
	};
}

export function resolveBinary(): string | undefined {
	const bin = process.env.CONTAINERKIT_BIN;
	if (bin) {
		return bin;
	}
	return undefined;
}

export type RunBinaryOptions = {
	bin?: string;
	args?: string[];
	progress?: RunProgress;
	onEvent?: (event: LiveEvent) => void;
	/** Receives stderr that should be shown. The default writes to this process's stderr. */
	writeStderr?: (chunk: string) => void;
};

export async function runWithBinary(doc: PlanDocument, binOrOptions?: string | RunBinaryOptions): Promise<unknown> {
	const options: RunBinaryOptions = typeof binOrOptions === "string" || binOrOptions === undefined ? { bin: binOrOptions } : binOrOptions;
	const progress = doc.progress ?? options.progress ?? "auto";
	const payload: PlanDocument = { ...doc, progress };
	const bin = options.bin ?? resolveBinary() ?? "containerkit";
	const write = options.writeStderr ?? ((chunk: string) => {
		process.stderr.write(chunk);
	});
	const router = createStderrRouter({ progress, onEvent: options.onEvent, write });
	const text = await spawnBin(bin, [...(options.args ?? []), "exec", "-"], JSON.stringify(payload) + "\n", {
		stderr(chunk) {
			router.push(chunk);
		},
		close() {
			router.flush();
		},
	});
	const trimmed = text.trim();
	if (!trimmed) {
		return undefined;
	}
	try {
		return JSON.parse(trimmed) as unknown;
	} catch {
		return trimmed;
	}
}

type SpawnHooks = {
	stderr?: (chunk: string) => void;
	close?: () => void;
};

function spawnBin(bin: string, args: string[], stdin: string, hooks?: SpawnHooks): Promise<string> {
	return new Promise((resolve, reject) => {
		const child = spawn(bin, args, { stdio: ["pipe", "pipe", "pipe"] });
		let stdout = "";
		let stderr = "";
		child.stdout?.on("data", (chunk: unknown) => {
			stdout += String(chunk);
		});
		child.stderr?.on("data", (chunk: unknown) => {
			const text = String(chunk);
			stderr += text;
			hooks?.stderr?.(text);
		});
		child.on("error", (err: Error) => {
			reject(wrapSpawnError(bin, err));
		});
		child.on("close", (code: number | null) => {
			hooks?.close?.();
			if (code !== 0) {
				reject(new Error(stderr.trim() || `${bin} exited ${code ?? "null"}`));
				return;
			}
			resolve(stdout);
		});
		child.stdin?.end(stdin);
	});
}

/**
 * Split runner stderr into progress and text.
 * JSON events go to `onEvent` and are not written when a handler is set.
 * Any other line is written, including a failure message. `off` writes nothing.
 */
export function createStderrRouter(options: {
	progress: RunProgress;
	onEvent?: (event: LiveEvent) => void;
	write: (chunk: string) => void;
}): { push(chunk: string): void; flush(): void } {
	let pending = "";
	const routeLine = (line: string) => {
		if (options.progress === "off") {
			return;
		}
		if (options.progress !== "json") {
			options.write(`${line}\n`);
			return;
		}
		const event = parseLiveEvent(line);
		if (!event) {
			if (line.trim() !== "") {
				options.write(`${line}\n`);
			}
			return;
		}
		options.onEvent?.(event);
		if (!options.onEvent) {
			options.write(`${line}\n`);
		}
	};
	return {
		push(chunk: string) {
			pending += chunk;
			const parts = pending.split("\n");
			pending = parts.pop() ?? "";
			for (const line of parts) {
				routeLine(line);
			}
		},
		flush() {
			if (pending === "") {
				return;
			}
			const line = pending;
			pending = "";
			routeLine(line);
		},
	};
}

function parseLiveEvent(line: string): LiveEvent | undefined {
	const trimmed = line.trim();
	if (!trimmed.startsWith("{")) {
		return undefined;
	}
	try {
		const value = JSON.parse(trimmed) as Partial<LiveEvent>;
		if (value.type !== "event") {
			return undefined;
		}
		if (value.kind !== "log" && value.kind !== "status" && value.kind !== "error") {
			return undefined;
		}
		return value as LiveEvent;
	} catch {
		return undefined;
	}
}

function wrapSpawnError(bin: string, err: Error): Error {
	const code = (err as Error & { code?: string }).code;
	if (code === "ENOENT") {
		return new Error(
			`containerkit/tsx: run() needs containerkit at ${bin} (set CONTAINERKIT_BIN or PATH) or setInvoke(); encode() is the no-I/O path`,
		);
	}
	return err;
}
