// Package pc provides the options for the platform config.
package pc

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// WithArchitecture sets the architecture for the platform
// Parameter:
//   - architecture: the architecture to be used for the platform
//
// Architecture field specifies the CPU architecture, for example `amd64` or `ppc64le`.
func WithArchitecture(architecture string) create.SetPlatformConfig {
	return func(options *ocispec.Platform) error {
		options.Architecture = architecture
		return nil
	}
}

// WithOS sets the OS for the platform
// Parameter:
//   - OS: the OS to be used for the platform
//
// OS specifies the operating system, for example `linux` or `windows`.
func WithOS(OS string) create.SetPlatformConfig {
	return func(options *ocispec.Platform) error {
		options.OS = OS
		return nil
	}
}

// WithOSVersion sets the OS version for the platform
// Parameter:
//   - OSVersion: the OS version to be used for the platform
//
// OSVersion is an optional field specifying the operating system version, for example on Windows `10.0.14393.1066`.
func WithOSVersion(OSVersion string) create.SetPlatformConfig {
	return func(options *ocispec.Platform) error {
		options.OSVersion = OSVersion
		return nil
	}
}

// WithOSFeatures sets the OS features for the platform
// Parameter:
//   - OSFeatures: the OS features to be used for the platform
//
// OSFeatures is an optional field specifying an array of strings, each listing a required OS feature (for example on Windows `win32k`).
func WithOSFeatures(OSFeatures ...string) create.SetPlatformConfig {
	return func(options *ocispec.Platform) error {
		if options.OSFeatures == nil {
			options.OSFeatures = make([]string, 0)
		}
		options.OSFeatures = append(options.OSFeatures, OSFeatures...)
		return nil
	}
}

// WithVariant sets the variant for the platform
// Parameter:
//   - variant: the variant to be used for the platform
//
// Variant is an optional field specifying a variant of the CPU, for example `v7` to specify ARMv7 when architecture is `arm`.
func WithVariant(variant string) create.SetPlatformConfig {
	return func(options *ocispec.Platform) error {
		options.Variant = variant
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the platform config
// and append the error to the platform config error collection
func Fail(err error) create.SetPlatformConfig {
	return func(options *ocispec.Platform) error {
		return errdefs.NewPlatformConfigError("platform_config", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the platform config
// and append the error to the platform config error collection
func Failf(stringFormat string, args ...any) create.SetPlatformConfig {
	return func(options *ocispec.Platform) error {
		return errdefs.NewPlatformConfigError("platform_config", fmt.Sprintf(stringFormat, args...))
	}
}
