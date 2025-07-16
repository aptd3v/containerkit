// Package ulimit provides a set of functions to configure the ulimits for the build
package ulimit

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

type SetUlimitConfig func(*types.UlimitsConfig) error

// WithSingle sets the single for the ulimit
// parameters:
//   - value: the value for the ulimit
func WithSingle(value int) SetUlimitConfig {
	return func(opt *types.UlimitsConfig) error {
		opt.Single = value
		return nil
	}
}

// WithSoft sets the soft for the ulimit
// parameters:
//   - value: the value for the ulimit
func WithSoft(value int) SetUlimitConfig {
	return func(opt *types.UlimitsConfig) error {
		opt.Soft = value
		return nil
	}
}

// WithHard sets the hard for the ulimit
// parameters:
//   - value: the value for the ulimit
func WithHard(value int) SetUlimitConfig {
	return func(opt *types.UlimitsConfig) error {
		opt.Hard = value
		return nil
	}
}

// Fail returns a SetUlimitConfig that returns an error
// parameters:
//   - err: the error to return
//
// note: this is useful for when you want to fail the ulimit config
// and append the error to the service config error collection
func Fail(err error) SetUlimitConfig {
	return func(opt *types.UlimitsConfig) error {
		return errdefs.NewServiceConfigError("ulimit", err.Error())
	}
}

// Failf returns a SetUlimitConfig that returns an error
// parameters:
//   - stringFormat: the string format to return
//   - args: the args to format the string with
//
// note: this is useful for when you want to fail the ulimit config
// and append the error to the service config error collection
func Failf(stringFormat string, args ...any) SetUlimitConfig {
	return func(opt *types.UlimitsConfig) error {
		return errdefs.NewServiceConfigError("ulimit", fmt.Sprintf(stringFormat, args...))
	}
}
