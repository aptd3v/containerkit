// Package hc provides the options for the host config.
package hc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	mountType "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
)

// WithMountPoint allows to create a custom mount point for the container in the host mount configuration.
// see mount.go for more details
// parameters:
//   - setMountOptionFn: the function to set the mount option one of the union type of MountSetter interface
func WithMountPoint(set ...mount.SetMountConfig) create.SetHostConfig {

	return func(opt *create.HostConfig) error {
		if opt.Mounts == nil {
			opt.Mounts = make([]mountType.Mount, 0)
		}
		mount := &mountType.Mount{}
		for _, set := range set {
			if set != nil {
				if err := set(mount); err != nil {
					return create.NewHostConfigError("mount", fmt.Sprintf("failed to set mount: %s", err))
				}
			}
		}
		opt.Mounts = append(opt.Mounts, *mount)
		return nil
	}
}

// WithMemoryLimit sets a memory limit for the container in the host configuration.
// parameters:
//   - memory: the memory limit in bytes
func WithMemoryLimit(memory int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.Memory = memory
		return nil
	}
}

// WithRestartAlways sets a restart policy that ensures the container is always restarted upon exit.
// parameters:
//   - maxRetryCount: the maximum number of retries before giving up
func WithRestartAlways(maxRetryCount int) create.SetHostConfig {
	return WithRestartPolicy(RestartPolicyAlways, maxRetryCount)
}

// WithAutoRemove sets the container to be automatically removed when it exits.
func WithAutoRemove() create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.AutoRemove = true
		return nil
	}
}

// WithPortBindings appends port mappings between the host and the container in the host configuration.
// parameters:
//   - protocol: the protocol to use (tcp, udp)
//   - hostIP: the IP address of the host
//   - hostPort: the port on the host
//   - containerPort: the port on the container
func WithPortBindings(protocol, hostIP, hostPort, containerPort string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.PortBindings == nil {
			opt.PortBindings = make(nat.PortMap)
		}
		cPort, err := nat.NewPort(protocol, containerPort)
		if err != nil {
			return fmt.Errorf("invalid container port: %s, %w", containerPort, err)
		}
		hostPort, err := nat.NewPort(protocol, hostPort)
		if err != nil {
			return fmt.Errorf("invalid host port: %s, %w", hostPort, err)
		}

		opt.PortBindings[cPort] = []nat.PortBinding{
			{
				HostIP:   hostIP,
				HostPort: hostPort.Port(),
			},
		}
		return nil
	}
}

// LookupDNS appends a DNS server to the host configuration for the container.
// parameters:
//   - dns: the DNS server to add
func WithDNSLookups(dns ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.DNS == nil {
			opt.DNS = make([]string, 0)
		}
		opt.DNS = append(opt.DNS, dns...)
		return nil
	}
}

// WithDNSOption appends a DNS option to the host configuration for the container.
// parameters:
//   - dnsOption: the DNS option to add
func WithDNSOptions(dnsOption ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.DNSOptions == nil {
			opt.DNSOptions = make(strslice.StrSlice, 0)
		}
		opt.DNSOptions = append(opt.DNSOptions, dnsOption...)
		return nil
	}
}

// WithDNSSearch appends a DNS search domain to the host configuration for the container.
// parameters:
//   - search: the DNS search domain to add
func WithDNSSearches(search ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.DNSSearch == nil {
			opt.DNSSearch = make(strslice.StrSlice, 0)
		}
		opt.DNSSearch = append(opt.DNSSearch, search...)
		return nil
	}
}

// WithExtraHosts appends extra hosts to the host configuration for the container.
// parameters:
//   - extraHosts: the extra hosts to add
func WithExtraHost(extraHosts ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.ExtraHosts == nil {
			opt.ExtraHosts = make([]string, 0)
		}
		opt.ExtraHosts = append(opt.ExtraHosts, extraHosts...)
		return nil
	}
}

// WithAddedGroups appends supplementary groups to the host configuration for the container.
// parameters:
//   - group: the group to add
func WithAddedGroups(group ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.GroupAdd == nil {
			opt.GroupAdd = make(strslice.StrSlice, 0)
		}
		opt.GroupAdd = append(opt.GroupAdd, group...)
		return nil
	}
}

// WithBind appends a volume binding to the host configuration for the container.
// parameters:
//   - bind: the volume binding to add e.g. "/host/path:/container/path:ro"
func WithVolumeBinds(bind ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.Binds == nil {
			opt.Binds = make([]string, 0)
		}
		if err := ValidateMounts(bind); err != nil {
			return err
		}
		opt.Binds = append(opt.Binds, bind...)
		return nil
	}
}

// ValidateMounts validates mount specifications in the format "/source/path:/target/path:mode"
// Whitespace is trimmed from all parts of the specification.
func ValidateMounts(mounts []string) error {
	var errMsgs []string
	seenTargets := make(map[string]bool)

	for i, mount := range mounts {
		if mount = strings.TrimSpace(mount); mount == "" {
			errMsgs = append(errMsgs, fmt.Sprintf("binds[%d]: empty mount specification", i))
			continue
		}

		parts := strings.Split(mount, ":")
		if len(parts) != 3 {
			errMsgs = append(errMsgs, fmt.Sprintf("binds[%d]: invalid format '%s' (must be '/source/path:/target/path:mode')", i, mount))
			continue
		}

		// Trim whitespace from all parts
		sourcePath := strings.TrimSpace(parts[0])
		targetPath := strings.TrimSpace(parts[1])
		mode := strings.TrimSpace(parts[2])

		// Validate source path
		if sourcePath == "" {
			errMsgs = append(errMsgs, fmt.Sprintf("binds[%d]: empty source path", i))
			continue
		}

		// Validate target path
		if targetPath == "" {
			errMsgs = append(errMsgs, fmt.Sprintf("binds[%d]: empty target path", i))
			continue
		}

		// Check for duplicate target paths
		if seenTargets[targetPath] {
			errMsgs = append(errMsgs, fmt.Sprintf("binds[%d]: duplicate target path '%s'", i, targetPath))
			continue
		}
		seenTargets[targetPath] = true

		// Validate mode
		if mode = strings.TrimSpace(mode); mode != "ro" && mode != "rw" {
			errMsgs = append(errMsgs, fmt.Sprintf("binds[%d]: invalid mode '%s' (must be 'ro' or 'rw')", i, mode))
			continue
		}
	}

	if len(errMsgs) > 0 {
		return errors.New(strings.Join(errMsgs, "\n"))
	}
	return nil
}

// WithUTSMode sets the UTS (Unix Timesharing System) namespace mode to be used for the container in the host configuration.
// parameters:
//   - mode: the UTS namespace mode to use
func WithUTSMode(mode string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.UTSMode = container.UTSMode(mode)
		return nil
	}
}

// WithUserNSMode sets the user namespace mode to be used for the container in the host configuration.
// parameters:
//   - mode: the user namespace mode to use
func WithUserNSMode(mode string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.UsernsMode = container.UsernsMode(mode)
		return nil
	}
}

// WithShmSize sets the size of the shared memory file system (/dev/shm) for the container in the host configuration.
// parameters:
//   - size: the size of the shared memory file system in bytes
func WithShmSize(size int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.ShmSize = size
		return nil
	}
}

// WithRuntime sets the runtime for the container in the host configuration.
// parameters:
//   - runtime: the runtime to use
func WithRuntime(runtime string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.Runtime = runtime
		return nil
	}
}

// WithConsoleSize sets the console size for the container in the host configuration.
// parameters:
//   - height: the height of the console
//   - width: the width of the console
func WithConsoleSize(height uint, width uint) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.ConsoleSize = [2]uint{height, width}
		return nil
	}
}

// WithIsolation sets the isolation mode to be used for the container in the host configuration.
// parameters:
//   - isolation: the isolation mode to use
//
// Note: This function applies isolation settings only in Windows environments.
func WithIsolation(isolation string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.Isolation = container.Isolation(isolation)
		return nil
	}
}

// WithCPUCount sets the number of CPUs for the container in the host configuration.
// parameters:
//   - count: the number of CPUs to use
func WithCPUCount(count int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CPUCount = count
		return nil
	}
}

// WithReadonlyPaths appends a list of paths to be marked as read-only in the host configuration.
// parameters:
//   - paths: the paths to mark as read-only
func WithReadonlyPaths(paths ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.ReadonlyPaths == nil {
			opt.ReadonlyPaths = make([]string, 0)
		}
		opt.ReadonlyPaths = append(opt.ReadonlyPaths, paths...)
		return nil
	}
}

// WithMaskedPaths appends a list of paths to be masked inside the container in the host configuration (this overrides the default set of paths).
// parameters:
//   - paths: the paths to mask
func WithMaskedPaths(paths ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.MaskedPaths == nil {
			opt.MaskedPaths = make([]string, 0)
		}
		opt.MaskedPaths = append(opt.MaskedPaths, paths...)
		return nil
	}
}

// WithNetworkMode sets the network mode for the container in the host configuration
// parameters:
//   - mode: the network mode to use
//
// Note: This function applies network settings only in Linux environments.
func WithNetworkMode(mode string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		// Handle container network namespace sharing
		opt.NetworkMode = container.NetworkMode(mode)
		return nil
	}
}

// WithVolumeDriver sets the volume driver for the container in the host configuration
// parameters:
//   - driver: the volume driver to use
//
// note: this is not used in compose
func WithVolumeDriver(driver string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.VolumeDriver = driver
		return nil
	}
}

// WithVolumesFrom appends a list of volumes to inherit from another container, specified in the form <container name>[:<ro|rw>].
// parameters:
//   - from: the container to inherit from
func WithVolumesFrom(from string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.VolumesFrom == nil {
			opt.VolumesFrom = make([]string, 0)
		}
		opt.VolumesFrom = append(opt.VolumesFrom, from)
		return nil
	}
}

// WithIpcMode sets the IPC mode for the container in the host configuration.
// parameters:
//   - mode: the IPC mode to use
func WithIpcMode(mode string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.IpcMode = container.IpcMode(mode)
		return nil
	}
}

// WithCgroup sets the cgroup for the container in the host configuration.
// parameters:
//   - cgroup: the cgroup to use
func WithCgroup(cgroup string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.Cgroup = container.CgroupSpec(cgroup)
		return nil
	}
}

// WithOomScoreAdj sets the OOM score adjustment for the container in the host configuration.
// parameters:
//   - score: the OOM score adjustment
func WithOomScoreAdj(score int) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.OomScoreAdj = score
		return nil
	}
}

// WithOomKillDisable sets the OOM kill disable flag for the container in the host configuration.
// parameters:
//   - oomKillDisable: the OOM kill disable flag to use
func WithOomKillDisable() create.SetHostConfig {
	oomKillDisable := true
	return func(opt *create.HostConfig) error {
		opt.OomKillDisable = &oomKillDisable
		return nil
	}
}

// WithPidMode sets the PID mode for the container in the host configuration.
// parameters:
//   - mode: the PID mode to use
func WithPidMode(mode string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.PidMode = container.PidMode(mode)
		return nil
	}
}

// WithPublishAllPorts sets the publish all ports flag for the container in the host configuration.
// parameters:
//   - publishAllPorts: the publish all ports flag to use
func WithPublishAllPorts(publishAllPorts bool) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.PublishAllPorts = publishAllPorts
		return nil
	}
}

// WithReadOnlyRootfs sets the readonly rootfs flag for the container in the host configuration.
// parameters:
//   - readonlyRootfs: the readonly rootfs flag to use
func WithReadOnlyRootfs() create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.ReadonlyRootfs = true
		return nil
	}
}

// WithSecurityOpts appends a list of string values to customize labels for MLS systems, such as SELinux.
// parameters:
//   - opts: the security options to add
func WithSecurityOpts(opts ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.SecurityOpt == nil {
			opt.SecurityOpt = make([]string, 0)
		}
		opt.SecurityOpt = append(opt.SecurityOpt, opts...)
		return nil
	}
}

// WithStorageOpt appends storage driver options per container to the host configuration.
// parameters:
//   - key: the key of the storage option
//   - value: the value of the storage option
func WithStorageOpt(key, value string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.StorageOpt == nil {
			opt.StorageOpt = make(map[string]string)
		}
		opt.StorageOpt[key] = value
		return nil
	}
}

// WithTmpfs appends to a map of tmpfs (mounts) used for the container
// parameters:
//   - key: the key of the tmpfs option
//   - value: the value of the tmpfs option
func WithTmpfs(key, value string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.Tmpfs == nil {
			opt.Tmpfs = make(map[string]string)
		}
		opt.Tmpfs[key] = value
		return nil
	}
}

// WithPrivileged sets the Privileged mode to the host configuration which allows the following:
// parameters:
//   - privileged: the privileged mode to use
func WithPrivileged() create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.Privileged = true
		return nil
	}
}

// WithDevice adds a device to the host configuration.
// parameters:
//   - device: the device to add
func WithAddedDevice(device string, pathInContainer string, permissions string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.Devices == nil {
			opt.Devices = make([]container.DeviceMapping, 0)
		}
		opt.Devices = append(opt.Devices, container.DeviceMapping{
			PathOnHost:        device,
			PathInContainer:   pathInContainer,
			CgroupPermissions: permissions,
		})
		return nil
	}
}

// WithContainerIDFile adds a containerIDFile to the host configuration.
// After running this command, the /path/to/container-id.txt file will contain the ID of the started container.
// parameters:
//   - containerIDFile: the containerIDFile to add
func WithContainerIDFile(containerIDFile string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.ContainerIDFile = containerIDFile
		return nil
	}
}

// WithCPUShares sets the CPU shares (relative weight) for the container
// parameters:
//   - shares: the CPU shares to use
func WithCPUShares(shares int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CPUShares = shares
		return nil
	}
}

// WithCPUPeriod sets the CPU CFS (Completely Fair Scheduler) period
// parameters:
//   - period: the CPU CFS (Completely Fair Scheduler) period
func WithCPUPeriod(period int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CPUPeriod = period
		return nil
	}
}

// WithCPUPercent sets the CPU percentage for the container
// parameters:
//   - percent: the CPU percentage to use
func WithCPUPercent(percent int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CPUPercent = percent
		return nil
	}
}

// WithCPUQuota sets the CPU CFS (Completely Fair Scheduler) quota
// parameters:
//   - quota: the CPU CFS (Completely Fair Scheduler) quota
func WithCPUQuota(quota int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CPUQuota = quota
		return nil
	}
}

// WithCpusetCpus sets the CPUs in which execution is allowed
// parameters:
//   - cpus: the CPUs in which execution is allowed
func WithCpusetCpus(cpus string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CpusetCpus = cpus
		return nil
	}
}

// WithMemoryReservation sets the memory soft limit
// parameters:
//   - memory: the memory soft limit
func WithMemoryReservation(memory int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.MemoryReservation = memory
		return nil
	}
}

// WithMemorySwap sets the total memory limit (memory + swap)
// parameters:
//   - memorySwap: the total memory limit (memory + swap)
func WithMemorySwap(memorySwap int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.MemorySwap = memorySwap
		return nil
	}
}

// WithUlimits sets ulimit options
// parameters:
//   - name: the name of the ulimit
//   - soft: the soft limit
//   - hard: the hard limit
func WithUlimits(name string, soft, hard int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.Ulimits == nil {
			opt.Ulimits = make([]*container.Ulimit, 0)
		}
		opt.Ulimits = append(opt.Ulimits, &container.Ulimit{
			Name: name,
			Soft: soft,
			Hard: hard,
		})
		return nil
	}
}

// WithInit sets the init flag for the container
// parameters:
//   - init: the init flag to use
//
// Run a custom init inside the container, if null, use the daemon's configured settings
func WithInit() create.SetHostConfig {
	init := true
	return func(opt *create.HostConfig) error {
		opt.Init = &init
		return nil
	}
}

// WithCPURealtimePeriod sets the CPU real-time period in microseconds.
// This option is only applicable when running containers on operating systems
// that support CPU real-time scheduler.
// parameters:
//   - period: the CPU real-time period in microseconds
func WithCPURealtimePeriod(period int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CPURealtimePeriod = period
		return nil
	}
}

// WithCPURealtimeRuntime sets the CPU real-time runtime in microseconds.
// This option is only applicable when running containers on operating systems
// that support CPU real-time scheduler.
// parameters:
//   - runtime: the CPU real-time runtime in microseconds
func WithCPURealtimeRuntime(runtime int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CPURealtimeRuntime = runtime
		return nil
	}
}

// WithCpusetMems sets the memory nodes in which execution is allowed.
// Only effective on NUMA systems.
// parameters:
//   - mems: the memory nodes in which execution is allowed
func WithCpusetMems(mems string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CpusetMems = mems
		return nil
	}
}

// WithMemorySwappiness tunes container memory swappiness (0 to 100).
// - A value of 0 turns off anonymous page swapping.
// - A value of 100 sets the host's swappiness value.
// - Values between 0 and 100 modify the swappiness level accordingly.
// parameters:
//   - swappiness: the swappiness level
func WithMemorySwappiness(swappiness int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.MemorySwappiness = &swappiness
		return nil
	}
}

// WithKernelMemory sets the kernel memory limit in bytes.
// This is the hard limit for kernel memory that cannot be swapped out.
// parameters:
//   - memory: the kernel memory limit in bytes
func WithKernelMemory(memory int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.KernelMemory = memory
		return nil
	}
}

// WithPidsLimit sets the container's PIDs limit.
// parameters:
//   - limit: the PIDs limit
func WithPidsLimit(limit int64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.PidsLimit = &limit
		return nil
	}
}

// WithBlkioWeight sets the block IO weight (relative weight) for the container.
// Weight is a value between 10 and 1000.
// parameters:
//   - weight: the block IO weight
func WithBlkioWeight(weight uint16) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.BlkioWeight = weight
		return nil
	}
}

// WithBlkioDeviceReadBps sets the block IO read rate limit for a device.
// parameters:
//   - devicePath: the path to the device
//   - rate: the read rate limit
func WithBlkioDeviceReadBps(devicePath string, rate uint64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.BlkioDeviceReadBps == nil {
			opt.BlkioDeviceReadBps = make([]*blkiodev.ThrottleDevice, 0)
		}
		opt.BlkioDeviceReadBps = append(opt.BlkioDeviceReadBps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rate,
		})
		return nil
	}
}

// WithBlkioDeviceWriteBps sets the block IO write rate limit for a device.
// parameters:
//   - devicePath: the path to the device
//   - rate: the write rate limit
func WithBlkioDeviceWriteBps(devicePath string, rate uint64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.BlkioDeviceWriteBps == nil {
			opt.BlkioDeviceWriteBps = make([]*blkiodev.ThrottleDevice, 0)
		}
		opt.BlkioDeviceWriteBps = append(opt.BlkioDeviceWriteBps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rate,
		})
		return nil
	}
}

// WithBlkioDeviceReadIOps sets the block IO read rate limit for a device.
// parameters:
//   - devicePath: the path to the device
//   - rate: the read rate limit
func WithBlkioDeviceReadIOps(devicePath string, rate uint64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.BlkioDeviceReadIOps == nil {
			opt.BlkioDeviceReadIOps = make([]*blkiodev.ThrottleDevice, 0)
		}
		opt.BlkioDeviceReadIOps = append(opt.BlkioDeviceReadIOps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rate,
		})
		return nil
	}
}

// WithBlkioDeviceWriteIOps sets the block IO write rate limit for a device.
// parameters:
//   - devicePath: the path to the device
//   - rate: the write rate limit
func WithBlkioDeviceWriteIOps(devicePath string, rate uint64) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.BlkioDeviceWriteIOps == nil {
			opt.BlkioDeviceWriteIOps = make([]*blkiodev.ThrottleDevice, 0)
		}
		opt.BlkioDeviceWriteIOps = append(opt.BlkioDeviceWriteIOps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rate,
		})
		return nil
	}
}

// WithSysctls sets the sysctls for the container
// parameters:
//   - sysctls: the sysctls to use
func WithSysctls(key, value string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.Sysctls == nil {
			opt.Sysctls = make(map[string]string)
		}
		opt.Sysctls[key] = value
		return nil
	}
}

// WithNetworkingSysctls sets the networking sysctls for the container
// applies net.ipv4.ip_forward 1 and net.ipv4.conf.all.rp_filter 1
func WithNetworkingSysctls() create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.Sysctls == nil {
			opt.Sysctls = make(map[string]string)
		}
		opt.Sysctls["net.ipv4.ip_forward"] = "1"
		opt.Sysctls["net.ipv4.conf.all.rp_filter"] = "1"
		return nil
	}
}

// WithDeviceCgroupRules sets the device cgroup rules for the container
// parameters:
//   - rules: the device cgroup rules to use
func WithDeviceCgroupRules(rules ...string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.DeviceCgroupRules == nil {
			opt.DeviceCgroupRules = make([]string, 0)
		}
		opt.DeviceCgroupRules = append(opt.DeviceCgroupRules, rules...)
		return nil
	}
}

// WithCgroupParent sets the cgroup parent for the container
// parameters:
//   - parent: the cgroup parent to use
func WithCgroupParent(parent string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		opt.CgroupParent = parent
		return nil
	}
}

// WithDeviceRequest sets the device request for the container
// parameters:
//   - driver: the driver to use
//   - count: the count of the device
//   - deviceIDs: the device IDs
//   - capabilities: the capabilities of the device
func WithDeviceRequest(driver string, count int, deviceIDs []string, capabilities [][]string) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		if opt.DeviceRequests == nil {
			opt.DeviceRequests = make([]container.DeviceRequest, 0)
		}

		opt.DeviceRequests = append(opt.DeviceRequests, container.DeviceRequest{
			Driver:       driver,
			Count:        count,
			DeviceIDs:    deviceIDs,
			Capabilities: capabilities,
		})
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the host config
// and append the error to the host config error collection
func Fail(err error) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		return create.NewContainerConfigError("host_config", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the host config
// and append the error to the host config error collection
func Failf(stringFormat string, args ...interface{}) create.SetHostConfig {
	return func(opt *create.HostConfig) error {
		return create.NewContainerConfigError("host_config", fmt.Sprintf(stringFormat, args...))
	}
}
