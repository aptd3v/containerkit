import { existsSync } from "node:fs";
import { basename, resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { isLiveRoot, runRoot } from "./root.js";

export { exportYAML } from "../render.js";
export {
	ContainerKitError,
	isCKAbortedError,
	isCKAlreadyExistsError,
	isCKCanceledError,
	isCKConflictError,
	isCKDataLossError,
	isCKDisabledError,
	isCKFailedPreconditionError,
	isCKIncompatibleError,
	isCKInternalError,
	isCKInvalidArgumentError,
	isCKInvalidProjectError,
	isCKLoginFailedError,
	isCKLoginRequiredError,
	isCKNotFoundError,
	isCKNotImplementedError,
	isCKNotModifiedError,
	isCKOutOfRangeError,
	isCKParsingFailedError,
	isCKPermissionDeniedError,
	isCKResourceExhaustedError,
	isCKUnauthorizedError,
	isCKUnavailableError,
	isCKUnknownError,
	isCKUnsupportedError,
	isCKUnsupportedFlagError,
	isCKWrongContextTypeError,
} from "./error.js";
export type { ErrorClass, ErrorFrame } from "./error.js";
export { root, useContainerState, useEffect, useMemo, useResource, useSession, useState } from "./root.js";
export type { LiveRoot, RootOptions, Running } from "./root.js";
export type { LiveEvent, Session } from "./session.js";

export async function start(entry = "docker.ts"): Promise<void> {
	const file = resolve(entry);
	const base = basename(file);
	if (base !== "docker.ts" && base !== "docker.tsx") {
		throw new Error("containerkit/live: the entry file must be docker.ts");
	}
	if (!existsSync(file)) {
		throw new Error(`containerkit/live: ${file} is required`);
	}
	const loaded = (await import(pathToFileURL(file).href)) as { default?: unknown };
	if (!isLiveRoot(loaded.default)) {
		throw new Error("containerkit/live: docker.ts must default-export root()");
	}
	await runRoot(loaded.default);
}
