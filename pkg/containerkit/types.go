package containerkit

import (
	"maps"
	"os"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/mount"
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
type MountType = mount.Type

const (
	MountBind      = mount.TypeBind
	MountVolume    = mount.TypeVolume
	MountTmpfs     = mount.TypeTmpfs
	MountNamedPipe = mount.TypeNamedPipe
	MountImage     = mount.TypeImage
	MountCluster   = mount.TypeCluster
)

// MountConsistency is a bind-mount consistency mode.
type MountConsistency = mount.Consistency

const (
	ConsistencyFull      = mount.ConsistencyFull
	ConsistencyCached    = mount.ConsistencyCached
	ConsistencyDelegated = mount.ConsistencyDelegated
	ConsistencyDefault   = mount.ConsistencyDefault
)

// SELinux relabel modes for a bind mount.
const (
	SELinuxShared  = "z"
	SELinuxPrivate = "Z"
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

// Mount describes one mount. Empty Type is bind, unless Volume, Tmpfs, or Image is set.
type Mount struct {
	Source      string
	Target      string
	Type        MountType
	ReadOnly    bool
	Consistency MountConsistency

	Bind   *BindOptions
	Volume *VolumeOptions
	Tmpfs  *TmpfsOptions
	Image  *ImageOptions
}

// BindOptions are options for a bind mount.
type BindOptions struct {
	Propagation            MountPropagation
	SELinux                string // z or Z
	NonRecursive           bool
	CreateHostPath         *bool
	ReadOnlyNonRecursive   bool
	ReadOnlyForceRecursive bool
}

// VolumeOptions are options for a named volume mount.
type VolumeOptions struct {
	NoCopy        bool
	Labels        map[string]string
	Subpath       string
	Driver        string
	DriverOptions map[string]string
}

// TmpfsOptions are options for a tmpfs mount.
type TmpfsOptions struct {
	SizeBytes int64
	Mode      os.FileMode
	Options   [][]string
}

// ImageOptions are options for an image mount.
type ImageOptions struct {
	Subpath string
}

// Protocol is a published port protocol.
type Protocol string

const (
	TCP  Protocol = "tcp"
	UDP  Protocol = "udp"
	SCTP Protocol = "sctp"
)

// PortBinding publishes one container port on the host.
type PortBinding struct {
	Protocol      Protocol
	HostIP        string
	HostPort      string
	ContainerPort string
}

// Device maps a host device into the container.
// Read, Write, and Mknod are the cgroup permissions.
type Device struct {
	HostPath      string
	ContainerPath string
	Read          bool
	Write         bool
	Mknod         bool
}

// Ulimit is one soft and hard resource limit.
type Ulimit struct {
	Name string
	Soft int64
	Hard int64
}

// ConsoleSize is the container console height and width in characters.
type ConsoleSize struct {
	Height uint
	Width  uint
}

// DeviceRequest asks a device driver for devices such as GPUs.
// Count -1 requests every device the driver offers.
// Capabilities is one required set, for example "gpu" or "gpu" and "compute".
// Leave DeviceIDs empty to let Count choose the devices.
type DeviceRequest struct {
	Driver       string
	Count        int
	DeviceIDs    []string
	Capabilities []string
	Options      map[string]string
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
	Context            string
	Dockerfile         string
	DockerfileInline   string
	Args               map[string]string
	Labels             map[string]string
	Tags               []string
	Target             string
	Network            string
	NoCache            bool
	Pull               bool
	Privileged         bool
	Platforms          []string
	CacheFrom          []string
	CacheTo            []string
	SSH                []BuildSSH
	AdditionalContexts map[string]string
	ExtraHosts         []string
	Isolation          string
	ShmSize            int64
	Entitlements       []string
	Ulimits            map[string]BuildUlimit
	Secrets            []BuildSecret
	Provenance         string
	SBOM               string
	NoCacheFilter      []string

	// err is set by dockerFile.WithInline when the dockerfile failed validation.
	// WithService reports it and does not apply the spec.
	err error
}

// BuildSSH is one SSH agent socket or key for a service build.
type BuildSSH struct {
	ID   string
	Path string
}

// BuildUlimit is a soft/hard ulimit for a service build.
type BuildUlimit struct {
	Soft int
	Hard int
}

// BuildSecret is a build-time secret reference.
type BuildSecret struct {
	Source string
	Target string
	UID    string
	GID    string
	Mode   *int64
}

func (b BuildSpec) apply(cfg *types.ServiceConfig) error {
	if b.err != nil {
		return b.err
	}
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
	bc.Isolation = firstNonEmpty(bc.Isolation, b.Isolation)
	bc.Provenance = firstNonEmpty(bc.Provenance, b.Provenance)
	bc.SBOM = firstNonEmpty(bc.SBOM, b.SBOM)
	bc.NoCache = bc.NoCache || b.NoCache
	bc.Pull = bc.Pull || b.Pull
	bc.Privileged = bc.Privileged || b.Privileged
	if b.ShmSize > 0 {
		bc.ShmSize = types.UnitBytes(b.ShmSize)
	}
	if len(b.Platforms) > 0 {
		bc.Platforms = append(bc.Platforms, b.Platforms...)
	}
	if len(b.CacheFrom) > 0 {
		bc.CacheFrom = append(bc.CacheFrom, b.CacheFrom...)
	}
	if len(b.CacheTo) > 0 {
		bc.CacheTo = append(bc.CacheTo, b.CacheTo...)
	}
	if len(b.NoCacheFilter) > 0 {
		bc.NoCacheFilter = append(bc.NoCacheFilter, b.NoCacheFilter...)
	}
	if len(b.Entitlements) > 0 {
		bc.Entitlements = append(bc.Entitlements, b.Entitlements...)
	}
	for _, key := range b.SSH {
		bc.SSH = append(bc.SSH, types.SSHKey{ID: key.ID, Path: key.Path})
	}
	if len(b.AdditionalContexts) > 0 {
		if bc.AdditionalContexts == nil {
			bc.AdditionalContexts = make(types.Mapping)
		}
		for k, v := range b.AdditionalContexts {
			bc.AdditionalContexts[k] = v
		}
	}
	if len(b.ExtraHosts) > 0 {
		hosts, err := convertExtraHosts(b.ExtraHosts)
		if err != nil {
			return err
		}
		if bc.ExtraHosts == nil {
			bc.ExtraHosts = hosts
		} else {
			for k, v := range hosts {
				bc.ExtraHosts[k] = append(bc.ExtraHosts[k], v...)
			}
		}
	}
	if len(b.Ulimits) > 0 {
		if bc.Ulimits == nil {
			bc.Ulimits = make(map[string]*types.UlimitsConfig)
		}
		for name, limit := range b.Ulimits {
			bc.Ulimits[name] = &types.UlimitsConfig{Soft: limit.Soft, Hard: limit.Hard}
		}
	}
	for _, secret := range b.Secrets {
		ref := types.ServiceSecretConfig{
			Source: secret.Source,
			Target: secret.Target,
			UID:    secret.UID,
			GID:    secret.GID,
		}
		if secret.Mode != nil {
			mode := types.FileMode(*secret.Mode)
			ref.Mode = &mode
		}
		bc.Secrets = append(bc.Secrets, ref)
	}
	return nil
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
			cfg.Deploy.Labels = make(map[string]string, len(d.Labels))
		}
		maps.Copy(cfg.Deploy.Labels, d.Labels)
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

func (n ProjectNetwork) apply(cfg *types.NetworkConfig) {
	cfg.Driver = n.Driver
	if len(n.DriverOptions) > 0 {
		cfg.DriverOpts = types.Options(maps.Clone(n.DriverOptions))
	}
	cfg.Internal = n.Internal
	cfg.Attachable = n.Attachable
	if n.EnableIPv6 {
		enable := true
		cfg.EnableIPv6 = &enable
	}
	if len(n.Labels) > 0 {
		cfg.Labels = types.Labels(maps.Clone(n.Labels))
	}
	if n.IPAMDriver != "" || len(n.IPAMPools) > 0 {
		cfg.Ipam.Driver = n.IPAMDriver
		for _, pool := range n.IPAMPools {
			if pool.Subnet == "" && pool.Gateway == "" && pool.IPRange == "" && len(pool.AuxiliaryAddresses) == 0 {
				continue
			}
			item := &types.IPAMPool{
				Subnet:  pool.Subnet,
				Gateway: pool.Gateway,
				IPRange: pool.IPRange,
			}
			if len(pool.AuxiliaryAddresses) > 0 {
				item.AuxiliaryAddresses = types.Mapping(maps.Clone(pool.AuxiliaryAddresses))
			}
			cfg.Ipam.Config = append(cfg.Ipam.Config, item)
		}
	}
}

func (v ProjectVolume) apply(cfg *types.VolumeConfig) {
	cfg.Driver = v.Driver
	if len(v.DriverOptions) > 0 {
		cfg.DriverOpts = types.Options(maps.Clone(v.DriverOptions))
	}
	if len(v.Labels) > 0 {
		cfg.Labels = types.Labels(maps.Clone(v.Labels))
	}
}

func (s ProjectSecret) apply(cfg *types.SecretConfig) {
	cfg.Name = s.Name
	cfg.File = s.File
	cfg.Content = s.Content
	cfg.Environment = s.Environment
	if s.External {
		cfg.External = true
	}
	cfg.Driver = s.Driver
	if len(s.DriverOptions) > 0 {
		cfg.DriverOpts = maps.Clone(s.DriverOptions)
	}
	cfg.TemplateDriver = s.TemplateDriver
}

// WatchAction is a Compose develop watch action.
type WatchAction string

const (
	WatchActionSync        WatchAction = "sync"
	WatchActionRebuild     WatchAction = "rebuild"
	WatchActionSyncRestart WatchAction = "sync+restart"
)
