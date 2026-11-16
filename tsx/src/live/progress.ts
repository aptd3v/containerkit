import type { LiveEvent } from "./session.js";
import { formatProgress } from "./session.js";

/** Same frames as the Go logger's live region. */
const spinFrames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];

const settledGlyph = "●";
const doneGlyph = "✓";
const errorGlyph = "✖";

/** Visible column of the status glyph in a formatEvent row. */
const glyphColumn = 9;

export type Progress = {
	push(event: LiveEvent): void;
	/** One spinner frame. The timer calls this. Tests call it directly. */
	tick(): void;
	/** Text that should scroll above the live rows, such as a console.log. */
	aside(text: string): void;
	/** True while a working row is on screen. */
	active(): boolean;
	stop(): void;
};

export type ProgressOptions = {
	/** Redraw working rows in place. A pipe prints one line per event. */
	tty?: boolean;
	write?: (chunk: string) => void;
	/** Terminal width. 0 means each line is one row. */
	columns?: number;
	/** Spinner period. 0 leaves ticking to the caller. */
	intervalMs?: number;
};

/**
 * Print session events the way the Go logger does on a terminal.
 * Working rows stay at the bottom and spin. Finished rows scroll above them.
 */
export function startProgress(options: ProgressOptions = {}): Progress {
	const tty = options.tty === true;
	const columns = options.columns ?? 0;
	const intervalMs = options.intervalMs ?? (tty ? 100 : 0);
	const live: string[] = [];
	const rows = new Map<string, string>();
	let pending: string[] = [];
	let asideRest = "";
	let drawn = 0;
	let spin = 0;
	let timer: ReturnType<typeof setInterval> | undefined;
	let stopped = false;
	let own = false;
	let restore: () => void = () => undefined;
	const lastDone = new Map<string, string>();
	const direct = options.write ?? ((chunk: string) => {
		process.stdout.write(chunk);
	});
	const write = (chunk: string) => {
		own = true;
		try {
			direct(chunk);
		} finally {
			own = false;
		}
	};

	const paint = () => {
		let chunk = "";
		if (drawn > 0) {
			chunk += `\x1b[${drawn}A`;
		}
		const held = pending;
		pending = [];
		let written = 0;
		for (const line of held) {
			written += writeRow(line);
		}
		let rowsHigh = 0;
		const glyph = spin < 0 ? settledGlyph : spinFrames[spin % spinFrames.length];
		for (const key of live) {
			const line = spin < 0 ? rows.get(key) ?? "" : spinLine(rows.get(key) ?? "", glyph);
			rowsHigh += writeRow(line);
		}
		const extra = Math.max(0, drawn - written - rowsHigh);
		for (let i = 0; i < extra; i++) {
			chunk += "\r\x1b[2K\n";
		}
		if (extra > 0) {
			chunk += `\x1b[${extra}A`;
		}
		drawn = rowsHigh;
		function writeRow(line: string): number {
			const high = visualRows(line, columns);
			if (high <= 1) {
				chunk += `\r\x1b[2K${line}\n`;
				return 1;
			}
			for (let i = 0; i < high; i++) {
				chunk += "\r\x1b[2K";
				if (i + 1 < high) {
					chunk += "\n";
				}
			}
			chunk += `\x1b[${high - 1}A\r${line}\n`;
			return high;
		}
		if (chunk) {
			write(chunk);
		}
	};

	const arm = () => {
		if (!tty || timer || intervalMs <= 0 || stopped) {
			return;
		}
		timer = setInterval(() => {
			if (live.length === 0) {
				clearInterval(timer);
				timer = undefined;
				return;
			}
			spin = (spin + 1) % spinFrames.length;
			paint();
		}, intervalMs);
		timer.unref?.();
	};

	const retireProjectUp = () => {
		let dropped = false;
		for (const key of [...live]) {
			const event = rowEvent.get(key);
			if (event && projectUp(event)) {
				remove(key);
				dropped = true;
			}
		}
		return dropped;
	};

	const rowEvent = new Map<string, LiveEvent>();

	const remove = (key: string) => {
		rowEvent.delete(key);
		rows.delete(key);
		const at = live.indexOf(key);
		if (at >= 0) {
			live.splice(at, 1);
		}
	};

	// A finished pull or build does not send done for every layer. Those
	// extracting/downloading rows would spin for the rest of the session.
	const settlePhase = (id: string, status: "done" | "error") => {
		if (!id) {
			return;
		}
		const prefix = `${id}\0`;
		const glyph = status === "error" ? errorGlyph : doneGlyph;
		const keys = live.filter((key) => key === id || key.startsWith(prefix));
		for (const key of keys) {
			const row = rows.get(key) ?? "";
			remove(key);
			if (row) {
				pending.push(spinLine(row, glyph));
			}
		}
		if (live.length === 0 && timer) {
			clearInterval(timer);
			timer = undefined;
		}
	};

	const upsert = (key: string, event: LiveEvent, line: string) => {
		if (!rows.has(key)) {
			live.push(key);
		}
		rows.set(key, line);
		rowEvent.set(key, event);
	};

	const commit = (line: string) => {
		pending.push(line);
		paint();
	};

	const progress: Progress = {
		push(event) {
			const line = formatProgress(event);
			if (!line) {
				return;
			}
			if (!tty) {
				write(`${line}\n`);
				return;
			}
			if (event.group === "compose" && event.status === "done" && repeated(event)) {
				remove(rowKey(event));
				paint();
				return;
			}
			if (event.kind === "log" && live.length > 0) {
				pending.push(line);
				if (!projectUp(event)) {
					retireProjectUp();
				}
				return;
			}
			if (!projectUp(event)) {
				retireProjectUp();
			}
			const key = rowKey(event);
			if (event.status === "working") {
				upsert(key, event, line);
				paint();
				arm();
				return;
			}
			if (event.status === "done" || event.status === "error") {
				remove(key);
				pending.push(line);
				settlePhase((event.source ?? "").trim(), event.status);
				paint();
				return;
			}
			commit(line);
		},
		tick() {
			if (!tty || live.length === 0) {
				if (pending.length > 0) {
					paint();
				}
				return;
			}
			spin = (spin + 1) % spinFrames.length;
			paint();
		},
		aside(text) {
			if (!tty) {
				write(text);
				return;
			}
			const all = asideRest + text;
			const lines = all.split("\n");
			asideRest = lines.pop() ?? "";
			let any = false;
			for (const line of lines) {
				if (line.length > 0) {
					pending.push(line);
					any = true;
				}
			}
			if (any) {
				paint();
			}
		},
		active() {
			return tty && (live.length > 0 || drawn > 0);
		},
		stop() {
			if (stopped) {
				return;
			}
			stopped = true;
			if (timer) {
				clearInterval(timer);
				timer = undefined;
			}
			spin = -1;
			if (asideRest) {
				pending.push(asideRest);
				asideRest = "";
			}
			if (live.length > 0 || pending.length > 0 || drawn > 0) {
				paint();
			}
			restore();
		},
	};
	if (tty && !options.write) {
		restore = captureStdout(progress, () => own);
	}
	return progress;

	function repeated(event: LiveEvent): boolean {
		const key = `${event.source ?? ""}\0${event.parent ?? ""}`;
		const sig = `${event.stream ?? ""}\0${event.detail ?? ""}`;
		if (lastDone.get(key) === sig) {
			return true;
		}
		lastDone.set(key, sig);
		return false;
	}
}

/** Identity of one live row. Matches the Go logger's event key. */
function rowKey(event: LiveEvent): string {
	let id = (event.source ?? "").trim();
	let parent = (event.parent ?? "").trim();
	if (parent === id) {
		parent = "";
	}
	if (id && parent) {
		return `${parent}\0${id}`;
	}
	if (id) {
		return id;
	}
	if (parent) {
		return parent;
	}
	return `${event.group ?? ""}\0${event.stream ?? ""}`;
}

function projectUp(event: LiveEvent): boolean {
	return event.group === "compose" && event.status === "working" && event.stream === "up" && (event.source ?? "").trim() !== "" && !(event.text ?? "").trim() && !(event.detail ?? "").trim() && !(typeof event.total === "number" && event.total > 0);
}

/** Swap the status glyph, leaving a line alone when that column is not the settled dot. */
function spinLine(line: string, glyph: string): string {
	const at = columnAt(line, glyphColumn);
	if (!at) {
		return line;
	}
	const current = line.slice(at.start, at.end);
	if (current !== settledGlyph && !spinFrames.includes(current)) {
		return line;
	}
	return line.slice(0, at.start) + glyph + line.slice(at.end);
}

function columnAt(line: string, column: number): { start: number; end: number } | undefined {
	let col = 0;
	for (let i = 0; i < line.length; ) {
		if (line[i] === "\u001b" && line[i + 1] === "[") {
			i += 2;
			while (i < line.length && !isCsiFinal(line[i])) {
				i++;
			}
			if (i < line.length) {
				i++;
			}
			continue;
		}
		const cp = line.codePointAt(i) ?? 0;
		const size = cp > 0xffff ? 2 : 1;
		if (col === column) {
			return { start: i, end: i + size };
		}
		col++;
		i += size;
	}
	return undefined;
}

function isCsiFinal(char: string): boolean {
	const code = char.charCodeAt(0);
	return (code >= 0x40 && code <= 0x7e);
}

function captureStdout(progress: Progress, inside: () => boolean): () => void {
	const stdout = process.stdout;
	const original = stdout.write.bind(stdout) as typeof stdout.write;
	const wrapped = ((chunk: string | Uint8Array, encoding?: BufferEncoding | ((err?: Error) => void), cb?: (err?: Error) => void) => {
		if (inside() || !progress.active()) {
			return original(chunk as never, encoding as never, cb as never);
		}
		const text = typeof chunk === "string" ? chunk : Buffer.from(chunk).toString("utf8");
		progress.aside(text);
		const done = typeof encoding === "function" ? encoding : cb;
		done?.();
		return true;
	}) as typeof stdout.write;
	stdout.write = wrapped;
	return () => {
		if (stdout.write === wrapped) {
			stdout.write = original;
		}
	};
}

function visualRows(line: string, width: number): number {
	if (width <= 0) {
		return 1;
	}
	const cols = visibleCols(line);
	if (cols === 0) {
		return 1;
	}
	return Math.ceil(cols / width);
}

function visibleCols(line: string): number {
	let n = 0;
	for (let i = 0; i < line.length; ) {
		if (line[i] === "\u001b" && line[i + 1] === "[") {
			i += 2;
			while (i < line.length && !isCsiFinal(line[i])) {
				i++;
			}
			if (i < line.length) {
				i++;
			}
			continue;
		}
		const cp = line.codePointAt(i) ?? 0;
		i += cp > 0xffff ? 2 : 1;
		n++;
	}
	return n;
}
