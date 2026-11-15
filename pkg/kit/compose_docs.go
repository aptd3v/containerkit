package kit

import (
	"fmt"
	"time"

	"github.com/aptd3v/containerkit/pkg/recursive"
	"github.com/aptd3v/containerkit/pkg/selinux"
)

// StartIntervalOpt is the value returned by [StartInterval].
// It implements [HealthOption].
type StartIntervalOpt struct{ d time.Duration }

// StartInterval sets Compose healthcheck.start_interval inside [Health].
func StartInterval(d time.Duration) StartIntervalOpt { return StartIntervalOpt{d: d} }

func (o StartIntervalOpt) applyHealth(h *health) error {
	if o.d < 0 {
		return cfgErr("", "start_interval", "StartInterval", fmt.Errorf("%w: start interval", errInvalid))
	}
	h.startInterval.assign(o.d)
	return nil
}

// AppProtocolOpt is the value returned by [AppProtocol].
// It implements [PortOption].
type AppProtocolOpt struct{ s string }

// AppProtocol sets Compose ports.app_protocol inside [Publish].
func AppProtocol(s string) AppProtocolOpt { return AppProtocolOpt{s: s} }

func (o AppProtocolOpt) applyPort(p *port) error {
	if o.s == "" {
		return cfgErr("", "app_protocol", "AppProtocol", fmt.Errorf("%w: app protocol", errEmpty))
	}
	p.appProtocol.assign(o.s)
	return nil
}

// PublishedOpt is the value returned by [Published].
// It implements [PortOption].
type PublishedOpt struct{ s string }

// Published sets Compose ports.published, including ranges such as "8083-9000".
func Published(s string) PublishedOpt { return PublishedOpt{s: s} }

func (o PublishedOpt) applyPort(p *port) error {
	if o.s == "" {
		return cfgErr("", "published", "Published", fmt.Errorf("%w: published port", errEmpty))
	}
	p.published.assign(o.s)
	return nil
}

func (o NameOpt) applyPort(p *port) error {
	if err := validateName(o.name); err != nil {
		return wrapPath("publish", err)
	}
	p.portName.assign(o.name)
	return nil
}

func (o ModeOpt) applyPort(p *port) error {
	if o.m == "" {
		return cfgErr("", "mode", "Mode", fmt.Errorf("%w: port mode", errEmpty))
	}
	p.portMode.assign(o.m)
	return nil
}

// LinkLocalOpt is the value returned by [LinkLocal].
// It implements [NetworkOption].
type LinkLocalOpt struct{ ip string }

// LinkLocal adds Compose networks.link_local_ips inside [Network].
func LinkLocal(ip string) LinkLocalOpt { return LinkLocalOpt{ip: ip} }

func (o LinkLocalOpt) applyNetwork(n *netAttach) error {
	if o.ip == "" {
		return cfgErr("", "link_local_ips", "LinkLocal", fmt.Errorf("%w: link local ip", errEmpty))
	}
	n.linkLocal = append(n.linkLocal, o.ip)
	return nil
}

// PriorityOpt is the value returned by [Priority].
// It implements [NetworkOption].
type PriorityOpt struct{ n int }

// Priority sets Compose networks.priority inside [Network].
func Priority(n int) PriorityOpt { return PriorityOpt{n: n} }

func (o PriorityOpt) applyNetwork(n *netAttach) error {
	n.priority.assign(o.n)
	return nil
}

// GatewayPriorityOpt is the value returned by [GatewayPriority].
// It implements [NetworkOption].
type GatewayPriorityOpt struct{ n int }

// GatewayPriority sets Compose networks.gw_priority inside [Network].
func GatewayPriority(n int) GatewayPriorityOpt { return GatewayPriorityOpt{n: n} }

func (o GatewayPriorityOpt) applyNetwork(n *netAttach) error {
	n.gwPriority.assign(o.n)
	return nil
}

// InterfaceNameOpt is the value returned by [InterfaceName].
// It implements [NetworkOption].
type InterfaceNameOpt struct{ s string }

// InterfaceName sets Compose networks.interface_name inside [Network].
func InterfaceName(s string) InterfaceNameOpt { return InterfaceNameOpt{s: s} }

func (o InterfaceNameOpt) applyNetwork(n *netAttach) error {
	if o.s == "" {
		return cfgErr("", "interface_name", "InterfaceName", fmt.Errorf("%w: interface name", errEmpty))
	}
	n.ifName.assign(o.s)
	return nil
}

// UseServiceNameAliasOpt is the value returned by [UseServiceNameAlias].
// It implements [NetworkOption].
type UseServiceNameAliasOpt struct{ v bool }

// UseServiceNameAlias sets Compose networks.use_service_name_alias inside [Network].
func UseServiceNameAlias(v bool) UseServiceNameAliasOpt { return UseServiceNameAliasOpt{v: v} }

func (o UseServiceNameAliasOpt) applyNetwork(n *netAttach) error {
	n.useAlias.assign(o.v)
	return nil
}

// IPRangeOpt is the value returned by [IPRange].
// It implements [ComposeNetOption].
type IPRangeOpt struct{ s string }

// IPRange sets Compose networks.ipam.config.ip_range inside [AddNetwork].
func IPRange(s string) IPRangeOpt { return IPRangeOpt{s: s} }

func (o IPRangeOpt) applyComposeNet(n *projectNet) error {
	if o.s == "" {
		return cfgErr("", "ip_range", "IPRange", fmt.Errorf("%w: ip range", errEmpty))
	}
	n.ipRange.assign(o.s)
	return nil
}

// AuxAddressOpt is the value returned by [AuxAddress].
// It implements [ComposeNetOption].
type AuxAddressOpt struct{ host, addr string }

// AuxAddress adds Compose networks.ipam.config.aux_addresses inside [AddNetwork].
func AuxAddress(host, addr string) AuxAddressOpt { return AuxAddressOpt{host: host, addr: addr} }

func (o AuxAddressOpt) applyComposeNet(n *projectNet) error {
	if o.host == "" {
		return cfgErr("", "aux_addresses", "AuxAddress", fmt.Errorf("%w: aux address name", errEmpty))
	}
	if n.aux == nil {
		n.aux = map[string]string{}
	}
	n.aux[o.host] = o.addr
	return nil
}

// IPAMDriverOpt is the value returned by [IPAMDriver].
// It implements [ComposeNetOption].
type IPAMDriverOpt struct{ d string }

// IPAMDriver sets Compose networks.ipam.driver inside [AddNetwork].
func IPAMDriver(d string) IPAMDriverOpt { return IPAMDriverOpt{d: d} }

func (o IPAMDriverOpt) applyComposeNet(n *projectNet) error {
	if o.d == "" {
		return cfgErr("", "ipam.driver", "IPAMDriver", fmt.Errorf("%w: ipam driver", errEmpty))
	}
	n.ipamDriver.assign(o.d)
	return nil
}

// IPAMOptKV is the value returned by [IPAMOpt].
// It implements [ComposeNetOption].
type IPAMOptKV struct{ key, value string }

// IPAMOpt adds Compose networks.ipam.options inside [AddNetwork].
func IPAMOpt(key, value string) IPAMOptKV { return IPAMOptKV{key: key, value: value} }

func (o IPAMOptKV) applyComposeNet(n *projectNet) error {
	if o.key == "" {
		return cfgErr("", "ipam.options", "IPAMOpt", fmt.Errorf("%w: ipam option key", errEmpty))
	}
	if n.ipamOpts == nil {
		n.ipamOpts = map[string]string{}
	}
	n.ipamOpts[o.key] = o.value
	return nil
}

// InitialSyncOpt is the value returned by [InitialSync].
// It implements [WatchTriggerOption].
type InitialSyncOpt struct{ v bool }

// InitialSync sets Compose develop.watch.initial_sync inside [WatchFiles].
func InitialSync(v bool) InitialSyncOpt { return InitialSyncOpt{v: v} }

func (o InitialSyncOpt) applyWatchTrigger(t *watchTrigger) error {
	t.initialSync.assign(o.v)
	return nil
}

func (o CommandOpt) applyWatchTrigger(t *watchTrigger) error {
	t.exec.cmd = append([]string(nil), o.cmd...)
	return nil
}

func (o CommandOpt) applyHook(h *serviceHook) error {
	if len(o.cmd) == 0 {
		return cfgErr("", "command", "Command", fmt.Errorf("%w: command", errEmpty))
	}
	h.cmd = append([]string(nil), o.cmd...)
	return nil
}

func (o UserOpt) applyWatchTrigger(t *watchTrigger) error {
	t.exec.user.assign(o.user)
	return nil
}

func (o UserOpt) applyHook(h *serviceHook) error {
	h.user.assign(o.user)
	return nil
}

func (o PrivilegedOpt) applyWatchTrigger(t *watchTrigger) error {
	t.exec.privileged.assign(o.v)
	return nil
}

func (o PrivilegedOpt) applyHook(h *serviceHook) error {
	h.privileged.assign(o.v)
	return nil
}

func (o WorkingDirOpt) applyWatchTrigger(t *watchTrigger) error {
	t.exec.workdir.assign(o.dir)
	return nil
}

func (o WorkingDirOpt) applyHook(h *serviceHook) error {
	h.workdir.assign(o.dir)
	return nil
}

func (o EnvOpt) applyWatchTrigger(t *watchTrigger) error {
	t.exec.env = append(t.exec.env, envVar{key: o.key, value: field[string]{set: true, val: o.value}})
	return nil
}

func (o EnvOpt) applyHook(h *serviceHook) error {
	h.env = append(h.env, envVar{key: o.key, value: field[string]{set: true, val: o.value}})
	return nil
}

// TmpfsMountOpt is the value returned by [TmpfsMount].
// It implements [ContainerOption], [JobOption], and [ServiceOption].
type TmpfsMountOpt struct{ path string }

// TmpfsMount adds a Compose service-level tmpfs path (short syntax).
// Long syntax uses [Volume] with Kind(Tmpfs).
func TmpfsMount(path string) TmpfsMountOpt { return TmpfsMountOpt{path: path} }

func (o TmpfsMountOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o TmpfsMountOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o TmpfsMountOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o TmpfsMountOpt) apply(s *spec) error {
	if o.path == "" {
		return cfgErr("", "tmpfs", "TmpfsMount", fmt.Errorf("%w: tmpfs path", errEmpty))
	}
	s.tmpfs = append(s.tmpfs, o.path)
	return nil
}

// RequiredOpt is the value returned by [Required].
// It implements [DependOption] and [EnvFileOption].
type RequiredOpt struct{ v bool }

// Required sets depends_on.required inside [DependsOn], or env_file.required inside [EnvFile].
func Required(v bool) RequiredOpt { return RequiredOpt{v: v} }

func (o RequiredOpt) applyDepend(d *depend) error {
	d.required.assign(o.v)
	return nil
}

func (o RequiredOpt) applyEnvFile(e *envFileRef) error {
	e.required.assign(o.v)
	return nil
}

// FormatOpt is the value returned by [Format].
// It implements [EnvFileOption].
type FormatOpt struct{ s string }

// Format sets Compose env_file.format inside [EnvFile] (raw or other).
func Format(s string) FormatOpt { return FormatOpt{s: s} }

func (o FormatOpt) applyEnvFile(e *envFileRef) error {
	if o.s == "" {
		return cfgErr("", "format", "Format", fmt.Errorf("%w: env file format", errEmpty))
	}
	e.format.assign(o.s)
	return nil
}

// SELinuxOpt is the value returned by [SELinux].
// It implements [VolumeOption].
type SELinuxOpt struct{ s string }

// SELinux sets Compose volumes.bind.selinux inside [Volume] ("z" or "Z").
func SELinux(s selinux.Label) SELinuxOpt { return SELinuxOpt{s: string(s)} }

func (o SELinuxOpt) applyVolume(v *volume) error {
	if o.s == "" {
		return cfgErr("", "selinux", "SELinux", fmt.Errorf("%w: selinux", errEmpty))
	}
	v.selinux.assign(o.s)
	return nil
}

// CreateHostPathOpt is the value returned by [CreateHostPath].
// It implements [VolumeOption].
type CreateHostPathOpt struct{ v bool }

// CreateHostPath sets Compose volumes.bind.create_host_path inside [Volume].
func CreateHostPath(v bool) CreateHostPathOpt { return CreateHostPathOpt{v: v} }

func (o CreateHostPathOpt) applyVolume(v *volume) error {
	v.createHostPath.assign(o.v)
	return nil
}

// RecursiveOpt is the value returned by [Recursive].
// It implements [VolumeOption].
type RecursiveOpt struct{ s string }

// Recursive sets Compose volumes.bind.recursive inside [Volume].
func Recursive(s recursive.Mode) RecursiveOpt { return RecursiveOpt{s: string(s)} }

func (o RecursiveOpt) applyVolume(v *volume) error {
	if o.s == "" {
		return cfgErr("", "recursive", "Recursive", fmt.Errorf("%w: recursive", errEmpty))
	}
	v.recursive.assign(o.s)
	return nil
}

// FileModeOpt is the value returned by [FileMode].
// It implements [FileConfigOption], [SecretOption], and [VolumeOption].
type FileModeOpt struct{ m uint32 }

// FileMode sets tmpfs/secret/config mode (octal file bits).
func FileMode(m uint32) FileModeOpt { return FileModeOpt{m: m} }

func (o FileModeOpt) applyVolume(v *volume) error {
	v.mode.assign(o.m)
	return nil
}

func (o FileModeOpt) applySecret(s *secretRef) error {
	s.mode.assign(o.m)
	return nil
}

func (o FileModeOpt) applyFileConfig(c *fileConfigRef) error {
	c.mode.assign(o.m)
	return nil
}

// UIDOpt is the value returned by [UID].
// It implements [FileConfigOption] and [SecretOption].
type UIDOpt struct{ s string }

// UID sets Compose secret/config uid.
func UID(s string) UIDOpt { return UIDOpt{s: s} }

func (o UIDOpt) applySecret(s *secretRef) error {
	s.uid.assign(o.s)
	return nil
}

func (o UIDOpt) applyFileConfig(c *fileConfigRef) error {
	c.uid.assign(o.s)
	return nil
}

// GIDOpt is the value returned by [GID].
// It implements [FileConfigOption] and [SecretOption].
type GIDOpt struct{ s string }

// GID sets Compose secret/config gid.
func GID(s string) GIDOpt { return GIDOpt{s: s} }

func (o GIDOpt) applySecret(s *secretRef) error {
	s.gid.assign(o.s)
	return nil
}

func (o GIDOpt) applyFileConfig(c *fileConfigRef) error {
	c.gid.assign(o.s)
	return nil
}

// ExtendsFileOpt is the value returned by [ExtendsFile].
// It implements [ServiceOption].
type ExtendsFileOpt struct{ p string }

// ExtendsFile sets Compose extends.file. Service only.
func ExtendsFile(path string) ExtendsFileOpt { return ExtendsFileOpt{p: path} }

func (o ExtendsFileOpt) applyService(s *Service) error {
	if o.p == "" {
		return cfgErr("", "extends.file", "ExtendsFile", fmt.Errorf("%w: extends file", errEmpty))
	}
	s.extendsFile.assign(o.p)
	return nil
}

// ProviderOptionOpt is the value returned by [ProviderOption].
// It implements [ServiceOption].
type ProviderOptionOpt struct{ key, value string }

// ProviderOption adds Compose provider.options. Service only.
func ProviderOption(key, value string) ProviderOptionOpt {
	return ProviderOptionOpt{key: key, value: value}
}

func (o ProviderOptionOpt) applyService(s *Service) error {
	if o.key == "" {
		return cfgErr("", "provider.options", "ProviderOption", fmt.Errorf("%w: option key", errEmpty))
	}
	if s.providerOpts == nil {
		s.providerOpts = map[string][]string{}
	}
	s.providerOpts[o.key] = append(s.providerOpts[o.key], o.value)
	return nil
}

// CredentialConfigOpt is the value returned by [CredentialConfig].
// It implements [ServiceOption].
type CredentialConfigOpt struct{ n string }

// CredentialConfig sets Compose credential_spec.config. Service only.
func CredentialConfig(name string) CredentialConfigOpt { return CredentialConfigOpt{n: name} }

func (o CredentialConfigOpt) applyService(s *Service) error {
	if o.n == "" {
		return cfgErr("", "credential_spec.config", "CredentialConfig", fmt.Errorf("%w: config", errEmpty))
	}
	s.credConfig.assign(o.n)
	return nil
}

// PerReplicaOpt is the value returned by [PerReplica].
// It implements [ContainerOption].
type PerReplicaOpt struct{ v bool }

// PerReplica sets Compose pre_start.per_replica inside [PreStart].
func PerReplica(v bool) PerReplicaOpt { return PerReplicaOpt{v: v} }

func (o PerReplicaOpt) applyContainer(c *Container) error {
	c.perReplica.assign(o.v)
	return nil
}

// TemplateDriverOpt is the value returned by [TemplateDriver].
// It implements [ComposeFileOption] and [ComposeSecretOption].
type TemplateDriverOpt struct{ s string }

// TemplateDriver sets Compose secrets/configs template_driver.
func TemplateDriver(s string) TemplateDriverOpt { return TemplateDriverOpt{s: s} }

func (o TemplateDriverOpt) applyComposeSecret(s *projectSecret) error {
	if o.s == "" {
		return cfgErr("", "template_driver", "TemplateDriver", fmt.Errorf("%w: template driver", errEmpty))
	}
	s.templateDriver.assign(o.s)
	return nil
}

func (o TemplateDriverOpt) applyComposeFile(f *projectFile) error {
	if o.s == "" {
		return cfgErr("", "template_driver", "TemplateDriver", fmt.Errorf("%w: template driver", errEmpty))
	}
	f.templateDriver.assign(o.s)
	return nil
}

// ProjectDirectoryOpt is the value returned by [ProjectDirectory].
// It implements [ImportOption], [IncludeOption], and [LoadProjectOption].
type ProjectDirectoryOpt struct{ s string }

// ProjectDirectory sets Compose include.project_directory inside [AddInclude]
// or the working directory on [LoadProject].
func ProjectDirectory(s string) ProjectDirectoryOpt { return ProjectDirectoryOpt{s: s} }

func (o ProjectDirectoryOpt) applyInclude(inc *includeRef) error {
	if o.s == "" {
		return cfgErr("", "project_directory", "ProjectDirectory", fmt.Errorf("%w: project directory", errEmpty))
	}
	inc.projectDir.assign(o.s)
	return nil
}

func (o PathOpt) applyInclude(inc *includeRef) error {
	if o.p == "" {
		return cfgErr("", "path", "Path", fmt.Errorf("%w: include path", errEmpty))
	}
	inc.paths = append(inc.paths, o.p)
	return nil
}

// DiscreteOpt is the value returned by [Discrete].
// It implements [DeployOption] and [ReserveOption].
type DiscreteOpt struct {
	kind  string
	value int64
}

// Discrete adds Compose deploy.resources.generic_resources.discrete_resource_spec.
func Discrete(kind string, value int64) DiscreteOpt { return DiscreteOpt{kind: kind, value: value} }

func (o DiscreteOpt) applyDeploy(d *deploy) error {
	if o.kind == "" {
		return cfgErr("", "generic_resources", "Discrete", fmt.Errorf("%w: kind", errEmpty))
	}
	d.generic = append(d.generic, discreteRes{kind: o.kind, value: o.value})
	return nil
}

func (o DiscreteOpt) applyReserve(r *resource) error {
	if o.kind == "" {
		return cfgErr("", "generic_resources", "Discrete", fmt.Errorf("%w: kind", errEmpty))
	}
	r.generic = append(r.generic, discreteRes{kind: o.kind, value: o.value})
	return nil
}

// AddIncludeOpt is the value returned by [AddInclude].
// It implements [ComposeOption].
type AddIncludeOpt struct{ opts []IncludeOption }

// AddInclude adds a top-level Compose include entry.
func AddInclude(opts ...IncludeOption) AddIncludeOpt { return AddIncludeOpt{opts: opts} }

func (o AddIncludeOpt) applyCompose(p *Compose) error {
	var inc includeRef
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("include.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyInclude(&inc); err != nil {
			errs = append(errs, wrapPath("include", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if len(inc.paths) == 0 {
		return cfgErr("", "include.path", "Path", fmt.Errorf("%w: include path", errRequired))
	}
	p.includes = append(p.includes, inc)
	return nil
}
