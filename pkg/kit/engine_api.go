package kit

import (
	"context"
	"fmt"
)

// EngineContainerListOption configures [ListContainers].
type EngineContainerListOption interface {
	applyEngineContainerListOption(*engineContainerListAct) error
}

// EngineContainerPruneOption configures [PruneContainers].
type EngineContainerPruneOption interface {
	applyEngineContainerPruneOption(*engineContainerPruneAct) error
}

// EngineRenameOption configures [RenameContainer].
type EngineRenameOption interface {
	applyEngineRenameOption(*engineRenameAct) error
}

// EngineResizeOption configures [ResizeContainer].
type EngineResizeOption interface {
	applyEngineResizeOption(*engineResizeAct) error
}

// EngineAttachOption configures [AttachContainer].
type EngineAttachOption interface {
	applyEngineAttachOption(*engineAttachAct) error
}

// EngineCommitOption configures [CommitContainer].
type EngineCommitOption interface {
	applyEngineCommitOption(*engineCommitAct) error
}

// EngineStatsOption configures [StatsContainer].
type EngineStatsOption interface {
	applyEngineStatsOption(*engineStatsAct) error
}

// EngineTopOption configures [TopContainer].
type EngineTopOption interface {
	applyEngineTopOption(*engineTopAct) error
}

// EngineStatPathOption configures [StatPath].
type EngineStatPathOption interface {
	applyEngineStatPathOption(*engineStatPathAct) error
}

// EngineCopyFromOption configures [CopyFrom].
type EngineCopyFromOption interface {
	applyEngineCopyFromOption(*engineCopyFromAct) error
}

// EngineCopyToOption configures [CopyTo].
type EngineCopyToOption interface {
	applyEngineCopyToOption(*engineCopyToAct) error
}

// EngineUpdateOption configures [UpdateContainer].
type EngineUpdateOption interface {
	applyEngineUpdateOption(*engineUpdateAct) error
}

// EngineExecCreateOption configures [CreateExec].
type EngineExecCreateOption interface {
	applyEngineExecCreateOption(*engineExecCreateAct) error
}

// EngineExecStartOption configures [StartExec].
type EngineExecStartOption interface {
	applyEngineExecStartOption(*engineExecStartAct) error
}

// EngineExecAttachOption configures [AttachExec].
type EngineExecAttachOption interface {
	applyEngineExecAttachOption(*engineExecAttachAct) error
}

// EngineImageListOption configures [ListImages].
type EngineImageListOption interface {
	applyEngineImageListOption(*engineImageListAct) error
}

// EngineImagePushOption configures [PushImage].
type EngineImagePushOption interface {
	applyEngineImagePushOption(*engineImagePushAct) error
}

// EngineImageRemoveOption configures [RemoveImage].
type EngineImageRemoveOption interface {
	applyEngineImageRemoveOption(*engineImageRemoveAct) error
}

// EngineImageTagOption configures [TagImage].
type EngineImageTagOption interface {
	applyEngineImageTagOption(*engineImageTagAct) error
}

// EngineImagePruneOption configures [PruneImages].
type EngineImagePruneOption interface {
	applyEngineImagePruneOption(*engineImagePruneAct) error
}

// EngineImageSearchOption configures [SearchImages].
type EngineImageSearchOption interface {
	applyEngineImageSearchOption(*engineImageSearchAct) error
}

// EngineImageImportOption configures [ImportImage].
type EngineImageImportOption interface {
	applyEngineImageImportOption(*engineImageImportAct) error
}

// EngineBuildCachePruneOption configures [PruneBuildCache].
type EngineBuildCachePruneOption interface {
	applyEngineBuildCachePruneOption(*engineBuildCachePruneAct) error
}

// EngineNetworkInspectOption configures [InspectNetwork].
type EngineNetworkInspectOption interface {
	applyEngineNetworkInspectOption(*engineNetworkInspectAct) error
}

// EngineNetworkListOption configures [ListNetworks].
type EngineNetworkListOption interface {
	applyEngineNetworkListOption(*engineNetworkListAct) error
}

// EngineNetworkPruneOption configures [PruneNetworks].
type EngineNetworkPruneOption interface {
	applyEngineNetworkPruneOption(*engineNetworkPruneAct) error
}

// EngineNetworkConnectOption configures [ConnectNetwork].
type EngineNetworkConnectOption interface {
	applyEngineNetworkConnectOption(*engineNetworkConnectAct) error
}

// EngineNetworkDisconnectOption configures [DisconnectNetwork].
type EngineNetworkDisconnectOption interface {
	applyEngineNetworkDisconnectOption(*engineNetworkDisconnectAct) error
}

// EngineVolumeListOption configures [ListEngineVolumes].
type EngineVolumeListOption interface {
	applyEngineVolumeListOption(*engineVolumeListAct) error
}

// EngineVolumeRemoveOption configures [RemoveVolume].
type EngineVolumeRemoveOption interface {
	applyEngineVolumeRemoveOption(*engineVolumeRemoveAct) error
}

// EngineVolumePruneOption configures [PruneVolumes].
type EngineVolumePruneOption interface {
	applyEngineVolumePruneOption(*engineVolumePruneAct) error
}

// EnginePingOption configures [Ping].
type EnginePingOption interface {
	applyEnginePingOption(*enginePingAct) error
}

// EngineEventsOption configures [ListenEvents].
type EngineEventsOption interface {
	applyEngineEventsOption(*engineEventsAct) error
}

// EngineDiskUsageOption configures [DiskUsage].
type EngineDiskUsageOption interface {
	applyEngineDiskUsageOption(*engineDiskUsageAct) error
}

// EngineLoginOption configures [Login].
type EngineLoginOption interface {
	applyEngineLoginOption(*engineLoginAct) error
}

// EngineSwarmInitOption configures [InitSwarm].
type EngineSwarmInitOption interface {
	applyEngineSwarmInitOption(*engineSwarmInitAct) error
}

// EngineSwarmJoinOption configures [JoinSwarm].
type EngineSwarmJoinOption interface {
	applyEngineSwarmJoinOption(*engineSwarmJoinAct) error
}

// EngineSwarmLeaveOption configures [LeaveSwarm].
type EngineSwarmLeaveOption interface {
	applyEngineSwarmLeaveOption(*engineSwarmLeaveAct) error
}

// EngineSwarmUnlockOption configures [UnlockSwarm].
type EngineSwarmUnlockOption interface {
	applyEngineSwarmUnlockOption(*engineSwarmUnlockAct) error
}

// EngineSwarmUpdateOption configures [UpdateSwarm].
type EngineSwarmUpdateOption interface {
	applyEngineSwarmUpdateOption(*engineSwarmUpdateAct) error
}

// EnginePluginListOption configures [ListPlugins].
type EnginePluginListOption interface {
	applyEnginePluginListOption(*enginePluginListAct) error
}

// EnginePluginRemoveOption configures [RemovePlugin].
type EnginePluginRemoveOption interface {
	applyEnginePluginRemoveOption(*enginePluginRemoveAct) error
}

// EnginePluginEnableOption configures [EnablePlugin].
type EnginePluginEnableOption interface {
	applyEnginePluginEnableOption(*enginePluginEnableAct) error
}

// EnginePluginDisableOption configures [DisablePlugin].
type EnginePluginDisableOption interface {
	applyEnginePluginDisableOption(*enginePluginDisableAct) error
}

// EngineNodeListOption configures [ListNodes].
type EngineNodeListOption interface {
	applyEngineNodeListOption(*engineNodeListAct) error
}

// EngineNodeRemoveOption configures [RemoveNode].
type EngineNodeRemoveOption interface {
	applyEngineNodeRemoveOption(*engineNodeRemoveAct) error
}

// EngineServiceListOption configures [ListSwarmServices].
type EngineServiceListOption interface {
	applyEngineServiceListOption(*engineServiceListAct) error
}

// EngineTaskListOption configures [ListTasks].
type EngineTaskListOption interface {
	applyEngineTaskListOption(*engineTaskListAct) error
}

// EngineSecretListOption configures [ListSecrets].
type EngineSecretListOption interface {
	applyEngineSecretListOption(*engineSecretListAct) error
}

// EngineConfigListOption configures [ListConfigs].
type EngineConfigListOption interface {
	applyEngineConfigListOption(*engineConfigListAct) error
}

// EngineCheckpointListOption configures [ListCheckpoints].
type EngineCheckpointListOption interface {
	applyEngineCheckpointListOption(*engineCheckpointListAct) error
}

// EngineCheckpointCreateOption configures [CreateCheckpoint].
type EngineCheckpointCreateOption interface {
	applyEngineCheckpointCreateOption(*engineCheckpointCreateAct) error
}

// EngineCheckpointRemoveOption configures [RemoveCheckpoint].
type EngineCheckpointRemoveOption interface {
	applyEngineCheckpointRemoveOption(*engineCheckpointRemoveAct) error
}

type engineContainerListAct struct{ opts ContainerListOptions }
type engineContainerPruneAct struct{ opts ContainerPruneOptions }
type engineRenameAct struct{ opts ContainerRenameOptions }
type engineResizeAct struct{ opts ContainerResizeOptions }
type engineAttachAct struct{ opts ContainerAttachOptions }
type engineCommitAct struct{ opts ContainerCommitOptions }
type engineStatsAct struct{ opts ContainerStatsOptions }
type engineTopAct struct{ opts ContainerTopOptions }
type engineStatPathAct struct{ opts ContainerStatPathOptions }
type engineCopyFromAct struct{ opts CopyFromContainerOptions }
type engineCopyToAct struct{ opts CopyToContainerOptions }
type engineUpdateAct struct{ opts ContainerUpdateOptions }
type engineExecCreateAct struct{ opts ExecCreateOptions }
type engineExecStartAct struct{ opts ExecStartOptions }
type engineExecAttachAct struct{ opts ExecAttachOptions }
type engineImageListAct struct{ opts ImageListOptions }
type engineImagePushAct struct {
	opts ImagePushOptions
	log  *Logger
}
type engineImageRemoveAct struct{ opts ImageRemoveOptions }
type engineImageTagAct struct{ opts ImageTagOptions }
type engineImagePruneAct struct{ opts ImagePruneOptions }
type engineImageSearchAct struct{ opts ImageSearchOptions }
type engineImageImportAct struct{ opts ImageImportOptions }
type engineBuildCachePruneAct struct{ opts BuildCachePruneOptions }
type engineNetworkInspectAct struct{ opts NetworkInspectOptions }
type engineNetworkListAct struct{ opts NetworkListOptions }
type engineNetworkPruneAct struct{ opts NetworkPruneOptions }
type engineNetworkConnectAct struct{ opts NetworkConnectOptions }
type engineNetworkDisconnectAct struct{ opts NetworkDisconnectOptions }
type engineVolumeListAct struct{ opts VolumeListOptions }
type engineVolumeRemoveAct struct{ opts VolumeRemoveOptions }
type engineVolumePruneAct struct{ opts VolumePruneOptions }
type enginePingAct struct{ opts PingOptions }
type engineEventsAct struct{ opts EventsListOptions }
type engineDiskUsageAct struct{ opts DiskUsageOptions }
type engineLoginAct struct{ opts RegistryLoginOptions }
type engineSwarmInitAct struct{ opts SwarmInitOptions }
type engineSwarmJoinAct struct{ opts SwarmJoinOptions }
type engineSwarmLeaveAct struct{ opts SwarmLeaveOptions }
type engineSwarmUnlockAct struct{ opts SwarmUnlockOptions }
type engineSwarmUpdateAct struct{ opts SwarmUpdateOptions }
type enginePluginListAct struct{ opts PluginListOptions }
type enginePluginRemoveAct struct{ opts PluginRemoveOptions }
type enginePluginEnableAct struct{ opts PluginEnableOptions }
type enginePluginDisableAct struct{ opts PluginDisableOptions }
type engineNodeListAct struct{ opts NodeListOptions }
type engineNodeRemoveAct struct{ opts NodeRemoveOptions }
type engineServiceListAct struct{ opts ServiceListOptions }
type engineTaskListAct struct{ opts TaskListOptions }
type engineSecretListAct struct{ opts SecretListOptions }
type engineConfigListAct struct{ opts ConfigListOptions }
type engineCheckpointListAct struct{ opts CheckpointListOptions }
type engineCheckpointCreateAct struct{ opts CheckpointCreateOptions }
type engineCheckpointRemoveAct struct{ opts CheckpointRemoveOptions }

// ListContainers wraps the corresponding Engine client method.
func ListContainers(ctx context.Context, api Engine, opts ...EngineContainerListOption) (ContainerListResult, error) {
	if err := engineAPI(api, "ListContainers"); err != nil {
		return ContainerListResult{}, err
	}
	var act engineContainerListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listcontainers.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineContainerListOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerList(ctx, act.opts)
	if err != nil {
		return ContainerListResult{}, cfgErr("engine", "listcontainers", "ListContainers", err)
	}
	return out, nil
}

// PruneContainers wraps the corresponding Engine client method.
func PruneContainers(ctx context.Context, api Engine, opts ...EngineContainerPruneOption) (ContainerPruneResult, error) {
	if err := engineAPI(api, "PruneContainers"); err != nil {
		return ContainerPruneResult{}, err
	}
	var act engineContainerPruneAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("prunecontainers.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineContainerPruneOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerPruneResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerPrune(ctx, act.opts)
	if err != nil {
		return ContainerPruneResult{}, cfgErr("engine", "prunecontainers", "PruneContainers", err)
	}
	return out, nil
}

// RenameContainer wraps the corresponding Engine client method.
func RenameContainer(ctx context.Context, api Engine, id string, opts ...EngineRenameOption) (ContainerRenameResult, error) {
	if err := engineAPI(api, "RenameContainer"); err != nil {
		return ContainerRenameResult{}, err
	}
	if id == "" {
		return ContainerRenameResult{}, cfgErr("engine", "id", "RenameContainer", fmt.Errorf("%w: id", errRequired))
	}
	var act engineRenameAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("renamecontainer.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineRenameOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerRenameResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerRename(ctx, id, act.opts)
	if err != nil {
		return ContainerRenameResult{}, cfgErr("engine", "renamecontainer", "RenameContainer", err)
	}
	return out, nil
}

// ResizeContainer wraps the corresponding Engine client method.
func ResizeContainer(ctx context.Context, api Engine, id string, opts ...EngineResizeOption) (ContainerResizeResult, error) {
	if err := engineAPI(api, "ResizeContainer"); err != nil {
		return ContainerResizeResult{}, err
	}
	if id == "" {
		return ContainerResizeResult{}, cfgErr("engine", "id", "ResizeContainer", fmt.Errorf("%w: id", errRequired))
	}
	var act engineResizeAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("resizecontainer.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineResizeOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerResizeResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerResize(ctx, id, act.opts)
	if err != nil {
		return ContainerResizeResult{}, cfgErr("engine", "resizecontainer", "ResizeContainer", err)
	}
	return out, nil
}

// AttachContainer wraps the corresponding Engine client method.
func AttachContainer(ctx context.Context, api Engine, id string, opts ...EngineAttachOption) (ContainerAttachResult, error) {
	if err := engineAPI(api, "AttachContainer"); err != nil {
		return ContainerAttachResult{}, err
	}
	if id == "" {
		return ContainerAttachResult{}, cfgErr("engine", "id", "AttachContainer", fmt.Errorf("%w: id", errRequired))
	}
	var act engineAttachAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("attachcontainer.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineAttachOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerAttachResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerAttach(ctx, id, act.opts)
	if err != nil {
		return ContainerAttachResult{}, cfgErr("engine", "attachcontainer", "AttachContainer", err)
	}
	return out, nil
}

// CommitContainer wraps the corresponding Engine client method.
func CommitContainer(ctx context.Context, api Engine, id string, opts ...EngineCommitOption) (ContainerCommitResult, error) {
	if err := engineAPI(api, "CommitContainer"); err != nil {
		return ContainerCommitResult{}, err
	}
	if id == "" {
		return ContainerCommitResult{}, cfgErr("engine", "id", "CommitContainer", fmt.Errorf("%w: id", errRequired))
	}
	var act engineCommitAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("commitcontainer.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineCommitOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerCommitResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerCommit(ctx, id, act.opts)
	if err != nil {
		return ContainerCommitResult{}, cfgErr("engine", "commitcontainer", "CommitContainer", err)
	}
	return out, nil
}

// DiffContainer wraps the corresponding Engine client method.
func DiffContainer(ctx context.Context, api Engine, id string) (ContainerDiffResult, error) {
	if err := engineAPI(api, "DiffContainer"); err != nil {
		return ContainerDiffResult{}, err
	}
	if id == "" {
		return ContainerDiffResult{}, cfgErr("engine", "id", "DiffContainer", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.ContainerDiff(ctx, id, ContainerDiffOptions{})
	if err != nil {
		return ContainerDiffResult{}, cfgErr("engine", "diffcontainer", "DiffContainer", err)
	}
	return out, nil
}

// ExportContainer wraps the corresponding Engine client method.
func ExportContainer(ctx context.Context, api Engine, id string) (ContainerExportResult, error) {
	if err := engineAPI(api, "ExportContainer"); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, cfgErr("engine", "id", "ExportContainer", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.ContainerExport(ctx, id, ContainerExportOptions{})
	if err != nil {
		return nil, cfgErr("engine", "exportcontainer", "ExportContainer", err)
	}
	return out, nil
}

// StatsContainer wraps the corresponding Engine client method.
func StatsContainer(ctx context.Context, api Engine, id string, opts ...EngineStatsOption) (ContainerStatsResult, error) {
	if err := engineAPI(api, "StatsContainer"); err != nil {
		return ContainerStatsResult{}, err
	}
	if id == "" {
		return ContainerStatsResult{}, cfgErr("engine", "id", "StatsContainer", fmt.Errorf("%w: id", errRequired))
	}
	var act engineStatsAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("statscontainer.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineStatsOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerStatsResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerStats(ctx, id, act.opts)
	if err != nil {
		return ContainerStatsResult{}, cfgErr("engine", "statscontainer", "StatsContainer", err)
	}
	return out, nil
}

// TopContainer wraps the corresponding Engine client method.
func TopContainer(ctx context.Context, api Engine, id string, opts ...EngineTopOption) (ContainerTopResult, error) {
	if err := engineAPI(api, "TopContainer"); err != nil {
		return ContainerTopResult{}, err
	}
	if id == "" {
		return ContainerTopResult{}, cfgErr("engine", "id", "TopContainer", fmt.Errorf("%w: id", errRequired))
	}
	var act engineTopAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("topcontainer.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineTopOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerTopResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerTop(ctx, id, act.opts)
	if err != nil {
		return ContainerTopResult{}, cfgErr("engine", "topcontainer", "TopContainer", err)
	}
	return out, nil
}

// StatPath wraps the corresponding Engine client method.
func StatPath(ctx context.Context, api Engine, id string, opts ...EngineStatPathOption) (ContainerStatPathResult, error) {
	if err := engineAPI(api, "StatPath"); err != nil {
		return ContainerStatPathResult{}, err
	}
	if id == "" {
		return ContainerStatPathResult{}, cfgErr("engine", "id", "StatPath", fmt.Errorf("%w: id", errRequired))
	}
	var act engineStatPathAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("statpath.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineStatPathOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerStatPathResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerStatPath(ctx, id, act.opts)
	if err != nil {
		return ContainerStatPathResult{}, cfgErr("engine", "statpath", "StatPath", err)
	}
	return out, nil
}

// CopyFrom wraps the corresponding Engine client method.
func CopyFrom(ctx context.Context, api Engine, id string, opts ...EngineCopyFromOption) (CopyFromContainerResult, error) {
	if err := engineAPI(api, "CopyFrom"); err != nil {
		return CopyFromContainerResult{}, err
	}
	if id == "" {
		return CopyFromContainerResult{}, cfgErr("engine", "id", "CopyFrom", fmt.Errorf("%w: id", errRequired))
	}
	var act engineCopyFromAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("copyfrom.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineCopyFromOption(&act))
	}
	if err := join(errs...); err != nil {
		return CopyFromContainerResult{}, wrapReceiver("engine", err)
	}
	out, err := api.CopyFromContainer(ctx, id, act.opts)
	if err != nil {
		return CopyFromContainerResult{}, cfgErr("engine", "copyfrom", "CopyFrom", err)
	}
	return out, nil
}

// CopyTo wraps the corresponding Engine client method.
func CopyTo(ctx context.Context, api Engine, id string, opts ...EngineCopyToOption) (CopyToContainerResult, error) {
	if err := engineAPI(api, "CopyTo"); err != nil {
		return CopyToContainerResult{}, err
	}
	if id == "" {
		return CopyToContainerResult{}, cfgErr("engine", "id", "CopyTo", fmt.Errorf("%w: id", errRequired))
	}
	var act engineCopyToAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("copyto.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineCopyToOption(&act))
	}
	if err := join(errs...); err != nil {
		return CopyToContainerResult{}, wrapReceiver("engine", err)
	}
	out, err := api.CopyToContainer(ctx, id, act.opts)
	if err != nil {
		return CopyToContainerResult{}, cfgErr("engine", "copyto", "CopyTo", err)
	}
	return out, nil
}

// UpdateContainer wraps the corresponding Engine client method.
func UpdateContainer(ctx context.Context, api Engine, id string, opts ...EngineUpdateOption) (ContainerUpdateResult, error) {
	if err := engineAPI(api, "UpdateContainer"); err != nil {
		return ContainerUpdateResult{}, err
	}
	if id == "" {
		return ContainerUpdateResult{}, cfgErr("engine", "id", "UpdateContainer", fmt.Errorf("%w: id", errRequired))
	}
	var act engineUpdateAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("updatecontainer.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineUpdateOption(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerUpdateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ContainerUpdate(ctx, id, act.opts)
	if err != nil {
		return ContainerUpdateResult{}, cfgErr("engine", "updatecontainer", "UpdateContainer", err)
	}
	return out, nil
}

// CreateExec wraps the corresponding Engine client method.
func CreateExec(ctx context.Context, api Engine, id string, opts ...EngineExecCreateOption) (ExecCreateResult, error) {
	if err := engineAPI(api, "CreateExec"); err != nil {
		return ExecCreateResult{}, err
	}
	if id == "" {
		return ExecCreateResult{}, cfgErr("engine", "id", "CreateExec", fmt.Errorf("%w: id", errRequired))
	}
	var act engineExecCreateAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("createexec.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineExecCreateOption(&act))
	}
	if err := join(errs...); err != nil {
		return ExecCreateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ExecCreate(ctx, id, act.opts)
	if err != nil {
		return ExecCreateResult{}, cfgErr("engine", "createexec", "CreateExec", err)
	}
	return out, nil
}

// InspectExec wraps the corresponding Engine client method.
func InspectExec(ctx context.Context, api Engine, id string) (ExecInspectResult, error) {
	if err := engineAPI(api, "InspectExec"); err != nil {
		return ExecInspectResult{}, err
	}
	if id == "" {
		return ExecInspectResult{}, cfgErr("engine", "id", "InspectExec", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.ExecInspect(ctx, id, ExecInspectOptions{})
	if err != nil {
		return ExecInspectResult{}, cfgErr("engine", "inspectexec", "InspectExec", err)
	}
	return out, nil
}

// StartExec wraps the corresponding Engine client method.
func StartExec(ctx context.Context, api Engine, id string, opts ...EngineExecStartOption) (ExecStartResult, error) {
	if err := engineAPI(api, "StartExec"); err != nil {
		return ExecStartResult{}, err
	}
	if id == "" {
		return ExecStartResult{}, cfgErr("engine", "id", "StartExec", fmt.Errorf("%w: id", errRequired))
	}
	var act engineExecStartAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("startexec.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineExecStartOption(&act))
	}
	if err := join(errs...); err != nil {
		return ExecStartResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ExecStart(ctx, id, act.opts)
	if err != nil {
		return ExecStartResult{}, cfgErr("engine", "startexec", "StartExec", err)
	}
	return out, nil
}

// AttachExec wraps the corresponding Engine client method.
func AttachExec(ctx context.Context, api Engine, id string, opts ...EngineExecAttachOption) (ExecAttachResult, error) {
	if err := engineAPI(api, "AttachExec"); err != nil {
		return ExecAttachResult{}, err
	}
	if id == "" {
		return ExecAttachResult{}, cfgErr("engine", "id", "AttachExec", fmt.Errorf("%w: id", errRequired))
	}
	var act engineExecAttachAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("attachexec.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineExecAttachOption(&act))
	}
	if err := join(errs...); err != nil {
		return ExecAttachResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ExecAttach(ctx, id, act.opts)
	if err != nil {
		return ExecAttachResult{}, cfgErr("engine", "attachexec", "AttachExec", err)
	}
	return out, nil
}

// ListImages wraps the corresponding Engine client method.
func ListImages(ctx context.Context, api Engine, opts ...EngineImageListOption) (ImageListResult, error) {
	if err := engineAPI(api, "ListImages"); err != nil {
		return ImageListResult{}, err
	}
	var act engineImageListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listimages.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineImageListOption(&act))
	}
	if err := join(errs...); err != nil {
		return ImageListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImageList(ctx, act.opts)
	if err != nil {
		return ImageListResult{}, cfgErr("engine", "listimages", "ListImages", err)
	}
	return out, nil
}

// PushImage pushes an image through Engine.
// The Engine call returns a progress stream. This function reads that stream
// to the end, so the push finishes here and the body is not a result.
func PushImage(ctx context.Context, api Engine, ref string, opts ...EngineImagePushOption) error {
	if err := engineAPI(api, "PushImage"); err != nil {
		return err
	}
	if ref == "" {
		return cfgErr("engine", "ref", "PushImage", fmt.Errorf("%w: image reference", errRequired))
	}
	var act engineImagePushAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("pushimage.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineImagePushOption(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("engine", err)
	}
	adoptLog(ctx, &act.log)
	out, err := api.ImagePush(ctx, ref, act.opts)
	if err != nil {
		return cfgErr("engine", ref, "PushImage", err)
	}
	if err := act.log.drainJSON(ctx, out, KindPush, ref); err != nil {
		return cfgErr("engine", ref, "PushImage", err)
	}
	return nil
}

// RemoveImage wraps the corresponding Engine client method.
func RemoveImage(ctx context.Context, api Engine, ref string, opts ...EngineImageRemoveOption) (ImageRemoveResult, error) {
	if err := engineAPI(api, "RemoveImage"); err != nil {
		return ImageRemoveResult{}, err
	}
	if ref == "" {
		return ImageRemoveResult{}, cfgErr("engine", "ref", "RemoveImage", fmt.Errorf("%w: image reference", errRequired))
	}
	var act engineImageRemoveAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("removeimage.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineImageRemoveOption(&act))
	}
	if err := join(errs...); err != nil {
		return ImageRemoveResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImageRemove(ctx, ref, act.opts)
	if err != nil {
		return ImageRemoveResult{}, cfgErr("engine", "removeimage", "RemoveImage", err)
	}
	return out, nil
}

// TagImage wraps the corresponding Engine client method.
func TagImage(ctx context.Context, api Engine, opts ...EngineImageTagOption) (ImageTagResult, error) {
	if err := engineAPI(api, "TagImage"); err != nil {
		return ImageTagResult{}, err
	}
	var act engineImageTagAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("tagimage.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineImageTagOption(&act))
	}
	if err := join(errs...); err != nil {
		return ImageTagResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImageTag(ctx, act.opts)
	if err != nil {
		return ImageTagResult{}, cfgErr("engine", "tagimage", "TagImage", err)
	}
	return out, nil
}

// PruneImages wraps the corresponding Engine client method.
func PruneImages(ctx context.Context, api Engine, opts ...EngineImagePruneOption) (ImagePruneResult, error) {
	if err := engineAPI(api, "PruneImages"); err != nil {
		return ImagePruneResult{}, err
	}
	var act engineImagePruneAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("pruneimages.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineImagePruneOption(&act))
	}
	if err := join(errs...); err != nil {
		return ImagePruneResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImagePrune(ctx, act.opts)
	if err != nil {
		return ImagePruneResult{}, cfgErr("engine", "pruneimages", "PruneImages", err)
	}
	return out, nil
}

// SearchImages wraps the corresponding Engine client method.
func SearchImages(ctx context.Context, api Engine, term string, opts ...EngineImageSearchOption) (ImageSearchResult, error) {
	if err := engineAPI(api, "SearchImages"); err != nil {
		return ImageSearchResult{}, err
	}
	if term == "" {
		return ImageSearchResult{}, cfgErr("engine", "term", "SearchImages", fmt.Errorf("%w: search term", errRequired))
	}
	var act engineImageSearchAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("searchimages.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineImageSearchOption(&act))
	}
	if err := join(errs...); err != nil {
		return ImageSearchResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImageSearch(ctx, term, act.opts)
	if err != nil {
		return ImageSearchResult{}, cfgErr("engine", term, "SearchImages", err)
	}
	return out, nil
}

// ImportImage wraps the corresponding Engine client method.
func ImportImage(ctx context.Context, api Engine, source ImageImportSource, ref string, opts ...EngineImageImportOption) (ImageImportResult, error) {
	if err := engineAPI(api, "ImportImage"); err != nil {
		return nil, err
	}
	var act engineImageImportAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("importimage.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineImageImportOption(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("engine", err)
	}
	out, err := api.ImageImport(ctx, source, ref, act.opts)
	if err != nil {
		return nil, cfgErr("engine", "importimage", "ImportImage", err)
	}
	return out, nil
}

// PruneBuildCache wraps the corresponding Engine client method.
func PruneBuildCache(ctx context.Context, api Engine, opts ...EngineBuildCachePruneOption) (BuildCachePruneResult, error) {
	if err := engineAPI(api, "PruneBuildCache"); err != nil {
		return BuildCachePruneResult{}, err
	}
	var act engineBuildCachePruneAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("prunebuildcache.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineBuildCachePruneOption(&act))
	}
	if err := join(errs...); err != nil {
		return BuildCachePruneResult{}, wrapReceiver("engine", err)
	}
	out, err := api.BuildCachePrune(ctx, act.opts)
	if err != nil {
		return BuildCachePruneResult{}, cfgErr("engine", "prunebuildcache", "PruneBuildCache", err)
	}
	return out, nil
}

// CancelBuild wraps the corresponding Engine client method.
func CancelBuild(ctx context.Context, api Engine, id string) (BuildCancelResult, error) {
	if err := engineAPI(api, "CancelBuild"); err != nil {
		return BuildCancelResult{}, err
	}
	if id == "" {
		return BuildCancelResult{}, cfgErr("engine", "id", "CancelBuild", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.BuildCancel(ctx, id, BuildCancelOptions{})
	if err != nil {
		return BuildCancelResult{}, cfgErr("engine", "cancelbuild", "CancelBuild", err)
	}
	return out, nil
}

// InspectNetwork wraps the corresponding Engine client method.
func InspectNetwork(ctx context.Context, api Engine, name string, opts ...EngineNetworkInspectOption) (NetworkInspectResult, error) {
	if err := engineAPI(api, "InspectNetwork"); err != nil {
		return NetworkInspectResult{}, err
	}
	if name == "" {
		return NetworkInspectResult{}, cfgErr("engine", "name", "InspectNetwork", fmt.Errorf("%w: name", errRequired))
	}
	var act engineNetworkInspectAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("inspectnetwork.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineNetworkInspectOption(&act))
	}
	if err := join(errs...); err != nil {
		return NetworkInspectResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NetworkInspect(ctx, name, act.opts)
	if err != nil {
		return NetworkInspectResult{}, cfgErr("engine", "inspectnetwork", "InspectNetwork", err)
	}
	return out, nil
}

// ListNetworks wraps the corresponding Engine client method.
func ListNetworks(ctx context.Context, api Engine, opts ...EngineNetworkListOption) (NetworkListResult, error) {
	if err := engineAPI(api, "ListNetworks"); err != nil {
		return NetworkListResult{}, err
	}
	var act engineNetworkListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listnetworks.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineNetworkListOption(&act))
	}
	if err := join(errs...); err != nil {
		return NetworkListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NetworkList(ctx, act.opts)
	if err != nil {
		return NetworkListResult{}, cfgErr("engine", "listnetworks", "ListNetworks", err)
	}
	return out, nil
}

// RemoveNetwork wraps the corresponding Engine client method.
func RemoveNetwork(ctx context.Context, api Engine, name string) (NetworkRemoveResult, error) {
	if err := engineAPI(api, "RemoveNetwork"); err != nil {
		return NetworkRemoveResult{}, err
	}
	if name == "" {
		return NetworkRemoveResult{}, cfgErr("engine", "name", "RemoveNetwork", fmt.Errorf("%w: name", errRequired))
	}
	out, err := api.NetworkRemove(ctx, name, NetworkRemoveOptions{})
	if err != nil {
		return NetworkRemoveResult{}, cfgErr("engine", "removenetwork", "RemoveNetwork", err)
	}
	return out, nil
}

// PruneNetworks wraps the corresponding Engine client method.
func PruneNetworks(ctx context.Context, api Engine, opts ...EngineNetworkPruneOption) (NetworkPruneResult, error) {
	if err := engineAPI(api, "PruneNetworks"); err != nil {
		return NetworkPruneResult{}, err
	}
	var act engineNetworkPruneAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("prunenetworks.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineNetworkPruneOption(&act))
	}
	if err := join(errs...); err != nil {
		return NetworkPruneResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NetworkPrune(ctx, act.opts)
	if err != nil {
		return NetworkPruneResult{}, cfgErr("engine", "prunenetworks", "PruneNetworks", err)
	}
	return out, nil
}

// ConnectNetwork wraps the corresponding Engine client method.
func ConnectNetwork(ctx context.Context, api Engine, name string, opts ...EngineNetworkConnectOption) (NetworkConnectResult, error) {
	if err := engineAPI(api, "ConnectNetwork"); err != nil {
		return NetworkConnectResult{}, err
	}
	if name == "" {
		return NetworkConnectResult{}, cfgErr("engine", "name", "ConnectNetwork", fmt.Errorf("%w: name", errRequired))
	}
	var act engineNetworkConnectAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("connectnetwork.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineNetworkConnectOption(&act))
	}
	if err := join(errs...); err != nil {
		return NetworkConnectResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NetworkConnect(ctx, name, act.opts)
	if err != nil {
		return NetworkConnectResult{}, cfgErr("engine", "connectnetwork", "ConnectNetwork", err)
	}
	return out, nil
}

// DisconnectNetwork wraps the corresponding Engine client method.
func DisconnectNetwork(ctx context.Context, api Engine, name string, opts ...EngineNetworkDisconnectOption) (NetworkDisconnectResult, error) {
	if err := engineAPI(api, "DisconnectNetwork"); err != nil {
		return NetworkDisconnectResult{}, err
	}
	if name == "" {
		return NetworkDisconnectResult{}, cfgErr("engine", "name", "DisconnectNetwork", fmt.Errorf("%w: name", errRequired))
	}
	var act engineNetworkDisconnectAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("disconnectnetwork.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineNetworkDisconnectOption(&act))
	}
	if err := join(errs...); err != nil {
		return NetworkDisconnectResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NetworkDisconnect(ctx, name, act.opts)
	if err != nil {
		return NetworkDisconnectResult{}, cfgErr("engine", "disconnectnetwork", "DisconnectNetwork", err)
	}
	return out, nil
}

// InspectVolume wraps the corresponding Engine client method.
func InspectVolume(ctx context.Context, api Engine, id string) (VolumeInspectResult, error) {
	if err := engineAPI(api, "InspectVolume"); err != nil {
		return VolumeInspectResult{}, err
	}
	if id == "" {
		return VolumeInspectResult{}, cfgErr("engine", "id", "InspectVolume", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.VolumeInspect(ctx, id, VolumeInspectOptions{})
	if err != nil {
		return VolumeInspectResult{}, cfgErr("engine", "inspectvolume", "InspectVolume", err)
	}
	return out, nil
}

// ListEngineVolumes wraps the corresponding Engine client method.
func ListEngineVolumes(ctx context.Context, api Engine, opts ...EngineVolumeListOption) (VolumeListResult, error) {
	if err := engineAPI(api, "ListEngineVolumes"); err != nil {
		return VolumeListResult{}, err
	}
	var act engineVolumeListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listenginevolumes.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineVolumeListOption(&act))
	}
	if err := join(errs...); err != nil {
		return VolumeListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.VolumeList(ctx, act.opts)
	if err != nil {
		return VolumeListResult{}, cfgErr("engine", "listenginevolumes", "ListEngineVolumes", err)
	}
	return out, nil
}

// RemoveVolume wraps the corresponding Engine client method.
func RemoveVolume(ctx context.Context, api Engine, id string, opts ...EngineVolumeRemoveOption) (VolumeRemoveResult, error) {
	if err := engineAPI(api, "RemoveVolume"); err != nil {
		return VolumeRemoveResult{}, err
	}
	if id == "" {
		return VolumeRemoveResult{}, cfgErr("engine", "id", "RemoveVolume", fmt.Errorf("%w: id", errRequired))
	}
	var act engineVolumeRemoveAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("removevolume.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineVolumeRemoveOption(&act))
	}
	if err := join(errs...); err != nil {
		return VolumeRemoveResult{}, wrapReceiver("engine", err)
	}
	out, err := api.VolumeRemove(ctx, id, act.opts)
	if err != nil {
		return VolumeRemoveResult{}, cfgErr("engine", "removevolume", "RemoveVolume", err)
	}
	return out, nil
}

// PruneVolumes wraps the corresponding Engine client method.
func PruneVolumes(ctx context.Context, api Engine, opts ...EngineVolumePruneOption) (VolumePruneResult, error) {
	if err := engineAPI(api, "PruneVolumes"); err != nil {
		return VolumePruneResult{}, err
	}
	var act engineVolumePruneAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("prunevolumes.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineVolumePruneOption(&act))
	}
	if err := join(errs...); err != nil {
		return VolumePruneResult{}, wrapReceiver("engine", err)
	}
	out, err := api.VolumePrune(ctx, act.opts)
	if err != nil {
		return VolumePruneResult{}, cfgErr("engine", "prunevolumes", "PruneVolumes", err)
	}
	return out, nil
}

// Ping wraps the corresponding Engine client method.
func Ping(ctx context.Context, api Engine, opts ...EnginePingOption) (PingResult, error) {
	if err := engineAPI(api, "Ping"); err != nil {
		return PingResult{}, err
	}
	var act enginePingAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("ping.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePingOption(&act))
	}
	if err := join(errs...); err != nil {
		return PingResult{}, wrapReceiver("engine", err)
	}
	out, err := api.Ping(ctx, act.opts)
	if err != nil {
		return PingResult{}, cfgErr("engine", "ping", "Ping", err)
	}
	return out, nil
}

// Info wraps the corresponding Engine client method.
func Info(ctx context.Context, api Engine) (SystemInfoResult, error) {
	if err := engineAPI(api, "Info"); err != nil {
		return SystemInfoResult{}, err
	}
	out, err := api.Info(ctx, InfoOptions{})
	if err != nil {
		return SystemInfoResult{}, cfgErr("engine", "info", "Info", err)
	}
	return out, nil
}

// ListenEvents wraps Engine.Events.
func ListenEvents(ctx context.Context, api Engine, opts ...EngineEventsOption) EventsResult {
	if engineAPI(api, "ListenEvents") != nil {
		return EventsResult{}
	}
	var act engineEventsAct
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		_ = opt.applyEngineEventsOption(&act)
	}
	return api.Events(ctx, act.opts)
}

// DiskUsage wraps the corresponding Engine client method.
func DiskUsage(ctx context.Context, api Engine, opts ...EngineDiskUsageOption) (DiskUsageResult, error) {
	if err := engineAPI(api, "DiskUsage"); err != nil {
		return DiskUsageResult{}, err
	}
	var act engineDiskUsageAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("diskusage.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineDiskUsageOption(&act))
	}
	if err := join(errs...); err != nil {
		return DiskUsageResult{}, wrapReceiver("engine", err)
	}
	out, err := api.DiskUsage(ctx, act.opts)
	if err != nil {
		return DiskUsageResult{}, cfgErr("engine", "diskusage", "DiskUsage", err)
	}
	return out, nil
}

// Login wraps the corresponding Engine client method.
func Login(ctx context.Context, api Engine, opts ...EngineLoginOption) (RegistryLoginResult, error) {
	if err := engineAPI(api, "Login"); err != nil {
		return RegistryLoginResult{}, err
	}
	var act engineLoginAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("login.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineLoginOption(&act))
	}
	if err := join(errs...); err != nil {
		return RegistryLoginResult{}, wrapReceiver("engine", err)
	}
	out, err := api.RegistryLogin(ctx, act.opts)
	if err != nil {
		return RegistryLoginResult{}, cfgErr("engine", "login", "Login", err)
	}
	return out, nil
}

// InspectDistribution wraps the corresponding Engine client method.
func InspectDistribution(ctx context.Context, api Engine, image string) (DistributionInspectResult, error) {
	if err := engineAPI(api, "InspectDistribution"); err != nil {
		return DistributionInspectResult{}, err
	}
	if image == "" {
		return DistributionInspectResult{}, cfgErr("engine", "image", "InspectDistribution", fmt.Errorf("%w: image", errRequired))
	}
	out, err := api.DistributionInspect(ctx, image, DistributionInspectOptions{})
	if err != nil {
		return DistributionInspectResult{}, cfgErr("engine", "inspectdistribution", "InspectDistribution", err)
	}
	return out, nil
}

// InitSwarm wraps the corresponding Engine client method.
func InitSwarm(ctx context.Context, api Engine, opts ...EngineSwarmInitOption) (SwarmInitResult, error) {
	if err := engineAPI(api, "InitSwarm"); err != nil {
		return SwarmInitResult{}, err
	}
	var act engineSwarmInitAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("initswarm.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineSwarmInitOption(&act))
	}
	if err := join(errs...); err != nil {
		return SwarmInitResult{}, wrapReceiver("engine", err)
	}
	out, err := api.SwarmInit(ctx, act.opts)
	if err != nil {
		return SwarmInitResult{}, cfgErr("engine", "initswarm", "InitSwarm", err)
	}
	return out, nil
}

// JoinSwarm wraps the corresponding Engine client method.
func JoinSwarm(ctx context.Context, api Engine, opts ...EngineSwarmJoinOption) (SwarmJoinResult, error) {
	if err := engineAPI(api, "JoinSwarm"); err != nil {
		return SwarmJoinResult{}, err
	}
	var act engineSwarmJoinAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("joinswarm.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineSwarmJoinOption(&act))
	}
	if err := join(errs...); err != nil {
		return SwarmJoinResult{}, wrapReceiver("engine", err)
	}
	out, err := api.SwarmJoin(ctx, act.opts)
	if err != nil {
		return SwarmJoinResult{}, cfgErr("engine", "joinswarm", "JoinSwarm", err)
	}
	return out, nil
}

// InspectSwarm wraps the corresponding Engine client method.
func InspectSwarm(ctx context.Context, api Engine) (SwarmInspectResult, error) {
	if err := engineAPI(api, "InspectSwarm"); err != nil {
		return SwarmInspectResult{}, err
	}
	out, err := api.SwarmInspect(ctx, SwarmInspectOptions{})
	if err != nil {
		return SwarmInspectResult{}, cfgErr("engine", "inspectswarm", "InspectSwarm", err)
	}
	return out, nil
}

// LeaveSwarm wraps the corresponding Engine client method.
func LeaveSwarm(ctx context.Context, api Engine, opts ...EngineSwarmLeaveOption) (SwarmLeaveResult, error) {
	if err := engineAPI(api, "LeaveSwarm"); err != nil {
		return SwarmLeaveResult{}, err
	}
	var act engineSwarmLeaveAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("leaveswarm.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineSwarmLeaveOption(&act))
	}
	if err := join(errs...); err != nil {
		return SwarmLeaveResult{}, wrapReceiver("engine", err)
	}
	out, err := api.SwarmLeave(ctx, act.opts)
	if err != nil {
		return SwarmLeaveResult{}, cfgErr("engine", "leaveswarm", "LeaveSwarm", err)
	}
	return out, nil
}

// UnlockSwarm wraps the corresponding Engine client method.
func UnlockSwarm(ctx context.Context, api Engine, opts ...EngineSwarmUnlockOption) (SwarmUnlockResult, error) {
	if err := engineAPI(api, "UnlockSwarm"); err != nil {
		return SwarmUnlockResult{}, err
	}
	var act engineSwarmUnlockAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("unlockswarm.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineSwarmUnlockOption(&act))
	}
	if err := join(errs...); err != nil {
		return SwarmUnlockResult{}, wrapReceiver("engine", err)
	}
	out, err := api.SwarmUnlock(ctx, act.opts)
	if err != nil {
		return SwarmUnlockResult{}, cfgErr("engine", "unlockswarm", "UnlockSwarm", err)
	}
	return out, nil
}

// SwarmUnlockKey wraps the corresponding Engine client method.
func SwarmUnlockKey(ctx context.Context, api Engine) (SwarmGetUnlockKeyResult, error) {
	if err := engineAPI(api, "SwarmUnlockKey"); err != nil {
		return SwarmGetUnlockKeyResult{}, err
	}
	out, err := api.SwarmGetUnlockKey(ctx)
	if err != nil {
		return SwarmGetUnlockKeyResult{}, cfgErr("engine", "swarmunlockkey", "SwarmUnlockKey", err)
	}
	return out, nil
}

// UpdateSwarm wraps the corresponding Engine client method.
func UpdateSwarm(ctx context.Context, api Engine, opts ...EngineSwarmUpdateOption) (SwarmUpdateResult, error) {
	if err := engineAPI(api, "UpdateSwarm"); err != nil {
		return SwarmUpdateResult{}, err
	}
	var act engineSwarmUpdateAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("updateswarm.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineSwarmUpdateOption(&act))
	}
	if err := join(errs...); err != nil {
		return SwarmUpdateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.SwarmUpdate(ctx, act.opts)
	if err != nil {
		return SwarmUpdateResult{}, cfgErr("engine", "updateswarm", "UpdateSwarm", err)
	}
	return out, nil
}

// ListPlugins wraps the corresponding Engine client method.
func ListPlugins(ctx context.Context, api Engine, opts ...EnginePluginListOption) (PluginListResult, error) {
	if err := engineAPI(api, "ListPlugins"); err != nil {
		return PluginListResult{}, err
	}
	var act enginePluginListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listplugins.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginListOption(&act))
	}
	if err := join(errs...); err != nil {
		return PluginListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.PluginList(ctx, act.opts)
	if err != nil {
		return PluginListResult{}, cfgErr("engine", "listplugins", "ListPlugins", err)
	}
	return out, nil
}

// InspectPlugin wraps the corresponding Engine client method.
func InspectPlugin(ctx context.Context, api Engine, name string) (PluginInspectResult, error) {
	if err := engineAPI(api, "InspectPlugin"); err != nil {
		return PluginInspectResult{}, err
	}
	if name == "" {
		return PluginInspectResult{}, cfgErr("engine", "name", "InspectPlugin", fmt.Errorf("%w: name", errRequired))
	}
	out, err := api.PluginInspect(ctx, name, PluginInspectOptions{})
	if err != nil {
		return PluginInspectResult{}, cfgErr("engine", "inspectplugin", "InspectPlugin", err)
	}
	return out, nil
}

// RemovePlugin wraps the corresponding Engine client method.
func RemovePlugin(ctx context.Context, api Engine, name string, opts ...EnginePluginRemoveOption) (PluginRemoveResult, error) {
	if err := engineAPI(api, "RemovePlugin"); err != nil {
		return PluginRemoveResult{}, err
	}
	if name == "" {
		return PluginRemoveResult{}, cfgErr("engine", "name", "RemovePlugin", fmt.Errorf("%w: name", errRequired))
	}
	var act enginePluginRemoveAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("removeplugin.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginRemoveOption(&act))
	}
	if err := join(errs...); err != nil {
		return PluginRemoveResult{}, wrapReceiver("engine", err)
	}
	out, err := api.PluginRemove(ctx, name, act.opts)
	if err != nil {
		return PluginRemoveResult{}, cfgErr("engine", "removeplugin", "RemovePlugin", err)
	}
	return out, nil
}

// EnablePlugin wraps the corresponding Engine client method.
func EnablePlugin(ctx context.Context, api Engine, name string, opts ...EnginePluginEnableOption) (PluginEnableResult, error) {
	if err := engineAPI(api, "EnablePlugin"); err != nil {
		return PluginEnableResult{}, err
	}
	if name == "" {
		return PluginEnableResult{}, cfgErr("engine", "name", "EnablePlugin", fmt.Errorf("%w: name", errRequired))
	}
	var act enginePluginEnableAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("enableplugin.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginEnableOption(&act))
	}
	if err := join(errs...); err != nil {
		return PluginEnableResult{}, wrapReceiver("engine", err)
	}
	out, err := api.PluginEnable(ctx, name, act.opts)
	if err != nil {
		return PluginEnableResult{}, cfgErr("engine", "enableplugin", "EnablePlugin", err)
	}
	return out, nil
}

// DisablePlugin wraps the corresponding Engine client method.
func DisablePlugin(ctx context.Context, api Engine, name string, opts ...EnginePluginDisableOption) (PluginDisableResult, error) {
	if err := engineAPI(api, "DisablePlugin"); err != nil {
		return PluginDisableResult{}, err
	}
	if name == "" {
		return PluginDisableResult{}, cfgErr("engine", "name", "DisablePlugin", fmt.Errorf("%w: name", errRequired))
	}
	var act enginePluginDisableAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("disableplugin.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginDisableOption(&act))
	}
	if err := join(errs...); err != nil {
		return PluginDisableResult{}, wrapReceiver("engine", err)
	}
	out, err := api.PluginDisable(ctx, name, act.opts)
	if err != nil {
		return PluginDisableResult{}, cfgErr("engine", "disableplugin", "DisablePlugin", err)
	}
	return out, nil
}

// ListNodes wraps the corresponding Engine client method.
func ListNodes(ctx context.Context, api Engine, opts ...EngineNodeListOption) (NodeListResult, error) {
	if err := engineAPI(api, "ListNodes"); err != nil {
		return NodeListResult{}, err
	}
	var act engineNodeListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listnodes.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineNodeListOption(&act))
	}
	if err := join(errs...); err != nil {
		return NodeListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NodeList(ctx, act.opts)
	if err != nil {
		return NodeListResult{}, cfgErr("engine", "listnodes", "ListNodes", err)
	}
	return out, nil
}

// InspectNode wraps the corresponding Engine client method.
func InspectNode(ctx context.Context, api Engine, id string) (NodeInspectResult, error) {
	if err := engineAPI(api, "InspectNode"); err != nil {
		return NodeInspectResult{}, err
	}
	if id == "" {
		return NodeInspectResult{}, cfgErr("engine", "id", "InspectNode", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.NodeInspect(ctx, id, NodeInspectOptions{})
	if err != nil {
		return NodeInspectResult{}, cfgErr("engine", "inspectnode", "InspectNode", err)
	}
	return out, nil
}

// RemoveNode wraps the corresponding Engine client method.
func RemoveNode(ctx context.Context, api Engine, id string, opts ...EngineNodeRemoveOption) (NodeRemoveResult, error) {
	if err := engineAPI(api, "RemoveNode"); err != nil {
		return NodeRemoveResult{}, err
	}
	if id == "" {
		return NodeRemoveResult{}, cfgErr("engine", "id", "RemoveNode", fmt.Errorf("%w: id", errRequired))
	}
	var act engineNodeRemoveAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("removenode.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineNodeRemoveOption(&act))
	}
	if err := join(errs...); err != nil {
		return NodeRemoveResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NodeRemove(ctx, id, act.opts)
	if err != nil {
		return NodeRemoveResult{}, cfgErr("engine", "removenode", "RemoveNode", err)
	}
	return out, nil
}

// ListSwarmServices wraps the corresponding Engine client method.
func ListSwarmServices(ctx context.Context, api Engine, opts ...EngineServiceListOption) (ServiceListResult, error) {
	if err := engineAPI(api, "ListSwarmServices"); err != nil {
		return ServiceListResult{}, err
	}
	var act engineServiceListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listswarmservices.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineServiceListOption(&act))
	}
	if err := join(errs...); err != nil {
		return ServiceListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ServiceList(ctx, act.opts)
	if err != nil {
		return ServiceListResult{}, cfgErr("engine", "listswarmservices", "ListSwarmServices", err)
	}
	return out, nil
}

// InspectSwarmService wraps the corresponding Engine client method.
func InspectSwarmService(ctx context.Context, api Engine, id string) (ServiceInspectResult, error) {
	if err := engineAPI(api, "InspectSwarmService"); err != nil {
		return ServiceInspectResult{}, err
	}
	if id == "" {
		return ServiceInspectResult{}, cfgErr("engine", "id", "InspectSwarmService", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.ServiceInspect(ctx, id, ServiceInspectOptions{})
	if err != nil {
		return ServiceInspectResult{}, cfgErr("engine", "inspectswarmservice", "InspectSwarmService", err)
	}
	return out, nil
}

// RemoveSwarmService wraps the corresponding Engine client method.
func RemoveSwarmService(ctx context.Context, api Engine, id string) (ServiceRemoveResult, error) {
	if err := engineAPI(api, "RemoveSwarmService"); err != nil {
		return ServiceRemoveResult{}, err
	}
	if id == "" {
		return ServiceRemoveResult{}, cfgErr("engine", "id", "RemoveSwarmService", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.ServiceRemove(ctx, id, ServiceRemoveOptions{})
	if err != nil {
		return ServiceRemoveResult{}, cfgErr("engine", "removeswarmservice", "RemoveSwarmService", err)
	}
	return out, nil
}

// ListTasks wraps the corresponding Engine client method.
func ListTasks(ctx context.Context, api Engine, opts ...EngineTaskListOption) (TaskListResult, error) {
	if err := engineAPI(api, "ListTasks"); err != nil {
		return TaskListResult{}, err
	}
	var act engineTaskListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listtasks.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineTaskListOption(&act))
	}
	if err := join(errs...); err != nil {
		return TaskListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.TaskList(ctx, act.opts)
	if err != nil {
		return TaskListResult{}, cfgErr("engine", "listtasks", "ListTasks", err)
	}
	return out, nil
}

// InspectTask wraps the corresponding Engine client method.
func InspectTask(ctx context.Context, api Engine, id string) (TaskInspectResult, error) {
	if err := engineAPI(api, "InspectTask"); err != nil {
		return TaskInspectResult{}, err
	}
	if id == "" {
		return TaskInspectResult{}, cfgErr("engine", "id", "InspectTask", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.TaskInspect(ctx, id, TaskInspectOptions{})
	if err != nil {
		return TaskInspectResult{}, cfgErr("engine", "inspecttask", "InspectTask", err)
	}
	return out, nil
}

// ListSecrets wraps the corresponding Engine client method.
func ListSecrets(ctx context.Context, api Engine, opts ...EngineSecretListOption) (SecretListResult, error) {
	if err := engineAPI(api, "ListSecrets"); err != nil {
		return SecretListResult{}, err
	}
	var act engineSecretListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listsecrets.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineSecretListOption(&act))
	}
	if err := join(errs...); err != nil {
		return SecretListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.SecretList(ctx, act.opts)
	if err != nil {
		return SecretListResult{}, cfgErr("engine", "listsecrets", "ListSecrets", err)
	}
	return out, nil
}

// InspectSecret wraps the corresponding Engine client method.
func InspectSecret(ctx context.Context, api Engine, id string) (SecretInspectResult, error) {
	if err := engineAPI(api, "InspectSecret"); err != nil {
		return SecretInspectResult{}, err
	}
	if id == "" {
		return SecretInspectResult{}, cfgErr("engine", "id", "InspectSecret", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.SecretInspect(ctx, id, SecretInspectOptions{})
	if err != nil {
		return SecretInspectResult{}, cfgErr("engine", "inspectsecret", "InspectSecret", err)
	}
	return out, nil
}

// RemoveSecret wraps the corresponding Engine client method.
func RemoveSecret(ctx context.Context, api Engine, id string) (SecretRemoveResult, error) {
	if err := engineAPI(api, "RemoveSecret"); err != nil {
		return SecretRemoveResult{}, err
	}
	if id == "" {
		return SecretRemoveResult{}, cfgErr("engine", "id", "RemoveSecret", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.SecretRemove(ctx, id, SecretRemoveOptions{})
	if err != nil {
		return SecretRemoveResult{}, cfgErr("engine", "removesecret", "RemoveSecret", err)
	}
	return out, nil
}

// ListConfigs wraps the corresponding Engine client method.
func ListConfigs(ctx context.Context, api Engine, opts ...EngineConfigListOption) (ConfigListResult, error) {
	if err := engineAPI(api, "ListConfigs"); err != nil {
		return ConfigListResult{}, err
	}
	var act engineConfigListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listconfigs.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineConfigListOption(&act))
	}
	if err := join(errs...); err != nil {
		return ConfigListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ConfigList(ctx, act.opts)
	if err != nil {
		return ConfigListResult{}, cfgErr("engine", "listconfigs", "ListConfigs", err)
	}
	return out, nil
}

// InspectConfig wraps the corresponding Engine client method.
func InspectConfig(ctx context.Context, api Engine, id string) (ConfigInspectResult, error) {
	if err := engineAPI(api, "InspectConfig"); err != nil {
		return ConfigInspectResult{}, err
	}
	if id == "" {
		return ConfigInspectResult{}, cfgErr("engine", "id", "InspectConfig", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.ConfigInspect(ctx, id, ConfigInspectOptions{})
	if err != nil {
		return ConfigInspectResult{}, cfgErr("engine", "inspectconfig", "InspectConfig", err)
	}
	return out, nil
}

// RemoveConfig wraps the corresponding Engine client method.
func RemoveConfig(ctx context.Context, api Engine, id string) (ConfigRemoveResult, error) {
	if err := engineAPI(api, "RemoveConfig"); err != nil {
		return ConfigRemoveResult{}, err
	}
	if id == "" {
		return ConfigRemoveResult{}, cfgErr("engine", "id", "RemoveConfig", fmt.Errorf("%w: id", errRequired))
	}
	out, err := api.ConfigRemove(ctx, id, ConfigRemoveOptions{})
	if err != nil {
		return ConfigRemoveResult{}, cfgErr("engine", "removeconfig", "RemoveConfig", err)
	}
	return out, nil
}

// ListCheckpoints wraps the corresponding Engine client method.
func ListCheckpoints(ctx context.Context, api Engine, id string, opts ...EngineCheckpointListOption) (CheckpointListResult, error) {
	if err := engineAPI(api, "ListCheckpoints"); err != nil {
		return CheckpointListResult{}, err
	}
	if id == "" {
		return CheckpointListResult{}, cfgErr("engine", "id", "ListCheckpoints", fmt.Errorf("%w: id", errRequired))
	}
	var act engineCheckpointListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("listcheckpoints.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineCheckpointListOption(&act))
	}
	if err := join(errs...); err != nil {
		return CheckpointListResult{}, wrapReceiver("engine", err)
	}
	out, err := api.CheckpointList(ctx, id, act.opts)
	if err != nil {
		return CheckpointListResult{}, cfgErr("engine", "listcheckpoints", "ListCheckpoints", err)
	}
	return out, nil
}

// CreateCheckpoint wraps the corresponding Engine client method.
func CreateCheckpoint(ctx context.Context, api Engine, id string, opts ...EngineCheckpointCreateOption) (CheckpointCreateResult, error) {
	if err := engineAPI(api, "CreateCheckpoint"); err != nil {
		return CheckpointCreateResult{}, err
	}
	if id == "" {
		return CheckpointCreateResult{}, cfgErr("engine", "id", "CreateCheckpoint", fmt.Errorf("%w: id", errRequired))
	}
	var act engineCheckpointCreateAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("createcheckpoint.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineCheckpointCreateOption(&act))
	}
	if err := join(errs...); err != nil {
		return CheckpointCreateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.CheckpointCreate(ctx, id, act.opts)
	if err != nil {
		return CheckpointCreateResult{}, cfgErr("engine", "createcheckpoint", "CreateCheckpoint", err)
	}
	return out, nil
}

// RemoveCheckpoint wraps the corresponding Engine client method.
func RemoveCheckpoint(ctx context.Context, api Engine, id string, opts ...EngineCheckpointRemoveOption) (CheckpointRemoveResult, error) {
	if err := engineAPI(api, "RemoveCheckpoint"); err != nil {
		return CheckpointRemoveResult{}, err
	}
	if id == "" {
		return CheckpointRemoveResult{}, cfgErr("engine", "id", "RemoveCheckpoint", fmt.Errorf("%w: id", errRequired))
	}
	var act engineCheckpointRemoveAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("removecheckpoint.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineCheckpointRemoveOption(&act))
	}
	if err := join(errs...); err != nil {
		return CheckpointRemoveResult{}, wrapReceiver("engine", err)
	}
	out, err := api.CheckpointRemove(ctx, id, act.opts)
	if err != nil {
		return CheckpointRemoveResult{}, cfgErr("engine", "removecheckpoint", "RemoveCheckpoint", err)
	}
	return out, nil
}
