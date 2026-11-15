package kit

import (
	"github.com/aptd3v/containerkit/pkg/cap"
	"github.com/aptd3v/containerkit/pkg/cascade"
	dependpkg "github.com/aptd3v/containerkit/pkg/depend"
	"github.com/aptd3v/containerkit/pkg/df"
	"github.com/aptd3v/containerkit/pkg/endpoint"
	"github.com/aptd3v/containerkit/pkg/ipc"
	"github.com/aptd3v/containerkit/pkg/logdrv"
	"github.com/aptd3v/containerkit/pkg/netmode"
	"github.com/aptd3v/containerkit/pkg/oci"
	"github.com/aptd3v/containerkit/pkg/portmode"
	"github.com/aptd3v/containerkit/pkg/proto"
	"github.com/aptd3v/containerkit/pkg/pull"
	restartpkg "github.com/aptd3v/containerkit/pkg/restart"
	"github.com/aptd3v/containerkit/pkg/role"
	"github.com/aptd3v/containerkit/pkg/signal"
	"github.com/aptd3v/containerkit/pkg/svcmode"
	"github.com/aptd3v/containerkit/pkg/vol"
	"github.com/aptd3v/containerkit/pkg/wait"
	"github.com/aptd3v/containerkit/pkg/watch"
)

// Capability is a Linux capability name. Known values live in package cap
// (for example cap.NET_ADMIN, cap.ALL).
type Capability = cap.Capability

const ALL = cap.ALL

// Proto is the IP protocol vocabulary (tcp/udp/sctp). Known values live in
// package proto. The setter is [Protocol]; Go forbids a type and a function
// both named Protocol.
type Proto = proto.Proto

const (
	TCP  = proto.TCP
	UDP  = proto.UDP
	SCTP = proto.SCTP
)

// Network modes for [NetworkMode]. The daemon address setter is [Host]; the
// host network namespace is [NetworkHost].
const (
	Bridge      = netmode.Bridge
	NetworkHost = netmode.Host
	NetworkNone = netmode.None
)

// IPC modes for [IPC].
const (
	IPCShareable = ipc.Shareable
	IPCPrivate   = ipc.Private
	IPCHost      = ipc.Host
)

// JSONFile is the json-file log driver for [Logging] / [Driver].
const JSONFile = logdrv.JSONFile

// Port publish modes for [Mode] inside [Publish].
const (
	PortIngress = portmode.Ingress
	PortHost    = portmode.Host
)

// Deploy modes for [Mode] inside [Deploy].
const (
	Replicated = svcmode.Replicated
	Global     = svcmode.Global
)

// EndpointVIP is deploy.endpoint_mode vip for [EndpointMode].
const EndpointVIP = endpoint.VIP

// Common stop/kill signals for [StopSignal].
const (
	SIGTERM = signal.SIGTERM
	SIGKILL = signal.SIGKILL
)

// RestartPolicy is the Engine restart-policy vocabulary. Known values live in
// package restart.
type RestartPolicy = restartpkg.Policy

const (
	NoRestart     = restartpkg.No
	OnFailure     = restartpkg.OnFailure
	Always        = restartpkg.Always
	UnlessStopped = restartpkg.UnlessStopped
)

// DependCondition is a Compose depends_on condition. Known values live in
// package depend. The setter is [Condition].
type DependCondition = dependpkg.Condition

const (
	Started               = dependpkg.Started
	Healthy               = dependpkg.Healthy
	CompletedSuccessfully = dependpkg.CompletedSuccessfully
)

// VolumeKind is Engine mount type / Compose volume type. Known values live in
// package vol.
type VolumeKind = vol.Kind

const (
	Bind        = vol.Bind
	NamedVolume = vol.Volume
	Tmpfs       = vol.Tmpfs
	NamedPipe   = vol.Pipe
	Cluster     = vol.Cluster
	ImageVolume = vol.Image
)

// WaitCondition is an Engine wait-for-container condition. Known values live
// in package wait.
type WaitCondition = wait.Condition

const (
	NotRunning = wait.NotRunning
	NextExit   = wait.NextExit
	Removed    = wait.Removed
)

// PullMode is a Compose pull_policy value. Known values live in package pull.
type PullMode = pull.Mode

const (
	PullAlways       = pull.Always
	PullNever        = pull.Never
	PullMissing      = pull.Missing
	PullIfNotPresent = pull.IfNotPresent
	PullBuild        = pull.Build
	PullRefresh      = pull.Refresh
	PullDaily        = pull.Daily
	PullWeekly       = pull.Weekly
)

// PullEvery returns Compose pull_policy every_<period> (for example "24h").
func PullEvery(period string) PullMode { return pull.Every(period) }

// WatchAction is a Compose develop.watch action. Known values live in package watch.
type WatchAction = watch.Action

const (
	Sync         = watch.Sync
	Rebuild      = watch.Rebuild
	RestartSync  = watch.SyncRestart
	ExecSync     = watch.SyncExec
	WatchRestart = watch.Restart
)

// Swarm node roles. Known values live in package role.
const (
	NodeRoleManager NodeRole = NodeRole(role.Manager)
	NodeRoleWorker  NodeRole = NodeRole(role.Worker)
)

// Cascade is a Compose start/up OnExit action. Known values live in package cascade.
type Cascade = cascade.Action

const (
	CascadeIgnore = cascade.Ignore
	CascadeStop   = cascade.Stop
	CascadeFail   = cascade.Fail
)

// OCIVersion is a Compose publish OCI manifest version. Known values live in package oci.
type OCIVersion = oci.Version

const (
	OCI1_0 = oci.V1_0
	OCI1_1 = oci.V1_1
)

// DiskUsageObject is an Engine DiskUsage type filter. Known values live in package df.
type DiskUsageObject = df.Object

const (
	UsageContainer  = df.Container
	UsageImage      = df.Image
	UsageVolume     = df.Volume
	UsageBuildCache = df.BuildCache
)
