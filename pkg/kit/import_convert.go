package kit

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

func composeFromProject(proj *composetypes.Project) (*Compose, error) {
	if proj == nil {
		return nil, cfgErr("compose", "import", "Import", fmt.Errorf("%w: project", errRequired))
	}
	p := &Compose{}
	if proj.Name != "" {
		if err := p.setName(proj.Name); err != nil {
			p.errs = addErr(p.errs, wrapReceiver("compose", err))
		}
	}
	for _, name := range mapKeys(proj.Services) {
		svc := proj.Services[name]
		p.services = append(p.services, serviceFromConfig(name, svc))
	}
	for _, name := range mapKeys(proj.Jobs) {
		job := proj.Jobs[name]
		p.jobs = append(p.jobs, jobFromConfig(name, job))
	}
	for _, name := range mapKeys(proj.Networks) {
		p.networks = append(p.networks, networkFromConfig(name, proj.Networks[name]))
	}
	for _, name := range mapKeys(proj.Volumes) {
		p.volumes = append(p.volumes, volumeFromConfig(name, proj.Volumes[name]))
	}
	for _, name := range mapKeys(proj.Secrets) {
		p.secrets = append(p.secrets, secretFromConfig(name, proj.Secrets[name]))
	}
	for _, name := range mapKeys(proj.Configs) {
		p.configs = append(p.configs, fileFromConfig(name, proj.Configs[name]))
	}
	for _, name := range mapKeys(proj.Models) {
		p.models = append(p.models, modelFromConfig(name, proj.Models[name]))
	}
	return p, nil
}

func serviceFromConfig(name string, cfg composetypes.ServiceConfig) *Service {
	s := &Service{}
	s.name.assign(name)
	specFromContainer(&s.spec, cfg.ContainerSpec)
	s.errs = append(s.errs, specFromWorkload(&s.spec, cfg.WorkloadSpec)...)
	serviceExtrasFromConfig(s, cfg)
	return s
}

func jobFromConfig(name string, cfg composetypes.JobConfig) *Job {
	j := &Job{}
	j.name.assign(name)
	specFromContainer(&j.spec, cfg.ContainerSpec)
	j.errs = append(j.errs, specFromWorkload(&j.spec, cfg.WorkloadSpec)...)
	if cfg.Build != nil {
		j.build.assign(buildFromConfig(*cfg.Build))
	}
	if cfg.PullPolicy != "" {
		j.pullPolicy.assign(cfg.PullPolicy)
	}
	j.envFiles = envFilesFromConfig(cfg.EnvFiles)
	j.secrets = secretsFromConfig(cfg.Secrets)
	j.configs = fileConfigsFromConfig(cfg.Configs)
	j.labelFiles = append([]string(nil), cfg.LabelFiles...)
	if cfg.UseAPISocket {
		j.useAPISocket.assign(true)
	}
	j.depends = dependsFromConfig(cfg.DependsOn)
	return j
}

func specFromContainer(s *spec, cfg composetypes.ContainerSpec) {
	if cfg.Image != "" {
		s.image.assign(cfg.Image)
	}
	if cfg.Command != nil {
		s.command.assign(append([]string(nil), cfg.Command...))
	}
	if cfg.Entrypoint != nil {
		s.entrypoint.assign(append([]string(nil), cfg.Entrypoint...))
	}
	if cfg.User != "" {
		s.user.assign(cfg.User)
	}
	if cfg.WorkingDir != "" {
		s.workdir.assign(cfg.WorkingDir)
	}
	if cfg.Hostname != "" {
		s.hostname.assign(cfg.Hostname)
	}
	if cfg.DomainName != "" {
		s.domainName.assign(cfg.DomainName)
	}
	if cfg.MacAddress != "" {
		s.macAddress.assign(cfg.MacAddress)
	}
	if cfg.StopSignal != "" {
		s.stopSignal.assign(cfg.StopSignal)
	}
	if cfg.Platform != "" {
		s.platform.assign(cfg.Platform)
	}
	if cfg.Privileged {
		s.privileged.assign(true)
	}
	if cfg.ReadOnly {
		s.readOnly.assign(true)
	}
	if cfg.MemLimit != 0 {
		s.memory.assign(int64(cfg.MemLimit))
	}
	if cfg.MemReservation != 0 {
		s.memReserve.assign(int64(cfg.MemReservation))
	}
	if cfg.MemSwapLimit != 0 {
		s.memSwap.assign(int64(cfg.MemSwapLimit))
	}
	if cfg.MemSwappiness != 0 {
		s.memSwapiness.assign(int64(cfg.MemSwappiness))
	}
	if cfg.ShmSize != 0 {
		s.shmSize.assign(int64(cfg.ShmSize))
	}
	if cfg.CPUS != 0 {
		s.cpus.assign(float64(cfg.CPUS))
	}
	if cfg.CPUShares != 0 {
		s.cpuShares.assign(cfg.CPUShares)
	}
	if cfg.CPUPeriod != 0 {
		s.cpuPeriod.assign(cfg.CPUPeriod)
	}
	if cfg.CPUQuota != 0 {
		s.cpuQuota.assign(cfg.CPUQuota)
	}
	if cfg.CPUSet != "" {
		s.cpusetCpus.assign(cfg.CPUSet)
	}
	if cfg.CPUCount != 0 {
		s.cpuCount.assign(cfg.CPUCount)
	}
	if cfg.CPUPercent != 0 {
		s.cpuPercent.assign(int64(cfg.CPUPercent))
	}
	if cfg.CPURTPeriod != 0 {
		s.cpuRTPeriod.assign(cfg.CPURTPeriod)
	}
	if cfg.CPURTRuntime != 0 {
		s.cpuRTRuntime.assign(cfg.CPURTRuntime)
	}
	if cfg.PidsLimit != 0 {
		s.pidsLimit.assign(cfg.PidsLimit)
	}
	if cfg.OomScoreAdj != 0 {
		s.oomScoreAdj.assign(int(cfg.OomScoreAdj))
	}
	if cfg.OomKillDisable {
		s.oomKillDisable.assign(true)
	}
	if cfg.Init != nil {
		s.init.assign(*cfg.Init)
	}
	if cfg.StopGracePeriod != nil {
		s.stopTimeout.assign(int(time.Duration(*cfg.StopGracePeriod) / time.Second))
	}
	for _, c := range cfg.CapAdd {
		s.capAdd = append(s.capAdd, Capability(c))
	}
	for _, c := range cfg.CapDrop {
		s.capDrop = append(s.capDrop, Capability(c))
	}
	if len(cfg.Labels) > 0 {
		s.labels = mapsCopy(cfg.Labels)
	}
	s.env = envFromMapping(cfg.Environment)
	s.volumes = volumesFromConfig(cfg.Volumes)
	s.tmpfs = append([]string(nil), cfg.Tmpfs...)
	s.volumesFrom = append([]string(nil), cfg.VolumesFrom...)
	if cfg.VolumeDriver != "" {
		s.volumeDrv.assign(cfg.VolumeDriver)
	}
	if cfg.NetworkMode != "" {
		s.networkMode.assign(cfg.NetworkMode)
	}
	s.networks = networksFromConfig(cfg.Networks)
	s.dns = append([]string(nil), cfg.DNS...)
	s.dnsOptions = append([]string(nil), cfg.DNSOpts...)
	s.dnsSearch = append([]string(nil), cfg.DNSSearch...)
	for host, ips := range cfg.ExtraHosts {
		ip := ""
		if len(ips) > 0 {
			ip = ips[0]
		}
		s.extraHosts = append(s.extraHosts, extraHost{hostname: host, ip: ip})
	}
	s.groupAdd = append([]string(nil), cfg.GroupAdd...)
	if cfg.Logging != nil {
		l := logging{}
		if cfg.Logging.Driver != "" {
			l.driver.assign(cfg.Logging.Driver)
		}
		l.options = mapsCopy(map[string]string(cfg.Logging.Options))
		s.logging.assign(l)
	}
	if cfg.Ipc != "" {
		s.ipc.assign(cfg.Ipc)
	}
	if cfg.Pid != "" {
		s.pid.assign(cfg.Pid)
	}
	if cfg.Uts != "" {
		s.uts.assign(cfg.Uts)
	}
	if cfg.UserNSMode != "" {
		s.userns.assign(cfg.UserNSMode)
	}
	if cfg.Cgroup != "" {
		s.cgroup.assign(cfg.Cgroup)
	}
	if cfg.CgroupParent != "" {
		s.cgroupParent.assign(cfg.CgroupParent)
	}
	if cfg.Isolation != "" {
		s.isolation.assign(cfg.Isolation)
	}
	if cfg.Runtime != "" {
		s.runtime.assign(cfg.Runtime)
	}
	s.securityOpt = append([]string(nil), cfg.SecurityOpt...)
	if len(cfg.Sysctls) > 0 {
		s.sysctls = mapsCopy(map[string]string(cfg.Sysctls))
	}
	if len(cfg.StorageOpt) > 0 {
		s.storageOpt = mapsCopy(cfg.StorageOpt)
	}
	if len(cfg.Annotations) > 0 {
		s.annotations = mapsCopy(map[string]string(cfg.Annotations))
	}
	if cfg.Umask != nil {
		s.umask.assign(*cfg.Umask)
	}
	s.ulimits = ulimitsFromConfig(cfg.Ulimits)
	s.devices = devicesFromConfig(cfg.Devices)
	s.devCgroup = append([]string(nil), cfg.DeviceCgroupRules...)
	s.gpus = gpusFromConfig(cfg.Gpus)
	s.models = modelsFromService(cfg.Models)
	if cfg.PullRefreshAfter != "" {
		s.pullRefreshAfter.assign(cfg.PullRefreshAfter)
	}
	blkioFromConfig(s, cfg.BlkioConfig)
}

func specFromWorkload(s *spec, cfg composetypes.WorkloadSpec) []error {
	if cfg.Tty {
		s.tty.assign(true)
	}
	if cfg.StdinOpen {
		s.interactive.assign(true)
	}
	if cfg.HealthCheck != nil {
		s.health.assign(healthFromConfig(*cfg.HealthCheck))
	}
	var errs []error
	for _, pc := range cfg.Ports {
		p, err := portFromConfig(pc)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s.ports = append(s.ports, p)
	}
	for _, e := range cfg.Expose {
		ex, err := exposeFromString(e)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s.expose = append(s.expose, ex)
	}
	return errs
}

func serviceExtrasFromConfig(s *Service, cfg composetypes.ServiceConfig) {
	if cfg.Restart != "" {
		s.restart.assign(restartFromString(cfg.Restart))
	}
	s.links = append([]string(nil), cfg.Links...)
	if cfg.Build != nil {
		s.build.assign(buildFromConfig(*cfg.Build))
	}
	if cfg.Deploy != nil {
		s.deploy.assign(deployFromConfig(*cfg.Deploy))
	}
	if cfg.Scale != nil {
		s.scale.assign(*cfg.Scale)
	}
	s.profiles = append([]string(nil), cfg.Profiles...)
	s.envFiles = envFilesFromConfig(cfg.EnvFiles)
	s.secrets = secretsFromConfig(cfg.Secrets)
	s.configs = fileConfigsFromConfig(cfg.Configs)
	if cfg.PullPolicy != "" {
		s.pullPolicy.assign(cfg.PullPolicy)
	}
	if cfg.ContainerName != "" {
		s.containerName.assign(cfg.ContainerName)
	}
	s.labelFiles = append([]string(nil), cfg.LabelFiles...)
	if cfg.UseAPISocket {
		s.useAPISocket.assign(true)
	}
	if cfg.Attach != nil {
		s.attach.assign(*cfg.Attach)
	}
	if cfg.Extends != nil {
		if cfg.Extends.Service != "" {
			s.extends.assign(cfg.Extends.Service)
		}
		if cfg.Extends.File != "" {
			s.extendsFile.assign(cfg.Extends.File)
		}
	}
	if cfg.CredentialSpec != nil {
		if cfg.CredentialSpec.File != "" {
			s.credFile.assign(cfg.CredentialSpec.File)
		}
		if cfg.CredentialSpec.Registry != "" {
			s.credRegistry.assign(cfg.CredentialSpec.Registry)
		}
		if cfg.CredentialSpec.Config != "" {
			s.credConfig.assign(cfg.CredentialSpec.Config)
		}
	}
	if cfg.Provider != nil {
		if cfg.Provider.Type != "" {
			s.provider.assign(cfg.Provider.Type)
		}
		if len(cfg.Provider.Options) > 0 {
			s.providerOpts = map[string][]string{}
			for k, vs := range cfg.Provider.Options {
				s.providerOpts[k] = append([]string(nil), vs...)
			}
		}
	}
	for _, h := range cfg.PreStart {
		sp := spec{}
		specFromContainer(&sp, h.ContainerSpec)
		if h.PerReplica {
			sp.perReplica.assign(true)
		}
		s.preStart = append(s.preStart, sp)
	}
	for _, h := range cfg.PostStart {
		s.postStart = append(s.postStart, hookFromConfig(h))
	}
	for _, h := range cfg.PreStop {
		s.preStop = append(s.preStop, hookFromConfig(h))
	}
	s.externalLinks = append([]string(nil), cfg.ExternalLinks...)
	if cfg.Develop != nil {
		s.develop.assign(developFromConfig(*cfg.Develop))
	}
	s.depends = dependsFromConfig(cfg.DependsOn)
}

func healthFromConfig(cfg composetypes.HealthCheckConfig) health {
	var h health
	if cfg.Disable {
		h.disabled.assign(true)
		return h
	}
	if len(cfg.Test) == 1 && cfg.Test[0] == "NONE" {
		h.disabled.assign(true)
		return h
	}
	if len(cfg.Test) > 0 {
		h.test.assign(append([]string(nil), cfg.Test...))
	}
	if cfg.Interval != nil {
		h.interval.assign(time.Duration(*cfg.Interval))
	}
	if cfg.Timeout != nil {
		h.timeout.assign(time.Duration(*cfg.Timeout))
	}
	if cfg.StartPeriod != nil {
		h.startPeriod.assign(time.Duration(*cfg.StartPeriod))
	}
	if cfg.StartInterval != nil {
		h.startInterval.assign(time.Duration(*cfg.StartInterval))
	}
	if cfg.Retries != nil {
		h.retries.assign(int(*cfg.Retries))
	}
	return h
}

func portFromConfig(pc composetypes.ServicePortConfig) (port, error) {
	var p port
	if pc.Protocol != "" {
		p.protocol.assign(Proto(pc.Protocol))
	}
	if pc.HostIP != "" {
		p.hostIP.assign(pc.HostIP)
	}
	if pc.Target != 0 {
		if pc.Target > math.MaxUint16 {
			return port{}, cfgErr("service", "ports.target", "ContainerPort", fmt.Errorf("%w: container port %d", errInvalid, pc.Target))
		}
		p.containerPort.assign(uint16(pc.Target))
	}
	if pc.Published != "" {
		if n, err := strconv.ParseUint(pc.Published, 10, 16); err == nil {
			p.hostPort.assign(uint16(n))
		} else {
			p.published.assign(pc.Published)
		}
	}
	if pc.Name != "" {
		p.portName.assign(pc.Name)
	}
	if pc.AppProtocol != "" {
		p.appProtocol.assign(pc.AppProtocol)
	}
	if pc.Mode != "" && pc.Mode != "ingress" {
		p.portMode.assign(pc.Mode)
	}
	return p, nil
}

func exposeFromString(s string) (exposePort, error) {
	port, proto, ok := strings.Cut(s, "/")
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return exposePort{}, cfgErr("service", "expose", "Expose", fmt.Errorf("%w: expose %q", errInvalid, s))
	}
	e := exposePort{port: uint16(n), protocol: TCP}
	if ok && proto != "" {
		e.protocol = Proto(proto)
	}
	return e, nil
}

func volumesFromConfig(in []composetypes.ServiceVolumeConfig) []volume {
	out := make([]volume, 0, len(in))
	for _, cv := range in {
		v := volume{}
		if cv.Type != "" {
			v.kind.assign(VolumeKind(cv.Type))
		}
		if cv.Source != "" {
			v.source.assign(cv.Source)
		}
		if cv.Target != "" {
			v.target.assign(cv.Target)
		}
		if cv.ReadOnly {
			v.readOnly.assign(true)
		}
		if cv.Tmpfs != nil {
			if cv.Tmpfs.Size != 0 {
				v.size.assign(int64(cv.Tmpfs.Size))
			}
			if cv.Tmpfs.Mode != 0 {
				v.mode.assign(cv.Tmpfs.Mode)
			}
		}
		if cv.Bind != nil {
			if cv.Bind.Propagation != "" {
				v.prop.assign(cv.Bind.Propagation)
			}
			if cv.Bind.SELinux != "" {
				v.selinux.assign(cv.Bind.SELinux)
			}
			v.createHostPath.assign(bool(cv.Bind.CreateHostPath))
			if cv.Bind.Recursive != "" {
				v.recursive.assign(cv.Bind.Recursive)
			}
		}
		if cv.Volume != nil {
			if cv.Volume.NoCopy {
				v.noCopy.assign(true)
			}
			if cv.Volume.Subpath != "" {
				v.subPath.assign(cv.Volume.Subpath)
			}
		}
		out = append(out, v)
	}
	return out
}

func networksFromConfig(in map[string]*composetypes.ServiceNetworkConfig) []netAttach {
	out := make([]netAttach, 0, len(in))
	for _, name := range mapKeys(in) {
		n := netAttach{}
		n.name.assign(name)
		cfg := in[name]
		if cfg != nil {
			n.aliases = append([]string(nil), cfg.Aliases...)
			n.driverOpts = mapsCopy(cfg.DriverOpts)
			if cfg.Ipv4Address != "" {
				n.ipv4.assign(cfg.Ipv4Address)
			}
			if cfg.Ipv6Address != "" {
				n.ipv6.assign(cfg.Ipv6Address)
			}
			if cfg.MacAddress != "" {
				n.mac.assign(cfg.MacAddress)
			}
			if cfg.Priority != 0 {
				n.priority.assign(cfg.Priority)
			}
			if cfg.GatewayPriority != 0 {
				n.gwPriority.assign(cfg.GatewayPriority)
			}
			if cfg.InterfaceName != "" {
				n.ifName.assign(cfg.InterfaceName)
			}
			n.linkLocal = append([]string(nil), cfg.LinkLocalIPs...)
			if cfg.UseServiceNameAlias != nil {
				n.useAlias.assign(*cfg.UseServiceNameAlias)
			}
		}
		out = append(out, n)
	}
	return out
}

func devicesFromConfig(in []composetypes.DeviceMapping) []device {
	out := make([]device, 0, len(in))
	for _, m := range in {
		d := device{}
		if m.Source != "" {
			d.host.assign(m.Source)
		}
		if m.Target != "" {
			d.container.assign(m.Target)
		}
		if m.Permissions != "" {
			d.perms.assign(m.Permissions)
		}
		out = append(out, d)
	}
	return out
}

func gpusFromConfig(in []composetypes.DeviceRequest) []gpu {
	out := make([]gpu, 0, len(in))
	for _, req := range in {
		out = append(out, gpuFromRequest(req))
	}
	return out
}

func gpuFromRequest(req composetypes.DeviceRequest) gpu {
	g := gpu{}
	if req.Driver != "" {
		g.driver.assign(req.Driver)
	}
	if req.Count != 0 {
		g.count.assign(int(req.Count))
	}
	g.ids = append([]string(nil), req.IDs...)
	g.capabilities = append([]string(nil), req.Capabilities...)
	return g
}

func modelsFromService(in map[string]*composetypes.ServiceModelConfig) []serviceModel {
	out := make([]serviceModel, 0, len(in))
	for _, name := range mapKeys(in) {
		m := serviceModel{}
		m.name.assign(name)
		if cfg := in[name]; cfg != nil {
			if cfg.EndpointVariable != "" {
				m.endpointVar.assign(cfg.EndpointVariable)
			}
			if cfg.ModelVariable != "" {
				m.modelVar.assign(cfg.ModelVariable)
			}
		}
		out = append(out, m)
	}
	return out
}

func ulimitsFromConfig(in map[string]*composetypes.UlimitsConfig) []ulimit {
	out := make([]ulimit, 0, len(in))
	for _, name := range mapKeys(in) {
		u := ulimit{}
		u.name.assign(name)
		if cfg := in[name]; cfg != nil {
			if cfg.Soft != 0 {
				u.soft.assign(int64(cfg.Soft))
			}
			if cfg.Hard != 0 {
				u.hard.assign(int64(cfg.Hard))
			}
			if cfg.Single != 0 && cfg.Soft == 0 && cfg.Hard == 0 {
				u.soft.assign(int64(cfg.Single))
				u.hard.assign(int64(cfg.Single))
			}
		}
		out = append(out, u)
	}
	return out
}

func blkioFromConfig(s *spec, bc *composetypes.BlkioConfig) {
	if bc == nil {
		return
	}
	if bc.Weight != 0 {
		s.blkioWeight.assign(bc.Weight)
	}
	for _, d := range bc.WeightDevice {
		s.blkioWDev = append(s.blkioWDev, blkioWeight{path: d.Path, weight: d.Weight})
	}
	for _, d := range bc.DeviceReadBps {
		s.blkioReadBps = append(s.blkioReadBps, blkioRate{path: d.Path, rate: uint64(d.Rate)})
	}
	for _, d := range bc.DeviceWriteBps {
		s.blkioWriteBps = append(s.blkioWriteBps, blkioRate{path: d.Path, rate: uint64(d.Rate)})
	}
	for _, d := range bc.DeviceReadIOps {
		s.blkioReadIOps = append(s.blkioReadIOps, blkioRate{path: d.Path, rate: uint64(d.Rate)})
	}
	for _, d := range bc.DeviceWriteIOps {
		s.blkioWriteIOps = append(s.blkioWriteIOps, blkioRate{path: d.Path, rate: uint64(d.Rate)})
	}
}

func envFromMapping(in composetypes.MappingWithEquals) []envVar {
	out := make([]envVar, 0, len(in))
	for _, k := range mapKeys(in) {
		ev := envVar{key: k}
		if v := in[k]; v != nil {
			ev.value.assign(*v)
		}
		out = append(out, ev)
	}
	return out
}

func envFilesFromConfig(in []composetypes.EnvFile) []envFileRef {
	out := make([]envFileRef, 0, len(in))
	for _, e := range in {
		ref := envFileRef{}
		ref.path.assign(e.Path)
		if !bool(e.Required) {
			ref.required.assign(false)
		}
		if e.Format != "" {
			ref.format.assign(e.Format)
		}
		out = append(out, ref)
	}
	return out
}

func secretsFromConfig(in []composetypes.ServiceSecretConfig) []secretRef {
	out := make([]secretRef, 0, len(in))
	for _, sc := range in {
		r := secretRef{}
		if sc.Source != "" {
			r.source.assign(sc.Source)
		}
		if sc.Target != "" {
			r.target.assign(sc.Target)
		}
		if sc.UID != "" {
			r.uid.assign(sc.UID)
		}
		if sc.GID != "" {
			r.gid.assign(sc.GID)
		}
		if sc.Mode != nil {
			r.mode.assign(uint32(*sc.Mode))
		}
		out = append(out, r)
	}
	return out
}

func fileConfigsFromConfig(in []composetypes.ServiceConfigObjConfig) []fileConfigRef {
	out := make([]fileConfigRef, 0, len(in))
	for _, sc := range in {
		r := fileConfigRef{}
		if sc.Source != "" {
			r.source.assign(sc.Source)
		}
		if sc.Target != "" {
			r.target.assign(sc.Target)
		}
		if sc.UID != "" {
			r.uid.assign(sc.UID)
		}
		if sc.GID != "" {
			r.gid.assign(sc.GID)
		}
		if sc.Mode != nil {
			r.mode.assign(uint32(*sc.Mode))
		}
		out = append(out, r)
	}
	return out
}

func dependsFromConfig(in composetypes.DependsOnConfig) []depend {
	out := make([]depend, 0, len(in))
	for _, name := range mapKeys(in) {
		d := depend{}
		d.name.assign(name)
		dep := in[name]
		if dep.Condition != "" {
			d.condition.assign(DependCondition(dep.Condition))
		}
		if dep.Restart {
			d.restart.assign(true)
		}
		if !dep.Required {
			d.required.assign(false)
		}
		out = append(out, d)
	}
	return out
}

func hookFromConfig(h composetypes.ServiceHook) serviceHook {
	out := serviceHook{cmd: append([]string(nil), h.Command...)}
	if h.User != "" {
		out.user.assign(h.User)
	}
	if h.Privileged {
		out.privileged.assign(true)
	}
	if h.WorkingDir != "" {
		out.workdir.assign(h.WorkingDir)
	}
	out.env = envFromMapping(h.Environment)
	return out
}

func developFromConfig(cfg composetypes.DevelopConfig) develop {
	var d develop
	for _, w := range cfg.Watch {
		t := watchTrigger{}
		if w.Path != "" {
			t.path.assign(w.Path)
		}
		if w.Action != "" {
			t.action.assign(WatchAction(w.Action))
		}
		if w.Target != "" {
			t.target.assign(w.Target)
		}
		t.include = append([]string(nil), w.Include...)
		t.ignore = append([]string(nil), w.Ignore...)
		if w.InitialSync {
			t.initialSync.assign(true)
		}
		t.exec = hookFromConfig(w.Exec)
		d.watch = append(d.watch, t)
	}
	return d
}

func buildFromConfig(cfg composetypes.BuildConfig) build {
	var b build
	if cfg.Context != "" {
		b.source.assign(cfg.Context)
	}
	if cfg.Dockerfile != "" {
		b.dockerfile.assign(cfg.Dockerfile)
	}
	if cfg.DockerfileInline != "" {
		b.inline.assign(cfg.DockerfileInline)
	}
	if cfg.Target != "" {
		b.target.assign(cfg.Target)
	}
	b.args = envFromMapping(cfg.Args)
	if len(cfg.Labels) > 0 {
		b.labels = mapsCopy(cfg.Labels)
	}
	if cfg.Pull {
		b.pull.assign(true)
	}
	if cfg.NoCache {
		b.noCache.assign(true)
	}
	if cfg.Network != "" {
		b.network.assign(cfg.Network)
	}
	if cfg.ShmSize > 0 {
		b.shmSize.assign(int64(cfg.ShmSize))
	}
	if cfg.Privileged {
		b.privileged.assign(true)
	}
	b.tags = append([]string(nil), cfg.Tags...)
	b.platforms = append([]string(nil), cfg.Platforms...)
	for host, ips := range cfg.ExtraHosts {
		ip := ""
		if len(ips) > 0 {
			ip = ips[0]
		}
		b.extraHosts = append(b.extraHosts, extraHost{hostname: host, ip: ip})
	}
	if cfg.Isolation != "" {
		b.isolation.assign(cfg.Isolation)
	}
	b.entitlements = append([]string(nil), cfg.Entitlements...)
	if cfg.Provenance != "" {
		b.provenance.assign(cfg.Provenance)
	}
	if cfg.SBOM != "" {
		b.sbom.assign(cfg.SBOM)
	}
	for _, k := range cfg.SSH {
		b.ssh = append(b.ssh, sshKey{id: k.ID, path: k.Path})
	}
	b.cacheFrom = append([]string(nil), cfg.CacheFrom...)
	b.cacheTo = append([]string(nil), cfg.CacheTo...)
	b.noCacheFilter = append([]string(nil), cfg.NoCacheFilter...)
	if len(cfg.AdditionalContexts) > 0 {
		b.extraContexts = mapsCopy(cfg.AdditionalContexts)
	}
	b.secrets = secretsFromConfig(cfg.Secrets)
	b.ulimits = ulimitsFromConfig(cfg.Ulimits)
	return b
}

func deployFromConfig(cfg composetypes.DeployConfig) deploy {
	var d deploy
	if cfg.Mode != "" {
		d.mode.assign(cfg.Mode)
	}
	if cfg.Replicas != nil {
		d.replicas.assign(*cfg.Replicas)
	}
	if len(cfg.Labels) > 0 {
		d.labels = mapsCopy(cfg.Labels)
	}
	if cfg.Resources.Limits != nil {
		lim := *cfg.Resources.Limits
		if lim.MemoryBytes != 0 {
			d.memory.assign(int64(lim.MemoryBytes))
		}
		if lim.NanoCPUs != 0 {
			d.cpus.assign(float64(lim.NanoCPUs))
		}
		if lim.Pids != 0 {
			d.pids.assign(lim.Pids)
		}
		for _, g := range lim.Devices {
			d.devices = append(d.devices, gpuFromRequest(g))
		}
		d.generic = genericFromConfig(lim.GenericResources)
	}
	if cfg.Resources.Reservations != nil {
		res := resourceFromConfig(*cfg.Resources.Reservations)
		d.reserve.assign(res)
	}
	if cfg.UpdateConfig != nil {
		d.update.assign(updateFromConfig(*cfg.UpdateConfig))
	}
	if cfg.RollbackConfig != nil {
		d.rollback.assign(updateFromConfig(*cfg.RollbackConfig))
	}
	if cfg.RestartPolicy != nil {
		d.restart.assign(deployRestartFromConfig(*cfg.RestartPolicy))
	}
	if len(cfg.Placement.Constraints) > 0 || len(cfg.Placement.Preferences) > 0 || cfg.Placement.MaxReplicas != 0 {
		pl := placement{constraints: append([]string(nil), cfg.Placement.Constraints...)}
		for _, pref := range cfg.Placement.Preferences {
			if pref.Spread != "" {
				pl.spreads = append(pl.spreads, pref.Spread)
			}
		}
		if cfg.Placement.MaxReplicas != 0 {
			pl.maxReplicas.assign(cfg.Placement.MaxReplicas)
		}
		d.placement.assign(pl)
	}
	if cfg.EndpointMode != "" {
		d.endpointMode.assign(cfg.EndpointMode)
	}
	return d
}

func resourceFromConfig(cfg composetypes.Resource) resource {
	var r resource
	if cfg.MemoryBytes != 0 {
		r.memory.assign(int64(cfg.MemoryBytes))
	}
	if cfg.NanoCPUs != 0 {
		r.cpus.assign(float64(cfg.NanoCPUs))
	}
	if cfg.Pids != 0 {
		r.pids.assign(cfg.Pids)
	}
	for _, g := range cfg.Devices {
		r.devices = append(r.devices, gpuFromRequest(g))
	}
	r.generic = genericFromConfig(cfg.GenericResources)
	return r
}

func genericFromConfig(in []composetypes.GenericResource) []discreteRes {
	var out []discreteRes
	for _, g := range in {
		if g.DiscreteResourceSpec == nil {
			continue
		}
		out = append(out, discreteRes{kind: g.DiscreteResourceSpec.Kind, value: g.DiscreteResourceSpec.Value})
	}
	return out
}

func updateFromConfig(cfg composetypes.UpdateConfig) updateCfg {
	var u updateCfg
	if cfg.Parallelism != nil {
		u.parallelism.assign(*cfg.Parallelism)
	}
	if cfg.Delay != 0 {
		u.delay.assign(time.Duration(cfg.Delay))
	}
	if cfg.FailureAction != "" {
		u.failureAction.assign(cfg.FailureAction)
	}
	if cfg.Monitor != 0 {
		u.monitor.assign(time.Duration(cfg.Monitor))
	}
	if cfg.MaxFailureRatio != 0 {
		u.maxFailureRatio.assign(cfg.MaxFailureRatio)
	}
	if cfg.Order != "" {
		u.order.assign(cfg.Order)
	}
	return u
}

func deployRestartFromConfig(cfg composetypes.RestartPolicy) restart {
	var r restart
	switch cfg.Condition {
	case "none":
		r.policy.assign(NoRestart)
	case "on-failure":
		r.policy.assign(OnFailure)
	case "any":
		r.policy.assign(Always)
	}
	if cfg.Delay != nil {
		r.delay.assign(time.Duration(*cfg.Delay))
	}
	if cfg.MaxAttempts != nil {
		r.maxRetries.assign(int(*cfg.MaxAttempts))
	}
	if cfg.Window != nil {
		r.window.assign(time.Duration(*cfg.Window))
	}
	return r
}

func restartFromString(s string) restart {
	var r restart
	kind, n, ok := strings.Cut(s, ":")
	if !ok {
		kind = s
	}
	switch kind {
	case "no", "none":
		r.policy.assign(NoRestart)
	case "on-failure":
		r.policy.assign(OnFailure)
		if ok {
			if v, err := strconv.Atoi(n); err == nil {
				r.maxRetries.assign(v)
			}
		}
	case "always":
		r.policy.assign(Always)
	case "unless-stopped":
		r.policy.assign(UnlessStopped)
	default:
		r.policy.assign(RestartPolicy(s))
	}
	return r
}

func networkFromConfig(name string, cfg composetypes.NetworkConfig) projectNet {
	n := projectNet{}
	n.name.assign(name)
	if cfg.Driver != "" {
		n.driver.assign(cfg.Driver)
	}
	if cfg.Internal {
		n.internal.assign(true)
	}
	if cfg.Attachable {
		n.attachable.assign(true)
	}
	if bool(cfg.External) {
		n.external.assign(true)
	}
	if cfg.EnableIPv4 != nil {
		n.ipv4.assign(*cfg.EnableIPv4)
	}
	if cfg.EnableIPv6 != nil {
		n.ipv6.assign(*cfg.EnableIPv6)
	}
	if len(cfg.Labels) > 0 {
		n.labels = mapsCopy(cfg.Labels)
	}
	n.driverOpts = mapsCopy(cfg.DriverOpts)
	if cfg.Ipam.Driver != "" {
		n.ipamDriver.assign(cfg.Ipam.Driver)
	}
	n.ipamOpts = mapsCopy(cfg.Ipam.Options)
	if len(cfg.Ipam.Config) > 0 {
		pool := cfg.Ipam.Config[0]
		if pool != nil {
			if pool.Subnet != "" {
				n.subnet.assign(pool.Subnet)
			}
			if pool.Gateway != "" {
				n.gateway.assign(pool.Gateway)
			}
			if pool.IPRange != "" {
				n.ipRange.assign(pool.IPRange)
			}
			if len(pool.AuxiliaryAddresses) > 0 {
				n.aux = mapsCopy(pool.AuxiliaryAddresses)
			}
		}
	}
	return n
}

func volumeFromConfig(name string, cfg composetypes.VolumeConfig) projectVol {
	v := projectVol{}
	v.name.assign(name)
	if cfg.Driver != "" {
		v.driver.assign(cfg.Driver)
	}
	if bool(cfg.External) {
		v.external.assign(true)
	}
	if len(cfg.Labels) > 0 {
		v.labels = mapsCopy(cfg.Labels)
	}
	v.driverOpts = mapsCopy(cfg.DriverOpts)
	return v
}

func secretFromConfig(name string, cfg composetypes.SecretConfig) projectSecret {
	s := projectSecret{}
	s.name.assign(name)
	if cfg.File != "" {
		s.file.assign(cfg.File)
	}
	if cfg.Environment != "" {
		s.env.assign(cfg.Environment)
	}
	if cfg.Content != "" {
		s.content.assign(cfg.Content)
	}
	if bool(cfg.External) {
		s.external.assign(true)
	}
	if cfg.Driver != "" {
		s.driver.assign(cfg.Driver)
	}
	if len(cfg.Labels) > 0 {
		s.labels = mapsCopy(cfg.Labels)
	}
	if cfg.TemplateDriver != "" {
		s.templateDriver.assign(cfg.TemplateDriver)
	}
	return s
}

func fileFromConfig(name string, cfg composetypes.ConfigObjConfig) projectFile {
	f := projectFile{}
	f.name.assign(name)
	if cfg.File != "" {
		f.file.assign(cfg.File)
	}
	if cfg.Environment != "" {
		f.env.assign(cfg.Environment)
	}
	if cfg.Content != "" {
		f.content.assign(cfg.Content)
	}
	if bool(cfg.External) {
		f.external.assign(true)
	}
	if cfg.Driver != "" {
		f.driver.assign(cfg.Driver)
	}
	if len(cfg.Labels) > 0 {
		f.labels = mapsCopy(cfg.Labels)
	}
	if cfg.TemplateDriver != "" {
		f.templateDriver.assign(cfg.TemplateDriver)
	}
	return f
}

func modelFromConfig(name string, cfg composetypes.ModelConfig) projectModel {
	m := projectModel{}
	m.name.assign(name)
	if cfg.Model != "" {
		m.model.assign(cfg.Model)
	}
	if cfg.ContextSize != 0 {
		m.contextSize.assign(cfg.ContextSize)
	}
	m.runtimeFlags = append([]string(nil), cfg.RuntimeFlags...)
	return m
}

func mapKeys[M ~map[string]T, T any](m M) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
