package kit

import (
	"context"
	"io"
	"net"
)

// Remaining [Engine] methods so fakeEngine implements the full client.

func (f *fakeEngine) ClientVersion() string {
	if f.clientVer != "" {
		return f.clientVer
	}
	return "1.56"
}

func (f *fakeEngine) DaemonHost() string {
	if f.daemonHost != "" {
		return f.daemonHost
	}
	return "unix:///var/run/docker.sock"
}

func (f *fakeEngine) ServerVersion(ctx context.Context, options ServerVersionOptions) (ServerVersionResult, error) {
	if f.serverVer.Version != "" {
		return f.serverVer, nil
	}
	return ServerVersionResult{Version: "29.0.0", APIVersion: "1.56"}, nil
}

func (f *fakeEngine) Dialer() func(context.Context) (net.Conn, error) {
	return func(context.Context) (net.Conn, error) { return nil, nil }
}

func (f *fakeEngine) DialHijack(ctx context.Context, url, proto string, meta map[string][]string) (net.Conn, error) {
	return nil, nil
}

func (f *fakeEngine) CheckpointCreate(ctx context.Context, container string, options CheckpointCreateOptions) (CheckpointCreateResult, error) {
	return CheckpointCreateResult{}, nil
}

func (f *fakeEngine) CheckpointRemove(ctx context.Context, container string, options CheckpointRemoveOptions) (CheckpointRemoveResult, error) {
	return CheckpointRemoveResult{}, nil
}

func (f *fakeEngine) CheckpointList(ctx context.Context, container string, options CheckpointListOptions) (CheckpointListResult, error) {
	return CheckpointListResult{}, nil
}

func (f *fakeEngine) ContainerList(ctx context.Context, options ContainerListOptions) (ContainerListResult, error) {
	f.listed = options
	return ContainerListResult{}, nil
}

func (f *fakeEngine) ContainerUpdate(ctx context.Context, container string, updateConfig ContainerUpdateOptions) (ContainerUpdateResult, error) {
	return ContainerUpdateResult{}, nil
}

func (f *fakeEngine) ContainerPrune(ctx context.Context, opts ContainerPruneOptions) (ContainerPruneResult, error) {
	return ContainerPruneResult{}, nil
}

func (f *fakeEngine) ContainerRename(ctx context.Context, container string, options ContainerRenameOptions) (ContainerRenameResult, error) {
	return ContainerRenameResult{}, nil
}

func (f *fakeEngine) ContainerResize(ctx context.Context, container string, options ContainerResizeOptions) (ContainerResizeResult, error) {
	return ContainerResizeResult{}, nil
}

func (f *fakeEngine) ContainerAttach(ctx context.Context, container string, options ContainerAttachOptions) (ContainerAttachResult, error) {
	return ContainerAttachResult{}, nil
}

func (f *fakeEngine) ContainerCommit(ctx context.Context, container string, options ContainerCommitOptions) (ContainerCommitResult, error) {
	return ContainerCommitResult{}, nil
}

func (f *fakeEngine) ContainerDiff(ctx context.Context, container string, options ContainerDiffOptions) (ContainerDiffResult, error) {
	return ContainerDiffResult{}, nil
}

func (f *fakeEngine) ContainerExport(ctx context.Context, container string, options ContainerExportOptions) (ContainerExportResult, error) {
	return nil, nil
}

func (f *fakeEngine) ContainerStats(ctx context.Context, container string, options ContainerStatsOptions) (ContainerStatsResult, error) {
	return ContainerStatsResult{}, nil
}

func (f *fakeEngine) ContainerTop(ctx context.Context, container string, options ContainerTopOptions) (ContainerTopResult, error) {
	return ContainerTopResult{}, nil
}

func (f *fakeEngine) ContainerStatPath(ctx context.Context, container string, options ContainerStatPathOptions) (ContainerStatPathResult, error) {
	return ContainerStatPathResult{}, nil
}

func (f *fakeEngine) CopyFromContainer(ctx context.Context, container string, options CopyFromContainerOptions) (CopyFromContainerResult, error) {
	return CopyFromContainerResult{}, nil
}

func (f *fakeEngine) CopyToContainer(ctx context.Context, container string, options CopyToContainerOptions) (CopyToContainerResult, error) {
	return CopyToContainerResult{}, nil
}

func (f *fakeEngine) ExecCreate(ctx context.Context, container string, options ExecCreateOptions) (ExecCreateResult, error) {
	return ExecCreateResult{}, nil
}

func (f *fakeEngine) ExecInspect(ctx context.Context, execID string, options ExecInspectOptions) (ExecInspectResult, error) {
	return ExecInspectResult{}, nil
}

func (f *fakeEngine) ExecResize(ctx context.Context, execID string, options ExecResizeOptions) (ExecResizeResult, error) {
	return ExecResizeResult{}, nil
}

func (f *fakeEngine) ExecStart(ctx context.Context, execID string, options ExecStartOptions) (ExecStartResult, error) {
	return ExecStartResult{}, nil
}

func (f *fakeEngine) ExecAttach(ctx context.Context, execID string, options ExecAttachOptions) (ExecAttachResult, error) {
	return ExecAttachResult{}, nil
}

func (f *fakeEngine) DistributionInspect(ctx context.Context, image string, options DistributionInspectOptions) (DistributionInspectResult, error) {
	return DistributionInspectResult{}, nil
}

func (f *fakeEngine) ImageSearch(ctx context.Context, term string, options ImageSearchOptions) (ImageSearchResult, error) {
	return ImageSearchResult{}, nil
}

func (f *fakeEngine) BuildCachePrune(ctx context.Context, opts BuildCachePruneOptions) (BuildCachePruneResult, error) {
	return BuildCachePruneResult{}, nil
}

func (f *fakeEngine) BuildCancel(ctx context.Context, id string, opts BuildCancelOptions) (BuildCancelResult, error) {
	return BuildCancelResult{}, nil
}

func (f *fakeEngine) ImageImport(ctx context.Context, source ImageImportSource, ref string, options ImageImportOptions) (ImageImportResult, error) {
	return nil, nil
}

func (f *fakeEngine) ImageList(ctx context.Context, options ImageListOptions) (ImageListResult, error) {
	return ImageListResult{}, nil
}

func (f *fakeEngine) ImagePush(ctx context.Context, ref string, options ImagePushOptions) (ImagePushResponse, error) {
	return nil, nil
}

func (f *fakeEngine) ImageRemove(ctx context.Context, image string, options ImageRemoveOptions) (ImageRemoveResult, error) {
	return ImageRemoveResult{}, nil
}

func (f *fakeEngine) ImageTag(ctx context.Context, options ImageTagOptions) (ImageTagResult, error) {
	return ImageTagResult{}, nil
}

func (f *fakeEngine) ImagePrune(ctx context.Context, opts ImagePruneOptions) (ImagePruneResult, error) {
	return ImagePruneResult{}, nil
}

func (f *fakeEngine) ImageInspect(ctx context.Context, image string, _ ...ImageInspectOption) (ImageInspectResult, error) {
	return ImageInspectResult{}, nil
}

func (f *fakeEngine) ImageHistory(ctx context.Context, image string, _ ...ImageHistoryOption) (ImageHistoryResult, error) {
	return ImageHistoryResult{}, nil
}

func (f *fakeEngine) ImageAttestations(ctx context.Context, image string, _ ...ImageAttestationsOption) (ImageAttestationsResult, error) {
	return ImageAttestationsResult{}, nil
}

func (f *fakeEngine) ImageLoad(ctx context.Context, input io.Reader, _ ...ImageLoadOption) (ImageLoadResult, error) {
	return nil, nil
}

func (f *fakeEngine) ImageSave(ctx context.Context, images []string, _ ...ImageSaveOption) (ImageSaveResult, error) {
	return nil, nil
}

func (f *fakeEngine) NetworkInspect(ctx context.Context, network string, options NetworkInspectOptions) (NetworkInspectResult, error) {
	return NetworkInspectResult{}, nil
}

func (f *fakeEngine) NetworkList(ctx context.Context, options NetworkListOptions) (NetworkListResult, error) {
	return NetworkListResult{}, nil
}

func (f *fakeEngine) NetworkRemove(ctx context.Context, network string, options NetworkRemoveOptions) (NetworkRemoveResult, error) {
	return NetworkRemoveResult{}, nil
}

func (f *fakeEngine) NetworkPrune(ctx context.Context, opts NetworkPruneOptions) (NetworkPruneResult, error) {
	return NetworkPruneResult{}, nil
}

func (f *fakeEngine) NetworkConnect(ctx context.Context, network string, options NetworkConnectOptions) (NetworkConnectResult, error) {
	return NetworkConnectResult{}, nil
}

func (f *fakeEngine) NetworkDisconnect(ctx context.Context, network string, options NetworkDisconnectOptions) (NetworkDisconnectResult, error) {
	return NetworkDisconnectResult{}, nil
}

func (f *fakeEngine) NodeInspect(ctx context.Context, nodeID string, options NodeInspectOptions) (NodeInspectResult, error) {
	return NodeInspectResult{}, nil
}

func (f *fakeEngine) NodeList(ctx context.Context, options NodeListOptions) (NodeListResult, error) {
	return NodeListResult{}, nil
}

func (f *fakeEngine) NodeUpdate(ctx context.Context, nodeID string, options NodeUpdateOptions) (NodeUpdateResult, error) {
	f.nodeUpdate = options
	return NodeUpdateResult{}, nil
}

func (f *fakeEngine) NodeRemove(ctx context.Context, nodeID string, options NodeRemoveOptions) (NodeRemoveResult, error) {
	return NodeRemoveResult{}, nil
}

func (f *fakeEngine) PluginCreate(ctx context.Context, createContext io.Reader, options PluginCreateOptions) (PluginCreateResult, error) {
	f.pluginCreate = options
	if createContext != nil {
		_, _ = io.Copy(io.Discard, createContext)
	}
	return PluginCreateResult{}, nil
}

func (f *fakeEngine) PluginInstall(ctx context.Context, name string, options PluginInstallOptions) (PluginInstallResult, error) {
	f.pluginInstall = options
	return PluginInstallResult{}, nil
}

func (f *fakeEngine) PluginInspect(ctx context.Context, name string, options PluginInspectOptions) (PluginInspectResult, error) {
	return PluginInspectResult{}, nil
}

func (f *fakeEngine) PluginList(ctx context.Context, options PluginListOptions) (PluginListResult, error) {
	f.pluginList = options
	return PluginListResult{}, nil
}

func (f *fakeEngine) PluginRemove(ctx context.Context, name string, options PluginRemoveOptions) (PluginRemoveResult, error) {
	f.pluginRemove = options
	return PluginRemoveResult{}, nil
}

func (f *fakeEngine) PluginEnable(ctx context.Context, name string, options PluginEnableOptions) (PluginEnableResult, error) {
	f.pluginEnable = options
	return PluginEnableResult{}, nil
}

func (f *fakeEngine) PluginDisable(ctx context.Context, name string, options PluginDisableOptions) (PluginDisableResult, error) {
	f.pluginDisable = options
	return PluginDisableResult{}, nil
}

func (f *fakeEngine) PluginUpgrade(ctx context.Context, name string, options PluginUpgradeOptions) (PluginUpgradeResult, error) {
	f.pluginUpgrade = options
	return nil, nil
}

func (f *fakeEngine) PluginPush(ctx context.Context, name string, options PluginPushOptions) (PluginPushResult, error) {
	f.pluginPush = options
	return PluginPushResult{}, nil
}

func (f *fakeEngine) PluginSet(ctx context.Context, name string, options PluginSetOptions) (PluginSetResult, error) {
	f.pluginSet = options
	return PluginSetResult{}, nil
}

func (f *fakeEngine) ServiceCreate(ctx context.Context, options ServiceCreateOptions) (ServiceCreateResult, error) {
	f.serviceCreate = options
	return ServiceCreateResult{ID: "svc1"}, nil
}

func (f *fakeEngine) ServiceInspect(ctx context.Context, serviceID string, options ServiceInspectOptions) (ServiceInspectResult, error) {
	return ServiceInspectResult{}, nil
}

func (f *fakeEngine) ServiceList(ctx context.Context, options ServiceListOptions) (ServiceListResult, error) {
	return ServiceListResult{}, nil
}

func (f *fakeEngine) ServiceUpdate(ctx context.Context, serviceID string, options ServiceUpdateOptions) (ServiceUpdateResult, error) {
	f.serviceUpdate = options
	return ServiceUpdateResult{}, nil
}

func (f *fakeEngine) ServiceRemove(ctx context.Context, serviceID string, options ServiceRemoveOptions) (ServiceRemoveResult, error) {
	return ServiceRemoveResult{}, nil
}

func (f *fakeEngine) ServiceLogs(ctx context.Context, serviceID string, options ServiceLogsOptions) (ServiceLogsResult, error) {
	return nil, nil
}

func (f *fakeEngine) TaskInspect(ctx context.Context, taskID string, options TaskInspectOptions) (TaskInspectResult, error) {
	return TaskInspectResult{}, nil
}

func (f *fakeEngine) TaskList(ctx context.Context, options TaskListOptions) (TaskListResult, error) {
	return TaskListResult{}, nil
}

func (f *fakeEngine) TaskLogs(ctx context.Context, taskID string, options TaskLogsOptions) (TaskLogsResult, error) {
	return nil, nil
}

func (f *fakeEngine) SwarmInit(ctx context.Context, options SwarmInitOptions) (SwarmInitResult, error) {
	return SwarmInitResult{}, nil
}

func (f *fakeEngine) SwarmJoin(ctx context.Context, options SwarmJoinOptions) (SwarmJoinResult, error) {
	return SwarmJoinResult{}, nil
}

func (f *fakeEngine) SwarmInspect(ctx context.Context, options SwarmInspectOptions) (SwarmInspectResult, error) {
	return SwarmInspectResult{}, nil
}

func (f *fakeEngine) SwarmUpdate(ctx context.Context, options SwarmUpdateOptions) (SwarmUpdateResult, error) {
	return SwarmUpdateResult{}, nil
}

func (f *fakeEngine) SwarmLeave(ctx context.Context, options SwarmLeaveOptions) (SwarmLeaveResult, error) {
	return SwarmLeaveResult{}, nil
}

func (f *fakeEngine) SwarmGetUnlockKey(ctx context.Context) (SwarmGetUnlockKeyResult, error) {
	return SwarmGetUnlockKeyResult{}, nil
}

func (f *fakeEngine) SwarmUnlock(ctx context.Context, options SwarmUnlockOptions) (SwarmUnlockResult, error) {
	return SwarmUnlockResult{}, nil
}

func (f *fakeEngine) Events(ctx context.Context, options EventsListOptions) EventsResult {
	f.events = options
	return EventsResult{}
}

func (f *fakeEngine) Info(ctx context.Context, options InfoOptions) (SystemInfoResult, error) {
	return SystemInfoResult{}, nil
}

func (f *fakeEngine) RegistryLogin(ctx context.Context, auth RegistryLoginOptions) (RegistryLoginResult, error) {
	f.login = auth
	return RegistryLoginResult{}, nil
}

func (f *fakeEngine) DiskUsage(ctx context.Context, options DiskUsageOptions) (DiskUsageResult, error) {
	f.diskUsage = options
	return DiskUsageResult{}, nil
}

func (f *fakeEngine) Ping(ctx context.Context, options PingOptions) (PingResult, error) {
	f.ping = options
	return PingResult{APIVersion: "1.56"}, nil
}

func (f *fakeEngine) VolumeInspect(ctx context.Context, volumeID string, options VolumeInspectOptions) (VolumeInspectResult, error) {
	return VolumeInspectResult{}, nil
}

func (f *fakeEngine) VolumeList(ctx context.Context, options VolumeListOptions) (VolumeListResult, error) {
	return VolumeListResult{}, nil
}

func (f *fakeEngine) VolumeUpdate(ctx context.Context, volumeID string, options VolumeUpdateOptions) (VolumeUpdateResult, error) {
	f.volumeUpdate = options
	return VolumeUpdateResult{}, nil
}

func (f *fakeEngine) VolumeRemove(ctx context.Context, volumeID string, options VolumeRemoveOptions) (VolumeRemoveResult, error) {
	return VolumeRemoveResult{}, nil
}

func (f *fakeEngine) VolumePrune(ctx context.Context, options VolumePruneOptions) (VolumePruneResult, error) {
	return VolumePruneResult{}, nil
}

func (f *fakeEngine) SecretCreate(ctx context.Context, options SecretCreateOptions) (SecretCreateResult, error) {
	f.secretCreate = options
	return SecretCreateResult{ID: "sec1"}, nil
}

func (f *fakeEngine) SecretInspect(ctx context.Context, id string, options SecretInspectOptions) (SecretInspectResult, error) {
	return SecretInspectResult{}, nil
}

func (f *fakeEngine) SecretList(ctx context.Context, options SecretListOptions) (SecretListResult, error) {
	return SecretListResult{}, nil
}

func (f *fakeEngine) SecretUpdate(ctx context.Context, id string, options SecretUpdateOptions) (SecretUpdateResult, error) {
	return SecretUpdateResult{}, nil
}

func (f *fakeEngine) SecretRemove(ctx context.Context, id string, options SecretRemoveOptions) (SecretRemoveResult, error) {
	return SecretRemoveResult{}, nil
}

func (f *fakeEngine) ConfigCreate(ctx context.Context, options ConfigCreateOptions) (ConfigCreateResult, error) {
	f.configCreate = options
	return ConfigCreateResult{ID: "cfg1"}, nil
}

func (f *fakeEngine) ConfigInspect(ctx context.Context, id string, options ConfigInspectOptions) (ConfigInspectResult, error) {
	return ConfigInspectResult{}, nil
}

func (f *fakeEngine) ConfigList(ctx context.Context, options ConfigListOptions) (ConfigListResult, error) {
	return ConfigListResult{}, nil
}

func (f *fakeEngine) ConfigUpdate(ctx context.Context, id string, options ConfigUpdateOptions) (ConfigUpdateResult, error) {
	return ConfigUpdateResult{}, nil
}

func (f *fakeEngine) ConfigRemove(ctx context.Context, id string, options ConfigRemoveOptions) (ConfigRemoveResult, error) {
	return ConfigRemoveResult{}, nil
}
