function needsQuote(s: string): boolean {
	if (s === "") {
		return true;
	}
	if (/[\n\r:#{}[\],&*!|>'"%@`]/.test(s) || s.includes(": ") || /^\s|\s$/.test(s)) {
		return true;
	}
	if (/^(?:true|false|null|~|yes|no|on|off)$/i.test(s)) {
		return true;
	}
	if (/^-?\d+(\.\d+)?$/.test(s)) {
		return true;
	}
	return false;
}

function fmtKey(k: string): string {
	return needsQuote(k) ? JSON.stringify(k) : k;
}

function fmtScalar(v: unknown): string {
	if (v === null || v === undefined) {
		return "null";
	}
	if (typeof v === "boolean") {
		return v ? "true" : "false";
	}
	if (typeof v === "number") {
		return Number.isFinite(v) ? String(v) : JSON.stringify(String(v));
	}
	if (typeof v === "string") {
		return needsQuote(v) ? JSON.stringify(v) : v;
	}
	return JSON.stringify(String(v));
}

function isPlainObject(v: unknown): v is Record<string, unknown> {
	return v !== null && typeof v === "object" && !Array.isArray(v);
}

export function toYAML(value: unknown): string {
	return dump(value, 0).replace(/\n+$/, "") + "\n";
}

function dump(value: unknown, indent: number): string {
	const pad = "  ".repeat(indent);
	if (Array.isArray(value)) {
		if (value.length === 0) {
			return "[]\n";
		}
		let out = "";
		for (const item of value) {
			if (isPlainObject(item)) {
				const entries = Object.entries(item).filter(([, v]) => v !== undefined);
				if (entries.length === 0) {
					out += `${pad}- {}\n`;
					continue;
				}
				const [firstKey, firstVal] = entries[0];
				if (isPlainObject(firstVal) || Array.isArray(firstVal)) {
					out += `${pad}- ${fmtKey(firstKey)}:\n${dump(firstVal, indent + 2)}`;
				} else {
					out += `${pad}- ${fmtKey(firstKey)}: ${fmtScalar(firstVal)}\n`;
				}
				for (const [k, v] of entries.slice(1)) {
					out += dumpPair(k, v, indent + 1);
				}
				continue;
			}
			if (Array.isArray(item)) {
				out += `${pad}-\n${dump(item, indent + 1)}`;
				continue;
			}
			out += `${pad}- ${fmtScalar(item)}\n`;
		}
		return out;
	}
	if (isPlainObject(value)) {
		const entries = Object.entries(value).filter(([, v]) => v !== undefined);
		if (entries.length === 0) {
			return "{}\n";
		}
		let out = "";
		for (const [k, v] of entries) {
			out += dumpPair(k, v, indent);
		}
		return out;
	}
	return `${pad}${fmtScalar(value)}\n`;
}

function dumpPair(k: string, v: unknown, indent: number): string {
	const pad = "  ".repeat(indent);
	if (v === undefined) {
		return "";
	}
	if (isPlainObject(v) || Array.isArray(v)) {
		if (isPlainObject(v) && Object.keys(v).length === 0) {
			return `${pad}${fmtKey(k)}: {}\n`;
		}
		if (Array.isArray(v) && v.length === 0) {
			return `${pad}${fmtKey(k)}: []\n`;
		}
		return `${pad}${fmtKey(k)}:\n${dump(v, indent + 1)}`;
	}
	return `${pad}${fmtKey(k)}: ${fmtScalar(v)}\n`;
}
