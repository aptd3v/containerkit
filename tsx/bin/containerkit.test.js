import assert from "node:assert/strict";
import { mkdtempSync, writeFileSync, chmodSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { packageName, binaryName, resolveContainerkit, run } from "./containerkit.js";

test("platform package names follow the Go archive names", () => {
	assert.equal(packageName("darwin", "arm64"), "@containerkit/darwin-arm64");
	assert.equal(packageName("darwin", "x64"), "@containerkit/darwin-amd64");
	assert.equal(packageName("linux", "arm64"), "@containerkit/linux-arm64");
	assert.equal(packageName("linux", "x64"), "@containerkit/linux-amd64");
	assert.equal(packageName("win32", "x64"), "@containerkit/windows-amd64");
	assert.equal(packageName("win32", "arm64"), "@containerkit/windows-arm64");
	assert.equal(binaryName("win32"), "containerkit.exe");
	assert.equal(binaryName("linux"), "containerkit");
});

test("resolveContainerkit returns the first path that exists", () => {
	const hit = resolveContainerkit("linux", "x64", (path) => path.endsWith("linux-amd64/bin/containerkit"));
	assert.equal(hit?.endsWith("npm/@containerkit/linux-amd64/bin/containerkit"), true);
	assert.equal(resolveContainerkit("linux", "x64", () => false), undefined);
});

test("run executes the selected binary", () => {
	const dir = mkdtempSync(join(tmpdir(), "containerkit-bin-"));
	const bin = join(dir, "containerkit");
	writeFileSync(bin, "#!/bin/sh\nprintf '%s\\n' \"$@\"\n");
	chmodSync(bin, 0o755);
	const previous = process.env.CONTAINERKIT_BIN;
	process.env.CONTAINERKIT_BIN = bin;
	try {
		assert.equal(resolveContainerkit(), bin);
		assert.equal(run(["version"]), 0);
	} finally {
		if (previous === undefined) {
			delete process.env.CONTAINERKIT_BIN;
		} else {
			process.env.CONTAINERKIT_BIN = previous;
		}
	}
});
