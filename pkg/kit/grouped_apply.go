package kit

// Apply methods so [Grouped] implements every option interface.

func (g Grouped) applyContainer(v *Container) error {
	return g.apply("container", func(o any) error {
		opt, ok := o.(ContainerOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyContainer(v)
	})
}

func (g Grouped) applyService(v *Service) error {
	return g.apply("service", func(o any) error {
		opt, ok := o.(ServiceOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyService(v)
	})
}

func (g Grouped) applyJob(v *Job) error {
	return g.apply("job", func(o any) error {
		opt, ok := o.(JobOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyJob(v)
	})
}

func (g Grouped) applyCompose(v *Compose) error {
	return g.apply("compose", func(o any) error {
		opt, ok := o.(ComposeOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyCompose(v)
	})
}

func (g Grouped) applyPort(v *port) error {
	return g.apply("port", func(o any) error {
		opt, ok := o.(PortOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyPort(v)
	})
}

func (g Grouped) applyHealth(v *health) error {
	return g.apply("health", func(o any) error {
		opt, ok := o.(HealthOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyHealth(v)
	})
}

func (g Grouped) applyRestart(v *restart) error {
	return g.apply("restart", func(o any) error {
		opt, ok := o.(RestartOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyRestart(v)
	})
}

func (g Grouped) applyDepend(v *depend) error {
	return g.apply("depend", func(o any) error {
		opt, ok := o.(DependOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyDepend(v)
	})
}

func (g Grouped) applyVolume(v *volume) error {
	return g.apply("volume", func(o any) error {
		opt, ok := o.(VolumeOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyVolume(v)
	})
}

func (g Grouped) applyNetwork(v *netAttach) error {
	return g.apply("network", func(o any) error {
		opt, ok := o.(NetworkOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyNetwork(v)
	})
}

func (g Grouped) applyLog(v *logging) error {
	return g.apply("log", func(o any) error {
		opt, ok := o.(LogOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyLog(v)
	})
}

func (g Grouped) applyDevice(v *device) error {
	return g.apply("device", func(o any) error {
		opt, ok := o.(DeviceOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyDevice(v)
	})
}

func (g Grouped) applyGPU(v *gpu) error {
	return g.apply("gPU", func(o any) error {
		opt, ok := o.(GPUOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyGPU(v)
	})
}

func (g Grouped) applyUlimit(v *ulimit) error {
	return g.apply("ulimit", func(o any) error {
		opt, ok := o.(UlimitOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyUlimit(v)
	})
}

func (g Grouped) applyBuild(v *build) error {
	return g.apply("build", func(o any) error {
		opt, ok := o.(BuildOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyBuild(v)
	})
}

func (g Grouped) applySecret(v *secretRef) error {
	return g.apply("secret", func(o any) error {
		opt, ok := o.(SecretOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applySecret(v)
	})
}

func (g Grouped) applyFileConfig(v *fileConfigRef) error {
	return g.apply("fileConfig", func(o any) error {
		opt, ok := o.(FileConfigOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyFileConfig(v)
	})
}

func (g Grouped) applyDeploy(v *deploy) error {
	return g.apply("deploy", func(o any) error {
		opt, ok := o.(DeployOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyDeploy(v)
	})
}

func (g Grouped) applyDevelop(v *develop) error {
	return g.apply("develop", func(o any) error {
		opt, ok := o.(DevelopOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyDevelop(v)
	})
}

func (g Grouped) applyWatchTrigger(v *watchTrigger) error {
	return g.apply("watchTrigger", func(o any) error {
		opt, ok := o.(WatchTriggerOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyWatchTrigger(v)
	})
}

func (g Grouped) applyModel(v *serviceModel) error {
	return g.apply("model", func(o any) error {
		opt, ok := o.(ModelOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyModel(v)
	})
}

func (g Grouped) applyUpdate(v *updateCfg) error {
	return g.apply("update", func(o any) error {
		opt, ok := o.(UpdateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyUpdate(v)
	})
}

func (g Grouped) applyPlacement(v *placement) error {
	return g.apply("placement", func(o any) error {
		opt, ok := o.(PlacementOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyPlacement(v)
	})
}

func (g Grouped) applyReserve(v *resource) error {
	return g.apply("reserve", func(o any) error {
		opt, ok := o.(ReserveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyReserve(v)
	})
}

func (g Grouped) applyComposeNet(v *projectNet) error {
	return g.apply("composeNet", func(o any) error {
		opt, ok := o.(ComposeNetOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeNet(v)
	})
}

func (g Grouped) applyComposeVol(v *projectVol) error {
	return g.apply("composeVol", func(o any) error {
		opt, ok := o.(ComposeVolOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeVol(v)
	})
}

func (g Grouped) applyComposeSecret(v *projectSecret) error {
	return g.apply("composeSecret", func(o any) error {
		opt, ok := o.(ComposeSecretOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeSecret(v)
	})
}

func (g Grouped) applyComposeFile(v *projectFile) error {
	return g.apply("composeFile", func(o any) error {
		opt, ok := o.(ComposeFileOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeFile(v)
	})
}

func (g Grouped) applyComposeModel(v *projectModel) error {
	return g.apply("composeModel", func(o any) error {
		opt, ok := o.(ComposeModelOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeModel(v)
	})
}

func (g Grouped) applyInclude(v *includeRef) error {
	return g.apply("include", func(o any) error {
		opt, ok := o.(IncludeOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyInclude(v)
	})
}

func (g Grouped) applyHook(v *serviceHook) error {
	return g.apply("hook", func(o any) error {
		opt, ok := o.(HookOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyHook(v)
	})
}

func (g Grouped) applyEnvFile(v *envFileRef) error {
	return g.apply("envFile", func(o any) error {
		opt, ok := o.(EnvFileOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnvFile(v)
	})
}

func (g Grouped) applyImport(v *importAct) error {
	return g.apply("import", func(o any) error {
		opt, ok := o.(ImportOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyImport(v)
	})
}

func (g Grouped) applyLogger(v *Logger) error {
	return g.apply("logger", func(o any) error {
		opt, ok := o.(LoggerOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyLogger(v)
	})
}

func (g Grouped) applyUp(v *upAct) error {
	return g.apply("up", func(o any) error {
		opt, ok := o.(UpOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyUp(v)
	})
}

func (g Grouped) applyDown(v *downAct) error {
	return g.apply("down", func(o any) error {
		opt, ok := o.(DownOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyDown(v)
	})
}

func (g Grouped) applyComposeBuild(v *composeBuildAct) error {
	return g.apply("composeBuild", func(o any) error {
		opt, ok := o.(ComposeBuildOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeBuild(v)
	})
}

func (g Grouped) applyComposeCreate(v *createAct) error {
	return g.apply("composeCreate", func(o any) error {
		opt, ok := o.(ComposeCreateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeCreate(v)
	})
}

func (g Grouped) applyComposeStart(v *startAct) error {
	return g.apply("composeStart", func(o any) error {
		opt, ok := o.(ComposeStartOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeStart(v)
	})
}

func (g Grouped) applyComposeStop(v *stopAct) error {
	return g.apply("composeStop", func(o any) error {
		opt, ok := o.(ComposeStopOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeStop(v)
	})
}

func (g Grouped) applyComposeRestart(v *restartAct) error {
	return g.apply("composeRestart", func(o any) error {
		opt, ok := o.(ComposeRestartOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposeRestart(v)
	})
}

func (g Grouped) applyPull(v *pullAct) error {
	return g.apply("pull", func(o any) error {
		opt, ok := o.(PullOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyPull(v)
	})
}

func (g Grouped) applyPush(v *pushAct) error {
	return g.apply("push", func(o any) error {
		opt, ok := o.(PushOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyPush(v)
	})
}

func (g Grouped) applyLogs(v *logsAct) error {
	return g.apply("logs", func(o any) error {
		opt, ok := o.(LogsOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyLogs(v)
	})
}

func (g Grouped) applyPs(v *psAct) error {
	return g.apply("ps", func(o any) error {
		opt, ok := o.(PsOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyPs(v)
	})
}

func (g Grouped) applyKill(v *killAct) error {
	return g.apply("kill", func(o any) error {
		opt, ok := o.(KillOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyKill(v)
	})
}

func (g Grouped) applyRemove(v *removeAct) error {
	return g.apply("remove", func(o any) error {
		opt, ok := o.(RemoveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyRemove(v)
	})
}

func (g Grouped) applyRun(v *runAct) error {
	return g.apply("run", func(o any) error {
		opt, ok := o.(RunOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyRun(v)
	})
}

func (g Grouped) applyPause(v *pauseAct) error {
	return g.apply("pause", func(o any) error {
		opt, ok := o.(PauseOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyPause(v)
	})
}

func (g Grouped) applyEvents(v *eventsAct) error {
	return g.apply("events", func(o any) error {
		opt, ok := o.(EventsOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEvents(v)
	})
}

func (g Grouped) applyComposePort(v *composePortAct) error {
	return g.apply("composePort", func(o any) error {
		opt, ok := o.(ComposePortOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposePort(v)
	})
}

func (g Grouped) applyWatch(v *watchAct) error {
	return g.apply("watch", func(o any) error {
		opt, ok := o.(WatchOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyWatch(v)
	})
}

func (g Grouped) applyWait(v *waitAct) error {
	return g.apply("wait", func(o any) error {
		opt, ok := o.(WaitOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyWait(v)
	})
}

func (g Grouped) applyScale(v *scaleAct) error {
	return g.apply("scale", func(o any) error {
		opt, ok := o.(ScaleOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyScale(v)
	})
}

func (g Grouped) applyCopy(v *copyAct) error {
	return g.apply("copy", func(o any) error {
		opt, ok := o.(CopyOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyCopy(v)
	})
}

func (g Grouped) applyAttach(v *attachAct) error {
	return g.apply("attach", func(o any) error {
		opt, ok := o.(AttachOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyAttach(v)
	})
}

func (g Grouped) applyPublish(v *publishAct) error {
	return g.apply("publish", func(o any) error {
		opt, ok := o.(PublishOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyPublish(v)
	})
}

func (g Grouped) applyImages(v *imagesAct) error {
	return g.apply("images", func(o any) error {
		opt, ok := o.(ImagesOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyImages(v)
	})
}

func (g Grouped) applyExport(v *exportAct) error {
	return g.apply("export", func(o any) error {
		opt, ok := o.(ExportOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyExport(v)
	})
}

func (g Grouped) applyCommit(v *commitAct) error {
	return g.apply("commit", func(o any) error {
		opt, ok := o.(CommitOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyCommit(v)
	})
}

func (g Grouped) applyViz(v *vizAct) error {
	return g.apply("viz", func(o any) error {
		opt, ok := o.(VizOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyViz(v)
	})
}

func (g Grouped) applyList(v *listAct) error {
	return g.apply("list", func(o any) error {
		opt, ok := o.(ListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyList(v)
	})
}

func (g Grouped) applyVolumesList(v *volumesListAct) error {
	return g.apply("volumesList", func(o any) error {
		opt, ok := o.(VolumesListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyVolumesList(v)
	})
}

func (g Grouped) applyEngineCreate(v *engineCreateAct) error {
	return g.apply("engineCreate", func(o any) error {
		opt, ok := o.(EngineCreateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineCreate(v)
	})
}

func (g Grouped) applyEngineStart(v *engineStartAct) error {
	return g.apply("engineStart", func(o any) error {
		opt, ok := o.(EngineStartOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineStart(v)
	})
}

func (g Grouped) applyEngineStop(v *engineStopAct) error {
	return g.apply("engineStop", func(o any) error {
		opt, ok := o.(EngineStopOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineStop(v)
	})
}

func (g Grouped) applyEngineInspect(v *engineInspectAct) error {
	return g.apply("engineInspect", func(o any) error {
		opt, ok := o.(EngineInspectOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineInspect(v)
	})
}

func (g Grouped) applyEngineWait(v *engineWaitAct) error {
	return g.apply("engineWait", func(o any) error {
		opt, ok := o.(EngineWaitOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineWait(v)
	})
}

func (g Grouped) applyGenerate(v *generateAct) error {
	return g.apply("generate", func(o any) error {
		opt, ok := o.(GenerateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyGenerate(v)
	})
}

func (g Grouped) applyLoadProject(v *loadProjectAct) error {
	return g.apply("loadProject", func(o any) error {
		opt, ok := o.(LoadProjectOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyLoadProject(v)
	})
}

func (g Grouped) applyEnginePull(v *enginePullAct) error {
	return g.apply("enginePull", func(o any) error {
		opt, ok := o.(EnginePullOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePull(v)
	})
}

func (g Grouped) applyEngineBuild(v *engineBuildAct) error {
	return g.apply("engineBuild", func(o any) error {
		opt, ok := o.(EngineBuildOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineBuild(v)
	})
}

func (g Grouped) applyEngineNetwork(v *engineNetworkAct) error {
	return g.apply("engineNetwork", func(o any) error {
		opt, ok := o.(EngineNetworkOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNetwork(v)
	})
}

func (g Grouped) applyEngineVolume(v *engineVolumeAct) error {
	return g.apply("engineVolume", func(o any) error {
		opt, ok := o.(EngineVolumeOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineVolume(v)
	})
}

func (g Grouped) applyClient(v *clientAct) error {
	return g.apply("client", func(o any) error {
		opt, ok := o.(ClientOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyClient(v)
	})
}

func (g Grouped) applyComposerClient(v *composerAct) error {
	return g.apply("composerClient", func(o any) error {
		opt, ok := o.(ComposerClientOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyComposerClient(v)
	})
}

func (g Grouped) applyTLS(v *tlsAct) error {
	return g.apply("tLS", func(o any) error {
		opt, ok := o.(TLSOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyTLS(v)
	})
}

func (g Grouped) applyEngineContainerListOption(v *engineContainerListAct) error {
	return g.apply("engineContainerList", func(o any) error {
		opt, ok := o.(EngineContainerListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineContainerListOption(v)
	})
}

func (g Grouped) applyEngineContainerPruneOption(v *engineContainerPruneAct) error {
	return g.apply("engineContainerPrune", func(o any) error {
		opt, ok := o.(EngineContainerPruneOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineContainerPruneOption(v)
	})
}

func (g Grouped) applyEngineRenameOption(v *engineRenameAct) error {
	return g.apply("engineRename", func(o any) error {
		opt, ok := o.(EngineRenameOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineRenameOption(v)
	})
}

func (g Grouped) applyEngineResizeOption(v *engineResizeAct) error {
	return g.apply("engineResize", func(o any) error {
		opt, ok := o.(EngineResizeOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineResizeOption(v)
	})
}

func (g Grouped) applyEngineAttachOption(v *engineAttachAct) error {
	return g.apply("engineAttach", func(o any) error {
		opt, ok := o.(EngineAttachOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineAttachOption(v)
	})
}

func (g Grouped) applyEngineCommitOption(v *engineCommitAct) error {
	return g.apply("engineCommit", func(o any) error {
		opt, ok := o.(EngineCommitOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineCommitOption(v)
	})
}

func (g Grouped) applyEngineStatsOption(v *engineStatsAct) error {
	return g.apply("engineStats", func(o any) error {
		opt, ok := o.(EngineStatsOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineStatsOption(v)
	})
}

func (g Grouped) applyEngineTopOption(v *engineTopAct) error {
	return g.apply("engineTop", func(o any) error {
		opt, ok := o.(EngineTopOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineTopOption(v)
	})
}

func (g Grouped) applyEngineStatPathOption(v *engineStatPathAct) error {
	return g.apply("engineStatPath", func(o any) error {
		opt, ok := o.(EngineStatPathOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineStatPathOption(v)
	})
}

func (g Grouped) applyEngineCopyFromOption(v *engineCopyFromAct) error {
	return g.apply("engineCopyFrom", func(o any) error {
		opt, ok := o.(EngineCopyFromOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineCopyFromOption(v)
	})
}

func (g Grouped) applyEngineCopyToOption(v *engineCopyToAct) error {
	return g.apply("engineCopyTo", func(o any) error {
		opt, ok := o.(EngineCopyToOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineCopyToOption(v)
	})
}

func (g Grouped) applyEngineUpdateOption(v *engineUpdateAct) error {
	return g.apply("engineUpdate", func(o any) error {
		opt, ok := o.(EngineUpdateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineUpdateOption(v)
	})
}

func (g Grouped) applyEngineExecCreateOption(v *engineExecCreateAct) error {
	return g.apply("engineExecCreate", func(o any) error {
		opt, ok := o.(EngineExecCreateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineExecCreateOption(v)
	})
}

func (g Grouped) applyEngineExecStartOption(v *engineExecStartAct) error {
	return g.apply("engineExecStart", func(o any) error {
		opt, ok := o.(EngineExecStartOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineExecStartOption(v)
	})
}

func (g Grouped) applyEngineExecAttachOption(v *engineExecAttachAct) error {
	return g.apply("engineExecAttach", func(o any) error {
		opt, ok := o.(EngineExecAttachOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineExecAttachOption(v)
	})
}

func (g Grouped) applyEngineImageListOption(v *engineImageListAct) error {
	return g.apply("engineImageList", func(o any) error {
		opt, ok := o.(EngineImageListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageListOption(v)
	})
}

func (g Grouped) applyEngineImagePushOption(v *engineImagePushAct) error {
	return g.apply("engineImagePush", func(o any) error {
		opt, ok := o.(EngineImagePushOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImagePushOption(v)
	})
}

func (g Grouped) applyEngineImageRemoveOption(v *engineImageRemoveAct) error {
	return g.apply("engineImageRemove", func(o any) error {
		opt, ok := o.(EngineImageRemoveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageRemoveOption(v)
	})
}

func (g Grouped) applyEngineImageTagOption(v *engineImageTagAct) error {
	return g.apply("engineImageTag", func(o any) error {
		opt, ok := o.(EngineImageTagOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageTagOption(v)
	})
}

func (g Grouped) applyEngineImagePruneOption(v *engineImagePruneAct) error {
	return g.apply("engineImagePrune", func(o any) error {
		opt, ok := o.(EngineImagePruneOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImagePruneOption(v)
	})
}

func (g Grouped) applyEngineImageSearchOption(v *engineImageSearchAct) error {
	return g.apply("engineImageSearch", func(o any) error {
		opt, ok := o.(EngineImageSearchOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageSearchOption(v)
	})
}

func (g Grouped) applyEngineImageImportOption(v *engineImageImportAct) error {
	return g.apply("engineImageImport", func(o any) error {
		opt, ok := o.(EngineImageImportOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageImportOption(v)
	})
}

func (g Grouped) applyEngineBuildCachePruneOption(v *engineBuildCachePruneAct) error {
	return g.apply("engineBuildCachePrune", func(o any) error {
		opt, ok := o.(EngineBuildCachePruneOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineBuildCachePruneOption(v)
	})
}

func (g Grouped) applyEngineNetworkInspectOption(v *engineNetworkInspectAct) error {
	return g.apply("engineNetworkInspect", func(o any) error {
		opt, ok := o.(EngineNetworkInspectOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNetworkInspectOption(v)
	})
}

func (g Grouped) applyEngineNetworkListOption(v *engineNetworkListAct) error {
	return g.apply("engineNetworkList", func(o any) error {
		opt, ok := o.(EngineNetworkListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNetworkListOption(v)
	})
}

func (g Grouped) applyEngineNetworkPruneOption(v *engineNetworkPruneAct) error {
	return g.apply("engineNetworkPrune", func(o any) error {
		opt, ok := o.(EngineNetworkPruneOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNetworkPruneOption(v)
	})
}

func (g Grouped) applyEngineNetworkConnectOption(v *engineNetworkConnectAct) error {
	return g.apply("engineNetworkConnect", func(o any) error {
		opt, ok := o.(EngineNetworkConnectOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNetworkConnectOption(v)
	})
}

func (g Grouped) applyEngineNetworkDisconnectOption(v *engineNetworkDisconnectAct) error {
	return g.apply("engineNetworkDisconnect", func(o any) error {
		opt, ok := o.(EngineNetworkDisconnectOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNetworkDisconnectOption(v)
	})
}

func (g Grouped) applyEngineVolumeListOption(v *engineVolumeListAct) error {
	return g.apply("engineVolumeList", func(o any) error {
		opt, ok := o.(EngineVolumeListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineVolumeListOption(v)
	})
}

func (g Grouped) applyEngineVolumeRemoveOption(v *engineVolumeRemoveAct) error {
	return g.apply("engineVolumeRemove", func(o any) error {
		opt, ok := o.(EngineVolumeRemoveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineVolumeRemoveOption(v)
	})
}

func (g Grouped) applyEngineVolumePruneOption(v *engineVolumePruneAct) error {
	return g.apply("engineVolumePrune", func(o any) error {
		opt, ok := o.(EngineVolumePruneOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineVolumePruneOption(v)
	})
}

func (g Grouped) applyEnginePingOption(v *enginePingAct) error {
	return g.apply("enginePing", func(o any) error {
		opt, ok := o.(EnginePingOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePingOption(v)
	})
}

func (g Grouped) applyEngineEventsOption(v *engineEventsAct) error {
	return g.apply("engineEvents", func(o any) error {
		opt, ok := o.(EngineEventsOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineEventsOption(v)
	})
}

func (g Grouped) applyEngineDiskUsageOption(v *engineDiskUsageAct) error {
	return g.apply("engineDiskUsage", func(o any) error {
		opt, ok := o.(EngineDiskUsageOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineDiskUsageOption(v)
	})
}

func (g Grouped) applyEngineLoginOption(v *engineLoginAct) error {
	return g.apply("engineLogin", func(o any) error {
		opt, ok := o.(EngineLoginOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineLoginOption(v)
	})
}

func (g Grouped) applyEngineSwarmInitOption(v *engineSwarmInitAct) error {
	return g.apply("engineSwarmInit", func(o any) error {
		opt, ok := o.(EngineSwarmInitOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineSwarmInitOption(v)
	})
}

func (g Grouped) applyEngineSwarmJoinOption(v *engineSwarmJoinAct) error {
	return g.apply("engineSwarmJoin", func(o any) error {
		opt, ok := o.(EngineSwarmJoinOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineSwarmJoinOption(v)
	})
}

func (g Grouped) applyEngineSwarmLeaveOption(v *engineSwarmLeaveAct) error {
	return g.apply("engineSwarmLeave", func(o any) error {
		opt, ok := o.(EngineSwarmLeaveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineSwarmLeaveOption(v)
	})
}

func (g Grouped) applyEngineSwarmUnlockOption(v *engineSwarmUnlockAct) error {
	return g.apply("engineSwarmUnlock", func(o any) error {
		opt, ok := o.(EngineSwarmUnlockOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineSwarmUnlockOption(v)
	})
}

func (g Grouped) applyEngineSwarmUpdateOption(v *engineSwarmUpdateAct) error {
	return g.apply("engineSwarmUpdate", func(o any) error {
		opt, ok := o.(EngineSwarmUpdateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineSwarmUpdateOption(v)
	})
}

func (g Grouped) applyEnginePluginListOption(v *enginePluginListAct) error {
	return g.apply("enginePluginList", func(o any) error {
		opt, ok := o.(EnginePluginListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginListOption(v)
	})
}

func (g Grouped) applyEnginePluginRemoveOption(v *enginePluginRemoveAct) error {
	return g.apply("enginePluginRemove", func(o any) error {
		opt, ok := o.(EnginePluginRemoveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginRemoveOption(v)
	})
}

func (g Grouped) applyEnginePluginEnableOption(v *enginePluginEnableAct) error {
	return g.apply("enginePluginEnable", func(o any) error {
		opt, ok := o.(EnginePluginEnableOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginEnableOption(v)
	})
}

func (g Grouped) applyEnginePluginDisableOption(v *enginePluginDisableAct) error {
	return g.apply("enginePluginDisable", func(o any) error {
		opt, ok := o.(EnginePluginDisableOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginDisableOption(v)
	})
}

func (g Grouped) applyEngineNodeListOption(v *engineNodeListAct) error {
	return g.apply("engineNodeList", func(o any) error {
		opt, ok := o.(EngineNodeListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNodeListOption(v)
	})
}

func (g Grouped) applyEngineNodeRemoveOption(v *engineNodeRemoveAct) error {
	return g.apply("engineNodeRemove", func(o any) error {
		opt, ok := o.(EngineNodeRemoveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNodeRemoveOption(v)
	})
}

func (g Grouped) applyEngineServiceListOption(v *engineServiceListAct) error {
	return g.apply("engineServiceList", func(o any) error {
		opt, ok := o.(EngineServiceListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineServiceListOption(v)
	})
}

func (g Grouped) applyEngineTaskListOption(v *engineTaskListAct) error {
	return g.apply("engineTaskList", func(o any) error {
		opt, ok := o.(EngineTaskListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineTaskListOption(v)
	})
}

func (g Grouped) applyEngineSecretListOption(v *engineSecretListAct) error {
	return g.apply("engineSecretList", func(o any) error {
		opt, ok := o.(EngineSecretListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineSecretListOption(v)
	})
}

func (g Grouped) applyEngineConfigListOption(v *engineConfigListAct) error {
	return g.apply("engineConfigList", func(o any) error {
		opt, ok := o.(EngineConfigListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineConfigListOption(v)
	})
}

func (g Grouped) applyEngineCheckpointListOption(v *engineCheckpointListAct) error {
	return g.apply("engineCheckpointList", func(o any) error {
		opt, ok := o.(EngineCheckpointListOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineCheckpointListOption(v)
	})
}

func (g Grouped) applyEngineCheckpointCreateOption(v *engineCheckpointCreateAct) error {
	return g.apply("engineCheckpointCreate", func(o any) error {
		opt, ok := o.(EngineCheckpointCreateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineCheckpointCreateOption(v)
	})
}

func (g Grouped) applyEngineCheckpointRemoveOption(v *engineCheckpointRemoveAct) error {
	return g.apply("engineCheckpointRemove", func(o any) error {
		opt, ok := o.(EngineCheckpointRemoveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineCheckpointRemoveOption(v)
	})
}

func (g Grouped) applyEnginePluginCreateOption(v *enginePluginCreateAct) error {
	return g.apply("enginePluginCreate", func(o any) error {
		opt, ok := o.(EnginePluginCreateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginCreateOption(v)
	})
}

func (g Grouped) applyEnginePluginInstallOption(v *enginePluginInstallAct) error {
	return g.apply("enginePluginInstall", func(o any) error {
		opt, ok := o.(EnginePluginInstallOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginInstallOption(v)
	})
}

func (g Grouped) applyEnginePluginPushOption(v *enginePluginPushAct) error {
	return g.apply("enginePluginPush", func(o any) error {
		opt, ok := o.(EnginePluginPushOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginPushOption(v)
	})
}

func (g Grouped) applyEnginePluginSetOption(v *enginePluginSetAct) error {
	return g.apply("enginePluginSet", func(o any) error {
		opt, ok := o.(EnginePluginSetOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginSetOption(v)
	})
}

func (g Grouped) applyEnginePluginUpgradeOption(v *enginePluginUpgradeAct) error {
	return g.apply("enginePluginUpgrade", func(o any) error {
		opt, ok := o.(EnginePluginUpgradeOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEnginePluginUpgradeOption(v)
	})
}

func (g Grouped) applyEngineExecResizeOption(v *engineExecResizeAct) error {
	return g.apply("engineExecResize", func(o any) error {
		opt, ok := o.(EngineExecResizeOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineExecResizeOption(v)
	})
}

func (g Grouped) applyEngineServiceLogsOption(v *engineServiceLogsAct) error {
	return g.apply("engineServiceLogs", func(o any) error {
		opt, ok := o.(EngineServiceLogsOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineServiceLogsOption(v)
	})
}

func (g Grouped) applyEngineTaskLogsOption(v *engineTaskLogsAct) error {
	return g.apply("engineTaskLogs", func(o any) error {
		opt, ok := o.(EngineTaskLogsOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineTaskLogsOption(v)
	})
}

func (g Grouped) applyEngineServiceCreateOption(v *engineServiceCreateAct) error {
	return g.apply("engineServiceCreate", func(o any) error {
		opt, ok := o.(EngineServiceCreateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineServiceCreateOption(v)
	})
}

func (g Grouped) applyEngineServiceUpdateOption(v *engineServiceUpdateAct) error {
	return g.apply("engineServiceUpdate", func(o any) error {
		opt, ok := o.(EngineServiceUpdateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineServiceUpdateOption(v)
	})
}

func (g Grouped) applyEngineSecretCreateOption(v *engineSecretCreateAct) error {
	return g.apply("engineSecretCreate", func(o any) error {
		opt, ok := o.(EngineSecretCreateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineSecretCreateOption(v)
	})
}

func (g Grouped) applyEngineSecretUpdateOption(v *engineSecretUpdateAct) error {
	return g.apply("engineSecretUpdate", func(o any) error {
		opt, ok := o.(EngineSecretUpdateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineSecretUpdateOption(v)
	})
}

func (g Grouped) applyEngineConfigCreateOption(v *engineConfigCreateAct) error {
	return g.apply("engineConfigCreate", func(o any) error {
		opt, ok := o.(EngineConfigCreateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineConfigCreateOption(v)
	})
}

func (g Grouped) applyEngineConfigUpdateOption(v *engineConfigUpdateAct) error {
	return g.apply("engineConfigUpdate", func(o any) error {
		opt, ok := o.(EngineConfigUpdateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineConfigUpdateOption(v)
	})
}

func (g Grouped) applyEngineNodeUpdateOption(v *engineNodeUpdateAct) error {
	return g.apply("engineNodeUpdate", func(o any) error {
		opt, ok := o.(EngineNodeUpdateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineNodeUpdateOption(v)
	})
}

func (g Grouped) applyEngineVolumeUpdateOption(v *engineVolumeUpdateAct) error {
	return g.apply("engineVolumeUpdate", func(o any) error {
		opt, ok := o.(EngineVolumeUpdateOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineVolumeUpdateOption(v)
	})
}

func (g Grouped) applyEngineImageInspectOption(v *engineImageInspectAct) error {
	return g.apply("engineImageInspect", func(o any) error {
		opt, ok := o.(EngineImageInspectOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageInspectOption(v)
	})
}

func (g Grouped) applyEngineImageHistoryOption(v *engineImageHistoryAct) error {
	return g.apply("engineImageHistory", func(o any) error {
		opt, ok := o.(EngineImageHistoryOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageHistoryOption(v)
	})
}

func (g Grouped) applyEngineImageLoadOption(v *engineImageLoadAct) error {
	return g.apply("engineImageLoad", func(o any) error {
		opt, ok := o.(EngineImageLoadOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageLoadOption(v)
	})
}

func (g Grouped) applyEngineImageSaveOption(v *engineImageSaveAct) error {
	return g.apply("engineImageSave", func(o any) error {
		opt, ok := o.(EngineImageSaveOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageSaveOption(v)
	})
}

func (g Grouped) applyEngineImageAttestOption(v *engineImageAttestAct) error {
	return g.apply("engineImageAttest", func(o any) error {
		opt, ok := o.(EngineImageAttestOption)
		if !ok {
			return errLogicSkip
		}
		return opt.applyEngineImageAttestOption(v)
	})
}

var (
	_ ContainerOption               = Grouped{}
	_ ServiceOption                 = Grouped{}
	_ JobOption                     = Grouped{}
	_ ComposeOption                 = Grouped{}
	_ PortOption                    = Grouped{}
	_ HealthOption                  = Grouped{}
	_ RestartOption                 = Grouped{}
	_ DependOption                  = Grouped{}
	_ VolumeOption                  = Grouped{}
	_ NetworkOption                 = Grouped{}
	_ LogOption                     = Grouped{}
	_ DeviceOption                  = Grouped{}
	_ GPUOption                     = Grouped{}
	_ UlimitOption                  = Grouped{}
	_ BuildOption                   = Grouped{}
	_ SecretOption                  = Grouped{}
	_ FileConfigOption              = Grouped{}
	_ DeployOption                  = Grouped{}
	_ DevelopOption                 = Grouped{}
	_ WatchTriggerOption            = Grouped{}
	_ ModelOption                   = Grouped{}
	_ UpdateOption                  = Grouped{}
	_ PlacementOption               = Grouped{}
	_ ReserveOption                 = Grouped{}
	_ ComposeNetOption              = Grouped{}
	_ ComposeVolOption              = Grouped{}
	_ ComposeSecretOption           = Grouped{}
	_ ComposeFileOption             = Grouped{}
	_ ComposeModelOption            = Grouped{}
	_ IncludeOption                 = Grouped{}
	_ HookOption                    = Grouped{}
	_ EnvFileOption                 = Grouped{}
	_ ImportOption                  = Grouped{}
	_ LoggerOption                  = Grouped{}
	_ UpOption                      = Grouped{}
	_ DownOption                    = Grouped{}
	_ ComposeBuildOption            = Grouped{}
	_ ComposeCreateOption           = Grouped{}
	_ ComposeStartOption            = Grouped{}
	_ ComposeStopOption             = Grouped{}
	_ ComposeRestartOption          = Grouped{}
	_ PullOption                    = Grouped{}
	_ PushOption                    = Grouped{}
	_ LogsOption                    = Grouped{}
	_ PsOption                      = Grouped{}
	_ KillOption                    = Grouped{}
	_ RemoveOption                  = Grouped{}
	_ RunOption                     = Grouped{}
	_ PauseOption                   = Grouped{}
	_ EventsOption                  = Grouped{}
	_ ComposePortOption             = Grouped{}
	_ WatchOption                   = Grouped{}
	_ WaitOption                    = Grouped{}
	_ ScaleOption                   = Grouped{}
	_ CopyOption                    = Grouped{}
	_ AttachOption                  = Grouped{}
	_ PublishOption                 = Grouped{}
	_ ImagesOption                  = Grouped{}
	_ ExportOption                  = Grouped{}
	_ CommitOption                  = Grouped{}
	_ VizOption                     = Grouped{}
	_ ListOption                    = Grouped{}
	_ VolumesListOption             = Grouped{}
	_ EngineCreateOption            = Grouped{}
	_ EngineStartOption             = Grouped{}
	_ EngineStopOption              = Grouped{}
	_ EngineInspectOption           = Grouped{}
	_ EngineWaitOption              = Grouped{}
	_ GenerateOption                = Grouped{}
	_ LoadProjectOption             = Grouped{}
	_ EnginePullOption              = Grouped{}
	_ EngineBuildOption             = Grouped{}
	_ EngineNetworkOption           = Grouped{}
	_ EngineVolumeOption            = Grouped{}
	_ ClientOption                  = Grouped{}
	_ ComposerClientOption          = Grouped{}
	_ TLSOption                     = Grouped{}
	_ EngineContainerListOption     = Grouped{}
	_ EngineContainerPruneOption    = Grouped{}
	_ EngineRenameOption            = Grouped{}
	_ EngineResizeOption            = Grouped{}
	_ EngineAttachOption            = Grouped{}
	_ EngineCommitOption            = Grouped{}
	_ EngineStatsOption             = Grouped{}
	_ EngineTopOption               = Grouped{}
	_ EngineStatPathOption          = Grouped{}
	_ EngineCopyFromOption          = Grouped{}
	_ EngineCopyToOption            = Grouped{}
	_ EngineUpdateOption            = Grouped{}
	_ EngineExecCreateOption        = Grouped{}
	_ EngineExecStartOption         = Grouped{}
	_ EngineExecAttachOption        = Grouped{}
	_ EngineImageListOption         = Grouped{}
	_ EngineImagePushOption         = Grouped{}
	_ EngineImageRemoveOption       = Grouped{}
	_ EngineImageTagOption          = Grouped{}
	_ EngineImagePruneOption        = Grouped{}
	_ EngineImageSearchOption       = Grouped{}
	_ EngineImageImportOption       = Grouped{}
	_ EngineBuildCachePruneOption   = Grouped{}
	_ EngineNetworkInspectOption    = Grouped{}
	_ EngineNetworkListOption       = Grouped{}
	_ EngineNetworkPruneOption      = Grouped{}
	_ EngineNetworkConnectOption    = Grouped{}
	_ EngineNetworkDisconnectOption = Grouped{}
	_ EngineVolumeListOption        = Grouped{}
	_ EngineVolumeRemoveOption      = Grouped{}
	_ EngineVolumePruneOption       = Grouped{}
	_ EnginePingOption              = Grouped{}
	_ EngineEventsOption            = Grouped{}
	_ EngineDiskUsageOption         = Grouped{}
	_ EngineLoginOption             = Grouped{}
	_ EngineSwarmInitOption         = Grouped{}
	_ EngineSwarmJoinOption         = Grouped{}
	_ EngineSwarmLeaveOption        = Grouped{}
	_ EngineSwarmUnlockOption       = Grouped{}
	_ EngineSwarmUpdateOption       = Grouped{}
	_ EnginePluginListOption        = Grouped{}
	_ EnginePluginRemoveOption      = Grouped{}
	_ EnginePluginEnableOption      = Grouped{}
	_ EnginePluginDisableOption     = Grouped{}
	_ EngineNodeListOption          = Grouped{}
	_ EngineNodeRemoveOption        = Grouped{}
	_ EngineServiceListOption       = Grouped{}
	_ EngineTaskListOption          = Grouped{}
	_ EngineSecretListOption        = Grouped{}
	_ EngineConfigListOption        = Grouped{}
	_ EngineCheckpointListOption    = Grouped{}
	_ EngineCheckpointCreateOption  = Grouped{}
	_ EngineCheckpointRemoveOption  = Grouped{}
	_ EnginePluginCreateOption      = Grouped{}
	_ EnginePluginInstallOption     = Grouped{}
	_ EnginePluginPushOption        = Grouped{}
	_ EnginePluginSetOption         = Grouped{}
	_ EnginePluginUpgradeOption     = Grouped{}
	_ EngineExecResizeOption        = Grouped{}
	_ EngineServiceLogsOption       = Grouped{}
	_ EngineTaskLogsOption          = Grouped{}
	_ EngineServiceCreateOption     = Grouped{}
	_ EngineServiceUpdateOption     = Grouped{}
	_ EngineSecretCreateOption      = Grouped{}
	_ EngineSecretUpdateOption      = Grouped{}
	_ EngineConfigCreateOption      = Grouped{}
	_ EngineConfigUpdateOption      = Grouped{}
	_ EngineNodeUpdateOption        = Grouped{}
	_ EngineVolumeUpdateOption      = Grouped{}
	_ EngineImageInspectOption      = Grouped{}
	_ EngineImageHistoryOption      = Grouped{}
	_ EngineImageLoadOption         = Grouped{}
	_ EngineImageSaveOption         = Grouped{}
	_ EngineImageAttestOption       = Grouped{}
)
