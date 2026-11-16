import { existsSync } from "node:fs";
import { basename, resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { isLiveRoot, runRoot, type LiveRoot } from "./root.js";
import { isMissingBinaryError } from "./session.js";

export { isMissingBinaryError };
export { ensureImage, hostPlatform, stopAndRemove } from "./helpers.js";

export { exportYAML } from "../render.js";
export {
	ContainerKitError,
	errorClass,
	errorFromFrame,
	isContainerKitError,
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
export type { ErrorClass, ErrorFrame, ErrorReport } from "./error.js";
export { root, runRoot, useContainerState, useEffect, useMemo, useResource, useSession, useState } from "./root.js";
export type { LiveRoot, OnceRoot, RenderRoot, RootOptions, Running } from "./root.js";
export type { ErrorLiveEvent, LiveEvent, LiveEventMap, LiveGroup, LiveStatus, LogLiveEvent, Session, SessionLogEvent, StatusLiveEvent } from "./session.js";
export { eventText } from "./session.js";
export type {
	BuildCachePrune,
	ClusterVolume,
	CommitResult,
	ComposeContainer,
	ComposeContainerState,
	ComposeHealthStatus,
	ComposeImageSummary,
	ComposeStack,
	ContainerConfig,
	ContainerInspect,
	ContainerInspectResult,
	ContainerList,
	ContainerMount,
	ContainerMountPoint,
	ContainerNetworkSettings,
	ContainerPort,
	ContainerPruneResult,
	ContainerRuntimeState,
	ContainerSummary,
	ContainerTop,
	ContainerUpdate,
	CreatedContainer,
	CreatedID,
	EndpointSettings,
	ExecCreate,
	GeneratedProject,
	ImageAttestations,
	HostConfig,
	ImageConfig,
	ImageDelete,
	ImageDescriptor,
	ImageHistoryItem,
	ImageHistoryList,
	ImageIdentity,
	ImageInspect,
	ImageList,
	ImageManifest,
	ImagePruneResult,
	ImageSearchHit,
	ImageSearchList,
	ImageSignature,
	ImageSigner,
	ImageSummary,
	LoadedProject,
	Network,
	NetworkCreate,
	NetworkInspect,
	NetworkInspectResult,
	NetworkList,
	NetworkPruneResult,
	PathStat,
	PathStatResult,
	Ping,
	Volume,
	VolumeCreate,
	VolumeList,
	VolumePruneResult,
	BuildCacheRecord,
	Checkpoint,
	CheckpointList,
	ConfigInspect,
	ConfigList,
	DiskUsage,
	DiskUsageSection,
	EnginePlugin,
	ImageRemoveResult,
	NodeInspect,
	NodeList,
	PluginConfig,
	PluginInspect,
	PluginList,
	PluginMount,
	PluginSettings,
	RegistryLogin,
	SecretInspect,
	SecretList,
	ServiceCreate,
	ServiceInspect,
	ServiceUpdate,
	ServiceList,
	SwarmConfig,
	SwarmConfigSpec,
	SwarmContainerSpec,
	SwarmDriver,
	SwarmEndpoint,
	SwarmInfo,
	SwarmInit,
	SwarmInspect,
	SwarmNode,
	SwarmNodeDescription,
	SwarmNodeSpec,
	SwarmNodeStatus,
	SwarmPort,
	SwarmResources,
	SwarmSecret,
	SwarmSecretSpec,
	SwarmService,
	SwarmServiceSpec,
	SwarmSpec,
	SwarmTask,
	SwarmTaskSpec,
	SwarmTaskState,
	SwarmTaskStatus,
	SwarmUpdateConfig,
	SwarmVersion,
	TaskInspect,
	TaskList,
} from "./results.js";

/**
 * Start `live` when `url` is the file Node is running.
 *
 * @remarks
 * `root()` starts on its own only for `docker.ts`, `docker.tsx`, and `docker.js`.
 * Other examples call this with `import.meta.url` so `node dist-examples/<example>/<example>.js` opens the session.
 * A failed once callback closes the session and exits the process.
 *
 * @param live - Handle returned by {@link root}
 * @param url - `import.meta.url` of the example file
 *
 * @example
 * ```ts
 * await runIfMain(live, import.meta.url);
 * ```
 */
export async function runIfMain(live: LiveRoot, url: string): Promise<void> {
	const entry = process.argv[1];
	if (!entry) {
		return;
	}
	if (url !== pathToFileURL(resolve(entry)).href) {
		return;
	}
	const running = await runRoot(live);
	if (live.failed) {
		await running.close();
		process.exit(1);
	}
}

/**
 * Load `entry` and run its default export.
 *
 * @param entry - Path to `docker.ts` or `docker.tsx`. The default is `docker.ts` in the working directory.
 * @throws Error when the file is not named `docker.ts` or `docker.tsx`, is missing, or does not default-export {@link root}
 */
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
