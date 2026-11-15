/**
 * JSON results from a live session.
 *
 * The objects are the Engine client and Compose API structs pinned in `go.mod`,
 * passed through `encoding/json`. A nil slice or map is `null`. A `time.Duration`
 * is an integer number of nanoseconds. Sizes are bytes. `-1` on a size or a count
 * means the daemon did not calculate that value.
 */

/**
 * Result of `listImages()`. `Items` is null when the Engine list is nil.
 * `SharedSize`, `Manifests`, and `Identity` stay unset unless the matching list option asked for them.
 */
export type ImageList = {
	/** Images in the local store. */
	Items: ImageSummary[] | null;
};

/**
 * One image from `listImages()`.
 * `Id` is the config digest. `RepoDigests` are the manifest digests that point at it.
 */
export type ImageSummary = {
	/** Containers using this image, running or stopped. `-1` when the count was not calculated. */
	Containers: number;
	/** Creation time as Unix seconds. */
	Created: number;
	/** Content-addressable image id, `sha256:…`, from the image config. */
	Id: string;
	/** User-defined metadata. Null when the image has no labels. */
	Labels: Record<string, string> | null;
	/**
	 * Parent image id. Set for images built locally. Empty when the image was pulled.
	 */
	ParentId: string;
	/** Manifest digests that reference this image. Usually present after a pull or a push. Null when none are known. */
	RepoDigests: string[] | null;
	/** `name:tag` references in the local cache. Null when the image is untagged; the id still addresses it. */
	RepoTags: string[] | null;
	/** Bytes of layers shared with other images. `-1` when shared size was not requested. */
	SharedSize: number;
	/** Total size of the image layers, in bytes. */
	Size: number;
	/** OCI descriptor of the image target. Present on a multi-platform image store. Experimental. */
	Descriptor?: ImageDescriptor;
	/** Platform manifests and attached data such as attestations. Present when the list asked for manifests. Experimental. */
	Manifests?: ImageManifest[];
};

/**
 * Result of `inspectImage()`.
 * `Created` here is an RFC 3339 string. The same field on {@link ImageSummary} is Unix seconds.
 */
export type ImageInspect = {
	/** Content-addressable image id, from the image config. Distinct from {@link ImageInspect.RepoDigests}. */
	Id: string;
	/** `name:tag` references in the local cache. Null when the image is untagged. */
	RepoTags: string[] | null;
	/** Manifest digests that reference this image. Null when the image has not been pulled or pushed. */
	RepoDigests: string[] | null;
	/** Message stored by commit or import. Omitted when unset. */
	Comment?: string;
	/** Creation time, RFC 3339 with nanoseconds. Omitted when the image has no created time. */
	Created?: string;
	/** Author from commit, or the deprecated Dockerfile `MAINTAINER`. Omitted when unset. */
	Author?: string;
	/** Execution config: user, env, command, healthcheck. Null when the image has no config. */
	Config: ImageConfig | null;
	/** CPU architecture the image runs on, such as `amd64` or `arm64`. */
	Architecture: string;
	/** CPU variant. Set for ARM images, such as `v7`. */
	Variant?: string;
	/** Operating system the image runs on, such as `linux`. */
	Os: string;
	/** Operating-system version. Set for Windows images. */
	OsVersion?: string;
	/** Size of the selected image variant, in bytes, including locally unpacked snapshot data on the containerd image store. */
	Size: number;
	/** Storage driver that holds the image filesystem. `Data` is driver-specific. */
	GraphDriver?: { Name: string; Data: Record<string, string> | null };
	/** Root filesystem type and layer ids, bottom-most first. */
	RootFS: { Type?: string; Layers?: string[] | null };
	/** Daemon-local metadata. `LastTagTime` is RFC 3339 and is not part of the image itself. */
	Metadata: { LastTagTime?: string };
	/** OCI descriptor of the image target. Present on a multi-platform image store. Experimental. */
	Descriptor?: ImageDescriptor;
	/**
	 * Platform manifests and attestations.
	 * Present when inspect asked for manifests, the daemon has a multi-platform store, and no single platform was selected.
	 * Experimental.
	 */
	Manifests?: ImageManifest[];
	/**
	 * Verified origin of the image. Tagging the image to another name does not change this.
	 * Omitted unless the daemon recorded identity for the image.
	 */
	Identity?: ImageIdentity;
};

/** Image config used as the base for a container. Ports and volumes are sets: each value is an empty object. */
export type ImageConfig = {
	/** User the container process runs as. `user` or `uid:gid`. */
	User?: string;
	/** Ports exposed by the image. Keys are `port/protocol`, such as `8080/tcp`. */
	ExposedPorts?: Record<string, object> | null;
	/** Environment entries, each `KEY=value`. */
	Env?: string[] | null;
	/** Exec-form entrypoint. Null when the image sets none. */
	Entrypoint?: string[] | null;
	/** Default arguments to the entrypoint. */
	Cmd?: string[] | null;
	/** Paths the image expects a container to write. Keys are container paths. */
	Volumes?: Record<string, object> | null;
	/** Working directory of the entrypoint process. */
	WorkingDir?: string;
	/** Arbitrary metadata stored in the image config. */
	Labels?: Record<string, string> | null;
	/** Signal sent to stop a container started from this image, such as `SIGTERM`. */
	StopSignal?: string;
	/**
	 * Windows legacy flag. When true, `Entrypoint` or `Cmd` is one already-escaped command line.
	 */
	ArgsEscaped?: boolean;
	/** Healthcheck from the image. Durations are nanoseconds. A zero duration means inherit. */
	Healthcheck?: ImageHealthcheck;
	/** `ONBUILD` instructions recorded on the image. */
	OnBuild?: string[] | null;
	/** Shell for shell-form `RUN`, `CMD`, and `ENTRYPOINT`. */
	Shell?: string[] | null;
};

/**
 * Healthcheck stored on an image or a container config.
 * Durations are nanoseconds. Zero means inherit the base value.
 */
export type ImageHealthcheck = {
	/**
	 * Probe to run.
	 * `[]` inherits. `["NONE"]` disables the check. `["CMD", ...args]` is exec form.
	 * `["CMD-SHELL", command]` runs `command` with the default shell.
	 */
	Test?: string[] | null;
	/** Wait between checks, in nanoseconds. */
	Interval?: number;
	/** How long a check may run before it is treated as hung, in nanoseconds. */
	Timeout?: number;
	/** Startup grace period, in nanoseconds. Failures during this period do not count toward {@link ImageHealthcheck.Retries}. */
	StartPeriod?: number;
	/** Time between checks during the start period, in nanoseconds. */
	StartInterval?: number;
	/** Consecutive failures required before the container is unhealthy. Zero means inherit. */
	Retries?: number;
};

/**
 * OCI descriptor (`application/vnd.oci.descriptor.v1+json`).
 * Field names are the OCI names, lowercase, unlike the Engine structs around them.
 */
export type ImageDescriptor = {
	/** Media type of the referenced object. */
	mediaType: string;
	/** Digest of the referenced content, such as `sha256:…`. */
	digest: string;
	/** Size of the blob, in bytes. */
	size: number;
	/** URLs the blob may be downloaded from. */
	urls?: string[];
	/** Arbitrary metadata for the referenced content. */
	annotations?: Record<string, string>;
	/** Embedded content, base64 in JSON, so the blob does not have to be fetched. */
	data?: string;
	/** Platform of a manifest this descriptor points at. */
	platform?: ImagePlatform;
	/** IANA media type of the artifact. */
	artifactType?: string;
};

/** OCI platform the image runs on. */
export type ImagePlatform = {
	/** CPU architecture, such as `amd64` or `arm64`. */
	architecture: string;
	/** Operating system, such as `linux` or `windows`. */
	os: string;
	/** Operating-system version, such as a Windows build number. */
	"os.version"?: string;
	/** Required OS features, such as `win32k` on Windows. */
	"os.features"?: string[];
	/** CPU variant, such as `v7` when `architecture` is `arm`. */
	variant?: string;
};

/**
 * One manifest inside an image index: a platform image, an attestation, or an unrecognized object.
 * Sizes count only content that is present on this daemon.
 */
export type ImageManifest = {
	/** Digest of this manifest. Same value as the manifest's own content id. */
	ID: string;
	/** OCI descriptor of this manifest. */
	Descriptor: ImageDescriptor;
	/** True when the config and layers this manifest references are fully present locally. */
	Available: boolean;
	/**
	 * Local sizes in bytes.
	 * `Content` is blobs in the content store. `Total` adds the kind-specific data, such as the unpacked snapshot.
	 */
	Size: { Content: number; Total: number };
	/** `image` for a runnable platform image, `attestation` for a provenance/SBOM blob, `unknown` otherwise. */
	Kind: "image" | "attestation" | "unknown";
	/** Set when {@link ImageManifest.Kind} is `image`. */
	ImageData?: {
		/** Platform this image manifest runs on. */
		Platform: ImagePlatform;
		/** Verified origin recorded for this manifest. */
		Identity?: ImageIdentity;
		/** Unpacked snapshot size in bytes. This can remain after the compressed blobs are gone. */
		Size: { Unpacked: number };
		/** Ids of containers using this platform image. */
		Containers: string[] | null;
	};
	/** Set when {@link ImageManifest.Kind} is `attestation`. `For` is the digest of the image manifest it describes. */
	AttestationData?: { For: string };
};

/**
 * Trusted origin of an image, verified by the daemon.
 * Retagging does not rewrite it.
 */
export type ImageIdentity = {
	/** Verified signatures. An entry with `Error` set failed verification; its other fields are empty. */
	Signature?: ImageSignature[];
	/** Registries the image was pulled from. A mirror pull records the original repository. A later push adds that repository too. */
	Pull?: { Repository?: string }[];
	/** Builds that produced the image. `Ref` looks the build up in BuildKit history. `CreatedAt` is RFC 3339. */
	Build?: { Ref?: string; CreatedAt?: string }[];
};

/** One verified signature on an image. */
export type ImageSignature = {
	/** Short description of the signature type. */
	Name?: string;
	/** Verified timestamps. `Type` is `Tlog` or `TimestampAuthority`. `Timestamp` is RFC 3339. */
	Timestamps?: { Type: string; URI: string; Timestamp: string }[];
	/** Signer the daemon recognizes specially. `DHI` is Docker Hardened Images. */
	KnownSigner?: string;
	/** Image reference stored on older hashed-record signatures. */
	DockerReference?: string;
	/** Certificate that signed the image. */
	Signer?: ImageSigner;
	/** Signature format, such as `bundle-v0.3` or `simplesigning-v1`. */
	SignatureType?: string;
	/** Verification failure. When this is set, the other fields on this signature are empty. */
	Error?: string;
	/** Verification warnings, such as falling back to cached trust roots. The signature can still be valid. */
	Warnings?: string[];
};

/**
 * Fulcio-style certificate summary for the signer.
 * The URI and digest fields describe the build that produced the signature.
 */
export type ImageSigner = {
	/** Issuer of the signing certificate. */
	CertificateIssuer: string;
	/** Subject alternative name on the signing certificate. */
	SubjectAlternativeName: string;
	/** OIDC issuer. Matches the `iss` claim, or the upstream issuer for a federated login. */
	Issuer?: string;
	/** URI of the build instructions that signed the image. */
	BuildSignerURI?: string;
	/** Immutable digest of those build instructions. */
	BuildSignerDigest?: string;
	/** Where the build ran: platform-hosted cloud, or customer-hosted infrastructure. */
	RunnerEnvironment?: string;
	/** Source repository URL the build used. */
	SourceRepositoryURI?: string;
	/** Immutable digest of the source the build used. */
	SourceRepositoryDigest?: string;
	/** Git ref the build ran from. */
	SourceRepositoryRef?: string;
	/** Immutable id of the source repository. */
	SourceRepositoryIdentifier?: string;
	/** URL of the owner of that source repository. */
	SourceRepositoryOwnerURI?: string;
	/** Immutable id of that owner. */
	SourceRepositoryOwnerIdentifier?: string;
	/** URL of the top-level build instructions. */
	BuildConfigURI?: string;
	/** Immutable digest of the top-level build instructions. */
	BuildConfigDigest?: string;
	/** Event that started the build. */
	BuildTrigger?: string;
	/** URL that identifies this build execution. */
	RunInvocationURI?: string;
	/** Repository visibility at the time the certificate was signed. */
	SourceRepositoryVisibilityAtSigning?: string;
};

/** Container lifecycle state used by Compose `ps` and by Engine container list and inspect. */
export type ComposeContainerState =
	| "created"
	| "running"
	| "paused"
	| "restarting"
	| "removing"
	| "exited"
	| "dead";

/**
 * Health of a container.
 * `""` means this response carried no health status. `none` means the container has no healthcheck.
 * `starting` is the grace period. `healthy` and `unhealthy` are probe results.
 */
export type ComposeHealthStatus = "" | "none" | "starting" | "healthy" | "unhealthy";

/**
 * One container from `composePs()`.
 * `State` is the lifecycle enum. `Status` is the human text, such as `Up 5 seconds`.
 */
export type ComposeContainer = {
	/** Container id. */
	ID: string;
	/** Compose container name, without a leading slash. */
	Name: string;
	/** Engine names for the container, including link aliases. Null when the daemon returned none. */
	Names: string[] | null;
	/** Image reference the container was created from. */
	Image: string;
	/** Command line shown by `ps`. */
	Command: string;
	/** Compose project name. */
	Project: string;
	/** Compose service name. */
	Service: string;
	/** Creation time as Unix seconds. */
	Created: number;
	/** Lifecycle state. */
	State: ComposeContainerState;
	/** Human-readable status, such as `Up 5 seconds` or `Exited (0) 2 hours ago`. */
	Status: string;
	/** Health status. Empty when the container is not running or has no recorded probe. */
	Health: ComposeHealthStatus;
	/** Exit code. Set when the container has exited or is dead. */
	ExitCode: number;
	/**
	 * Published ports.
	 * `URL` is the host address, `TargetPort` is the container port, `PublishedPort` is the host port.
	 */
	Publishers: { URL: string; TargetPort: number; PublishedPort: number; Protocol: string }[] | null;
	/** Container labels, including the Compose project and service labels. */
	Labels: Record<string, string> | null;
	/** Writable-layer size in bytes. Omitted unless the list asked for sizes. */
	SizeRw?: number;
	/** Total root-filesystem size in bytes. Omitted unless the list asked for sizes. */
	SizeRootFs?: number;
	/** Mount names. A nameless mount is reported by its source path. */
	Mounts: string[] | null;
	/** Networks the container is attached to. */
	Networks: string[] | null;
	/** How many of {@link ComposeContainer.Mounts} use the local volume driver. */
	LocalVolumes: number;
};

/** Result of `create()`. Warnings are daemon notices. The container still exists when this returns. */
export type CreatedContainer = {
	/** Id of the created container. */
	ID: string;
	/** Daemon warnings from create. Null when there were none. */
	Warnings: string[] | null;
};

/** Result of `createExec()`. */
export type ExecCreate = {
	/** Id to pass to later exec calls for this process. */
	ID: string;
};

/** Result of `createNetwork()`. The warning list field is `Warning`, singular. */
export type NetworkCreate = {
	/** Id of the created network. */
	ID: string;
	/** Daemon warnings from create. Null when there were none. */
	Warning: string[] | null;
};

/**
 * Result of `ping()`.
 * Header values from the daemon: API version, OS, experimental flag, and builder.
 */
export type Ping = {
	/** Engine API version the daemon speaks, such as `1.56`. */
	APIVersion: string;
	/** Daemon OS, `linux` or `windows`. */
	OSType: string;
	/** True when the daemon is running with experimental features enabled. */
	Experimental: boolean;
	/** Builder implementation. `2` is BuildKit. */
	BuilderVersion: string;
	/**
	 * Swarm membership from the ping response.
	 * Null when this API version sent no Swarm header.
	 * `NodeState` is `inactive`, `pending`, `active`, `error`, or `locked`.
	 * `ControlAvailable` is true on a manager.
	 */
	SwarmStatus: { NodeState: string; ControlAvailable: boolean } | null;
};

/** Result of `load()`. The loaded project's name. Loading does not start containers. */
export type LoadedProject = {
	/** Compose project name read from the file. Empty when the file set none. */
	name: string;
};

/** A published port on a container from `listContainers()`. */
export type ContainerPort = {
	/** Host address the port is bound to. Omitted when the mapping has no specific host IP. */
	IP?: string;
	/** Port inside the container. */
	PrivatePort: number;
	/** Port on the host. Omitted when the container port is exposed and not published. */
	PublicPort?: number;
	/** Protocol: `tcp`, `udp`, or `sctp`. */
	Type: string;
};

/** A mount reported on a container. This describes a mount in use, which is separate from the requested {@link ContainerMount}. */
export type ContainerMountPoint = {
	/** Mount type: `bind`, `volume`, `tmpfs`, `npipe`, `image`, or `cluster`. */
	Type?: string;
	/** Volume name, when the mount is a volume. */
	Name?: string;
	/** Host path, or the volume's storage path under the daemon's volume directory. Empty for tmpfs. */
	Source: string;
	/** Path inside the container. */
	Destination: string;
	/** Volume driver that created the volume. */
	Driver?: string;
	/** Options supplied when the mount was created, comma-separated. */
	Mode: string;
	/** True when the mount is writable. */
	RW: boolean;
	/** Linux mount propagation, such as `rprivate` or `shared`. Unused on Windows. */
	Propagation: string;
};

/** One container from `listContainers()`. */
export type ContainerSummary = {
	/** Container id. */
	Id: string;
	/** Names, each with a leading slash. The first canonical name has no extra slash. Null when the daemon returned none. */
	Names: string[] | null;
	/** Image reference the container was created from. */
	Image: string;
	/** Image id, `sha256:…`. */
	ImageID: string;
	/** Descriptor of the platform manifest used to create the container. */
	ImageManifestDescriptor?: ImageDescriptor;
	/** Command line shown by `ps`. */
	Command: string;
	/** Creation time as Unix seconds. */
	Created: number;
	/** Published and exposed ports. */
	Ports: ContainerPort[] | null;
	/** Writable-layer size in bytes. Omitted unless the list asked for sizes. */
	SizeRw?: number;
	/** Total root-filesystem size in bytes. Omitted unless the list asked for sizes. */
	SizeRootFs?: number;
	/** Container labels. */
	Labels: Record<string, string> | null;
	/** Lifecycle state. */
	State: ComposeContainerState;
	/** Human-readable status, such as `Up 5 seconds`. */
	Status: string;
	/** Network mode and runtime annotations copied onto the summary. */
	HostConfig: {
		/** Network mode, such as `bridge`, `host`, `none`, or `container:<id>`. */
		NetworkMode?: string;
		/** Runtime annotations. Omitted when the container has none. */
		Annotations?: Record<string, string> | null;
	};
	/** Current health. Omitted when the container has no healthcheck state. */
	Health?: { Status: ComposeHealthStatus; FailingStreak: number };
	/** Networks the container is attached to, keyed by network name. */
	NetworkSettings: { Networks: Record<string, EndpointSettings | null> | null } | null;
	/** Mounts in use by the container. */
	Mounts: ContainerMountPoint[] | null;
};

/** Result of `listContainers()`. `Items` is null when the Engine list is nil. */
export type ContainerList = {
	/** Containers matching the list filters. */
	Items: ContainerSummary[] | null;
};

/**
 * Full container record inside {@link ContainerInspectResult}.
 * `Created` is RFC 3339. The same field on {@link ContainerSummary} is Unix seconds.
 * `SizeRw` and `SizeRootFs` are set only when inspect was asked for sizes.
 */
export type ContainerInspect = {
	/** Container id. */
	Id: string;
	/** Creation time, RFC 3339 with nanoseconds. */
	Created: string;
	/** Executable path of the container's main process. */
	Path: string;
	/** Arguments to {@link ContainerInspect.Path}. */
	Args: string[] | null;
	/** Running state, including the health log. Null when the daemon sent no state. */
	State: ContainerRuntimeState | null;
	/** Image id the container was created from. */
	Image: string;
	/** Host path of the container's `resolv.conf`. */
	ResolvConfPath: string;
	/** Host path of the container's hostname file. */
	HostnamePath: string;
	/** Host path of the container's `hosts` file. */
	HostsPath: string;
	/** Host path of the container's log file. */
	LogPath: string;
	/** Container name, with a leading slash. */
	Name: string;
	/** How many times the daemon has restarted this container. */
	RestartCount: number;
	/** Storage driver holding the container filesystem, such as `overlay2`. */
	Driver: string;
	/** Runtime platform, such as `linux`. */
	Platform: string;
	/** SELinux mount label. */
	MountLabel: string;
	/** SELinux process label. */
	ProcessLabel: string;
	/** AppArmor profile name. Empty when AppArmor is not applied. */
	AppArmorProfile: string;
	/** Ids of exec instances created in this container. */
	ExecIDs: string[] | null;
	/** Host-dependent settings: binds, resources, restart policy, namespaces. */
	HostConfig: HostConfig | null;
	/** Storage driver metadata. `Data` is driver-specific and informational. */
	GraphDriver?: { Name: string; Data: Record<string, string> | null };
	/** Snapshotter used for the root filesystem, when the daemon reports one. */
	Storage?: { RootFS?: { Snapshot?: { Name?: string } } };
	/** Writable-layer size in bytes. Set when inspect was asked for sizes. */
	SizeRw?: number;
	/** Total root-filesystem size in bytes. Set when inspect was asked for sizes. */
	SizeRootFs?: number;
	/** Mounts in use. */
	Mounts: ContainerMountPoint[] | null;
	/** Portable config: command, env, image name, labels. */
	Config: ContainerConfig | null;
	/** Network namespace and per-network endpoints. */
	NetworkSettings: ContainerNetworkSettings | null;
	/** Descriptor of the platform manifest used to create the container. */
	ImageManifestDescriptor?: ImageDescriptor;
};

/**
 * Result of `inspect()`.
 * `Container` is the decoded inspect document. `Raw` is that same document before decoding, or null.
 */
export type ContainerInspectResult = {
	/** Decoded inspect document. */
	Container: ContainerInspect;
	/** Inspect JSON as the daemon sent it, decoded again into the same shape. Null when the client kept no raw body. */
	Raw: ContainerInspect | null;
};

/** Runtime state from container inspect. */
export type ContainerRuntimeState = {
	/** Lifecycle state. */
	Status: ComposeContainerState;
	/** True while the main process is running. */
	Running: boolean;
	/** True while the container is paused. */
	Paused: boolean;
	/** True while a restart is in progress. */
	Restarting: boolean;
	/** True when the kernel OOM killer stopped the container. */
	OOMKilled: boolean;
	/** True when the daemon has marked the container dead. */
	Dead: boolean;
	/** Host pid of the main process. `0` when the container is not running. */
	Pid: number;
	/** Exit code of the main process. */
	ExitCode: number;
	/** Error recorded against the container. Empty when the last transition had none. */
	Error: string;
	/** RFC 3339 time the container last started. */
	StartedAt: string;
	/** RFC 3339 time the container last finished. The zero time is `0001-01-01T00:00:00Z` when it has not finished. */
	FinishedAt: string;
	/** Health probes. Omitted when the container has no healthcheck. The log is oldest first. */
	Health?: {
		/** Current health. */
		Status: ComposeHealthStatus;
		/** Consecutive failed probes. */
		FailingStreak: number;
		/**
		 * Recent probe results.
		 * `ExitCode` is `0` healthy, `1` unhealthy, `2` reserved and treated as unhealthy. Any other code means the probe itself failed.
		 */
		Log: { Start: string; End: string; ExitCode: number; Output: string }[] | null;
	};
};

/**
 * Portable container config from inspect.
 * Host-dependent settings are on {@link HostConfig}.
 */
export type ContainerConfig = {
	/** Hostname inside the container. */
	Hostname: string;
	/** DNS domain name inside the container. */
	Domainname: string;
	/** User the command runs as. `user` or `user:group`. */
	User: string;
	/** True when stdin was attached at create time. */
	AttachStdin: boolean;
	/** True when stdout was attached at create time. */
	AttachStdout: boolean;
	/** True when stderr was attached at create time. */
	AttachStderr: boolean;
	/** Exposed ports. Keys are `port/protocol`. */
	ExposedPorts?: Record<string, object> | null;
	/** True when the container has a terminal. */
	Tty: boolean;
	/** True when stdin is held open. */
	OpenStdin: boolean;
	/** True when stdin closes after the first attached client disconnects. */
	StdinOnce: boolean;
	/** Environment entries, each `KEY=value`. */
	Env: string[] | null;
	/** Command, exec form. */
	Cmd: string[] | null;
	/** Healthcheck. Durations are nanoseconds. */
	Healthcheck?: ImageHealthcheck;
	/** Windows legacy flag. When true, the command is one already-escaped command line. */
	ArgsEscaped?: boolean;
	/** Image name as it was given at create time. This can be a tag rather than an id. */
	Image: string;
	/** Declared volumes. Keys are container paths. */
	Volumes: Record<string, object> | null;
	/** Working directory of the command. */
	WorkingDir: string;
	/** Entrypoint, exec form. */
	Entrypoint: string[] | null;
	/** True when the container was created with networking disabled. */
	NetworkDisabled?: boolean;
	/** `ONBUILD` metadata copied from the image. */
	OnBuild?: string[] | null;
	/** Container labels. */
	Labels: Record<string, string> | null;
	/** Stop signal, such as `SIGTERM`. */
	StopSignal?: string;
	/** Seconds to wait after the stop signal before the daemon kills the container. */
	StopTimeout?: number;
	/** Shell for shell-form commands. */
	Shell?: string[] | null;
};

/**
 * Host-dependent container settings from inspect.
 * Resource numbers follow the Engine: memory is bytes, `NanoCpus` is billionths of a CPU,
 * and the CFS period and quota are microseconds.
 */
export type HostConfig = {
	/** Binds, each `host:container[:options]`. */
	Binds: string[] | null;
	/** Host file the container id is written to. */
	ContainerIDFile: string;
	/** Log driver and its options. `Type` is the driver name, such as `json-file`. */
	LogConfig: { Type: string; Config: Record<string, string> | null };
	/** Network mode: `bridge`, `host`, `none`, a network name, or `container:<id>`. */
	NetworkMode: string;
	/** Port bindings keyed by `port/protocol`. Each entry is a host IP and host port. */
	PortBindings: Record<string, { HostIp: string; HostPort: string }[] | null> | null;
	/**
	 * Restart policy.
	 * `Name` is `no`, `always`, `on-failure`, or `unless-stopped`.
	 * `MaximumRetryCount` applies to `on-failure`.
	 */
	RestartPolicy: { Name: string; MaximumRetryCount: number };
	/** True when the container is removed after it exits. */
	AutoRemove: boolean;
	/** Volume driver used for anonymous volumes. */
	VolumeDriver: string;
	/** Containers to inherit volumes from, each `name[:ro|rw]`. */
	VolumesFrom: string[] | null;
	/** Initial console size, `[height, width]`. */
	ConsoleSize: [number, number];
	/** Runtime annotations. Omitted when unset. */
	Annotations?: Record<string, string> | null;
	/** Kernel capabilities added on top of the default set. */
	CapAdd: string[] | null;
	/** Kernel capabilities removed from the default set. */
	CapDrop: string[] | null;
	/** Cgroup namespace: `private`, `host`, or empty for the daemon default. */
	CgroupnsMode: string;
	/** DNS server addresses. */
	Dns: string[] | null;
	/** Resolver options, such as `ndots:1`. */
	DnsOptions: string[] | null;
	/** DNS search domains. */
	DnsSearch: string[] | null;
	/** Extra hosts, each `hostname:ip`. */
	ExtraHosts: string[] | null;
	/** Supplementary group ids. */
	GroupAdd: string[] | null;
	/** IPC namespace: `private`, `host`, `shareable`, `none`, or `container:<id>`. */
	IpcMode: string;
	/** Cgroup to join. `container:<id>` uses another container's cgroup. */
	Cgroup: string;
	/** Legacy links, each `name:alias`. */
	Links: string[] | null;
	/** OOM-killer preference. Higher is more likely to be killed. */
	OomScoreAdj: number;
	/** PID namespace: empty for private, `host`, or `container:<id>`. */
	PidMode: string;
	/** True when the container is privileged. */
	Privileged: boolean;
	/** True when every exposed port is published on an ephemeral host port. */
	PublishAllPorts: boolean;
	/** True when the root filesystem is read-only. */
	ReadonlyRootfs: boolean;
	/** Security options, such as SELinux labels. */
	SecurityOpt: string[] | null;
	/** Storage-driver options for this container. */
	StorageOpt?: Record<string, string> | null;
	/** tmpfs mounts, keyed by the container path. The value is the mount option string. */
	Tmpfs?: Record<string, string> | null;
	/** UTS namespace: empty for private, or `host`. */
	UTSMode: string;
	/** User namespace: empty for private, or `host`. */
	UsernsMode: string;
	/** Shared memory size, in bytes. */
	ShmSize: number;
	/** Namespaced sysctls, such as `net.ipv4.ip_forward`. */
	Sysctls?: Record<string, string> | null;
	/** OCI runtime name. Empty uses the daemon default. */
	Runtime?: string;
	/** Initial process umask. Omitted when unset. */
	Umask?: number;
	/** Windows isolation: empty, `default`, `process`, or `hyperv`. */
	Isolation: string;
	/** Relative CPU weight against other containers. */
	CpuShares: number;
	/** Memory limit in bytes. `0` means no limit. */
	Memory: number;
	/** CPU quota in billionths of a CPU. `1000000000` is one CPU. */
	NanoCpus: number;
	/** Parent cgroup. */
	CgroupParent: string;
	/** Block IO weight relative to other containers. */
	BlkioWeight: number;
	/** Per-device block IO weights. `Weight` is relative. */
	BlkioWeightDevice: { Path: string; Weight: number }[] | null;
	/** Per-device read limit in bytes per second. */
	BlkioDeviceReadBps: { Path: string; Rate: number }[] | null;
	/** Per-device write limit in bytes per second. */
	BlkioDeviceWriteBps: { Path: string; Rate: number }[] | null;
	/** Per-device read limit in I/O operations per second. */
	BlkioDeviceReadIOps: { Path: string; Rate: number }[] | null;
	/** Per-device write limit in I/O operations per second. */
	BlkioDeviceWriteIOps: { Path: string; Rate: number }[] | null;
	/** CFS scheduler period, in microseconds. */
	CpuPeriod: number;
	/** CFS scheduler quota, in microseconds, within each period. */
	CpuQuota: number;
	/** Real-time scheduler period, in microseconds. */
	CpuRealtimePeriod: number;
	/** Real-time scheduler runtime, in microseconds, within each period. */
	CpuRealtimeRuntime: number;
	/** CPUs the container may run on, such as `0-2` or `0,1`. */
	CpusetCpus: string;
	/** Memory nodes the container may use, such as `0-2` or `0,1`. */
	CpusetMems: string;
	/** Devices mapped from the host. `CgroupPermissions` is a string such as `rwm`. */
	Devices: { PathOnHost: string; PathInContainer: string; CgroupPermissions: string }[] | null;
	/** Rules added to the device cgroup. */
	DeviceCgroupRules: string[] | null;
	/**
	 * Device-driver requests, used for GPUs.
	 * `Count` of `-1` asks for every device. `Capabilities` is an OR of AND-lists, such as `[["gpu"]]`.
	 */
	DeviceRequests: {
		/** Device driver name. */
		Driver: string;
		/** How many devices to request. `-1` is all of them. */
		Count: number;
		/** Device ids the driver recognizes. */
		DeviceIDs: string[] | null;
		/** OR of AND-lists of capabilities. */
		Capabilities: string[][] | null;
		/** Options passed to the driver. */
		Options: Record<string, string> | null;
	}[] | null;
	/** Soft memory limit in bytes. */
	MemoryReservation: number;
	/** Memory plus swap, in bytes. `-1` allows unlimited swap. */
	MemorySwap: number;
	/** Swappiness, `0` through `100`. Null when unset. */
	MemorySwappiness: number | null;
	/** True when the OOM killer is disabled for this container. Null when unset. */
	OomKillDisable: boolean | null;
	/** Process limit. `0` or `-1` means unlimited. Null when unset. */
	PidsLimit: number | null;
	/** Resource limits. `Hard` and `Soft` use the unit of that limit name. */
	Ulimits: { Name: string; Hard: number; Soft: number }[] | null;
	/** Windows CPU count. */
	CpuCount: number;
	/** Windows CPU percent. */
	CpuPercent: number;
	/** Windows maximum I/O operations per second for the system drive. */
	IOMaximumIOps: number;
	/** Windows maximum I/O bandwidth for the system drive, in bytes per second. */
	IOMaximumBandwidth: number;
	/** Mounts requested at create time. */
	Mounts?: ContainerMount[];
	/** Paths hidden inside the container. This replaces the default masked set. */
	MaskedPaths: string[] | null;
	/** Paths made read-only inside the container. This replaces the default read-only set. */
	ReadonlyPaths: string[] | null;
	/** True to run an init as pid 1. Null uses the daemon's configured default. */
	Init?: boolean;
};

/** A mount requested on a container. `Type` selects which options object is set. */
export type ContainerMount = {
	/** `bind`, `volume`, `tmpfs`, `npipe`, `image`, or `cluster`. */
	Type?: string;
	/** Host path or volume name. Empty for tmpfs. */
	Source?: string;
	/** Path inside the container. */
	Target?: string;
	/** True to mount read-only, recursively when the runtime can. */
	ReadOnly?: boolean;
	/** Consistency hint: `consistent`, `cached`, `delegated`, or `default`. */
	Consistency?: string;
	/** Set when {@link ContainerMount.Type} is `bind`. */
	BindOptions?: {
		/** Propagation: `rprivate`, `private`, `rshared`, `shared`, `rslave`, or `slave`. */
		Propagation?: string;
		/** True to skip mounting submounts. */
		NonRecursive?: boolean;
		/** True to create the host source path when it is missing. */
		CreateMountpoint?: boolean;
		/** True to make this mount read-only while leaving it recursive. */
		ReadOnlyNonRecursive?: boolean;
		/** True to fail the mount when it cannot be made recursively read-only. */
		ReadOnlyForceRecursive?: boolean;
	};
	/** Set when {@link ContainerMount.Type} is `volume`. */
	VolumeOptions?: {
		/** True to skip copying image data into a fresh volume. */
		NoCopy?: boolean;
		/** Labels stored on the volume. */
		Labels?: Record<string, string> | null;
		/** Subdirectory of the volume to mount. */
		Subpath?: string;
		/** Volume driver used to create the volume. */
		DriverConfig?: { Name?: string; Options?: Record<string, string> | null };
	};
	/** Set when {@link ContainerMount.Type} is `image`. `Subpath` is the path inside that image. */
	ImageOptions?: { Subpath?: string };
	/** Set when {@link ContainerMount.Type} is `tmpfs`. */
	TmpfsOptions?: {
		/** Size in bytes. */
		SizeBytes?: number;
		/** Permission bits applied when the tmpfs is created. */
		Mode?: number;
		/** Mount flags. A one-element array is a flag. A two-element array is a key and a value. */
		Options?: string[][] | null;
	};
	/** Set when {@link ContainerMount.Type} is `cluster`. The object is empty; the volume spec lives on the Swarm volume. */
	ClusterOptions?: Record<string, never>;
};

/** Network namespace and endpoints from container inspect. */
export type ContainerNetworkSettings = {
	/** Id of the container's network namespace. */
	SandboxID: string;
	/** Path that identifies the network namespace, such as a netns path. */
	SandboxKey: string;
	/** Published ports keyed by `port/protocol`. */
	Ports: Record<string, { HostIp: string; HostPort: string }[] | null> | null;
	/** Attachment to each network, keyed by network name. */
	Networks: Record<string, EndpointSettings | null> | null;
};

/** One container endpoint on a network. Addresses are strings. An unset address is empty. */
export type EndpointSettings = {
	/** Requested addresses. Null when the daemon assigned them. */
	IPAMConfig: { IPv4Address?: string; IPv6Address?: string; LinkLocalIPs?: string[] | null } | null;
	/** Legacy links on this endpoint. */
	Links: string[] | null;
	/** Extra DNS names for this endpoint. */
	Aliases: string[] | null;
	/** Driver options for this endpoint. */
	DriverOpts: Record<string, string> | null;
	/** Which endpoint supplies the default gateway. The highest priority wins. Equal priorities are ordered by network name. */
	GwPriority: number;
	/** Network id. */
	NetworkID: string;
	/** Endpoint id. */
	EndpointID: string;
	/** IPv4 gateway. */
	Gateway: string;
	/** IPv4 address. */
	IPAddress: string;
	/** MAC address. A value set at create time is kept; otherwise the daemon generates one. */
	MacAddress: string;
	/** IPv4 prefix length. */
	IPPrefixLen: number;
	/** IPv6 gateway. */
	IPv6Gateway: string;
	/** IPv6 address. */
	GlobalIPv6Address: string;
	/** IPv6 prefix length. */
	GlobalIPv6PrefixLen: number;
	/** DNS names for this endpoint. The first name is used for PTR records. */
	DNSNames: string[] | null;
};

/** A network from `listNetworks()` or the network half of `inspectNetwork()`. */
export type Network = {
	/** Network name. */
	Name: string;
	/** Network id. */
	Id: string;
	/** Creation time, RFC 3339 with nanoseconds. */
	Created: string;
	/** `local` for this daemon, or `swarm` for a cluster-wide network. */
	Scope: string;
	/** Driver, such as `bridge` or `overlay`. */
	Driver: string;
	/** True when the network has IPv4. */
	EnableIPv4: boolean;
	/** True when the network has IPv6. */
	EnableIPv6: boolean;
	/** Address pools. `Config` entries use CIDR strings. An omitted prefix or address is absent. */
	IPAM: {
		/** IPAM driver name. */
		Driver: string;
		/** Options for that driver. */
		Options: Record<string, string> | null;
		/** Pools. `Subnet` and `IPRange` are CIDRs. `Gateway` is an address. */
		Config: { Subnet?: string; IPRange?: string; Gateway?: string; AuxiliaryAddresses?: Record<string, string> | null }[] | null;
	};
	/** True when the network has no outbound connectivity. */
	Internal: boolean;
	/** True when a swarm-scoped network can be attached manually from a worker. */
	Attachable: boolean;
	/** True when this network is the swarm routing mesh. */
	Ingress: boolean;
	/** Config-only network this network copies its configuration from. */
	ConfigFrom: { Network: string };
	/** True when this network only holds configuration and cannot run containers. */
	ConfigOnly: boolean;
	/** Driver options set at create time. */
	Options: Record<string, string> | null;
	/** User labels. */
	Labels: Record<string, string> | null;
	/** Overlay peers. Omitted for other drivers. */
	Peers?: { Name: string; IP: string }[] | null;
};

/** Result of `listNetworks()`. */
export type NetworkList = {
	/** Networks on this daemon. */
	Items: Network[] | null;
};

/**
 * A network from `inspectNetwork()`, including who is attached.
 * `Containers` is keyed by container id. `Services` is present on swarm-scoped networks.
 */
export type NetworkInspect = Network & {
	/** Endpoints on this network, keyed by container id. */
	Containers: Record<
		string,
		{
			/** Container name. */
			Name: string;
			/** Endpoint id. */
			EndpointID: string;
			/** MAC address. */
			MacAddress: string;
			/** IPv4 address and prefix, such as `172.19.0.2/16`. */
			IPv4Address: string;
			/** IPv6 address and prefix. Empty when the network has no IPv6. */
			IPv6Address: string;
		}
	> | null;
	/** Swarm services using this network. Omitted on a local-scoped network. */
	Services?: Record<
		string,
		{
			/** Virtual IP of the service on this network. */
			VIP: string;
			/** Ports the service publishes on this network. */
			Ports: string[] | null;
			/** Index of this node in the service's local load balancer. */
			LocalLBIndex: number;
			/** Tasks attached to the service. `Info` is driver-specific. */
			Tasks: { Name: string; EndpointID: string; EndpointIP: string; Info: Record<string, string> | null }[] | null;
		}
	> | null;
	/** Runtime allocation counts. Present when inspect asked for verbose status. */
	Status?: {
		IPAM: {
			/** Allocation counts keyed by subnet CIDR. */
			Subnets?: Record<
				string,
				{
					/** Addresses currently allocated. */
					IPsInUse: number;
					/** Addresses still available from the dynamic pool. */
					DynamicIPsAvailable: number;
				}
			> | null;
		};
	};
};

/**
 * Result of `inspectNetwork()`.
 * `Network` is the decoded inspect document. `Raw` is that same document before decoding, or null.
 */
export type NetworkInspectResult = {
	/** Decoded network inspect document. */
	Network: NetworkInspect;
	/** Inspect JSON as the daemon sent it, decoded again into the same shape. Null when the client kept no raw body. */
	Raw: NetworkInspect | null;
};

/**
 * A volume from `listEngineVolumes()`, `createVolume()`, or `composeVolumeList()`.
 * Compose volume list returns this same Engine volume shape for the project's volumes.
 */
export type Volume = {
	/** Volume name. */
	Name: string;
	/** Volume driver, such as `local`. */
	Driver: string;
	/** Host path where the volume is mounted. */
	Mountpoint: string;
	/** Creation time, RFC 3339 with nanoseconds. Omitted when the driver does not report it. */
	CreatedAt?: string;
	/** User labels. */
	Labels: Record<string, string> | null;
	/** `local` for this daemon, or `global` for a cluster volume. */
	Scope: string;
	/** Driver options used at create time. */
	Options: Record<string, string> | null;
	/** Driver-specific status. Omitted when the driver does not report it. */
	Status?: Record<string, unknown> | null;
	/**
	 * Usage from a disk-usage query. Omitted on a plain list.
	 * `RefCount` is `-1` when the count is unavailable. `Size` is bytes, or `-1` for a driver other than `local`.
	 */
	UsageData?: { RefCount: number; Size: number };
	/** Swarm CSI volume. Present when {@link Volume.Scope} is `global`. */
	ClusterVolume?: ClusterVolume;
};

/** Swarm CSI cluster volume attached to a {@link Volume}. */
export type ClusterVolume = {
	/** Swarm object id. Distinct from the volume name, and distinct from the CSI plugin's volume id. */
	ID: string;
	/** Swarm object version. */
	Version?: { Index?: number };
	/** RFC 3339 time the Swarm object was created. */
	CreatedAt?: string;
	/** RFC 3339 time the Swarm object was last updated. */
	UpdatedAt?: string;
	/** Desired state. */
	Spec: {
		/** Volume group. Services can schedule any volume in the group. An empty group is one shared group. */
		Group?: string;
		/** How tasks may use the volume. `Scope` is `single` or `multi`. `Sharing` is `none`, `readonly`, `onewriter`, or `all`. */
		AccessMode?: {
			/** `single` is one node at a time. `multi` is many nodes at once. */
			Scope?: string;
			/** `none`, `readonly`, `onewriter`, or `all`. */
			Sharing?: string;
			/** Filesystem mount options. Set for a mount volume. */
			MountVolume?: { FsType?: string; MountFlags?: string[] | null };
			/** Present for a block volume. The object is empty. */
			BlockVolume?: Record<string, never>;
		};
		/** Topologies the volume must be reachable from. `Segments` are constraints such as zone. */
		AccessibilityRequirements?: {
			/** The volume must be reachable from at least one of these topologies. */
			Requisite?: { Segments?: Record<string, string> | null }[] | null;
			/** Topologies to prefer when more than one can reach the volume. */
			Preferred?: { Segments?: Record<string, string> | null }[] | null;
		};
		/** Requested size in bytes. `0` leaves that bound unspecified. */
		CapacityRange?: { RequiredBytes: number; LimitBytes: number };
		/** Swarm secrets passed to the CSI plugin. `Key` is the plugin key. `Secret` is the secret name or id. */
		Secrets?: { Key: string; Secret: string }[] | null;
		/** Desired availability: `active`, `pause`, or `drain`. */
		Availability?: string;
	};
	/**
	 * Per-node publish state.
	 * `State` is `pending-publish`, `published`, `pending-node-unpublish`, or `pending-controller-unpublish`.
	 */
	PublishStatus?: { NodeID?: string; State?: string; PublishContext?: Record<string, string> | null }[] | null;
	/** Status reported by the CSI plugin. `CapacityBytes` of `0` means unknown. */
	Info?: {
		/** Capacity in bytes. `0` means the plugin did not report it. */
		CapacityBytes?: number;
		/** Context the plugin stored when it created the volume. */
		VolumeContext?: Record<string, string> | null;
		/** Volume id as the CSI plugin knows it. Empty until creation succeeds. */
		VolumeID?: string;
		/** Topologies the volume is actually reachable from. */
		AccessibleTopology?: { Segments?: Record<string, string> | null }[] | null;
	};
};

/** Result of `listEngineVolumes()`. */
export type VolumeList = {
	/** Volumes matching the list filters. */
	Items: Volume[] | null;
	/** Warnings from the volume drivers. Null when there were none. */
	Warnings: string[] | null;
};

/** Result of `createVolume()`. */
export type VolumeCreate = {
	/** The volume as the daemon reports it after create. */
	Volume: Volume;
};

/**
 * One hit from `searchImages()`.
 * `is_automated` is always false; registries no longer report that flag.
 */
export type ImageSearchHit = {
	/** Repository name, such as `library/alpine`. */
	name: string;
	/** Star count reported by the registry. */
	star_count: number;
	/** True for an official repository. */
	is_official: boolean;
	/** Always false. Registries no longer mark results as automated. */
	is_automated: boolean;
	/** Registry description of the repository. */
	description: string;
};

/** Result of `searchImages()`. */
export type ImageSearchList = {
	/** Registry search hits. */
	Items: ImageSearchHit[] | null;
};

/** One layer from `historyImage()`. */
export type ImageHistoryItem = {
	/** Layer id. `<missing>` when the layer is not present locally. */
	Id: string;
	/** Creation time as Unix seconds. */
	Created: number;
	/** Command that created the layer. */
	CreatedBy: string;
	/** Comment stored on the layer. */
	Comment: string;
	/** Layer size in bytes. */
	Size: number;
	/** Tags that point at this layer. Null when none do. */
	Tags: string[] | null;
};

/** Result of `historyImage()`. */
export type ImageHistoryList = {
	/** Layers, newest first as the daemon returns them. */
	Items: ImageHistoryItem[] | null;
};

/**
 * Result of `topContainer()`.
 * `Titles` is the header row of the `ps` inside the container. Each `Processes` row lines up with those titles.
 */
export type ContainerTop = {
	/** Column titles from the process listing. */
	Titles: string[] | null;
	/** One row per process. Each cell lines up with {@link ContainerTop.Titles}. */
	Processes: string[][] | null;
};

/**
 * File metadata from `statPath()`.
 * `mode` is a Go `os.FileMode` bitfield. `mtime` is RFC 3339.
 */
export type PathStat = {
	/** File or directory name. */
	name: string;
	/** Size in bytes. */
	size: number;
	/** Go file mode bitfield, including the type bits. */
	mode: number;
	/** Modification time, RFC 3339. */
	mtime: string;
	/** Target path when this entry is a symlink. Empty otherwise. */
	linkTarget: string;
};

/** Result of `statPath()`. */
export type PathStatResult = {
	/** Metadata for the path inside the container. */
	Stat: PathStat;
};

/**
 * One project from `composeList()`.
 * `Status` is `Starting`, `Running`, `Updating`, `Removing`, `Unknown`, or `Failed`.
 */
export type ComposeStack = {
	/** Project id. */
	ID: string;
	/** Project name. */
	Name: string;
	/** `Starting`, `Running`, `Updating`, `Removing`, `Unknown`, or `Failed`. */
	Status: string;
	/** Compose files that define the project. */
	ConfigFiles: string;
	/** Why the project is in {@link ComposeStack.Status}, when the daemon recorded a reason. */
	Reason: string;
};

/**
 * One image from `composeImages()`.
 * The map is keyed by the container's canonical name. A container with no names uses the first 12 characters of its id.
 * `Created` is null when the image record is gone and only the container's own image reference remains.
 * `LastTagTime` is the zero time `0001-01-01T00:00:00Z` when the image has never been tagged locally.
 */
export type ComposeImageSummary = {
	/** Image id. */
	ID: string;
	/** Repository name, without the tag. */
	Repository: string;
	/** Tag. */
	Tag: string;
	/** Platform the image runs on. */
	Platform: ImagePlatform;
	/** Image size in bytes. */
	Size: number;
	/** Creation time, RFC 3339. Null when it is unknown. */
	Created: string | null;
	/** Last local tag time, RFC 3339. The zero time when the image has not been tagged locally. */
	LastTagTime: string;
};

/** One image removed or untagged by `pruneImages()`. A single prune entry sets one of the two fields. */
export type ImageDelete = {
	/** Id of an image that was deleted. */
	Deleted?: string;
	/** Id of an image that was untagged and left in the store. */
	Untagged?: string;
};

/** Result of `pruneImages()`. */
export type ImagePruneResult = {
	/** Images the prune removed, and the bytes it reclaimed. */
	Report: {
		/** Deleted and untagged images. Null when the prune removed nothing. */
		ImagesDeleted: ImageDelete[] | null;
		/** Disk space reclaimed, in bytes. */
		SpaceReclaimed: number;
	};
};

/** Result of `pruneContainers()`. */
export type ContainerPruneResult = {
	/** Containers the prune removed, and the bytes it reclaimed. */
	Report: {
		/** Ids of deleted containers. Null when the prune removed nothing. */
		ContainersDeleted: string[] | null;
		/** Disk space reclaimed, in bytes. */
		SpaceReclaimed: number;
	};
};

/** Result of `pruneNetworks()`. */
export type NetworkPruneResult = {
	/** Networks the prune removed. */
	Report: {
		/** Ids of deleted networks. Null when the prune removed nothing. */
		NetworksDeleted: string[] | null;
	};
};

/** Result of `pruneVolumes()`. Named volumes are included only when the prune asked for them. */
export type VolumePruneResult = {
	/** Volumes the prune removed, and the bytes it reclaimed. */
	Report: {
		/** Names of deleted volumes. Null when the prune removed nothing. */
		VolumesDeleted: string[] | null;
		/** Disk space reclaimed, in bytes. */
		SpaceReclaimed: number;
	};
};

/** Result of `login()`. `Status` is the registry's message, such as `Login Succeeded`. */
export type RegistryLogin = {
	Auth: {
		/** Opaque token for later requests. Omitted when the registry did not issue one. */
		IdentityToken?: string;
		/** Authentication status message. */
		Status: string;
	};
};

/** Result of `removeImage()`. Each item sets `Deleted` or `Untagged`. */
export type ImageRemoveResult = {
	/** Deleted and untagged images, in the order the daemon reported them. Null when nothing was removed. */
	Items: ImageDelete[] | null;
};

/**
 * Result of `diskUsage()`.
 * Counts and sizes are always present. `Items` is null unless the call asked for a verbose report.
 * Sizes are bytes.
 */
export type DiskUsage = {
	Containers: DiskUsageSection<ContainerSummary>;
	Images: DiskUsageSection<ImageSummary>;
	BuildCache: DiskUsageSection<BuildCacheRecord>;
	Volumes: DiskUsageSection<Volume>;
};

/** Counts for one kind of object in `diskUsage()`. */
export type DiskUsageSection<T> = {
	/** Objects currently in use. */
	ActiveCount: number;
	/** Objects of this kind. */
	TotalCount: number;
	/** Bytes that a prune could reclaim. */
	Reclaimable: number;
	/** Bytes used by this kind. */
	TotalSize: number;
	/** One record per object. Null when the call was not verbose. */
	Items: T[] | null;
};

/** One build-cache record from a verbose `diskUsage()`. */
export type BuildCacheRecord = {
	/** Cache record id. */
	ID: string;
	/**
	 * Parent cache record ids.
	 * The Engine type's JSON name includes a leading space.
	 */
	" Parents"?: string[];
	/** Cache record type. */
	Type: string;
	/** Build step that produced the record. */
	Description: string;
	/** True when a build is using this record. */
	InUse: boolean;
	/** True when another record shares this data. */
	Shared: boolean;
	/** Size in bytes. */
	Size: number;
	/** Creation time, RFC 3339. */
	CreatedAt: string;
	/** Last use time, RFC 3339. Null when the record has not been used. */
	LastUsedAt: string | null;
	/** Times this record has been used. */
	UsageCount: number;
};

/** Result of `listCheckpoints()`. */
export type CheckpointList = {
	/** Checkpoints for the container. Null when the container has none. */
	Items: Checkpoint[] | null;
};

/** One checkpoint from `listCheckpoints()`. */
export type Checkpoint = {
	/** Checkpoint name. */
	Name: string;
};

/** Object version stored by Swarm. Omitted when the index is zero. */
export type SwarmVersion = {
	Index?: number;
};

/** Result of `listConfigs()`. */
export type ConfigList = {
	/** Swarm configs. Null when the swarm has none. */
	Items: SwarmConfig[] | null;
};

/** Result of `inspectConfig()`. */
export type ConfigInspect = {
	/** Decoded config. */
	Config: SwarmConfig;
	/** Inspect JSON as the daemon sent it. Null when the client kept no raw body. */
	Raw: SwarmConfig | null;
};

/** One Swarm config from `listConfigs()` or `inspectConfig()`. */
export type SwarmConfig = {
	/** Config id. */
	ID: string;
	/** Object version. Omitted when the index is zero. */
	Version?: SwarmVersion;
	/** Creation time, RFC 3339. Omitted when unset. */
	CreatedAt?: string;
	/** Last update time, RFC 3339. Omitted when unset. */
	UpdatedAt?: string;
	Spec: SwarmConfigSpec;
};

/** Desired state of a Swarm config. `Data` is base64. */
export type SwarmConfigSpec = {
	/** Config name. Omitted when unset. */
	Name?: string;
	/** User labels. Null when the config has none. */
	Labels: Record<string, string> | null;
	/** Config payload, base64. Omitted when empty. */
	Data?: string;
	/** Template driver. Omitted when the payload is not a template. */
	Templating?: SwarmDriver;
};

/** Result of `listSecrets()`. */
export type SecretList = {
	/** Swarm secrets. Null when the swarm has none. */
	Items: SwarmSecret[] | null;
};

/** Result of `inspectSecret()`. */
export type SecretInspect = {
	/** Decoded secret. */
	Secret: SwarmSecret;
	/** Inspect JSON as the daemon sent it. Null when the client kept no raw body. */
	Raw: SwarmSecret | null;
};

/** One Swarm secret from `listSecrets()` or `inspectSecret()`. The payload is not returned after create. */
export type SwarmSecret = {
	/** Secret id. */
	ID: string;
	Version?: SwarmVersion;
	CreatedAt?: string;
	UpdatedAt?: string;
	Spec: SwarmSecretSpec;
};

/** Desired state of a Swarm secret. */
export type SwarmSecretSpec = {
	Name?: string;
	/** User labels. Null when the secret has none. */
	Labels: Record<string, string> | null;
	/** Secret payload, base64. Present on create, omitted by list and inspect. */
	Data?: string;
	/** External secret store. Omitted when the built-in store holds the secret. */
	Driver?: SwarmDriver;
	/** Template driver. Omitted when the payload is not a template. */
	Templating?: SwarmDriver;
};

/** A Swarm driver, such as a log driver or a secret backend. */
export type SwarmDriver = {
	Name?: string;
	Options?: Record<string, string>;
};

/** Result of `listPlugins()`. */
export type PluginList = {
	/** Installed plugins. Null when none are installed. */
	Items: EnginePlugin[] | null;
};

/** Result of `inspectPlugin()`. */
export type PluginInspect = {
	/** Decoded plugin. */
	Plugin: EnginePlugin;
	/** Inspect JSON as the daemon sent it. Null when the client kept no raw body. */
	Raw: EnginePlugin | null;
};

/** One Engine plugin from `listPlugins()` or `inspectPlugin()`. */
export type EnginePlugin = {
	/** Plugin id. Omitted when the daemon did not send one. */
	Id?: string;
	/** Plugin name. */
	Name: string;
	/** True when the plugin is running. */
	Enabled: boolean;
	/** Reference used to push or pull the plugin. */
	PluginReference?: string;
	Config: PluginConfig;
	Settings: PluginSettings;
};

/** Static config of an Engine plugin. */
export type PluginConfig = {
	Args: { Description: string; Name: string; Settable: string[] | null; Value: string[] | null };
	Description: string;
	Documentation: string;
	Entrypoint: string[] | null;
	Env: { Description: string; Name: string; Settable: string[] | null; Value: string | null }[] | null;
	Interface: {
		ProtocolScheme?: string;
		Socket: string;
		/** Interface types, such as `docker.volumedriver/1.0`. */
		Types: string[] | null;
	};
	IpcHost: boolean;
	Linux: {
		AllowAllDevices: boolean;
		Capabilities: string[] | null;
		Devices: { Description: string; Name: string; Path: string | null; Settable: string[] | null }[] | null;
	};
	Mounts: PluginMount[] | null;
	Network: { Type: string };
	PidHost: boolean;
	PropagatedMount: string;
	User?: { GID?: number; UID?: number };
	WorkDir: string;
	rootfs?: { diff_ids: string[] | null; type?: string };
};

/** User-configurable settings of an Engine plugin. */
export type PluginSettings = {
	Args: string[] | null;
	Devices: { Description: string; Name: string; Path: string | null; Settable: string[] | null }[] | null;
	Env: string[] | null;
	Mounts: PluginMount[] | null;
};

/** One mount declared by a plugin. */
export type PluginMount = {
	Description: string;
	Destination: string;
	Name: string;
	Options: string[] | null;
	Settable: string[] | null;
	Source: string | null;
	Type: string;
};

/** Result of `listNodes()`. */
export type NodeList = {
	/** Swarm nodes. Null when the swarm has none. */
	Items: SwarmNode[] | null;
};

/** Result of `inspectNode()`. */
export type NodeInspect = {
	Node: SwarmNode;
	Raw: SwarmNode | null;
};

/** One Swarm node from `listNodes()` or `inspectNode()`. */
export type SwarmNode = {
	ID: string;
	Version?: SwarmVersion;
	CreatedAt?: string;
	UpdatedAt?: string;
	Spec?: SwarmNodeSpec;
	Description?: SwarmNodeDescription;
	Status?: SwarmNodeStatus;
	/** Present when this node is a manager. */
	ManagerStatus?: { Leader?: boolean; Reachability?: string; Addr?: string };
};

/** Desired role and availability of a Swarm node. */
export type SwarmNodeSpec = {
	Name?: string;
	Labels: Record<string, string> | null;
	/** `worker` or `manager`. */
	Role?: string;
	/** `active`, `pause`, or `drain`. */
	Availability?: string;
};

/** Properties the agent reported for a Swarm node. */
export type SwarmNodeDescription = {
	Hostname?: string;
	Platform?: { Architecture?: string; OS?: string };
	Resources?: {
		NanoCPUs?: number;
		MemoryBytes?: number;
		GenericResources?: { NamedResourceSpec?: { Kind?: string; Value?: string }; DiscreteResourceSpec?: { Kind?: string; Value?: number } }[];
	};
	Engine?: {
		EngineVersion?: string;
		Labels?: Record<string, string>;
		Plugins?: { Type?: string; Name?: string }[];
	};
	TLSInfo?: { TrustRoot?: string; CertIssuerSubject?: string; CertIssuerPublicKey?: string };
};

/** Current state of a Swarm node, as a manager sees it. */
export type SwarmNodeStatus = {
	/** `unknown`, `down`, `ready`, or `disconnected`. */
	State?: string;
	Message?: string;
	/** Address the manager uses to reach the node. */
	Addr?: string;
};

/** Result of `listSwarmServices()`. */
export type ServiceList = {
	/** Swarm services. Null when the swarm has none. */
	Items: SwarmService[] | null;
};

/** Result of `inspectSwarmService()`. */
export type ServiceInspect = {
	Service: SwarmService;
	Raw: SwarmService | null;
};

/** One Swarm service from `listSwarmServices()` or `inspectSwarmService()`. */
export type SwarmService = {
	ID: string;
	Version?: SwarmVersion;
	CreatedAt?: string;
	UpdatedAt?: string;
	Spec?: SwarmServiceSpec;
	/** Previous spec, while an update is in progress. */
	PreviousSpec?: SwarmServiceSpec;
	Endpoint?: SwarmEndpoint;
	UpdateStatus?: {
		/** `updating`, `paused`, `completed`, `rollback_started`, `rollback_paused`, or `rollback_completed`. */
		State?: string;
		StartedAt?: string;
		CompletedAt?: string;
		Message?: string;
	};
	/** Task counts. Present when the list asked for status. */
	ServiceStatus?: { RunningTasks: number; DesiredTasks: number; CompletedTasks: number };
	/** Present on a replicated or global job. */
	JobStatus?: { JobIteration: SwarmVersion; LastExecution?: string };
};

/** Desired state of a Swarm service. */
export type SwarmServiceSpec = {
	Name?: string;
	Labels: Record<string, string> | null;
	TaskTemplate?: SwarmTaskSpec;
	Mode?: {
		Replicated?: { Replicas?: number };
		Global?: Record<string, never>;
		ReplicatedJob?: { MaxConcurrent?: number; TotalCompletions?: number };
		GlobalJob?: Record<string, never>;
	};
	UpdateConfig?: SwarmUpdateConfig;
	RollbackConfig?: SwarmUpdateConfig;
	EndpointSpec?: { Mode?: "vip" | "dnsrr"; Ports?: SwarmPort[] };
};

/** Rolling update settings on a Swarm service. Durations are nanoseconds. */
export type SwarmUpdateConfig = {
	/** Tasks updated at once. `0` means unlimited. */
	Parallelism: number;
	/** Delay between updates, in nanoseconds. */
	Delay?: number;
	/** `pause`, `continue`, or `rollback`. */
	FailureAction?: string;
	/** How long a new task is watched, in nanoseconds. */
	Monitor?: number;
	/** Fraction of tasks that may fail before {@link SwarmUpdateConfig.FailureAction} runs. */
	MaxFailureRatio: number;
	/** `stop-first` or `start-first`. */
	Order: string;
};

/** Result of `listTasks()`. */
export type TaskList = {
	/** Swarm tasks. Null when the swarm has none. */
	Items: SwarmTask[] | null;
};

/** Result of `inspectTask()`. */
export type TaskInspect = {
	Task: SwarmTask;
	Raw: SwarmTask | null;
};

/** One Swarm task from `listTasks()` or `inspectTask()`. */
export type SwarmTask = {
	ID: string;
	Version?: SwarmVersion;
	CreatedAt?: string;
	UpdatedAt?: string;
	Name?: string;
	Labels: Record<string, string> | null;
	Spec?: SwarmTaskSpec;
	ServiceID?: string;
	Slot?: number;
	NodeID?: string;
	Status?: SwarmTaskStatus;
	DesiredState?: SwarmTaskState;
	NetworksAttachments?: { Network?: { ID: string }; Addresses?: string[] }[];
	GenericResources?: {
		NamedResourceSpec?: { Kind?: string; Value?: string };
		DiscreteResourceSpec?: { Kind?: string; Value?: number };
	}[];
	JobIteration?: SwarmVersion;
	/** Volumes attached to this task. Null when the task has none. */
	Volumes: { ID?: string; Source?: string; Target?: string }[] | null;
};

/**
 * Task lifecycle state.
 * `new`, `allocated`, `pending`, `assigned`, `accepted`, `preparing`, `ready`, `starting`,
 * `running`, `complete`, `shutdown`, `failed`, `rejected`, `remove`, or `orphaned`.
 */
export type SwarmTaskState =
	| "new"
	| "allocated"
	| "pending"
	| "assigned"
	| "accepted"
	| "preparing"
	| "ready"
	| "starting"
	| "running"
	| "complete"
	| "shutdown"
	| "failed"
	| "rejected"
	| "remove"
	| "orphaned";

/** How a Swarm task is built and where it may run. */
export type SwarmTaskSpec = {
	ContainerSpec?: SwarmContainerSpec;
	PluginSpec?: { name?: string; remote?: string; disabled?: boolean; env?: string[] };
	NetworkAttachmentSpec?: { ContainerID: string };
	Resources?: { Limits?: SwarmResources; Reservations?: SwarmResources };
	RestartPolicy?: { Condition?: "none" | "on-failure" | "any"; Delay?: number; MaxAttempts?: number; Window?: number };
	Placement?: { Constraints?: string[]; Platforms?: { Architecture?: string; OS?: string }[] };
	Networks?: { Target?: string; Aliases?: string[] }[];
	LogDriver?: SwarmDriver;
	/** Incremented to force a new task even when the spec is otherwise unchanged. */
	ForceUpdate: number;
	/** `container`, `plugin`, or `attachment`. */
	Runtime?: string;
};

/** Container a Swarm task starts. Durations are nanoseconds. */
export type SwarmContainerSpec = {
	Image?: string;
	Labels?: Record<string, string>;
	Command?: string[];
	Args?: string[];
	Hostname?: string;
	Env?: string[];
	Dir?: string;
	User?: string;
	Groups?: string[];
	Privileges?: {
		CredentialSpec?: { Config: string; File: string; Registry: string };
		SELinuxContext?: { Disable: boolean; User: string; Role: string; Type: string; Level: string };
		Seccomp?: { Mode?: string; Profile?: string };
		AppArmor?: { Mode?: string };
		NoNewPrivileges: boolean;
	};
	Init?: boolean;
	StopSignal?: string;
	TTY?: boolean;
	OpenStdin?: boolean;
	ReadOnly?: boolean;
	Mounts?: { Type?: string; Source?: string; Target?: string; ReadOnly?: boolean }[];
	StopGracePeriod?: number;
	Healthcheck?: {
		Test?: string[];
		Interval?: number;
		Timeout?: number;
		StartPeriod?: number;
		StartInterval?: number;
		Retries?: number;
	};
	Hosts?: string[];
	DNSConfig?: { Nameservers?: string[]; Search?: string[]; Options?: string[] };
	Secrets?: { File?: { Name: string; UID: string; GID: string; Mode: number }; SecretID: string; SecretName: string }[];
	Configs?: {
		File?: { Name: string; UID: string; GID: string; Mode: number };
		Runtime?: Record<string, never>;
		ConfigID: string;
		ConfigName: string;
	}[];
	Isolation?: string;
	CapabilityAdd?: string[];
	CapabilityDrop?: string[];
	Sysctls?: Record<string, string>;
	Ulimits?: { Name: string; Hard: number; Soft: number }[];
	/** Adjustment to the container's oom_score_adj. */
	OomScoreAdj?: number;
};

/** CPU and memory requested or limited for a Swarm task. */
export type SwarmResources = {
	NanoCPUs?: number;
	MemoryBytes?: number;
	Pids?: number;
};

/** Current state of a Swarm task. */
export type SwarmTaskStatus = {
	/** Time of this status, RFC 3339. */
	Timestamp?: string;
	State?: SwarmTaskState;
	Message?: string;
	Err?: string;
	ContainerStatus?: { ContainerID: string; PID: number; ExitCode: number };
	PortStatus?: { Ports?: SwarmPort[] };
};

/** A port published by a Swarm service or task. */
export type SwarmPort = {
	Name?: string;
	Protocol?: string;
	/** Port inside the container. */
	TargetPort?: number;
	/** Port on the swarm hosts. */
	PublishedPort?: number;
	/** `ingress` or `host`. */
	PublishMode?: string;
};

/** Published ports and virtual IPs of a Swarm service. */
export type SwarmEndpoint = {
	Spec?: { Mode?: "vip" | "dnsrr"; Ports?: SwarmPort[] };
	Ports?: SwarmPort[];
	VirtualIPs?: { NetworkID?: string; Addr?: string }[];
};

/** Result of `inspectSwarm()`. */
export type SwarmInspect = {
	Swarm: SwarmInfo;
};

/** The swarm this daemon belongs to, including the join tokens. */
export type SwarmInfo = {
	ID: string;
	Version?: SwarmVersion;
	CreatedAt?: string;
	UpdatedAt?: string;
	Spec: SwarmSpec;
	TLSInfo: { TrustRoot?: string; CertIssuerSubject?: string; CertIssuerPublicKey?: string };
	RootRotationInProgress: boolean;
	/** Address pool prefixes, such as `10.0.0.0/8`. Null when the swarm uses the default pool. */
	DefaultAddrPool: string[] | null;
	SubnetSize: number;
	DataPathPort: number;
	JoinTokens: { Worker: string; Manager: string };
};

/** Cluster settings from `inspectSwarm()`. Durations are nanoseconds. */
export type SwarmSpec = {
	Name?: string;
	Labels: Record<string, string> | null;
	Orchestration?: { TaskHistoryRetentionLimit?: number };
	Raft?: {
		SnapshotInterval?: number;
		KeepOldSnapshots?: number;
		LogEntriesForSlowFollowers?: number;
		ElectionTick: number;
		HeartbeatTick: number;
	};
	Dispatcher?: { HeartbeatPeriod?: number };
	CAConfig?: { NodeCertExpiry?: number; SigningCACert?: string; SigningCAKey?: string; ForceRotate?: number };
	TaskDefaults?: { LogDriver?: SwarmDriver };
	EncryptionConfig: { AutoLockManagers: boolean };
};
