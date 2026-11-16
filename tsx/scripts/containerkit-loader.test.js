import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { test } from "node:test";
import { mapSpecifier } from "./containerkit-loader.mjs";

test("the bare containerkit name maps onto the published package", () => {
	assert.equal(mapSpecifier("containerkit"), "@aptd3v/containerkit");
	assert.equal(mapSpecifier("containerkit/jsx-runtime"), "@aptd3v/containerkit/jsx-runtime");
	assert.equal(mapSpecifier("containerkit/live"), "@aptd3v/containerkit/live");
	assert.equal(mapSpecifier("./index.js"), "./index.js");
	assert.equal(mapSpecifier("@aptd3v/containerkit"), "@aptd3v/containerkit");
});

test("node loads containerkit/jsx-runtime through the loader", () => {
	const result = spawnSync(process.execPath, [
		"--import",
		"./scripts/register-containerkit.mjs",
		"--input-type=module",
		"--eval",
		"await import('containerkit/jsx-runtime')",
	], { encoding: "utf8" });
	assert.equal(result.status, 0, result.stderr);
});
