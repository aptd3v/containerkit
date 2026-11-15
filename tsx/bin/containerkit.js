#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { existsSync, realpathSync } from "node:fs";
import { createRequire } from "node:module";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));

// packageName is the platform package npm installs for this machine.
// Go calls x64 amd64 and calls Windows windows. Node does not.
export function packageName(platform = process.platform, arch = process.arch) {
	const os = platform === "win32" ? "windows" : platform;
	const cpu = arch === "x64" ? "amd64" : arch;
	return `@containerkit/${os}-${cpu}`;
}

export function binaryName(platform = process.platform) {
	return platform === "win32" ? "containerkit.exe" : "containerkit";
}

function isThisScript(path) {
	try {
		return realpathSync(path) === realpathSync(fileURLToPath(import.meta.url));
	} catch {
		return false;
	}
}

// candidateBinaries lists where the Go binary may be. The installed platform
// package is first. The repo copy is the path crossbuild writes for a local install.
export function candidateBinaries(platform = process.platform, arch = process.arch) {
	const name = packageName(platform, arch);
	const exe = binaryName(platform);
	const slug = name.slice("@containerkit/".length);
	const found = [];
	const fromEnv = process.env.CONTAINERKIT_BIN;
	if (fromEnv && !isThisScript(fromEnv)) {
		found.push(fromEnv);
	}
	try {
		const require = createRequire(import.meta.url);
		found.push(require.resolve(`${name}/bin/${exe}`));
	} catch {
		// The optional package is not installed in this tree.
	}
	found.push(join(here, "..", "..", "npm", "@containerkit", slug, "bin", exe));
	return found;
}

export function resolveContainerkit(platform = process.platform, arch = process.arch, exists = existsSync) {
	for (const path of candidateBinaries(platform, arch)) {
		if (exists(path)) {
			return path;
		}
	}
	return undefined;
}

export function run(argv = process.argv.slice(2)) {
	const bin = resolveContainerkit();
	if (!bin) {
		process.stderr.write(
			`containerkit: no binary for ${packageName()}. Install that package, run go run ./tools/crossbuild, or set CONTAINERKIT_BIN.\n`,
		);
		return 1;
	}
	const child = spawnSync(bin, argv, { stdio: "inherit" });
	if (child.error) {
		process.stderr.write(`${child.error.message}\n`);
		return 1;
	}
	return child.status ?? 1;
}

function invokedDirectly() {
	const entry = process.argv[1];
	if (!entry) {
		return false;
	}
	return isThisScript(entry);
}

if (invokedDirectly()) {
	process.exit(run());
}
