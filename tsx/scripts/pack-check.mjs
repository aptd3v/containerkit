import { execSync } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const dest = mkdtempSync(join(tmpdir(), "containerkit-pack-"));
try {
	execSync("npm pack --dry-run --pack-destination " + JSON.stringify(dest), {
		stdio: "inherit",
	});
} finally {
	rmSync(dest, { recursive: true, force: true });
}
