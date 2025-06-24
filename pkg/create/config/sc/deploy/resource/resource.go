// Package resource provides functions to set the resource configuration for a service deploy
package resource

import (
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource/device"
	"github.com/compose-spec/compose-go/types"
)

type SetResourceConfig func(opt *types.Resource) error

// WithNanoCPUs sets the number of nano CPUs for the resource
// parameters:
//   - nanoCPUs: the number of nano CPUs for the resource
func WithNanoCPUs(nanoCPUs string) SetResourceConfig {
	return func(opt *types.Resource) error {
		opt.NanoCPUs = nanoCPUs
		return nil
	}
}

// WithMemoryBytes sets the memory bytes for the resource
// parameters:
//   - bytes: the memory bytes for the resource
func WithMemoryBytes(bytes uint64) SetResourceConfig {
	return func(opt *types.Resource) error {
		opt.MemoryBytes = types.UnitBytes(bytes)
		return nil
	}
}

// WithPids sets the number of pids for the resource
// parameters:
//   - pids: the number of pids for the resource
func WithPids(pids int64) SetResourceConfig {
	return func(opt *types.Resource) error {
		opt.Pids = pids
		return nil
	}
}

// WithDevice appends the device to the resource
// parameters:
//   - setters: the setters for the device
func WithDevice(setters ...device.SetDeviceConfig) SetResourceConfig {
	return func(opt *types.Resource) error {
		if len(setters) == 0 {
			return nil
		}
		if opt.Devices == nil {
			opt.Devices = []types.DeviceRequest{}
		}
		device := types.DeviceRequest{}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(&device); err != nil {
				return err
			}
		}
		opt.Devices = append(opt.Devices, device)
		return nil
	}
}

// WithGenericResource appends the generic resource to the resource
// parameters:
//   - kind: the kind of the generic resource
//   - value: the value of the generic resource
func WithGenericResource(kind string, value int64) SetResourceConfig {
	return func(opt *types.Resource) error {
		if opt.GenericResources == nil {
			opt.GenericResources = []types.GenericResource{}
		}
		generic := []types.GenericResource{
			{
				DiscreteResourceSpec: &types.DiscreteGenericResource{
					Kind:  kind,
					Value: value,
				},
			},
		}
		opt.GenericResources = append(opt.GenericResources, generic...)
		return nil
	}
}
