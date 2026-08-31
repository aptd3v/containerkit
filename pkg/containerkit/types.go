package containerkit

import (
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/hc"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/hc/mount"
	"github.com/compose-spec/compose-go/v2/types"
)

// Capability is a Linux capability granted or dropped on a container.
type Capability = hc.Capability

const (
	AUDIT_WRITE        = hc.AUDIT_WRITE
	CHOWN              = hc.CHOWN
	DAC_OVERRIDE       = hc.DAC_OVERRIDE
	FOWNER             = hc.FOWNER
	FSETID             = hc.FSETID
	KILL               = hc.KILL
	MKNOD              = hc.MKNOD
	NET_BIND_SERVICE   = hc.NET_BIND_SERVICE
	NET_RAW            = hc.NET_RAW
	SETFCAP            = hc.SETFCAP
	SETGID             = hc.SETGID
	SETPCAP            = hc.SETPCAP
	SETUID             = hc.SETUID
	SYS_CHROOT         = hc.SYS_CHROOT
	AUDIT_CONTROL      = hc.AUDIT_CONTROL
	AUDIT_READ         = hc.AUDIT_READ
	BLOCK_SUSPEND      = hc.BLOCK_SUSPEND
	BPF                = hc.BPF
	CHECKPOINT_RESTORE = hc.CHECKPOINT_RESTORE
	DAC_READ_SEARCH    = hc.DAC_READ_SEARCH
	SYS_ADMIN          = hc.SYS_ADMIN
	SYS_BOOT           = hc.SYS_BOOT
	SYS_MODULE         = hc.SYS_MODULE
	SYS_PACCT          = hc.SYS_PACCT
	SYS_RAWIO          = hc.SYS_RAWIO
	SYS_TIME           = hc.SYS_TIME
	SYS_TTY_CONFIG     = hc.SYS_TTY_CONFIG
	SYSLOG             = hc.SYSLOG
	IPC_LOCK           = hc.IPC_LOCK
	IPC_OWNER          = hc.IPC_OWNER
	LEASE              = hc.LEASE
	LINUX_IMMUTABLE    = hc.LINUX_IMMUTABLE
	SYS_NICE           = hc.SYS_NICE
	SYS_PTRACE         = hc.SYS_PTRACE
	SYS_RESOURCE       = hc.SYS_RESOURCE
	WAKE_ALARM         = hc.WAKE_ALARM
	NET_ADMIN          = hc.NET_ADMIN
	NET_BROADCAST      = hc.NET_BROADCAST
	MAC_ADMIN          = hc.MAC_ADMIN
	MAC_OVERRIDE       = hc.MAC_OVERRIDE
	PERFMON            = hc.PERFMON
	ALL                = hc.ALL
)

// RestartPolicy is a Docker restart policy name.
type RestartPolicy = hc.RestartPolicy

const (
	RestartPolicyNo            = hc.RestartPolicyNo
	RestartPolicyOnFailure     = hc.RestartPolicyOnFailure
	RestartPolicyAlways        = hc.RestartPolicyAlways
	RestartPolicyUnlessStopped = hc.RestartPolicyUnlessStopped
)

// MountPropagation is a bind-mount propagation mode.
type MountPropagation = mount.Propagation

const (
	PropagationRPrivate = mount.PropagationRPrivate
	PropagationPrivate  = mount.PropagationPrivate
	PropagationRShared  = mount.PropagationRShared
	PropagationShared   = mount.PropagationShared
	PropagationRSlave   = mount.PropagationRSlave
	PropagationSlave    = mount.PropagationSlave
)

// MountType is a volume/bind/tmpfs mount type.
type MountType = mount.MountType

const (
	MountBind      = mount.MountTypeBind
	MountVolume    = mount.MountTypeVolume
	MountTmpfs     = mount.MountTypeTmpfs
	MountNamedPipe = mount.MountTypeNamedPipe
)

// Health is a container healthcheck. Timeout/Interval/StartPeriod are seconds
// unless the matching *D field is a duration string ("2s", "1m").
type Health struct {
	Test         []string
	Timeout      int
	Interval     int
	StartPeriod  int
	Retries      int
	TimeoutD     string
	IntervalD    string
	StartPeriodD string
}

// Mount describes a single host mount point.
type Mount struct {
	Source   string
	Target   string
	Type     MountType
	ReadOnly bool
}

// EndpointOpts are optional settings for a network endpoint.
type EndpointOpts struct {
	Aliases       []string
	Links         []string
	MacAddress    string
	DriverOptions map[string]string
	IPv4          string
	IPv6          string
	LinkLocalIPs  []string
}

// BuildSpec is Compose service build configuration.
type BuildSpec struct {
	Context          string
	Dockerfile       string
	DockerfileInline string
	Args             map[string]string
	Labels           map[string]string
	Tags             []string
	Target           string
	Network          string
	NoCache          bool
	Pull             bool
	Privileged       bool
	Platforms        []string
	CacheFrom        []string
	CacheTo          []string
}

func (b BuildSpec) apply(cfg *types.ServiceConfig) {
	if cfg.Build == nil {
		cfg.Build = &types.BuildConfig{}
	}
	bc := cfg.Build
	if b.Context != "" {
		bc.Context = b.Context
	}
	if b.Dockerfile != "" {
		bc.Dockerfile = b.Dockerfile
	}
	if b.DockerfileInline != "" {
		bc.DockerfileInline = b.DockerfileInline
	}
	if len(b.Args) > 0 {
		if bc.Args == nil {
			bc.Args = make(types.MappingWithEquals)
		}
		for k, v := range b.Args {
			val := v
			bc.Args[k] = &val
		}
	}
	if len(b.Labels) > 0 {
		if bc.Labels == nil {
			bc.Labels = make(types.Labels)
		}
		for k, v := range b.Labels {
			bc.Labels[k] = v
		}
	}
	if len(b.Tags) > 0 {
		bc.Tags = append(bc.Tags, b.Tags...)
	}
	bc.Target = firstNonEmpty(bc.Target, b.Target)
	bc.Network = firstNonEmpty(bc.Network, b.Network)
	bc.NoCache = bc.NoCache || b.NoCache
	bc.Pull = bc.Pull || b.Pull
	bc.Privileged = bc.Privileged || b.Privileged
	if len(b.Platforms) > 0 {
		bc.Platforms = append(bc.Platforms, b.Platforms...)
	}
	if len(b.CacheFrom) > 0 {
		bc.CacheFrom = append(bc.CacheFrom, b.CacheFrom...)
	}
	if len(b.CacheTo) > 0 {
		bc.CacheTo = append(bc.CacheTo, b.CacheTo...)
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// Resources are deploy resource limits or reservations.
type Resources struct {
	NanoCPUs    int64
	MemoryBytes uint64
	Pids        int64
}

func (r Resources) apply(dst *types.Resource) {
	if r.NanoCPUs != 0 {
		dst.NanoCPUs = types.NanoCPUs(r.NanoCPUs)
	}
	if r.MemoryBytes != 0 {
		dst.MemoryBytes = types.UnitBytes(r.MemoryBytes)
	}
	if r.Pids != 0 {
		dst.Pids = r.Pids
	}
}

// Deploy is Compose deploy configuration.
type Deploy struct {
	Mode         string
	Replicas     *int
	Labels       map[string]string
	Limits       *Resources
	Reservations *Resources
}

func (d Deploy) apply(cfg *types.ServiceConfig) {
	if cfg.Deploy == nil {
		cfg.Deploy = &types.DeployConfig{}
	}
	if d.Mode != "" {
		cfg.Deploy.Mode = d.Mode
	}
	if d.Replicas != nil {
		cfg.Deploy.Replicas = d.Replicas
	}
	if len(d.Labels) > 0 {
		if cfg.Deploy.Labels == nil {
			cfg.Deploy.Labels = make(map[string]string)
		}
		for k, v := range d.Labels {
			cfg.Deploy.Labels[k] = v
		}
	}
	if d.Limits != nil {
		if cfg.Deploy.Resources.Limits == nil {
			cfg.Deploy.Resources.Limits = &types.Resource{}
		}
		d.Limits.apply(cfg.Deploy.Resources.Limits)
	}
	if d.Reservations != nil {
		if cfg.Deploy.Resources.Reservations == nil {
			cfg.Deploy.Resources.Reservations = &types.Resource{}
		}
		d.Reservations.apply(cfg.Deploy.Resources.Reservations)
	}
}

// ServiceSecret is a service-level secret reference.
type ServiceSecret struct {
	Source string
	Target string
	UID    string
	GID    string
	Mode   *int64
}

func (s ServiceSecret) apply(cfg *types.ServiceConfig) {
	sec := types.ServiceSecretConfig{
		Source: s.Source,
		Target: s.Target,
		UID:    s.UID,
		GID:    s.GID,
	}
	if s.Mode != nil {
		m := types.FileMode(*s.Mode)
		sec.Mode = &m
	}
	cfg.Secrets = append(cfg.Secrets, sec)
}

// ProjectNetwork is extra config for Project.WithNetwork.
type ProjectNetwork struct {
	Driver        string
	DriverOptions map[string]string
	Internal      bool
	Attachable    bool
	EnableIPv6    bool
	Labels        map[string]string
	IPAMDriver    string
	IPAMPools     []IPAMPool
}

// IPAMPool is a project network IPAM pool.
type IPAMPool struct {
	Subnet             string
	Gateway            string
	IPRange            string
	AuxiliaryAddresses map[string]string
}

// ProjectVolume is extra config for Project.WithVolume.
type ProjectVolume struct {
	Driver        string
	DriverOptions map[string]string
	Labels        map[string]string
}

// ProjectSecret is extra config for Project.WithSecret.
type ProjectSecret struct {
	Name           string
	File           string
	Content        string
	Environment    string
	External       bool
	Driver         string
	DriverOptions  map[string]string
	TemplateDriver string
}

// WatchAction is a Compose develop watch action.
type WatchAction string

const (
	WatchActionSync        WatchAction = "sync"
	WatchActionRebuild     WatchAction = "rebuild"
	WatchActionSyncRestart WatchAction = "sync+restart"
)
