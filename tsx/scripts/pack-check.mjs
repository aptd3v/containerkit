import { execSync } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const required = [
	"package.json",
	"README.md",
	"GUIDE.md",
	"LICENSE",
	"tsconfig.jsx.json",
	"bin/containerkit.js",
	"dist/index.js",
	"dist/index.d.ts",
	"dist/jsx-runtime.js",
	"dist/docker-file.js",
	"dist/live/index.js",
	"dist/hooks/index.js",
	"dist/vocab.js",
	"docs/live.md",
	"docs/README.md",
];

const forbidden = ["examples/", "src/", "dist-test/", "dist-examples/", "node_modules/", "testdata/", ".env"];

const dest = mkdtempSync(join(tmpdir(), "containerkit-pack-"));
try {
	const raw = execSync("npm pack --dry-run --json --pack-destination " + JSON.stringify(dest), {
		encoding: "utf8",
	});
	const packed = JSON.parse(raw);
	const files = (packed[0]?.files ?? []).map((file) => file.path);
	const missing = required.filter((path) => !files.includes(path));
	const leaked = files.filter((path) => forbidden.some((prefix) => path === prefix.slice(0, -1) || path.startsWith(prefix)));
	const tests = files.filter((path) => path.includes(".test."));
	if (missing.length > 0 || leaked.length > 0 || tests.length > 0) {
		const lines = [];
		if (missing.length > 0) {
			lines.push(`missing from the tarball:\n${missing.join("\n")}`);
		}
		if (leaked.length > 0) {
			lines.push(`tarball includes files that must stay out:\n${leaked.join("\n")}`);
		}
		if (tests.length > 0) {
			lines.push(`tarball includes tests:\n${tests.join("\n")}`);
		}
		throw new Error(lines.join("\n"));
	}
	process.stdout.write(`pack ok: ${files.length} files\n`);
} finally {
	rmSync(dest, { recursive: true, force: true });
}
