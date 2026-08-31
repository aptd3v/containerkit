package containerkit

import (
	"os"

	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/cc"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/hc"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/pc"
)

func (c *Container) Env(key string, value string) *Container {
	return c.WithContainerConfig(cc.WithEnv(key, value))
}

func (c *Container) EnvMap(env map[string]string) *Container {
	return c.WithContainerConfig(cc.WithEnvMap(env))
}

func (c *Container) ExposedPort(protocol string, port string) *Container {
	return c.WithContainerConfig(cc.WithExposedPort(protocol, port))
}

func (c *Container) HostName(hostname string) *Container {
	return c.WithContainerConfig(cc.WithHostName(hostname))
}

func (c *Container) DomainName(domainname string) *Container {
	return c.WithContainerConfig(cc.WithDomainName(domainname))
}

func (c *Container) Image(image string) *Container {
	return c.WithContainerConfig(cc.WithImage(image))
}

func (c *Container) Imagef(stringFormat string, args ...any) *Container {
	return c.WithContainerConfig(cc.WithImagef(stringFormat, args...))
}

func (c *Container) Command(cmd ...string) *Container {
	return c.WithContainerConfig(cc.WithCommand(cmd...))
}

func (c *Container) User(user string) *Container {
	return c.WithContainerConfig(cc.WithUser(user))
}

func (c *Container) AttachedStdin() *Container {
	return c.WithContainerConfig(cc.WithAttachedStdin())
}

func (c *Container) AttachedStdout() *Container {
	return c.WithContainerConfig(cc.WithAttachedStdout())
}

func (c *Container) AttachedStderr() *Container {
	return c.WithContainerConfig(cc.WithAttachedStderr())
}

func (c *Container) Tty() *Container {
	return c.WithContainerConfig(cc.WithTty())
}

func (c *Container) StdinOpen() *Container {
	return c.WithContainerConfig(cc.WithStdinOpen())
}

func (c *Container) StdinOnce() *Container {
	return c.WithContainerConfig(cc.WithStdinOnce())
}

func (c *Container) EscapedArgs() *Container {
	return c.WithContainerConfig(cc.WithEscapedArgs())
}

func (c *Container) Volume(volume string) *Container {
	return c.WithContainerConfig(cc.WithVolume(volume))
}

func (c *Container) WorkingDir(dir string) *Container {
	return c.WithContainerConfig(cc.WithWorkingDir(dir))
}

func (c *Container) DisabledNetwork() *Container {
	return c.WithContainerConfig(cc.WithDisabledNetwork())
}

func (c *Container) OnBuild(args ...string) *Container {
	return c.WithContainerConfig(cc.WithOnBuild(args...))
}

func (c *Container) Label(label, value string) *Container {
	return c.WithContainerConfig(cc.WithLabel(label, value))
}

func (c *Container) StopSignal(signal string) *Container {
	return c.WithContainerConfig(cc.WithStopSignal(signal))
}

func (c *Container) Entrypoint(entrypoint ...string) *Container {
	return c.WithContainerConfig(cc.WithEntrypoint(entrypoint...))
}

func (c *Container) Shell(shell ...string) *Container {
	return c.WithContainerConfig(cc.WithShell(shell...))
}

func (c *Container) StopTimeout(timeout int) *Container {
	return c.WithContainerConfig(cc.WithStopTimeout(timeout))
}

func (c *Container) MemoryLimit(memory int) *Container {
	return c.WithHostConfig(hc.WithMemoryLimit(memory))
}

func (c *Container) MemoryLimitString(memory string) *Container {
	return c.WithHostConfig(hc.WithMemoryLimit(memory))
}

func (c *Container) RestartAlways() *Container {
	return c.WithHostConfig(hc.WithRestartAlways())
}

func (c *Container) AutoRemove() *Container {
	return c.WithHostConfig(hc.WithAutoRemove())
}

func (c *Container) PortBindings(protocol, hostIP, hostPort, containerPort string) *Container {
	return c.WithHostConfig(hc.WithPortBindings(protocol, hostIP, hostPort, containerPort))
}

func (c *Container) DNSLookups(dns ...string) *Container {
	return c.WithHostConfig(hc.WithDNSLookups(dns...))
}

func (c *Container) DNSOptions(dnsOption ...string) *Container {
	return c.WithHostConfig(hc.WithDNSOptions(dnsOption...))
}

func (c *Container) DNSSearches(search ...string) *Container {
	return c.WithHostConfig(hc.WithDNSSearches(search...))
}

func (c *Container) ExtraHost(extraHosts ...string) *Container {
	return c.WithHostConfig(hc.WithExtraHost(extraHosts...))
}

func (c *Container) AddedGroups(group ...string) *Container {
	return c.WithHostConfig(hc.WithAddedGroups(group...))
}

func (c *Container) VolumeBinds(binds ...string) *Container {
	return c.WithHostConfig(hc.WithVolumeBinds(binds...))
}

func (c *Container) UTSMode(mode string) *Container {
	return c.WithHostConfig(hc.WithUTSMode(mode))
}

func (c *Container) UserNSMode(mode string) *Container {
	return c.WithHostConfig(hc.WithUserNSMode(mode))
}

func (c *Container) ShmSize(size int) *Container {
	return c.WithHostConfig(hc.WithShmSize(size))
}

func (c *Container) ShmSizeString(size string) *Container {
	return c.WithHostConfig(hc.WithShmSize(size))
}

func (c *Container) Runtime(runtime string) *Container {
	return c.WithHostConfig(hc.WithRuntime(runtime))
}

func (c *Container) ConsoleSize(height uint, width uint) *Container {
	return c.WithHostConfig(hc.WithConsoleSize(height, width))
}

func (c *Container) Isolation(isolation string) *Container {
	return c.WithHostConfig(hc.WithIsolation(isolation))
}

func (c *Container) CPUCount(count int64) *Container {
	return c.WithHostConfig(hc.WithCPUCount(count))
}

func (c *Container) ReadonlyPaths(paths ...string) *Container {
	return c.WithHostConfig(hc.WithReadonlyPaths(paths...))
}

func (c *Container) MaskedPaths(paths ...string) *Container {
	return c.WithHostConfig(hc.WithMaskedPaths(paths...))
}

func (c *Container) NetworkMode(mode string) *Container {
	return c.WithHostConfig(hc.WithNetworkMode(mode))
}

func (c *Container) VolumeDriver(driver string) *Container {
	return c.WithHostConfig(hc.WithVolumeDriver(driver))
}

func (c *Container) VolumesFrom(from string) *Container {
	return c.WithHostConfig(hc.WithVolumesFrom(from))
}

func (c *Container) IpcMode(mode string) *Container {
	return c.WithHostConfig(hc.WithIpcMode(mode))
}

func (c *Container) Cgroup(cgroup string) *Container {
	return c.WithHostConfig(hc.WithCgroup(cgroup))
}

func (c *Container) OomScoreAdj(score int) *Container {
	return c.WithHostConfig(hc.WithOomScoreAdj(score))
}

func (c *Container) OomKillDisable() *Container {
	return c.WithHostConfig(hc.WithOomKillDisable())
}

func (c *Container) PidMode(mode string) *Container {
	return c.WithHostConfig(hc.WithPidMode(mode))
}

func (c *Container) PublishAllPorts() *Container {
	return c.WithHostConfig(hc.WithPublishAllPorts())
}

func (c *Container) ReadOnlyRootfs() *Container {
	return c.WithHostConfig(hc.WithReadOnlyRootfs())
}

func (c *Container) SecurityOpts(opts ...string) *Container {
	return c.WithHostConfig(hc.WithSecurityOpts(opts...))
}

func (c *Container) StorageOpt(key, value string) *Container {
	return c.WithHostConfig(hc.WithStorageOpt(key, value))
}

func (c *Container) Tmpfs(key, value string) *Container {
	return c.WithHostConfig(hc.WithTmpfs(key, value))
}

func (c *Container) Privileged() *Container {
	return c.WithHostConfig(hc.WithPrivileged())
}

func (c *Container) AddedDevice(device string, pathInContainer string, permissions string) *Container {
	return c.WithHostConfig(hc.WithAddedDevice(device, pathInContainer, permissions))
}

func (c *Container) ContainerIDFile(path string) *Container {
	return c.WithHostConfig(hc.WithContainerIDFile(path))
}

func (c *Container) CPUShares(shares int64) *Container {
	return c.WithHostConfig(hc.WithCPUShares(shares))
}

func (c *Container) CPUPeriod(period int) *Container {
	return c.WithHostConfig(hc.WithCPUPeriod(period))
}

func (c *Container) CPUPeriodString(period string) *Container {
	return c.WithHostConfig(hc.WithCPUPeriod(period))
}

func (c *Container) CPUPercent(percent int64) *Container {
	return c.WithHostConfig(hc.WithCPUPercent(percent))
}

func (c *Container) CPUQuota(quota int) *Container {
	return c.WithHostConfig(hc.WithCPUQuota(quota))
}

func (c *Container) CPUQuotaString(quota string) *Container {
	return c.WithHostConfig(hc.WithCPUQuota(quota))
}

func (c *Container) CpusetCpus(cpus string) *Container {
	return c.WithHostConfig(hc.WithCpusetCpus(cpus))
}

func (c *Container) MemoryReservation(memory int) *Container {
	return c.WithHostConfig(hc.WithMemoryReservation(memory))
}

func (c *Container) MemoryReservationString(memory string) *Container {
	return c.WithHostConfig(hc.WithMemoryReservation(memory))
}

func (c *Container) MemorySwap(memorySwap int) *Container {
	return c.WithHostConfig(hc.WithMemorySwap(memorySwap))
}

func (c *Container) MemorySwapString(memorySwap string) *Container {
	return c.WithHostConfig(hc.WithMemorySwap(memorySwap))
}

func (c *Container) Ulimits(name string, soft, hard int64) *Container {
	return c.WithHostConfig(hc.WithUlimits(name, soft, hard))
}

func (c *Container) Init() *Container {
	return c.WithHostConfig(hc.WithInit())
}

func (c *Container) CPURealtimePeriod(period int) *Container {
	return c.WithHostConfig(hc.WithCPURealtimePeriod(period))
}

func (c *Container) CPURealtimePeriodString(period string) *Container {
	return c.WithHostConfig(hc.WithCPURealtimePeriod(period))
}

func (c *Container) CPURealtimeRuntime(runtime int64) *Container {
	return c.WithHostConfig(hc.WithCPURealtimeRuntime(runtime))
}

func (c *Container) CpusetMems(mems string) *Container {
	return c.WithHostConfig(hc.WithCpusetMems(mems))
}

func (c *Container) MemorySwappiness(swappiness int64) *Container {
	return c.WithHostConfig(hc.WithMemorySwappiness(swappiness))
}

func (c *Container) KernelMemory(memory int) *Container {
	return c.WithHostConfig(hc.WithKernelMemory(memory))
}

func (c *Container) KernelMemoryString(memory string) *Container {
	return c.WithHostConfig(hc.WithKernelMemory(memory))
}

func (c *Container) PidsLimit(limit int64) *Container {
	return c.WithHostConfig(hc.WithPidsLimit(limit))
}

func (c *Container) BlkioWeight(weight uint16) *Container {
	return c.WithHostConfig(hc.WithBlkioWeight(weight))
}

func (c *Container) BlkioDeviceReadBps(devicePath string, rate int) *Container {
	return c.WithHostConfig(hc.WithBlkioDeviceReadBps(devicePath, rate))
}

func (c *Container) BlkioDeviceReadBpsString(devicePath, rate string) *Container {
	return c.WithHostConfig(hc.WithBlkioDeviceReadBps(devicePath, rate))
}

func (c *Container) BlkioDeviceWriteBps(devicePath string, rate int) *Container {
	return c.WithHostConfig(hc.WithBlkioDeviceWriteBps(devicePath, rate))
}

func (c *Container) BlkioDeviceWriteBpsString(devicePath, rate string) *Container {
	return c.WithHostConfig(hc.WithBlkioDeviceWriteBps(devicePath, rate))
}

func (c *Container) BlkioDeviceReadIOps(devicePath string, rate uint64) *Container {
	return c.WithHostConfig(hc.WithBlkioDeviceReadIOps(devicePath, rate))
}

func (c *Container) BlkioDeviceWriteIOps(devicePath string, rate uint64) *Container {
	return c.WithHostConfig(hc.WithBlkioDeviceWriteIOps(devicePath, rate))
}

func (c *Container) Sysctls(key, value string) *Container {
	return c.WithHostConfig(hc.WithSysctls(key, value))
}

func (c *Container) DeviceCgroupRules(rules ...string) *Container {
	return c.WithHostConfig(hc.WithDeviceCgroupRules(rules...))
}

func (c *Container) CgroupParent(parent string) *Container {
	return c.WithHostConfig(hc.WithCgroupParent(parent))
}

func (c *Container) DeviceRequest(driver string, count int, deviceIDs []string, capabilities [][]string) *Container {
	return c.WithHostConfig(hc.WithDeviceRequest(driver, count, deviceIDs, capabilities))
}

func (c *Container) LogDriver(driver string, options map[string]string) *Container {
	return c.WithHostConfig(hc.WithLogDriver(driver, options))
}

func (c *Container) AddedCapabilities(caps ...Capability) *Container {
	return c.WithHostConfig(hc.WithAddedCapabilities(caps...))
}

func (c *Container) DroppedCapabilities(caps ...Capability) *Container {
	return c.WithHostConfig(hc.WithDroppedCapabilities(caps...))
}

func (c *Container) DroppedAllCapabilities() *Container {
	return c.WithHostConfig(hc.WithDroppedAllCapabilities())
}

func (c *Container) DroppedSensitiveCapabilities() *Container {
	return c.WithHostConfig(hc.WithDroppedSensitiveCapabilities())
}

func (c *Container) RestartPolicy(mode RestartPolicy, maxRetryCount int) *Container {
	return c.WithHostConfig(hc.WithRestartPolicy(mode, maxRetryCount))
}

func (c *Container) RestartPolicyAlways() *Container {
	return c.WithHostConfig(hc.WithRestartPolicyAlways())
}

func (c *Container) RestartPolicyOnFailure(maxRetryCount int) *Container {
	return c.WithHostConfig(hc.WithRestartPolicyOnFailure(maxRetryCount))
}

func (c *Container) RestartPolicyUnlessStopped() *Container {
	return c.WithHostConfig(hc.WithRestartPolicyUnlessStopped())
}

func (c *Container) RestartPolicyNever() *Container {
	return c.WithHostConfig(hc.WithRestartPolicyNever())
}

func (c *Container) RWHostBindMount(source string, target string) *Container {
	return c.WithHostConfig(hc.WithRWHostBindMount(source, target))
}

func (c *Container) ROHostBindMount(source string, target string) *Container {
	return c.WithHostConfig(hc.WithROHostBindMount(source, target))
}

func (c *Container) TmpfsMount(target string, sizeBytes int, mode os.FileMode) *Container {
	return c.WithHostConfig(hc.WithTmpfsMount(target, sizeBytes, mode))
}

func (c *Container) RONamedVolumeMount(name, target string) *Container {
	return c.WithHostConfig(hc.WithRONamedVolumeMount(name, target))
}

func (c *Container) RWNamedVolumeMount(name, target string) *Container {
	return c.WithHostConfig(hc.WithRWNamedVolumeMount(name, target))
}

func (c *Container) HostBindMountRecursiveRO(source, target string) *Container {
	return c.WithHostConfig(hc.WithHostBindMountRecursiveRO(source, target))
}

func (c *Container) TmpfsMountUIDGID(target string, sizeBytes int, uid, gid string) *Container {
	return c.WithHostConfig(hc.WithTmpfsMountUIDGID(target, sizeBytes, uid, gid))
}

func (c *Container) RWNamedVolumeMountWithLabel(name, target, labelKey, labelValue string) *Container {
	return c.WithHostConfig(hc.WithRWNamedVolumeMountWithLabel(name, target, labelKey, labelValue))
}

func (c *Container) RWNamedVolumeSubPath(name, target, subPath string) *Container {
	return c.WithHostConfig(hc.WithRWNamedVolumeSubPath(name, target, subPath))
}

func (c *Container) BindMountWithPropagation(source, target string, propagation MountPropagation) *Container {
	return c.WithHostConfig(hc.WithBindMountWithPropagation(source, target, propagation))
}

func (c *Container) TmpfsMountExec(target string, sizeBytes int) *Container {
	return c.WithHostConfig(hc.WithTmpfsMountExec(target, sizeBytes))
}

func (c *Container) TmpfsMountCustomOptions(target string, sizeBytes int, flags ...[]string) *Container {
	return c.WithHostConfig(hc.WithTmpfsMountCustomOptions(target, sizeBytes, flags...))
}

func (c *Container) NonRecursiveBindMount(source, target string, readonly bool) *Container {
	return c.WithHostConfig(hc.WithNonRecursiveBindMount(source, target, readonly))
}

func (c *Container) Architecture(architecture string) *Container {
	return c.WithPlatformConfig(pc.WithArchitecture(architecture))
}

func (c *Container) OS(OS string) *Container {
	return c.WithPlatformConfig(pc.WithOS(OS))
}

func (c *Container) OSVersion(OSVersion string) *Container {
	return c.WithPlatformConfig(pc.WithOSVersion(OSVersion))
}

func (c *Container) OSFeatures(OSFeatures ...string) *Container {
	return c.WithPlatformConfig(pc.WithOSFeatures(OSFeatures...))
}

func (c *Container) Variant(variant string) *Container {
	return c.WithPlatformConfig(pc.WithVariant(variant))
}
