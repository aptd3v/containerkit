package hc_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/aptd3v/go-contain/pkg/tools"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	mountType "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

var (
	nilPortMap nat.PortMap
	boolTrue   = true
	i64100     = int64(100)
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *container.HostConfig
		setFn    create.SetHostConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &container.HostConfig{},
			field:    "MemoryReservation",
			setFn:    hc.WithMemoryReservation("100Error"),
			wantErr:  true,
			message:  "WithMemoryReservation error",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			field:    "MemoryReservation",
			setFn:    hc.WithMemoryReservation("100M"),
			wantErr:  false,
			message:  "WithMemoryReservation ok",
			expected: int64(100 * 1024 * 1024),
		},
		{
			config:   &container.HostConfig{},
			field:    "MemorySwap",
			setFn:    hc.WithMemorySwap("100M"),
			wantErr:  false,
			message:  "WithMemorySwap ok",
			expected: int64(100 * 1024 * 1024),
		},
		{
			config:   &container.HostConfig{},
			field:    "MemorySwap",
			setFn:    hc.WithMemorySwap("100Error"),
			wantErr:  true,
			message:  "WithMemorySwap error",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			field:    "ShmSize",
			setFn:    hc.WithShmSize("100Error"),
			wantErr:  true,
			message:  "WithShmSize error",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			field:    "MemoryReservation",
			setFn:    hc.WithMemoryLimit("100Error"),
			wantErr:  true,
			message:  "WithMemoryLimit error",
			expected: nil,
		},

		{
			config:   &container.HostConfig{},
			field:    "KernelMemory",
			setFn:    hc.WithKernelMemory("100Error"),
			wantErr:  true,
			message:  "WithKernelMemory error",
			expected: nil,
		},

		{
			config:   &container.HostConfig{},
			field:    "KernelMemory",
			setFn:    hc.WithKernelMemory("100M"),
			wantErr:  false,
			message:  "WithKernelMemory ok",
			expected: int64(100 * 1024 * 1024),
		},

		{
			config:   &container.HostConfig{},
			setFn:    hc.WithMemoryLimit("100M"),
			field:    "Memory",
			wantErr:  false,
			message:  "WithMemoryLimit ok",
			expected: int64(100 * 1024 * 1024),
		},

		{
			config:   &container.HostConfig{},
			setFn:    hc.WithShmSize("100M"),
			field:    "ShmSize",
			wantErr:  false,
			message:  "WithShmSize ok",
			expected: int64(100 * 1024 * 1024),
		},

		{
			config: &container.HostConfig{},
			setFn: hc.WithMountPoint(
				mount.WithSource("/tmp/test"),
				mount.WithTarget("/tmp/test"),
				mount.WithType("bind"),
				mount.WithReadOnly(),
			),
			field:   "Mounts",
			wantErr: false,
			message: "WithMountPoint ok",
			expected: []mountType.Mount{
				{
					Source:   "/tmp/test",
					Target:   "/tmp/test",
					Type:     "bind",
					ReadOnly: true,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: hc.WithMountPoint(
				mount.Fail(errors.New("test")),
			),
			field:    "Mounts",
			wantErr:  true,
			message:  "WithMountPoint error",
			expected: nil,
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartAlways(10),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartAlways ok",
			expected: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 10,
			},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithMemoryLimit(1024),
			field:    "Memory",
			wantErr:  false,
			message:  "WithMemoryLimit ok",
			expected: int64(1024),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithAutoRemove(),
			field:    "AutoRemove",
			wantErr:  false,
			message:  "WithAutoRemove ok",
			expected: true,
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithPortBindings("tcp", "0.0.0.0", "8080", "8080"),
			field:   "PortBindings",
			wantErr: false,
			message: "WithPortBindings ok",
			expected: nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
			},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "0.0.0.0", "808012", ""),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings error",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "0.0.0.0", "", ""),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings empty host and container port",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("", "", "", ""),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings empty values",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "", "8080", "8080"),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings empty host IP",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "0.0.0.0", "8080", ""),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings empty container port",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("", "0.0.0.0", "8080", "8080"),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings empty protocol",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "0.0.0.0", "", "8080"),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings empty host port",
			expected: nilPortMap,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "0.0.0.0", "8080", "invalid"),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings invalid container port",
			expected: nilPortMap,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "0.0.0.0", "invalid", "8080"),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings invalid host port",
			expected: nilPortMap,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithDNSLookups("1.1.1.1", "8.8.8.8"),
			field:    "DNS",
			wantErr:  false,
			message:  "WithDNSLookups ok",
			expected: []string{"1.1.1.1", "8.8.8.8"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithDNSOptions("dns-option1", "dns-option2"),
			field:    "DNSOptions",
			wantErr:  false,
			message:  "WithDNSOptions ok",
			expected: []string{"dns-option1", "dns-option2"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithDNSSearches("dns-search1", "dns-search2"),
			field:    "DNSSearch",
			wantErr:  false,
			message:  "WithDNSSearches ok",
			expected: []string{"dns-search1", "dns-search2"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithExtraHost("test.com"),
			field:    "ExtraHosts",
			wantErr:  false,
			message:  "WithExtraHost ok",
			expected: []string{"test.com"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithAddedGroups("group1", "group2"),
			field:    "GroupAdd",
			wantErr:  false,
			message:  "WithAddedGroups ok",
			expected: []string{"group1", "group2"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumeBinds("/host/path:/container/path:ro"),
			field:    "Binds",
			wantErr:  false,
			message:  "WithVolumeBinds ok",
			expected: []string{"/host/path:/container/path:ro"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumeBinds("/host/path:/container/path:ro", "/host/path2:/container/path2:rw"),
			field:    "Binds",
			wantErr:  false,
			message:  "WithVolumeBinds ok",
			expected: []string{"/host/path:/container/path:ro", "/host/path2:/container/path2:rw"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumeBinds("/host/path:/container/path:ro", "/host/path2:/container/path2:rw", "invalid"),
			field:    "Binds",
			wantErr:  true,
			message:  "WithVolumeBinds invalid",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumeBinds("/host/path:/container/path:ro", "/host/path:/container/path:rw", "invalid"),
			field:    "Binds",
			wantErr:  true,
			message:  "WithVolumeBinds duplicate target path",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumeBinds("/host/path:/container/path:invalid"),
			field:    "Binds",
			wantErr:  true,
			message:  "WithVolumeBinds bad modes",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumeBinds(":/container/path:ro", "/container/path2::rw"),
			field:    "Binds",
			wantErr:  true,
			message:  "WithVolumeBinds empty source and target",
			expected: []string{},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumeBinds(""),
			field:    "Binds",
			wantErr:  true,
			message:  "WithVolumeBinds empty bind",
			expected: []string{},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithUTSMode("host"),
			field:    "UTSMode",
			wantErr:  false,
			message:  "WithUTSMode ok",
			expected: container.UTSMode("host"),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithUserNSMode("host"),
			field:    "UsernsMode",
			wantErr:  false,
			message:  "WithUserNSMode ok",
			expected: container.UsernsMode("host"),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithShmSize(1024),
			field:    "ShmSize",
			wantErr:  false,
			message:  "WithShmSize ok",
			expected: int64(1024),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithRuntime("runc"),
			field:    "Runtime",
			wantErr:  false,
			message:  "WithRuntime ok",
			expected: "runc",
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithConsoleSize(10, 10),
			field:    "ConsoleSize",
			wantErr:  false,
			message:  "WithConsoleSize ok",
			expected: [2]uint{10, 10},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithIsolation("default"),
			field:    "Isolation",
			wantErr:  false,
			message:  "WithIsolation ok",
			expected: container.Isolation("default"),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCPUCount(1),
			field:    "CPUCount",
			wantErr:  false,
			message:  "WithCPUCount ok",
			expected: int64(1),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithReadonlyPaths("/host/path"),
			field:    "ReadonlyPaths",
			wantErr:  false,
			message:  "WithReadonlyPaths ok",
			expected: []string{"/host/path"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithMaskedPaths("/host/path", "/host/path2"),
			field:    "MaskedPaths",
			wantErr:  false,
			message:  "WithMaskedPaths ok",
			expected: []string{"/host/path", "/host/path2"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithNetworkMode("host"),
			field:    "NetworkMode",
			wantErr:  false,
			message:  "WithNetworkMode ok",
			expected: container.NetworkMode("host"),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumeDriver("local"),
			field:    "VolumeDriver",
			wantErr:  false,
			message:  "WithVolumeDriver ok",
			expected: "local",
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithVolumesFrom("container1"),
			field:    "VolumesFrom",
			wantErr:  false,
			message:  "WithVolumesFrom ok",
			expected: []string{"container1"},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithIpcMode("host"),
			field:    "IpcMode",
			wantErr:  false,
			message:  "WithIpcMode ok",
			expected: container.IpcMode("host"),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCgroup("host"),
			field:    "Cgroup",
			wantErr:  false,
			message:  "WithCgroup ok",
			expected: container.CgroupSpec("host"),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithOomScoreAdj(100),
			field:    "OomScoreAdj",
			wantErr:  false,
			message:  "WithOomScoreAdj ok",
			expected: int(100),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithOomKillDisable(),
			field:    "OomKillDisable",
			wantErr:  false,
			message:  "WithOomKillDisable ok",
			expected: &boolTrue,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPidMode("host"),
			field:    "PidMode",
			wantErr:  false,
			message:  "WithPidMode ok",
			expected: container.PidMode("host"),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPublishAllPorts(),
			field:    "PublishAllPorts",
			wantErr:  false,
			message:  "WithPublishAllPorts ok",
			expected: boolTrue,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithReadOnlyRootfs(),
			field:    "ReadonlyRootfs",
			wantErr:  false,
			message:  "WithReadOnlyRootfs ok",
			expected: true,
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithSecurityOpts("test1", "test2"),
			field:   "SecurityOpt",
			wantErr: false,
			message: "WithSecurityOpts ok",
			expected: []string{
				"test1",
				"test2",
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithStorageOpt("test1", "test2"),
			field:   "StorageOpt",
			wantErr: false,
			message: "WithStorageOpt ok",
			expected: map[string]string{
				"test1": "test2",
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithTmpfs("test1", "test2"),
			field:   "Tmpfs",
			wantErr: false,
			message: "WithTmpfs ok",
			expected: map[string]string{
				"test1": "test2",
			},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPrivileged(),
			field:    "Privileged",
			wantErr:  false,
			message:  "WithPrivileged ok",
			expected: true,
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithAddedDevice("/dev/null", "/dev/null", "rwm"),
			field:   "Devices",
			wantErr: false,
			message: "WithAddedDevice ok",
			expected: []container.DeviceMapping{
				{
					PathOnHost:        "/dev/null",
					PathInContainer:   "/dev/null",
					CgroupPermissions: "rwm",
				},
			},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithContainerIDFile("/path/to/container-id.txt"),
			field:    "ContainerIDFile",
			wantErr:  false,
			message:  "WithContainerIDFile ok",
			expected: "/path/to/container-id.txt",
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCPUShares(100),
			field:    "CPUShares",
			wantErr:  false,
			message:  "WithCPUShares ok",
			expected: int64(100),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCPUPeriod(100),
			field:    "CPUPeriod",
			wantErr:  false,
			message:  "WithCPUPeriod ok",
			expected: int64(100),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCPUQuota(100),
			field:    "CPUQuota",
			wantErr:  false,
			message:  "WithCPUQuota ok",
			expected: int64(100),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCPUPercent(100),
			field:    "CPUPercent",
			wantErr:  false,
			message:  "WithCPUPercent ok",
			expected: int64(100),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithMemoryReservation(100),
			field:    "MemoryReservation",
			wantErr:  false,
			message:  "WithMemoryReservation ok",
			expected: int64(100),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithMemorySwap(100),
			field:    "MemorySwap",
			wantErr:  false,
			message:  "WithMemorySwap ok",
			expected: int64(100),
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithUlimits("nproc", 100, 200),
			field:   "Ulimits",
			wantErr: false,
			message: "WithUlimits ok",
			expected: []*container.Ulimit{
				{
					Name: "nproc",
					Soft: 100,
					Hard: 200,
				},
			},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithInit(),
			field:    "Init",
			wantErr:  false,
			message:  "WithInit ok",
			expected: &boolTrue,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCPURealtimePeriod(100),
			field:    "CPURealtimePeriod",
			wantErr:  false,
			message:  "WithCPURealtimePeriod ok",
			expected: int64(100),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCPURealtimeRuntime(100),
			field:    "CPURealtimeRuntime",
			wantErr:  false,
			message:  "WithCPURealtimeRuntime ok",
			expected: int64(100),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCpusetMems("0,1"),
			field:    "CpusetMems",
			wantErr:  false,
			message:  "WithCpusetMems ok",
			expected: "0,1",
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithMemorySwappiness(100),
			field:    "MemorySwappiness",
			wantErr:  false,
			message:  "WithMemorySwappiness ok",
			expected: &i64100,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCpusetCpus("0,1"),
			field:    "CpusetCpus",
			wantErr:  false,
			message:  "WithCpusetCpus ok",
			expected: "0,1",
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithKernelMemory(100),
			field:    "KernelMemory",
			wantErr:  false,
			message:  "WithKernelMemory ok",
			expected: i64100,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPidsLimit(100),
			field:    "PidsLimit",
			wantErr:  false,
			message:  "WithPidsLimit ok",
			expected: &i64100,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithBlkioWeight(100),
			field:    "BlkioWeight",
			wantErr:  false,
			message:  "WithBlkioWeight ok",
			expected: uint16(100),
		},
		{
			config: &container.HostConfig{},
			setFn: tools.Group(
				hc.WithBlkioDeviceReadBps("/dev/0", 100),
				hc.WithBlkioDeviceReadBps("/dev/1", 100),
			),
			field:   "BlkioDeviceReadBps",
			wantErr: false,
			message: "WithBlkioDeviceReadBps ok",
			expected: []*blkiodev.ThrottleDevice{
				{
					Path: "/dev/0",
					Rate: 100,
				},
				{
					Path: "/dev/1",
					Rate: 100,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: tools.Group(
				hc.WithBlkioDeviceWriteBps("/dev/0", 100),
				hc.WithBlkioDeviceWriteBps("/dev/1", 100),
			),
			field:   "BlkioDeviceWriteBps",
			wantErr: false,
			message: "WithBlkioDeviceWriteBps ok",
			expected: []*blkiodev.ThrottleDevice{
				{
					Path: "/dev/0",
					Rate: 100,
				},
				{
					Path: "/dev/1",
					Rate: 100,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: tools.Group(
				hc.WithBlkioDeviceReadIOps("/dev/0", 100),
				hc.WithBlkioDeviceReadIOps("/dev/1", 100),
			),
			field:   "BlkioDeviceReadIOps",
			wantErr: false,
			message: "WithBlkioDeviceReadIOps ok",
			expected: []*blkiodev.ThrottleDevice{
				{
					Path: "/dev/0",
					Rate: 100,
				},
				{
					Path: "/dev/1",
					Rate: 100,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: tools.Group(
				hc.WithBlkioDeviceWriteIOps("/dev/0", 100),
				hc.WithBlkioDeviceWriteIOps("/dev/1", 100),
			),
			field:   "BlkioDeviceWriteIOps",
			wantErr: false,
			message: "WithBlkioDeviceWriteIOps ok",
			expected: []*blkiodev.ThrottleDevice{
				{
					Path: "/dev/0",
					Rate: 100,
				},
				{
					Path: "/dev/1",
					Rate: 100,
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithSysctls("net.ipv4.ip_forward", "1"),
			field:   "Sysctls",
			wantErr: false,
			message: "WithSysctls ok",
			expected: map[string]string{
				"net.ipv4.ip_forward": "1",
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithDeviceCgroupRules("c 1:3 rwm"),
			field:   "DeviceCgroupRules",
			wantErr: false,
			message: "WithDeviceCgroupRules ok",
			expected: []string{
				"c 1:3 rwm",
			},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithCgroupParent("test"),
			field:    "CgroupParent",
			wantErr:  false,
			message:  "WithCgroupParent ok",
			expected: "test",
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithDeviceRequest("nvidia", 1, []string{"1"}, [][]string{{"gpu"}}),
			field:   "DeviceRequests",
			wantErr: false,
			message: "WithDeviceRequest ok",
			expected: []container.DeviceRequest{
				{
					Driver:       "nvidia",
					Count:        1,
					DeviceIDs:    []string{"1"},
					Capabilities: [][]string{{"gpu"}},
				},
			},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.Fail(errors.New("test")),
			field:    "",
			wantErr:  true,
			message:  "fail test",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.Failf("test %s", "test"),
			field:    "",
			wantErr:  true,
			message:  "failf test",
			expected: nil,
		},

		{
			config:  &container.HostConfig{},
			setFn:   hc.WithLogDriver("json-file", map[string]string{"max-size": "10m", "max-file": "3"}),
			field:   "LogConfig",
			wantErr: false,
			message: "WithLogDriver ok",
			expected: container.LogConfig{
				Type: "json-file",
				Config: map[string]string{
					"max-size": "10m",
					"max-file": "3",
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithAddedCapabilities(hc.NET_ADMIN, hc.SYS_ADMIN),
			field:   "CapAdd",
			wantErr: false,
			message: "WithAddedCapabilities ok",
			expected: strslice.StrSlice{
				"NET_ADMIN",
				"SYS_ADMIN",
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithDroppedCapabilities(hc.NET_ADMIN, hc.SYS_ADMIN),
			field:   "CapDrop",
			wantErr: false,
			message: "WithDroppedCapabilities ok",
			expected: strslice.StrSlice{
				"NET_ADMIN",
				"SYS_ADMIN",
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithDroppedAllCapabilities(),
			field:   "CapDrop",
			wantErr: false,
			message: "WithDroppedAllCapabilities ok",
			expected: strslice.StrSlice{
				"ALL",
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithDroppedSensitiveCapabilities(),
			field:   "CapDrop",
			wantErr: false,
			message: "WithDroppedSensitiveCapabilities ok",
			expected: strslice.StrSlice{
				"SYS_ADMIN",
				"SYS_MODULE",
				"SYS_PTRACE",
				"SYS_TIME",
				"SYSLOG",
				"MAC_ADMIN",
				"MAC_OVERRIDE",
				"CHECKPOINT_RESTORE",
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicy(hc.RestartPolicyNo, 0),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy no ok",
			expected: container.RestartPolicy{
				Name:              "no",
				MaximumRetryCount: 0,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicy(hc.RestartPolicyOnFailure, 10),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy on-failure ok",
			expected: container.RestartPolicy{
				Name:              "on-failure",
				MaximumRetryCount: 10,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicy(hc.RestartPolicyUnlessStopped, 10),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy ok",
			expected: container.RestartPolicy{
				Name:              "unless-stopped",
				MaximumRetryCount: 10,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicy(hc.RestartPolicyAlways, 10),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy ok",
			expected: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 10,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicy("", 0),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy  fallback disabled ok",
			expected: container.RestartPolicy{
				Name:              "no",
				MaximumRetryCount: 0,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicyAlways(),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy always ok",
			expected: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicyOnFailure(10),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy on-failure ok",
			expected: container.RestartPolicy{
				Name:              "on-failure",
				MaximumRetryCount: 10,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicyUnlessStopped(),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy unless-stopped ok",
			expected: container.RestartPolicy{
				Name:              "unless-stopped",
				MaximumRetryCount: 0,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRestartPolicyNever(),
			field:   "RestartPolicy",
			wantErr: false,
			message: "WithRestartPolicy never ok",
			expected: container.RestartPolicy{
				Name:              "no",
				MaximumRetryCount: 0,
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRWHostBindMount("/tmp/test", "/tmp/test"),
			field:   "Mounts",
			wantErr: false,
			message: "WithRWHostBindMount ok",
			expected: []mountType.Mount{
				{
					Source:   "/tmp/test",
					Target:   "/tmp/test",
					Type:     "bind",
					ReadOnly: false,
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithROHostBindMount("/tmp/test", "/tmp/test"),
			field:   "Mounts",
			wantErr: false,
			message: "WithROHostBindMount ok",
			expected: []mountType.Mount{
				{
					Source:   "/tmp/test",
					Target:   "/tmp/test",
					Type:     "bind",
					ReadOnly: true,
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithTmpfsMount("/tmp/test", 1024, 0755),
			field:   "Mounts",
			wantErr: false,
			message: "WithTmpfsMount ok",
			expected: []mountType.Mount{
				{
					Target:   "/tmp/test",
					Type:     "tmpfs",
					ReadOnly: false,
					TmpfsOptions: &mountType.TmpfsOptions{
						SizeBytes: 1024,
						Mode:      0755,
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRONamedVolumeMount("test", "/tmp/test"),
			field:   "Mounts",
			wantErr: false,
			message: "WithRONamedVolumeMount ok",
			expected: []mountType.Mount{
				{
					Source:   "test",
					Target:   "/tmp/test",
					Type:     "volume",
					ReadOnly: true,
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRWNamedVolumeMount("test", "/tmp/test"),
			field:   "Mounts",
			wantErr: false,
			message: "WithRWNamedVolumeMount ok",
			expected: []mountType.Mount{
				{
					Source:   "test",
					Target:   "/tmp/test",
					Type:     "volume",
					ReadOnly: false,
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithHostBindMountRecursiveRO("/tmp/test", "/tmp/test"),
			field:   "Mounts",
			wantErr: false,
			message: "WithHostBindMountRecursiveReadOnly ok",
			expected: []mountType.Mount{
				{
					Source:   "/tmp/test",
					Target:   "/tmp/test",
					Type:     "bind",
					ReadOnly: true,
					BindOptions: &mountType.BindOptions{
						ReadOnlyForceRecursive: true,
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithTmpfsMountUIDGID("/tmp/test", 1024, "1000", "1000"),
			field:   "Mounts",
			wantErr: false,
			message: "WithTmpfsMountUIDGID ok",
			expected: []mountType.Mount{
				{
					Target:   "/tmp/test",
					Type:     "tmpfs",
					ReadOnly: false,
					TmpfsOptions: &mountType.TmpfsOptions{
						SizeBytes: 1024,
						Options: [][]string{
							{"uid", "1000"},
							{"gid", "1000"},
						},
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRWNamedVolumeMountWithLabel("test", "/tmp/test", "label", "value"),
			field:   "Mounts",
			wantErr: false,
			message: "WithRWNamedVolumeMountWithLabel ok",
			expected: []mountType.Mount{
				{
					Source:   "test",
					Target:   "/tmp/test",
					Type:     "volume",
					ReadOnly: false,
					VolumeOptions: &mountType.VolumeOptions{
						Labels: map[string]string{
							"label": "value",
						},
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithRWNamedVolumeSubPath("test", "/tmp/test", "/some/subpath"),
			field:   "Mounts",
			wantErr: false,
			message: "WithRWNamedVolumeSubPath ok",
			expected: []mountType.Mount{
				{
					Source:   "test",
					Target:   "/tmp/test",
					Type:     "volume",
					ReadOnly: false,
					VolumeOptions: &mountType.VolumeOptions{
						Subpath: "/some/subpath",
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithBindMountWithPropagation("/tmp/test", "/tmp/test", mount.PropagationShared),
			field:   "Mounts",
			wantErr: false,
			message: "WithBindMountWithPropagation ok",
			expected: []mountType.Mount{
				{
					Source:   "/tmp/test",
					Target:   "/tmp/test",
					Type:     "bind",
					ReadOnly: false,
					BindOptions: &mountType.BindOptions{
						Propagation: "shared",
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithTmpfsMountExec("/tmp/test", 1024),
			field:   "Mounts",
			wantErr: false,
			message: "WithTmpfsMountExec ok",
			expected: []mountType.Mount{
				{
					Target:   "/tmp/test",
					Type:     "tmpfs",
					ReadOnly: false,
					TmpfsOptions: &mountType.TmpfsOptions{
						SizeBytes: 1024,
						Options: [][]string{
							{"exec"},
						},
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithTmpfsMountCustomOptions("/tmp/test", 1024, []string{"exec"}, []string{"foo", "bar"}),
			field:   "Mounts",
			wantErr: false,
			message: "WithTmpfsMountCustomOptions ok",
			expected: []mountType.Mount{
				{
					Target:   "/tmp/test",
					Type:     "tmpfs",
					ReadOnly: false,
					TmpfsOptions: &mountType.TmpfsOptions{
						SizeBytes: 1024,
						Options: [][]string{
							{"exec"},
							{"foo", "bar"},
						},
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithNonRecursiveBindMount("/tmp/test", "/tmp/test", true),
			field:   "Mounts",
			wantErr: false,
			message: "WithNonRecursiveBindMount readonly ok",
			expected: []mountType.Mount{
				{
					Source:   "/tmp/test",
					Target:   "/tmp/test",
					Type:     "bind",
					ReadOnly: true,
					BindOptions: &mountType.BindOptions{
						NonRecursive: true,
					},
				},
			},
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithNonRecursiveBindMount("/tmp/test", "/tmp/test", false),
			field:   "Mounts",
			wantErr: false,
			message: "WithNonRecursiveBindMount readwrite ok",
			expected: []mountType.Mount{
				{
					Source:   "/tmp/test",
					Target:   "/tmp/test",
					Type:     "bind",
					ReadOnly: false,
					BindOptions: &mountType.BindOptions{
						NonRecursive: true,
					},
				},
			},
		},
	}

	for _, test := range tests {
		err := test.setFn(test.config)
		if test.wantErr {
			assert.Error(t, err)
			assert.True(t, errdefs.IsHostConfigError(err), "expected container config error")
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}
