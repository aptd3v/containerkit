package containerkit

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	dockerNet "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Project is a wrapper around types.Project
// It provides methods to create and manage a docker Compose project.
type Project struct {
	wrapped *types.Project
	errs    []error
}

// SetServiceConfig is a function that sets the service config
type SetServiceConfig func(service *types.ServiceConfig) error

// NewProject creates a new project
// with the given name. It initializes the project with an empty service list.
// The name is used as the project name in the compose file.
func NewProject(name string) *Project {
	return &Project{
		wrapped: &types.Project{
			Name:     name,
			Services: types.Services{},
			Volumes:  make(types.Volumes),
			Networks: make(types.Networks),
			Secrets:  make(types.Secrets),
			Configs:  make(types.Configs),
		},
	}
}

// ForEachService iterates over each service in the project and calls the provided function to provide the ability to mutate a service
// parameters:
//   - fn: the function to call for each service
//
// returns an error if the function returns an error
func (p *Project) ForEachService(fn func(name string, service *types.ServiceConfig) error) error {
	if fn == nil {
		return errdefs.NewProjectConfigError(p.wrapped.Name, "ForEachService function is nil")
	}
	if p.wrapped.Services == nil {
		return errdefs.NewProjectConfigError(p.wrapped.Name, "project has no services")
	}
	for name, service := range p.wrapped.Services {
		if err := fn(name, &service); err != nil {
			return err
		}
	}
	return nil
}

// GetService returns a service from the project
// parameters:
//   - name: the name of the service
//
// returns the service and an error if the service is not found or the project has no services
func (p *Project) GetService(name string) (*types.ServiceConfig, error) {
	if p.wrapped.Services == nil {
		return nil, errdefs.NewProjectConfigError("project", "project has no services")
	}
	service, ok := p.wrapped.Services[name]
	if !ok {
		return nil, errdefs.NewProjectConfigError("project", fmt.Sprintf("service %s not found", name))
	}
	return &service, nil
}

// WithService defines a new service in the project
// parameters:
//   - name: the name of the service
//   - service: the container to create the service from
//   - setters: the setters to apply to the service
func (p *Project) WithService(name string, service *Container, extras ...any) *Project {
	if err := service.Validate(); err != nil {
		p.errs = append(p.errs, errdefs.NewServiceConfigError(name, err.Error()))
		return p
	}

	if _, ok := p.wrapped.Services[name]; ok {
		p.errs = append(p.errs, errdefs.NewServiceConfigError(name, "service already exists"))
		return p
	}

	config := service.Config
	//if the container, host, or network config is nil, we need to set it to an empty object to avoid nil pointer dereference
	if config.Container == nil {
		config.Container = &container.Config{}
	}
	if config.Host == nil {
		config.Host = &container.HostConfig{}
	}
	if config.Network == nil {
		config.Network = &dockerNet.NetworkingConfig{}
	}
	if config.Platform == nil {
		config.Platform = &ocispec.Platform{}
	}

	volumes, tmpfsMounts := convertVolumes(config.Container.Volumes, config.Host)
	// Compose-only keys that have no container field are still unset:
	// configs, credential_spec, extends, external_links, provider, pull_policy,
	// label_file, use_api_socket, post_start, pre_stop, and scale.
	serv := types.ServiceConfig{
		Name: name,
		//container
		Image:       config.Container.Image,
		Command:     types.ShellCommand(config.Container.Cmd),
		Environment: types.NewMappingWithEquals(config.Container.Env),
		Tty:         config.Container.Tty,
		Expose:      convertExposedPorts(config.Container.ExposedPorts),
		HealthCheck: convertHealthCheck(config.Container.Healthcheck),
		Entrypoint:  types.ShellCommand(config.Container.Entrypoint),
		StdinOpen:   config.Container.OpenStdin,
		StopSignal:  config.Container.StopSignal,
		WorkingDir:  config.Container.WorkingDir,
		Labels:      config.Container.Labels,
		DomainName:  config.Container.Domainname,
		Hostname:    config.Container.Hostname,
		User:        config.Container.User,

		//blkio
		BlkioConfig: convertBlkioConfig(config.Host),

		//cap
		CapAdd:  config.Host.CapAdd,
		CapDrop: config.Host.CapDrop,

		//cgroup
		Cgroup:            string(config.Host.Cgroup),
		CgroupParent:      string(config.Host.CgroupParent),
		DeviceCgroupRules: config.Host.DeviceCgroupRules,

		//cpu
		CPUCount:     int64(config.Host.CPUCount),
		CPUPercent:   float32(config.Host.CPUPercent),
		CPUPeriod:    int64(config.Host.CPUPeriod),
		CPUQuota:     int64(config.Host.CPUQuota),
		CPUShares:    int64(config.Host.CPUShares),
		CPUSet:       string(config.Host.CpusetCpus),
		CPUS:         nanoCPUsToCPUS(config.Host.NanoCPUs),
		CPURTRuntime: int64(config.Host.CPURealtimeRuntime),
		CPURTPeriod:  int64(config.Host.CPURealtimePeriod),

		//pids
		Pid: string(config.Host.PidMode),

		//memory
		MemReservation: types.UnitBytes(config.Host.MemoryReservation),
		MemSwapLimit:   types.UnitBytes(config.Host.MemorySwap),
		MemLimit:       types.UnitBytes(config.Host.Memory),
		ShmSize:        types.UnitBytes(config.Host.ShmSize),

		//dns
		DNS:       config.Host.DNS,
		DNSSearch: config.Host.DNSSearch,
		DNSOpts:   config.Host.DNSOptions,

		//oom
		OomScoreAdj: int64(config.Host.OomScoreAdj),
		//Devices
		Devices: convertDevices(config.Host.Devices),
		Gpus:    convertDeviceRequests(config.Host.DeviceRequests),

		//groups
		GroupAdd: config.Host.GroupAdd,
		Links:    config.Host.Links,

		//init
		Init: config.Host.Init,

		//ipc
		Ipc: string(config.Host.IpcMode),

		//isolation
		Isolation: string(config.Host.Isolation),

		//mac address
		MacAddress: config.Container.MacAddress,

		//network
		NetworkMode: string(config.Host.NetworkMode),
		Networks:    convertNetworks(config.Network),

		//logging
		Logging: convertLogging(&config.Host.LogConfig),

		//volumes
		VolumesFrom:  convertVolumesFrom(config.Host.VolumesFrom),
		Volumes:      volumes,
		VolumeDriver: config.Host.VolumeDriver,

		Ports:       convertPortsBindings(config.Host.PortBindings),
		Platform:    formatPlatform(config.Platform),
		Privileged:  config.Host.Privileged,
		ReadOnly:    config.Host.ReadonlyRootfs,
		Restart:     convertRestart(config.Host.RestartPolicy),
		Runtime:     string(config.Host.Runtime),
		SecurityOpt: config.Host.SecurityOpt,
		StorageOpt:  config.Host.StorageOpt,
		Sysctls:     config.Host.Sysctls,
		Tmpfs:       append(convertTmpfs(config.Host.Tmpfs), tmpfsMounts...),
		Ulimits:     convertUlimits(config.Host.Ulimits),
		UserNSMode:  string(config.Host.UsernsMode),
		Uts:         string(config.Host.UTSMode),
	}
	if config.Container.NetworkDisabled && serv.NetworkMode == "" {
		serv.NetworkMode = "none"
	}

	extraHosts, err := convertExtraHosts(config.Host.ExtraHosts)
	if err != nil {
		p.errs = append(p.errs, errdefs.NewServiceConfigError(name, err.Error()))
	} else {
		serv.ExtraHosts = extraHosts
	}

	// the following is nil if not set and needs to stay that way so docker cant determine if it is set or not
	memSwappines := config.Host.MemorySwappiness
	pidsLimit := config.Host.PidsLimit
	oomKillDisable := config.Host.OomKillDisable
	if memSwappines != nil {
		serv.MemSwappiness = types.UnitBytes(*memSwappines)
	}
	if pidsLimit != nil {
		serv.PidsLimit = *pidsLimit
	}
	if oomKillDisable != nil {
		serv.OomKillDisable = *oomKillDisable
	}
	if config.Container.StopTimeout != nil {
		t := types.Duration(time.Second * time.Duration(*config.Container.StopTimeout))
		serv.StopGracePeriod = &t
	}

	for _, extra := range extras {
		if extra == nil {
			continue
		}
		if err := applyServiceExtra(&serv, extra); err != nil {
			p.errs = append(p.errs, errdefs.NewServiceConfigError(name, err.Error()))
			continue
		}
	}
	//swarm mode wants unique container names so we need to only set container name if deploy is not set
	if serv.Deploy != nil {
		serv.ContainerName = ""
	} else {
		serv.ContainerName = service.Name
	}
	if p.wrapped.Services == nil {
		p.wrapped.Services = make(types.Services, 0)
	}
	p.wrapped.Services[name] = serv
	return p
}

// WithVolume defines a new volume in the project
// parameters:
//   - name: the name of the volume
//   - volume: the volume to create the volume from
func (p *Project) WithVolume(name string, spec ...ProjectVolume) *Project {
	if p.wrapped.Volumes == nil {
		p.wrapped.Volumes = make(types.Volumes, 0)
	}
	vol := types.VolumeConfig{
		Name: name,
	}
	if len(spec) > 0 {
		for _, setter := range applyProjectVolume(spec[0]) {
			if setter == nil {
				continue
			}
			if err := setter(&vol); err != nil {
				p.errs = append(p.errs, errdefs.NewProjectConfigError("volume", err.Error()))
			}
		}
	}
	p.wrapped.Volumes[name] = vol
	return p
}

// WithSecret defines a new secret in the project
func (p *Project) WithSecret(key string, spec ...ProjectSecret) *Project {
	if p.wrapped.Secrets == nil {
		p.wrapped.Secrets = make(types.Secrets, 0)
	}
	secret := types.SecretConfig{}
	if len(spec) > 0 {
		for _, setter := range applyProjectSecret(spec[0]) {
			if setter == nil {
				continue
			}
			if err := setter(&secret); err != nil {
				p.errs = append(p.errs, errdefs.NewProjectConfigError("secret", err.Error()))
			}
		}
	}
	p.wrapped.Secrets[key] = secret
	return p
}

// WithNetwork defines a new network in the project
func (p *Project) WithNetwork(name string, spec ...ProjectNetwork) *Project {
	if p.wrapped.Networks == nil {
		p.wrapped.Networks = make(types.Networks, 0)
	}
	net := types.NetworkConfig{
		Name: name,
	}
	if len(spec) > 0 {
		for _, setter := range applyProjectNetwork(name, spec[0]) {
			if setter == nil {
				continue
			}
			if err := setter(&net); err != nil {
				p.errs = append(p.errs, errdefs.NewProjectConfigError("network", err.Error()))
			}
		}
	}
	p.wrapped.Networks[name] = net
	return p
}

// Validate validates the project
// returns an error if the project has errors
func (p *Project) Validate() error {
	errs := []error{}
	// a basic project must have either a image or a build context
	if len(p.wrapped.Services) == 0 {
		errs = append(errs, errdefs.NewProjectConfigError("project", "project must have at least one service"))
	}
	for _, service := range p.wrapped.Services {
		if service.Image == "" && service.Build == nil {
			errs = append(errs, errdefs.NewProjectConfigError("project", fmt.Sprintf("service %s must have either a image or a build context", service.Name)))
			continue
		}
	}
	if len(p.errs) > 0 {
		errs = append(errs, errdefs.NewProjectConfigError("project", errors.Join(p.errs...).Error()))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Marshal marshals the project to a yaml bytes slice
func (p *Project) Marshal() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p.wrapped.MarshalYAML()
}

// Export exports the project to a file
// parameters:
//   - file: the file path to export the project to
//   - perm: the permission of the file
func (p *Project) Export(file string, perm os.FileMode) error {
	if err := p.Validate(); err != nil {
		return err
	}
	yaml, err := p.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(file, yaml, perm)
}

// Unwrap returns the underlying types.Project
func (p *Project) Unwrap() *types.Project {
	return p.wrapped
}

// convertDevices converts the devices from the container config to the compose config
func convertDevices(devices []container.DeviceMapping) []types.DeviceMapping {
	deviceRules := make([]types.DeviceMapping, 0, len(devices))
	for _, device := range devices {
		deviceRules = append(deviceRules, types.DeviceMapping{
			Source:      device.PathOnHost,
			Target:      device.PathInContainer,
			Permissions: string(device.CgroupPermissions),
		})
	}
	return deviceRules
}

// convertVolumesFrom converts the volumes from the container config to the compose config
func convertVolumesFrom(volumes []string) []string {
	volumeRules := make([]string, 0, len(volumes))
	volumeRules = append(volumeRules, volumes...)
	return volumeRules
}

// convertExposedPorts converts the exposed ports from the container config to the compose config
func convertExposedPorts(exposedPorts map[nat.Port]struct{}) types.StringOrNumberList {
	if len(exposedPorts) == 0 {
		return nil
	}
	ports := make(types.StringOrNumberList, 0, len(exposedPorts))
	for port := range exposedPorts {
		ports = append(ports, port.Port())
	}
	return ports
}

// convertHealthCheck converts the health check from the container config to the compose config
func convertHealthCheck(healthcheck *container.HealthConfig) *types.HealthCheckConfig {
	if healthcheck == nil {
		return nil
	}
	if len(healthcheck.Test) == 1 && healthcheck.Test[0] == "NONE" {
		return &types.HealthCheckConfig{Disable: true}
	}
	test := types.HealthCheckTest{}
	if len(healthcheck.Test) > 0 {
		test = append(test, healthcheck.Test...)
	}
	timeout := types.Duration(healthcheck.Timeout)
	interval := types.Duration(healthcheck.Interval)
	retries := uint64(healthcheck.Retries)
	startPeriod := types.Duration(healthcheck.StartPeriod)
	cfg := &types.HealthCheckConfig{
		Test:        test,
		Timeout:     &timeout,
		Interval:    &interval,
		Retries:     &retries,
		StartPeriod: &startPeriod,
	}
	if healthcheck.StartInterval > 0 {
		startInterval := types.Duration(healthcheck.StartInterval)
		cfg.StartInterval = &startInterval
	}
	return cfg
}

// convertBlkioConfig converts the blkio config from the container config to the compose config
func convertBlkioConfig(hostConfig *container.HostConfig) *types.BlkioConfig {
	weightDevice := make([]types.WeightDevice, 0, len(hostConfig.BlkioWeightDevice))
	for _, device := range hostConfig.BlkioWeightDevice {
		weightDevice = append(weightDevice, types.WeightDevice{
			Path:   device.Path,
			Weight: device.Weight,
		})
	}
	deviceReadBps := make([]types.ThrottleDevice, 0, len(hostConfig.BlkioDeviceReadBps))
	for _, device := range hostConfig.BlkioDeviceReadBps {
		deviceReadBps = append(deviceReadBps, types.ThrottleDevice{
			Path: device.Path,
			Rate: types.UnitBytes(device.Rate),
		})
	}
	deviceWriteBps := make([]types.ThrottleDevice, 0, len(hostConfig.BlkioDeviceWriteBps))
	for _, device := range hostConfig.BlkioDeviceWriteBps {
		deviceWriteBps = append(deviceWriteBps, types.ThrottleDevice{
			Path: device.Path,
			Rate: types.UnitBytes(device.Rate),
		})
	}
	deviceReadIOps := make([]types.ThrottleDevice, 0, len(hostConfig.BlkioDeviceReadIOps))
	for _, device := range hostConfig.BlkioDeviceReadIOps {
		deviceReadIOps = append(deviceReadIOps, types.ThrottleDevice{
			Path: device.Path,
			Rate: types.UnitBytes(device.Rate),
		})
	}
	deviceWriteIOps := make([]types.ThrottleDevice, 0, len(hostConfig.BlkioDeviceWriteIOps))
	for _, device := range hostConfig.BlkioDeviceWriteIOps {
		deviceWriteIOps = append(deviceWriteIOps, types.ThrottleDevice{
			Path: device.Path,
			Rate: types.UnitBytes(device.Rate),
		})
	}
	if len(weightDevice) == 0 &&
		len(deviceReadBps) == 0 &&
		len(deviceWriteBps) == 0 &&
		len(deviceReadIOps) == 0 &&
		len(deviceWriteIOps) == 0 &&
		hostConfig.BlkioWeight == 0 {
		//all values zeroed so return nil so compose yaml does not get a empty object {}
		return nil
	}
	return &types.BlkioConfig{
		Weight:          hostConfig.BlkioWeight,
		WeightDevice:    weightDevice,
		DeviceReadBps:   deviceReadBps,
		DeviceWriteBps:  deviceWriteBps,
		DeviceReadIOps:  deviceReadIOps,
		DeviceWriteIOps: deviceWriteIOps,
	}
}

// convertLogging converts the logging config from the container config to the compose config
func convertLogging(logConfig *container.LogConfig) *types.LoggingConfig {
	if logConfig == nil || logConfig.Type == "" {
		// if no logging config is set, return nil so compose does not get a empty object {}
		return nil
	}
	return &types.LoggingConfig{
		Driver:  logConfig.Type,
		Options: logConfig.Config,
	}
}

// convertVolumes converts anonymous container volumes, binds, and mounts into compose service volumes.
// Tmpfs mounts that carry engine options Compose's long form cannot store are returned as short tmpfs strings.
func convertVolumes(anonymous map[string]struct{}, hostConfig *container.HostConfig) ([]types.ServiceVolumeConfig, types.StringList) {
	volumeRules := make([]types.ServiceVolumeConfig, 0)
	var tmpfsMounts types.StringList

	if len(anonymous) > 0 {
		targets := make([]string, 0, len(anonymous))
		for target := range anonymous {
			if target == "" {
				continue
			}
			targets = append(targets, target)
		}
		slices.Sort(targets)
		for _, target := range targets {
			volumeRules = append(volumeRules, types.ServiceVolumeConfig{
				Type:   types.VolumeTypeVolume,
				Target: target,
			})
		}
	}

	binds := hostConfig.Binds
	// binds are validated in the hc package and we can assume they are valid
	for _, bind := range binds {
		// Parse "source:target:mode" (mode may be comma-separated e.g. ro,z)
		last := strings.LastIndex(bind, ":")
		if last == -1 {
			continue
		}
		mode := strings.TrimSpace(bind[last+1:])
		rest := strings.TrimSpace(bind[:last])
		sep := ":/"
		idx := strings.Index(rest, sep)
		if idx == -1 {
			continue
		}
		sourcePath := strings.TrimSpace(rest[:idx])
		targetPath := strings.TrimSpace(rest[idx+len(sep):])
		if targetPath != "" && targetPath[0] != '/' {
			targetPath = "/" + targetPath
		}

		readOnly := strings.Contains(mode, "ro") || strings.Contains(mode, "readonly")
		config := types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   sourcePath,
			Target:   targetPath,
			ReadOnly: readOnly,
		}
		// SELinux labels z (shared) or Z (private) for bind mounts
		for _, part := range strings.Split(mode, ",") {
			part = strings.TrimSpace(part)
			if part == "z" || part == "Z" {
				config.Bind = &types.ServiceVolumeBind{SELinux: part}
				break
			}
		}
		volumeRules = append(volumeRules, config)
	}

	for _, mnt := range hostConfig.Mounts {
		if mnt.Type == "tmpfs" && mnt.TmpfsOptions != nil && len(mnt.TmpfsOptions.Options) > 0 {
			tmpfsMounts = append(tmpfsMounts, formatTmpfsMount(mnt.Target, mnt.ReadOnly, mnt.TmpfsOptions))
			continue
		}
		config := types.ServiceVolumeConfig{
			Source:      mnt.Source,
			Target:      mnt.Target,
			Type:        string(mnt.Type),
			ReadOnly:    mnt.ReadOnly,
			Consistency: string(mnt.Consistency),
		}
		if mnt.BindOptions != nil {
			config.Bind = &types.ServiceVolumeBind{
				Propagation:    string(mnt.BindOptions.Propagation),
				CreateHostPath: mnt.BindOptions.CreateMountpoint,
				Recursive: bindRecursiveMode(
					mnt.BindOptions.NonRecursive,
					mnt.BindOptions.ReadOnlyForceRecursive,
					mnt.BindOptions.ReadOnlyNonRecursive,
				),
			}
		}
		if mnt.VolumeOptions != nil {
			vol := &types.ServiceVolumeVolume{
				NoCopy:  mnt.VolumeOptions.NoCopy,
				Subpath: mnt.VolumeOptions.Subpath,
			}
			if len(mnt.VolumeOptions.Labels) > 0 {
				vol.Labels = mnt.VolumeOptions.Labels
			}
			config.Volume = vol
		}
		if mnt.ImageOptions != nil && mnt.ImageOptions.Subpath != "" {
			config.Image = &types.ServiceVolumeImage{SubPath: mnt.ImageOptions.Subpath}
		}
		if mnt.TmpfsOptions != nil {
			config.Tmpfs = &types.ServiceVolumeTmpfs{
				Size: types.UnitBytes(mnt.TmpfsOptions.SizeBytes),
				Mode: uint32(mnt.TmpfsOptions.Mode),
			}
		}

		volumeRules = append(volumeRules, config)
	}
	return volumeRules, tmpfsMounts
}

// convertPortsBindings converts the ports and bindings from the container config to the compose config
func convertPortsBindings(portBindings map[nat.Port][]nat.PortBinding) []types.ServicePortConfig {
	ports := make([]types.ServicePortConfig, 0, len(portBindings))
	for port, bindings := range portBindings {
		for _, binding := range bindings {
			ports = append(ports, types.ServicePortConfig{
				Target:    uint32(port.Int()),
				HostIP:    binding.HostIP,
				Published: binding.HostPort,
				Protocol:  string(port.Proto()),
			})
		}
	}
	return ports
}

// convertNetworks converts the networks from the container config to the compose config
func convertNetworks(networks *dockerNet.NetworkingConfig) map[string]*types.ServiceNetworkConfig {
	networkRules := make(map[string]*types.ServiceNetworkConfig, len(networks.EndpointsConfig))
	for name, endpointConfig := range networks.EndpointsConfig {
		config := &types.ServiceNetworkConfig{
			GatewayPriority: endpointConfig.GwPriority,
			Aliases:         endpointConfig.Aliases,
			MacAddress:      endpointConfig.MacAddress,
		}
		if endpointConfig.IPAMConfig != nil {
			config.Ipv4Address = endpointConfig.IPAMConfig.IPv4Address
			config.Ipv6Address = endpointConfig.IPAMConfig.IPv6Address
			config.LinkLocalIPs = endpointConfig.IPAMConfig.LinkLocalIPs
		}
		if len(endpointConfig.DriverOpts) > 0 {
			config.DriverOpts = endpointConfig.DriverOpts
		}
		networkRules[name] = config
	}
	return networkRules
}

// convertTmpfs converts the tmpfs from the container config to the compose config
func convertTmpfs(tmpfs map[string]string) types.StringList {
	tmpfsRules := make(types.StringList, 0, len(tmpfs))
	for path, size := range tmpfs {
		tmpfsRules = append(tmpfsRules, fmt.Sprintf("%s:%s", path, size))
	}
	return tmpfsRules
}

// convertExtraHosts converts Docker HostConfig ExtraHosts into compose extra_hosts.
func convertExtraHosts(hosts []string) (types.HostsList, error) {
	if len(hosts) == 0 {
		return nil, nil
	}
	list, err := types.NewHostsList(hosts)
	if err != nil {
		return nil, fmt.Errorf("extra hosts: %w", err)
	}
	return list, nil
}

func nanoCPUsToCPUS(nano int64) float32 {
	if nano == 0 {
		return 0
	}
	return float32(nano) / 1e9
}

func convertRestart(policy container.RestartPolicy) string {
	name := string(policy.Name)
	if policy.Name == container.RestartPolicyOnFailure && policy.MaximumRetryCount > 0 {
		return fmt.Sprintf("%s:%d", name, policy.MaximumRetryCount)
	}
	return name
}

func formatPlatform(p *ocispec.Platform) string {
	if p == nil {
		return ""
	}
	switch {
	case p.OS != "" && p.Architecture != "" && p.Variant != "":
		return p.OS + "/" + p.Architecture + "/" + p.Variant
	case p.OS != "" && p.Architecture != "":
		return p.OS + "/" + p.Architecture
	case p.Architecture != "" && p.Variant != "":
		return p.Architecture + "/" + p.Variant
	case p.OS != "":
		return p.OS
	case p.Architecture != "":
		return p.Architecture
	default:
		return p.Variant
	}
}

func convertDeviceRequests(reqs []container.DeviceRequest) []types.DeviceRequest {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]types.DeviceRequest, 0, len(reqs))
	for _, req := range reqs {
		gpu := types.DeviceRequest{
			Driver: req.Driver,
			Count:  types.DeviceCount(req.Count),
			IDs:    req.DeviceIDs,
		}
		if len(req.Options) > 0 {
			gpu.Options = req.Options
		}
		for _, caps := range req.Capabilities {
			gpu.Capabilities = append(gpu.Capabilities, caps...)
		}
		out = append(out, gpu)
	}
	return out
}

func bindRecursiveMode(nonRecursive, forceReadOnly, writableSubmounts bool) string {
	switch {
	case nonRecursive:
		return "disabled"
	case forceReadOnly:
		return "readonly"
	case writableSubmounts:
		return "writable"
	default:
		return ""
	}
}

func formatTmpfsMount(target string, readOnly bool, opts *mount.TmpfsOptions) string {
	if opts == nil {
		return target
	}
	parts := make([]string, 0, 2+len(opts.Options))
	if opts.SizeBytes > 0 {
		parts = append(parts, fmt.Sprintf("size=%d", opts.SizeBytes))
	}
	if opts.Mode != 0 {
		parts = append(parts, fmt.Sprintf("mode=%o", opts.Mode))
	}
	for _, opt := range opts.Options {
		switch len(opt) {
		case 0:
			continue
		case 1:
			if opt[0] != "" {
				parts = append(parts, opt[0])
			}
		default:
			parts = append(parts, opt[0]+"="+opt[1])
		}
	}
	if readOnly {
		parts = append(parts, "ro")
	}
	if len(parts) == 0 {
		return target
	}
	return target + ":" + strings.Join(parts, ",")
}

// convertUlimits converts the ulimits from the container config to the compose config
func convertUlimits(ulimits []*container.Ulimit) map[string]*types.UlimitsConfig {
	ulimitRules := make(map[string]*types.UlimitsConfig, len(ulimits))
	for _, ulimit := range ulimits {
		if ulimit == nil {
			continue
		}
		ulimitRules[ulimit.Name] = &types.UlimitsConfig{
			Soft: int(ulimit.Soft),
			Hard: int(ulimit.Hard),
		}
	}
	return ulimitRules
}
