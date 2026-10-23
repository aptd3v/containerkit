package containerkit

import (
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"
)

func TestFluentMethodsPopulateConfig(t *testing.T) {
	c := NewContainer("unit", "ctr")
	require.Equal(t, "unit-ctr", c.Name)

	c.Env("A", "1").EnvMap(map[string]string{"B": "2"}).ExposedPort("tcp", "80").HostName("host").DomainName("example.test").Image("alpine:latest").Imagef("%s:%s", "alpine", "latest").Command("sleep", "infinity").User("root").AttachedStdin().AttachedStdout().AttachedStderr().Tty().StdinOpen().StdinOnce().EscapedArgs().Volume("/data").WorkingDir("/work").DisabledNetwork().OnBuild("echo onbuild").Label("k", "v").StopSignal("SIGTERM").Entrypoint("/bin/sh", "-c").Shell("/bin/sh").StopTimeout(10).MemoryLimit(64*MiB).RestartAlways().AutoRemove().PortBinding(PortBinding{Protocol: TCP, HostIP: "127.0.0.1", HostPort: "8080", ContainerPort: "80"}).DNSLookups("1.1.1.1").DNSOptions("ndots:1").DNSSearches("example.test").ExtraHost("foo", "127.0.0.1").AddedGroups("audio").VolumeBinds("/tmp:/tmp:rw").UTSMode(UTSHost).UserNSMode(UserNSHost).ShmSize(MiB).Runtime("runc").ConsoleSize(ConsoleSize{Height: 24, Width: 80}).Isolation(IsolationDefault).CPUCount(1).ReadonlyPaths("/proc").MaskedPaths("/sys").NetworkMode(NetworkBridge).VolumeDriver("local").VolumesFrom("other", true).IpcMode(IpcShareable).Cgroup(CgroupHost).OomScoreAdj(100).OomKillDisable().PidMode(PidHost).PublishAllPorts().ReadOnlyRootfs().SecurityOpts("no-new-privileges").StorageOpt("size", "1G").Tmpfs("/run", "rw,size=64m").Privileged().AddedDevice(Device{HostPath: "/dev/null", ContainerPath: "/dev/null", Read: true, Write: true, Mknod: true}).ContainerIDFile("/tmp/cid").CPUShares(512).CPUPeriod(100*time.Millisecond).CPUPercent(50).CPUQuota(50*time.Millisecond).CpusetCpus(mustCpuset("0")).MemoryReservation(32*MiB).MemorySwap(128*MiB).Ulimits(Ulimit{Name: "nofile", Soft: 1024, Hard: 2048}).Init().CPURealtimePeriod(time.Millisecond).CPURealtimeRuntime(500).CpusetMems(mustCpuset("0")).MemorySwappiness(60).KernelMemory(16*MiB).PidsLimit(100).BlkioWeight(500).BlkioDeviceReadBps("/dev/zero", KiB).BlkioDeviceWriteBps("/dev/zero", KiB).BlkioDeviceReadIOps("/dev/zero", 10).BlkioDeviceWriteIOps("/dev/zero", 10).Sysctls("net.ipv4.ip_forward", "1").DeviceCgroupRules("c 1:3 rwm").CgroupParent("slice").DeviceRequest(DeviceRequest{Driver: "nvidia", Count: 1, DeviceIDs: []string{"0"}, Capabilities: []string{"gpu"}}).LogDriver("json-file", map[string]string{"max-size": "10m"}).AddedCapabilities(NET_BIND_SERVICE).DroppedCapabilities(MKNOD).DroppedAllCapabilities().DroppedSensitiveCapabilities().RestartPolicy(RestartPolicyOnFailure, 3).RestartPolicyAlways().RestartPolicyOnFailure(2).RestartPolicyUnlessStopped().RestartPolicyNever().RWHostBindMount("/tmp", "/mnt/rw").ROHostBindMount("/tmp", "/mnt/ro").TmpfsMount("/mnt/tmp", 1024, 0755).RONamedVolumeMount("data", "/mnt/vol-ro").RWNamedVolumeMount("data", "/mnt/vol-rw").HostBindMountRecursiveRO("/tmp", "/mnt/rec").TmpfsMountUIDGID("/mnt/tmp2", 2048, "0", "0").RWNamedVolumeMountWithLabel("data", "/mnt/vol-l", "foo", "bar").RWNamedVolumeSubPath("data", "/mnt/sub", "sub").BindMountWithPropagation("/tmp", "/mnt/prop", PropagationRPrivate).TmpfsMountExec("/mnt/tmpx", 4096).TmpfsMountCustomOptions("/mnt/tmpc", 8192, []string{"noexec"}).NonRecursiveBindMount("/tmp", "/mnt/nr", true).Architecture("amd64").OS("linux").OSVersion("6.0").OSFeatures("feat").Variant("v1")

	require.NoError(t, c.Validate())
	require.Equal(t, "alpine:latest", c.Config.Container.Image)
	require.Equal(t, "linux", c.Config.Platform.OS)
	require.Empty(t, c.Config.Host.Binds)
	mounts := map[string]mount.Mount{}
	for _, m := range c.Config.Host.Mounts {
		mounts[m.Target] = m
	}
	require.Equal(t, mount.TypeBind, mounts["/tmp"].Type)
	require.Equal(t, "/tmp", mounts["/tmp"].Source)
	require.False(t, mounts["/tmp"].ReadOnly)
	require.Equal(t, mount.TypeBind, mounts["/mnt/rw"].Type)
	require.False(t, mounts["/mnt/rw"].ReadOnly)
	require.True(t, mounts["/mnt/ro"].ReadOnly)
	require.Equal(t, mount.TypeTmpfs, mounts["/mnt/tmp"].Type)
	require.Equal(t, int64(1024), mounts["/mnt/tmp"].TmpfsOptions.SizeBytes)
	require.Equal(t, os.FileMode(0755), mounts["/mnt/tmp"].TmpfsOptions.Mode)
	require.True(t, mounts["/mnt/vol-ro"].ReadOnly)
	require.Equal(t, mount.TypeVolume, mounts["/mnt/vol-rw"].Type)
	require.True(t, mounts["/mnt/rec"].BindOptions.ReadOnlyForceRecursive)
	require.Equal(t, [][]string{{"uid", "0"}, {"gid", "0"}}, mounts["/mnt/tmp2"].TmpfsOptions.Options)
	require.Equal(t, "bar", mounts["/mnt/vol-l"].VolumeOptions.Labels["foo"])
	require.Equal(t, "sub", mounts["/mnt/sub"].VolumeOptions.Subpath)
	require.Equal(t, mount.PropagationRPrivate, mounts["/mnt/prop"].BindOptions.Propagation)
	require.Equal(t, [][]string{{"exec"}}, mounts["/mnt/tmpx"].TmpfsOptions.Options)
	require.Equal(t, [][]string{{"noexec"}}, mounts["/mnt/tmpc"].TmpfsOptions.Options)
	require.True(t, mounts["/mnt/nr"].BindOptions.NonRecursive)
	require.True(t, mounts["/mnt/nr"].ReadOnly)
}

func TestFluentInvalidMemoryStringRecordsError(t *testing.T) {
	_, err := ParseByteSize("not-a-size")
	require.Error(t, err)
}
