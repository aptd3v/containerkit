package kit

import (
	"fmt"
	"time"

	"github.com/aptd3v/containerkit/pkg/signal"
)

// NameOpt is the value returned by [Name].
// It implements [ComposeFileOption], [ComposeModelOption], [ComposeNetOption], [ComposeOption],
// [ComposeSecretOption], [ComposeVolOption], [ContainerOption], [DependOption],
// [EngineConfigCreateOption], [EngineConfigUpdateOption], [EngineCreateOption],
// [EngineNodeUpdateOption], [EnginePluginCreateOption], [EnginePluginInstallOption],
// [EnginePluginUpgradeOption], [EngineRenameOption], [EngineSecretCreateOption],
// [EngineSecretUpdateOption], [EngineServiceCreateOption], [EngineServiceUpdateOption],
// [EngineVolumeOption], [FileConfigOption], [GenerateOption], [ImportOption], [JobOption],
// [LoadProjectOption], [ModelOption], [NetworkOption], [PortOption], [RunOption],
// [SecretOption], [ServiceOption], [UlimitOption], and [VolumeOption].
type NameOpt struct{ name string }

// Name sets the receiver's name: container, service, job, project, depends_on,
// volume, network, secret, or config-file.
func Name(name string) NameOpt { return NameOpt{name: name} }

func (o NameOpt) applyContainer(c *Container) error { return c.setName(o.name) }
func (o NameOpt) applyService(s *Service) error     { return s.setName(o.name) }
func (o NameOpt) applyJob(j *Job) error             { return j.setName(o.name) }
func (o NameOpt) applyCompose(p *Compose) error     { return p.setName(o.name) }
func (o NameOpt) applyDepend(d *depend) error       { return d.setName(o.name) }
func (o NameOpt) applyVolume(v *volume) error       { return v.setName(o.name) }
func (o NameOpt) applyNetwork(n *netAttach) error   { return n.setName(o.name) }
func (o NameOpt) applySecret(s *secretRef) error    { return s.setName(o.name) }
func (o NameOpt) applyFileConfig(c *fileConfigRef) error {
	return c.setName(o.name)
}
func (o NameOpt) applyComposeNet(n *projectNet) error       { return n.setName(o.name) }
func (o NameOpt) applyComposeVol(v *projectVol) error       { return v.setName(o.name) }
func (o NameOpt) applyComposeSecret(s *projectSecret) error { return s.setName(o.name) }
func (o NameOpt) applyComposeFile(f *projectFile) error     { return f.setName(o.name) }
func (o NameOpt) applyComposeModel(m *projectModel) error {
	if err := validateName(o.name); err != nil {
		return wrapPath("model", err)
	}
	m.name.assign(o.name)
	return nil
}
func (o NameOpt) applyUlimit(u *ulimit) error { return u.setName(o.name) }

// ImageOpt is the value returned by [Image].
// It implements [ComposeModelOption], [ContainerOption], [EngineBuildOption],
// [EngineServiceCreateOption], [EngineServiceUpdateOption], [JobOption], and [ServiceOption].
type ImageOpt struct{ image string }

// Image sets the image reference on a container, service, job, or Engine build tag.
func Image(image string) ImageOpt { return ImageOpt{image: image} }

func (o ImageOpt) applyContainer(c *Container) error { return c.setImage(o.image) }
func (o ImageOpt) applyService(s *Service) error     { return s.setImage(o.image) }
func (o ImageOpt) applyJob(j *Job) error             { return j.setImage(o.image) }
func (o ImageOpt) applyComposeModel(m *projectModel) error {
	if o.image == "" {
		return cfgErr("", "model", "Image", fmt.Errorf("%w: model image", errEmpty))
	}
	m.model.assign(o.image)
	return nil
}

// CommandOpt is the value returned by [Command].
// It implements [ContainerOption], [EngineExecCreateOption], [EngineServiceCreateOption],
// [EngineServiceUpdateOption], [HookOption], [JobOption], [RunOption], [ServiceOption], and
// [WatchTriggerOption].
type CommandOpt struct{ cmd []string }

// Command sets the command (argv) for a container, service, job, [Compose.Run],
// or [CreateExec] / [ExecTerminal].
func Command(cmd ...string) CommandOpt { return CommandOpt{cmd: cmd} }

func (o CommandOpt) applyContainer(c *Container) error { return c.setCommand(o.cmd) }
func (o CommandOpt) applyService(s *Service) error     { return s.setCommand(o.cmd) }
func (o CommandOpt) applyJob(j *Job) error             { return j.setCommand(o.cmd) }
func (o CommandOpt) applyRun(a *runAct) error {
	a.command = append([]string(nil), o.cmd...)
	return nil
}
func (o CommandOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.Cmd = append([]string(nil), o.cmd...)
	return nil
}

// EntrypointOpt is the value returned by [Entrypoint].
// It implements [ContainerOption], [JobOption], [RunOption], and [ServiceOption].
type EntrypointOpt struct{ cmd []string }

// Entrypoint sets the entrypoint (argv).
func Entrypoint(cmd ...string) EntrypointOpt { return EntrypointOpt{cmd: cmd} }

func (o EntrypointOpt) applyContainer(c *Container) error { return c.setEntrypoint(o.cmd) }
func (o EntrypointOpt) applyService(s *Service) error     { return s.setEntrypoint(o.cmd) }
func (o EntrypointOpt) applyJob(j *Job) error             { return j.setEntrypoint(o.cmd) }
func (o EntrypointOpt) applyRun(a *runAct) error {
	a.entry = append([]string(nil), o.cmd...)
	return nil
}

// EnvOpt is the value returned by [Env].
// It implements [BuildOption], [ComposeBuildOption], [ComposeCreateOption], [ContainerOption],
// [EngineBuildOption], [EngineServiceCreateOption], [EngineServiceUpdateOption], [HookOption],
// [ImportOption], [JobOption], [RunOption], [ServiceOption], [UpOption], [WatchOption], and
// [WatchTriggerOption].
type EnvOpt struct{ key, value string }

// Env adds an environment variable at the outer scope, a build-arg inside
// [Build] or [Compose.Build], or a run-time variable on [Compose.Run].
// The value is treated as secret in errors.
func Env(key, value string) EnvOpt { return EnvOpt{key: key, value: value} }

func (o EnvOpt) applyContainer(c *Container) error { return c.addEnv(o.key, o.value) }
func (o EnvOpt) applyService(s *Service) error     { return s.addEnv(o.key, o.value) }
func (o EnvOpt) applyJob(j *Job) error             { return j.addEnv(o.key, o.value) }
func (o EnvOpt) applyBuild(b *build) error         { return b.addArg(o.key, o.value) }

// LabelOpt is the value returned by [Label].
// It implements [BuildOption], [ComposeFileOption], [ComposeNetOption], [ComposeSecretOption],
// [ComposeVolOption], [ContainerOption], [DeployOption], [EngineBuildOption],
// [EngineConfigCreateOption], [EngineConfigUpdateOption], [EngineNetworkOption],
// [EngineNodeUpdateOption], [EngineSecretCreateOption], [EngineSecretUpdateOption],
// [EngineServiceCreateOption], [EngineServiceUpdateOption], [EngineVolumeOption],
// [FileConfigOption], [JobOption], [RunOption], [SecretOption], [ServiceOption], and
// [VolumeOption].
type LabelOpt struct{ key, value string }

// Label adds a label on the receiver: container/service/job, or nested
// volume, network, build, deploy, or project resource.
func Label(key, value string) LabelOpt { return LabelOpt{key: key, value: value} }

func (o LabelOpt) applyContainer(c *Container) error { return c.addLabel(o.key, o.value) }
func (o LabelOpt) applyService(s *Service) error     { return s.addLabel(o.key, o.value) }
func (o LabelOpt) applyJob(j *Job) error             { return j.addLabel(o.key, o.value) }
func (o LabelOpt) applyVolume(v *volume) error       { return v.addLabel(o.key, o.value) }
func (o LabelOpt) applyBuild(b *build) error         { return b.addLabel(o.key, o.value) }
func (o LabelOpt) applyDeploy(d *deploy) error       { return d.addLabel(o.key, o.value) }
func (o LabelOpt) applySecret(s *secretRef) error    { return s.addLabel(o.key, o.value) }
func (o LabelOpt) applyFileConfig(c *fileConfigRef) error {
	return c.addLabel(o.key, o.value)
}
func (o LabelOpt) applyComposeNet(n *projectNet) error { return n.addLabel(o.key, o.value) }
func (o LabelOpt) applyComposeVol(v *projectVol) error { return v.addLabel(o.key, o.value) }
func (o LabelOpt) applyComposeSecret(s *projectSecret) error {
	return s.addLabel(o.key, o.value)
}
func (o LabelOpt) applyComposeFile(f *projectFile) error { return f.addLabel(o.key, o.value) }

// UserOpt is the value returned by [User].
// It implements [ContainerOption], [EngineExecCreateOption], [HookOption], [JobOption],
// [RunOption], [ServiceOption], and [WatchTriggerOption].
type UserOpt struct{ user string }

// User sets the user the process runs as.
func User(user string) UserOpt { return UserOpt{user: user} }

func (o UserOpt) applyContainer(c *Container) error { return c.setUser(o.user) }
func (o UserOpt) applyService(s *Service) error     { return s.setUser(o.user) }
func (o UserOpt) applyJob(j *Job) error             { return j.setUser(o.user) }
func (o UserOpt) applyRun(a *runAct) error          { a.user = o.user; return nil }

// WorkingDirOpt is the value returned by [WorkingDir].
// It implements [ContainerOption], [HookOption], [JobOption], [RunOption],
// [ServiceOption], and [WatchTriggerOption].
type WorkingDirOpt struct{ dir string }

// WorkingDir sets a container, service, job, or hook working directory.
// Compose project directory on [LoadProject] is [ProjectDirectory].
func WorkingDir(dir string) WorkingDirOpt { return WorkingDirOpt{dir: dir} }

func (o WorkingDirOpt) applyContainer(c *Container) error { return c.setWorkingDir(o.dir) }
func (o WorkingDirOpt) applyService(s *Service) error     { return s.setWorkingDir(o.dir) }
func (o WorkingDirOpt) applyJob(j *Job) error             { return j.setWorkingDir(o.dir) }
func (o WorkingDirOpt) applyRun(a *runAct) error          { a.workdir = o.dir; return nil }

// Hostname sets the container hostname.
func Hostname(host string) SpecOpt {
	return specFn(func(s *spec) error { return s.setHostname(host) })
}

// DomainName sets the NIS domain name (Engine Domainname / Compose domainname).
func DomainName(d string) SpecOpt {
	return specFn(func(s *spec) error { return s.setDomainName(d) })
}

// MacAddressOpt is the value returned by [MacAddress].
// It implements [ContainerOption], [JobOption], [NetworkOption], and [ServiceOption].
type MacAddressOpt struct{ m string }

// MacAddress sets the container MAC address, or the endpoint MAC inside [Network].
func MacAddress(m string) MacAddressOpt { return MacAddressOpt{m: m} }

func (o MacAddressOpt) applyContainer(c *Container) error { return c.setMacAddress(o.m) }
func (o MacAddressOpt) applyService(s *Service) error     { return s.setMacAddress(o.m) }
func (o MacAddressOpt) applyJob(j *Job) error             { return j.setMacAddress(o.m) }
func (o MacAddressOpt) applyNetwork(n *netAttach) error {
	n.mac.assign(o.m)
	return nil
}

// StopSignalOpt is the value returned by [StopSignal].
// It implements [ContainerOption], [EngineStopOption], [JobOption], [KillOption], and
// [ServiceOption].
type StopSignalOpt struct{ sig string }

// StopSignal sets the signal used to stop or kill the container.
func StopSignal(sig signal.Signal) StopSignalOpt { return StopSignalOpt{sig: string(sig)} }

func (o StopSignalOpt) applyContainer(c *Container) error { return c.setStopSignal(o.sig) }
func (o StopSignalOpt) applyService(s *Service) error     { return s.setStopSignal(o.sig) }
func (o StopSignalOpt) applyJob(j *Job) error             { return j.setStopSignal(o.sig) }
func (o StopSignalOpt) applyKill(a *killAct) error        { a.signal = o.sig; return nil }
func (o StopSignalOpt) applyEngineStop(a *engineStopAct) error {
	a.signal = o.sig
	return nil
}

// PlatformOpt is the value returned by [Platform].
// It implements [BuildOption], [ContainerOption], [EngineBuildOption], [EnginePullOption],
// [JobOption], and [ServiceOption].
type PlatformOpt struct{ p string }

// Platform sets the target platform (Compose platform / Engine create, pull, or build).
func Platform(p string) PlatformOpt { return PlatformOpt{p: p} }

func (o PlatformOpt) applyContainer(c *Container) error { return c.setPlatform(o.p) }
func (o PlatformOpt) applyService(s *Service) error     { return s.setPlatform(o.p) }
func (o PlatformOpt) applyJob(j *Job) error             { return j.setPlatform(o.p) }

// PrivilegedOpt is the value returned by [Privileged].
// It implements [BuildOption], [ContainerOption], [EngineExecCreateOption], [HookOption],
// [JobOption], [RunOption], [ServiceOption], and [WatchTriggerOption].
type PrivilegedOpt struct{ v bool }

// Privileged sets privileged mode, including explicit false. Inside [Build]
// it maps to compose build.privileged.
func Privileged(v bool) PrivilegedOpt { return PrivilegedOpt{v: v} }

func (o PrivilegedOpt) applyContainer(c *Container) error { return c.setPrivileged(o.v) }
func (o PrivilegedOpt) applyService(s *Service) error     { return s.setPrivileged(o.v) }
func (o PrivilegedOpt) applyJob(j *Job) error             { return j.setPrivileged(o.v) }
func (o PrivilegedOpt) applyBuild(b *build) error         { return b.setPrivileged(o.v) }

// MemoryOpt is the value returned by [Memory].
// It implements [BuildOption], [ComposeBuildOption], [ComposeCreateOption], [ContainerOption],
// [DeployOption], [EngineBuildOption], [JobOption], [ReserveOption], [ServiceOption],
// [UpOption], and [WatchOption].
type MemoryOpt struct{ n int64 }

// Memory sets the memory limit in bytes, including explicit 0.
// Inside [Deploy] it maps to deploy.resources.limits.memory.
func Memory(bytes int64) MemoryOpt { return MemoryOpt{n: bytes} }

func (o MemoryOpt) applyContainer(c *Container) error { return c.setMemory(o.n) }
func (o MemoryOpt) applyService(s *Service) error     { return s.setMemory(o.n) }
func (o MemoryOpt) applyJob(j *Job) error             { return j.setMemory(o.n) }
func (o MemoryOpt) applyDeploy(d *deploy) error       { return d.setMemory(o.n) }
func (o MemoryOpt) applyBuild(b *build) error         { return b.setShmSize(o.n) }

// TtyOpt is the value returned by [Tty].
// It implements [ContainerOption], [EngineExecAttachOption], [EngineExecCreateOption],
// [EngineExecStartOption], [JobOption], [RunOption], and [ServiceOption].
type TtyOpt struct{ v bool }

// Tty attaches a TTY, including explicit false.
func Tty(v bool) TtyOpt { return TtyOpt{v: v} }

func (o TtyOpt) applyContainer(c *Container) error { return c.setTty(o.v) }
func (o TtyOpt) applyService(s *Service) error     { return s.setTty(o.v) }
func (o TtyOpt) applyJob(j *Job) error             { return j.setTty(o.v) }
func (o TtyOpt) applyRun(a *runAct) error          { a.tty.assign(o.v); return nil }

// InteractiveOpt is the value returned by [Interactive].
// It implements [ContainerOption], [JobOption], [RunOption], and [ServiceOption].
type InteractiveOpt struct{ v bool }

// Interactive opens stdin, including explicit false.
func Interactive(v bool) InteractiveOpt { return InteractiveOpt{v: v} }

func (o InteractiveOpt) applyContainer(c *Container) error { return c.setInteractive(o.v) }
func (o InteractiveOpt) applyService(s *Service) error     { return s.setInteractive(o.v) }
func (o InteractiveOpt) applyJob(j *Job) error             { return j.setInteractive(o.v) }
func (o InteractiveOpt) applyRun(a *runAct) error          { a.stdin.assign(o.v); return nil }

// ReadOnlyOpt is the value returned by [ReadOnly].
// It implements [ContainerOption], [JobOption], [ServiceOption], and [VolumeOption].
type ReadOnlyOpt struct{ v bool }

// ReadOnly mounts the root filesystem read-only at the outer scope, or the
// nested [Volume] read-only inside Volume(...).
func ReadOnly(v bool) ReadOnlyOpt { return ReadOnlyOpt{v: v} }

func (o ReadOnlyOpt) applyContainer(c *Container) error { return c.setReadOnly(o.v) }
func (o ReadOnlyOpt) applyService(s *Service) error     { return s.setReadOnly(o.v) }
func (o ReadOnlyOpt) applyJob(j *Job) error             { return j.setReadOnly(o.v) }
func (o ReadOnlyOpt) applyVolume(v *volume) error       { return v.setReadOnly(o.v) }

// Init runs a signal-forwarding init, including explicit false.
func Init(v bool) SpecOpt {
	return specFn(func(s *spec) error { return s.setInit(v) })
}

// StopTimeout sets the stop timeout. Zero is explicit.
func StopTimeout(d time.Duration) SpecOpt {
	return specFn(func(s *spec) error {
		sec := int(d / time.Second)
		if d > 0 && sec == 0 {
			return cfgErr("", "stop_timeout", "StopTimeout", errInvalid)
		}
		return s.setStopTimeout(sec)
	})
}

// CapAddOpt is the value returned by [CapAdd].
// It implements [ContainerOption], [JobOption], [RunOption], and [ServiceOption].
type CapAddOpt struct{ c Capability }

// CapAdd adds a Linux capability. Pass cap.ALL or a kernel name such as cap.NET_ADMIN.
func CapAdd(c Capability) CapAddOpt { return CapAddOpt{c: c} }

func (o CapAddOpt) applyContainer(c *Container) error { return c.addCap(o.c) }
func (o CapAddOpt) applyService(s *Service) error     { return s.addCap(o.c) }
func (o CapAddOpt) applyJob(j *Job) error             { return j.addCap(o.c) }
func (o CapAddOpt) applyRun(a *runAct) error {
	a.capAdd = append(a.capAdd, string(o.c))
	return nil
}

// CapDropOpt is the value returned by [CapDrop].
// It implements [ContainerOption], [JobOption], [RunOption], and [ServiceOption].
type CapDropOpt struct{ c Capability }

// CapDrop drops a Linux capability. Pass cap.ALL or a kernel name such as cap.MKNOD.
func CapDrop(c Capability) CapDropOpt { return CapDropOpt{c: c} }

func (o CapDropOpt) applyContainer(c *Container) error { return c.dropCap(o.c) }
func (o CapDropOpt) applyService(s *Service) error     { return s.dropCap(o.c) }
func (o CapDropOpt) applyJob(j *Job) error             { return j.dropCap(o.c) }
func (o CapDropOpt) applyRun(a *runAct) error {
	a.capDrop = append(a.capDrop, string(o.c))
	return nil
}

// AutoRemoveOpt is the value returned by [AutoRemove].
// It implements [ContainerOption] and [RunOption].
type AutoRemoveOpt struct{ v bool }

// AutoRemove removes the container on exit. Engine-only; converting to a
// Compose service reports an unsupported mapping.
func AutoRemove(v bool) AutoRemoveOpt { return AutoRemoveOpt{v: v} }

func (o AutoRemoveOpt) applyContainer(c *Container) error { return c.setAutoRemove(o.v) }

// SourceOpt is the value returned by [Source].
// It implements [BuildOption], [DeviceOption], [EngineBuildOption], [EngineCopyFromOption],
// [EngineImageTagOption], [FileConfigOption], [SecretOption], and [VolumeOption].
type SourceOpt struct{ s string }

// Source is the origin: bind path, named-volume name, secret/config file
// source, or build context inside [Build].
func Source(s string) SourceOpt { return SourceOpt{s: s} }

func (o SourceOpt) applyVolume(v *volume) error { return v.setSource(o.s) }
func (o SourceOpt) applySecret(s *secretRef) error {
	return s.setSource(o.s)
}
func (o SourceOpt) applyFileConfig(c *fileConfigRef) error { return c.setSource(o.s) }
func (o SourceOpt) applyBuild(b *build) error              { return b.setSource(o.s) }
func (o SourceOpt) applyEngineBuild(a *engineBuildAct) error {
	a.contextDir = o.s
	return nil
}
func (o SourceOpt) applyDevice(d *device) error { return d.setHost(o.s) }

// TargetOpt is the value returned by [Target].
// It implements [BuildOption], [DeviceOption], [EngineBuildOption], [EngineCopyToOption],
// [EngineImageTagOption], [FileConfigOption], [SecretOption], [VolumeOption], and
// [WatchTriggerOption].
type TargetOpt struct{ s string }

// Target is the destination: mount path, secret/config target, or build stage
// inside [Build].
func Target(s string) TargetOpt { return TargetOpt{s: s} }

func (o TargetOpt) applyVolume(v *volume) error            { return v.setTarget(o.s) }
func (o TargetOpt) applySecret(s *secretRef) error         { return s.setTarget(o.s) }
func (o TargetOpt) applyFileConfig(c *fileConfigRef) error { return c.setTarget(o.s) }
func (o TargetOpt) applyBuild(b *build) error              { return b.setTarget(o.s) }
func (o TargetOpt) applyDevice(d *device) error            { return d.setContainer(o.s) }

// DriverValue is the value returned by [Driver].
// The type is not named DriverOpt because that name is already the [DriverOpt] setter.
// It implements [ComposeNetOption], [ComposeSecretOption], [ComposeVolOption],
// [EngineNetworkOption], [EngineVolumeOption], [GPUOption], [LogOption], [SecretOption], and
// [VolumeOption].
type DriverValue struct{ s string }

// Driver sets the driver name for logging, devices, GPUs, volumes, networks,
// or secrets. Known log drivers live in package logdrv; network drivers in netdrv.
func Driver[T ~string](s T) DriverValue { return DriverValue{s: string(s)} }

func (o DriverValue) applyLog(l *logging) error           { return l.setDriver(o.s) }
func (o DriverValue) applyVolume(v *volume) error         { return v.setDriver(o.s) }
func (o DriverValue) applyGPU(g *gpu) error               { return g.setDriver(o.s) }
func (o DriverValue) applyComposeNet(n *projectNet) error { return n.setDriver(o.s) }
func (o DriverValue) applyComposeVol(v *projectVol) error { return v.setDriver(o.s) }
func (o DriverValue) applyComposeSecret(s *projectSecret) error {
	return s.setDriver(o.s)
}
func (o DriverValue) applySecret(s *secretRef) error { return s.setDriver(o.s) }

// FileOpt is the value returned by [File].
// It implements [ImportOption] and [LoadProjectOption].
type FileOpt struct{ path string }

// File adds a compose YAML path on [ImportFile] / [LoadProject]. Secret and
// config content use [FromFile]. A Dockerfile path uses [DockerfilePath].
// An export archive uses [OutputFile].
func File(path string) FileOpt { return FileOpt{path: path} }

// FromFileOpt is the value returned by [FromFile].
// It implements [ComposeFileOption], [ComposeSecretOption], [FileConfigOption], and [SecretOption].
type FromFileOpt struct{ path string }

// FromFile sets the filesystem source of a secret or config object.
func FromFile(path string) FromFileOpt { return FromFileOpt{path: path} }

func (o FromFileOpt) filePath() (string, error) {
	if o.path == "" {
		return "", cfgErr("", "file", "FromFile", fmt.Errorf("%w: file", errEmpty))
	}
	return o.path, nil
}

func (o FromFileOpt) applySecret(s *secretRef) error {
	p, err := o.filePath()
	if err != nil {
		return err
	}
	return s.setFile(p)
}
func (o FromFileOpt) applyFileConfig(c *fileConfigRef) error {
	p, err := o.filePath()
	if err != nil {
		return err
	}
	return c.setFile(p)
}
func (o FromFileOpt) applyComposeSecret(s *projectSecret) error {
	p, err := o.filePath()
	if err != nil {
		return err
	}
	return s.setFile(p)
}
func (o FromFileOpt) applyComposeFile(f *projectFile) error {
	p, err := o.filePath()
	if err != nil {
		return err
	}
	return f.setFile(p)
}

// DockerfilePathOpt is the value returned by [DockerfilePath].
// It implements [BuildOption] and [EngineBuildOption].
type DockerfilePathOpt struct{ path string }

// DockerfilePath sets the Dockerfile path on [Build] and [BuildImage].
func DockerfilePath(path string) DockerfilePathOpt { return DockerfilePathOpt{path: path} }

func (o DockerfilePathOpt) applyBuild(b *build) error {
	if o.path == "" {
		return cfgErr("", "dockerfile", "DockerfilePath", fmt.Errorf("%w: dockerfile", errEmpty))
	}
	return b.setDockerfile(o.path)
}

// ExternalOpt is the value returned by [External].
// It implements [ComposeFileOption], [ComposeNetOption], [ComposeSecretOption],
// [ComposeVolOption], [FileConfigOption], and [SecretOption].
type ExternalOpt struct{ v bool }

// External marks a project network, volume, secret, or config as unmanaged.
func External(v bool) ExternalOpt { return ExternalOpt{v: v} }

func (o ExternalOpt) applyComposeNet(n *projectNet) error { n.external.assign(o.v); return nil }
func (o ExternalOpt) applyComposeVol(v *projectVol) error { v.external.assign(o.v); return nil }
func (o ExternalOpt) applyComposeSecret(s *projectSecret) error {
	s.external.assign(o.v)
	return nil
}
func (o ExternalOpt) applyComposeFile(f *projectFile) error { f.external.assign(o.v); return nil }
func (o ExternalOpt) applySecret(s *secretRef) error        { s.external.assign(o.v); return nil }
func (o ExternalOpt) applyFileConfig(c *fileConfigRef) error {
	c.external.assign(o.v)
	return nil
}
