import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, renameSync, writeFileSync } from "node:fs";
import { join } from "node:path";

// memoKey identifies a factory plus its dependency list. Object keys are sorted,
// so { a, b } and { b, a } are one entry. Array order still matters.
export function memoKey(factory: () => unknown, deps: readonly unknown[]): string | undefined {
	let body: string;
	try {
		body = JSON.stringify(canonicalize(deps));
	} catch {
		return undefined;
	}
	return createHash("sha256").update(factory.toString()).update("\0").update(body).digest("hex");
}

export function readMemo(dir: string, key: string): { value: unknown } | undefined {
	try {
		const parsed = JSON.parse(readFileSync(join(dir, `${key}.json`), "utf8")) as { value?: unknown };
		return { value: parsed.value };
	} catch {
		return undefined;
	}
}

export function writeMemo(dir: string, key: string, value: unknown): void {
	mkdirSync(dir, { recursive: true });
	const file = join(dir, `${key}.json`);
	const tmp = `${file}.${process.pid}.tmp`;
	writeFileSync(tmp, JSON.stringify({ value }));
	renameSync(tmp, file);
}

function canonicalize(value: unknown): unknown {
	if (value === null) {
		return null;
	}
	const kind = typeof value;
	if (kind === "string" || kind === "boolean") {
		return value;
	}
	if (kind === "number") {
		if (!Number.isFinite(value)) {
			throw new TypeError("memo dep is not finite");
		}
		return value;
	}
	if (kind !== "object") {
		throw new TypeError("memo dep is not JSON");
	}
	if (Array.isArray(value)) {
		return value.map((item) => canonicalize(item));
	}
	const record = value as Record<string, unknown>;
	const out: Record<string, unknown> = {};
	for (const name of Object.keys(record).sort()) {
		out[name] = canonicalize(record[name]);
	}
	return out;
}
