import assert from "node:assert/strict";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { assertReady, missingBins, publish, registryDependencies } from "./publish-optional.mjs";

test("file platform packages become their version", () => {
	const pkg = {
		name: "containerkit",
		version: "0.8.1",
		optionalDependencies: {
			"@containerkit/darwin-arm64": "file:../npm/@containerkit/darwin-arm64",
			"left-alone": "1.2.3",
		},
	};
	const next = registryDependencies(pkg);
	assert.equal(next.optionalDependencies["@containerkit/darwin-arm64"], "0.8.1");
	assert.equal(next.optionalDependencies["left-alone"], "1.2.3");
	assert.equal(pkg.optionalDependencies["@containerkit/darwin-arm64"].startsWith("file:"), true);
});

test("a private package or a version mismatch cannot publish", () => {
	assert.throws(() => assertReady({ private: true, version: "0.8.1" }, "0.8.1"), /private/);
	assert.throws(() => assertReady({ version: "0.8.0" }, "0.8.1\n"), /0.8.0/);
	assertReady({ version: "0.8.1" }, "0.8.1\n");
});

test("missing platform binaries are reported from the given root", () => {
	const root = mkdtempSync(join(tmpdir(), "containerkit-bins-"));
	const missing = missingBins(root);
	assert.equal(missing.length, 6);
	assert.equal(missing.some((path) => path.endsWith(join("windows-amd64", "bin", "containerkit.exe"))), true);
});

test("publish refuses to run without a token", () => {
	const prevNode = process.env.NODE_AUTH_TOKEN;
	const prevNpm = process.env.NPM_TOKEN;
	delete process.env.NODE_AUTH_TOKEN;
	delete process.env.NPM_TOKEN;
	try {
		assert.throws(() => publish(), /NODE_AUTH_TOKEN or NPM_TOKEN/);
	} finally {
		if (prevNode === undefined) delete process.env.NODE_AUTH_TOKEN;
		else process.env.NODE_AUTH_TOKEN = prevNode;
		if (prevNpm === undefined) delete process.env.NPM_TOKEN;
		else process.env.NPM_TOKEN = prevNpm;
	}
});
