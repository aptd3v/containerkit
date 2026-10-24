package kit

import (
	"reflect"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/aptd3v/containerkit/pkg/kit/types"
	composetypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	nilPortMap nat.PortMap
	boolTrue   = true
	i64100     = int64(100)
)

func TestHostConfigSetters(t *testing.T) {
	tests := []struct {
		config   *container.HostConfig
		setFn    func(*Container)
		field    string
		wantErr  bool
		message  string
		expected any
	}{

		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUQuota(100 * time.Millisecond)
			},
			field:    "CPUQuota",
			wantErr:  false,
			message:  "CPUQuota ok 100ms",
			expected: int64(100000),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUQuota(-time.Millisecond)
			},
			field:    "CPUQuota",
			wantErr:  true,
			message:  "CPUQuota error negative",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUQuota(100 * time.Millisecond)
			},
			field:    "CPUQuota",
			wantErr:  false,
			message:  "CPUQuota ok 100ms",
			expected: int64(100000),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.Sysctls("", "1")
			},
			field:    "Sysctls",
			wantErr:  true,
			message:  "Sysctls error empty key",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceReadBps("", 100*types.MiB)
			},
			field:    "BlkioDeviceReadBps",
			wantErr:  true,
			message:  "BlkioDeviceReadBps error empty device path",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceWriteBps("", 100*types.MiB)
			},
			field:    "BlkioDeviceWriteBps",
			wantErr:  true,
			message:  "BlkioDeviceWriteBps error empty device path",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceReadIOps("", 100)
			},
			field:    "BlkioDeviceReadIOps",
			wantErr:  true,
			message:  "BlkioDeviceReadIOps error empty device path",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceWriteIOps("", 100)
			},
			field:    "BlkioDeviceWriteIOps",
			wantErr:  true,
			message:  "BlkioDeviceWriteIOps error empty device path",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceWriteIOps("/dev/sda", 100)
			},
			field:   "BlkioDeviceWriteIOps",
			wantErr: false,
			message: "BlkioDeviceWriteIOps ok 100KiB",
			expected: []*blkiodev.ThrottleDevice{
				{
					Path: "/dev/sda",
					Rate: 100,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceWriteBps("/dev/sda", 100*types.MiB)
			},
			field:   "BlkioDeviceWriteBps",
			wantErr: false,
			message: "BlkioDeviceWriteBps ok 100MiB",
			expected: []*blkiodev.ThrottleDevice{
				{
					Path: "/dev/sda",
					Rate: 100 * 1024 * 1024,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceReadIOps("/dev/sda", 100)
			},
			field:   "BlkioDeviceReadIOps",
			wantErr: false,
			message: "BlkioDeviceReadIOps ok 100KiB",
			expected: []*blkiodev.ThrottleDevice{
				{
					Path: "/dev/sda",
					Rate: 100,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceReadBps("/dev/sda", 100*types.MiB)
			},
			field:   "BlkioDeviceReadBps",
			wantErr: false,
			message: "BlkioDeviceReadBps ok 100MB",
			expected: []*blkiodev.ThrottleDevice{
				{
					Path: "/dev/sda",
					Rate: 100 * 1024 * 1024,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUCount(0)
			},
			field:    "CPUCount",
			wantErr:  true,
			message:  "CPUCount error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ReadonlyPaths("")
			},
			field:    "ReadonlyPaths",
			wantErr:  true,
			message:  "ReadonlyPaths error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ReadonlyPaths("bad/path")
			},
			field:    "ReadonlyPaths",
			wantErr:  true,
			message:  "ReadonlyPaths error bad/path",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.MaskedPaths("")
			},
			field:    "MaskedPaths",
			wantErr:  true,
			message:  "MaskedPaths error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.MaskedPaths("bad/path")
			},
			field:    "MaskedPaths",
			wantErr:  true,
			message:  "MaskedPaths error bad/path",
			expected: nil,
		},

		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.NetworkMode(types.NetworkContainer("123"))
			},
			field:    "NetworkMode",
			wantErr:  false,
			message:  "NetworkMode error container:123",
			expected: container.NetworkMode("container:123"),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.VolumeDriver("")
			},
			field:    "VolumeDriver",
			wantErr:  true,
			message:  "VolumeDriver error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.OomScoreAdj(-1001)
			},
			field:    "OomScoreAdj",
			wantErr:  true,
			message:  "OomScoreAdj error negative",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.StorageOpt("", "value")
			},
			field:    "StorageOpt",
			wantErr:  true,
			message:  "StorageOpt error empty key",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedDevice(types.Device{PathOnHost: "/dev/null", PathInContainer: "/dev/null"})
			},
			field:    "Devices",
			wantErr:  true,
			message:  "AddedDevice error invalid permissions",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedDevice(types.Device{PathOnHost: "/dev/null", PathInContainer: "/dev/null"})
			},
			field:    "Devices",
			wantErr:  true,
			message:  "AddedDevice error duplicate permissions",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedDevice(types.Device{PathOnHost: "/dev/null", PathInContainer: "/dev/null"})
			},
			field:    "Devices",
			wantErr:  true,
			message:  "AddedDevice error empty permissions",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedDevice(types.Device{PathOnHost: "/dev/null", CgroupPermissions: "rw"})
			},
			field:    "Devices",
			wantErr:  true,
			message:  "AddedDevice error empty path",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedDevice(types.Device{PathInContainer: "/dev/null", CgroupPermissions: "rw"})
			},
			field:    "Devices",
			wantErr:  true,
			message:  "AddedDevice error empty device",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedDevice(types.Device{PathOnHost: "/dev/null", PathInContainer: "/dev/null", CgroupPermissions: "rw"})
			},
			field:   "Devices",
			wantErr: false,
			message: "AddedDevice ok",
			expected: []container.DeviceMapping{
				{
					PathOnHost:        "/dev/null",
					PathInContainer:   "/dev/null",
					CgroupPermissions: "rw",
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ContainerIDFile("")
			},
			field:    "ContainerIDFile",
			wantErr:  true,
			message:  "ContainerIDFile error empty",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUShares(0)
			},
			field:    "CPUShares",
			wantErr:  true,
			message:  "CPUShares error zero",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUPeriod(10 * time.Hour)
			},
			field:    "CPUPeriod",
			wantErr:  true,
			message:  "CPUPeriod error value too high",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "CPUPercent",
			setFn: func(c *Container) {
				c.CPUPercent(-1)
			},
			wantErr:  true,
			message:  "CPUPercent error negative",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "Ulimits",
			setFn: func(c *Container) {
				c.Ulimits(types.Ulimit{Name: "nofile", Soft: 100, Hard: 10})
			},
			wantErr:  true,
			message:  "Ulimits soft greater than hard error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "Ulimits",
			setFn: func(c *Container) {
				c.Ulimits(types.Ulimit{Name: "", Soft: 1024, Hard: 1024})
			},
			wantErr:  true,
			message:  "Ulimits error empty name",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "Ulimits",
			setFn: func(c *Container) {
				c.Ulimits(types.Ulimit{Name: "nofile", Soft: -1, Hard: 1024})
			},
			wantErr:  true,
			message:  "Ulimits error negative soft limit",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "CPURealtimePeriod",
			setFn: func(c *Container) {
				c.CPURealtimePeriod(700 * time.Millisecond)
			},
			wantErr:  false,
			message:  "CPURealtimePeriod ok 700ms",
			expected: int64(700 * 1000),
		},
		{
			config: &container.HostConfig{},
			field:  "CPURealtimePeriod",
			setFn: func(c *Container) {
				c.CPURealtimePeriod(time.Microsecond)
			},
			wantErr:  true,
			message:  "CPURealtimePeriod error 1us",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "CPURealtimePeriod",
			setFn: func(c *Container) {
				c.CPURealtimePeriod(time.Second)
			},
			wantErr:  false,
			message:  "CPURealtimePeriod ok 1s",
			expected: int64(1 * 1000 * 1000),
		},
		{
			config: &container.HostConfig{},
			field:  "CPURealtimePeriod",
			setFn: func(c *Container) {
				c.CPURealtimePeriod(999 * time.Millisecond)
			},
			wantErr:  false,
			message:  "CPURealtimePeriod ok 999ms",
			expected: int64(999 * 1000),
		},
		{
			config: &container.HostConfig{},
			field:  "CPURealtimePeriod",
			setFn: func(c *Container) {
				c.CPURealtimePeriod(time.Millisecond)
			},
			wantErr:  false,
			message:  "CPURealtimePeriod ok 1ms",
			expected: int64(1000),
		},
		{
			field: "CPURealtimeRuntime",
			setFn: func(c *Container) {
				c.CPURealtimeRuntime(-2)
			},
			wantErr:  true,
			message:  "CPURealtimeRuntime error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "CpusetCpus",
			setFn: func(c *Container) {
				c.CpusetCpus(mustCpuset("0,1,3-5"))
			},
			wantErr:  false,
			message:  "CpusetCpus ok 0,1,3-5",
			expected: "0,1,3-5",
		},
		{
			config: &container.HostConfig{},
			field:  "BlkioWeight",
			setFn: func(c *Container) {
				c.BlkioWeight(0)
			},
			wantErr:  true,
			message:  "BlkioWeight error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "PidsLimit",
			setFn: func(c *Container) {
				c.PidsLimit(-2)
			},
			wantErr:  true,
			message:  "PidsLimit error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "CpusetMems",
			setFn: func(c *Container) {
				c.MemorySwappiness(-1)
			},
			wantErr:  true,
			message:  "MemorySwappiness error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			field:  "CpusetMems",
			setFn: func(c *Container) {
				c.CpusetMems(mustCpuset("0,1,3-5"))
			},
			wantErr:  false,
			message:  "CpusetMems ok 0,1,3-5",
			expected: "0,1,3-5",
		},
		{
			config: &container.HostConfig{},
			field:  "CpusetMems",
			setFn: func(c *Container) {
				c.CpusetMems(mustCpuset("0,2,3,4"))
			},
			wantErr:  false,
			message:  "CpusetMems ok 0,2,3,4",
			expected: "0,2,3,4",
		},
		{
			config: &container.HostConfig{},
			field:  "MemoryReservation",
			setFn: func(c *Container) {
				c.MemoryReservation(100 * types.MiB)
			},
			wantErr:  false,
			message:  "MemoryReservation ok",
			expected: int64(100 * 1024 * 1024),
		},
		{
			config: &container.HostConfig{},
			field:  "MemorySwap",
			setFn: func(c *Container) {
				c.MemorySwap(100 * types.MiB)
			},
			wantErr:  false,
			message:  "MemorySwap ok",
			expected: int64(100 * 1024 * 1024),
		},
		{
			config: &container.HostConfig{},
			field:  "KernelMemory",
			setFn: func(c *Container) {
				c.KernelMemory(100 * types.MiB)
			},
			wantErr:  false,
			message:  "KernelMemory ok",
			expected: int64(100 * 1024 * 1024),
		},

		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.MemoryLimit(100 * types.MiB)
			},
			field:    "Memory",
			wantErr:  false,
			message:  "MemoryLimit ok",
			expected: int64(100 * 1024 * 1024),
		},

		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ShmSize(100 * types.MiB)
			},
			field:    "ShmSize",
			wantErr:  false,
			message:  "ShmSize ok",
			expected: int64(100 * 1024 * 1024),
		},

		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartAlways()
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartAlways ok",
			expected: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.MemoryLimit(1024)
			},
			field:    "Memory",
			wantErr:  false,
			message:  "MemoryLimit ok",
			expected: int64(1024),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AutoRemove()
			},
			field:    "AutoRemove",
			wantErr:  false,
			message:  "AutoRemove ok",
			expected: true,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "8080"})
			},
			field:   "PortBindings",
			wantErr: false,
			message: "PortBinding ok",
			expected: nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "808012", ContainerPort: ""})
			},
			field:    "PortBindings",
			wantErr:  true,
			message:  "PortBinding error",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "", ContainerPort: ""})
			},
			field:    "PortBindings",
			wantErr:  true,
			message:  "PortBinding empty host and container port",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: "", HostIP: "", HostPort: "", ContainerPort: ""})
			},
			field:    "PortBindings",
			wantErr:  true,
			message:  "PortBinding empty values",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: types.TCP, HostPort: "8080", ContainerPort: "8080"})
			},
			field:   "PortBindings",
			wantErr: false,
			message: "PortBinding empty host IP defaults to 0.0.0.0",
			expected: nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: ""})
			},
			field:    "PortBindings",
			wantErr:  true,
			message:  "PortBinding empty container port",
			expected: nil,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "8080"})
			},
			field:   "PortBindings",
			wantErr: false,
			message: "PortBinding empty protocol defaults to TCP",
			expected: nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "", ContainerPort: "8080"})
			},
			field:    "PortBindings",
			wantErr:  true,
			message:  "PortBinding empty host port",
			expected: nilPortMap,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "invalid"})
			},
			field:    "PortBindings",
			wantErr:  true,
			message:  "PortBinding invalid container port",
			expected: nilPortMap,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "invalid", ContainerPort: "8080"})
			},
			field:    "PortBindings",
			wantErr:  true,
			message:  "PortBinding invalid host port",
			expected: nilPortMap,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.DNSLookups("1.1.1.1", "8.8.8.8")
			},
			field:    "DNS",
			wantErr:  false,
			message:  "DNSLookups ok",
			expected: []string{"1.1.1.1", "8.8.8.8"},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.DNSOptions("dns-option1", "dns-option2")
			},
			field:    "DNSOptions",
			wantErr:  false,
			message:  "DNSOptions ok",
			expected: []string{"dns-option1", "dns-option2"},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.DNSSearches("dns-search1", "dns-search2")
			},
			field:    "DNSSearch",
			wantErr:  false,
			message:  "DNSSearches ok",
			expected: []string{"dns-search1", "dns-search2"},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ExtraHost("test.com", "127.0.0.1")
			},
			field:    "ExtraHosts",
			wantErr:  false,
			message:  "ExtraHost ok",
			expected: []string{"test.com:127.0.0.1"},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedGroups("group1", "group2")
			},
			field:    "GroupAdd",
			wantErr:  false,
			message:  "AddedGroups ok",
			expected: []string{"group1", "group2"},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.UTSMode(types.UTSHost)
			},
			field:    "UTSMode",
			wantErr:  false,
			message:  "UTSMode ok",
			expected: container.UTSMode("host"),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.UserNSMode(types.UserNSHost)
			},
			field:    "UsernsMode",
			wantErr:  false,
			message:  "UserNSMode ok",
			expected: container.UsernsMode("host"),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ShmSize(1024)
			},
			field:    "ShmSize",
			wantErr:  false,
			message:  "ShmSize ok",
			expected: int64(1024),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.Runtime("runc")
			},
			field:    "Runtime",
			wantErr:  false,
			message:  "Runtime ok",
			expected: "runc",
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ConsoleSize(types.ConsoleSize{10, 10})
			},
			field:    "ConsoleSize",
			wantErr:  false,
			message:  "ConsoleSize ok",
			expected: [2]uint{10, 10},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.Isolation(types.IsolationDefault)
			},
			field:    "Isolation",
			wantErr:  false,
			message:  "Isolation ok",
			expected: container.Isolation("default"),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUCount(1)
			},
			field:    "CPUCount",
			wantErr:  false,
			message:  "CPUCount ok",
			expected: int64(1),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ReadonlyPaths("/host/path")
			},
			field:    "ReadonlyPaths",
			wantErr:  false,
			message:  "ReadonlyPaths ok",
			expected: []string{"/host/path"},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.MaskedPaths("/host/path", "/host/path2")
			},
			field:    "MaskedPaths",
			wantErr:  false,
			message:  "MaskedPaths ok",
			expected: []string{"/host/path", "/host/path2"},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.NetworkMode(types.NetworkHost)
			},
			field:    "NetworkMode",
			wantErr:  false,
			message:  "NetworkMode ok",
			expected: container.NetworkMode("host"),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.VolumeDriver("local")
			},
			field:    "VolumeDriver",
			wantErr:  false,
			message:  "VolumeDriver ok",
			expected: "local",
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.VolumesFrom("container1", false)
			},
			field:    "VolumesFrom",
			wantErr:  false,
			message:  "VolumesFrom ok",
			expected: []string{"container1"},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.IpcMode(types.IpcHost)
			},
			field:    "IpcMode",
			wantErr:  false,
			message:  "IpcMode ok",
			expected: container.IpcMode("host"),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.Cgroup(types.CgroupHost)
			},
			field:    "CgroupnsMode",
			wantErr:  false,
			message:  "Cgroup ok",
			expected: container.CgroupnsModeHost,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.OomScoreAdj(100)
			},
			field:    "OomScoreAdj",
			wantErr:  false,
			message:  "OomScoreAdj ok",
			expected: int(100),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.OomKillDisable()
			},
			field:    "OomKillDisable",
			wantErr:  false,
			message:  "OomKillDisable ok",
			expected: &boolTrue,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PidMode(types.PidHost)
			},
			field:    "PidMode",
			wantErr:  false,
			message:  "PidMode ok",
			expected: container.PidMode("host"),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PublishAllPorts()
			},
			field:    "PublishAllPorts",
			wantErr:  false,
			message:  "PublishAllPorts ok",
			expected: boolTrue,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ReadOnlyRootfs()
			},
			field:    "ReadonlyRootfs",
			wantErr:  false,
			message:  "ReadOnlyRootfs ok",
			expected: true,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.SecurityOpts("test1", "test2")
			},
			field:   "SecurityOpt",
			wantErr: false,
			message: "SecurityOpts ok",
			expected: []string{
				"test1",
				"test2",
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.StorageOpt("test1", "test2")
			},
			field:   "StorageOpt",
			wantErr: false,
			message: "StorageOpt ok",
			expected: map[string]string{
				"test1": "test2",
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.Tmpfs("test1", "test2")
			},
			field:   "Tmpfs",
			wantErr: false,
			message: "Tmpfs ok",
			expected: map[string]string{
				"test1": "test2",
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.Privileged()
			},
			field:    "Privileged",
			wantErr:  false,
			message:  "Privileged ok",
			expected: true,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedDevice(types.Device{PathOnHost: "/dev/null", PathInContainer: "/dev/null", CgroupPermissions: "rwm"})
			},
			field:   "Devices",
			wantErr: false,
			message: "AddedDevice ok",
			expected: []container.DeviceMapping{
				{
					PathOnHost:        "/dev/null",
					PathInContainer:   "/dev/null",
					CgroupPermissions: "rwm",
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.ContainerIDFile("/path/to/container-id.txt")
			},
			field:    "ContainerIDFile",
			wantErr:  false,
			message:  "ContainerIDFile ok",
			expected: "/path/to/container-id.txt",
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUShares(100)
			},
			field:    "CPUShares",
			wantErr:  false,
			message:  "CPUShares ok",
			expected: int64(100),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUPeriod(time.Millisecond)
			},
			field:    "CPUPeriod",
			wantErr:  false,
			message:  "CPUPeriod ok",
			expected: int64(1000),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUPeriod(100 * time.Millisecond)
			},
			field:    "CPUPeriod",
			wantErr:  false,
			message:  "CPUPeriod ok",
			expected: int64(100 * 1000),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUQuota(100 * time.Microsecond)
			},
			field:    "CPUQuota",
			wantErr:  false,
			message:  "CPUQuota ok",
			expected: int64(100),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPUPercent(100)
			},
			field:    "CPUPercent",
			wantErr:  false,
			message:  "CPUPercent ok",
			expected: int64(100),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.MemoryReservation(100)
			},
			field:    "MemoryReservation",
			wantErr:  false,
			message:  "MemoryReservation ok",
			expected: int64(100),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.MemorySwap(100)
			},
			field:    "MemorySwap",
			wantErr:  false,
			message:  "MemorySwap ok",
			expected: int64(100),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.Ulimits(types.Ulimit{Name: "nproc", Soft: 100, Hard: 200})
			},
			field:   "Ulimits",
			wantErr: false,
			message: "Ulimits ok",
			expected: []*container.Ulimit{
				{
					Name: "nproc",
					Soft: 100,
					Hard: 200,
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.Init()
			},
			field:    "Init",
			wantErr:  false,
			message:  "Init ok",
			expected: &boolTrue,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPURealtimePeriod(100 * time.Millisecond)
			},
			field:    "CPURealtimePeriod",
			wantErr:  false,
			message:  "CPURealtimePeriod ok",
			expected: int64(100 * 1000),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CPURealtimeRuntime(100)
			},
			field:    "CPURealtimeRuntime",
			wantErr:  false,
			message:  "CPURealtimeRuntime ok",
			expected: int64(100),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CpusetMems(mustCpuset("0,1"))
			},
			field:    "CpusetMems",
			wantErr:  false,
			message:  "CpusetMems ok",
			expected: "0,1",
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.MemorySwappiness(100)
			},
			field:    "MemorySwappiness",
			wantErr:  false,
			message:  "MemorySwappiness ok",
			expected: &i64100,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CpusetCpus(mustCpuset("0,1"))
			},
			field:    "CpusetCpus",
			wantErr:  false,
			message:  "CpusetCpus ok",
			expected: "0,1",
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.KernelMemory(100)
			},
			field:    "KernelMemory",
			wantErr:  false,
			message:  "KernelMemory ok",
			expected: i64100,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.PidsLimit(100)
			},
			field:    "PidsLimit",
			wantErr:  false,
			message:  "PidsLimit ok",
			expected: &i64100,
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioWeight(100)
			},
			field:    "BlkioWeight",
			wantErr:  false,
			message:  "BlkioWeight ok",
			expected: uint16(100),
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.BlkioDeviceReadBps("/dev/0", 100)
				c.BlkioDeviceReadBps("/dev/1", 100)
			},

			field:   "BlkioDeviceReadBps",
			wantErr: false,
			message: "BlkioDeviceReadBps ok",
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
			setFn: func(c *Container) {
				c.BlkioDeviceWriteBps("/dev/0", 100)
				c.BlkioDeviceWriteBps("/dev/1", 100)
			},

			field:   "BlkioDeviceWriteBps",
			wantErr: false,
			message: "BlkioDeviceWriteBps ok",
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
			setFn: func(c *Container) {
				c.BlkioDeviceReadIOps("/dev/0", 100)
				c.BlkioDeviceReadIOps("/dev/1", 100)
			},

			field:   "BlkioDeviceReadIOps",
			wantErr: false,
			message: "BlkioDeviceReadIOps ok",
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
			setFn: func(c *Container) {
				c.BlkioDeviceWriteIOps("/dev/0", 100)
				c.BlkioDeviceWriteIOps("/dev/1", 100)
			},

			field:   "BlkioDeviceWriteIOps",
			wantErr: false,
			message: "BlkioDeviceWriteIOps ok",
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
			setFn: func(c *Container) {
				c.Sysctls("net.ipv4.ip_forward", "1")
			},
			field:   "Sysctls",
			wantErr: false,
			message: "Sysctls ok",
			expected: map[string]string{
				"net.ipv4.ip_forward": "1",
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.DeviceCgroupRules("c 1:3 rwm")
			},
			field:   "DeviceCgroupRules",
			wantErr: false,
			message: "DeviceCgroupRules ok",
			expected: []string{
				"c 1:3 rwm",
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.CgroupParent("test")
			},
			field:    "CgroupParent",
			wantErr:  false,
			message:  "CgroupParent ok",
			expected: "test",
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.DeviceRequest(types.DeviceRequest{Driver: "nvidia", Count: 1, DeviceIDs: []string{"1"}, Capabilities: [][]string{{"gpu"}}})
			},
			field:   "DeviceRequests",
			wantErr: false,
			message: "DeviceRequest ok",
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
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.LogDriver("json-file", map[string]string{"max-size": "10m", "max-file": "3"})
			},
			field:   "LogConfig",
			wantErr: false,
			message: "LogDriver ok",
			expected: container.LogConfig{
				Type: "json-file",
				Config: map[string]string{
					"max-size": "10m",
					"max-file": "3",
				},
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.AddedCapabilities(types.NET_ADMIN, types.SYS_ADMIN)
			},
			field:   "CapAdd",
			wantErr: false,
			message: "AddedCapabilities ok",
			expected: strslice.StrSlice{
				"NET_ADMIN",
				"SYS_ADMIN",
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.DroppedCapabilities(types.NET_ADMIN, types.SYS_ADMIN)
			},
			field:   "CapDrop",
			wantErr: false,
			message: "DroppedCapabilities ok",
			expected: strslice.StrSlice{
				"NET_ADMIN",
				"SYS_ADMIN",
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.DroppedAllCapabilities()
			},
			field:   "CapDrop",
			wantErr: false,
			message: "DroppedAllCapabilities ok",
			expected: strslice.StrSlice{
				"ALL",
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.DroppedSensitiveCapabilities()
			},
			field:   "CapDrop",
			wantErr: false,
			message: "DroppedSensitiveCapabilities ok",
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
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartPolicy(types.RestartPolicyNo, 0)
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartPolicy no ok",
			expected: container.RestartPolicy{
				Name:              "no",
				MaximumRetryCount: 0,
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartPolicy(types.RestartPolicyOnFailure, 10)
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartPolicy on-failure ok",
			expected: container.RestartPolicy{
				Name:              "on-failure",
				MaximumRetryCount: 10,
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartPolicy(types.RestartPolicyUnlessStopped, 10)
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartPolicy ok",
			expected: container.RestartPolicy{
				Name:              "unless-stopped",
				MaximumRetryCount: 10,
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartPolicy(types.RestartPolicyAlways, 10)
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartPolicy ok",
			expected: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 10,
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartPolicy("", 0)
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartPolicy fallback disabled ok",
			expected: container.RestartPolicy{
				Name:              "no",
				MaximumRetryCount: 0,
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartPolicyAlways()
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartPolicy always ok",
			expected: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartPolicyOnFailure(10)
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartPolicy on-failure ok",
			expected: container.RestartPolicy{
				Name:              "on-failure",
				MaximumRetryCount: 10,
			},
		},
		{
			config: &container.HostConfig{},
			setFn: func(c *Container) {
				c.RestartPolicyUnlessStopped()
			},
			field:   "RestartPolicy",
			wantErr: false,
			message: "RestartPolicy unless-stopped ok",
			expected: container.RestartPolicy{
				Name:              "unless-stopped",
				MaximumRetryCount: 0,
			},
		},
	}
	for _, test := range tests {
		if test.config == nil {
			test.config = &container.HostConfig{}
		}
		c := NewContainer("t")
		c.Config.Host = test.config
		test.setFn(c)
		if test.wantErr {
			assert.NotEmpty(t, c.Errors, test.message)
			assert.True(t, errdefs.IsHostConfigError(c.Errors[0]), test.message)
			continue
		}
		assert.Empty(t, c.Errors, test.message)
		if test.field != "" {
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}

func TestVolumeBinds(t *testing.T) {
	c := NewContainer("app").Image("alpine").VolumeBinds(
		"/src:/dst",
		"data:/var/lib:rw",
		"/host:/data:ro,rshared,cached",
		"./rel:/opt:z",
		`C:\data:/app:ro`,
		"/etc/vector/vector.yml:/etc/vector/vector.yml:readonly",
	)
	require.NoError(t, c.Validate())
	require.Empty(t, c.Config.Host.Binds)
	require.Len(t, c.Config.Host.Mounts, 6)

	require.Equal(t, mount.TypeBind, c.Config.Host.Mounts[0].Type)
	require.Equal(t, "/src", c.Config.Host.Mounts[0].Source)
	require.Equal(t, "/dst", c.Config.Host.Mounts[0].Target)
	require.False(t, c.Config.Host.Mounts[0].ReadOnly)

	require.Equal(t, mount.TypeVolume, c.Config.Host.Mounts[1].Type)
	require.Equal(t, "data", c.Config.Host.Mounts[1].Source)
	require.Equal(t, "/var/lib", c.Config.Host.Mounts[1].Target)

	require.True(t, c.Config.Host.Mounts[2].ReadOnly)
	require.Equal(t, mount.ConsistencyCached, c.Config.Host.Mounts[2].Consistency)
	require.Equal(t, mount.PropagationRShared, c.Config.Host.Mounts[2].BindOptions.Propagation)

	require.Equal(t, "./rel", c.Config.Host.Mounts[3].Source)
	require.Equal(t, "z", c.mountNotes["/opt"].selinux)
	require.Equal(t, `C:\data`, c.Config.Host.Mounts[4].Source)
	require.True(t, c.Config.Host.Mounts[4].ReadOnly)
	require.True(t, c.Config.Host.Mounts[5].ReadOnly)

	c.VolumeBinds("/other:/dst:rw")
	require.Error(t, c.Validate())
	var hostErr *errdefs.HostConfigError
	require.ErrorAs(t, c.Errors[0], &hostErr)
	require.Equal(t, "volume_binds", hostErr.Field)
	require.Contains(t, hostErr.Error(), "duplicate target path '/dst'")

	bad := NewContainer("bad").VolumeBinds("/src:/dst:ro,rw", "", ":/missing:ro", "/only")
	require.Error(t, bad.Validate())
	require.Len(t, bad.Errors, 4)
	require.Empty(t, bad.Config.Host.Mounts)

	p := NewProject("p").WithService(NewService("app", NewContainer("svc").Image("alpine").VolumeBinds("/tmp:/opt/bind:ro,z", "data:/var/lib:rw")))
	require.NoError(t, p.Validate())
	svc, err := p.GetService("app")
	require.NoError(t, err)
	var bind, named *composetypes.ServiceVolumeConfig
	for i := range svc.Volumes {
		switch svc.Volumes[i].Target {
		case "/opt/bind":
			bind = &svc.Volumes[i]
		case "/var/lib":
			named = &svc.Volumes[i]
		}
	}
	require.NotNil(t, bind)
	require.Equal(t, "bind", bind.Type)
	require.True(t, bind.ReadOnly)
	require.Equal(t, "z", bind.Bind.SELinux)
	require.NotNil(t, named)
	require.Equal(t, "volume", named.Type)
	require.Equal(t, "data", named.Source)

	kept := NewContainer("kept").PortBinding(types.PortBinding{})
	require.True(t, errdefs.IsHostConfigError(NewProject("p2").WithService(NewService("kept", kept.Image("alpine"))).Validate()))
}
