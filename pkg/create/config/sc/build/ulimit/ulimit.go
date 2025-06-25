// Package ulimit provides a set of functions to configure the ulimits for the build
package ulimit

import "github.com/compose-spec/compose-go/v2/types"

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
