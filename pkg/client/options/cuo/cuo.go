// package cuo provides options for the container update.
package cuo

import (
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
)

// SetContainerUpdateOption is a function that sets a parameter for the container update.
type SetContainerUpdateOption func(*container.UpdateConfig) error

// WithRestartPolicy sets the container restart policy behaviour.
func WithRestartPolicy(name hc.RestartPolicy, maxRetry int) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.RestartPolicy = container.RestartPolicy{
			Name:              container.RestartPolicyMode(name),
			MaximumRetryCount: maxRetry,
		}
		return nil
	}
}

// WithCPUShares sets the container cpu shares behaviour.
func WithCPUShares(shares int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CPUShares = shares
		return nil
	}
}

// WithMemory sets the container memory behaviour.
func WithMemory(memory int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.Memory = memory
		return nil
	}
}

// WithNanoCPUs sets the container nano cpus behaviour.
func WithNanoCPUs(nanoCPUs int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.NanoCPUs = nanoCPUs
		return nil
	}
}

// WithCgroupParent sets the container cgroup parent behaviour.
func WithCgroupParent(cgroupParent string) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CgroupParent = cgroupParent
		return nil
	}
}

// WithBlkioWeight sets the container blkio weight behaviour.
func WithBlkioWeight(weight uint16) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.BlkioWeight = weight
		return nil
	}
}

// WithBlkioWeightDevice sets the container blkio weight device behaviour.
func WithBlkioWeightDevice(path string, weight uint16) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		if o.BlkioWeightDevice == nil {
			o.BlkioWeightDevice = []*blkiodev.WeightDevice{}
		}
		o.BlkioWeightDevice = append(o.BlkioWeightDevice, &blkiodev.WeightDevice{Path: path, Weight: weight})
		return nil
	}
}

// WithBlkioDeviceReadBps sets the container blkio device read bps behaviour.
func WithBlkioDeviceReadBps(path string, rate uint64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		if o.BlkioDeviceReadBps == nil {
			o.BlkioDeviceReadBps = []*blkiodev.ThrottleDevice{}
		}
		o.BlkioDeviceReadBps = append(o.BlkioDeviceReadBps, &blkiodev.ThrottleDevice{Path: path, Rate: rate})
		return nil
	}
}

// WithBlkioDeviceWriteBps sets the container blkio device write bps behaviour.
func WithBlkioDeviceWriteBps(path string, rate uint64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		if o.BlkioDeviceWriteBps == nil {
			o.BlkioDeviceWriteBps = []*blkiodev.ThrottleDevice{}
		}
		o.BlkioDeviceWriteBps = append(o.BlkioDeviceWriteBps, &blkiodev.ThrottleDevice{Path: path, Rate: rate})
		return nil
	}
}

// WithBlkioDeviceReadIOps sets the container blkio device read iops behaviour.
func WithBlkioDeviceReadIOps(path string, rate uint64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		if o.BlkioDeviceReadIOps == nil {
			o.BlkioDeviceReadIOps = []*blkiodev.ThrottleDevice{}
		}
		o.BlkioDeviceReadIOps = append(o.BlkioDeviceReadIOps, &blkiodev.ThrottleDevice{Path: path, Rate: rate})
		return nil
	}
}

// WithBlkioDeviceWriteIOps sets the container blkio device write iops behaviour.
func WithBlkioDeviceWriteIOps(path string, rate uint64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		if o.BlkioDeviceWriteIOps == nil {
			o.BlkioDeviceWriteIOps = []*blkiodev.ThrottleDevice{}
		}
		o.BlkioDeviceWriteIOps = append(o.BlkioDeviceWriteIOps, &blkiodev.ThrottleDevice{Path: path, Rate: rate})
		return nil
	}
}

// WithCPUPeriod sets the container cpu period behaviour.
func WithCPUPeriod(period int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CPUPeriod = period
		return nil
	}
}

// WithCPUQuota sets the container cpu quota behaviour.
func WithCPUQuota(quota int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CPUQuota = quota
		return nil
	}
}

// WithCPURealtimePeriod sets the container cpu realtime period behaviour.
func WithCPURealtimePeriod(period int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CPURealtimePeriod = period
		return nil
	}
}

// WithCPURealtimeRuntime sets the container cpu realtime runtime behaviour.
func WithCPURealtimeRuntime(runtime int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CPURealtimeRuntime = runtime
		return nil
	}
}

// WithCpusetCpus sets the container cpuset cpus behaviour.
func WithCpusetCpus(cpusetCpus string) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CpusetCpus = cpusetCpus
		return nil
	}
}

// WithCpusetMems sets the container cpuset mems behaviour.
func WithCpusetMems(cpusetMems string) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CpusetMems = cpusetMems
		return nil
	}
}

// WithDevices appends the devices for the container.
func WithDevices(pathOnHost, pathInContainer, cgroupPermissions string) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		if o.Devices == nil {
			o.Devices = make([]container.DeviceMapping, 0)
		}
		o.Devices = append(o.Devices, container.DeviceMapping{
			PathOnHost:        pathOnHost,
			PathInContainer:   pathInContainer,
			CgroupPermissions: cgroupPermissions,
		})
		return nil
	}
}

// WithDeviceCgroupRules appends the device cgroup rules for the container.
func WithDeviceCgroupRules(rules []string) SetContainerUpdateOption {

	return func(o *container.UpdateConfig) error {
		o.DeviceCgroupRules = rules
		return nil
	}
}

// WithDeviceRequests appends the device requests for the container.
func WithDeviceRequests(driver string, count int, deviceIDs []string, capabilities [][]string) SetContainerUpdateOption {

	return func(o *container.UpdateConfig) error {
		if o.DeviceRequests == nil {
			o.DeviceRequests = make([]container.DeviceRequest, 0)
		}
		o.DeviceRequests = append(o.DeviceRequests, container.DeviceRequest{
			Driver:       driver,
			Count:        count,
			DeviceIDs:    deviceIDs,
			Capabilities: capabilities,
		})
		return nil
	}
}

// WithKernelMemory sets the container kernel memory behaviour.
//
// Deprecated kernel 5.4 deprecated kmem.limit_in_bytes.
func WithKernelMemory(kernelMemory int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.KernelMemory = kernelMemory
		return nil
	}
}

// WithKernelMemoryTCP sets the container kernel memory TCP behaviour
// Kernel memory TCP limit (in bytes)
func WithKernelMemoryTCP(kernelMemoryTCP int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.KernelMemoryTCP = kernelMemoryTCP
		return nil
	}
}

// WithMemoryReservation sets the container memory reservation
// Memory soft limit (in bytes)
func WithMemoryReservation(memoryReservation int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.MemoryReservation = memoryReservation
		return nil
	}
}

// WithMemorySwap sets the container memory swap behaviour
// Total memory usage (memory + swap); set `-1` to enable unlimited swap
func WithMemorySwap(memorySwap int64) SetContainerUpdateOption {

	return func(o *container.UpdateConfig) error {
		o.MemorySwap = memorySwap
		return nil
	}
}

// WithMemorySwappiness sets the container memory swappiness behaviour
func WithMemorySwappiness(memorySwappiness int64) SetContainerUpdateOption {

	return func(o *container.UpdateConfig) error {
		o.MemorySwappiness = &memorySwappiness
		return nil
	}
}

// WithOomKillDisable sets the OOM Whether to disable OOM Killer or not
func WithOomKillDisable() SetContainerUpdateOption {
	oomKillDisable := true
	return func(o *container.UpdateConfig) error {
		o.OomKillDisable = &oomKillDisable
		return nil
	}
}

// WithPidsLimit sets the maximum number of processes for the container.
// Set `0` or `-1` for unlimited, or `null` to not change.
func WithPidsLimit(pidsLimit int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.PidsLimit = &pidsLimit
		return nil
	}
}

// WithUlimits appends the ulimits for the container.
func WithUlimits(name string, soft, hard int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		if o.Ulimits == nil {
			o.Ulimits = []*container.Ulimit{}
		}
		o.Ulimits = append(o.Ulimits, &container.Ulimit{
			Name: name,
			Soft: soft,
			Hard: hard,
		})
		return nil
	}
}

// WithCPUCount sets the CPU count for the container. applicable to windows.
func WithCPUCount(cpuCount int64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.CPUCount = cpuCount
		return nil
	}
}

// WithCPUPercent sets the CPU percentage for the container.
func WithCPUPercent(cpuPercent int64) SetContainerUpdateOption {

	return func(o *container.UpdateConfig) error {
		o.CPUPercent = cpuPercent
		return nil
	}
}

// WithIOMaximumIOps sets the maximum IO in IO per second for the container system drive.
func WithIOMaximumIOps(ioMaximumIOps uint64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.IOMaximumIOps = ioMaximumIOps
		return nil
	}
}

// WithIOMaximumBandwidth sets the maximum IO in bytes per second for the container system drive.
func WithIOMaximumBandwidth(ioMaximumBandwidth uint64) SetContainerUpdateOption {
	return func(o *container.UpdateConfig) error {
		o.IOMaximumBandwidth = ioMaximumBandwidth
		return nil
	}
}
