import assert from "node:assert/strict";
import { test } from "node:test";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { binaryInvoke, createStderrRouter } from "./runner.js";
import type { LiveEvent } from "./live/session.js";
import { connect } from "./plan.js";
import { listImages } from "./generated-actions.js";

const fake = resolve(dirname(fileURLToPath(import.meta.url)), "..", "testdata", "fake-containerkit.mjs");

test("binaryInvoke execs a runner and parses JSON stdout", async () => {
	const invoke = binaryInvoke({ bin: process.execPath, args: [fake] });
	const doc = listImages(connect({ fromEnv: true }), { all: true });
	const out = (await invoke("exec", doc)) as { ok: boolean; bytes: number };
	assert.equal(out.ok, true);
	assert.ok(out.bytes > 0);
});

test("json stderr becomes onEvent and is not forwarded", () => {
	const forwarded: string[] = [];
	const seen: LiveEvent[] = [];
	const router = createStderrRouter({
		progress: "json",
		onEvent: (event) => {
			seen.push(event);
		},
		write: (chunk) => {
			forwarded.push(chunk);
		},
	});
	router.push('{"type":"event","kind":"log","line":"hi","group":"log"}\nnot json\n');
	router.flush();
	assert.equal(seen.length, 1);
	assert.equal(seen[0]?.kind, "log");
	assert.equal(seen[0]?.line, "hi");
	assert.deepEqual(forwarded, ["not json\n"]);
});
