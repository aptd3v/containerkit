import assert from "node:assert/strict";
import { mkdtempSync, readFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { assertReady, missingBins, publish, registryDependencies } from "./publish-optional.mjs";

const platform = JSON.parse(
	readFileSync(new URL("../../npm/@aptd3v/containerkit-darwin-arm64/package.json", import.meta.url), "utf8"),
);

test("file platform packages become their version", () => {
	const pkg = {
		name: "containerkit",
		version: platform.version,
		optionalDependencies: {
			"@aptd3v/containerkit-darwin-arm64": "file:../npm/@aptd3v/containerkit-darwin-arm64",
			"left-alone": "1.2.3",
		},
	};
	const next = registryDependencies(pkg);
	assert.equal(next.optionalDependencies["@aptd3v/containerkit-darwin-arm64"], platform.version);
	assert.equal(next.optionalDependencies["left-alone"], "1.2.3");
	assert.equal(pkg.optionalDependencies["@aptd3v/containerkit-darwin-arm64"].startsWith("file:"), true);
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
