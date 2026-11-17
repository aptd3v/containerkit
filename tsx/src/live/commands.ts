import * as generated from "../generated-actions.js";
import { buildDockerfile } from "../index.js";
import type { Client, PlanDocument } from "../plan.js";
import { Flow } from "../taskflow.js";
import type {
	ComposeContainer,
	ComposeImageSummary,
	ComposeStack,
	ContainerInspectResult,
	ContainerList,
	ContainerPruneResult,
	ContainerTop,
	CreatedContainer,
	ExecCreate,
	ImageHistoryList,
	ImageInspect,
	ImageList,
	ImagePruneResult,
	ImageSearchList,
	LoadedProject,
	NetworkCreate,
	NetworkInspectResult,
	NetworkList,
	NetworkPruneResult,
	PathStatResult,
	Ping,
	Volume,
	VolumeCreate,
	VolumeList,
	VolumePruneResult,
	RegistryLogin,
	ImageRemoveResult,
	DiskUsage,
	CommitResult,
	CreatedID,
	ServiceCreate,
	SwarmInit,
	ContainerUpdate,
	ServiceUpdate,
	BuildCachePrune,
	ImageAttestations,
	GeneratedProject,
	CheckpointList,
	ConfigList,
	ConfigInspect,
	SecretList,
	SecretInspect,
	PluginList,
	PluginInspect,
	NodeList,
	NodeInspect,
	ServiceList,
	ServiceInspect,
	TaskList,
	TaskInspect,
	SwarmInspect,
} from "./results.js";

type Command = (cli: Client, ...args: never[]) => PlanDocument;

export const commandByName: Record<string, Command> = {};

for (const [name, value] of Object.entries(generated)) {
	if (typeof value === "function") {
		commandByName[name] = value as Command;
	}
}
commandByName.buildDockerfile = buildDockerfile;
commandByName.flow = Flow as Command;

type SessionResult = {
	listImages: ImageList;
	inspectImage: ImageInspect;
	composePs: ComposeContainer[] | null;
	composeExec: number;
	composeRun: number;
	composeWait: number;
	composePort: [string, number];
	composeViz: string;
	load: LoadedProject;
	create: CreatedContainer;
	run: CreatedContainer;
	containerRun: CreatedContainer;
	createExec: ExecCreate;
	createNetwork: NetworkCreate;
	ping: Ping;
	listContainers: ContainerList;
	inspect: ContainerInspectResult;
	listNetworks: NetworkList;
	inspectNetwork: NetworkInspectResult;
	listEngineVolumes: VolumeList;
	createVolume: VolumeCreate;
	searchImages: ImageSearchList;
	historyImage: ImageHistoryList;
	topContainer: ContainerTop;
	statPath: PathStatResult;
	composeList: ComposeStack[] | null;
	composeImages: Record<string, ComposeImageSummary> | null;
	composeVolumeList: Volume[] | null;
	pruneImages: ImagePruneResult;
	pruneContainers: ContainerPruneResult;
	pruneNetworks: NetworkPruneResult;
	pruneVolumes: VolumePruneResult;
	login: RegistryLogin;
	diskUsage: DiskUsage;
	removeImage: ImageRemoveResult;
	listCheckpoints: CheckpointList;
	listConfigs: ConfigList;
	inspectConfig: ConfigInspect;
	listSecrets: SecretList;
	inspectSecret: SecretInspect;
	listPlugins: PluginList;
	inspectPlugin: PluginInspect;
	listNodes: NodeList;
	inspectNode: NodeInspect;
	listSwarmServices: ServiceList;
	inspectSwarmService: ServiceInspect;
	listTasks: TaskList;
	inspectTask: TaskInspect;
	inspectSwarm: SwarmInspect;
	commitContainer: CommitResult;
	createEngineConfig: CreatedID;
	createEngineSecret: CreatedID;
	createSwarmService: ServiceCreate;
	initSwarm: SwarmInit;
	updateContainer: ContainerUpdate;
	updateSwarmService: ServiceUpdate;
	pruneBuildCache: BuildCachePrune;
	attestImage: ImageAttestations;
	composeGenerate: GeneratedProject;
	composeLoadProject: GeneratedProject;
};

/** Actions whose JSON result is only `{"ok":true}`. Progress arrives as session events. */
type SessionVoid =
	| "buildImage"
	| "composeAttach"
	| "composeBuild"
	| "composeCommit"
	| "composeCopy"
	| "composeCreate"
	| "composeDown"
	| "composeEvents"
	| "composeExport"
	| "composeKill"
	| "composePause"
	| "composePublishOCI"
	| "composePull"
	| "composePush"
	| "composeRemove"
	| "composeRestart"
	| "composeScale"
	| "composeStart"
	| "composeStop"
	| "composeUnpause"
	| "composeUp"
	| "composeWatch"
	| "connectNetwork"
	| "createCheckpoint"
	| "createPlugin"
	| "disablePlugin"
	| "disconnectNetwork"
	| "enablePlugin"
	| "joinSwarm"
	| "kill"
	| "leaveSwarm"
	| "logs"
	| "pullImage"
	| "pushImage"
	| "remove"
	| "removeCheckpoint"
	| "removeNode"
	| "removePlugin"
	| "removeVolume"
	| "renameContainer"
	| "resizeContainer"
	| "resizeExec"
	| "restart"
	| "setPlugin"
	| "copyTo"
	| "start"
	| "startExec"
	| "stop"
	| "tagImage"
	| "unlockSwarm"
	| "updateEngineConfig"
	| "updateEngineSecret"
	| "updateNode"
	| "updateSwarm"
	| "updateVolume";

type DropClient<F, R = unknown> = F extends (cli: Client, ...args: infer A) => PlanDocument
	? (...args: A) => Promise<R>
	: never;

type SessionReturn<K> = K extends keyof SessionResult ? SessionResult[K] : K extends SessionVoid ? void : unknown;

export type SessionCommands = {
	[K in keyof typeof generated]: DropClient<(typeof generated)[K], SessionReturn<K>>;
} & {
	buildDockerfile: DropClient<typeof buildDockerfile, void>;
	flow: DropClient<typeof Flow, void>;
};
