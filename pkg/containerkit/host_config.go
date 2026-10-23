package containerkit

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
)

// MemoryLimit sets the container memory limit.
func (c *Container) MemoryLimit(memory ByteSize) *Container {
	c.Config.Host.Memory = int64(memory)
	return c
}

// RestartAlways sets a restart policy that ensures the container is always restarted upon exit.
// This policy restarts the container indefinitely, ignoring any retry limits.
func (c *Container) RestartAlways() *Container { return c.RestartPolicy(RestartPolicyAlways, 0) }

// AutoRemove sets the container to be automatically removed when it exits.
func (c *Container) AutoRemove() *Container {
	c.Config.Host.AutoRemove = true
	return c
}

// PortBinding publishes one container port on the host.
func (c *Container) PortBinding(binding PortBinding) *Container {
	if binding.ContainerPort == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("port_bindings", "empty container port"))
		return c
	}
	if binding.HostPort == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("port_bindings", "empty host port"))
		return c
	}
	if binding.HostIP == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("port_bindings", "empty host IP"))
		return c
	}
	if binding.Protocol == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("port_bindings", "empty protocol"))
		return c
	}

	cPort, err := nat.NewPort(string(binding.Protocol), binding.ContainerPort)
	if err != nil {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("port_bindings", err.Error()))
		return c
	}
	parsedHost, err := nat.NewPort(string(binding.Protocol), binding.HostPort)
	if err != nil {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("port_bindings", err.Error()))
		return c
	}

	if c.Config.Host.PortBindings == nil {
		c.Config.Host.PortBindings = make(nat.PortMap)
	}
	c.Config.Host.PortBindings[cPort] = []nat.PortBinding{
		{
			HostIP:   binding.HostIP,
			HostPort: parsedHost.Port(),
		},
	}
	return c
}

// DNSLookups appends a DNS server to the host configuration for the container.
// parameters:
//   - dns: the DNS server to add
func (c *Container) DNSLookups(dns ...string) *Container {
	if c.Config.Host.DNS == nil {
		c.Config.Host.DNS = make([]string, 0, len(dns))
	}
	c.Config.Host.DNS = append(c.Config.Host.DNS, dns...)
	return c
}

// DNSOptions appends a DNS option to the host configuration for the container.
// parameters:
//   - dnsOption: the DNS option to add
func (c *Container) DNSOptions(dnsOption ...string) *Container {
	if c.Config.Host.DNSOptions == nil {
		c.Config.Host.DNSOptions = make(strslice.StrSlice, 0, len(dnsOption))
	}
	c.Config.Host.DNSOptions = append(c.Config.Host.DNSOptions, dnsOption...)
	return c
}

// DNSSearches appends a DNS search domain to the host configuration for the container.
// parameters:
//   - search: the DNS search domain to add
func (c *Container) DNSSearches(search ...string) *Container {
	if c.Config.Host.DNSSearch == nil {
		c.Config.Host.DNSSearch = make(strslice.StrSlice, 0, len(search))
	}
	c.Config.Host.DNSSearch = append(c.Config.Host.DNSSearch, search...)
	return c
}

// ExtraHost adds a hostname and address to the container hosts file.
func (c *Container) ExtraHost(hostname, ip string) *Container {
	if hostname == "" || ip == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("extra_hosts", "hostname and ip are required"))
		return c
	}
	if c.Config.Host.ExtraHosts == nil {
		c.Config.Host.ExtraHosts = make([]string, 0, 1)
	}
	c.Config.Host.ExtraHosts = append(c.Config.Host.ExtraHosts, hostname+":"+ip)
	return c
}

// AddedGroups appends supplementary groups to the host configuration for the container.
// parameters:
//   - group: the group to add
func (c *Container) AddedGroups(group ...string) *Container {
	if c.Config.Host.GroupAdd == nil {
		c.Config.Host.GroupAdd = make(strslice.StrSlice, 0, len(group))
	}
	c.Config.Host.GroupAdd = append(c.Config.Host.GroupAdd, group...)
	return c
}

// VolumeBinds parses short mount strings into Mount values.
// Format is source:target[:mode]. Mode is optional and defaults to read-write.
// A source with no path separator is a named volume. Absolute, relative, and Windows paths are binds.
// Mode tokens: ro, rw, readonly, z, Z, consistent, cached, delegated, rshared, shared, rslave, slave, rprivate, private.
func (c *Container) VolumeBinds(binds ...string) *Container {
	for i, spec := range binds {
		parsed, err := parseVolumeBind(spec)
		if err != nil {
			c.Errors = append(c.Errors, errdefs.NewHostConfigError("volume_binds", fmt.Sprintf("binds[%d]: %s", i, err)))
			continue
		}
		if c.mountTargetTaken(parsed.Target) {
			c.Errors = append(c.Errors, errdefs.NewHostConfigError("volume_binds", fmt.Sprintf("binds[%d]: duplicate target path '%s'", i, parsed.Target)))
			continue
		}
		c.Mount(parsed)
	}
	return c
}

func (c *Container) mountTargetTaken(target string) bool {
	if c == nil || c.Config == nil || c.Config.Host == nil {
		return false
	}
	for _, existing := range c.Config.Host.Mounts {
		if existing.Target == target {
			return true
		}
	}
	return false
}

func parseVolumeBind(spec string) (Mount, error) {
	source, target, mode, err := splitBindSpec(spec)
	if err != nil {
		return Mount{}, err
	}
	parsed := Mount{Source: source, Target: target, Type: MountBind}
	if !isBindSource(source) {
		parsed.Type = MountVolume
	}
	if err := applyBindMode(&parsed, mode); err != nil {
		return Mount{}, err
	}
	return parsed, nil
}

func isBindSource(source string) bool {
	if strings.HasPrefix(source, "/") || strings.HasPrefix(source, ".") || strings.HasPrefix(source, "~") {
		return true
	}
	if strings.Contains(source, "/") || strings.Contains(source, `\`) {
		return true
	}
	return len(source) >= 2 && source[1] == ':'
}

func splitBindSpec(spec string) (source, target, mode string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", "", "", errors.New("empty mount specification")
	}
	idx := strings.Index(spec, ":/")
	if idx == -1 {
		return "", "", "", fmt.Errorf("invalid format '%s' (must be 'source:target[:mode]')", spec)
	}
	source = strings.TrimSpace(spec[:idx])
	rest := spec[idx+2:]
	if source == "" {
		return "", "", "", errors.New("empty source path")
	}
	if colon := strings.LastIndex(rest, ":"); colon != -1 {
		maybe := strings.TrimSpace(rest[colon+1:])
		rest = strings.TrimSpace(rest[:colon])
		if rest == "" {
			return "", "", "", errors.New("empty target path")
		}
		if !validBindMode(maybe) {
			return "", "", "", fmt.Errorf("invalid mode '%s'", maybe)
		}
		return source, "/" + rest, maybe, nil
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", "", "", errors.New("empty target path")
	}
	return source, "/" + rest, "", nil
}

func applyBindMode(parsed *Mount, mode string) error {
	if mode == "" {
		return nil
	}
	var (
		readOnly    *bool
		selinux     string
		consistency MountConsistency
		propagation MountPropagation
	)
	for _, part := range strings.Split(mode, ",") {
		part = strings.TrimSpace(part)
		switch part {
		case "ro", "readonly":
			if readOnly != nil && !*readOnly {
				return fmt.Errorf("invalid mode '%s'", mode)
			}
			value := true
			readOnly = &value
		case "rw":
			if readOnly != nil && *readOnly {
				return fmt.Errorf("invalid mode '%s'", mode)
			}
			value := false
			readOnly = &value
		case "z", "Z":
			if selinux != "" && selinux != part {
				return fmt.Errorf("invalid mode '%s'", mode)
			}
			selinux = part
		case "consistent", "cached", "delegated":
			if consistency != "" && consistency != MountConsistency(part) {
				return fmt.Errorf("invalid mode '%s'", mode)
			}
			consistency = MountConsistency(part)
		default:
			if propagation != "" && propagation != MountPropagation(part) {
				return fmt.Errorf("invalid mode '%s'", mode)
			}
			propagation = MountPropagation(part)
		}
	}
	if readOnly != nil {
		parsed.ReadOnly = *readOnly
	}
	parsed.Consistency = consistency
	if selinux != "" || propagation != "" {
		parsed.Bind = &BindOptions{SELinux: selinux, Propagation: propagation}
	}
	return nil
}

// validBindModeOptions is the set of bind mount mode tokens accepted by Docker/Compose.
// - ro, rw, readonly: read-only / read-write (readonly is alias for ro)
// - z, Z: SELinux labeling (shared / private)
// - consistent, cached, delegated: consistency (e.g. Docker Desktop macOS)
// - rshared, shared, rslave, slave, rprivate, private: bind propagation
var validBindModeOptions = map[string]struct{}{
	"ro": {}, "rw": {}, "readonly": {},
	"z": {}, "Z": {},
	"consistent": {}, "cached": {}, "delegated": {},
	"rshared": {}, "shared": {}, "rslave": {}, "slave": {}, "rprivate": {}, "private": {},
}

func validBindMode(mode string) bool {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return false
	}
	for _, part := range strings.Split(mode, ",") {
		part = strings.TrimSpace(part)
		if _, ok := validBindModeOptions[part]; !ok {
			return false
		}
	}
	return true
}

// ShmSize sets the size of /dev/shm.
func (c *Container) ShmSize(size ByteSize) *Container {
	c.Config.Host.ShmSize = int64(size)
	return c
}

func (c *Container) Runtime(runtime string) *Container {
	c.Config.Host.Runtime = runtime
	return c
}

// ConsoleSize sets the container console height and width.
func (c *Container) ConsoleSize(size ConsoleSize) *Container {
	c.Config.Host.ConsoleSize = [2]uint{size.Height, size.Width}
	return c
}

// CPUCount sets the number of CPUs allocated to the container in the host configuration.
//
// Parameters:
//   - count: the number of CPUs to allocate to the container.
//
// Returns an error if count is less than 1.
func (c *Container) CPUCount(count int64) *Container {
	if count < 1 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpu_count", "CPU count must be at least 1"))
		return c
	}
	c.Config.Host.CPUCount = count
	return c
}

// ReadonlyPaths appends a list of paths to be marked as read-only inside the container.
//
// These paths are mounted as read-only from the host and are inaccessible for writing
// by any process in the container.
//
// Parameters:
//   - paths: one or more absolute paths to mark as read-only
//
// Returns an error if any path is empty or not absolute.
func (c *Container) ReadonlyPaths(paths ...string) *Container {
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			c.Errors = append(c.Errors, errdefs.NewHostConfigError("readonly_paths", "path cannot be empty"))
			return c
		}
		if !strings.HasPrefix(path, "/") {
			c.Errors = append(c.Errors, errdefs.NewHostConfigError("readonly_paths", fmt.Sprintf("path must be absolute: %q", path)))
			return c
		}
	}
	if c.Config.Host.ReadonlyPaths == nil {
		c.Config.Host.ReadonlyPaths = make([]string, 0, len(paths))
	}
	c.Config.Host.ReadonlyPaths = append(c.Config.Host.ReadonlyPaths, paths...)
	return c
}

// MaskedPaths appends a list of paths to be masked inside the container.
//
// Masked paths are mounted as read-only and inaccessible from within the container,
// overriding the default masked paths (e.g., "/proc/kcore", "/proc/latency_stats").
//
// Parameters:
//   - paths: one or more absolute paths to mask
//
// Returns an error if any path is empty or not absolute.
func (c *Container) MaskedPaths(paths ...string) *Container {
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			c.Errors = append(c.Errors, errdefs.NewHostConfigError("masked_paths", "path cannot be empty"))
			return c
		}
		if !strings.HasPrefix(path, "/") {
			c.Errors = append(c.Errors, errdefs.NewHostConfigError("masked_paths", fmt.Sprintf("path must be absolute: %q", path)))
			return c
		}
	}
	if c.Config.Host.MaskedPaths == nil {
		c.Config.Host.MaskedPaths = make([]string, 0, len(paths))
	}
	c.Config.Host.MaskedPaths = append(c.Config.Host.MaskedPaths, paths...)
	return c
}

// VolumeDriver sets the volume driver for the container in the host configuration.
//
// The volume driver specifies the plugin or mechanism used to manage volumes mounted into the container.
// This option is typically only used with the Docker API directly, not with Docker Compose.
//
// Parameters:
//   - driver: the name of the volume driver to use.
//
// Returns an error if the driver name is empty or contains only whitespace.
func (c *Container) VolumeDriver(driver string) *Container {
	driver = strings.TrimSpace(driver)
	if driver == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("volume_driver", "volume driver cannot be empty"))
		return c
	}
	c.Config.Host.VolumeDriver = driver
	return c
}

// VolumesFrom inherits volumes from another container.
// readOnly maps to the :ro suffix. A writable inherit stores the container name alone.
func (c *Container) VolumesFrom(name string, readOnly bool) *Container {
	if strings.TrimSpace(name) == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("volumes_from", "container name is empty"))
		return c
	}
	spec := name
	if readOnly {
		spec += ":ro"
	}
	if c.Config.Host.VolumesFrom == nil {
		c.Config.Host.VolumesFrom = make([]string, 0, 1)
	}
	c.Config.Host.VolumesFrom = append(c.Config.Host.VolumesFrom, spec)
	return c
}

// OomScoreAdj sets the Out-Of-Memory (OOM) score adjustment for the container.
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
func (c *Container) OomScoreAdj(score int) *Container {
	if score < -1000 || score > 1000 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("oom_score_adj", "value must be between -1000 and 1000"))
		return c
	}
	c.Config.Host.OomScoreAdj = score
	return c
}

// OomKillDisable sets the OOM kill disable flag for the container in the host configuration.
// parameters:
//   - oomKillDisable: the OOM kill disable flag to use
func (c *Container) OomKillDisable() *Container {
	oomKillDisable := true
	c.Config.Host.OomKillDisable = &oomKillDisable
	return c
}

func (c *Container) PublishAllPorts() *Container {
	c.Config.Host.PublishAllPorts = true
	return c
}

func (c *Container) ReadOnlyRootfs() *Container {
	c.Config.Host.ReadonlyRootfs = true
	return c
}

// SecurityOpts appends security options to the container's host configuration.
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
func (c *Container) SecurityOpts(opts ...string) *Container {
	if c.Config.Host.SecurityOpt == nil {
		c.Config.Host.SecurityOpt = make([]string, 0, len(opts))
	}
	c.Config.Host.SecurityOpt = append(c.Config.Host.SecurityOpt, opts...)
	return c
}

// StorageOpt appends a storage driver option for the container's host configuration.
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
func (c *Container) StorageOpt(key, value string) *Container {
	key = strings.TrimSpace(key)
	if key == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("storage_opt", "storage option key cannot be empty"))
		return c
	}
	if c.Config.Host.StorageOpt == nil {
		c.Config.Host.StorageOpt = make(map[string]string)
	}
	c.Config.Host.StorageOpt[key] = value
	return c
}

// Tmpfs appends to a map of tmpfs (mounts) used for the container
// parameters:
//   - key: the key of the tmpfs option
//   - value: the value of the tmpfs option
func (c *Container) Tmpfs(key, value string) *Container {
	if c.Config.Host.Tmpfs == nil {
		c.Config.Host.Tmpfs = make(map[string]string)
	}
	c.Config.Host.Tmpfs[key] = value
	return c
}

func (c *Container) Privileged() *Container {
	c.Config.Host.Privileged = true
	return c
}

// AddedDevice maps a host device into the container.
func (c *Container) AddedDevice(device Device) *Container {
	if strings.TrimSpace(device.HostPath) == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("devices", "host device path cannot be empty"))
		return c
	}
	if strings.TrimSpace(device.ContainerPath) == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("devices", "container device path cannot be empty"))
		return c
	}
	permissions := devicePermissions(device)
	if permissions == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("devices", "device permissions are required"))
		return c
	}

	if c.Config.Host.Devices == nil {
		c.Config.Host.Devices = make([]container.DeviceMapping, 0)
	}
	c.Config.Host.Devices = append(c.Config.Host.Devices, container.DeviceMapping{
		PathOnHost:        device.HostPath,
		PathInContainer:   device.ContainerPath,
		CgroupPermissions: permissions,
	})
	return c
}

func devicePermissions(device Device) string {
	var permissions string
	if device.Read {
		permissions += "r"
	}
	if device.Write {
		permissions += "w"
	}
	if device.Mknod {
		permissions += "m"
	}
	return permissions
}

// ContainerIDFile sets the path to a file where the container ID will be written after creation.
//
// After `client.ContainerCreate`, Docker will write the container's ID to this file.
// This is useful for external tooling or scripts that need to reference the container after it starts.
//
// Parameters:
//   - path: absolute or relative path to the container ID file.
//
// Returns an error if the path is empty or only whitespace.
func (c *Container) ContainerIDFile(path string) *Container {
	if strings.TrimSpace(path) == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("container_id_file", "path cannot be empty"))
		return c
	}
	c.Config.Host.ContainerIDFile = path
	return c
}

// CPUShares sets the CPU shares (relative weight) for the container.
//
// CPU shares define the relative CPU time available to the container compared to others.
// For example, 1024 is the default (normal priority), 512 is half the CPU weight, 2048 is double.
//
// Parameters:
//   - shares: the number of CPU shares (must be a positive integer).
//
// Returns an error if the value is less than or equal to zero.
func (c *Container) CPUShares(shares int64) *Container {
	if shares <= 0 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpu_shares", "CPU shares must be a positive integer"))
		return c
	}
	c.Config.Host.CPUShares = shares
	return c
}

// CPUPeriod sets the CFS period. Zero leaves it unset.
// A positive period must be from 1ms through 1s.
func (c *Container) CPUPeriod(period time.Duration) *Container {
	if period == 0 {
		c.Config.Host.CPUPeriod = 0
		return c
	}
	micros := period.Microseconds()
	if micros < 1000 || micros > 1_000_000 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpu_period", "CPU period must be between 1ms and 1s"))
		return c
	}
	c.Config.Host.CPUPeriod = micros
	return c
}

// CPUPercent sets the CPU percentage limit for the container.
//
// This option specifies the maximum amount of CPU the container can use as a percentage
// (e.g., 50 means the container can use up to 50% of one CPU core).
//
// Parameters:
//   - percent: the desired CPU limit as a percentage (0–100, or more for multi-core allocation).
//
// Returns an error if the percent is negative.
func (c *Container) CPUPercent(percent int64) *Container {
	if percent < 0 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpu_percent", "CPU percent must be non-negative"))
		return c
	}
	c.Config.Host.CPUPercent = percent
	return c
}

// CPUQuota sets the CFS quota. Zero means no quota.
func (c *Container) CPUQuota(quota time.Duration) *Container {
	if quota < 0 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpu_quota", "quota must be non-negative"))
		return c
	}
	c.Config.Host.CPUQuota = quota.Microseconds()
	return c
}

// CpusetCpus sets the CPUs in which execution is allowed.
func (c *Container) CpusetCpus(cpus Cpuset) *Container {
	if cpus.list == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpuset_cpus", "cpus is required"))
		return c
	}
	c.Config.Host.CpusetCpus = cpus.list
	return c
}

// MemoryReservation sets the memory soft limit.
func (c *Container) MemoryReservation(memory ByteSize) *Container {
	c.Config.Host.MemoryReservation = int64(memory)
	return c
}

// MemorySwap sets the total memory limit, including swap.
func (c *Container) MemorySwap(memorySwap ByteSize) *Container {
	c.Config.Host.MemorySwap = int64(memorySwap)
	return c
}

// Ulimits appends one resource limit.
func (c *Container) Ulimits(limit Ulimit) *Container {
	if limit.Name == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("ulimits", "ulimit name cannot be empty"))
		return c
	}
	if limit.Soft < 0 || limit.Hard < 0 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("ulimits", "ulimit values must be non-negative"))
		return c
	}
	if limit.Soft > limit.Hard {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("ulimits", fmt.Sprintf("soft limit (%d) cannot be greater than hard limit (%d)", limit.Soft, limit.Hard)))
		return c
	}
	if c.Config.Host.Ulimits == nil {
		c.Config.Host.Ulimits = make([]*container.Ulimit, 0)
	}
	c.Config.Host.Ulimits = append(c.Config.Host.Ulimits, &container.Ulimit{
		Name: limit.Name,
		Soft: limit.Soft,
		Hard: limit.Hard,
	})
	return c
}

// Init sets the init flag for the container
// parameters:
//   - init: the init flag to use
//
// Note: use to run a custom init inside the container, if null, use the daemon's configured settings
func (c *Container) Init() *Container {
	init := true
	c.Config.Host.Init = &init
	return c
}

// CPURealtimePeriod sets the real-time scheduling period. Zero leaves it unset.
// A positive period must be from 1ms through 1s.
func (c *Container) CPURealtimePeriod(period time.Duration) *Container {
	if period == 0 {
		c.Config.Host.CPURealtimePeriod = 0
		return c
	}
	micro := period.Microseconds()
	if micro < 1000 || micro > 1_000_000 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpu_realtime_period", "must be between 1ms and 1s"))
		return c
	}
	c.Config.Host.CPURealtimePeriod = micro
	return c
}

// CPURealtimeRuntime sets the CPU real-time runtime in microseconds.
// A value of -1 disables the limit (default).
// parameters:
//   - runtime: the CPU real-time runtime in microseconds
func (c *Container) CPURealtimeRuntime(runtime int64) *Container {
	if runtime < -1 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpu_realtime_runtime", "runtime must be -1 or >= 0"))
		return c
	}
	c.Config.Host.CPURealtimeRuntime = runtime
	return c
}

// CpusetMems sets the memory nodes in which execution is allowed.
// Only effective on NUMA systems.
func (c *Container) CpusetMems(mems Cpuset) *Container {
	if mems.list == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("cpuset_mems", "mems is required"))
		return c
	}
	c.Config.Host.CpusetMems = mems.list
	return c
}

// MemorySwappiness tunes container memory swappiness (0 to 100).
// - A value of 0 turns off anonymous page swapping.
// - A value of 100 sets the host's swappiness value.
// - Values between 0 and 100 modify the swappiness level accordingly.
// parameters:
//   - swappiness: the swappiness level
func (c *Container) MemorySwappiness(swappiness int64) *Container {
	if swappiness < 0 || swappiness > 100 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("memory_swappiness", "swappiness must be between 0 and 100"))
		return c
	}
	c.Config.Host.MemorySwappiness = &swappiness
	return c
}

// KernelMemory sets the kernel memory limit.
func (c *Container) KernelMemory(memory ByteSize) *Container {
	c.Config.Host.KernelMemory = int64(memory)
	return c
}

// PidsLimit sets the container's PIDs limit.
// parameters:
//   - limit: the PIDs limit
func (c *Container) PidsLimit(limit int64) *Container {
	if limit < -1 || limit == 0 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("pids_limit", "limit must be -1 (unlimited) or a positive integer"))
		return c
	}
	c.Config.Host.PidsLimit = &limit
	return c
}

// BlkioWeight sets the block IO weight (relative weight) for the container.
// Weight is a value between 10 and 1000.
// parameters:
//   - weight: the block IO weight
func (c *Container) BlkioWeight(weight uint16) *Container {
	if weight < 10 || weight > 1000 {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("blkio_weight", "weight must be between 10 and 1000"))
		return c
	}
	c.Config.Host.BlkioWeight = weight
	return c
}

// BlkioDeviceReadBps limits the read rate for one device.
func (c *Container) BlkioDeviceReadBps(devicePath string, rate ByteSize) *Container {
	if devicePath == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("blkio_device_read_bps", "device path cannot be empty"))
		return c
	}
	if c.Config.Host.BlkioDeviceReadBps == nil {
		c.Config.Host.BlkioDeviceReadBps = make([]*blkiodev.ThrottleDevice, 0)
	}
	c.Config.Host.BlkioDeviceReadBps = append(c.Config.Host.BlkioDeviceReadBps, &blkiodev.ThrottleDevice{
		Path: devicePath,
		Rate: uint64(rate),
	})
	return c
}

// BlkioDeviceWriteBps limits the write rate for one device.
func (c *Container) BlkioDeviceWriteBps(devicePath string, rate ByteSize) *Container {
	if devicePath == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("blkio_device_write_bps", "device path cannot be empty"))
		return c
	}
	if c.Config.Host.BlkioDeviceWriteBps == nil {
		c.Config.Host.BlkioDeviceWriteBps = make([]*blkiodev.ThrottleDevice, 0)
	}
	c.Config.Host.BlkioDeviceWriteBps = append(c.Config.Host.BlkioDeviceWriteBps, &blkiodev.ThrottleDevice{
		Path: devicePath,
		Rate: uint64(rate),
	})
	return c
}

// BlkioDeviceReadIOps appends a block IO read operations per second (IOPS) limit
// for a specific device to the container's host configuration.
//
// Parameters:
//   - devicePath: the device path (e.g., "/dev/sda")
//   - rate: the maximum read IOPS limit
func (c *Container) BlkioDeviceReadIOps(devicePath string, rate uint64) *Container {
	if c.Config.Host.BlkioDeviceReadIOps == nil {
		c.Config.Host.BlkioDeviceReadIOps = make([]*blkiodev.ThrottleDevice, 0)
	}
	if devicePath == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("blkio_device_write_iops", "device path cannot be empty"))
		return c
	}
	c.Config.Host.BlkioDeviceReadIOps = append(c.Config.Host.BlkioDeviceReadIOps, &blkiodev.ThrottleDevice{
		Path: devicePath,
		Rate: rate,
	})
	return c
}

// BlkioDeviceWriteIOps appends a block IO write operations per second (IOPS) limit
// for a specific device to the container's host configuration.
//
// Parameters:
//   - devicePath: the device path (e.g., "/dev/sda")
//   - rate: the maximum write IOPS limit
func (c *Container) BlkioDeviceWriteIOps(devicePath string, rate uint64) *Container {
	if c.Config.Host.BlkioDeviceWriteIOps == nil {
		c.Config.Host.BlkioDeviceWriteIOps = make([]*blkiodev.ThrottleDevice, 0)
	}
	if devicePath == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("blkio_device_write_iops", "device path cannot be empty"))
		return c
	}
	c.Config.Host.BlkioDeviceWriteIOps = append(c.Config.Host.BlkioDeviceWriteIOps, &blkiodev.ThrottleDevice{
		Path: devicePath,
		Rate: rate,
	})
	return c
}

// Sysctls adds or updates a sysctl key-value pair in the container's host configuration.
// Sysctls allow tuning of kernel parameters inside the container.
//
// Parameters:
//   - key: the sysctl parameter name (e.g., "net.ipv4.ip_forward")
//   - value: the sysctl parameter value (e.g., "1")
//
// Note: The effectiveness of sysctls depends on the container runtime and host kernel support.
func (c *Container) Sysctls(key, value string) *Container {
	if c.Config.Host.Sysctls == nil {
		c.Config.Host.Sysctls = make(map[string]string)
	}
	if key == "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("sysctls", "key cannot be empty"))
		return c
	}
	c.Config.Host.Sysctls[key] = value
	return c
}

// DeviceCgroupRules appends device cgroup rules to the container's host configuration.
// Device cgroup rules control access to devices inside the container.
//
// Parameters:
//   - rules: one or more device cgroup rule strings (e.g., "c 1:3 rwm").
//
// Note: Rules must follow the device cgroup format accepted by the Linux kernel.
func (c *Container) DeviceCgroupRules(rules ...string) *Container {
	if c.Config.Host.DeviceCgroupRules == nil {
		c.Config.Host.DeviceCgroupRules = make([]string, 0, len(rules))
	}
	c.Config.Host.DeviceCgroupRules = append(c.Config.Host.DeviceCgroupRules, rules...)
	return c
}

func (c *Container) CgroupParent(parent string) *Container {
	c.Config.Host.CgroupParent = parent
	return c
}

// DeviceRequest appends a device request, such as a GPU, to the container.
func (c *Container) DeviceRequest(req DeviceRequest) *Container {
	if c.Config.Host.DeviceRequests == nil {
		c.Config.Host.DeviceRequests = make([]container.DeviceRequest, 0)
	}
	var caps [][]string
	if len(req.Capabilities) > 0 {
		caps = [][]string{req.Capabilities}
	}
	c.Config.Host.DeviceRequests = append(c.Config.Host.DeviceRequests, container.DeviceRequest{
		Driver:       req.Driver,
		Count:        req.Count,
		DeviceIDs:    req.DeviceIDs,
		Capabilities: caps,
		Options:      req.Options,
	})
	return c
}

func (c *Container) LogDriver(driver string, options map[string]string) *Container {
	c.Config.Host.LogConfig = container.LogConfig{
		Type:   driver,
		Config: options,
	}
	return c
}
