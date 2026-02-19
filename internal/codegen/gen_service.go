package codegen

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/dave/jennifer/jen"
)

// serviceFuncName returns an exported Go name for a service (e.g. "api" -> "Api", "my-service" -> "MyService").
func serviceFuncName(serviceName string) string {
	parts := strings.Split(serviceName, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// genContainerExpr returns the container expression (second argument to WithService):
// create.NewContainer(...).WithContainerConfig(...).WithHostConfig(...).WithNetworkConfig(...).WithPlatformConfig(...).
func genContainerExpr(name string, svc *types.ServiceConfig) *jen.Statement {
	containerName := name
	if svc.ContainerName != "" {
		containerName = svc.ContainerName
	}
	ccParts := genContainerConfig(svc)
	hcParts := genHostConfig(svc)
	ncParts := genNetworkConfig(svc)
	pcParts := genPlatformConfig(svc)

	container := jen.Qual(pkgCreate, "NewContainer").Call(jen.Lit(containerName))
	if len(ccParts) > 0 {
		container = container.Dot("WithContainerConfig").Call(ccParts...)
	}
	if len(hcParts) > 0 {
		container = container.Dot("WithHostConfig").Call(hcParts...)
	}
	if len(ncParts) > 0 {
		container = container.Dot("WithNetworkConfig").Call(ncParts...)
	}
	if len(pcParts) > 0 {
		container = container.Dot("WithPlatformConfig").Call(pcParts...)
	}
	return container
}

// genServiceFunc returns the container function, optional service-config function, and the WithService call.
// Container function: With<Name>Container() *create.Container.
// Service-config function (if any): With<Name>ServiceConfig() create.SetServiceConfig { return tools.Group(...) }.
// Call: project.WithService(name, WithXxxContainer(), WithXxxServiceConfig()) or project.WithService(name, WithXxxContainer()).
func genServiceFunc(name string, svc *types.ServiceConfig) (containerFuncDef *jen.Statement, serviceConfigFuncDef *jen.Statement, callStmt *jen.Statement) {
	containerFnName := "With" + serviceFuncName(name) + "Container"
	containerExpr := genContainerExpr(name, svc)
	containerFuncDef = jen.Func().Id(containerFnName).Params().
		Op("*").Qual(pkgCreate, "Container").
		Block(jen.Return(containerExpr))

	scParts := genServiceLevelConfig(svc)
	callArgs := []jen.Code{jen.Lit(name), jen.Id(containerFnName).Call()}
	if len(scParts) > 0 {
		configFnName := "With" + serviceFuncName(name) + "ServiceConfig"
		serviceConfigFuncDef = jen.Func().Id(configFnName).Params().
			Qual(pkgCreate, "SetServiceConfig").
			Block(jen.Return(jen.Qual(pkgTools, "Group").Call(scParts...)))
		callArgs = append(callArgs, jen.Id(configFnName).Call())
	}
	callStmt = jen.Id("project").Dot("WithService").Call(callArgs...)
	return containerFuncDef, serviceConfigFuncDef, callStmt
}

func genContainerConfig(svc *types.ServiceConfig) []jen.Code {
	var parts []jen.Code
	if svc.Image != "" {
		parts = append(parts, jen.Qual(pkgCC, "WithImage").Call(jen.Lit(svc.Image)))
	}
	for _, e := range svc.Expose {
		portStr := strings.TrimSpace(e)
		if portStr == "" {
			continue
		}
		protocol := "tcp"
		port := portStr
		if idx := strings.Index(portStr, "/"); idx >= 0 {
			protocol = strings.TrimSpace(portStr[idx+1:])
			port = strings.TrimSpace(portStr[:idx])
		}
		if port != "" {
			parts = append(parts, jen.Qual(pkgCC, "WithExposedPort").Call(jen.Lit(protocol), jen.Lit(port)))
		}
	}
	if svc.StopGracePeriod != nil {
		secs := int(time.Duration(*svc.StopGracePeriod).Seconds())
		if secs > 0 {
			parts = append(parts, jen.Qual(pkgCC, "WithStopTimeout").Call(jen.Lit(secs)))
		}
	}
	for k, v := range svc.Environment {
		val := ""
		if v != nil {
			val = *v
		}
		parts = append(parts, jen.Qual(pkgCC, "WithEnv").Call(jen.Lit(k), jen.Lit(val)))
	}
	if len(svc.Command) > 0 {
		args := make([]jen.Code, len(svc.Command))
		for i, c := range svc.Command {
			args[i] = jen.Lit(c)
		}
		parts = append(parts, jen.Qual(pkgCC, "WithCommand").Call(args...))
	}
	if len(svc.Entrypoint) > 0 {
		args := make([]jen.Code, len(svc.Entrypoint))
		for i, e := range svc.Entrypoint {
			args[i] = jen.Lit(e)
		}
		parts = append(parts, jen.Qual(pkgCC, "WithEntrypoint").Call(args...))
	}
	if svc.User != "" {
		parts = append(parts, jen.Qual(pkgCC, "WithUser").Call(jen.Lit(svc.User)))
	}
	if svc.WorkingDir != "" {
		parts = append(parts, jen.Qual(pkgCC, "WithWorkingDir").Call(jen.Lit(svc.WorkingDir)))
	}
	if svc.Hostname != "" {
		parts = append(parts, jen.Qual(pkgCC, "WithHostName").Call(jen.Lit(svc.Hostname)))
	}
	if svc.DomainName != "" {
		parts = append(parts, jen.Qual(pkgCC, "WithDomainName").Call(jen.Lit(svc.DomainName)))
	}
	if svc.StopSignal != "" {
		parts = append(parts, jen.Qual(pkgCC, "WithStopSignal").Call(jen.Lit(svc.StopSignal)))
	}
	for k, v := range svc.Labels {
		parts = append(parts, jen.Qual(pkgCC, "WithLabel").Call(jen.Lit(k), jen.Lit(v)))
	}
	if svc.HealthCheck != nil && !svc.HealthCheck.Disable {
		hcParts := genHealthCheck(svc.HealthCheck)
		if len(hcParts) > 0 {
			parts = append(parts, jen.Qual(pkgCC, "WithHealthCheck").Call(hcParts...))
		}
	}
	if svc.Tty {
		parts = append(parts, jen.Qual(pkgCC, "WithTty").Call())
	}
	if svc.StdinOpen {
		parts = append(parts, jen.Qual(pkgCC, "WithStdinOpen").Call())
	}
	return parts
}

func genHealthCheck(hc *types.HealthCheckConfig) []jen.Code {
	var parts []jen.Code
	if len(hc.Test) > 0 {
		args := make([]jen.Code, len(hc.Test))
		for i, t := range hc.Test {
			args[i] = jen.Lit(t)
		}
		parts = append(parts, jen.Qual(pkgHealth, "WithTest").Call(args...))
	}
	if hc.Interval != nil {
		parts = append(parts, jen.Qual(pkgHealth, "WithInterval").Call(jen.Lit(hc.Interval.String())))
	}
	if hc.Timeout != nil {
		parts = append(parts, jen.Qual(pkgHealth, "WithTimeout").Call(jen.Lit(hc.Timeout.String())))
	}
	if hc.StartPeriod != nil {
		parts = append(parts, jen.Qual(pkgHealth, "WithStartPeriod").Call(jen.Lit(hc.StartPeriod.String())))
	}
	if hc.Retries != nil {
		parts = append(parts, jen.Qual(pkgHealth, "WithRetries").Call(jen.Lit(int(*hc.Retries))))
	}
	return parts
}

func genHostConfig(svc *types.ServiceConfig) []jen.Code {
	var parts []jen.Code
	for _, p := range svc.Ports {
		protocol := p.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		hostIP := p.HostIP
		if hostIP == "" {
			hostIP = "0.0.0.0"
		}
		published := p.Published
		if published == "" {
			published = strconv.Itoa(int(p.Target))
		}
		target := strconv.Itoa(int(p.Target))
		parts = append(parts, jen.Qual(pkgHC, "WithPortBindings").Call(jen.Lit(protocol), jen.Lit(hostIP), jen.Lit(published), jen.Lit(target)))
	}
	for _, v := range svc.Volumes {
		if stmt := genVolumeMount(v); stmt != nil {
			parts = append(parts, stmt)
		}
	}
	restartMaxRetry := 0
	if strings.HasPrefix(svc.Restart, "on-failure") {
		if idx := strings.Index(svc.Restart, ":"); idx >= 0 && idx+1 < len(svc.Restart) {
			if n, err := strconv.Atoi(strings.TrimSpace(svc.Restart[idx+1:])); err == nil && n > 0 {
				restartMaxRetry = n
			}
		}
	}
	switch {
	case svc.Restart == "always":
		parts = append(parts, jen.Qual(pkgHC, "WithRestartPolicyAlways").Call())
	case svc.Restart == "unless-stopped":
		parts = append(parts, jen.Qual(pkgHC, "WithRestartPolicyUnlessStopped").Call())
	case svc.Restart == "on-failure" || strings.HasPrefix(svc.Restart, "on-failure:"):
		parts = append(parts, jen.Qual(pkgHC, "WithRestartPolicyOnFailure").Call(jen.Lit(restartMaxRetry)))
	case svc.Restart == "no" || svc.Restart == "":
		parts = append(parts, jen.Qual(pkgHC, "WithRestartPolicyNever").Call())
	}
	if svc.Privileged {
		parts = append(parts, jen.Qual(pkgHC, "WithPrivileged").Call())
	}
	if svc.ReadOnly {
		parts = append(parts, jen.Qual(pkgHC, "WithReadOnlyRootfs").Call())
	}
	if svc.MemLimit > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithMemoryLimit").Call(jen.Lit(int(svc.MemLimit))))
	}
	if svc.ShmSize > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithShmSize").Call(jen.Lit(int(svc.ShmSize))))
	}
	if len(svc.DNS) > 0 {
		args := make([]jen.Code, len(svc.DNS))
		for i, d := range svc.DNS {
			args[i] = jen.Lit(d)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithDNSLookups").Call(args...))
	}
	if len(svc.ExtraHosts) > 0 {
		extraList := svc.ExtraHosts.AsList(":")
		args := make([]jen.Code, len(extraList))
		for i, h := range extraList {
			args[i] = jen.Lit(h)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithExtraHost").Call(args...))
	}
	if svc.Init != nil && *svc.Init {
		parts = append(parts, jen.Qual(pkgHC, "WithInit").Call())
	}
	// cgroup, cpu, memory, dns, devices, logging, security, ulimits, etc.
	if svc.CgroupParent != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithCgroupParent").Call(jen.Lit(svc.CgroupParent)))
	}
	if svc.Cgroup != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithCgroup").Call(jen.Lit(svc.Cgroup)))
	}
	if svc.CPUCount > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithCPUCount").Call(jen.Lit(svc.CPUCount)))
	}
	if svc.CPUPercent > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithCPUPercent").Call(jen.Lit(int(svc.CPUPercent))))
	}
	if svc.CPUPeriod > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithCPUPeriod").Call(jen.Lit(svc.CPUPeriod)))
	}
	if svc.CPUQuota > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithCPUQuota").Call(jen.Lit(svc.CPUQuota)))
	}
	if svc.CPUShares > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithCPUShares").Call(jen.Lit(svc.CPUShares)))
	}
	if svc.CPUSet != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithCpusetCpus").Call(jen.Lit(svc.CPUSet)))
	}
	if len(svc.DeviceCgroupRules) > 0 {
		args := make([]jen.Code, len(svc.DeviceCgroupRules))
		for i, r := range svc.DeviceCgroupRules {
			args[i] = jen.Lit(r)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithDeviceCgroupRules").Call(args...))
	}
	for _, d := range svc.Devices {
		perm := d.Permissions
		if perm == "" {
			perm = "rwm"
		}
		parts = append(parts, jen.Qual(pkgHC, "WithAddedDevice").Call(jen.Lit(d.Source), jen.Lit(d.Target), jen.Lit(perm)))
	}
	if len(svc.DNSOpts) > 0 {
		args := make([]jen.Code, len(svc.DNSOpts))
		for i, o := range svc.DNSOpts {
			args[i] = jen.Lit(o)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithDNSOptions").Call(args...))
	}
	if len(svc.DNSSearch) > 0 {
		args := make([]jen.Code, len(svc.DNSSearch))
		for i, s := range svc.DNSSearch {
			args[i] = jen.Lit(s)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithDNSSearches").Call(args...))
	}
	if len(svc.GroupAdd) > 0 {
		args := make([]jen.Code, len(svc.GroupAdd))
		for i, g := range svc.GroupAdd {
			args[i] = jen.Lit(g)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithAddedGroups").Call(args...))
	}
	if svc.Ipc != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithIpcMode").Call(jen.Lit(svc.Ipc)))
	}
	if svc.Isolation != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithIsolation").Call(jen.Lit(svc.Isolation)))
	}
	if svc.LogDriver != "" {
		logOpts := svc.LogOpt
		if logOpts == nil {
			logOpts = make(map[string]string)
		}
		d := jen.Dict{}
		for k, v := range logOpts {
			d[jen.Lit(k)] = jen.Lit(v)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithLogDriver").Call(jen.Lit(svc.LogDriver), jen.Map(jen.String()).String().Values(d)))
	}
	if svc.MemReservation > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithMemoryReservation").Call(jen.Lit(int(svc.MemReservation))))
	}
	if svc.MemSwapLimit > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithMemorySwap").Call(jen.Lit(int(svc.MemSwapLimit))))
	}
	if svc.NetworkMode != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithNetworkMode").Call(jen.Lit(svc.NetworkMode)))
	}
	if svc.OomScoreAdj != 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithOomScoreAdj").Call(jen.Lit(int(svc.OomScoreAdj))))
	}
	if svc.OomKillDisable {
		parts = append(parts, jen.Qual(pkgHC, "WithOomKillDisable").Call())
	}
	if svc.Pid != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithPidMode").Call(jen.Lit(svc.Pid)))
	}
	if svc.PidsLimit > 0 {
		parts = append(parts, jen.Qual(pkgHC, "WithPidsLimit").Call(jen.Lit(svc.PidsLimit)))
	}
	if svc.Runtime != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithRuntime").Call(jen.Lit(svc.Runtime)))
	}
	if len(svc.SecurityOpt) > 0 {
		args := make([]jen.Code, len(svc.SecurityOpt))
		for i, o := range svc.SecurityOpt {
			args[i] = jen.Lit(o)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithSecurityOpts").Call(args...))
	}
	for k, v := range svc.StorageOpt {
		parts = append(parts, jen.Qual(pkgHC, "WithStorageOpt").Call(jen.Lit(k), jen.Lit(v)))
	}
	for k, v := range svc.Sysctls {
		parts = append(parts, jen.Qual(pkgHC, "WithSysctls").Call(jen.Lit(k), jen.Lit(v)))
	}
	for name, u := range svc.Ulimits {
		if u == nil {
			continue
		}
		soft, hard := int(u.Soft), int(u.Hard)
		if u.Single != 0 {
			soft, hard = int(u.Single), int(u.Single)
		}
		parts = append(parts, jen.Qual(pkgHC, "WithUlimits").Call(jen.Lit(name), jen.Lit(int(soft)), jen.Lit(int(hard))))
	}
	if svc.UserNSMode != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithUserNSMode").Call(jen.Lit(svc.UserNSMode)))
	}
	if svc.Uts != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithUTSMode").Call(jen.Lit(svc.Uts)))
	}
	for _, from := range svc.VolumesFrom {
		parts = append(parts, jen.Qual(pkgHC, "WithVolumesFrom").Call(jen.Lit(from)))
	}
	if svc.VolumeDriver != "" {
		parts = append(parts, jen.Qual(pkgHC, "WithVolumeDriver").Call(jen.Lit(svc.VolumeDriver)))
	}
	// cap_add / cap_drop via create API
	if len(svc.CapAdd) > 0 {
		args := make([]jen.Code, len(svc.CapAdd))
		for i, c := range svc.CapAdd {
			args[i] = jen.Qual(pkgHC, "Capability").Parens(jen.Lit(c))
		}
		parts = append(parts, jen.Qual(pkgHC, "WithAddedCapabilities").Call(args...))
	}
	if len(svc.CapDrop) > 0 {
		args := make([]jen.Code, len(svc.CapDrop))
		for i, c := range svc.CapDrop {
			args[i] = jen.Qual(pkgHC, "Capability").Parens(jen.Lit(c))
		}
		parts = append(parts, jen.Qual(pkgHC, "WithDroppedCapabilities").Call(args...))
	}
	// blkio
	if svc.BlkioConfig != nil {
		if svc.BlkioConfig.Weight > 0 {
			parts = append(parts, jen.Qual(pkgHC, "WithBlkioWeight").Call(jen.Lit(svc.BlkioConfig.Weight)))
		}
		for _, td := range svc.BlkioConfig.DeviceReadBps {
			parts = append(parts, jen.Qual(pkgHC, "WithBlkioDeviceReadBps").Call(jen.Lit(td.Path), jen.Lit(int(td.Rate))))
		}
		for _, td := range svc.BlkioConfig.DeviceWriteBps {
			parts = append(parts, jen.Qual(pkgHC, "WithBlkioDeviceWriteBps").Call(jen.Lit(td.Path), jen.Lit(int(td.Rate))))
		}
		for _, td := range svc.BlkioConfig.DeviceReadIOps {
			parts = append(parts, jen.Qual(pkgHC, "WithBlkioDeviceReadIOps").Call(jen.Lit(td.Path), jen.Lit(int(td.Rate))))
		}
		for _, td := range svc.BlkioConfig.DeviceWriteIOps {
			parts = append(parts, jen.Qual(pkgHC, "WithBlkioDeviceWriteIOps").Call(jen.Lit(td.Path), jen.Lit(int(td.Rate))))
		}
	}
	return parts
}

func genVolumeMount(v types.ServiceVolumeConfig) *jen.Statement {
	source := v.Source
	target := v.Target
	readOnly := v.ReadOnly
	switch v.Type {
	case "bind", "":
		if source == "" {
			return nil // TODO: volume missing source - skip or comment
		}
		if readOnly {
			return jen.Qual(pkgHC, "WithROHostBindMount").Call(jen.Lit(source), jen.Lit(target))
		}
		return jen.Qual(pkgHC, "WithRWHostBindMount").Call(jen.Lit(source), jen.Lit(target))
	case "volume":
		if source == "" {
			source = "anonymous"
		}
		if readOnly {
			return jen.Qual(pkgHC, "WithRONamedVolumeMount").Call(jen.Lit(source), jen.Lit(target))
		}
		return jen.Qual(pkgHC, "WithRWNamedVolumeMount").Call(jen.Lit(source), jen.Lit(target))
	case "tmpfs":
		var size int
		mode := 0o777
		if v.Tmpfs != nil {
			size = int(v.Tmpfs.Size)
			if v.Tmpfs.Mode != 0 {
				mode = int(v.Tmpfs.Mode)
			}
		}
		return jen.Qual(pkgHC, "WithTmpfsMount").Call(jen.Lit(target), jen.Lit(size), jen.Lit(mode))
	default:
		return nil // TODO: volume type
	}
}

func genNetworkConfig(svc *types.ServiceConfig) []jen.Code {
	var parts []jen.Code
	netNames := make([]string, 0, len(svc.Networks))
	for name := range svc.Networks {
		netNames = append(netNames, name)
	}
	sort.Strings(netNames)
	for i, netName := range netNames {
		netCfg := svc.Networks[netName]
		args := []jen.Code{jen.Lit(netName)}
		if netCfg != nil {
			if len(netCfg.Aliases) > 0 {
				aliasArgs := make([]jen.Code, len(netCfg.Aliases))
				for j, a := range netCfg.Aliases {
					aliasArgs[j] = jen.Lit(a)
				}
				args = append(args, jen.Qual(pkgEndpoint, "WithAliases").Call(aliasArgs...))
			}
			var ipamParts []jen.Code
			if netCfg.Ipv4Address != "" {
				ipamParts = append(ipamParts, jen.Qual(pkgEndpointIPAM, "WithIPv4Address").Call(jen.Lit(netCfg.Ipv4Address)))
			}
			if netCfg.Ipv6Address != "" {
				ipamParts = append(ipamParts, jen.Qual(pkgEndpointIPAM, "WithIPv6Address").Call(jen.Lit(netCfg.Ipv6Address)))
			}
			if len(netCfg.LinkLocalIPs) > 0 {
				ipArgs := make([]jen.Code, len(netCfg.LinkLocalIPs))
				for j, ip := range netCfg.LinkLocalIPs {
					ipArgs[j] = jen.Lit(ip)
				}
				ipamParts = append(ipamParts, jen.Qual(pkgEndpointIPAM, "WithLinkLocalIPs").Call(ipArgs...))
			}
			if len(ipamParts) > 0 {
				args = append(args, jen.Qual(pkgEndpoint, "WithIPAMConfig").Call(ipamParts...))
			}
			macAddr := netCfg.MacAddress
			if macAddr == "" && i == 0 && svc.MacAddress != "" {
				macAddr = svc.MacAddress
			}
			if macAddr != "" {
				args = append(args, jen.Qual(pkgEndpoint, "WithMacAddress").Call(jen.Lit(macAddr)))
			}
			for k, v := range netCfg.DriverOpts {
				args = append(args, jen.Qual(pkgEndpoint, "WithDriverOptions").Call(jen.Lit(k), jen.Lit(v)))
			}
		} else if i == 0 && svc.MacAddress != "" {
			// Service-level mac_address (Docker API v1.44+): set on first endpoint instead of deprecated cc.WithMacAddress
			args = append(args, jen.Qual(pkgEndpoint, "WithMacAddress").Call(jen.Lit(svc.MacAddress)))
		}
		parts = append(parts, jen.Qual(pkgNC, "WithEndpoint").Call(args...))
	}
	return parts
}

func genPlatformConfig(svc *types.ServiceConfig) []jen.Code {
	var parts []jen.Code
	if svc.Platform == "" {
		return parts
	}
	// Platform may be "arch", "os/arch", or "os/arch/variant"
	segments := strings.Split(svc.Platform, "/")
	switch len(segments) {
	case 1:
		parts = append(parts, jen.Qual(pkgPC, "WithArchitecture").Call(jen.Lit(segments[0])))
	case 2:
		parts = append(parts, jen.Qual(pkgPC, "WithOS").Call(jen.Lit(segments[0])))
		parts = append(parts, jen.Qual(pkgPC, "WithArchitecture").Call(jen.Lit(segments[1])))
	case 3:
		parts = append(parts, jen.Qual(pkgPC, "WithOS").Call(jen.Lit(segments[0])))
		parts = append(parts, jen.Qual(pkgPC, "WithArchitecture").Call(jen.Lit(segments[1])))
		parts = append(parts, jen.Qual(pkgPC, "WithVariant").Call(jen.Lit(segments[2])))
	default:
		parts = append(parts, jen.Qual(pkgPC, "WithArchitecture").Call(jen.Lit(svc.Platform)))
	}
	return parts
}

func genServiceLevelConfig(svc *types.ServiceConfig) []jen.Code {
	var parts []jen.Code
	for k, v := range svc.Annotations {
		parts = append(parts, jen.Qual(pkgSC, "WithAnnotation").Call(jen.Lit(k), jen.Lit(v)))
	}
	if svc.Attach != nil && !*svc.Attach {
		parts = append(parts, jen.Qual(pkgSC, "WithNoAttach").Call())
	}
	if svc.Develop != nil {
		for _, w := range svc.Develop.Watch {
			developArgs := []jen.Code{jen.Lit(string(w.Action)), jen.Lit(w.Path), jen.Lit(w.Target)}
			for _, ig := range w.Ignore {
				developArgs = append(developArgs, jen.Lit(ig))
			}
			parts = append(parts, jen.Qual(pkgSC, "WithDevelop").Call(developArgs...))
		}
	}
	for _, sec := range svc.Secrets {
		secretParts := []jen.Code{
			jen.Qual(pkgSecretSvc, "WithSource").Call(jen.Lit(sec.Source)),
			jen.Qual(pkgSecretSvc, "WithTarget").Call(jen.Lit(sec.Target)),
		}
		if sec.UID != "" {
			secretParts = append(secretParts, jen.Qual(pkgSecretSvc, "WithUID").Call(jen.Lit(sec.UID)))
		}
		if sec.GID != "" {
			secretParts = append(secretParts, jen.Qual(pkgSecretSvc, "WithGID").Call(jen.Lit(sec.GID)))
		}
		if sec.Mode != nil {
			secretParts = append(secretParts, jen.Qual(pkgSecretSvc, "WithMode").Call(jen.Lit(int(*sec.Mode))))
		}
		parts = append(parts, jen.Qual(pkgSC, "WithSecret").Call(secretParts...))
	}
	if svc.Build != nil {
		buildParts := genBuild(svc.Build)
		if len(buildParts) > 0 {
			parts = append(parts, jen.Qual(pkgSC, "WithBuild").Call(buildParts...))
		}
	}
	if len(svc.Profiles) > 0 {
		args := make([]jen.Code, len(svc.Profiles))
		for i, p := range svc.Profiles {
			args[i] = jen.Lit(p)
		}
		parts = append(parts, jen.Qual(pkgSC, "WithProfiles").Call(args...))
	}
	for depName, dep := range svc.DependsOn {
		if dep.Condition == "service_healthy" {
			parts = append(parts, jen.Qual(pkgSC, "WithDependsOnHealthy").Call(jen.Lit(depName)))
		} else {
			parts = append(parts, jen.Qual(pkgSC, "WithDependsOn").Call(jen.Lit(depName)))
		}
	}
	for _, ef := range svc.EnvFiles {
		path := ef.Path
		if path == "" {
			continue
		}
		parts = append(parts, jen.Qual(pkgSC, "WithEnvFile").Call(jen.Lit(path)))
	}
	if svc.Deploy != nil {
		deployParts := genDeploy(svc.Deploy)
		if len(deployParts) > 0 {
			parts = append(parts, jen.Qual(pkgSC, "WithDeploy").Call(deployParts...))
		}
	}
	return parts
}

func genBuild(b *types.BuildConfig) []jen.Code {
	var parts []jen.Code
	if b.Context != "" {
		parts = append(parts, jen.Qual(pkgBuild, "WithContext").Call(jen.Lit(b.Context)))
	}
	if b.Dockerfile != "" {
		parts = append(parts, jen.Qual(pkgBuild, "WithDockerfile").Call(jen.Lit(b.Dockerfile)))
	}
	if b.DockerfileInline != "" {
		parts = append(parts, jen.Qual(pkgBuild, "WithDockerfileInline").Call(jen.Lit(b.DockerfileInline)))
	}
	if b.Target != "" {
		parts = append(parts, jen.Qual(pkgBuild, "WithTarget").Call(jen.Lit(b.Target)))
	}
	for k, v := range b.Args {
		val := ""
		if v != nil {
			val = *v
		}
		parts = append(parts, jen.Qual(pkgBuild, "WithArgs").Call(jen.Lit(k), jen.Lit(val)))
	}
	if len(b.CacheFrom) > 0 {
		args := make([]jen.Code, len(b.CacheFrom))
		for i, c := range b.CacheFrom {
			args[i] = jen.Lit(c)
		}
		parts = append(parts, jen.Qual(pkgBuild, "WithCacheFrom").Call(args...))
	}
	if len(b.CacheTo) > 0 {
		args := make([]jen.Code, len(b.CacheTo))
		for i, c := range b.CacheTo {
			args[i] = jen.Lit(c)
		}
		parts = append(parts, jen.Qual(pkgBuild, "WithCacheTo").Call(args...))
	}
	if b.NoCache {
		parts = append(parts, jen.Qual(pkgBuild, "WithNoCache").Call())
	}
	for _, sshKey := range b.SSH {
		idStr, pathStr := sshKey.ID, sshKey.Path
		if pathStr == "" && strings.Contains(idStr, "=") {
			if idx := strings.Index(idStr, "="); idx >= 0 {
				idStr = strings.TrimSpace(sshKey.ID[:idx])
				pathStr = strings.TrimSpace(sshKey.ID[idx+1:])
			}
		}
		if idStr != "" {
			parts = append(parts, jen.Qual(pkgBuild, "WithSSHKey").Call(jen.Lit(idStr), jen.Lit(pathStr)))
		}
	}
	for k, v := range b.Labels {
		parts = append(parts, jen.Qual(pkgBuild, "WithLabels").Call(jen.Lit(k), jen.Lit(v)))
	}
	if b.Network != "" {
		parts = append(parts, jen.Qual(pkgBuild, "WithNetwork").Call(jen.Lit(b.Network)))
	}
	if b.Isolation != "" {
		parts = append(parts, jen.Qual(pkgBuild, "WithIsolation").Call(jen.Lit(b.Isolation)))
	}
	if b.Pull {
		parts = append(parts, jen.Qual(pkgBuild, "WithPull").Call())
	}
	for _, sec := range b.Secrets {
		secretParts := []jen.Code{
			jen.Qual(pkgSecretSvc, "WithSource").Call(jen.Lit(sec.Source)),
			jen.Qual(pkgSecretSvc, "WithTarget").Call(jen.Lit(sec.Target)),
		}
		if sec.UID != "" {
			secretParts = append(secretParts, jen.Qual(pkgSecretSvc, "WithUID").Call(jen.Lit(sec.UID)))
		}
		if sec.GID != "" {
			secretParts = append(secretParts, jen.Qual(pkgSecretSvc, "WithGID").Call(jen.Lit(sec.GID)))
		}
		if sec.Mode != nil {
			secretParts = append(secretParts, jen.Qual(pkgSecretSvc, "WithMode").Call(jen.Lit(int(*sec.Mode))))
		}
		parts = append(parts, jen.Qual(pkgBuild, "WithSecret").Call(secretParts...))
	}
	if len(b.Tags) > 0 {
		args := make([]jen.Code, len(b.Tags))
		for i, t := range b.Tags {
			args[i] = jen.Lit(t)
		}
		parts = append(parts, jen.Qual(pkgBuild, "WithTags").Call(args...))
	}
	for name, u := range b.Ulimits {
		if u == nil {
			continue
		}
		var ulimitParts []jen.Code
		if u.Single != 0 {
			ulimitParts = append(ulimitParts, jen.Qual(pkgBuildUlimit, "WithSingle").Call(jen.Lit(u.Single)))
		} else {
			if u.Soft != 0 {
				ulimitParts = append(ulimitParts, jen.Qual(pkgBuildUlimit, "WithSoft").Call(jen.Lit(u.Soft)))
			}
			if u.Hard != 0 {
				ulimitParts = append(ulimitParts, jen.Qual(pkgBuildUlimit, "WithHard").Call(jen.Lit(u.Hard)))
			}
		}
		if len(ulimitParts) > 0 {
			parts = append(parts, jen.Qual(pkgBuild, "WithUlimit").Call(append([]jen.Code{jen.Lit(name)}, ulimitParts...)...))
		}
	}
	if len(b.Platforms) > 0 {
		args := make([]jen.Code, len(b.Platforms))
		for i, p := range b.Platforms {
			args[i] = jen.Lit(p)
		}
		parts = append(parts, jen.Qual(pkgBuild, "WithPlatforms").Call(args...))
	}
	if b.Privileged {
		parts = append(parts, jen.Qual(pkgBuild, "WithPrivileged").Call())
	}
	for host, ips := range b.ExtraHosts {
		if len(ips) > 0 {
			args := make([]jen.Code, len(ips)+1)
			args[0] = jen.Lit(host)
			for i, ip := range ips {
				args[i+1] = jen.Lit(ip)
			}
			parts = append(parts, jen.Qual(pkgBuild, "WithExtraHosts").Call(args...))
		}
	}
	for k, v := range b.AdditionalContexts {
		parts = append(parts, jen.Qual(pkgBuild, "WithAdditionalContexts").Call(jen.Lit(k), jen.Lit(v)))
	}
	return parts
}

func genDeploy(d *types.DeployConfig) []jen.Code {
	var parts []jen.Code
	if d.Replicas != nil {
		parts = append(parts, jen.Qual(pkgDeploy, "WithReplicas").Call(jen.Lit(*d.Replicas)))
	}
	if d.Mode != "" {
		parts = append(parts, jen.Qual(pkgDeploy, "WithMode").Call(jen.Lit(d.Mode)))
	}
	for k, v := range d.Labels {
		parts = append(parts, jen.Qual(pkgDeploy, "WithLabel").Call(jen.Lit(k), jen.Lit(v)))
	}
	if d.UpdateConfig != nil {
		uc := d.UpdateConfig
		var updateParts []jen.Code
		if uc.Parallelism != nil {
			updateParts = append(updateParts, jen.Qual(pkgUpdate, "WithParallelism").Call(jen.Lit(int(*uc.Parallelism))))
		}
		if uc.Delay != 0 {
			updateParts = append(updateParts, jen.Qual(pkgUpdate, "WithDelay").Call(jen.Lit(int(time.Duration(uc.Delay).Seconds()))))
		}
		if uc.FailureAction != "" {
			updateParts = append(updateParts, jen.Qual(pkgUpdate, "WithFailureAction").Call(jen.Lit(uc.FailureAction)))
		}
		if uc.Monitor != 0 {
			updateParts = append(updateParts, jen.Qual(pkgUpdate, "WithMonitor").Call(jen.Lit(int(time.Duration(uc.Monitor).Seconds()))))
		}
		if uc.MaxFailureRatio != 0 {
			updateParts = append(updateParts, jen.Qual(pkgUpdate, "WithMaxFailureRatio").Call(jen.Lit(uc.MaxFailureRatio)))
		}
		if uc.Order != "" {
			updateParts = append(updateParts, jen.Qual(pkgUpdate, "WithOrder").Call(jen.Lit(uc.Order)))
		}
		if len(updateParts) > 0 {
			parts = append(parts, jen.Qual(pkgDeploy, "WithUpdateConfig").Call(updateParts...))
		}
	}
	if d.RollbackConfig != nil {
		rc := d.RollbackConfig
		var rollbackParts []jen.Code
		if rc.Parallelism != nil {
			rollbackParts = append(rollbackParts, jen.Qual(pkgUpdate, "WithParallelism").Call(jen.Lit(int(*rc.Parallelism))))
		}
		if rc.Delay != 0 {
			rollbackParts = append(rollbackParts, jen.Qual(pkgUpdate, "WithDelay").Call(jen.Lit(int(time.Duration(rc.Delay).Seconds()))))
		}
		if rc.FailureAction != "" {
			rollbackParts = append(rollbackParts, jen.Qual(pkgUpdate, "WithFailureAction").Call(jen.Lit(rc.FailureAction)))
		}
		if rc.Monitor != 0 {
			rollbackParts = append(rollbackParts, jen.Qual(pkgUpdate, "WithMonitor").Call(jen.Lit(int(time.Duration(rc.Monitor).Seconds()))))
		}
		if rc.MaxFailureRatio != 0 {
			rollbackParts = append(rollbackParts, jen.Qual(pkgUpdate, "WithMaxFailureRatio").Call(jen.Lit(rc.MaxFailureRatio)))
		}
		if rc.Order != "" {
			rollbackParts = append(rollbackParts, jen.Qual(pkgUpdate, "WithOrder").Call(jen.Lit(rc.Order)))
		}
		if len(rollbackParts) > 0 {
			parts = append(parts, jen.Qual(pkgDeploy, "WithRollbackConfig").Call(rollbackParts...))
		}
	}
	if d.Resources.Limits != nil {
		if d.Resources.Limits.MemoryBytes > 0 {
			parts = append(parts, jen.Qual(pkgDeploy, "WithResourceLimits").Call(
				jen.Qual(pkgResource, "WithMemoryBytes").Call(jen.Lit(int(d.Resources.Limits.MemoryBytes))),
			))
		}
		if d.Resources.Limits.NanoCPUs > 0 {
			parts = append(parts, jen.Qual(pkgDeploy, "WithResourceLimits").Call(
				jen.Qual(pkgResource, "WithNanoCPUs").Call(jen.Lit(int(d.Resources.Limits.NanoCPUs))),
			))
		}
		for _, dev := range d.Resources.Limits.Devices {
			deviceParts := genDeviceRequest(dev)
			if len(deviceParts) > 0 {
				parts = append(parts, jen.Qual(pkgDeploy, "WithResourceLimits").Call(jen.Qual(pkgResource, "WithDevice").Call(deviceParts...)))
			}
		}
	}
	if d.Resources.Reservations != nil {
		if d.Resources.Reservations.MemoryBytes > 0 {
			parts = append(parts, jen.Qual(pkgDeploy, "WithResourceReservations").Call(
				jen.Qual(pkgResource, "WithMemoryBytes").Call(jen.Lit(int(d.Resources.Reservations.MemoryBytes))),
			))
		}
		if d.Resources.Reservations.NanoCPUs > 0 {
			parts = append(parts, jen.Qual(pkgDeploy, "WithResourceReservations").Call(
				jen.Qual(pkgResource, "WithNanoCPUs").Call(jen.Lit(int(d.Resources.Reservations.NanoCPUs))),
			))
		}
		for _, dev := range d.Resources.Reservations.Devices {
			deviceParts := genDeviceRequest(dev)
			if len(deviceParts) > 0 {
				parts = append(parts, jen.Qual(pkgDeploy, "WithResourceReservations").Call(jen.Qual(pkgResource, "WithDevice").Call(deviceParts...)))
			}
		}
	}
	return parts
}

func genDeviceRequest(dev types.DeviceRequest) []jen.Code {
	var parts []jen.Code
	if len(dev.Capabilities) > 0 {
		args := make([]jen.Code, len(dev.Capabilities))
		for i, c := range dev.Capabilities {
			args[i] = jen.Lit(c)
		}
		parts = append(parts, jen.Qual(pkgDevice, "WithCapabilities").Call(args...))
	}
	if dev.Driver != "" {
		parts = append(parts, jen.Qual(pkgDevice, "WithDriver").Call(jen.Lit(dev.Driver)))
	}
	if dev.Count != 0 {
		parts = append(parts, jen.Qual(pkgDevice, "WithCount").Call(jen.Lit(int(dev.Count))))
	}
	if len(dev.IDs) > 0 {
		args := make([]jen.Code, len(dev.IDs))
		for i, id := range dev.IDs {
			args[i] = jen.Lit(id)
		}
		parts = append(parts, jen.Qual(pkgDevice, "WithIDs").Call(args...))
	}
	return parts
}
