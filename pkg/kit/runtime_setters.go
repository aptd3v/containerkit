package kit

import (
	"fmt"
	"io"
	"strings"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
	"github.com/aptd3v/containerkit/pkg/cascade"
	"github.com/aptd3v/containerkit/pkg/oci"
	"github.com/aptd3v/containerkit/pkg/progress"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ForceOpt is the value returned by [Force].
// It implements [EngineBuildOption], [EngineImageRemoveOption],
// [EngineNetworkDisconnectOption], [EngineNodeRemoveOption], [EnginePluginDisableOption],
// [EnginePluginRemoveOption], [EngineSwarmLeaveOption], [EngineVolumeRemoveOption], and
// [RemoveOption].
type ForceOpt struct{ v bool }

// Force skips confirmation on remove, including explicit false.
func Force(v bool) ForceOpt { return ForceOpt{v: v} }

func (o ForceOpt) applyRemove(a *removeAct) error { a.force = o.v; return nil }

// VolumesOpt is the value returned by [Volumes].
// It implements [DownOption], [EngineDiskUsageOption], and [RemoveOption].
type VolumesOpt struct{ v bool }

// Volumes removes volumes on down/rm, or includes volumes on [DiskUsage],
// including explicit false.
func Volumes(v bool) VolumesOpt { return VolumesOpt{v: v} }

func (o VolumesOpt) applyDown(a *downAct) error     { a.volumes = o.v; return nil }
func (o VolumesOpt) applyRemove(a *removeAct) error { a.volumes = o.v; return nil }
func (o VolumesOpt) applyEngineDiskUsageOption(a *engineDiskUsageAct) error {
	a.opts.Volumes = o.v
	return nil
}

// LinksOpt is the value returned by [Links].
// It implements [RemoveOption].
type LinksOpt struct{ v bool }

// Links removes specified linked containers with [Container.Remove].
func Links(v bool) LinksOpt { return LinksOpt{v: v} }

func (o LinksOpt) applyRemove(a *removeAct) error { a.links = o.v; return nil }

// DetachOpt is the value returned by [Detach].
// It implements [EngineExecStartOption], [RunOption], and [UpOption].
type DetachOpt struct{ v bool }

// Detach runs without attaching, including explicit false.
func Detach(v bool) DetachOpt { return DetachOpt{v: v} }

func (o DetachOpt) applyRun(a *runAct) error { a.detach = o.v; return nil }
func (o DetachOpt) applyUp(a *upAct) error {
	if o.v {
		a.start.consumer = nil
	}
	return nil
}

// QuietOpt is the value returned by [Quiet].
// It implements [ComposeBuildOption], [ComposeCreateOption], [EngineBuildOption],
// [EngineImageLoadOption], [LoggerOption], [PullOption], [PushOption], [RunOption], [UpOption],
// and [WatchOption].
type QuietOpt struct{ v bool }

// Quiet suppresses progress output. On [NewLogger] it hides working events.
func Quiet(v bool) QuietOpt { return QuietOpt{v: v} }

func (o QuietOpt) applyUp(a *upAct) error {
	a.create.quietPull.assign(o.v)
	return nil
}
func (o QuietOpt) applyComposeCreate(a *createAct) error { a.quietPull.assign(o.v); return nil }
func (o QuietOpt) applyComposeBuild(a *composeBuildAct) error {
	a.quiet = o.v
	return nil
}
func (o QuietOpt) applyPull(a *pullAct) error { a.quiet = o.v; return nil }
func (o QuietOpt) applyPush(a *pushAct) error { a.quiet = o.v; return nil }

// FollowOpt is the value returned by [Follow].
// It implements [EngineServiceLogsOption], [EngineTaskLogsOption], and [LogsOption].
type FollowOpt struct{ v bool }

// Follow streams logs until cancelled.
func Follow(v bool) FollowOpt { return FollowOpt{v: v} }

func (o FollowOpt) applyLogs(a *logsAct) error { a.follow = o.v; return nil }

// TailOpt is the value returned by [Tail].
// It implements [EngineServiceLogsOption], [EngineTaskLogsOption], and [LogsOption].
type TailOpt struct{ n string }

// Tail limits log lines from the end ("all" or a count).
func Tail(n string) TailOpt { return TailOpt{n: n} }

func (o TailOpt) applyLogs(a *logsAct) error { a.tail = o.n; return nil }

// SinceOpt is the value returned by [Since].
// It implements [EngineContainerListOption], [EngineEventsOption], [EngineServiceLogsOption],
// [EngineTaskLogsOption], [EventsOption], and [LogsOption].
type SinceOpt struct{ s string }

// Since selects logs/events after a timestamp or duration.
func Since(s string) SinceOpt { return SinceOpt{s: s} }

func (o SinceOpt) applyLogs(a *logsAct) error     { a.since = o.s; return nil }
func (o SinceOpt) applyEvents(a *eventsAct) error { a.since = o.s; return nil }

// UntilOpt is the value returned by [Until].
// It implements [EngineEventsOption], [EngineServiceLogsOption], [EngineTaskLogsOption],
// [EventsOption], and [LogsOption].
type UntilOpt struct{ s string }

// Until selects logs/events before a timestamp or duration.
func Until(s string) UntilOpt { return UntilOpt{s: s} }

func (o UntilOpt) applyLogs(a *logsAct) error     { a.until = o.s; return nil }
func (o UntilOpt) applyEvents(a *eventsAct) error { a.until = o.s; return nil }

// TimestampsOpt is the value returned by [Timestamps].
// It implements [EngineServiceLogsOption], [EngineTaskLogsOption], [LoggerOption], and
// [LogsOption].
type TimestampsOpt struct{ v bool }

// Timestamps includes timestamps in container logs and [NewLogger] output.
func Timestamps(v bool) TimestampsOpt { return TimestampsOpt{v: v} }

func (o TimestampsOpt) applyLogs(a *logsAct) error { a.timestamps = o.v; return nil }

// DetailsOpt is the value returned by [Details].
// It implements [EngineServiceLogsOption], [EngineTaskLogsOption], and [LogsOption].
type DetailsOpt struct{ v bool }

// Details includes extra log attributes on Engine logs.
func Details(v bool) DetailsOpt { return DetailsOpt{v: v} }

func (o DetailsOpt) applyLogs(a *logsAct) error { a.details = o.v; return nil }

// AllOpt is the value returned by [All].
// It implements [CopyOption], [EngineBuildCachePruneOption], [EngineContainerListOption],
// [EngineImageListOption], [EngineImagePushOption], [EnginePullOption],
// [EngineVolumePruneOption], [KillOption], [ListOption], and [PsOption].
type AllOpt struct{ v bool }

// All includes stopped containers, every image tag, or every prune candidate.
// LoadProject unused resources use [AllResources].
func All(v bool) AllOpt { return AllOpt{v: v} }

func (o AllOpt) applyPs(a *psAct) error     { a.all = o.v; return nil }
func (o AllOpt) applyKill(a *killAct) error { a.all = o.v; return nil }
func (o AllOpt) applyList(a *listAct) error { a.all = o.v; return nil }
func (o AllOpt) applyCopy(a *copyAct) error { a.all = o.v; return nil }

// WaitOpt is the value returned by [Wait].
// It implements [ComposeStartOption] and [UpOption].
type WaitOpt struct{ v bool }

// Wait blocks until containers are running or healthy.
func Wait(v bool) WaitOpt { return WaitOpt{v: v} }

func (o WaitOpt) applyUp(a *upAct) error              { a.start.wait = o.v; return nil }
func (o WaitOpt) applyComposeStart(a *startAct) error { a.wait = o.v; return nil }

// RemoveOrphansOpt is the value returned by [RemoveOrphans].
// It implements [ComposeCreateOption], [DownOption], [KillOption], [RunOption], and [UpOption].
type RemoveOrphansOpt struct{ v bool }

// RemoveOrphans removes containers for undefined services.
func RemoveOrphans(v bool) RemoveOrphansOpt { return RemoveOrphansOpt{v: v} }

func (o RemoveOrphansOpt) applyUp(a *upAct) error                { a.create.removeOrphan = o.v; return nil }
func (o RemoveOrphansOpt) applyDown(a *downAct) error            { a.removeOrphan = o.v; return nil }
func (o RemoveOrphansOpt) applyComposeCreate(a *createAct) error { a.removeOrphan = o.v; return nil }
func (o RemoveOrphansOpt) applyKill(a *killAct) error            { a.removeOrphan = o.v; return nil }
func (o RemoveOrphansOpt) applyRun(a *runAct) error              { a.create.removeOrphan = o.v; return nil }

// NoDepsOpt is the value returned by [NoDeps].
// It implements [ComposeRestartOption] and [RunOption].
type NoDepsOpt struct{ v bool }

// NoDeps ignores service dependencies.
func NoDeps(v bool) NoDepsOpt { return NoDepsOpt{v: v} }

func (o NoDepsOpt) applyComposeRestart(a *restartAct) error { a.noDeps = o.v; return nil }
func (o NoDepsOpt) applyRun(a *runAct) error                { a.noDeps = o.v; return nil }

// IndexOpt is the value returned by [Index].
// It implements [AttachOption], [CommitOption], [ComposePortOption], [CopyOption],
// [ExportOption], [LogsOption], and [RunOption].
type IndexOpt struct{ n int }

// Index selects a replica index (1-based).
func Index(n int) IndexOpt { return IndexOpt{n: n} }

func (o IndexOpt) applyLogs(a *logsAct) error               { a.index = o.n; return nil }
func (o IndexOpt) applyRun(a *runAct) error                 { a.index = o.n; return nil }
func (o IndexOpt) applyCopy(a *copyAct) error               { a.index = o.n; return nil }
func (o IndexOpt) applyAttach(a *attachAct) error           { a.index = o.n; return nil }
func (o IndexOpt) applyComposePort(a *composePortAct) error { a.index = o.n; return nil }
func (o IndexOpt) applyExport(a *exportAct) error           { a.index = o.n; return nil }
func (o IndexOpt) applyCommit(a *commitAct) error           { a.index = o.n; return nil }

// AssumeYesOpt is the value returned by [AssumeYes].
// It implements [ComposeCreateOption], [PublishOption], [RunOption], and [UpOption].
type AssumeYesOpt struct{ v bool }

// AssumeYes answers yes to prompts.
func AssumeYes(v bool) AssumeYesOpt { return AssumeYesOpt{v: v} }

func (o AssumeYesOpt) applyUp(a *upAct) error                { a.create.assumeYes = o.v; return nil }
func (o AssumeYesOpt) applyComposeCreate(a *createAct) error { a.assumeYes = o.v; return nil }
func (o AssumeYesOpt) applyPublish(a *publishAct) error      { a.assumeYes = o.v; return nil }
func (o AssumeYesOpt) applyRun(a *runAct) error              { a.create.assumeYes = o.v; return nil }

// RecreateOpt is the value returned by [Recreate].
// It implements [ComposeCreateOption], [RunOption], and [UpOption].
type RecreateOpt struct{ s string }

// Recreate sets the recreate strategy on up/create ("never", "diverged", "force").
func Recreate(strategy string) RecreateOpt { return RecreateOpt{s: strategy} }

func (o RecreateOpt) applyUp(a *upAct) error { a.create.recreate = o.s; return nil }
func (o RecreateOpt) applyComposeCreate(a *createAct) error {
	a.recreate = o.s
	return nil
}
func (o RecreateOpt) applyRun(a *runAct) error { a.create.recreate = o.s; return nil }

// ServicesOpt is the value returned by [Services].
// It implements [ComposeBuildOption], [ComposeCreateOption], [ComposeRestartOption],
// [ComposeStartOption], [ComposeStopOption], [DownOption], [EventsOption], [ImagesOption],
// [KillOption], [LoadProjectOption], [LogsOption], [PauseOption], [PsOption], [RemoveOption],
// [ScaleOption], [UpOption], [VolumesListOption], [WaitOption], and [WatchOption].
type ServicesOpt struct{ names []string }

// Services selects service names for a Compose runtime action.
func Services(names ...string) ServicesOpt { return ServicesOpt{names: names} }

func (o ServicesOpt) set(dst *[]string) {
	*dst = append(*dst, o.names...)
}

func (o ServicesOpt) applyUp(a *upAct) error {
	o.set(&a.create.services)
	o.set(&a.start.services)
	return nil
}
func (o ServicesOpt) applyDown(a *downAct) error                 { o.set(&a.services); return nil }
func (o ServicesOpt) applyComposeBuild(a *composeBuildAct) error { o.set(&a.services); return nil }
func (o ServicesOpt) applyComposeCreate(a *createAct) error      { o.set(&a.services); return nil }
func (o ServicesOpt) applyComposeStart(a *startAct) error        { o.set(&a.services); return nil }
func (o ServicesOpt) applyComposeStop(a *stopAct) error          { o.set(&a.services); return nil }
func (o ServicesOpt) applyComposeRestart(a *restartAct) error    { o.set(&a.services); return nil }
func (o ServicesOpt) applyLogs(a *logsAct) error                 { o.set(&a.services); return nil }
func (o ServicesOpt) applyPs(a *psAct) error                     { o.set(&a.services); return nil }
func (o ServicesOpt) applyKill(a *killAct) error                 { o.set(&a.services); return nil }
func (o ServicesOpt) applyRemove(a *removeAct) error             { o.set(&a.services); return nil }
func (o ServicesOpt) applyPause(a *pauseAct) error               { o.set(&a.services); return nil }
func (o ServicesOpt) applyEvents(a *eventsAct) error             { o.set(&a.services); return nil }
func (o ServicesOpt) applyWatch(a *watchAct) error               { o.set(&a.services); return nil }
func (o ServicesOpt) applyWait(a *waitAct) error                 { o.set(&a.services); return nil }
func (o ServicesOpt) applyScale(a *scaleAct) error               { o.set(&a.services); return nil }
func (o ServicesOpt) applyImages(a *imagesAct) error             { o.set(&a.services); return nil }
func (o ServicesOpt) applyVolumesList(a *volumesListAct) error   { o.set(&a.services); return nil }

// RemoveImagesOpt is the value returned by [RemoveImages].
// It implements [DownOption].
type RemoveImagesOpt struct{ mode string }

// RemoveImages sets down image removal ("all" or "local").
func RemoveImages(mode string) RemoveImagesOpt { return RemoveImagesOpt{mode: mode} }

func (o RemoveImagesOpt) applyDown(a *downAct) error { a.images = o.mode; return nil }

// StopOpt is the value returned by [Stop].
// It implements [RemoveOption].
type StopOpt struct{ v bool }

// Stop first stops containers before remove.
func Stop(v bool) StopOpt { return StopOpt{v: v} }

func (o StopOpt) applyRemove(a *removeAct) error { a.stop = o.v; return nil }

// IgnoreFailuresOpt is the value returned by [IgnoreFailures].
// It implements [PullOption] and [PushOption].
type IgnoreFailuresOpt struct{ v bool }

// IgnoreFailures continues pull/push after errors.
func IgnoreFailures(v bool) IgnoreFailuresOpt { return IgnoreFailuresOpt{v: v} }

func (o IgnoreFailuresOpt) applyPull(a *pullAct) error { a.ignoreFailures = o.v; return nil }
func (o IgnoreFailuresOpt) applyPush(a *pushAct) error { a.ignoreFailures = o.v; return nil }

// ProgressOpt is the value returned by [Progress].
// It implements [ComposeBuildOption], [ComposeCreateOption], [LoggerOption], [UpOption], and
// [WatchOption].
type ProgressOpt struct{ s string }

// Progress sets compose build progress or [NewLogger] mode. Known values live
// in package progress (auto, plain, tty, quiet, json).
func Progress(s progress.Mode) ProgressOpt { return ProgressOpt{s: string(s)} }

func (o ProgressOpt) applyComposeBuild(a *composeBuildAct) error { a.progress = o.s; return nil }
func (o ProgressOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o ProgressOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o ProgressOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

// PushOpt is the value returned by [Push].
// It implements [ComposeBuildOption], [ComposeCreateOption], [UpOption], and [WatchOption].
type PushOpt struct{ v bool }

// Push pushes images after build.
func Push(v bool) PushOpt { return PushOpt{v: v} }

func (o PushOpt) applyComposeBuild(a *composeBuildAct) error { a.push = o.v; return nil }
func (o PushOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o PushOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o PushOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

// DepsOpt is the value returned by [Deps].
// It implements [ComposeBuildOption], [ComposeCreateOption], [UpOption], and [WatchOption].
type DepsOpt struct{ v bool }

// Deps also builds selected services' dependencies.
func Deps(v bool) DepsOpt { return DepsOpt{v: v} }

func (o DepsOpt) applyComposeBuild(a *composeBuildAct) error { a.deps = o.v; return nil }
func (o DepsOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o DepsOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o DepsOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

// BuilderOpt is the value returned by [Builder].
// It implements [ComposeBuildOption], [ComposeCreateOption], [UpOption], and [WatchOption].
type BuilderOpt struct{ s string }

// Builder selects a buildx builder for compose build.
func Builder(s string) BuilderOpt { return BuilderOpt{s: s} }

func (o BuilderOpt) applyComposeBuild(a *composeBuildAct) error { a.builder = o.s; return nil }
func (o BuilderOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o BuilderOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o BuilderOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

// OutputStreamOpt is the value returned by [OutputStream].
// It implements [ComposeBuildOption], [ComposeCreateOption], [ComposerClientOption],
// [LoggerOption], [UpOption], and [WatchOption].
type OutputStreamOpt struct{ w io.Writer }

// OutputStream sets the [NewLogger] destination or compose build progress writer.
func OutputStream(w io.Writer) OutputStreamOpt { return OutputStreamOpt{w: w} }

func (o OutputStreamOpt) applyComposeBuild(a *composeBuildAct) error { a.out = o.w; return nil }
func (o OutputStreamOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o OutputStreamOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o OutputStreamOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o OutputStreamOpt) applyComposerClient(a *composerAct) error {
	if o.w == nil {
		return cfgErr("", "output", "OutputStream", fmt.Errorf("%w: writer", errRequired))
	}
	a.out = o.w
	return nil
}

// DownOnExitOpt is the value returned by [DownOnExit].
// It implements [WaitOption].
type DownOnExitOpt struct{ v bool }

// DownOnExit runs down when a waited container exits.
func DownOnExit(v bool) DownOnExitOpt { return DownOnExitOpt{v: v} }

func (o DownOnExitOpt) applyWait(a *waitAct) error { a.downExit = o.v; return nil }

// PruneOpt is the value returned by [Prune].
// It implements [WatchOption].
type PruneOpt struct{ v bool }

// Prune removes dangling resources on watch.
func Prune(v bool) PruneOpt { return PruneOpt{v: v} }

func (o PruneOpt) applyWatch(a *watchAct) error { a.prune = o.v; return nil }

// CopyPathOpt is the value returned by [CopyPath].
// It implements [CopyOption].
type CopyPathOpt struct{ src, dst string }

// CopyPath sets source and destination for [Compose.Copy].
func CopyPath(source, destination string) CopyPathOpt {
	return CopyPathOpt{src: source, dst: destination}
}

func (o CopyPathOpt) applyCopy(a *copyAct) error {
	if o.src == "" || o.dst == "" {
		return cfgErr("", "copy", "CopyPath", fmt.Errorf("%w: source and destination", errRequired))
	}
	a.source = o.src
	a.destination = o.dst
	return nil
}

// ServiceNameOpt is the value returned by [ServiceName].
// It implements [AttachOption], [CommitOption], [ExportOption], and [RunOption].
type ServiceNameOpt struct{ s string }

// ServiceName selects a service for run/exec/attach/export/commit.
func ServiceName(s string) ServiceNameOpt { return ServiceNameOpt{s: s} }

func (o ServiceNameOpt) applyRun(a *runAct) error       { a.service = o.s; return nil }
func (o ServiceNameOpt) applyAttach(a *attachAct) error { a.service = o.s; return nil }
func (o ServiceNameOpt) applyExport(a *exportAct) error { a.service = o.s; return nil }
func (o ServiceNameOpt) applyCommit(a *commitAct) error { a.service = o.s; return nil }

// CheckpointOpt is the value returned by [Checkpoint].
// It implements [EngineCheckpointCreateOption], [EngineCheckpointRemoveOption], and
// [EngineStartOption].
type CheckpointOpt struct{ id string }

// Checkpoint starts from a checkpoint id.
func Checkpoint(id string) CheckpointOpt { return CheckpointOpt{id: id} }

func (o CheckpointOpt) applyEngineStart(a *engineStartAct) error { a.checkpoint = o.id; return nil }

// CheckpointDirOpt is the value returned by [CheckpointDir].
// It implements [EngineCheckpointCreateOption], [EngineCheckpointListOption],
// [EngineCheckpointRemoveOption], and [EngineStartOption].
type CheckpointDirOpt struct{ dir string }

// CheckpointDir sets the checkpoint directory.
func CheckpointDir(dir string) CheckpointDirOpt { return CheckpointDirOpt{dir: dir} }

func (o CheckpointDirOpt) applyEngineStart(a *engineStartAct) error {
	a.checkpointDir = o.dir
	return nil
}

// IncludeSizeOpt is the value returned by [IncludeSize].
// It implements [EngineContainerListOption] and [EngineInspectOption].
type IncludeSizeOpt struct{ v bool }

// IncludeSize includes filesystem size on inspect.
func IncludeSize(v bool) IncludeSizeOpt { return IncludeSizeOpt{v: v} }

func (o IncludeSizeOpt) applyEngineInspect(a *engineInspectAct) error { a.size = o.v; return nil }

// WaitForOpt is the value returned by [WaitFor].
// It implements [EngineWaitOption].
type WaitForOpt struct{ c WaitCondition }

// WaitFor sets the Engine wait condition.
func WaitFor(c WaitCondition) WaitForOpt { return WaitForOpt{c: c} }

func (o WaitForOpt) applyEngineWait(a *engineWaitAct) error { a.condition = o.c; return nil }

func (o TimeoutOpt) applyUp(a *upAct) error {
	a.create.timeout = durPtr(o.d)
	a.start.waitTimeout = o.d
	return nil
}
func (o TimeoutOpt) applyDown(a *downAct) error              { a.timeout = durPtr(o.d); return nil }
func (o TimeoutOpt) applyComposeCreate(a *createAct) error   { a.timeout = durPtr(o.d); return nil }
func (o TimeoutOpt) applyComposeStart(a *startAct) error     { a.waitTimeout = o.d; return nil }
func (o TimeoutOpt) applyComposeStop(a *stopAct) error       { a.timeout = durPtr(o.d); return nil }
func (o TimeoutOpt) applyComposeRestart(a *restartAct) error { a.timeout = durPtr(o.d); return nil }
func (o TimeoutOpt) applyRun(a *runAct) error                { a.create.timeout = durPtr(o.d); return nil }
func (o TimeoutOpt) applyEngineStop(a *engineStopAct) error {
	a.timeout = secPtr(o.d)
	return nil
}

func (o NameOpt) applyEngineCreate(a *engineCreateAct) error { a.name = o.name; return nil }
func (o NameOpt) applyRun(a *runAct) error                   { a.name = o.name; return nil }

func (o PullOpt) applyComposeBuild(a *composeBuildAct) error { a.pull = o.v; return nil }
func (o PullOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o PullOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o PullOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o PullOpt) applyRun(a *runAct) error { return o.applyComposeCreate(&a.create) }

func (o NoCacheOpt) applyComposeBuild(a *composeBuildAct) error { a.noCache = o.v; return nil }
func (o NoCacheOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o NoCacheOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o NoCacheOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o NoCacheOpt) applyRun(a *runAct) error { return o.applyComposeCreate(&a.create) }

func (o MemoryOpt) applyComposeBuild(a *composeBuildAct) error { a.memory = o.n; return nil }
func (o MemoryOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o MemoryOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o MemoryOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

func (o AutoRemoveOpt) applyRun(a *runAct) error { a.autoRm = o.v; return nil }

func (o PrivilegedOpt) applyRun(a *runAct) error { a.priv = o.v; return nil }

func (o EnvOpt) applyRun(a *runAct) error {
	a.env = append(a.env, o.key+"="+o.value)
	return nil
}

func (o EnvOpt) applyComposeBuild(a *composeBuildAct) error {
	if o.key == "" {
		return secretErr("", "env", "Env", fmt.Errorf("%w: build arg key", errEmpty))
	}
	if a.args == nil {
		a.args = composetypes.MappingWithEquals{}
	}
	v := o.value
	a.args[o.key] = &v
	return nil
}
func (o EnvOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o EnvOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o EnvOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

func (o LabelOpt) applyRun(a *runAct) error {
	if a.labels == nil {
		a.labels = composetypes.Labels{}
	}
	a.labels[o.key] = o.value
	return nil
}

func (o ProtocolOpt) applyComposePort(a *composePortAct) error {
	a.protocol = string(o.p)
	return nil
}

// OutputFileOpt is the value returned by [OutputFile].
// It implements [ExportOption].
type OutputFileOpt struct{ path string }

// OutputFile sets the archive path for [Compose.Export].
func OutputFile(path string) OutputFileOpt { return OutputFileOpt{path: path} }

func (o OutputFileOpt) applyExport(a *exportAct) error {
	if o.path == "" {
		return cfgErr("", "export", "OutputFile", fmt.Errorf("%w: output file", errEmpty))
	}
	a.output = o.path
	return nil
}

func (o AttachOpt) applyRun(a *runAct) error {
	a.detach = !o.v
	return nil
}

func (o AttachStdoutOpt) applyLogs(a *logsAct) error { a.stdout.assign(o.v); return nil }
func (o AttachStderrOpt) applyLogs(a *logsAct) error { a.stderr.assign(o.v); return nil }

func (o AllOpt) applyEnginePull(a *enginePullAct) error { a.all = o.v; return nil }

func (o QuietOpt) applyEngineBuild(a *engineBuildAct) error { a.quiet = o.v; return nil }
func (o QuietOpt) applyRun(a *runAct) error                 { a.create.quietPull.assign(o.v); return nil }
func (o QuietOpt) applyWatch(a *watchAct) error {
	a.ensureBuild().quiet = o.v
	return nil
}

func (o NoCacheOpt) applyEngineBuild(a *engineBuildAct) error { a.noCache = o.v; return nil }
func (o PullOpt) applyEngineBuild(a *engineBuildAct) error    { a.pull = o.v; return nil }
func (o MemoryOpt) applyEngineBuild(a *engineBuildAct) error  { a.memory = o.n; return nil }
func (o ForceOpt) applyEngineBuild(a *engineBuildAct) error   { a.forceRm = o.v; return nil }

func (o ImageOpt) applyEngineBuild(a *engineBuildAct) error {
	a.tags = append(a.tags, o.image)
	return nil
}
func (o TagOpt) applyEngineBuild(a *engineBuildAct) error {
	if o.t == "" {
		return cfgErr("", "tag", "Tag", fmt.Errorf("%w: tag", errEmpty))
	}
	a.tags = append(a.tags, o.t)
	return nil
}

func (o DockerfilePathOpt) applyEngineBuild(a *engineBuildAct) error {
	if o.path == "" {
		return cfgErr("", "dockerfile", "DockerfilePath", fmt.Errorf("%w: dockerfile", errEmpty))
	}
	a.dockerfile = o.path
	return nil
}

// TarOpt is the value returned by [Tar].
// It implements [EngineBuildOption].
type TarOpt struct{ r io.Reader }

// Tar supplies a pre-packed build-context archive for [BuildImage].
func Tar(r io.Reader) TarOpt { return TarOpt{r: r} }

func (o TarOpt) applyEngineBuild(a *engineBuildAct) error {
	a.archive = o.r
	return nil
}
func (o TargetOpt) applyEngineBuild(a *engineBuildAct) error {
	a.target = o.s
	return nil
}

func (o EnvOpt) applyEngineBuild(a *engineBuildAct) error {
	if a.args == nil {
		a.args = map[string]*string{}
	}
	v := o.value
	a.args[o.key] = &v
	return nil
}

func (o LabelOpt) applyEngineBuild(a *engineBuildAct) error {
	if a.labels == nil {
		a.labels = map[string]string{}
	}
	a.labels[o.key] = o.value
	return nil
}
func (o LabelOpt) applyEngineNetwork(a *engineNetworkAct) error {
	if a.labels == nil {
		a.labels = map[string]string{}
	}
	a.labels[o.key] = o.value
	return nil
}
func (o LabelOpt) applyEngineVolume(a *engineVolumeAct) error {
	if a.labels == nil {
		a.labels = map[string]string{}
	}
	a.labels[o.key] = o.value
	return nil
}

func (o DriverValue) applyEngineNetwork(a *engineNetworkAct) error { a.driver = o.s; return nil }
func (o DriverValue) applyEngineVolume(a *engineVolumeAct) error   { a.driver = o.s; return nil }

func (o NameOpt) applyEngineVolume(a *engineVolumeAct) error { a.name = o.name; return nil }
func (o NameOpt) applyGenerate(a *generateAct) error         { a.name = o.name; return nil }

func (o DriverOptKV) applyEngineNetwork(a *engineNetworkAct) error {
	if o.key == "" {
		return cfgErr("", "option", "DriverOpt", fmt.Errorf("%w: network option key", errEmpty))
	}
	if a.options == nil {
		a.options = map[string]string{}
	}
	a.options[o.key] = o.value
	return nil
}
func (o DriverOptKV) applyEngineVolume(a *engineVolumeAct) error {
	if o.key == "" {
		return cfgErr("", "option", "DriverOpt", fmt.Errorf("%w: volume option key", errEmpty))
	}
	if a.DriverValue == nil {
		a.DriverValue = map[string]string{}
	}
	a.DriverValue[o.key] = o.value
	return nil
}

func (o InternalOpt) applyEngineNetwork(a *engineNetworkAct) error {
	a.internal = o.v
	return nil
}
func (o AttachableOpt) applyEngineNetwork(a *engineNetworkAct) error {
	a.attachable = o.v
	return nil
}

func (o PlatformOpt) applyEnginePull(a *enginePullAct) error {
	p, err := parseOCIPlatform(o.p)
	if err != nil {
		return err
	}
	a.platforms = append(a.platforms, p)
	return nil
}
func (o PlatformOpt) applyEngineBuild(a *engineBuildAct) error {
	p, err := parseOCIPlatform(o.p)
	if err != nil {
		return err
	}
	a.platforms = append(a.platforms, p)
	return nil
}

// InheritOpt is the value returned by [Inherit].
// It implements [ComposeCreateOption], [RunOption], and [UpOption].
type InheritOpt struct{ v bool }

// Inherit reuses anonymous volumes from the previous container.
func Inherit(v bool) InheritOpt { return InheritOpt{v: v} }

func (o InheritOpt) applyUp(a *upAct) error { a.create.inherit.assign(o.v); return nil }
func (o InheritOpt) applyComposeCreate(a *createAct) error {
	a.inherit.assign(o.v)
	return nil
}
func (o InheritOpt) applyRun(a *runAct) error { a.create.inherit.assign(o.v); return nil }

// IgnoreOrphansOpt is the value returned by [IgnoreOrphans].
// It implements [ComposeCreateOption], [RunOption], and [UpOption].
type IgnoreOrphansOpt struct{ v bool }

// IgnoreOrphans leaves containers for undefined services in place.
func IgnoreOrphans(v bool) IgnoreOrphansOpt { return IgnoreOrphansOpt{v: v} }

func (o IgnoreOrphansOpt) applyUp(a *upAct) error { a.create.ignoreOrphan = o.v; return nil }
func (o IgnoreOrphansOpt) applyComposeCreate(a *createAct) error {
	a.ignoreOrphan = o.v
	return nil
}
func (o IgnoreOrphansOpt) applyRun(a *runAct) error { a.create.ignoreOrphan = o.v; return nil }

// IgnoreBuildableOpt is the value returned by [IgnoreBuildable].
// It implements [PullOption].
type IgnoreBuildableOpt struct{ v bool }

// IgnoreBuildable skips images that can be built locally on pull.
func IgnoreBuildable(v bool) IgnoreBuildableOpt { return IgnoreBuildableOpt{v: v} }

func (o IgnoreBuildableOpt) applyPull(a *pullAct) error { a.ignoreBuildable = o.v; return nil }

// ImageMandatoryOpt is the value returned by [ImageMandatory].
// It implements [PushOption].
type ImageMandatoryOpt struct{ v bool }

// ImageMandatory fails push when a service has no image name.
func ImageMandatory(v bool) ImageMandatoryOpt { return ImageMandatoryOpt{v: v} }

func (o ImageMandatoryOpt) applyPush(a *pushAct) error { a.imageMandatory = o.v; return nil }

// RecreateDependenciesOpt is the value returned by [RecreateDependencies].
// It implements [ComposeCreateOption], [RunOption], and [UpOption].
type RecreateDependenciesOpt struct{ s string }

// RecreateDependencies sets the recreate strategy for dependency services.
func RecreateDependencies(strategy string) RecreateDependenciesOpt {
	return RecreateDependenciesOpt{s: strategy}
}

func (o RecreateDependenciesOpt) applyUp(a *upAct) error { a.create.recreateDeps = o.s; return nil }
func (o RecreateDependenciesOpt) applyComposeCreate(a *createAct) error {
	a.recreateDeps = o.s
	return nil
}
func (o RecreateDependenciesOpt) applyRun(a *runAct) error { a.create.recreateDeps = o.s; return nil }

// EventConsumerOpt is the value returned by [EventConsumer].
// It implements [EventsOption].
type EventConsumerOpt struct {
	fn func(composeapi.Event) error
}

// EventConsumer receives Compose events from [Compose.Events].
func EventConsumer(fn func(composeapi.Event) error) EventConsumerOpt {
	return EventConsumerOpt{fn: fn}
}

func (o EventConsumerOpt) applyEvents(a *eventsAct) error { a.consumer = o.fn; return nil }

// ConsumeLogsOpt is the value returned by [ConsumeLogs].
// It implements [ComposeStartOption], [UpOption], and [WatchOption].
type ConsumeLogsOpt struct{ c composeapi.LogConsumer }

// ConsumeLogs attaches a log consumer on up/start/watch.
func ConsumeLogs(c composeapi.LogConsumer) ConsumeLogsOpt { return ConsumeLogsOpt{c: c} }

func (o ConsumeLogsOpt) applyUp(a *upAct) error { a.start.consumer = o.c; return nil }
func (o ConsumeLogsOpt) applyComposeStart(a *startAct) error {
	a.consumer = o.c
	return nil
}
func (o ConsumeLogsOpt) applyWatch(a *watchAct) error { a.logTo = o.c; return nil }

// ContainersOpt is the value returned by [Containers].
// It implements [GenerateOption].
type ContainersOpt struct{ ids []string }

// Containers selects Engine container IDs for [Generate].
func Containers(ids ...string) ContainersOpt { return ContainersOpt{ids: ids} }

func (o ContainersOpt) applyGenerate(a *generateAct) error {
	a.containers = append(a.containers, o.ids...)
	return nil
}

// FollowLinkOpt is the value returned by [FollowLink].
// It implements [CopyOption].
type FollowLinkOpt struct{ v bool }

// FollowLink follows symlinks on [Compose.Copy].
func FollowLink(v bool) FollowLinkOpt { return FollowLinkOpt{v: v} }

func (o FollowLinkOpt) applyCopy(a *copyAct) error { a.followLink = o.v; return nil }

// CopyUIDGIDOpt is the value returned by [CopyUIDGID].
// It implements [CopyOption] and [EngineCopyToOption].
type CopyUIDGIDOpt struct{ v bool }

// CopyUIDGID preserves uid/gid on [Compose.Copy].
func CopyUIDGID(v bool) CopyUIDGIDOpt { return CopyUIDGIDOpt{v: v} }

func (o CopyUIDGIDOpt) applyCopy(a *copyAct) error { a.copyUIDGID = o.v; return nil }

// DetachKeysOpt is the value returned by [DetachKeys].
// It implements [AttachOption], [EngineAttachOption], and [EngineExecCreateOption].
type DetachKeysOpt struct{ s string }

// DetachKeys sets the attach detach key sequence.
func DetachKeys(s string) DetachKeysOpt { return DetachKeysOpt{s: s} }

func (o DetachKeysOpt) applyAttach(a *attachAct) error { a.detachKeys = o.s; return nil }

// NoStdinOpt is the value returned by [NoStdin].
// It implements [AttachOption].
type NoStdinOpt struct{ v bool }

// NoStdin disables stdin on attach.
func NoStdin(v bool) NoStdinOpt { return NoStdinOpt{v: v} }

func (o NoStdinOpt) applyAttach(a *attachAct) error { a.noStdin = o.v; return nil }

// ProxyOpt is the value returned by [Proxy].
// It implements [AttachOption].
type ProxyOpt struct{ v bool }

// Proxy enables attach stream proxying.
func Proxy(v bool) ProxyOpt { return ProxyOpt{v: v} }

func (o ProxyOpt) applyAttach(a *attachAct) error { a.proxy = o.v; return nil }

// AuthorOpt is the value returned by [Author].
// It implements [CommitOption] and [EngineCommitOption].
type AuthorOpt struct{ s string }

// Author sets the image author on [Compose.Commit].
func Author(s string) AuthorOpt { return AuthorOpt{s: s} }

func (o AuthorOpt) applyCommit(a *commitAct) error { a.author = o.s; return nil }

// CommentOpt is the value returned by [Comment].
// It implements [CommitOption] and [EngineCommitOption].
type CommentOpt struct{ s string }

// Comment sets the commit message on [Compose.Commit].
func Comment(s string) CommentOpt { return CommentOpt{s: s} }

func (o CommentOpt) applyCommit(a *commitAct) error { a.comment = o.s; return nil }

// ReferenceOpt is the value returned by [Reference].
// It implements [CommitOption] and [EngineCommitOption].
type ReferenceOpt struct{ s string }

// Reference sets the commit image reference.
func Reference(s string) ReferenceOpt { return ReferenceOpt{s: s} }

func (o ReferenceOpt) applyCommit(a *commitAct) error { a.reference = o.s; return nil }

// PauseOpt is the value returned by [Pause].
// It implements [CommitOption].
type PauseOpt struct{ v bool }

// Pause pauses the container during [Compose.Commit], including explicit false.
func Pause(v bool) PauseOpt { return PauseOpt{v: v} }

func (o PauseOpt) applyCommit(a *commitAct) error { a.pause.assign(o.v); return nil }

// IncludeNetworksOpt is the value returned by [IncludeNetworks].
// It implements [VizOption].
type IncludeNetworksOpt struct{ v bool }

// IncludeNetworks includes network names in [Compose.Viz].
func IncludeNetworks(v bool) IncludeNetworksOpt { return IncludeNetworksOpt{v: v} }

func (o IncludeNetworksOpt) applyViz(a *vizAct) error { a.networks = o.v; return nil }

// IncludePortsOpt is the value returned by [IncludePorts].
// It implements [VizOption].
type IncludePortsOpt struct{ v bool }

// IncludePorts includes published ports in [Compose.Viz].
func IncludePorts(v bool) IncludePortsOpt { return IncludePortsOpt{v: v} }

func (o IncludePortsOpt) applyViz(a *vizAct) error { a.ports = o.v; return nil }

// IncludeImageOpt is the value returned by [IncludeImage].
// It implements [VizOption].
type IncludeImageOpt struct{ v bool }

// IncludeImage includes image names in [Compose.Viz].
func IncludeImage(v bool) IncludeImageOpt { return IncludeImageOpt{v: v} }

func (o IncludeImageOpt) applyViz(a *vizAct) error { a.image = o.v; return nil }

// IndentOpt is the value returned by [Indent].
// It implements [VizOption].
type IndentOpt struct{ s string }

// Indent sets graphviz indentation for [Compose.Viz].
func Indent(s string) IndentOpt { return IndentOpt{s: s} }

func (o IndentOpt) applyViz(a *vizAct) error { a.indent = o.s; return nil }

// ResolveDigestsOpt is the value returned by [ResolveDigests].
// It implements [PublishOption].
type ResolveDigestsOpt struct{ v bool }

// ResolveDigests resolves image references to digests on publish.
func ResolveDigests(v bool) ResolveDigestsOpt { return ResolveDigestsOpt{v: v} }

func (o ResolveDigestsOpt) applyPublish(a *publishAct) error { a.resolveDigest = o.v; return nil }

// ApplicationOpt is the value returned by [Application].
// It implements [PublishOption].
type ApplicationOpt struct{ v bool }

// Application publishes as an application artifact.
func Application(v bool) ApplicationOpt { return ApplicationOpt{v: v} }

func (o ApplicationOpt) applyPublish(a *publishAct) error { a.application = o.v; return nil }

// WithEnvironmentOpt is the value returned by [WithEnvironment].
// It implements [PublishOption].
type WithEnvironmentOpt struct{ v bool }

// WithEnvironment includes environment on publish.
func WithEnvironment(v bool) WithEnvironmentOpt { return WithEnvironmentOpt{v: v} }

func (o WithEnvironmentOpt) applyPublish(a *publishAct) error { a.withEnv = o.v; return nil }

// RegistryAuthOpt is the value returned by [RegistryAuth].
// It implements [EngineImagePushOption], [EngineImageSearchOption],
// [EnginePluginInstallOption], [EnginePluginPushOption], [EnginePluginUpgradeOption],
// [EnginePullOption], [EngineServiceCreateOption], and [EngineServiceUpdateOption].
type RegistryAuthOpt struct{ s string }

// RegistryAuth sets registry credentials for [PullImage] and swarm service create/update.
func RegistryAuth(s string) RegistryAuthOpt { return RegistryAuthOpt{s: s} }

func (o RegistryAuthOpt) applyEnginePull(a *enginePullAct) error { a.auth = o.s; return nil }

// PrintOpt is the value returned by [Print].
// It implements [ComposeBuildOption], [ComposeCreateOption], [UpOption], and [WatchOption].
type PrintOpt struct{ v bool }

// Print prints the build config without building.
func Print(v bool) PrintOpt { return PrintOpt{v: v} }

func (o PrintOpt) applyComposeBuild(a *composeBuildAct) error { a.print = o.v; return nil }
func (o PrintOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o PrintOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o PrintOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

// CheckOpt is the value returned by [Check].
// It implements [ComposeBuildOption], [ComposeCreateOption], [UpOption], and [WatchOption].
type CheckOpt struct{ v bool }

// Check validates build configuration without building.
func Check(v bool) CheckOpt { return CheckOpt{v: v} }

func (o CheckOpt) applyComposeBuild(a *composeBuildAct) error { a.check = o.v; return nil }
func (o CheckOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o CheckOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o CheckOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

// UseAliasesOpt is the value returned by [UseAliases].
// It implements [RunOption].
type UseAliasesOpt struct{ v bool }

// UseAliases uses network aliases on [Compose.Run].
func UseAliases(v bool) UseAliasesOpt { return UseAliasesOpt{v: v} }

func (o UseAliasesOpt) applyRun(a *runAct) error { a.useAlias = o.v; return nil }

// AttachToOpt is the value returned by [AttachTo].
// It implements [ComposeStartOption] and [UpOption].
type AttachToOpt struct{ names []string }

// AttachTo selects services to attach on start/up.
func AttachTo(names ...string) AttachToOpt { return AttachToOpt{names: names} }

func (o AttachToOpt) applyUp(a *upAct) error {
	a.start.attachTo = append(a.start.attachTo, o.names...)
	return nil
}
func (o AttachToOpt) applyComposeStart(a *startAct) error {
	a.attachTo = append(a.attachTo, o.names...)
	return nil
}

// ExitFromOpt is the value returned by [ExitFrom].
// It implements [ComposeStartOption] and [UpOption].
type ExitFromOpt struct{ s string }

// ExitFrom returns the exit code of the named service on up/start.
func ExitFrom(service string) ExitFromOpt { return ExitFromOpt{s: service} }

func (o ExitFromOpt) applyUp(a *upAct) error              { a.start.exitFrom = o.s; return nil }
func (o ExitFromOpt) applyComposeStart(a *startAct) error { a.exitFrom = o.s; return nil }

// WatchChangesOpt is the value returned by [WatchChanges].
// It implements [ComposeStartOption] and [UpOption].
type WatchChangesOpt struct{ v bool }

// WatchChanges enables file watch during start/up.
func WatchChanges(v bool) WatchChangesOpt { return WatchChangesOpt{v: v} }

func (o WatchChangesOpt) applyUp(a *upAct) error              { a.start.watch = o.v; return nil }
func (o WatchChangesOpt) applyComposeStart(a *startAct) error { a.watch = o.v; return nil }

// OnExitOpt is the value returned by [OnExit].
// It implements [ComposeStartOption] and [UpOption].
type OnExitOpt struct{ a cascade.Action }

// OnExit sets cascade behavior when a container stops on up/start.
func OnExit(a cascade.Action) OnExitOpt { return OnExitOpt{a: a} }

func (o OnExitOpt) applyUp(a *upAct) error {
	a.start.onExit = composeapi.Cascade(o.a)
	return nil
}
func (o OnExitOpt) applyComposeStart(a *startAct) error {
	a.onExit = composeapi.Cascade(o.a)
	return nil
}

// NavigationMenuOpt is the value returned by [NavigationMenu].
// It implements [ComposeStartOption] and [UpOption].
type NavigationMenuOpt struct{ v bool }

// NavigationMenu toggles the compose navigation menu on up/start, including explicit false.
func NavigationMenu(v bool) NavigationMenuOpt { return NavigationMenuOpt{v: v} }

func (o NavigationMenuOpt) applyUp(a *upAct) error              { a.start.navMenu = o.v; return nil }
func (o NavigationMenuOpt) applyComposeStart(a *startAct) error { a.navMenu = o.v; return nil }

// AttestationsOpt is the value returned by [Attestations].
// It implements [ComposeBuildOption], [ComposeCreateOption], [UpOption], and [WatchOption].
type AttestationsOpt struct{ v bool }

// Attestations enables build attestations on [Compose.Build], including explicit false.
func Attestations(v bool) AttestationsOpt { return AttestationsOpt{v: v} }

func (o AttestationsOpt) applyComposeBuild(a *composeBuildAct) error {
	a.attestations = o.v
	return nil
}
func (o AttestationsOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o AttestationsOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o AttestationsOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

// OCIOpt is the value returned by [OCI].
// It implements [PublishOption].
type OCIOpt struct{ v oci.Version }

// OCI sets the publish OCI manifest version.
func OCI(v oci.Version) OCIOpt { return OCIOpt{v: v} }

func (o OCIOpt) applyPublish(a *publishAct) error {
	a.ociVersion = composeapi.OCIVersion(o.v)
	return nil
}

// ChangeOpt is the value returned by [Change].
// It implements [CommitOption].
type ChangeOpt struct{ s string }

// Change adds a Dockerfile instruction applied on [Compose.Commit].
func Change(instruction string) ChangeOpt { return ChangeOpt{s: instruction} }

func (o ChangeOpt) applyCommit(a *commitAct) error {
	if o.s == "" {
		return cfgErr("", "changes", "Change", fmt.Errorf("%w: commit change", errEmpty))
	}
	a.changes = append(a.changes, o.s)
	return nil
}

func (o SSHOpt) applyComposeBuild(a *composeBuildAct) error {
	if o.id == "" {
		return cfgErr("", "ssh", "SSH", fmt.Errorf("%w: ssh id", errEmpty))
	}
	a.sshs = append(a.sshs, composetypes.SSHKey{ID: o.id, Path: o.path})
	return nil
}
func (o SSHOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o SSHOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o SSHOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

func (o ProvenanceOpt) applyComposeBuild(a *composeBuildAct) error {
	a.provenance = o.s
	return nil
}
func (o ProvenanceOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o ProvenanceOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o ProvenanceOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

func (o SBOMOpt) applyComposeBuild(a *composeBuildAct) error {
	a.sbom = o.s
	return nil
}
func (o SBOMOpt) applyComposeCreate(a *createAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}
func (o SBOMOpt) applyUp(a *upAct) error { return o.applyComposeCreate(&a.create) }
func (o SBOMOpt) applyWatch(a *watchAct) error {
	return o.applyComposeBuild(a.ensureBuild())
}

func parseOCIPlatform(s string) (ocispec.Platform, error) {
	var zero ocispec.Platform
	if s == "" {
		return zero, cfgErr("", "platform", "Platform", fmt.Errorf("%w: platform", errEmpty))
	}
	parts := strings.Split(s, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return zero, cfgErr("", "platform", "Platform", fmt.Errorf("%w: platform %q", errInvalid, s))
	}
	p := ocispec.Platform{OS: parts[0], Architecture: parts[1]}
	if len(parts) > 2 {
		p.Variant = parts[2]
	}
	return p, nil
}

var (
	_ UpOption            = OnExit(cascade.Stop)
	_ ComposeStartOption  = OnExit(cascade.Fail)
	_ UpOption            = NavigationMenu(true)
	_ ComposeStartOption  = NavigationMenu(true)
	_ ComposeBuildOption  = Attestations(true)
	_ ComposeBuildOption  = SSH("default", "")
	_ ComposeBuildOption  = Provenance("mode=min")
	_ ComposeBuildOption  = SBOM("mode=min")
	_ ComposeBuildOption  = Env("FOO", "bar")
	_ ComposeCreateOption = Pull(true)
	_ ComposeCreateOption = SSH("default", "")
	_ WatchOption         = Pull(true)
	_ WatchOption         = Attestations(true)
	_ PublishOption       = OCI(oci.V1_1)
	_ CommitOption        = Change("CMD /bin/true")
	_ RunOption           = Recreate("force")
	_ RunOption           = Timeout(0)
)
