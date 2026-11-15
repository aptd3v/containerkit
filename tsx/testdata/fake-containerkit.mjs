#!/usr/bin/env node
import { readFileSync } from "node:fs";
import { createInterface } from "node:readline";

const cmd = process.argv[2] ?? "";
if (cmd === "session") {
	const rl = createInterface({ input: process.stdin });
	for await (const line of rl) {
		if (!line.trim()) {
			continue;
		}
		const msg = JSON.parse(line);
		process.stdout.write(JSON.stringify({
			id: msg.id,
			type: "result",
			ok: true,
			value: { ok: true, op: msg.plan?.actions?.[0]?.op ?? "" },
		}) + "\n");
	}
	process.exit(0);
}
let input = "";
try {
	input = readFileSync(0, "utf8");
} catch {
	input = "";
}
if (cmd === "validate") {
	process.exit(0);
}
if (cmd === "exec") {
	process.stdout.write(JSON.stringify({ ok: true, bytes: input.length }) + "\n");
	process.exit(0);
}
process.stderr.write(`unknown command ${cmd}\n`);
process.exit(2);
