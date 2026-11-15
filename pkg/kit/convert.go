package kit

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"
	"time"

	blkiotypes "github.com/aptd3v/containerkit/internal/engine/blkiodev/types"
	containertypes "github.com/aptd3v/containerkit/internal/engine/container/types"
	mounttypes "github.com/aptd3v/containerkit/internal/engine/mount/types"
	networktypes "github.com/aptd3v/containerkit/internal/engine/network/types"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
)

func (s spec) applyEngine(cfg *containertypes.Config, host *containertypes.HostConfig, netcfg *networktypes.NetworkingConfig) error {
	if s.hostname.ok() {
		cfg.Hostname = s.hostname.val
	}
	if s.domainName.ok() {
		cfg.Domainname = s.domainName.val
	}
	if s.user.ok() {
		cfg.User = s.user.val
	}
	if s.image.ok() {
		cfg.Image = s.image.val
	}
	if s.command.ok() {
		cfg.Cmd = append([]string(nil), s.command.val...)
	}
	if s.entrypoint.ok() {
		cfg.Entrypoint = append([]string(nil), s.entrypoint.val...)
	}
	if s.workdir.ok() {
		cfg.WorkingDir = s.workdir.val
	}
	if s.tty.ok() {
		cfg.Tty = s.tty.val
	}
	if s.interactive.ok() {
		cfg.OpenStdin = s.interactive.val
	}
	if s.stopSignal.ok() {
		cfg.StopSignal = s.stopSignal.val
	}
	if s.networkDisabled.ok() {
		cfg.NetworkDisabled = s.networkDisabled.val
	}
	if len(s.env) > 0 {
		cfg.Env = make([]string, 0, len(s.env))
		for _, e := range s.env {
			cfg.Env = append(cfg.Env, e.key+"="+e.value.val)
		}
	}
	if s.labels != nil {
		cfg.Labels = mapsCopy(s.labels)
	}
	if s.stopTimeout.ok() {
		sec := s.stopTimeout.val
		cfg.StopTimeout = &sec
	}
	if err := s.applyEngineHealth(cfg); err != nil {
		return err
	}
	if err := s.applyEnginePorts(cfg, host); err != nil {
		return err
	}
	s.applyEngineExpose(cfg)
	if s.privileged.ok() {
		host.Privileged = s.privileged.val
	}
	if s.readOnly.ok() {
		host.ReadonlyRootfs = s.readOnly.val
	}
	if s.memory.ok() {
		host.Memory = s.memory.val
	}
	if s.init.ok() {
		v := s.init.val
		host.Init = &v
	}
	if len(s.capAdd) > 0 {
		host.CapAdd = caps(s.capAdd)
	}
	if len(s.capDrop) > 0 {
		host.CapDrop = caps(s.capDrop)
	}
	if s.restart.ok() {
		host.RestartPolicy = containertypes.RestartPolicy{
			Name:              container.RestartPolicyMode(s.restart.val.policy.val),
			MaximumRetryCount: 0,
		}
		if s.restart.val.maxRetries.ok() {
			host.RestartPolicy.MaximumRetryCount = s.restart.val.maxRetries.val
		}
	}
	if err := s.applyEngineMounts(cfg, host); err != nil {
		return err
	}
	if len(s.tmpfs) > 0 {
		if host.Tmpfs == nil {
			host.Tmpfs = map[string]string{}
		}
		for _, p := range s.tmpfs {
			host.Tmpfs[p] = ""
		}
	}
	s.applyEngineResources(host)
	s.applyEngineNamespaces(host)
	if err := s.applyEngineNetwork(host, netcfg); err != nil {
		return err
	}
	if err := s.applyEngineDNS(host); err != nil {
		return err
	}
	s.applyEngineDevices(host)
	if s.logging.ok() {
		host.LogConfig = containertypes.LogConfig{
			Type:   s.logging.val.driver.val,
			Config: mapsCopy(s.logging.val.options),
		}
	}
	return nil
}

func (s spec) applyEngineExpose(cfg *containertypes.Config) {
	if len(s.expose) == 0 {
		return
	}
	if cfg.ExposedPorts == nil {
		cfg.ExposedPorts = networktypes.PortSet{}
	}
	for _, e := range s.expose {
		proto := network.IPProtocol(e.protocol)
		if proto == "" {
			proto = network.TCP
		}
		p, ok := network.PortFrom(e.port, proto)
		if ok {
			cfg.ExposedPorts[p] = struct{}{}
		}
	}
}

func (s spec) applyEngineMounts(cfg *containertypes.Config, host *containertypes.HostConfig) error {
	if len(s.volumesFrom) > 0 {
		host.VolumesFrom = append([]string(nil), s.volumesFrom...)
	}
	if s.volumeDrv.ok() {
		host.VolumeDriver = s.volumeDrv.val
	}
	for _, v := range s.volumes {
		kind := v.kindOrDefault()
		m := mounttypes.Mount{
			Type:     mount.Type(kind),
			ReadOnly: v.readOnly.ok() && v.readOnly.val,
		}
		if v.source.ok() {
			m.Source = v.source.val
		}
		if v.target.ok() {
			m.Target = v.target.val
		}
		if kind == Tmpfs {
			m.TmpfsOptions = &mounttypes.TmpfsOptions{}
			if v.size.ok() {
				m.TmpfsOptions.SizeBytes = v.size.val
			}
			if v.mode.ok() {
				m.TmpfsOptions.Mode = os.FileMode(v.mode.val)
			}
		}
		if kind == Bind && (v.prop.ok() || v.createHostPath.ok() || v.recursive.ok()) {
			bo := &mounttypes.BindOptions{}
			if v.prop.ok() {
				bo.Propagation = mount.Propagation(v.prop.val)
			}
			if v.createHostPath.ok() {
				bo.CreateMountpoint = v.createHostPath.val
			}
			if v.recursive.ok() && (v.recursive.val == "disabled" || v.recursive.val == "false") {
				bo.NonRecursive = true
			}
			m.BindOptions = bo
		}
		if kind == NamedVolume {
			vo := &mounttypes.VolumeOptions{}
			used := false
			if v.noCopy.ok() {
				vo.NoCopy = v.noCopy.val
				used = true
			}
			if v.subPath.ok() {
				vo.Subpath = v.subPath.val
				used = true
			}
			if v.labels != nil {
				vo.Labels = mapsCopy(v.labels)
				used = true
			}
			if v.driver.ok() {
				vo.DriverConfig = &mounttypes.Driver{Name: v.driver.val, Options: mapsCopy(v.DriverValue)}
				used = true
			}
			if used {
				m.VolumeOptions = vo
			}
		}
		if kind == NamedVolume && !v.source.ok() && v.target.ok() {
			if cfg.Volumes == nil {
				cfg.Volumes = map[string]struct{}{}
			}
			cfg.Volumes[v.target.val] = struct{}{}
		}
		host.Mounts = append(host.Mounts, m)
	}
	return nil
}

func (s spec) applyEngineResources(host *containertypes.HostConfig) {
	if s.memReserve.ok() {
		host.MemoryReservation = s.memReserve.val
	}
	if s.memSwap.ok() {
		host.MemorySwap = s.memSwap.val
	}
	if s.memSwapiness.ok() {
		v := s.memSwapiness.val
		host.MemorySwappiness = &v
	}
	if s.shmSize.ok() {
		host.ShmSize = s.shmSize.val
	}
	if s.cpuShares.ok() {
		host.CPUShares = s.cpuShares.val
	}
	if s.cpuPeriod.ok() {
		host.CPUPeriod = s.cpuPeriod.val
	}
	if s.cpuQuota.ok() {
		host.CPUQuota = s.cpuQuota.val
	}
	if s.cpus.ok() {
		host.NanoCPUs = int64(s.cpus.val * 1e9)
	}
	if s.cpusetCpus.ok() {
		host.CpusetCpus = s.cpusetCpus.val
	}
	if s.cpusetMems.ok() {
		host.CpusetMems = s.cpusetMems.val
	}
	if s.cpuCount.ok() {
		host.CPUCount = s.cpuCount.val
	}
	if s.cpuPercent.ok() {
		host.CPUPercent = s.cpuPercent.val
	}
	if s.cpuRTPeriod.ok() {
		host.CPURealtimePeriod = s.cpuRTPeriod.val
	}
	if s.cpuRTRuntime.ok() {
		host.CPURealtimeRuntime = s.cpuRTRuntime.val
	}
	if s.pidsLimit.ok() {
		v := s.pidsLimit.val
		host.PidsLimit = &v
	}
	if s.oomScoreAdj.ok() {
		host.OomScoreAdj = s.oomScoreAdj.val
	}
	if s.oomKillDisable.ok() {
		v := s.oomKillDisable.val
		host.OomKillDisable = &v
	}
	if s.blkioWeight.ok() {
		host.BlkioWeight = s.blkioWeight.val
	}
	if s.cgroupParent.ok() {
		host.CgroupParent = s.cgroupParent.val
	}
	if len(s.ulimits) > 0 {
		for _, u := range s.ulimits {
			ul := &containertypes.Ulimit{Name: u.name.val}
			if u.soft.ok() {
				ul.Soft = u.soft.val
			}
			if u.hard.ok() {
				ul.Hard = u.hard.val
			}
			host.Ulimits = append(host.Ulimits, ul)
		}
	}
	if s.sysctls != nil {
		host.Sysctls = mapsCopy(s.sysctls)
	}
	if len(s.securityOpt) > 0 {
		host.SecurityOpt = append([]string(nil), s.securityOpt...)
	}
	if s.storageOpt != nil {
		host.StorageOpt = mapsCopy(s.storageOpt)
	}
	if len(s.groupAdd) > 0 {
		host.GroupAdd = append([]string(nil), s.groupAdd...)
	}
	if s.annotations != nil {
		host.Annotations = mapsCopy(s.annotations)
	}
	if s.umask.ok() {
		v := s.umask.val
		host.Umask = &v
	}
	if s.ioMaxIOps.ok() {
		host.IOMaximumIOps = s.ioMaxIOps.val
	}
	if s.ioMaxBandwidth.ok() {
		host.IOMaximumBandwidth = s.ioMaxBandwidth.val
	}
	if s.consoleH.ok() || s.consoleW.ok() {
		var h, w uint
		if s.consoleH.ok() {
			h = s.consoleH.val
		}
		if s.consoleW.ok() {
			w = s.consoleW.val
		}
		host.ConsoleSize = [2]uint{h, w}
	}
	s.applyEngineBlkio(host)
}

func (s spec) applyEngineNamespaces(host *containertypes.HostConfig) {
	if s.ipc.ok() {
		host.IpcMode = containertypes.IpcMode(s.ipc.val)
	}
	if s.pid.ok() {
		host.PidMode = containertypes.PidMode(s.pid.val)
	}
	if s.uts.ok() {
		host.UTSMode = containertypes.UTSMode(s.uts.val)
	}
	if s.userns.ok() {
		host.UsernsMode = containertypes.UsernsMode(s.userns.val)
	}
	if s.cgroup.ok() {
		host.Cgroup = containertypes.CgroupSpec(s.cgroup.val)
	}
	if s.cgroupns.ok() {
		host.CgroupnsMode = containertypes.CgroupnsMode(s.cgroupns.val)
	}
	if s.isolation.ok() {
		host.Isolation = containertypes.Isolation(s.isolation.val)
	}
	if s.runtime.ok() {
		host.Runtime = s.runtime.val
	}
	if len(s.links) > 0 {
		host.Links = append([]string(nil), s.links...)
	}
}

func (s spec) applyEngineNetwork(host *containertypes.HostConfig, netcfg *networktypes.NetworkingConfig) error {
	if s.networkMode.ok() {
		host.NetworkMode = containertypes.NetworkMode(s.networkMode.val)
	}
	if s.networkDisabled.ok() && s.networkDisabled.val && !s.networkMode.ok() {
		host.NetworkMode = "none"
	}
	if len(s.networks) == 0 {
		return nil
	}
	if netcfg.EndpointsConfig == nil {
		netcfg.EndpointsConfig = map[string]*networktypes.EndpointSettings{}
	}
	for _, n := range s.networks {
		ep := &networktypes.EndpointSettings{
			Aliases:    append([]string(nil), n.aliases...),
			DriverOpts: mapsCopy(n.driverOpts),
		}
		if n.gwPriority.ok() {
			ep.GwPriority = n.gwPriority.val
		} else if n.priority.ok() {
			ep.GwPriority = n.priority.val
		}
		ipam := &networktypes.EndpointIPAMConfig{}
		used := false
		if n.ipv4.ok() {
			addr, err := netip.ParseAddr(n.ipv4.val)
			if err != nil {
				return cfgErr("", "network.ipv4", "IPv4", fmt.Errorf("%w: ipv4 %q", errInvalid, n.ipv4.val))
			}
			ipam.IPv4Address = addr
			used = true
		}
		if n.ipv6.ok() {
			addr, err := netip.ParseAddr(n.ipv6.val)
			if err != nil {
				return cfgErr("", "network.ipv6", "IPv6", fmt.Errorf("%w: ipv6 %q", errInvalid, n.ipv6.val))
			}
			ipam.IPv6Address = addr
			used = true
		}
		for _, ip := range n.linkLocal {
			addr, err := netip.ParseAddr(ip)
			if err != nil {
				return cfgErr("", "network.link_local_ips", "LinkLocal", fmt.Errorf("%w: link local ip %q", errInvalid, ip))
			}
			ipam.LinkLocalIPs = append(ipam.LinkLocalIPs, addr)
			used = true
		}
		if used {
			ep.IPAMConfig = ipam
		}
		if n.mac.ok() {
			hw, err := net.ParseMAC(n.mac.val)
			if err != nil {
				return cfgErr("", "network.mac_address", "MacAddress", fmt.Errorf("%w: mac %q", errInvalid, n.mac.val))
			}
			ep.MacAddress = networktypes.HardwareAddr(hw)
		}
		netcfg.EndpointsConfig[n.name.val] = ep
	}
	return nil
}

func (s spec) applyEngineDNS(host *containertypes.HostConfig) error {
	for _, d := range s.dns {
		addr, err := netip.ParseAddr(d)
		if err != nil {
			return cfgErr("", "dns", "DNS", fmt.Errorf("%w: dns %q", errInvalid, d))
		}
		host.DNS = append(host.DNS, addr)
	}
	if len(s.dnsOptions) > 0 {
		host.DNSOptions = append([]string(nil), s.dnsOptions...)
	}
	if len(s.dnsSearch) > 0 {
		host.DNSSearch = append([]string(nil), s.dnsSearch...)
	}
	for _, h := range s.extraHosts {
		host.ExtraHosts = append(host.ExtraHosts, h.hostname+":"+h.ip)
	}
	return nil
}

func (s spec) applyEngineBlkio(host *containertypes.HostConfig) {
	for _, d := range s.blkioWDev {
		host.BlkioWeightDevice = append(host.BlkioWeightDevice, &blkiotypes.WeightDevice{Path: d.path, Weight: d.weight})
	}
	for _, d := range s.blkioReadBps {
		host.BlkioDeviceReadBps = append(host.BlkioDeviceReadBps, &blkiotypes.ThrottleDevice{Path: d.path, Rate: d.rate})
	}
	for _, d := range s.blkioWriteBps {
		host.BlkioDeviceWriteBps = append(host.BlkioDeviceWriteBps, &blkiotypes.ThrottleDevice{Path: d.path, Rate: d.rate})
	}
	for _, d := range s.blkioReadIOps {
		host.BlkioDeviceReadIOps = append(host.BlkioDeviceReadIOps, &blkiotypes.ThrottleDevice{Path: d.path, Rate: d.rate})
	}
	for _, d := range s.blkioWriteIOps {
		host.BlkioDeviceWriteIOps = append(host.BlkioDeviceWriteIOps, &blkiotypes.ThrottleDevice{Path: d.path, Rate: d.rate})
	}
}

func (s spec) applyEngineDevices(host *containertypes.HostConfig) {
	for _, d := range s.devices {
		m := containertypes.DeviceMapping{PathOnHost: d.host.val, CgroupPermissions: "rwm"}
		if d.container.ok() {
			m.PathInContainer = d.container.val
		} else {
			m.PathInContainer = d.host.val
		}
		if d.perms.ok() {
			m.CgroupPermissions = d.perms.val
		}
		host.Devices = append(host.Devices, m)
	}
	if len(s.devCgroup) > 0 {
		host.DeviceCgroupRules = append([]string(nil), s.devCgroup...)
	}
	for _, g := range s.gpus {
		req := containertypes.DeviceRequest{}
		if g.driver.ok() {
			req.Driver = g.driver.val
		}
		if g.count.ok() {
			req.Count = g.count.val
		} else {
			req.Count = -1
		}
		req.DeviceIDs = append([]string(nil), g.ids...)
		if len(g.capabilities) > 0 {
			req.Capabilities = [][]string{append([]string(nil), g.capabilities...)}
		} else {
			req.Capabilities = [][]string{{"gpu"}}
		}
		host.DeviceRequests = append(host.DeviceRequests, req)
	}
}

func (s spec) applyEngineHealth(cfg *containertypes.Config) error {
	if !s.health.ok() {
		return nil
	}
	h := s.health.val
	hc := &containertypes.HealthConfig{}
	if h.disabled.ok() && h.disabled.val {
		hc.Test = []string{"NONE"}
		cfg.Healthcheck = hc
		return nil
	}
	if h.test.ok() {
		hc.Test = append([]string(nil), h.test.val...)
	}
	if h.interval.ok() {
		if err := checkDuration("health.interval", "Interval", h.interval.val); err != nil {
			return err
		}
		hc.Interval = h.interval.val
	}
	if h.timeout.ok() {
		if err := checkDuration("health.timeout", "Timeout", h.timeout.val); err != nil {
			return err
		}
		hc.Timeout = h.timeout.val
	}
	if h.startPeriod.ok() {
		if err := checkDuration("health.start_period", "StartPeriod", h.startPeriod.val); err != nil {
			return err
		}
		hc.StartPeriod = h.startPeriod.val
	}
	if h.retries.ok() {
		hc.Retries = h.retries.val
	}
	cfg.Healthcheck = hc
	return nil
}

func (s spec) applyEnginePorts(cfg *containertypes.Config, host *containertypes.HostConfig) error {
	if len(s.ports) == 0 {
		return nil
	}
	if cfg.ExposedPorts == nil {
		cfg.ExposedPorts = networktypes.PortSet{}
	}
	if host.PortBindings == nil {
		host.PortBindings = networktypes.PortMap{}
	}
	for _, p := range s.ports {
		proto := network.IPProtocol(p.proto())
		ep, ok := network.PortFrom(p.containerPort.val, proto)
		if !ok {
			return cfgErr("", "publish.container_port", "ContainerPort", fmt.Errorf("%w: port protocol", errInvalid))
		}
		cfg.ExposedPorts[ep] = struct{}{}
		var addr netip.Addr
		if ip := p.hostIPOrDefault(); ip != "" {
			parsed, err := netip.ParseAddr(ip)
			if err != nil {
				return cfgErr("", "publish.host_ip", "HostIP", fmt.Errorf("%w: host ip %q", errInvalid, ip))
			}
			addr = parsed
		}
		host.PortBindings[ep] = []networktypes.PortBinding{{
			HostIP:   addr,
			HostPort: p.hostPortString(),
		}}
	}
	return nil
}

func checkDuration(path, setting string, d time.Duration) error {
	if d != 0 && d < container.MinimumDuration {
		return cfgErr("", path, setting, fmt.Errorf("%w: duration %s below engine minimum", errInvalid, d))
	}
	return nil
}

func caps(in []Capability) []string {
	out := make([]string, len(in))
	for i, c := range in {
		out[i] = string(c)
	}
	return out
}

func mapsCopy[M ~map[string]string](in M) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func portProtoString(p Proto) string {
	if p == "" {
		return string(TCP)
	}
	return string(p)
}

func formatExpose(e exposePort) string {
	return strconv.FormatUint(uint64(e.port), 10) + "/" + portProtoString(e.protocol)
}
