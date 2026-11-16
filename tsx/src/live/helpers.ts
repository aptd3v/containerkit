import { isCKNotFoundError } from "./error.js";
import type { Session } from "./session.js";

/**
 * Docker platform for this machine.
 *
 * @returns `linux/` plus the process architecture. `x64` is reported as `amd64`.
 *
 * @example
 * ```ts
 * hostPlatform(); // "linux/arm64" on Apple silicon
 * ```
 */
export function hostPlatform(): string {
	const arch = process.arch === "x64" ? "amd64" : process.arch;
	return `linux/${arch}`;
}

/**
 * Pull `ref` when the daemon does not already have it.
 *
 * @param session - Open session
 * @param ref - Image reference, such as `nginx:alpine`
 * @param platform - Platform passed to the pull. Defaults to {@link hostPlatform}
 * @throws {@link ContainerKitError} when the image is missing for a reason other than `notFound`, or when the pull fails
 */
export async function ensureImage(session: Session, ref: string, platform = hostPlatform()): Promise<void> {
	try {
		await session.inspectImage(ref);
	} catch (err) {
		if (!isCKNotFoundError(err)) {
			throw err;
		}
		await session.pullImage(ref, { platform });
	}
}

/**
 * Stop `id` and remove it.
 *
 * @param session - Open session
 * @param id - Container id or name
 * @throws {@link ContainerKitError} when stop or remove fails for a reason other than `notFound`. A container that is already gone is not an error.
 */
export async function stopAndRemove(session: Session, id: string): Promise<void> {
	try {
		await session.stop(id);
	} catch (err) {
		if (!isCKNotFoundError(err)) {
			throw err;
		}
	}
	try {
		await session.remove(id, { force: true });
	} catch (err) {
		if (!isCKNotFoundError(err)) {
			throw err;
		}
	}
}
