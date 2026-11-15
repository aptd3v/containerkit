import { spawn } from "node:child_process";
import type { Invoke, PlanDocument } from "./plan.js";

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

export async function runWithBinary(doc: PlanDocument, bin?: string): Promise<unknown> {
	const path = bin ?? resolveBinary() ?? "containerkit";
	return binaryInvoke(path)("exec", doc);
}

function spawnBin(bin: string, args: string[], stdin: string): Promise<string> {
	return new Promise((resolve, reject) => {
		const child = spawn(bin, args, { stdio: ["pipe", "pipe", "pipe"] });
		let stdout = "";
		let stderr = "";
		child.stdout?.on("data", (chunk: unknown) => {
			stdout += String(chunk);
		});
		child.stderr?.on("data", (chunk: unknown) => {
			stderr += String(chunk);
		});
		child.on("error", (err: Error) => {
			reject(wrapSpawnError(bin, err));
		});
		child.on("close", (code: number | null) => {
			if (code !== 0) {
				reject(new Error(stderr.trim() || `${bin} exited ${code ?? "null"}`));
				return;
			}
			resolve(stdout);
		});
		child.stdin?.end(stdin);
	});
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
