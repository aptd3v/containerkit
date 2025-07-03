package create

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/network"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/secrets/projectsecret"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/volume"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
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
		return NewProjectConfigError(p.wrapped.Name, "ForEachService function is nil")
	}
	if p.wrapped.Services == nil {
		return NewProjectConfigError(p.wrapped.Name, "project has no services")
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
		return nil, NewProjectConfigError("project", "project has no services")
	}
	service, ok := p.wrapped.Services[name]
	if !ok {
		return nil, NewProjectConfigError("project", fmt.Sprintf("service %s not found", name))
	}
	return &service, nil
}

// WithService defines a new service in the project
// parameters:
//   - name: the name of the service
//   - service: the container to create the service from
//   - setters: the setters to apply to the service
func (p *Project) WithService(name string, service *Container, setters ...SetServiceConfig) *Project {
	if err := service.Validate(); err != nil {
		p.errs = append(p.errs, NewServiceConfigError(name, err.Error()))
		return p
	}

	if _, ok := p.wrapped.Services[name]; ok {
		p.errs = append(p.errs, NewServiceConfigError(name, "service already exists"))
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
		CPURTRuntime: int64(config.Host.CPURealtimeRuntime),
		CPURTPeriod:  int64(config.Host.CPURealtimePeriod),

		//pids
		Pid: string(config.Host.PidMode),

		//memory
		MemReservation: types.UnitBytes(config.Host.MemoryReservation),
		MemSwapLimit:   types.UnitBytes(config.Host.MemorySwap),
		MemLimit:       types.UnitBytes(config.Host.MemoryReservation),
		ShmSize:        types.UnitBytes(config.Host.ShmSize),

		//dns
		DNS:       config.Host.DNS,
		DNSSearch: config.Host.DNSSearch,
		DNSOpts:   config.Host.DNSOptions,

		//oom
		OomScoreAdj: int64(config.Host.OomScoreAdj),
		//Devices
		Devices: convertDevices(config.Host.Devices),

		//groups
		GroupAdd: config.Host.GroupAdd,

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
		VolumesFrom: convertVolumesFrom(config.Host.VolumesFrom),
		Volumes:     convertVolumes(config.Host),

		Ports:       convertPortsBindings(config.Host.PortBindings),
		Platform:    config.Platform.Architecture,
		Privileged:  config.Host.Privileged,
		ReadOnly:    config.Host.ReadonlyRootfs,
		Restart:     string(config.Host.RestartPolicy.Name),
		Runtime:     string(config.Host.Runtime),
		SecurityOpt: config.Host.SecurityOpt,
		Sysctls:     config.Host.Sysctls,
		Tmpfs:       convertTmpfs(config.Host.Tmpfs),
		Ulimits:     convertUlimits(config.Host.Ulimits),
		UserNSMode:  string(config.Host.UsernsMode),
		Uts:         string(config.Host.UTSMode),
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

	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&serv); err != nil {
			p.errs = append(p.errs, NewServiceConfigError(name, err.Error()))
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
func (p *Project) WithVolume(name string, setters ...volume.SetVolumeProjectConfig) *Project {
	if p.wrapped.Volumes == nil {
		p.wrapped.Volumes = make(types.Volumes, 0)
	}
	volume := types.VolumeConfig{
		Name: name,
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&volume); err != nil {
			p.errs = append(p.errs, NewProjectConfigError("volume", err.Error()))
		}
	}
	p.wrapped.Volumes[name] = volume
	return p
}

// WithSecret defines a new secret in the project
func (p *Project) WithSecret(key string, setters ...projectsecret.SetProjectSecretConfig) *Project {
	if p.wrapped.Secrets == nil {
		p.wrapped.Secrets = make(types.Secrets, 0)
	}
	secret := types.SecretConfig{}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&secret); err != nil {
			p.errs = append(p.errs, NewProjectConfigError("secret", err.Error()))
		}
	}
	p.wrapped.Secrets[key] = secret
	return p
}

// WithNetwork defines a new network in the project
func (p *Project) WithNetwork(name string, setters ...network.SetNetworkProjectConfig) *Project {
	if p.wrapped.Networks == nil {
		p.wrapped.Networks = make(types.Networks, 0)
	}
	network := types.NetworkConfig{
		Name: name,
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&network); err != nil {
			p.errs = append(p.errs, NewProjectConfigError("network", err.Error()))
		}
	}
	p.wrapped.Networks[name] = network
	return p
}

// Validate validates the project
// returns an error if the project has errors
func (p *Project) Validate() error {
	// a basic project must have either a image or a build context
	if len(p.wrapped.Services) == 0 {
		return NewProjectConfigError("project", "project must have at least one service")
	}
	for _, service := range p.wrapped.Services {
		if service.Image == "" && service.Build == nil {
			p.errs = append(p.errs, NewProjectConfigError("project", fmt.Sprintf("service %s must have either a image or a build context", service.Name)))
			continue
		}
	}
	if len(p.errs) > 0 {
		return NewProjectConfigError("project", errors.Join(p.errs...).Error())
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
	test := types.HealthCheckTest{}
	if len(healthcheck.Test) > 0 {
		test = append(test, healthcheck.Test...)
	}
	timeout := types.Duration(healthcheck.Timeout)
	interval := types.Duration(healthcheck.Interval)
	retries := uint64(healthcheck.Retries)
	startPeriod := types.Duration(healthcheck.StartPeriod)
	return &types.HealthCheckConfig{
		Test:        test,
		Timeout:     &timeout,
		Interval:    &interval,
		Retries:     &retries,
		StartPeriod: &startPeriod,
	}
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

// convertVolumes converts the volumes and binds from the container config to the compose config
func convertVolumes(hostConfig *container.HostConfig) []types.ServiceVolumeConfig {
	volumeRules := make([]types.ServiceVolumeConfig, 0)

	binds := hostConfig.Binds
	//binds are validated in the hc package and we can assume they are valid
	for _, bind := range binds {
		// Trim whitespace from all parts
		parts := strings.Split(bind, ":")
		sourcePath := strings.TrimSpace(parts[0])
		targetPath := strings.TrimSpace(parts[1])
		mode := strings.TrimSpace(parts[2])

		volumeRules = append(volumeRules, types.ServiceVolumeConfig{
			Type:     "bind", // Always bind type for WithVolumeBinds
			Source:   sourcePath,
			Target:   targetPath,
			ReadOnly: mode == "ro",
		})
	}

	mounts := hostConfig.Mounts
	for _, mount := range mounts {
		config := types.ServiceVolumeConfig{
			Source:      mount.Source,
			Target:      mount.Target,
			Type:        string(mount.Type),
			ReadOnly:    mount.ReadOnly,
			Consistency: string(mount.Consistency),
		}
		if mount.BindOptions != nil {
			config.Bind = &types.ServiceVolumeBind{
				Propagation:    string(mount.BindOptions.Propagation),
				CreateHostPath: mount.BindOptions.CreateMountpoint,
			}
		}
		if mount.VolumeOptions != nil {
			config.Volume = &types.ServiceVolumeVolume{
				NoCopy: mount.VolumeOptions.NoCopy,
			}
		}
		if mount.TmpfsOptions != nil {
			config.Tmpfs = &types.ServiceVolumeTmpfs{
				Size: types.UnitBytes(mount.TmpfsOptions.SizeBytes),
				Mode: uint32(mount.TmpfsOptions.Mode),
			}
		}

		volumeRules = append(volumeRules, config)
	}
	return volumeRules
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
			Priority:   endpointConfig.Copy().GwPriority,
			Aliases:    endpointConfig.Aliases,
			MacAddress: endpointConfig.MacAddress,
		}
		if endpointConfig.IPAMConfig != nil {
			config.Ipv4Address = endpointConfig.IPAMConfig.IPv4Address
			config.Ipv6Address = endpointConfig.IPAMConfig.IPv6Address
			config.LinkLocalIPs = endpointConfig.IPAMConfig.LinkLocalIPs
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
