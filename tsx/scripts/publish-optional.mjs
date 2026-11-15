import { existsSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const packageDir = join(dirname(fileURLToPath(import.meta.url)), "..");
const packagePath = join(packageDir, "package.json");
const backupPath = join(packageDir, "package.json.publish-backup");
const repoRoot = join(packageDir, "..");

// registryDependencies replaces file: platform packages with their versions.
// npm publish would otherwise keep the file: paths, which a public install cannot use.
export function registryDependencies(pkg, root = repoRoot) {
	const optional = { ...pkg.optionalDependencies };
	for (const [name, spec] of Object.entries(optional)) {
		if (!name.startsWith("@containerkit/") || typeof spec !== "string" || !spec.startsWith("file:")) {
			continue;
		}
		const localPath = join(root, "npm", ...name.split("/"), "package.json");
		const local = JSON.parse(readFileSync(localPath, "utf8"));
		if (local.name !== name || local.version !== pkg.version) {
			throw new Error(`${name} is ${local.version}, containerkit is ${pkg.version}`);
		}
		optional[name] = local.version;
	}
	return { ...pkg, optionalDependencies: optional };
}

export function rewritePackage(raw, root = repoRoot) {
	return JSON.stringify(registryDependencies(JSON.parse(raw), root), null, 2) + "\n";
}

function main() {
	const cmd = process.argv[2];
	if (cmd === "rewrite") {
		if (existsSync(backupPath)) {
			throw new Error("package.json.publish-backup already exists");
		}
		const raw = readFileSync(packagePath, "utf8");
		writeFileSync(backupPath, raw);
		writeFileSync(packagePath, rewritePackage(raw));
		return;
	}
	if (cmd === "restore") {
		if (!existsSync(backupPath)) {
			return;
		}
		writeFileSync(packagePath, readFileSync(backupPath));
		unlinkSync(backupPath);
		return;
	}
	throw new Error("usage: publish-optional.mjs rewrite|restore");
}

function invokedDirectly() {
	return (process.argv[1] ?? "").endsWith("publish-optional.mjs");
}

if (invokedDirectly()) {
	try {
		main();
	} catch (err) {
		process.stderr.write(`${err instanceof Error ? err.message : err}\n`);
		process.exit(1);
	}
}
