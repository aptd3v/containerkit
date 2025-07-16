// Package volume provides functions to set the volume configuration for a project
package volume

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

type SetVolumeProjectConfig func(*types.VolumeConfig) error

// WithDriver sets the driver for the volume
// parameters:
//   - driver: the driver for the volume
func WithDriver(driver string) SetVolumeProjectConfig {
	return func(opt *types.VolumeConfig) error {
		opt.Driver = driver
		return nil
	}
}

// WithDriverOption appends a driver option to the  project volume
// parameters:
//   - key: the key of the driver option
//   - value: the value of the driver option
func WithDriverOptions(key, value string) SetVolumeProjectConfig {
	return func(opt *types.VolumeConfig) error {
		if opt.DriverOpts == nil {
			opt.DriverOpts = make(map[string]string)
		}
		opt.DriverOpts[key] = value
		return nil
	}
}

// WithLabel appends a label to the volume
// parameters:
//   - key: the key of the label
//   - value: the value of the label
func WithLabel(key, value string) SetVolumeProjectConfig {
	return func(opt *types.VolumeConfig) error {
		if opt.Labels == nil {
			opt.Labels = make(map[string]string)
		}
		opt.Labels[key] = value
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the volume config
// and append the error to the service config error collection
func Fail(err error) SetVolumeProjectConfig {
	return func(opt *types.VolumeConfig) error {
		return errdefs.NewServiceConfigError("secrets", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the volume config
// and append the error to the service config error collection
func Failf(stringFormat string, args ...any) SetVolumeProjectConfig {
	return func(opt *types.VolumeConfig) error {
		return errdefs.NewServiceConfigError("secrets", fmt.Sprintf(stringFormat, args...))
	}
}
