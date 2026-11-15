package kit

import (
	"fmt"
	"io"

	"github.com/aptd3v/containerkit/pkg/df"
)

// FilterOpt is the value returned by [Filter].
// It implements [EngineBuildCachePruneOption], [EngineConfigListOption],
// [EngineContainerListOption], [EngineContainerPruneOption], [EngineEventsOption],
// [EngineImageListOption], [EngineImagePruneOption], [EngineImageSearchOption],
// [EngineNetworkListOption], [EngineNetworkPruneOption], [EngineNodeListOption],
// [EnginePluginListOption], [EngineSecretListOption], [EngineServiceListOption],
// [EngineTaskListOption], [EngineVolumeListOption], and [EngineVolumePruneOption].
type FilterOpt struct {
	term   string
	values []string
}

// Filter adds an Engine API filter term (list/prune/search/events).
func Filter(term string, values ...string) FilterOpt { return FilterOpt{term: term, values: values} }

func (o FilterOpt) add(dst *Filters) error {
	if o.term == "" {
		return cfgErr("", "filter", "Filter", fmt.Errorf("%w: filter term", errEmpty))
	}
	if *dst == nil {
		*dst = Filters{}
	}
	(*dst).Add(o.term, o.values...)
	return nil
}

func (o FilterOpt) applyEngineContainerListOption(a *engineContainerListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineContainerPruneOption(a *engineContainerPruneAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineImageListOption(a *engineImageListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineImagePruneOption(a *engineImagePruneAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineImageSearchOption(a *engineImageSearchAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineNetworkListOption(a *engineNetworkListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineNetworkPruneOption(a *engineNetworkPruneAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineVolumeListOption(a *engineVolumeListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineVolumePruneOption(a *engineVolumePruneAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineEventsOption(a *engineEventsAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineBuildCachePruneOption(a *engineBuildCachePruneAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEnginePluginListOption(a *enginePluginListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineNodeListOption(a *engineNodeListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineServiceListOption(a *engineServiceListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineTaskListOption(a *engineTaskListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineSecretListOption(a *engineSecretListAct) error {
	return o.add(&a.opts.Filters)
}
func (o FilterOpt) applyEngineConfigListOption(a *engineConfigListAct) error {
	return o.add(&a.opts.Filters)
}

// LimitOpt is the value returned by [Limit].
// It implements [EngineContainerListOption] and [EngineImageSearchOption].
type LimitOpt struct{ n int }

// Limit caps list/search results.
func Limit(n int) LimitOpt { return LimitOpt{n: n} }

func (o LimitOpt) applyEngineContainerListOption(a *engineContainerListAct) error {
	a.opts.Limit = o.n
	return nil
}
func (o LimitOpt) applyEngineImageSearchOption(a *engineImageSearchAct) error {
	a.opts.Limit = o.n
	return nil
}

func (o AllOpt) applyEngineContainerListOption(a *engineContainerListAct) error {
	a.opts.All = o.v
	return nil
}
func (o AllOpt) applyEngineImageListOption(a *engineImageListAct) error {
	a.opts.All = o.v
	return nil
}
func (o AllOpt) applyEngineVolumePruneOption(a *engineVolumePruneAct) error {
	a.opts.All = o.v
	return nil
}
func (o AllOpt) applyEngineBuildCachePruneOption(a *engineBuildCachePruneAct) error {
	a.opts.All = o.v
	return nil
}
func (o AllOpt) applyEngineImagePushOption(a *engineImagePushAct) error {
	a.opts.All = o.v
	return nil
}

func (o IncludeSizeOpt) applyEngineContainerListOption(a *engineContainerListAct) error {
	a.opts.Size = o.v
	return nil
}

func (o ForceOpt) applyEngineImageRemoveOption(a *engineImageRemoveAct) error {
	a.opts.Force = o.v
	return nil
}
func (o ForceOpt) applyEngineVolumeRemoveOption(a *engineVolumeRemoveAct) error {
	a.opts.Force = o.v
	return nil
}
func (o ForceOpt) applyEngineNetworkDisconnectOption(a *engineNetworkDisconnectAct) error {
	a.opts.Force = o.v
	return nil
}
func (o ForceOpt) applyEnginePluginRemoveOption(a *enginePluginRemoveAct) error {
	a.opts.Force = o.v
	return nil
}
func (o ForceOpt) applyEnginePluginDisableOption(a *enginePluginDisableAct) error {
	a.opts.Force = o.v
	return nil
}
func (o ForceOpt) applyEngineNodeRemoveOption(a *engineNodeRemoveAct) error {
	a.opts.Force = o.v
	return nil
}
func (o ForceOpt) applyEngineSwarmLeaveOption(a *engineSwarmLeaveAct) error {
	a.opts.Force = o.v
	return nil
}

func (o NameOpt) applyEngineRenameOption(a *engineRenameAct) error {
	a.opts.NewName = o.name
	return nil
}

func (o SourceOpt) applyEngineCopyFromOption(a *engineCopyFromAct) error {
	a.opts.SourcePath = o.s
	return nil
}
func (o SourceOpt) applyEngineImageTagOption(a *engineImageTagAct) error {
	a.opts.Source = o.s
	return nil
}

func (o TargetOpt) applyEngineCopyToOption(a *engineCopyToAct) error {
	a.opts.DestinationPath = o.s
	return nil
}
func (o TargetOpt) applyEngineImageTagOption(a *engineImageTagAct) error {
	a.opts.Target = o.s
	return nil
}

func (o ConsoleSizeOpt) applyEngineResizeOption(a *engineResizeAct) error {
	a.opts.Height = o.h
	a.opts.Width = o.w
	return nil
}
func (o ConsoleSizeOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.ConsoleSize.Height = o.h
	a.opts.ConsoleSize.Width = o.w
	return nil
}
func (o ConsoleSizeOpt) applyEngineExecAttachOption(a *engineExecAttachAct) error {
	a.opts.ConsoleSize.Height = o.h
	a.opts.ConsoleSize.Width = o.w
	return nil
}

func (o SinceOpt) applyEngineContainerListOption(a *engineContainerListAct) error {
	a.opts.Since = o.s
	return nil
}
func (o SinceOpt) applyEngineEventsOption(a *engineEventsAct) error {
	a.opts.Since = o.s
	return nil
}
func (o UntilOpt) applyEngineEventsOption(a *engineEventsAct) error {
	a.opts.Until = o.s
	return nil
}

func (o RegistryAuthOpt) applyEngineImagePushOption(a *engineImagePushAct) error {
	a.opts.RegistryAuth = o.s
	return nil
}
func (o RegistryAuthOpt) applyEngineImageSearchOption(a *engineImageSearchAct) error {
	a.opts.RegistryAuth = o.s
	return nil
}

func (o CommentOpt) applyEngineCommitOption(a *engineCommitAct) error {
	a.opts.Comment = o.s
	return nil
}
func (o AuthorOpt) applyEngineCommitOption(a *engineCommitAct) error {
	a.opts.Author = o.s
	return nil
}
func (o ReferenceOpt) applyEngineCommitOption(a *engineCommitAct) error {
	a.opts.Reference = o.s
	return nil
}

func (o CheckpointOpt) applyEngineCheckpointCreateOption(a *engineCheckpointCreateAct) error {
	a.opts.CheckpointID = o.id
	return nil
}
func (o CheckpointOpt) applyEngineCheckpointRemoveOption(a *engineCheckpointRemoveAct) error {
	a.opts.CheckpointID = o.id
	return nil
}
func (o CheckpointDirOpt) applyEngineCheckpointCreateOption(a *engineCheckpointCreateAct) error {
	a.opts.CheckpointDir = o.dir
	return nil
}
func (o CheckpointDirOpt) applyEngineCheckpointListOption(a *engineCheckpointListAct) error {
	a.opts.CheckpointDir = o.dir
	return nil
}
func (o CheckpointDirOpt) applyEngineCheckpointRemoveOption(a *engineCheckpointRemoveAct) error {
	a.opts.CheckpointDir = o.dir
	return nil
}

func (o CopyUIDGIDOpt) applyEngineCopyToOption(a *engineCopyToAct) error {
	a.opts.CopyUIDGID = o.v
	return nil
}
func (o DetachKeysOpt) applyEngineAttachOption(a *engineAttachAct) error {
	a.opts.DetachKeys = o.s
	return nil
}
func (o DetachKeysOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.DetachKeys = o.s
	return nil
}

func (o PrivilegedOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.Privileged = o.v
	return nil
}
func (o TtyOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.TTY = o.v
	return nil
}
func (o TtyOpt) applyEngineExecStartOption(a *engineExecStartAct) error {
	a.opts.TTY = o.v
	return nil
}
func (o TtyOpt) applyEngineExecAttachOption(a *engineExecAttachAct) error {
	a.opts.TTY = o.v
	return nil
}
func (o UserOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.User = o.user
	return nil
}
func (o DetachOpt) applyEngineExecStartOption(a *engineExecStartAct) error {
	a.opts.Detach = o.v
	return nil
}

// StreamOpt is the value returned by [Stream].
// It implements [EngineAttachOption] and [EngineStatsOption].
type StreamOpt struct{ v bool }

// Stream requests a streaming Engine response.
func Stream(v bool) StreamOpt { return StreamOpt{v: v} }

func (o StreamOpt) applyEngineAttachOption(a *engineAttachAct) error {
	a.opts.Stream = o.v
	return nil
}
func (o StreamOpt) applyEngineStatsOption(a *engineStatsAct) error {
	a.opts.Stream = o.v
	return nil
}

// VerboseOpt is the value returned by [Verbose].
// It implements [EngineDiskUsageOption] and [EngineNetworkInspectOption].
type VerboseOpt struct{ v bool }

// Verbose requests extra Engine inspect/list detail.
func Verbose(v bool) VerboseOpt { return VerboseOpt{v: v} }

func (o VerboseOpt) applyEngineNetworkInspectOption(a *engineNetworkInspectAct) error {
	a.opts.Verbose = o.v
	return nil
}
func (o VerboseOpt) applyEngineDiskUsageOption(a *engineDiskUsageAct) error {
	a.opts.Verbose = o.v
	return nil
}

// ForceNegotiateOpt is the value returned by [ForceNegotiate].
// It implements [EnginePingOption].
type ForceNegotiateOpt struct{ v bool }

// ForceNegotiate re-negotiates the Engine API version on [Ping], including explicit false.
func ForceNegotiate(v bool) ForceNegotiateOpt { return ForceNegotiateOpt{v: v} }

func (o ForceNegotiateOpt) applyEnginePingOption(a *enginePingAct) error {
	a.opts.ForceNegotiate = o.v
	return nil
}

// ImagesOpt is the value returned by [Images].
// It implements [EngineDiskUsageOption].
type ImagesOpt struct{ v bool }

// Images includes image disk usage on [DiskUsage], including explicit false.
func Images(v bool) ImagesOpt { return ImagesOpt{v: v} }

func (o ImagesOpt) applyEngineDiskUsageOption(a *engineDiskUsageAct) error {
	a.opts.Images = o.v
	return nil
}

// BuildCacheOpt is the value returned by [BuildCache].
// It implements [EngineDiskUsageOption].
type BuildCacheOpt struct{ v bool }

// BuildCache includes build-cache disk usage on [DiskUsage], including explicit false.
func BuildCache(v bool) BuildCacheOpt { return BuildCacheOpt{v: v} }

func (o BuildCacheOpt) applyEngineDiskUsageOption(a *engineDiskUsageAct) error {
	a.opts.BuildCache = o.v
	return nil
}

// UsageOpt is the value returned by [Usage].
// It implements [EngineDiskUsageOption].
type UsageOpt struct{ o df.Object }

// Usage selects a [DiskUsage] object type. Call more than once to include several types.
func Usage(o df.Object) UsageOpt { return UsageOpt{o: o} }

func (o UsageOpt) applyEngineDiskUsageOption(a *engineDiskUsageAct) error {
	switch o.o {
	case df.Container:
		a.opts.Containers = true
	case df.Image:
		a.opts.Images = true
	case df.Volume:
		a.opts.Volumes = true
	case df.BuildCache:
		a.opts.BuildCache = true
	default:
		return cfgErr("", "type", "Usage", fmt.Errorf("%w: disk usage type %q", errInvalid, o.o))
	}
	return nil
}

// UsernameOpt is the value returned by [Username].
// It implements [EngineLoginOption].
type UsernameOpt struct{ s string }

// Username sets the registry username for [Login].
func Username(s string) UsernameOpt { return UsernameOpt{s: s} }

func (o UsernameOpt) applyEngineLoginOption(a *engineLoginAct) error {
	if o.s == "" {
		return cfgErr("", "username", "Username", fmt.Errorf("%w: username", errEmpty))
	}
	a.opts.Username = o.s
	return nil
}

// PasswordOpt is the value returned by [Password].
// It implements [EngineLoginOption].
type PasswordOpt struct{ s string }

// Password sets the registry password for [Login]. The value is treated as secret in errors.
func Password(s string) PasswordOpt { return PasswordOpt{s: s} }

func (o PasswordOpt) applyEngineLoginOption(a *engineLoginAct) error {
	if o.s == "" {
		return secretErr("", "password", "Password", fmt.Errorf("%w: password", errEmpty))
	}
	a.opts.Password = o.s
	return nil
}

// ServerAddressOpt is the value returned by [ServerAddress].
// It implements [EngineLoginOption].
type ServerAddressOpt struct{ s string }

// ServerAddress sets the registry host for [Login].
func ServerAddress(s string) ServerAddressOpt { return ServerAddressOpt{s: s} }

func (o ServerAddressOpt) applyEngineLoginOption(a *engineLoginAct) error {
	if o.s == "" {
		return cfgErr("", "serveraddress", "ServerAddress", fmt.Errorf("%w: server address", errEmpty))
	}
	a.opts.ServerAddress = o.s
	return nil
}

// IdentityTokenOpt is the value returned by [IdentityToken].
// It implements [EngineLoginOption].
type IdentityTokenOpt struct{ s string }

// IdentityToken sets the identity token for [Login]. The value is treated as secret in errors.
func IdentityToken(s string) IdentityTokenOpt { return IdentityTokenOpt{s: s} }

func (o IdentityTokenOpt) applyEngineLoginOption(a *engineLoginAct) error {
	if o.s == "" {
		return secretErr("", "identitytoken", "IdentityToken", fmt.Errorf("%w: identity token", errEmpty))
	}
	a.opts.IdentityToken = o.s
	return nil
}

// RegistryTokenOpt is the value returned by [RegistryToken].
// It implements [EngineLoginOption].
type RegistryTokenOpt struct{ s string }

// RegistryToken sets the registry token for [Login]. The value is treated as secret in errors.
func RegistryToken(s string) RegistryTokenOpt { return RegistryTokenOpt{s: s} }

func (o RegistryTokenOpt) applyEngineLoginOption(a *engineLoginAct) error {
	if o.s == "" {
		return secretErr("", "registrytoken", "RegistryToken", fmt.Errorf("%w: registry token", errEmpty))
	}
	a.opts.RegistryToken = o.s
	return nil
}

// ContentOpt is the value returned by [Content].
// It implements [EngineConfigCreateOption], [EngineConfigUpdateOption], [EngineCopyToOption],
// [EnginePluginCreateOption], [EngineSecretCreateOption], and [EngineSecretUpdateOption].
type ContentOpt struct{ r io.Reader }

// Content supplies the archive body for [CopyTo], [CreatePlugin], or swarm secret/config data.
func Content(r io.Reader) ContentOpt { return ContentOpt{r: r} }

func (o ContentOpt) applyEngineCopyToOption(a *engineCopyToAct) error {
	a.opts.Content = o.r
	return nil
}

// PathOpt is the value returned by [Path].
// It implements [EngineStatPathOption], [IncludeOption], and [WatchTriggerOption].
type PathOpt struct{ p string }

// Path sets a filesystem path for [StatPath].
func Path(p string) PathOpt { return PathOpt{p: p} }

func (o PathOpt) applyEngineStatPathOption(a *engineStatPathAct) error {
	a.opts.Path = o.p
	return nil
}

// ContainerIDOpt is the value returned by [ContainerID].
// It implements [EngineNetworkConnectOption] and [EngineNetworkDisconnectOption].
type ContainerIDOpt struct{ id string }

// ContainerID selects a container for network connect/disconnect.
func ContainerID(id string) ContainerIDOpt { return ContainerIDOpt{id: id} }

func (o ContainerIDOpt) applyEngineNetworkConnectOption(a *engineNetworkConnectAct) error {
	a.opts.Container = o.id
	return nil
}
func (o ContainerIDOpt) applyEngineNetworkDisconnectOption(a *engineNetworkDisconnectAct) error {
	a.opts.Container = o.id
	return nil
}

var (
	_ EngineContainerListOption = Filter("status", "running")
	_ EngineContainerListOption = All(true)
	_ EngineContainerListOption = IncludeSize(true)
	_ EngineContainerListOption = Limit(1)
	_ EngineImageRemoveOption   = Force(true)
	_ EngineRenameOption        = Name("renamed")
	_ EngineCopyFromOption      = Source("/tmp/a")
	_ EngineCopyToOption        = Target("/tmp/b")
	_ EngineCopyToOption        = Content(nil)
	_ EnginePingOption          = NegotiateAPIVersion()
	_ EnginePingOption          = ForceNegotiate(true)
	_ EngineDiskUsageOption     = Verbose(true)
	_ EngineDiskUsageOption     = Images(true)
	_ EngineDiskUsageOption     = BuildCache(true)
	_ EngineDiskUsageOption     = Volumes(true)
	_ EngineDiskUsageOption     = Usage(df.Container)
	_ EngineLoginOption         = Username("user")
	_ EngineLoginOption         = Password("secret")
	_ EngineLoginOption         = ServerAddress("registry.example")
	_ EngineLoginOption         = IdentityToken("id")
	_ EngineLoginOption         = RegistryToken("reg")
	_ EngineEventsOption        = Since("1h")
	_ EngineEventsOption        = Until("now")
	_ EngineEventsOption        = Filter("type", "container")
)
