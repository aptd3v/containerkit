import { execSync } from "node:child_process";
import { existsSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const packageDir = join(dirname(fileURLToPath(import.meta.url)), "..");
const packagePath = join(packageDir, "package.json");
const backupPath = join(packageDir, "package.json.publish-backup");
const repoRoot = join(packageDir, "..");

// Keep this list aligned with scripts/version.py PLATFORMS.
const platforms = [
	["darwin-amd64", "containerkit"],
	["darwin-arm64", "containerkit"],
	["linux-amd64", "containerkit"],
	["linux-arm64", "containerkit"],
	["windows-amd64", "containerkit.exe"],
	["windows-arm64", "containerkit.exe"],
];

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

export function assertReady(pkg, versionText) {
	if (pkg.private) {
		throw new Error("package is private; not publishing");
	}
	const version = String(versionText).trim();
	if (pkg.version !== version) {
		throw new Error(`package.json is ${pkg.version}, VERSION is ${version}`);
	}
}

export function missingBins(root = repoRoot) {
	const missing = [];
	for (const [slug, exe] of platforms) {
		const path = join(root, "npm", "@containerkit", slug, "bin", exe);
		if (!existsSync(path)) {
			missing.push(path);
		}
	}
	return missing;
}

function check() {
	const pkg = JSON.parse(readFileSync(packagePath, "utf8"));
	const version = readFileSync(join(repoRoot, "VERSION"), "utf8");
	assertReady(pkg, version);
	const missing = missingBins();
	if (missing.length > 0) {
		throw new Error(`missing platform binaries:\n${missing.join("\n")}`);
	}
}

export function publish() {
	const token = process.env.NODE_AUTH_TOKEN || process.env.NPM_TOKEN;
	if (!token) {
		throw new Error("NODE_AUTH_TOKEN or NPM_TOKEN is required");
	}
	check();
	const env = { ...process.env, NODE_AUTH_TOKEN: token, NPM_TOKEN: token };
	for (const [slug] of platforms) {
		execSync("npm publish --access public", {
			cwd: join(repoRoot, "npm", "@containerkit", slug),
			stdio: "inherit",
			env,
		});
	}
	execSync("npm publish --access public", {
		cwd: packageDir,
		stdio: "inherit",
		env,
	});
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
	if (cmd === "check") {
		check();
		return;
	}
	if (cmd === "publish") {
		publish();
		return;
	}
	throw new Error("usage: publish-optional.mjs rewrite|restore|check|publish");
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
