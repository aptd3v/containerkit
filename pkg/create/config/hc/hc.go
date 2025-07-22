// Package hc provides the options for the host config.
package hc

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	mountType "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
)

// WithMountPoint allows to create a custom mount point for the container in the host mount configuration.
// see mount.go for more details
// parameters:
//   - setters: the setter functions to set config for the mount point
func WithMountPoint(setters ...mount.SetMountConfig) create.SetHostConfig {

	return func(opt *container.HostConfig) error {
		if opt.Mounts == nil {
			opt.Mounts = make([]mountType.Mount, 0)
		}
		mount := &mountType.Mount{}
		for _, set := range setters {
			if set != nil {
				if err := set(mount); err != nil {
					return errdefs.NewHostConfigError("mount", fmt.Sprintf("failed to set mount: %s", err))
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
//
// note: the memory limit can be specified as an integer in bytes or as a string with a unit (e.g. "100M", "1G")
func WithMemoryLimit[T int | string](memory T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		switch v := any(memory).(type) {
		case int:
			opt.Memory = int64(v)
		case string:
			// parse the memory limit
			memory, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("memory", err.Error())
			}
			opt.Memory = memory
		}
		return nil
	}
}

// WithRestartAlways sets a restart policy that ensures the container is always restarted upon exit.
// This policy restarts the container indefinitely, ignoring any retry limits.
func WithRestartAlways() create.SetHostConfig {
	return WithRestartPolicy(RestartPolicyAlways, 0)
}

// WithAutoRemove sets the container to be automatically removed when it exits.
func WithAutoRemove() create.SetHostConfig {
	return func(opt *container.HostConfig) error {
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
	return func(opt *container.HostConfig) error {

		if containerPort == "" {
			return errdefs.NewHostConfigError("port", "empty container port")
		}
		if hostPort == "" {
			return errdefs.NewHostConfigError("port", "empty host port")
		}
		if hostIP == "" {
			return errdefs.NewHostConfigError("port", "empty host IP")
		}
		if protocol == "" {
			return errdefs.NewHostConfigError("port", "empty protocol")
		}

		cPort, err := nat.NewPort(protocol, containerPort)
		if err != nil {
			return errdefs.NewHostConfigError("port", err.Error())
		}
		hostPort, err := nat.NewPort(protocol, hostPort)
		if err != nil {
			return errdefs.NewHostConfigError("port", err.Error())
		}

		if opt.PortBindings == nil {
			opt.PortBindings = make(nat.PortMap)
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
	return func(opt *container.HostConfig) error {
		if opt.DNS == nil {
			opt.DNS = make([]string, 0, len(dns))
		}
		opt.DNS = append(opt.DNS, dns...)
		return nil
	}
}

// WithDNSOption appends a DNS option to the host configuration for the container.
// parameters:
//   - dnsOption: the DNS option to add
func WithDNSOptions(dnsOption ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.DNSOptions == nil {
			opt.DNSOptions = make(strslice.StrSlice, 0, len(dnsOption))
		}
		opt.DNSOptions = append(opt.DNSOptions, dnsOption...)
		return nil
	}
}

// WithDNSSearch appends a DNS search domain to the host configuration for the container.
// parameters:
//   - search: the DNS search domain to add
func WithDNSSearches(search ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.DNSSearch == nil {
			opt.DNSSearch = make(strslice.StrSlice, 0, len(search))
		}
		opt.DNSSearch = append(opt.DNSSearch, search...)
		return nil
	}
}

// WithExtraHosts appends extra hosts to the host configuration for the container.
// parameters:
//   - extraHosts: the extra hosts to add
func WithExtraHost(extraHosts ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.ExtraHosts == nil {
			opt.ExtraHosts = make([]string, 0, len(extraHosts))
		}
		opt.ExtraHosts = append(opt.ExtraHosts, extraHosts...)
		return nil
	}
}

// WithAddedGroups appends supplementary groups to the host configuration for the container.
// parameters:
//   - group: the group to add
func WithAddedGroups(group ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.GroupAdd == nil {
			opt.GroupAdd = make(strslice.StrSlice, 0, len(group))
		}
		opt.GroupAdd = append(opt.GroupAdd, group...)
		return nil
	}
}

// WithBind appends a volume binding to the host configuration for the container.
// parameters:
//   - binds: the volume binding to add e.g. "/host/path:/container/path:ro"
func WithVolumeBinds(binds ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.Binds == nil {
			opt.Binds = make([]string, 0, len(binds))
		}
		if err := validateMounts(binds); err != nil {
			return err
		}
		opt.Binds = append(opt.Binds, binds...)
		return nil
	}
}

// ValidateMounts validates mount specifications in the format "/source/path:/target/path:mode"
// Whitespace is trimmed from all parts of the specification.
func validateMounts(mounts []string) error {
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
		return errdefs.NewHostConfigError("mounts", strings.Join(errMsgs, "\n"))
	}
	return nil
}

// WithUTSMode sets the UTS (Unix Timesharing System) namespace mode to be used for the container in the host configuration.
// parameters:
//   - mode: the UTS namespace mode to use
func WithUTSMode(mode string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if mode == "" {
			return errdefs.NewHostConfigError("uts_mode", "mode cannot be empty")
		}
		opt.UTSMode = container.UTSMode(mode)
		return nil
	}
}

// WithUserNSMode sets the user namespace mode for the container in the host configuration.
// parameters:
//   - mode: the user namespace mode to use, e.g., "host", "private", "container:<id>"
func WithUserNSMode(mode string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if mode == "" {
			return errdefs.NewHostConfigError("userns_mode", "mode cannot be empty")
		}
		opt.UsernsMode = container.UsernsMode(mode)
		return nil
	}
}

// WithShmSize sets the size of the shared memory file system (/dev/shm) for the container in the host configuration.
// parameters:
//   - size: the size of the shared memory file system in bytes
//
// note: the size can be specified as an integer in bytes or as a string with a unit (e.g. "100M", "1G")
func WithShmSize[T int | string](size T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		switch v := any(size).(type) {
		case int:
			opt.ShmSize = int64(v)
		case string:
			shmSize, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("shm_size", err.Error())
			}
			opt.ShmSize = shmSize
		}
		return nil
	}
}

// WithRuntime sets the OCI runtime for the container in the host configuration.
// parameters:
//   - runtime: the runtime to use, e.g., "runc", "nvidia", or custom runtimes installed on the host.
func WithRuntime(runtime string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.Runtime = runtime
		return nil
	}
}

// WithConsoleSize sets the console size for the container in the host configuration.
// parameters:
//   - height: the height of the console
//   - width: the width of the console
func WithConsoleSize(height uint, width uint) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.ConsoleSize = [2]uint{height, width}
		return nil
	}
}

// WithIsolation sets the isolation mode for the container in the host configuration.
// Parameters:
//   - isolation: the isolation mode to use. Valid values typically include "default", "process", and "hyperv".
//
// Note: This function applies isolation settings only on Windows environments.
func WithIsolation(isolation string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		switch isolation {
		case "", "default", "process", "hyperv":
			opt.Isolation = container.Isolation(isolation)
			return nil
		default:
			return errdefs.NewHostConfigError("isolation", "invalid isolation mode")
		}
	}
}

// WithCPUCount sets the number of CPUs allocated to the container in the host configuration.
//
// Parameters:
//   - count: the number of CPUs to allocate to the container.
//
// Returns an error if count is less than 1.
func WithCPUCount(count int64) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if count < 1 {
			return errdefs.NewHostConfigError("cpu_count", "CPU count must be at least 1")
		}
		opt.CPUCount = count
		return nil
	}
}

// WithReadonlyPaths appends a list of paths to be marked as read-only inside the container.
//
// These paths are mounted as read-only from the host and are inaccessible for writing
// by any process in the container.
//
// Parameters:
//   - paths: one or more absolute paths to mark as read-only
//
// Returns an error if any path is empty or not absolute.
func WithReadonlyPaths(paths ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		for _, path := range paths {
			if strings.TrimSpace(path) == "" {
				return errdefs.NewHostConfigError("readonly_paths", "path cannot be empty")
			}
			if !strings.HasPrefix(path, "/") {
				return errdefs.NewHostConfigError("readonly_paths", fmt.Sprintf("path must be absolute: %q", path))
			}
		}
		if opt.ReadonlyPaths == nil {
			opt.ReadonlyPaths = make([]string, 0, len(paths))
		}
		opt.ReadonlyPaths = append(opt.ReadonlyPaths, paths...)
		return nil
	}
}

// WithMaskedPaths appends a list of paths to be masked inside the container.
//
// Masked paths are mounted as read-only and inaccessible from within the container,
// overriding the default masked paths (e.g., "/proc/kcore", "/proc/latency_stats").
//
// Parameters:
//   - paths: one or more absolute paths to mask
//
// Returns an error if any path is empty or not absolute.
func WithMaskedPaths(paths ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		for _, path := range paths {
			if strings.TrimSpace(path) == "" {
				return errdefs.NewHostConfigError("masked_paths", "path cannot be empty")
			}
			if !strings.HasPrefix(path, "/") {
				return errdefs.NewHostConfigError("masked_paths", fmt.Sprintf("path must be absolute: %q", path))
			}
		}
		if opt.MaskedPaths == nil {
			opt.MaskedPaths = make([]string, 0, len(paths))
		}
		opt.MaskedPaths = append(opt.MaskedPaths, paths...)
		return nil
	}
}

// WithNetworkMode sets the network mode for the container in the host configuration.
//
// Accepts standard Docker network modes such as:
//   - "bridge"    (default)
//   - "host"      (shares host's network stack)
//   - "none"      (no network)
//   - "container:<id|name>" (joins another container's network namespace)
//
// Returns an error if the mode is empty or not one of the supported formats.
func WithNetworkMode(mode string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		mode = strings.TrimSpace(mode)
		if mode == "" ||
			mode == "bridge" ||
			mode == "host" ||
			mode == "none" ||
			strings.HasPrefix(mode, "container:") {
			opt.NetworkMode = container.NetworkMode(mode)
			return nil
		}
		return errdefs.NewHostConfigError("network_mode", fmt.Sprintf("invalid network mode: %q", mode))
	}
}

// WithVolumeDriver sets the volume driver for the container in the host configuration.
//
// The volume driver specifies the plugin or mechanism used to manage volumes mounted into the container.
// This option is typically only used with the Docker API directly, not with Docker Compose.
//
// Parameters:
//   - driver: the name of the volume driver to use.
//
// Returns an error if the driver name is empty or contains only whitespace.
func WithVolumeDriver(driver string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		driver = strings.TrimSpace(driver)
		if driver == "" {
			return errdefs.NewHostConfigError("volume_driver", "volume driver cannot be empty")
		}
		opt.VolumeDriver = driver
		return nil
	}
}

// WithVolumesFrom appends a list of volumes to inherit from another container, specified in the form <container name>[:<ro|rw>].
// parameters:
//   - from: the container to inherit from
func WithVolumesFrom(from string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.VolumesFrom == nil {
			opt.VolumesFrom = make([]string, 0, len(from))
		}
		opt.VolumesFrom = append(opt.VolumesFrom, from)
		return nil
	}
}

// WithIpcMode sets the IPC (Inter-Process Communication) mode for the container.
//
// IPC mode controls how processes inside the container share memory and other IPC resources.
// Common valid values include:
//   - ""              (default — isolated IPC namespace)
//   - "host"          (use the host's IPC namespace)
//   - "private"       (use a private IPC namespace)
//   - "shareable"     (allows other containers to join this container's IPC namespace)
//   - "container:<id>" (join another container’s IPC namespace)
//
// Parameters:
//   - mode: the IPC mode to use.
//
// Returns an error if the mode is invalid or improperly formatted.
func WithIpcMode(mode string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if mode == "" || mode == "host" || mode == "private" || mode == "shareable" || strings.HasPrefix(mode, "container:") {
			opt.IpcMode = container.IpcMode(mode)
			return nil
		}
		return errdefs.NewHostConfigError("ipc_mode", fmt.Sprintf("invalid IPC mode: %q", mode))
	}
}

// WithCgroup sets the cgroup namespace mode for the container in the host configuration.
//
// The cgroup namespace determines how the container is isolated in terms of resource control.
// Valid values include:
//   - ""           (default behavior — Docker decides)
//   - "host"       (container shares the host's cgroup namespace)
//   - "private"    (container gets its own cgroup namespace)
//   - "none"       (disable cgroup namespace, if supported)
//
// Parameters:
//   - cgroup: the cgroup namespace mode to use
//
// Returns an error if the cgroup mode is invalid.
func WithCgroup(cgroup string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		valid := map[string]bool{
			"": true, "host": true, "private": true, "none": true,
		}
		if !valid[cgroup] {
			return errdefs.NewHostConfigError("cgroup", fmt.Sprintf("invalid cgroup mode: %q", cgroup))
		}
		opt.Cgroup = container.CgroupSpec(cgroup)
		return nil
	}
}

// WithOomScoreAdj sets the Out-Of-Memory (OOM) score adjustment for the container.
//
// The OOM score adjustment ranges from -1000 to 1000 and tells the kernel how likely
// it is to kill the container’s process when the system is under memory pressure.
// - A value closer to -1000 means the process is protected from OOM killing.
// - A value closer to 1000 makes the process more likely to be killed.
//
// Parameters:
//   - score: an integer between -1000 and 1000 (inclusive)
//
// Returns an error if the score is outside the valid range.
func WithOomScoreAdj(score int) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if score < -1000 || score > 1000 {
			return errdefs.NewHostConfigError("oom_score_adj", "value must be between -1000 and 1000")
		}
		opt.OomScoreAdj = score
		return nil
	}
}

// WithOomKillDisable sets the OOM kill disable flag for the container in the host configuration.
// parameters:
//   - oomKillDisable: the OOM kill disable flag to use
func WithOomKillDisable() create.SetHostConfig {
	oomKillDisable := true
	return func(opt *container.HostConfig) error {
		opt.OomKillDisable = &oomKillDisable
		return nil
	}
}

// WithPidMode sets the PID mode for the container in the host configuration.
//
// PID mode controls the process ID namespace isolation of the container.
// Common values include "host", "container:<name|id>", or an empty string for private namespace.
//
// parameters:
//   - mode: the PID mode to use
func WithPidMode(mode string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.PidMode = container.PidMode(mode)
		return nil
	}
}

// WithPublishAllPorts sets the publish all ports flag to true for the container
func WithPublishAllPorts() create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.PublishAllPorts = true
		return nil
	}
}

// WithReadOnlyRootfs sets the readonly rootfs flag for the container in the host configuration.
// parameters:
//   - readonlyRootfs: the readonly rootfs flag to use
func WithReadOnlyRootfs() create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.ReadonlyRootfs = true
		return nil
	}
}

// WithSecurityOpts appends security options to the container's host configuration.
//
// These options customize security labels or settings used by Mandatory Access Control (MAC) systems
// like SELinux, AppArmor, or Seccomp. They allow fine-tuning container security contexts,
// such as specifying custom SELinux labels or disabling certain security profiles.
//
// Docker itself does minimal validation on these strings, so this function simply appends them.
// Users should provide valid options according to their security system's documentation.
//
// parameters:
//   - opts: a variadic list of security options to append (e.g., SELinux labels or profile overrides)
func WithSecurityOpts(opts ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.SecurityOpt == nil {
			opt.SecurityOpt = make([]string, 0, len(opts))
		}
		opt.SecurityOpt = append(opt.SecurityOpt, opts...)
		return nil
	}
}

// WithStorageOpt appends a storage driver option for the container's host configuration.
// Storage options are key-value pairs passed to the container storage driver to configure
// specific behaviors like size limits, encryption, or performance settings.
//
// Since storage options are driver-specific and Docker itself does minimal validation,
// this function trims whitespace and requires a non-empty key to avoid invalid or
// meaningless options that might cause runtime errors or unexpected behavior.
//
// parameters:
//   - key: the storage option key (must be non-empty after trimming whitespace)
//   - value: the value for the storage option
func WithStorageOpt(key, value string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		key = strings.TrimSpace(key)
		if key == "" {
			return errdefs.NewHostConfigError("storage_opt", "storage option key cannot be empty")
		}
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
	return func(opt *container.HostConfig) error {
		if opt.Tmpfs == nil {
			opt.Tmpfs = make(map[string]string)
		}
		opt.Tmpfs[key] = value
		return nil
	}
}

// WithPrivileged enables privileged mode for the container.
// Privileged mode grants the container extended Linux capabilities and access to devices.
func WithPrivileged() create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.Privileged = true
		return nil
	}
}

// WithAddedDevice adds a device mapping to the container's host configuration.
//
// This allows the container to access a device from the host (e.g., /dev/snd, /dev/ttyUSB0).
//
// Parameters:
//   - device: the path to the device on the host (e.g., "/dev/snd").
//   - pathInContainer: the path the device will be available at inside the container (e.g., "/dev/snd").
//   - permissions: cgroup permissions for the device ("r", "w", "m", or combinations like "rw").
//
// Returns an error if any parameter is empty or permissions contain invalid characters.
func WithAddedDevice(device string, pathInContainer string, permissions string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if strings.TrimSpace(device) == "" {
			return errdefs.NewHostConfigError("device", "host device path cannot be empty")
		}
		if strings.TrimSpace(pathInContainer) == "" {
			return errdefs.NewHostConfigError("device", "container device path cannot be empty")
		}
		if !isValidDevicePermission(permissions) {
			return errdefs.NewHostConfigError("device", "invalid device permissions (must be combination of 'r', 'w', 'm')")
		}

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

// isValidDevicePermission checks if permissions contain only valid cgroup characters: r (read), w (write), m (mknod).
func isValidDevicePermission(p string) bool {
	seen := map[rune]bool{}
	for _, r := range p {
		switch r {
		case 'r', 'w', 'm':
			if seen[r] {
				return false // disallow duplicates
			}
			seen[r] = true
		default:
			return false
		}
	}
	return len(p) > 0
}

// WithContainerIDFile sets the path to a file where the container ID will be written after creation.
//
// After `client.ContainerCreate`, Docker will write the container's ID to this file.
// This is useful for external tooling or scripts that need to reference the container after it starts.
//
// Parameters:
//   - path: absolute or relative path to the container ID file.
//
// Returns an error if the path is empty or only whitespace.
func WithContainerIDFile(path string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if strings.TrimSpace(path) == "" {
			return errdefs.NewHostConfigError("container_id_file", "path cannot be empty")
		}
		opt.ContainerIDFile = path
		return nil
	}
}

// WithCPUShares sets the CPU shares (relative weight) for the container.
//
// CPU shares define the relative CPU time available to the container compared to others.
// For example, 1024 is the default (normal priority), 512 is half the CPU weight, 2048 is double.
//
// Parameters:
//   - shares: the number of CPU shares (must be a positive integer).
//
// Returns an error if the value is less than or equal to zero.
func WithCPUShares(shares int64) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if shares <= 0 {
			return errdefs.NewHostConfigError("cpu_shares", "CPU shares must be a positive integer")
		}
		opt.CPUShares = shares
		return nil
	}
}

// WithCPUPeriod sets the CPU CFS (Completely Fair Scheduler) period in microseconds.
//
// This defines the time window used by the CFS quota system to restrict CPU usage.
// It is typically used with CPUQuota to control CPU bandwidth allocation.
//
// Accepts either:
//   - an int value in microseconds (e.g., 100000 for 100ms), or
//   - a duration string (e.g., "100ms", "1s").
//
// Valid values range from 1,000 to 1,000,000 microseconds (1ms to 1s), inclusive.
// Returns an error if the input is out of range or invalid.
func WithCPUPeriod[T int | string](period T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		var micros int64

		switch v := any(period).(type) {
		case int:
			micros = int64(v)
		case string:
			dur, err := time.ParseDuration(v)
			if err != nil {
				return errdefs.NewHostConfigError("cpu_period", err.Error())
			}
			micros = dur.Microseconds()
		}

		if micros < 1000 || micros > 1_000_000 {
			return errdefs.NewHostConfigError("cpu_period", "CPU period must be between 1,000 and 1,000,000 microseconds")
		}

		opt.CPUPeriod = micros
		return nil
	}
}

// WithCPUPercent sets the CPU percentage limit for the container.
//
// This option specifies the maximum amount of CPU the container can use as a percentage
// (e.g., 50 means the container can use up to 50% of one CPU core).
//
// Parameters:
//   - percent: the desired CPU limit as a percentage (0–100, or more for multi-core allocation).
//
// Returns an error if the percent is negative.
func WithCPUPercent(percent int64) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if percent < 0 {
			return errdefs.NewHostConfigError("cpu_percent", "CPU percent must be non-negative")
		}
		opt.CPUPercent = percent
		return nil
	}
}

// WithCPUQuota sets the CPU CFS (Completely Fair Scheduler) quota for the container.
//
// The CPU quota limits the total CPU time that all tasks in a container can use during one period.
// A value of 0 means no quota (no limit).
//
// Parameters:
//   - quota: the CPU quota in microseconds (e.g., 100000 for 100ms)
func WithCPUQuota(quota int64) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.CPUQuota = quota
		return nil
	}
}

// WithCpusetCpus sets the CPUs in which execution is allowed
// parameters:
//   - cpus: the CPUs in which execution is allowed
func WithCpusetCpus(cpus string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if cpus == "" {
			return errdefs.NewHostConfigError("cpuset_cpus", "cpus is required")
		}
		if err := validateCpuset(cpus); err != nil {
			return errdefs.NewHostConfigError("cpuset_cpus", err.Error())
		}
		opt.CpusetCpus = cpus
		return nil
	}
}

// WithMemoryReservation sets the memory soft limit
// parameters:
//   - memory: the memory soft limit
//
// note: the memory limit can be specified as an integer in bytes or as a string with a unit (e.g. "100M", "1G")
func WithMemoryReservation[T int | string](memory T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		switch v := any(memory).(type) {
		case int:
			opt.MemoryReservation = int64(v)
		case string:
			memory, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("memory_reservation", err.Error())
			}
			opt.MemoryReservation = memory

		}
		return nil
	}
}

// WithMemorySwap sets the total memory limit (memory + swap)
// parameters:
//   - memorySwap: the total memory limit (memory + swap)
//
// note: the memory limit can be specified as an integer in bytes or as a string with a unit (e.g. "100M", "1G")
func WithMemorySwap[T int | string](memorySwap T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		switch v := any(memorySwap).(type) {
		case int:
			opt.MemorySwap = int64(v)
		case string:
			memorySwap, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("memory_swap", err.Error())
			}
			opt.MemorySwap = memorySwap
		}
		return nil
	}
}

// WithUlimits adds a user resource limit (ulimit) to the container's host configuration.
//
// Ulimits define resource constraints for processes running inside the container,
// such as the maximum number of open files ("nofile") or processes ("nproc").
//
// Parameters:
//   - name: the name of the resource to limit (e.g., "nofile", "nproc").
//   - soft: the soft limit, which is the value enforced for running processes.
//   - hard: the hard limit, which is the maximum value to which the soft limit can be raised.
//
// Returns an error if:
//   - the name is empty,
//   - any limit is negative,
//   - or the soft limit is greater than the hard limit.

func WithUlimits(name string, soft, hard int64) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if name == "" {
			return errdefs.NewHostConfigError("ulimit", "ulimit name cannot be empty")
		}
		if soft < 0 || hard < 0 {
			return errdefs.NewHostConfigError("ulimit", "ulimit values must be non-negative")
		}
		if soft > hard {
			return errdefs.NewHostConfigError("ulimit", fmt.Sprintf("soft limit (%d) cannot be greater than hard limit (%d)", soft, hard))
		}
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
// Note: use to run a custom init inside the container, if null, use the daemon's configured settings
func WithInit() create.SetHostConfig {
	init := true
	return func(opt *container.HostConfig) error {
		opt.Init = &init
		return nil
	}
}

// WithCPURealtimePeriod sets the CPU real-time period in microseconds.
// This controls the scheduling period for real-time tasks in the container.
//
// Accepts either:
//   - an int value representing microseconds directly (e.g. 100000 for 100ms), or
//   - a duration string (e.g. "1ms", "1s", "750ms").
//
// Valid values range from 1,000 to 1,000,000 microseconds (1ms to 1s), inclusive.
func WithCPURealtimePeriod[T int | string](period T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		switch v := any(period).(type) {
		case int:
			opt.CPURealtimePeriod = int64(v)
		case string:
			period, err := time.ParseDuration(v)
			if err != nil {
				return errdefs.NewHostConfigError("cpu_realtime_period", err.Error())
			}
			micro := period.Microseconds()
			if micro < 1000 || micro > 1_000_000 {
				return errdefs.NewHostConfigError("cpu_realtime_period", "must be between 1000 and 1,000,000 microseconds")
			}

			opt.CPURealtimePeriod = int64(period.Microseconds())
		}
		return nil
	}
}

// WithCPURealtimeRuntime sets the CPU real-time runtime in microseconds.
// A value of -1 disables the limit (default).
// parameters:
//   - runtime: the CPU real-time runtime in microseconds
func WithCPURealtimeRuntime(runtime int64) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if runtime < -1 {
			return errdefs.NewHostConfigError("cpu_realtime_runtime", "runtime must be -1 or >= 0")
		}
		opt.CPURealtimeRuntime = runtime
		return nil
	}
}

// WithCpusetMems sets the memory nodes in which execution is allowed.
// Only effective on NUMA systems.
// parameters:
//   - mems: the memory nodes in which execution is allowed
func WithCpusetMems(mems string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if mems == "" {
			return errdefs.NewHostConfigError("cpuset_mems", "mems is required")
		}
		if err := validateCpuset(mems); err != nil {
			return errdefs.NewHostConfigError("cpuset_mems", err.Error())
		}
		opt.CpusetMems = mems
		return nil
	}
}
func validateCpuset(mems string) error {
	ranges := strings.Split(mems, ",")
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			return fmt.Errorf("empty cpuset range")
		}
		parts := strings.Split(r, "-")
		if len(parts) == 1 {
			if _, err := strconv.Atoi(parts[0]); err != nil {
				return fmt.Errorf("invalid cpuset value: %s", r)
			}
		} else if len(parts) == 2 {
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil || start > end {
				return fmt.Errorf("invalid cpuset range: %s", r)
			}
		} else {
			return fmt.Errorf("invalid cpuset format: %s", r)
		}
	}
	return nil
}

// WithMemorySwappiness tunes container memory swappiness (0 to 100).
// - A value of 0 turns off anonymous page swapping.
// - A value of 100 sets the host's swappiness value.
// - Values between 0 and 100 modify the swappiness level accordingly.
// parameters:
//   - swappiness: the swappiness level
func WithMemorySwappiness(swappiness int64) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if swappiness < 0 || swappiness > 100 {
			return errdefs.NewHostConfigError("memory_swappiness", "swappiness must be between 0 and 100")
		}

		opt.MemorySwappiness = &swappiness
		return nil
	}
}

// WithKernelMemory sets the kernel memory limit in bytes.
// This is the hard limit for kernel memory that cannot be swapped out.
// parameters:
//   - memory: the kernel memory limit in bytes
//
// note: the kernel memory limit can be specified as an integer in bytes or as a string with a unit (e.g. "100M", "1G")
func WithKernelMemory[T int | string](memory T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		switch v := any(memory).(type) {
		case int:
			opt.KernelMemory = int64(v)
		case string:
			kernelMemory, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("kernel_memory", err.Error())
			}
			opt.KernelMemory = kernelMemory
		}
		return nil
	}
}

// WithPidsLimit sets the container's PIDs limit.
// parameters:
//   - limit: the PIDs limit
func WithPidsLimit(limit int64) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if limit < -1 || limit == 0 {
			return errdefs.NewHostConfigError("pids_limit", "limit must be -1 (unlimited) or a positive integer")
		}
		opt.PidsLimit = &limit
		return nil
	}
}

// WithBlkioWeight sets the block IO weight (relative weight) for the container.
// Weight is a value between 10 and 1000.
// parameters:
//   - weight: the block IO weight
func WithBlkioWeight(weight uint16) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if weight < 10 || weight > 1000 {
			return errdefs.NewHostConfigError("blkio_weight", "weight must be between 10 and 1000")
		}
		opt.BlkioWeight = weight
		return nil
	}
}

// WithBlkioDeviceReadBps appends a block IO read bandwidth throttle limit
// for a specific device to the container's host configuration.
// It limits the read rate (in bytes per second) on the specified device within the container.
//
// Parameters:
//   - devicePath: the device path (e.g., "/dev/sda")
//   - rate: the maximum read rate limit, either as an int (bytes per second)
//     or a human-readable string with units (e.g., "10MiB", "500KiB").
//
// Returns an error if the string rate cannot be parsed.
func WithBlkioDeviceReadBps[T int | string](devicePath string, rate T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.BlkioDeviceReadBps == nil {
			opt.BlkioDeviceReadBps = make([]*blkiodev.ThrottleDevice, 0)
		}
		if devicePath == "" {
			return errdefs.NewHostConfigError("blkio_device_read_bps", "device path cannot be empty")
		}

		var rateVal uint64
		switch v := any(rate).(type) {
		case int:
			rateVal = uint64(v)
		case string:
			parsedBytes, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("blkio_device_read_bps", err.Error())
			}
			rateVal = uint64(parsedBytes)
		}

		opt.BlkioDeviceReadBps = append(opt.BlkioDeviceReadBps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rateVal,
		})
		return nil
	}
}

// WithBlkioDeviceWriteBps appends a block IO write bandwidth throttle limit
// for a specific device to the container's host configuration.
// It limits the write rate (in bytes per second) on the specified device within the container.
//
// Parameters:
//   - devicePath: the device path (e.g., "/dev/sda")
//   - rate: the maximum write rate limit, either as an int (bytes per second)
//     or a human-readable string with units (e.g., "10MiB", "500KiB").
//
// Returns an error if the string rate cannot be parsed.
func WithBlkioDeviceWriteBps[T int | string](devicePath string, rate T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.BlkioDeviceWriteBps == nil {
			opt.BlkioDeviceWriteBps = make([]*blkiodev.ThrottleDevice, 0)
		}
		if devicePath == "" {
			return errdefs.NewHostConfigError("blkio_device_write_bps", "device path cannot be empty")
		}

		var rateVal uint64
		switch v := any(rate).(type) {
		case int:
			rateVal = uint64(v)
		case string:
			parsedBytes, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("blkio_device_write_bps", err.Error())
			}
			rateVal = uint64(parsedBytes)
		}

		opt.BlkioDeviceWriteBps = append(opt.BlkioDeviceWriteBps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rateVal,
		})
		return nil
	}
}

// WithBlkioDeviceReadIOps appends a block IO read operations per second (IOPS) limit
// for a specific device to the container's host configuration.
//
// Parameters:
//   - devicePath: the device path (e.g., "/dev/sda")
//   - rate: the maximum read IOPS limit, either as an int or a string with units
//     (e.g., "10MiB", "1KiB").
//
// Returns an error if the string rate cannot be parsed.
func WithBlkioDeviceReadIOps[T int | string](devicePath string, rate T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.BlkioDeviceReadIOps == nil {
			opt.BlkioDeviceReadIOps = make([]*blkiodev.ThrottleDevice, 0)
		}
		if devicePath == "" {
			return errdefs.NewHostConfigError("blkio_device_write_iops", "device path cannot be empty")
		}

		var rateVal uint64
		switch v := any(rate).(type) {
		case int:
			rateVal = uint64(v)
		case string:
			parsedBytes, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("blkio_device_read_iops", err.Error())
			}
			rateVal = uint64(parsedBytes)
		}

		opt.BlkioDeviceReadIOps = append(opt.BlkioDeviceReadIOps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rateVal,
		})
		return nil
	}
}

// WithBlkioDeviceWriteIOps appends a block IO write operations per second (IOPS) limit
// for a specific device to the container's host configuration.
//
// Parameters:
//   - devicePath: the device path (e.g., "/dev/sda")
//   - rate: the maximum write IOPS limit, either as an int or a string with units
//     (e.g., "10MiB", "1KiB").
//
// Returns an error if the string rate cannot be parsed.
func WithBlkioDeviceWriteIOps[T int | string](devicePath string, rate T) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.BlkioDeviceWriteIOps == nil {
			opt.BlkioDeviceWriteIOps = make([]*blkiodev.ThrottleDevice, 0)
		}
		if devicePath == "" {
			return errdefs.NewHostConfigError("blkio_device_write_iops", "device path cannot be empty")
		}

		var rateVal uint64
		switch v := any(rate).(type) {
		case int:
			rateVal = uint64(v)
		case string:
			parsedBytes, err := units.RAMInBytes(v)
			if err != nil {
				return errdefs.NewHostConfigError("blkio_device_write_iops", err.Error())
			}
			rateVal = uint64(parsedBytes)
		}

		opt.BlkioDeviceWriteIOps = append(opt.BlkioDeviceWriteIOps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rateVal,
		})
		return nil
	}
}

// WithSysctls adds or updates a sysctl key-value pair in the container's host configuration.
// Sysctls allow tuning of kernel parameters inside the container.
//
// Parameters:
//   - key: the sysctl parameter name (e.g., "net.ipv4.ip_forward")
//   - value: the sysctl parameter value (e.g., "1")
//
// Note: The effectiveness of sysctls depends on the container runtime and host kernel support.
func WithSysctls(key, value string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.Sysctls == nil {
			opt.Sysctls = make(map[string]string)
		}
		if key == "" {
			return errdefs.NewHostConfigError("sysctls", "key cannot be empty")
		}

		opt.Sysctls[key] = value
		return nil
	}
}

// WithDeviceCgroupRules appends device cgroup rules to the container's host configuration.
// Device cgroup rules control access to devices inside the container.
//
// Parameters:
//   - rules: one or more device cgroup rule strings (e.g., "c 1:3 rwm").
//
// Note: Rules must follow the device cgroup format accepted by the Linux kernel.
func WithDeviceCgroupRules(rules ...string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.DeviceCgroupRules == nil {
			opt.DeviceCgroupRules = make([]string, 0, len(rules))
		}
		opt.DeviceCgroupRules = append(opt.DeviceCgroupRules, rules...)
		return nil
	}
}

// WithCgroupParent sets the cgroup parent for the container in the host configuration.
// This determines the parent cgroup under which the container's cgroup will be created.
//
// Parameters:
//   - parent: the cgroup parent path (e.g., "docker", "system.slice")
func WithCgroupParent(parent string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.CgroupParent = parent
		return nil
	}
}

// WithDeviceRequest adds a device request to the container's host configuration.
// Device requests specify special device access requirements, such as GPUs.
//
// Parameters:
//   - driver: the device driver name (e.g., "nvidia")
//   - count: the number of devices to request (use -1 for all available)
//   - deviceIDs: specific device IDs to request (empty for any)
//   - capabilities: list of capability sets required (e.g., [][]string{{"gpu"}, {"compute"}})
//
// Note: Device requests are commonly used for GPU access and require appropriate drivers.
func WithDeviceRequest(driver string, count int, deviceIDs []string, capabilities [][]string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
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

// WithLogDriver sets the log driver and its options for the container.
//
// Parameters:
//   - driver: the log driver to use (e.g., "json-file", "syslog", "fluentd")
//   - options: a map of driver-specific options to configure the log driver
//
// Note: The supported log drivers and options depend on the container runtime and host configuration.
func WithLogDriver(driver string, options map[string]string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.LogConfig = container.LogConfig{
			Type:   driver,
			Config: options,
		}
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the host config
// and append the error to the host config error collection
func Fail(err error) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		return errdefs.NewHostConfigError("host_config", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the host config
// and append the error to the host config error collection
func Failf(stringFormat string, args ...any) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		return errdefs.NewHostConfigError("host_config", fmt.Sprintf(stringFormat, args...))
	}
}
