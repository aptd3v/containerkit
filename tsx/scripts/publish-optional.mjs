import { execSync } from "node:child_process";
import { chmodSync, copyFileSync, existsSync, mkdirSync, readFileSync, rmSync, unlinkSync, writeFileSync } from "node:fs";
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
	if (!pkg.optionalDependencies) {
		return pkg;
	}
	const optional = { ...pkg.optionalDependencies };
	for (const [name, spec] of Object.entries(optional)) {
		if (!name.startsWith("@aptd3v/containerkit-") || typeof spec !== "string" || !spec.startsWith("file:")) {
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
		const path = join(root, "npm", "@aptd3v", `containerkit-${slug}`, "bin", exe);
		if (!existsSync(path)) {
			missing.push(path);
		}
	}
	return missing;
}

// stageBins copies the six Go binaries next to bin/containerkit.js so the
// published package carries the runner. clearStagedBins removes that copy.
export function stageBins(root = repoRoot) {
	const missing = missingBins(root);
	if (missing.length > 0) {
		throw new Error(`missing platform binaries:\n${missing.join("\n")}`);
	}
	for (const [slug, exe] of platforms) {
		const destDir = join(packageDir, "bin", slug);
		mkdirSync(destDir, { recursive: true });
		const dest = join(destDir, exe);
		copyFileSync(join(root, "npm", "@aptd3v", `containerkit-${slug}`, "bin", exe), dest);
		chmodSync(dest, 0o755);
	}
}

export function clearStagedBins() {
	for (const [slug] of platforms) {
		rmSync(join(packageDir, "bin", slug), { recursive: true, force: true });
	}
}

function check() {
	const pkg = JSON.parse(readFileSync(packagePath, "utf8"));
	const version = readFileSync(join(repoRoot, "VERSION"), "utf8");
	assertReady(pkg, version);
	if (pkg.optionalDependencies) {
		throw new Error("package must not depend on a separate runner");
	}
	const missing = missingBins();
	if (missing.length > 0) {
		throw new Error(`missing platform binaries:\n${missing.join("\n")}`);
	}
}

// publishAuth is "oidc" when GitHub Actions can mint an npm token, and "token"
// for a local publish. The tag workflow uses OIDC and does not pass a secret.
export function publishAuth(env = process.env) {
	const token = env.NODE_AUTH_TOKEN || env.NPM_TOKEN;
	const oidc = env.ACTIONS_ID_TOKEN_REQUEST_URL && env.ACTIONS_ID_TOKEN_REQUEST_TOKEN;
	if (!token && !oidc) {
		throw new Error("NODE_AUTH_TOKEN or NPM_TOKEN is required");
	}
	return oidc ? "oidc" : "token";
}

export function publish() {
	const mode = publishAuth();
	check();
	const env = { ...process.env };
	if (mode === "token") {
		const token = env.NODE_AUTH_TOKEN || env.NPM_TOKEN;
		env.NODE_AUTH_TOKEN = token;
		env.NPM_TOKEN = token;
	}
	execSync("npm publish --access public", {
		cwd: packageDir,
		stdio: "inherit",
		env,
	});
}

function main() {
	const cmd = process.argv[2];
	if (cmd === "prepack") {
		if (existsSync(backupPath)) {
			throw new Error("package.json.publish-backup already exists");
		}
		const raw = readFileSync(packagePath, "utf8");
		writeFileSync(backupPath, raw);
		writeFileSync(packagePath, rewritePackage(raw));
		try {
			stageBins();
		} catch (err) {
			writeFileSync(packagePath, raw);
			unlinkSync(backupPath);
			clearStagedBins();
			throw err;
		}
		return;
	}
	if (cmd === "postpack") {
		if (existsSync(backupPath)) {
			writeFileSync(packagePath, readFileSync(backupPath));
			unlinkSync(backupPath);
		}
		clearStagedBins();
		return;
	}
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
	throw new Error("usage: publish-optional.mjs prepack|postpack|rewrite|restore|check|publish");
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
