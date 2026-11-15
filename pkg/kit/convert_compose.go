package kit

import (
	"fmt"
	"time"

	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

func (s spec) applyCompose(svc *composetypes.ServiceConfig) error {
	if s.image.ok() {
		svc.Image = s.image.val
	}
	if s.command.ok() {
		svc.Command = append([]string(nil), s.command.val...)
	}
	if s.entrypoint.ok() {
		svc.Entrypoint = append([]string(nil), s.entrypoint.val...)
	}
	if s.user.ok() {
		svc.User = s.user.val
	}
	if s.workdir.ok() {
		svc.WorkingDir = s.workdir.val
	}
	if s.hostname.ok() {
		svc.Hostname = s.hostname.val
	}
	if s.domainName.ok() {
		svc.DomainName = s.domainName.val
	}
	if s.macAddress.ok() {
		svc.MacAddress = s.macAddress.val
	}
	if s.stopSignal.ok() {
		svc.StopSignal = s.stopSignal.val
	}
	if s.platform.ok() {
		svc.Platform = s.platform.val
	}
	if s.privileged.ok() {
		svc.Privileged = s.privileged.val
	}
	if s.readOnly.ok() {
		svc.ReadOnly = s.readOnly.val
	}
	if s.tty.ok() {
		svc.Tty = s.tty.val
	}
	if s.interactive.ok() {
		svc.StdinOpen = s.interactive.val
	}
	if s.memory.ok() {
		svc.MemLimit = composetypes.UnitBytes(s.memory.val)
	}
	if s.memReserve.ok() {
		svc.MemReservation = composetypes.UnitBytes(s.memReserve.val)
	}
	if s.memSwap.ok() {
		svc.MemSwapLimit = composetypes.UnitBytes(s.memSwap.val)
	}
	if s.memSwapiness.ok() {
		svc.MemSwappiness = composetypes.UnitBytes(s.memSwapiness.val)
	}
	if s.shmSize.ok() {
		svc.ShmSize = composetypes.UnitBytes(s.shmSize.val)
	}
	if s.cpus.ok() {
		svc.CPUS = float32(s.cpus.val)
	}
	if s.cpuShares.ok() {
		svc.CPUShares = s.cpuShares.val
	}
	if s.cpuPeriod.ok() {
		svc.CPUPeriod = s.cpuPeriod.val
	}
	if s.cpuQuota.ok() {
		svc.CPUQuota = s.cpuQuota.val
	}
	if s.cpusetCpus.ok() {
		svc.CPUSet = s.cpusetCpus.val
	}
	if s.cpuCount.ok() {
		svc.CPUCount = s.cpuCount.val
	}
	if s.cpuPercent.ok() {
		svc.CPUPercent = float32(s.cpuPercent.val)
	}
	if s.cpuRTPeriod.ok() {
		svc.CPURTPeriod = s.cpuRTPeriod.val
	}
	if s.cpuRTRuntime.ok() {
		svc.CPURTRuntime = s.cpuRTRuntime.val
	}
	if s.pidsLimit.ok() {
		svc.PidsLimit = s.pidsLimit.val
	}
	if s.oomScoreAdj.ok() {
		svc.OomScoreAdj = int64(s.oomScoreAdj.val)
	}
	if s.oomKillDisable.ok() {
		svc.OomKillDisable = s.oomKillDisable.val
	}
	if s.init.ok() {
		v := s.init.val
		svc.Init = &v
	}
	if s.stopTimeout.ok() {
		d := composetypes.Duration(time.Duration(s.stopTimeout.val) * time.Second)
		svc.StopGracePeriod = &d
	}
	if len(s.capAdd) > 0 {
		svc.CapAdd = caps(s.capAdd)
	}
	if len(s.capDrop) > 0 {
		svc.CapDrop = caps(s.capDrop)
	}
	if s.labels != nil {
		svc.Labels = composetypes.Labels(mapsCopy(s.labels))
	}
	if len(s.env) > 0 {
		env := composetypes.MappingWithEquals{}
		for _, e := range s.env {
			v := e.value.val
			env[e.key] = &v
		}
		svc.Environment = env
	}
	if s.restart.ok() {
		name := string(s.restart.val.policy.val)
		if s.restart.val.maxRetries.ok() && s.restart.val.policy.val == OnFailure {
			name = fmt.Sprintf("on-failure:%d", s.restart.val.maxRetries.val)
		}
		svc.Restart = name
	}
	if err := s.applyComposeHealth(svc); err != nil {
		return err
	}
	for _, p := range s.ports {
		pc := composetypes.ServicePortConfig{
			Mode:      "ingress",
			HostIP:    p.hostIPOrDefault(),
			Target:    uint32(p.containerPort.val),
			Published: p.hostPortString(),
			Protocol:  string(p.proto()),
		}
		if p.portMode.ok() {
			pc.Mode = p.portMode.val
		}
		if p.portName.ok() {
			pc.Name = p.portName.val
		}
		if p.appProtocol.ok() {
			pc.AppProtocol = p.appProtocol.val
		}
		svc.Ports = append(svc.Ports, pc)
	}
	for _, e := range s.expose {
		svc.Expose = append(svc.Expose, formatExpose(e))
	}
	s.applyComposeMounts(svc)
	if len(s.tmpfs) > 0 {
		svc.Tmpfs = composetypes.StringList(append([]string(nil), s.tmpfs...))
	}
	s.applyComposeNetwork(svc)
	s.applyComposeDNS(svc)
	s.applyComposeDevices(svc)
	if s.logging.ok() {
		svc.Logging = &composetypes.LoggingConfig{
			Driver:  s.logging.val.driver.val,
			Options: composetypes.Options(mapsCopy(s.logging.val.options)),
		}
	}
	if s.ipc.ok() {
		svc.Ipc = s.ipc.val
	}
	if s.pid.ok() {
		svc.Pid = s.pid.val
	}
	if s.uts.ok() {
		svc.Uts = s.uts.val
	}
	if s.userns.ok() {
		svc.UserNSMode = s.userns.val
	}
	if s.cgroup.ok() {
		svc.Cgroup = s.cgroup.val
	}
	if s.cgroupParent.ok() {
		svc.CgroupParent = s.cgroupParent.val
	}
	if s.isolation.ok() {
		svc.Isolation = s.isolation.val
	}
	if s.runtime.ok() {
		svc.Runtime = s.runtime.val
	}
	if s.volumeDrv.ok() {
		svc.VolumeDriver = s.volumeDrv.val
	}
	if len(s.volumesFrom) > 0 {
		svc.VolumesFrom = append([]string(nil), s.volumesFrom...)
	}
	if len(s.groupAdd) > 0 {
		svc.GroupAdd = append([]string(nil), s.groupAdd...)
	}
	if len(s.securityOpt) > 0 {
		svc.SecurityOpt = append([]string(nil), s.securityOpt...)
	}
	if s.sysctls != nil {
		svc.Sysctls = composetypes.Mapping(mapsCopy(s.sysctls))
	}
	if s.storageOpt != nil {
		svc.StorageOpt = mapsCopy(s.storageOpt)
	}
	if s.annotations != nil {
		svc.Annotations = composetypes.Mapping(mapsCopy(s.annotations))
	}
	if s.umask.ok() {
		v := s.umask.val
		svc.Umask = &v
	}
	if len(s.links) > 0 {
		svc.Links = append([]string(nil), s.links...)
	}
	if len(s.ulimits) > 0 {
		svc.Ulimits = map[string]*composetypes.UlimitsConfig{}
		for _, u := range s.ulimits {
			cfg := &composetypes.UlimitsConfig{}
			if u.soft.ok() {
				cfg.Soft = int(u.soft.val)
			}
			if u.hard.ok() {
				cfg.Hard = int(u.hard.val)
			}
			svc.Ulimits[u.name.val] = cfg
		}
	}
	if s.networkMode.ok() {
		svc.NetworkMode = s.networkMode.val
	}
	if s.networkDisabled.ok() && s.networkDisabled.val && !s.networkMode.ok() {
		svc.NetworkMode = "none"
	}
	if s.pullRefreshAfter.ok() {
		svc.PullRefreshAfter = s.pullRefreshAfter.val
	}
	if len(s.models) > 0 {
		svc.Models = map[string]*composetypes.ServiceModelConfig{}
		for _, m := range s.models {
			cfg := &composetypes.ServiceModelConfig{}
			if m.endpointVar.ok() {
				cfg.EndpointVariable = m.endpointVar.val
			}
			if m.modelVar.ok() {
				cfg.ModelVariable = m.modelVar.val
			}
			svc.Models[m.name.val] = cfg
		}
	}
	s.applyComposeBlkio(svc)
	return nil
}

func (s spec) applyComposeBlkio(svc *composetypes.ServiceConfig) {
	if s.blkioWeight.ok() && len(s.blkioWDev) == 0 && len(s.blkioReadBps) == 0 && len(s.blkioWriteBps) == 0 && len(s.blkioReadIOps) == 0 && len(s.blkioWriteIOps) == 0 {
		svc.BlkioConfig = &composetypes.BlkioConfig{Weight: s.blkioWeight.val}
		return
	}
	if !s.blkioWeight.ok() && len(s.blkioWDev) == 0 && len(s.blkioReadBps) == 0 && len(s.blkioWriteBps) == 0 && len(s.blkioReadIOps) == 0 && len(s.blkioWriteIOps) == 0 {
		return
	}
	bc := &composetypes.BlkioConfig{}
	if s.blkioWeight.ok() {
		bc.Weight = s.blkioWeight.val
	}
	for _, d := range s.blkioWDev {
		bc.WeightDevice = append(bc.WeightDevice, composetypes.WeightDevice{Path: d.path, Weight: d.weight})
	}
	for _, d := range s.blkioReadBps {
		bc.DeviceReadBps = append(bc.DeviceReadBps, composetypes.ThrottleDevice{Path: d.path, Rate: composetypes.UnitBytes(d.rate)})
	}
	for _, d := range s.blkioWriteBps {
		bc.DeviceWriteBps = append(bc.DeviceWriteBps, composetypes.ThrottleDevice{Path: d.path, Rate: composetypes.UnitBytes(d.rate)})
	}
	for _, d := range s.blkioReadIOps {
		bc.DeviceReadIOps = append(bc.DeviceReadIOps, composetypes.ThrottleDevice{Path: d.path, Rate: composetypes.UnitBytes(d.rate)})
	}
	for _, d := range s.blkioWriteIOps {
		bc.DeviceWriteIOps = append(bc.DeviceWriteIOps, composetypes.ThrottleDevice{Path: d.path, Rate: composetypes.UnitBytes(d.rate)})
	}
	svc.BlkioConfig = bc
}

func (s spec) applyComposeMounts(svc *composetypes.ServiceConfig) {
	for _, v := range s.volumes {
		kind := v.kindOrDefault()
		cv := composetypes.ServiceVolumeConfig{
			Type:     string(kind),
			ReadOnly: v.readOnly.ok() && v.readOnly.val,
		}
		if v.source.ok() {
			cv.Source = v.source.val
		}
		if v.target.ok() {
			cv.Target = v.target.val
		}
		if kind == Tmpfs {
			t := &composetypes.ServiceVolumeTmpfs{}
			if v.size.ok() {
				t.Size = composetypes.UnitBytes(v.size.val)
			}
			if v.mode.ok() {
				t.Mode = v.mode.val
			}
			cv.Tmpfs = t
		}
		if kind == Bind && (v.prop.ok() || v.selinux.ok() || v.createHostPath.ok() || v.recursive.ok()) {
			b := &composetypes.ServiceVolumeBind{}
			if v.prop.ok() {
				b.Propagation = v.prop.val
			}
			if v.selinux.ok() {
				b.SELinux = v.selinux.val
			}
			if v.createHostPath.ok() {
				b.CreateHostPath = composetypes.OptOut(v.createHostPath.val)
			}
			if v.recursive.ok() {
				b.Recursive = v.recursive.val
			}
			cv.Bind = b
		}
		if kind == NamedVolume && (v.noCopy.ok() || v.subPath.ok()) {
			vo := &composetypes.ServiceVolumeVolume{}
			if v.noCopy.ok() {
				vo.NoCopy = v.noCopy.val
			}
			if v.subPath.ok() {
				vo.Subpath = v.subPath.val
			}
			cv.Volume = vo
		}
		svc.Volumes = append(svc.Volumes, cv)
	}
}

func (s spec) applyComposeNetwork(svc *composetypes.ServiceConfig) {
	if len(s.networks) == 0 {
		return
	}
	svc.Networks = map[string]*composetypes.ServiceNetworkConfig{}
	for _, n := range s.networks {
		cfg := &composetypes.ServiceNetworkConfig{
			Aliases:    append([]string(nil), n.aliases...),
			DriverOpts: composetypes.Options(mapsCopy(n.driverOpts)),
		}
		if n.ipv4.ok() {
			cfg.Ipv4Address = n.ipv4.val
		}
		if n.ipv6.ok() {
			cfg.Ipv6Address = n.ipv6.val
		}
		if n.mac.ok() {
			cfg.MacAddress = n.mac.val
		}
		if n.priority.ok() {
			cfg.Priority = n.priority.val
		}
		if n.gwPriority.ok() {
			cfg.GatewayPriority = n.gwPriority.val
		}
		if n.ifName.ok() {
			cfg.InterfaceName = n.ifName.val
		}
		if len(n.linkLocal) > 0 {
			cfg.LinkLocalIPs = append([]string(nil), n.linkLocal...)
		}
		if n.useAlias.ok() {
			v := n.useAlias.val
			cfg.UseServiceNameAlias = &v
		}
		svc.Networks[n.name.val] = cfg
	}
}

func (s spec) applyComposeDNS(svc *composetypes.ServiceConfig) {
	if len(s.dns) > 0 {
		svc.DNS = composetypes.StringList(append([]string(nil), s.dns...))
	}
	if len(s.dnsOptions) > 0 {
		svc.DNSOpts = append([]string(nil), s.dnsOptions...)
	}
	if len(s.dnsSearch) > 0 {
		svc.DNSSearch = composetypes.StringList(append([]string(nil), s.dnsSearch...))
	}
	if len(s.extraHosts) > 0 {
		hosts := composetypes.HostsList{}
		for _, h := range s.extraHosts {
			hosts[h.hostname] = []string{h.ip}
		}
		svc.ExtraHosts = hosts
	}
}

func (s spec) applyComposeDevices(svc *composetypes.ServiceConfig) {
	for _, d := range s.devices {
		m := composetypes.DeviceMapping{Source: d.host.val}
		if d.container.ok() {
			m.Target = d.container.val
		} else {
			m.Target = d.host.val
		}
		if d.perms.ok() {
			m.Permissions = d.perms.val
		}
		svc.Devices = append(svc.Devices, m)
	}
	if len(s.devCgroup) > 0 {
		svc.DeviceCgroupRules = append([]string(nil), s.devCgroup...)
	}
	for _, g := range s.gpus {
		svc.Gpus = append(svc.Gpus, composeDeviceRequest(g))
	}
}

func (s spec) applyComposeHealth(svc *composetypes.ServiceConfig) error {
	if !s.health.ok() {
		return nil
	}
	h := s.health.val
	hc := &composetypes.HealthCheckConfig{}
	if h.disabled.ok() && h.disabled.val {
		hc.Disable = true
		svc.HealthCheck = hc
		return nil
	}
	if h.test.ok() {
		hc.Test = append([]string(nil), h.test.val...)
	}
	if h.interval.ok() {
		d := composetypes.Duration(h.interval.val)
		hc.Interval = &d
	}
	if h.timeout.ok() {
		d := composetypes.Duration(h.timeout.val)
		hc.Timeout = &d
	}
	if h.startPeriod.ok() {
		d := composetypes.Duration(h.startPeriod.val)
		hc.StartPeriod = &d
	}
	if h.startInterval.ok() {
		d := composetypes.Duration(h.startInterval.val)
		hc.StartInterval = &d
	}
	if h.retries.ok() {
		if h.retries.val < 0 {
			return cfgErr("", "health.retries", "Retries", fmt.Errorf("%w: retries", errInvalid))
		}
		n := uint64(h.retries.val)
		hc.Retries = &n
	}
	svc.HealthCheck = hc
	return nil
}

func applyDepends(dst *composetypes.DependsOnConfig, depends []depend) {
	if len(depends) == 0 {
		return
	}
	if *dst == nil {
		*dst = composetypes.DependsOnConfig{}
	}
	for _, d := range depends {
		dep := composetypes.ServiceDependency{Required: true}
		if d.condition.ok() && d.condition.val != "" {
			dep.Condition = string(d.condition.val)
		} else {
			dep.Condition = string(Started)
		}
		if d.restart.ok() {
			dep.Restart = d.restart.val
		}
		if d.required.ok() {
			dep.Required = d.required.val
		}
		(*dst)[d.name.val] = dep
	}
}

func applyBuild(b build) *composetypes.BuildConfig {
	out := &composetypes.BuildConfig{}
	if b.source.ok() {
		out.Context = b.source.val
	}
	if b.inline.ok() {
		out.DockerfileInline = b.inline.val
	} else if b.dockerfile.ok() {
		out.Dockerfile = b.dockerfile.val
	}
	if b.target.ok() {
		out.Target = b.target.val
	}
	if len(b.args) > 0 {
		out.Args = composetypes.MappingWithEquals{}
		for _, a := range b.args {
			v := a.value.val
			out.Args[a.key] = &v
		}
	}
	if b.labels != nil {
		out.Labels = composetypes.Labels(mapsCopy(b.labels))
	}
	if b.pull.ok() {
		out.Pull = b.pull.val
	}
	if b.noCache.ok() {
		out.NoCache = b.noCache.val
	}
	if b.network.ok() {
		out.Network = b.network.val
	}
	if b.shmSize.ok() {
		out.ShmSize = composetypes.UnitBytes(b.shmSize.val)
	}
	if b.privileged.ok() {
		out.Privileged = b.privileged.val
	}
	if len(b.tags) > 0 {
		out.Tags = composetypes.StringList(append([]string(nil), b.tags...))
	}
	if len(b.platforms) > 0 {
		out.Platforms = composetypes.StringList(append([]string(nil), b.platforms...))
	}
	if b.isolation.ok() {
		out.Isolation = b.isolation.val
	}
	if len(b.extraHosts) > 0 {
		hosts := composetypes.HostsList{}
		for _, h := range b.extraHosts {
			hosts[h.hostname] = []string{h.ip}
		}
		out.ExtraHosts = hosts
	}
	if len(b.entitlements) > 0 {
		out.Entitlements = append([]string(nil), b.entitlements...)
	}
	if b.provenance.ok() {
		out.Provenance = b.provenance.val
	}
	if b.sbom.ok() {
		out.SBOM = b.sbom.val
	}
	for _, k := range b.ssh {
		out.SSH = append(out.SSH, composetypes.SSHKey{ID: k.id, Path: k.path})
	}
	if len(b.cacheFrom) > 0 {
		out.CacheFrom = composetypes.StringList(append([]string(nil), b.cacheFrom...))
	}
	if len(b.cacheTo) > 0 {
		out.CacheTo = composetypes.StringList(append([]string(nil), b.cacheTo...))
	}
	if len(b.noCacheFilter) > 0 {
		out.NoCacheFilter = composetypes.StringList(append([]string(nil), b.noCacheFilter...))
	}
	if len(b.extraContexts) > 0 {
		out.AdditionalContexts = composetypes.Mapping(mapsCopy(b.extraContexts))
	}
	if len(b.secrets) > 0 {
		out.Secrets = applySecrets(b.secrets)
	}
	if len(b.ulimits) > 0 {
		out.Ulimits = map[string]*composetypes.UlimitsConfig{}
		for _, u := range b.ulimits {
			cfg := &composetypes.UlimitsConfig{}
			if u.soft.ok() {
				cfg.Soft = int(u.soft.val)
			}
			if u.hard.ok() {
				cfg.Hard = int(u.hard.val)
			}
			out.Ulimits[u.name.val] = cfg
		}
	}
	return out
}

func applySecrets(refs []secretRef) []composetypes.ServiceSecretConfig {
	out := make([]composetypes.ServiceSecretConfig, 0, len(refs))
	for _, r := range refs {
		sc := composetypes.ServiceSecretConfig{Source: r.source.val}
		if r.target.ok() {
			sc.Target = r.target.val
		}
		if r.uid.ok() {
			sc.UID = r.uid.val
		}
		if r.gid.ok() {
			sc.GID = r.gid.val
		}
		if r.mode.ok() {
			m := composetypes.FileMode(r.mode.val)
			sc.Mode = &m
		}
		out = append(out, sc)
	}
	return out
}

func applyFileConfigs(refs []fileConfigRef) []composetypes.ServiceConfigObjConfig {
	out := make([]composetypes.ServiceConfigObjConfig, 0, len(refs))
	for _, r := range refs {
		sc := composetypes.ServiceConfigObjConfig{Source: r.source.val}
		if r.target.ok() {
			sc.Target = r.target.val
		}
		if r.uid.ok() {
			sc.UID = r.uid.val
		}
		if r.gid.ok() {
			sc.GID = r.gid.val
		}
		if r.mode.ok() {
			m := composetypes.FileMode(r.mode.val)
			sc.Mode = &m
		}
		out = append(out, sc)
	}
	return out
}

func applyServiceExtras(s *Service, svc *composetypes.ServiceConfig) {
	applyDepends(&svc.DependsOn, s.depends)
	if s.build.ok() {
		svc.Build = applyBuild(s.build.val)
	}
	if s.deploy.ok() {
		d := s.deploy.val
		dc := &composetypes.DeployConfig{}
		if d.mode.ok() {
			dc.Mode = d.mode.val
		}
		if d.replicas.ok() {
			n := d.replicas.val
			dc.Replicas = &n
		}
		if d.labels != nil {
			dc.Labels = composetypes.Labels(mapsCopy(d.labels))
		}
		if d.memory.ok() || d.cpus.ok() || d.pids.ok() || len(d.devices) > 0 || len(d.generic) > 0 {
			res := composetypes.Resource{}
			if d.memory.ok() {
				res.MemoryBytes = composetypes.UnitBytes(d.memory.val)
			}
			if d.cpus.ok() {
				res.NanoCPUs = composetypes.NanoCPUs(d.cpus.val)
			}
			if d.pids.ok() {
				res.Pids = d.pids.val
			}
			for _, g := range d.devices {
				res.Devices = append(res.Devices, composeDeviceRequest(g))
			}
			res.GenericResources = composeGeneric(d.generic)
			dc.Resources.Limits = &res
		}
		if d.reserve.ok() {
			r := d.reserve.val
			res := composetypes.Resource{}
			if r.memory.ok() {
				res.MemoryBytes = composetypes.UnitBytes(r.memory.val)
			}
			if r.cpus.ok() {
				res.NanoCPUs = composetypes.NanoCPUs(r.cpus.val)
			}
			if r.pids.ok() {
				res.Pids = r.pids.val
			}
			for _, g := range r.devices {
				res.Devices = append(res.Devices, composeDeviceRequest(g))
			}
			res.GenericResources = composeGeneric(r.generic)
			dc.Resources.Reservations = &res
		}
		if d.update.ok() {
			dc.UpdateConfig = composeUpdate(d.update.val)
		}
		if d.rollback.ok() {
			dc.RollbackConfig = composeUpdate(d.rollback.val)
		}
		if d.restart.ok() {
			dc.RestartPolicy = composeDeployRestart(d.restart.val)
		}
		if d.placement.ok() {
			p := d.placement.val
			pl := composetypes.Placement{
				Constraints: append([]string(nil), p.constraints...),
			}
			for _, s := range p.spreads {
				pl.Preferences = append(pl.Preferences, composetypes.PlacementPreferences{Spread: s})
			}
			if p.maxReplicas.ok() {
				pl.MaxReplicas = p.maxReplicas.val
			}
			dc.Placement = pl
		}
		if d.endpointMode.ok() {
			dc.EndpointMode = d.endpointMode.val
		}
		svc.Deploy = dc
	}
	if s.scale.ok() {
		n := s.scale.val
		svc.Scale = &n
	}
	if len(s.profiles) > 0 {
		svc.Profiles = append([]string(nil), s.profiles...)
	}
	if s.containerName.ok() {
		svc.ContainerName = s.containerName.val
	}
	if s.pullPolicy.ok() {
		svc.PullPolicy = s.pullPolicy.val
	}
	for _, p := range s.envFiles {
		svc.EnvFiles = append(svc.EnvFiles, composeEnvFile(p))
	}
	if len(s.secrets) > 0 {
		svc.Secrets = applySecrets(s.secrets)
	}
	if len(s.configs) > 0 {
		svc.Configs = applyFileConfigs(s.configs)
	}
	if len(s.labelFiles) > 0 {
		svc.LabelFiles = append([]string(nil), s.labelFiles...)
	}
	if s.useAPISocket.ok() {
		svc.UseAPISocket = s.useAPISocket.val
	}
	if s.attach.ok() {
		v := s.attach.val
		svc.Attach = &v
	}
	if s.extends.ok() || s.extendsFile.ok() {
		ex := &composetypes.ExtendsConfig{}
		if s.extends.ok() {
			ex.Service = s.extends.val
		}
		if s.extendsFile.ok() {
			ex.File = s.extendsFile.val
		}
		svc.Extends = ex
	}
	if s.credFile.ok() || s.credRegistry.ok() || s.credConfig.ok() {
		cs := &composetypes.CredentialSpecConfig{}
		if s.credFile.ok() {
			cs.File = s.credFile.val
		}
		if s.credRegistry.ok() {
			cs.Registry = s.credRegistry.val
		}
		if s.credConfig.ok() {
			cs.Config = s.credConfig.val
		}
		svc.CredentialSpec = cs
	}
	if s.provider.ok() || len(s.providerOpts) > 0 {
		svc.Provider = &composetypes.ServiceProviderConfig{
			Type:    s.provider.val,
			Options: composetypes.MultiOptions(s.providerOpts),
		}
	}
	for _, spec := range s.preStart {
		tmp := composetypes.ServiceConfig{}
		if err := spec.applyCompose(&tmp); err != nil {
			continue
		}
		hook := composetypes.PreStartHook{ContainerSpec: tmp.ContainerSpec}
		if spec.perReplica.ok() {
			hook.PerReplica = spec.perReplica.val
		}
		svc.PreStart = append(svc.PreStart, hook)
	}
	for _, h := range s.postStart {
		svc.PostStart = append(svc.PostStart, composeHook(h))
	}
	for _, h := range s.preStop {
		svc.PreStop = append(svc.PreStop, composeHook(h))
	}
	if len(s.externalLinks) > 0 {
		svc.ExternalLinks = append([]string(nil), s.externalLinks...)
	}
	if s.develop.ok() {
		dev := &composetypes.DevelopConfig{}
		for _, w := range s.develop.val.watch {
			t := composetypes.Trigger{Path: w.path.val}
			if w.action.ok() {
				t.Action = composetypes.WatchAction(w.action.val)
			}
			if w.target.ok() {
				t.Target = w.target.val
			}
			t.Include = append([]string(nil), w.include...)
			t.Ignore = append([]string(nil), w.ignore...)
			if w.initialSync.ok() {
				t.InitialSync = w.initialSync.val
			}
			if len(w.exec.cmd) > 0 || w.exec.user.ok() || w.exec.privileged.ok() || w.exec.workdir.ok() || len(w.exec.env) > 0 {
				t.Exec = composeHook(w.exec)
			}
			dev.Watch = append(dev.Watch, t)
		}
		svc.Develop = dev
	}
}

func composeUpdate(u updateCfg) *composetypes.UpdateConfig {
	out := &composetypes.UpdateConfig{}
	if u.parallelism.ok() {
		n := u.parallelism.val
		out.Parallelism = &n
	}
	if u.delay.ok() {
		out.Delay = composetypes.Duration(u.delay.val)
	}
	if u.failureAction.ok() {
		out.FailureAction = u.failureAction.val
	}
	if u.monitor.ok() {
		out.Monitor = composetypes.Duration(u.monitor.val)
	}
	if u.maxFailureRatio.ok() {
		out.MaxFailureRatio = u.maxFailureRatio.val
	}
	if u.order.ok() {
		out.Order = u.order.val
	}
	return out
}

func composeDeployRestart(r restart) *composetypes.RestartPolicy {
	out := &composetypes.RestartPolicy{}
	if r.policy.ok() {
		switch r.policy.val {
		case NoRestart:
			out.Condition = "none"
		case OnFailure:
			out.Condition = "on-failure"
		default:
			out.Condition = "any"
		}
	}
	if r.delay.ok() {
		d := composetypes.Duration(r.delay.val)
		out.Delay = &d
	}
	if r.maxRetries.ok() {
		n := uint64(r.maxRetries.val)
		out.MaxAttempts = &n
	}
	if r.window.ok() {
		d := composetypes.Duration(r.window.val)
		out.Window = &d
	}
	return out
}

func applyJobExtras(j *Job, job *composetypes.JobConfig) {
	applyDepends(&job.DependsOn, j.depends)
	if j.build.ok() {
		job.Build = applyBuild(j.build.val)
	}
	if j.pullPolicy.ok() {
		job.PullPolicy = j.pullPolicy.val
	}
	for _, p := range j.envFiles {
		job.EnvFiles = append(job.EnvFiles, composeEnvFile(p))
	}
	if len(j.secrets) > 0 {
		job.Secrets = applySecrets(j.secrets)
	}
	if len(j.configs) > 0 {
		job.Configs = applyFileConfigs(j.configs)
	}
	if len(j.labelFiles) > 0 {
		job.LabelFiles = append([]string(nil), j.labelFiles...)
	}
	if j.useAPISocket.ok() {
		job.UseAPISocket = j.useAPISocket.val
	}
}

// engineNetworkName is the Docker network name. Compose keys such as
// "default" stay as the project key so services can attach to them, but
// Docker rejects a network whose name is default, bridge, host, or none.
func engineNetworkName(project, key string) string {
	switch key {
	case "default", "bridge", "host", "none":
		if project == "" {
			project = "compose"
		}
		return project + "_" + key
	default:
		return key
	}
}

func projectNetwork(n projectNet) composetypes.NetworkConfig {
	out := composetypes.NetworkConfig{Name: n.name.val}
	if n.driver.ok() {
		out.Driver = n.driver.val
	}
	if n.internal.ok() {
		out.Internal = n.internal.val
	}
	if n.attachable.ok() {
		out.Attachable = n.attachable.val
	}
	if n.external.ok() {
		out.External = composetypes.External(n.external.val)
	}
	if n.ipv4.ok() {
		v := n.ipv4.val
		out.EnableIPv4 = &v
	}
	if n.ipv6.ok() {
		v := n.ipv6.val
		out.EnableIPv6 = &v
	}
	if n.labels != nil {
		out.Labels = composetypes.Labels(mapsCopy(n.labels))
	}
	if n.driverOpts != nil {
		out.DriverOpts = composetypes.Options(mapsCopy(n.driverOpts))
	}
	if n.subnet.ok() || n.gateway.ok() || n.ipRange.ok() || len(n.aux) > 0 {
		pool := &composetypes.IPAMPool{}
		if n.subnet.ok() {
			pool.Subnet = n.subnet.val
		}
		if n.gateway.ok() {
			pool.Gateway = n.gateway.val
		}
		if n.ipRange.ok() {
			pool.IPRange = n.ipRange.val
		}
		if len(n.aux) > 0 {
			pool.AuxiliaryAddresses = composetypes.Mapping(mapsCopy(n.aux))
		}
		out.Ipam.Config = []*composetypes.IPAMPool{pool}
	}
	if n.ipamDriver.ok() {
		out.Ipam.Driver = n.ipamDriver.val
	}
	if len(n.ipamOpts) > 0 {
		out.Ipam.Options = composetypes.Options(mapsCopy(n.ipamOpts))
	}
	return out
}

func projectVolume(v projectVol) composetypes.VolumeConfig {
	out := composetypes.VolumeConfig{Name: v.name.val}
	if v.driver.ok() {
		out.Driver = v.driver.val
	}
	if v.external.ok() {
		out.External = composetypes.External(v.external.val)
	}
	if v.labels != nil {
		out.Labels = composetypes.Labels(mapsCopy(v.labels))
	}
	if v.driverOpts != nil {
		out.DriverOpts = composetypes.Options(mapsCopy(v.driverOpts))
	}
	return out
}

func projectSecretCfg(s projectSecret) composetypes.SecretConfig {
	out := composetypes.SecretConfig{Name: s.name.val}
	if s.file.ok() {
		out.File = s.file.val
	}
	if s.env.ok() {
		out.Environment = s.env.val
	}
	if s.content.ok() {
		out.Content = s.content.val
	}
	if s.external.ok() {
		out.External = composetypes.External(s.external.val)
	}
	if s.driver.ok() {
		out.Driver = s.driver.val
	}
	if s.labels != nil {
		out.Labels = composetypes.Labels(mapsCopy(s.labels))
	}
	if s.templateDriver.ok() {
		out.TemplateDriver = s.templateDriver.val
	}
	return out
}

func projectFileCfg(f projectFile) composetypes.ConfigObjConfig {
	out := composetypes.ConfigObjConfig{Name: f.name.val}
	if f.file.ok() {
		out.File = f.file.val
	}
	if f.env.ok() {
		out.Environment = f.env.val
	}
	if f.content.ok() {
		out.Content = f.content.val
	}
	if f.external.ok() {
		out.External = composetypes.External(f.external.val)
	}
	if f.driver.ok() {
		out.Driver = f.driver.val
	}
	if f.labels != nil {
		out.Labels = composetypes.Labels(mapsCopy(f.labels))
	}
	if f.templateDriver.ok() {
		out.TemplateDriver = f.templateDriver.val
	}
	return out
}

func composeHook(h serviceHook) composetypes.ServiceHook {
	out := composetypes.ServiceHook{Command: append([]string(nil), h.cmd...)}
	if h.user.ok() {
		out.User = h.user.val
	}
	if h.privileged.ok() {
		out.Privileged = h.privileged.val
	}
	if h.workdir.ok() {
		out.WorkingDir = h.workdir.val
	}
	if len(h.env) > 0 {
		env := composetypes.MappingWithEquals{}
		for _, e := range h.env {
			v := e.value.val
			env[e.key] = &v
		}
		out.Environment = env
	}
	return out
}

func composeEnvFile(e envFileRef) composetypes.EnvFile {
	out := composetypes.EnvFile{Path: e.path.val, Required: composetypes.OptOut(true)}
	if e.required.ok() {
		out.Required = composetypes.OptOut(e.required.val)
	}
	if e.format.ok() {
		out.Format = e.format.val
	}
	return out
}

func composeGeneric(items []discreteRes) []composetypes.GenericResource {
	if len(items) == 0 {
		return nil
	}
	out := make([]composetypes.GenericResource, 0, len(items))
	for _, d := range items {
		kind, value := d.kind, d.value
		out = append(out, composetypes.GenericResource{
			DiscreteResourceSpec: &composetypes.DiscreteGenericResource{Kind: kind, Value: value},
		})
	}
	return out
}

func composeDeviceRequest(g gpu) composetypes.DeviceRequest {
	req := composetypes.DeviceRequest{}
	if g.driver.ok() {
		req.Driver = g.driver.val
	}
	if g.count.ok() {
		req.Count = composetypes.DeviceCount(g.count.val)
	} else {
		req.Count = -1
	}
	req.IDs = append([]string(nil), g.ids...)
	if len(g.capabilities) > 0 {
		req.Capabilities = append([]string(nil), g.capabilities...)
	} else {
		req.Capabilities = []string{"gpu"}
	}
	return req
}
