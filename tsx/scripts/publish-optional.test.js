import assert from "node:assert/strict";
import { test } from "node:test";
import { registryDependencies } from "./publish-optional.mjs";

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
