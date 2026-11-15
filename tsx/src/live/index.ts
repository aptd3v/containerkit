import { existsSync } from "node:fs";
import { basename, resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { isLiveRoot, runRoot, type LiveRoot } from "./root.js";

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
export { root, runRoot, useContainerState, useEffect, useMemo, useResource, useSession, useState } from "./root.js";
export type { LiveRoot, OnceRoot, RenderRoot, RootOptions, Running } from "./root.js";
export type { LiveEvent, Session } from "./session.js";
export type {
	ClusterVolume,
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
	CreatedContainer,
	EndpointSettings,
	ExecCreate,
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
	ServiceInspect,
	ServiceList,
	SwarmConfig,
	SwarmConfigSpec,
	SwarmContainerSpec,
	SwarmDriver,
	SwarmEndpoint,
	SwarmInfo,
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
 * `root()` starts on its own only for `docker.ts`, `docker.tsx`, and `docker.js`.
 * Other examples call this with `import.meta.url` so `node dist-examples/<example>.js` opens the session.
 */
export async function runIfMain(live: LiveRoot, url: string): Promise<void> {
	const entry = process.argv[1];
	if (!entry) {
		return;
	}
	if (url !== pathToFileURL(resolve(entry)).href) {
		return;
	}
	await runRoot(live);
}

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
