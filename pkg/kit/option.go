package kit

// ContainerOption configures a [Container].
type ContainerOption interface {
	applyContainer(*Container) error
}

// ServiceOption configures a [Service].
type ServiceOption interface {
	applyService(*Service) error
}

// JobOption configures a Compose [Job].
type JobOption interface {
	applyJob(*Job) error
}

// ComposeOption configures a [Compose].
type ComposeOption interface {
	applyCompose(*Compose) error
}

// PortOption configures a nested [Publish] scope.
type PortOption interface {
	applyPort(*port) error
}

// HealthOption configures a nested [Health] scope.
type HealthOption interface {
	applyHealth(*health) error
}

// RestartOption configures a nested [Restart] scope.
type RestartOption interface {
	applyRestart(*restart) error
}

// DependOption configures a nested [DependsOn] scope.
type DependOption interface {
	applyDepend(*depend) error
}

// VolumeOption configures a nested [Volume] scope (bind, named, tmpfs).
type VolumeOption interface {
	applyVolume(*volume) error
}

// NetworkOption configures a nested [Network] attachment.
type NetworkOption interface {
	applyNetwork(*netAttach) error
}

// LogOption configures a nested [Logging] scope.
type LogOption interface {
	applyLog(*logging) error
}

// DeviceOption configures a nested [Device] mapping.
type DeviceOption interface {
	applyDevice(*device) error
}

// GPUOption configures a nested [GPU] device request.
type GPUOption interface {
	applyGPU(*gpu) error
}

// UlimitOption configures a nested [Ulimit] scope.
type UlimitOption interface {
	applyUlimit(*ulimit) error
}

// BuildOption configures a nested [Build] scope.
type BuildOption interface {
	applyBuild(*build) error
}

// SecretOption configures a nested [Secret] reference or definition.
type SecretOption interface {
	applySecret(*secretRef) error
}

// FileConfigOption configures a nested Compose config-file reference or definition.
type FileConfigOption interface {
	applyFileConfig(*fileConfigRef) error
}

// DeployOption configures a nested [Deploy] scope.
type DeployOption interface {
	applyDeploy(*deploy) error
}

// DevelopOption configures a nested [Develop] scope.
type DevelopOption interface {
	applyDevelop(*develop) error
}

// WatchTriggerOption configures a Compose develop.watch entry inside [Develop].
type WatchTriggerOption interface {
	applyWatchTrigger(*watchTrigger) error
}

// ModelOption configures a Compose models entry inside [Model].
type ModelOption interface {
	applyModel(*serviceModel) error
}

// UpdateOption configures nested [Update] / [Rollback] deploy settings.
type UpdateOption interface {
	applyUpdate(*updateCfg) error
}

// PlacementOption configures nested [Placement] deploy constraints.
type PlacementOption interface {
	applyPlacement(*placement) error
}

// ReserveOption configures deploy.resources.reservations inside [Reserve].
type ReserveOption interface {
	applyReserve(*resource) error
}

// ComposeNetOption configures a top-level Compose [AddNetwork] definition.
type ComposeNetOption interface {
	applyComposeNet(*projectNet) error
}

// ComposeVolOption configures a top-level Compose [AddVolume] definition.
type ComposeVolOption interface {
	applyComposeVol(*projectVol) error
}

// ComposeSecretOption configures a top-level Compose [AddSecret] definition.
type ComposeSecretOption interface {
	applyComposeSecret(*projectSecret) error
}

// ComposeFileOption configures a top-level Compose [AddFile] definition.
type ComposeFileOption interface {
	applyComposeFile(*projectFile) error
}

// ComposeModelOption configures a top-level Compose [AddModel] definition.
type ComposeModelOption interface {
	applyComposeModel(*projectModel) error
}

// IncludeOption configures a top-level Compose [AddInclude] entry.
type IncludeOption interface {
	applyInclude(*includeRef) error
}

// HookOption configures a Compose post_start / pre_stop command.
type HookOption interface {
	applyHook(*serviceHook) error
}

// EnvFileOption configures long-syntax env_file attributes inside [EnvFile].
type EnvFileOption interface {
	applyEnvFile(*envFileRef) error
}

// ImportOption configures [ImportYAML], [ImportFile], and [Import].
type ImportOption interface {
	applyImport(*importAct) error
}

// LoggerOption configures [NewLogger].
type LoggerOption interface {
	applyLogger(*Logger) error
}

// UpOption configures [Compose.Up].
type UpOption interface {
	applyUp(*upAct) error
}

// DownOption configures [Compose.Down].
type DownOption interface {
	applyDown(*downAct) error
}

// ComposeBuildOption configures [Compose.Build].
type ComposeBuildOption interface {
	applyComposeBuild(*composeBuildAct) error
}

// ComposeCreateOption configures [Compose.Create].
type ComposeCreateOption interface {
	applyComposeCreate(*createAct) error
}

// ComposeStartOption configures [Compose.Start].
type ComposeStartOption interface {
	applyComposeStart(*startAct) error
}

// ComposeStopOption configures [Compose.Stop].
type ComposeStopOption interface {
	applyComposeStop(*stopAct) error
}

// ComposeRestartOption configures [Compose.Restart].
type ComposeRestartOption interface {
	applyComposeRestart(*restartAct) error
}

// PullOption configures [Compose.Pull].
type PullOption interface {
	applyPull(*pullAct) error
}

// PushOption configures [Compose.Push].
type PushOption interface {
	applyPush(*pushAct) error
}

// LogsOption configures [Compose.Logs] and [Container.Logs].
type LogsOption interface {
	applyLogs(*logsAct) error
}

// PsOption configures [Compose.Ps].
type PsOption interface {
	applyPs(*psAct) error
}

// KillOption configures [Compose.Kill] and [Container.Kill].
type KillOption interface {
	applyKill(*killAct) error
}

// RemoveOption configures [Compose.Remove] and [Container.Remove].
type RemoveOption interface {
	applyRemove(*removeAct) error
}

// RunOption configures [Compose.Run] and [Compose.Exec].
type RunOption interface {
	applyRun(*runAct) error
}

// PauseOption configures [Compose.Pause] and [Compose.Unpause].
type PauseOption interface {
	applyPause(*pauseAct) error
}

// EventsOption configures [Compose.Events].
type EventsOption interface {
	applyEvents(*eventsAct) error
}

// ComposePortOption configures [Compose.Port].
type ComposePortOption interface {
	applyComposePort(*composePortAct) error
}

// WatchOption configures [Compose.Watch].
type WatchOption interface {
	applyWatch(*watchAct) error
}

// WaitOption configures [Compose.Wait].
type WaitOption interface {
	applyWait(*waitAct) error
}

// ScaleOption configures [Compose.Scale].
type ScaleOption interface {
	applyScale(*scaleAct) error
}

// CopyOption configures [Compose.Copy].
type CopyOption interface {
	applyCopy(*copyAct) error
}

// AttachOption configures [Compose.Attach].
type AttachOption interface {
	applyAttach(*attachAct) error
}

// PublishOption configures [Compose.PublishOCI].
type PublishOption interface {
	applyPublish(*publishAct) error
}

// ImagesOption configures [Compose.Images].
type ImagesOption interface {
	applyImages(*imagesAct) error
}

// ExportOption configures [Compose.Export].
type ExportOption interface {
	applyExport(*exportAct) error
}

// CommitOption configures [Compose.Commit].
type CommitOption interface {
	applyCommit(*commitAct) error
}

// VizOption configures [Compose.Viz].
type VizOption interface {
	applyViz(*vizAct) error
}

// ListOption configures [List].
type ListOption interface {
	applyList(*listAct) error
}

// VolumesListOption configures [Compose.VolumeList].
type VolumesListOption interface {
	applyVolumesList(*volumesListAct) error
}

// EngineCreateOption configures [Container.Create].
type EngineCreateOption interface {
	applyEngineCreate(*engineCreateAct) error
}

// EngineStartOption configures [Container.Start].
type EngineStartOption interface {
	applyEngineStart(*engineStartAct) error
}

// EngineStopOption configures [Container.Stop] and [Container.Restart].
type EngineStopOption interface {
	applyEngineStop(*engineStopAct) error
}

// EngineInspectOption configures [Container.Inspect].
type EngineInspectOption interface {
	applyEngineInspect(*engineInspectAct) error
}

// EngineWaitOption configures [Container.Wait].
type EngineWaitOption interface {
	applyEngineWait(*engineWaitAct) error
}

// GenerateOption configures [Generate].
type GenerateOption interface {
	applyGenerate(*generateAct) error
}

// LoadProjectOption configures [LoadProject].
type LoadProjectOption interface {
	applyLoadProject(*loadProjectAct) error
}

// EnginePullOption configures [PullImage].
type EnginePullOption interface {
	applyEnginePull(*enginePullAct) error
}

// EngineBuildOption configures [BuildImage].
type EngineBuildOption interface {
	applyEngineBuild(*engineBuildAct) error
}

// EngineNetworkOption configures [CreateNetwork].
type EngineNetworkOption interface {
	applyEngineNetwork(*engineNetworkAct) error
}

// EngineVolumeOption configures [CreateVolume].
type EngineVolumeOption interface {
	applyEngineVolume(*engineVolumeAct) error
}

// ClientOption configures [NewEngine].
type ClientOption interface {
	applyClient(*clientAct) error
}

// ComposerClientOption configures [NewComposer]. Engine [ClientOption]
// values also implement this interface, as do [Log], [MaxConcurrency],
// [DryRun], and [OutputStream].
type ComposerClientOption interface {
	applyComposerClient(*composerAct) error
}

// TLSOption configures nested [TLS] client credentials.
type TLSOption interface {
	applyTLS(*tlsAct) error
}

// SpecOpt is the value returned by [BlkioWeightDevice], [BlkioReadBps], [BlkioWriteBps],
// [BlkioReadIOps], [BlkioWriteIOps], [DeviceCgroupRule], [IPC], [PID], [UTS], [UserNS],
// [Cgroup], [CgroupParent], [CgroupNS], [Runtime], [Umask], [NetworkDisabled], [DNS],
// [DNSOption], [DNSSearch], [Link], [MemoryReservation], [MemorySwap], [MemorySwappiness],
// [ShmSize], [CPUShares], [CPUPeriod], [CPUQuota], [CPUSet], [CPUSetMems], [IOMaximumIOps],
// [IOMaximumBandwidth], [CPURealtimePeriod], [CPURealtimeRuntime], [CPUCount], [CPUPercent],
// [OomScoreAdj], [OomKillDisable], [BlkioWeight], [Sysctl], [SecurityOpt], [StorageOpt],
// [GroupAdd], [Annotation], [Hostname], [DomainName], [Init], [StopTimeout], [VolumesFrom], and
// [VolumeDriver].
// It implements [ContainerOption], [JobOption], and [ServiceOption].
type SpecOpt struct {
	fn func(*spec) error
}

// specFn returns a SpecOpt that applies the given function t
func specFn(fn func(*spec) error) SpecOpt { return SpecOpt{fn: fn} }

func (o SpecOpt) applyContainer(c *Container) error {
	if o.fn == nil {
		return nil
	}
	return o.fn(&c.spec)
}
func (o SpecOpt) applyService(s *Service) error {
	if o.fn == nil {
		return nil
	}
	return o.fn(&s.spec)
}
func (o SpecOpt) applyJob(j *Job) error {
	if o.fn == nil {
		return nil
	}
	return o.fn(&j.spec)
}
