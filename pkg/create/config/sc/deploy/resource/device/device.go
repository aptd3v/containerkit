// Package device provides functions to set the device configuration for a service deploys resource
package device

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

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

// Failf is a function that returns a setter that always returns the given error
//
// note: this is useful for when you want to fail the device config
// and append the error to the service config error collection
func Failf(stringFormat string, args ...any) SetDeviceConfig {
	return func(opt *types.DeviceRequest) error {
		return errdefs.NewServiceConfigError("device", fmt.Sprintf(stringFormat, args...))
	}
}

// Fail is a function that returns a setter that always returns the given error
//
// note: this is useful for when you want to fail the device config
// and append the error to the service config error collection
func Fail(err error) SetDeviceConfig {
	return func(opt *types.DeviceRequest) error {
		return errdefs.NewServiceConfigError("device", err.Error())
	}
}
