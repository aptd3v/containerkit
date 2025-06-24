package create

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"gopkg.in/yaml.v3"
)

// Project is a wrapper around types.Project
// It provides methods to create and manage a docker Compose project.
type Project struct {
	wrapped *types.Project
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
			Services: []types.ServiceConfig{},
		},
	}
}

// NewService defines a new service in the project
// parameters:
//   - name: the name of the service
//   - service: the container to create the service from
//   - setters: the setters to apply to the service
func (p *Project) NewService(name string, service *Container, setters ...SetServiceConfig) error {
	if err := service.Validate(); err != nil {
		//TODO: proper error type handling
		return err
	}
	config := service.Config

	serv := types.ServiceConfig{
		Name:        name,
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

		//blkio
		BlkioConfig: convertBlkioConfig(config.Host.HostConfig),

		//cap
		CapAdd:  config.Host.HostConfig.CapAdd,
		CapDrop: config.Host.HostConfig.CapDrop,

		//cgroup
		Cgroup:            string(config.Host.HostConfig.Cgroup),
		CgroupParent:      string(config.Host.HostConfig.CgroupParent),
		DeviceCgroupRules: config.Host.HostConfig.DeviceCgroupRules,

		//cpu
		CPUCount:     int64(config.Host.HostConfig.CPUCount),
		CPUPercent:   float32(config.Host.HostConfig.CPUPercent),
		CPUPeriod:    int64(config.Host.HostConfig.CPUPeriod),
		CPUQuota:     int64(config.Host.HostConfig.CPUQuota),
		CPUShares:    int64(config.Host.HostConfig.CPUShares),
		CPUSet:       string(config.Host.HostConfig.CpusetCpus),
		CPURTRuntime: int64(config.Host.HostConfig.CPURealtimeRuntime),
		CPURTPeriod:  int64(config.Host.HostConfig.CPURealtimePeriod),

		//pids
		Pid: string(config.Host.HostConfig.PidMode),

		//memory
		MemReservation: types.UnitBytes(config.Host.HostConfig.MemoryReservation),
		MemSwapLimit:   types.UnitBytes(config.Host.HostConfig.MemorySwap),
		MemLimit:       types.UnitBytes(config.Host.HostConfig.MemoryReservation),
		//dns
		DNS:       config.Host.HostConfig.DNS,
		DNSSearch: config.Host.HostConfig.DNSSearch,
		DNSOpts:   config.Host.HostConfig.DNSOptions,

		//oom
		OomScoreAdj: int64(config.Host.HostConfig.OomScoreAdj),
		//Devices
		Devices: convertDevices(config.Host.HostConfig.Devices),

		//groups
		GroupAdd: config.Host.HostConfig.GroupAdd,

		//init
		Init: config.Host.HostConfig.Init,

		//ipc
		Ipc: string(config.Host.HostConfig.IpcMode),

		//isolation
		Isolation: string(config.Host.HostConfig.Isolation),

		//mac address
		MacAddress: config.Container.MacAddress,

		//network
		NetworkMode: string(config.Host.HostConfig.NetworkMode),
		Networks:    convertNetworks(config.Network.NetworkingConfig),

		//logging
		Logging: convertLogging(&config.Host.HostConfig.LogConfig),

		//volumes
		VolumesFrom: convertVolumesFrom(config.Host.HostConfig.VolumesFrom),

		//volumes
		Volumes:     convertVolumes(config.Host.HostConfig),
		Ports:       convertPortsBindings(config.Host.HostConfig.PortBindings),
		Platform:    config.Platform.Architecture,
		Privileged:  config.Host.HostConfig.Privileged,
		ReadOnly:    config.Host.HostConfig.ReadonlyRootfs,
		Restart:     string(config.Host.HostConfig.RestartPolicy.Name),
		Runtime:     string(config.Host.HostConfig.Runtime),
		SecurityOpt: config.Host.HostConfig.SecurityOpt,
		ShmSize:     types.UnitBytes(config.Host.HostConfig.ShmSize),
		Sysctls:     config.Host.HostConfig.Sysctls,
		Tmpfs:       convertTmpfs(config.Host.HostConfig.Tmpfs),
		Ulimits:     convertUlimits(config.Host.HostConfig.Ulimits),
		User:        config.Container.User,
		UserNSMode:  string(config.Host.HostConfig.UsernsMode),
		Uts:         string(config.Host.HostConfig.UTSMode),
	}

	// the following is nil if not set and needs to stay that way so docker cant determine if it is set or not
	memSwappines := config.Host.HostConfig.MemorySwappiness
	pidsLimit := config.Host.HostConfig.PidsLimit
	oomKillDisable := config.Host.HostConfig.OomKillDisable
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
			return NewServiceConfigError(name, err.Error())
		}
	}
	//swarm mode wants unique container names so we need to only set container name if deploy is not set
	if serv.Deploy != nil {
		serv.ContainerName = ""
	} else {
		serv.ContainerName = config.Name
	}
	p.wrapped.Services = append(p.wrapped.Services, serv)
	return nil
}

// Marshal marshals the project to a yaml string
func (p *Project) Marshal() ([]byte, error) {
	return yaml.Marshal(p.wrapped)
}

// Export exports the project to a file
// parameters:
//   - file: the file path to export the project to
//   - perm: the permission of the file
func (p *Project) Export(file string, perm os.FileMode) error {
	yaml, err := p.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(file, yaml, perm)
}

// convertDevices converts the devices from the container config to the compose config
func convertDevices(devices []container.DeviceMapping) []string {
	deviceRules := make([]string, 0, len(devices))
	for _, device := range devices {
		deviceRules = append(deviceRules, fmt.Sprintf("%s:%s:%s", device.PathOnHost, device.PathInContainer, device.CgroupPermissions))
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
func convertNetworks(networks *network.NetworkingConfig) map[string]*types.ServiceNetworkConfig {
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
