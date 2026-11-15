import assert from "node:assert/strict";
import { test } from "node:test";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { binaryInvoke } from "./runner.js";
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
