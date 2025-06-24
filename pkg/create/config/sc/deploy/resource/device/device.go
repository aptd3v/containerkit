// Package device provides functions to set the device configuration for a service deploys resource
package device

import "github.com/compose-spec/compose-go/types"

type SetDeviceConfig func(opt *types.DeviceRequest) error

// WithCapabilities appends the capabilities for the device
// parameters:
//   - capabilities: the capabilities for the device
func WithCapabilities(capabilities ...string) SetDeviceConfig {
	return func(opt *types.DeviceRequest) error {
		if opt.Capabilities == nil {
			opt.Capabilities = []string{}
		}
		opt.Capabilities = append(opt.Capabilities, capabilities...)
		return nil
	}
}

// WithDriver sets the driver for the device
// parameters:
//   - driver: the driver for the device
func WithDriver(driver string) SetDeviceConfig {
	return func(opt *types.DeviceRequest) error {
		opt.Driver = driver
		return nil
	}
}

// WithCount sets the count for the device
// parameters:
//   - count: the count for the device
func WithCount(count int64) SetDeviceConfig {
	return func(opt *types.DeviceRequest) error {
		opt.Count = types.DeviceCount(count)
		return nil
	}
}

// WithIDs sets the ids for the device
// parameters:
//   - ids: the ids for the device
func WithIDs(ids ...string) SetDeviceConfig {
	return func(opt *types.DeviceRequest) error {
		if opt.IDs == nil {
			opt.IDs = []string{}
		}
		opt.IDs = append(opt.IDs, ids...)
		return nil
	}
}
