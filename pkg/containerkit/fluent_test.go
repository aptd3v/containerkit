package containerkit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFluentMethodsPopulateConfig(t *testing.T) {
	c := NewContainer("unit", "ctr")
	require.Equal(t, "unit-ctr", c.Name)

	c.
		Env("A", "1").
		EnvMap(map[string]string{"B": "2"}).
		ExposedPort("tcp", "80").
		HostName("host").
		DomainName("example.test").
		Image("alpine:latest").
		Imagef("%s:%s", "alpine", "latest").
		Command("sleep", "infinity").
		User("root").
		AttachedStdin().
		AttachedStdout().
		AttachedStderr().
		Tty().
		StdinOpen().
		StdinOnce().
		EscapedArgs().
		Volume("/data").
		WorkingDir("/work").
		DisabledNetwork().
		OnBuild("echo onbuild").
		Label("k", "v").
		StopSignal("SIGTERM").
		Entrypoint("/bin/sh", "-c").
		Shell("/bin/sh").
		StopTimeout(10).
		MemoryLimit(64*1024*1024).
		MemoryLimitString("64m").
		RestartAlways().
		AutoRemove().
		PortBindings("tcp", "127.0.0.1", "8080", "80").
		DNSLookups("1.1.1.1").
		DNSOptions("ndots:1").
		DNSSearches("example.test").
		ExtraHost("foo:127.0.0.1").
		AddedGroups("audio").
		VolumeBinds("/tmp:/tmp:rw").
		UTSMode("host").
		UserNSMode("host").
		ShmSize(1024*1024).
		ShmSizeString("1m").
		Runtime("runc").
		ConsoleSize(24, 80).
		Isolation("default").
		CPUCount(1).
		ReadonlyPaths("/proc").
		MaskedPaths("/sys").
		NetworkMode("bridge").
		VolumeDriver("local").
		VolumesFrom("other:ro").
		IpcMode("shareable").
		Cgroup("host").
		OomScoreAdj(100).
		OomKillDisable().
		PidMode("host").
		PublishAllPorts().
		ReadOnlyRootfs().
		SecurityOpts("no-new-privileges").
		StorageOpt("size", "1G").
		Tmpfs("/run", "rw,size=64m").
		Privileged().
		AddedDevice("/dev/null", "/dev/null", "rwm").
		ContainerIDFile("/tmp/cid").
		CPUShares(512).
		CPUPeriod(100000).
		CPUPeriodString("100ms").
		CPUPercent(50).
		CPUQuota(50000).
		CPUQuotaString("50ms").
		CpusetCpus("0").
		MemoryReservation(32*1024*1024).
		MemoryReservationString("32m").
		MemorySwap(128*1024*1024).
		MemorySwapString("128m").
		Ulimits("nofile", 1024, 2048).
		Init().
		CPURealtimePeriod(1000).
		CPURealtimePeriodString("1ms").
		CPURealtimeRuntime(500).
		CpusetMems("0").
		MemorySwappiness(60).
		KernelMemory(16*1024*1024).
		KernelMemoryString("16m").
		PidsLimit(100).
		BlkioWeight(500).
		BlkioDeviceReadBps("/dev/zero", 1024).
		BlkioDeviceReadBpsString("/dev/zero", "1kb").
		BlkioDeviceWriteBps("/dev/zero", 1024).
		BlkioDeviceWriteBpsString("/dev/zero", "1kb").
		BlkioDeviceReadIOps("/dev/zero", 10).
		BlkioDeviceWriteIOps("/dev/zero", 10).
		Sysctls("net.ipv4.ip_forward", "1").
		DeviceCgroupRules("c 1:3 rwm").
		CgroupParent("slice").
		DeviceRequest("nvidia", 1, []string{"0"}, [][]string{{"gpu"}}).
		LogDriver("json-file", map[string]string{"max-size": "10m"}).
		AddedCapabilities(NET_BIND_SERVICE).
		DroppedCapabilities(MKNOD).
		DroppedAllCapabilities().
		DroppedSensitiveCapabilities().
		RestartPolicy(RestartPolicyOnFailure, 3).
		RestartPolicyAlways().
		RestartPolicyOnFailure(2).
		RestartPolicyUnlessStopped().
		RestartPolicyNever().
		RWHostBindMount("/tmp", "/mnt/rw").
		ROHostBindMount("/tmp", "/mnt/ro").
		TmpfsMount("/mnt/tmp", 1024, 0755).
		RONamedVolumeMount("data", "/mnt/vol-ro").
		RWNamedVolumeMount("data", "/mnt/vol-rw").
		HostBindMountRecursiveRO("/tmp", "/mnt/rec").
		TmpfsMountUIDGID("/mnt/tmp2", 2048, "0", "0").
		RWNamedVolumeMountWithLabel("data", "/mnt/vol-l", "foo", "bar").
		RWNamedVolumeSubPath("data", "/mnt/sub", "sub").
		BindMountWithPropagation("/tmp", "/mnt/prop", PropagationRPrivate).
		TmpfsMountExec("/mnt/tmpx", 4096).
		TmpfsMountCustomOptions("/mnt/tmpc", 8192, []string{"noexec"}).
		NonRecursiveBindMount("/tmp", "/mnt/nr", true).
		Architecture("amd64").
		OS("linux").
		OSVersion("6.0").
		OSFeatures("feat").
		Variant("v1")

	require.NoError(t, c.Validate())
	require.Equal(t, "alpine:latest", c.Config.Container.Image)
	require.Equal(t, "linux", c.Config.Platform.OS)
	require.NotEmpty(t, c.Config.Host.Binds)
	require.NotEmpty(t, c.Config.Host.Mounts)
}

func TestFluentEmptySettersNoop(t *testing.T) {
	c := NewContainer("x")
	require.Equal(t, c, c.WithContainerConfig())
	require.Equal(t, c, c.WithHostConfig())
	require.Equal(t, c, c.WithNetworkConfig())
	require.Equal(t, c, c.WithPlatformConfig())
	require.Equal(t, c, c.WithContainerConfig(nil))
	require.Equal(t, c, c.WithHostConfig(nil))
	require.Equal(t, c, c.WithNetworkConfig(nil))
	require.Equal(t, c, c.WithPlatformConfig(nil))
}

func TestFluentInvalidMemoryStringRecordsError(t *testing.T) {
	c := NewContainer("bad").MemoryLimitString("not-a-size")
	require.Error(t, c.Validate())
}
