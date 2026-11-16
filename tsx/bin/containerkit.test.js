import assert from "node:assert/strict";
import { mkdtempSync, writeFileSync, chmodSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { platformSlug, binaryName, resolveContainerkit, run } from "./containerkit.js";

test("platform directories follow the Go archive names", () => {
	assert.equal(platformSlug("darwin", "arm64"), "darwin-arm64");
	assert.equal(platformSlug("darwin", "x64"), "darwin-amd64");
	assert.equal(platformSlug("linux", "arm64"), "linux-arm64");
	assert.equal(platformSlug("linux", "x64"), "linux-amd64");
	assert.equal(platformSlug("win32", "x64"), "windows-amd64");
	assert.equal(platformSlug("win32", "arm64"), "windows-arm64");
	assert.equal(binaryName("win32"), "containerkit.exe");
	assert.equal(binaryName("linux"), "containerkit");
});

test("resolveContainerkit returns the first path that exists", () => {
	const bundled = resolveContainerkit("linux", "x64", (path) => path.endsWith(`${join("bin", "linux-amd64", "containerkit")}`));
	assert.equal(bundled?.endsWith(join("bin", "linux-amd64", "containerkit")), true);
	const built = resolveContainerkit("linux", "x64", (path) => path.endsWith(join("linux-amd64", "bin", "containerkit")));
	assert.equal(built?.endsWith(join("npm", "@aptd3v", "containerkit-linux-amd64", "bin", "containerkit")), true);
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
