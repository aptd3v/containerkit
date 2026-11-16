#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { existsSync, realpathSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));

// platformSlug is the directory of the bundled Go binary.
// Go calls x64 amd64 and calls Windows windows. Node does not.
export function platformSlug(platform = process.platform, arch = process.arch) {
	const os = platform === "win32" ? "windows" : platform;
	const cpu = arch === "x64" ? "amd64" : arch;
	return `${os}-${cpu}`;
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

// candidateBinaries lists where the Go binary may be. The copy beside this
// script is the one packed into @aptd3v/containerkit. The repo path is what
// crossbuild writes before that copy exists.
export function candidateBinaries(platform = process.platform, arch = process.arch) {
	const exe = binaryName(platform);
	const slug = platformSlug(platform, arch);
	const found = [];
	const fromEnv = process.env.CONTAINERKIT_BIN;
	if (fromEnv && !isThisScript(fromEnv)) {
		found.push(fromEnv);
	}
	found.push(join(here, slug, exe));
	found.push(join(here, "..", "..", "npm", "@aptd3v", `containerkit-${slug}`, "bin", exe));
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
			`containerkit: no binary for ${platformSlug()}. Run go run ./tools/crossbuild, or set CONTAINERKIT_BIN.\n`,
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
