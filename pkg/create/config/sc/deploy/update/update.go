// Package update provides functions to set the update configuration for a service deploy
package update

import (
	"fmt"
	"time"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

type SetUpdateConfig func(opt *types.UpdateConfig) error

// WithParallelism sets the parallelism for the update
// parameters:
//   - parallelism: the parallelism for the update
func WithParallelism(parallelism uint64) SetUpdateConfig {
	return func(opt *types.UpdateConfig) error {
		opt.Parallelism = &parallelism
		return nil
	}
}

// WithDelay sets the delay for the update
// parameters:
//   - delay: the delay in seconds for the update
func WithDelay(delay uint64) SetUpdateConfig {
	return func(opt *types.UpdateConfig) error {
		opt.Delay = types.Duration(time.Duration(delay) * time.Second)
		return nil
	}
}

// WithFailureAction sets the failure action for the update
// parameters:
//   - failureAction: the failure action for the update
func WithFailureAction(failureAction string) SetUpdateConfig {
	return func(opt *types.UpdateConfig) error {
		opt.FailureAction = failureAction
		return nil
	}
}

// WithMonitor sets the monitor for the update
// parameters:
//   - monitor: the monitor in seconds for the update
func WithMonitor(monitor uint64) SetUpdateConfig {
	return func(opt *types.UpdateConfig) error {
		opt.Monitor = types.Duration(time.Duration(monitor) * time.Second)
		return nil
	}
}

// WithMaxFailureRatio sets the max failure ratio for the update
// parameters:
//   - ratio: the max failure ratio for the update
func WithMaxFailureRatio(ratio float32) SetUpdateConfig {
	return func(opt *types.UpdateConfig) error {
		opt.MaxFailureRatio = ratio
		return nil
	}
}

// WithOrder sets the order for the update
// parameters:
//   - order: the order for the update
func WithOrder(order string) SetUpdateConfig {
	return func(opt *types.UpdateConfig) error {
		opt.Order = order
		return nil
	}
}

// Fail is a function that returns a setter that always returns the given error
//
// note: this is useful for when you want to fail the update config
// and append the error to the service config error collection
func Fail(err error) SetUpdateConfig {
	return func(opt *types.UpdateConfig) error {
		return errdefs.NewServiceConfigError("update", err.Error())
	}
}

// Failf is a function that returns a setter that always returns the given error
//
// note: this is useful for when you want to fail the update config
// and append the error to the service config error collection
func Failf(stringFormat string, args ...any) SetUpdateConfig {
	return func(opt *types.UpdateConfig) error {
		return errdefs.NewServiceConfigError("update", fmt.Sprintf(stringFormat, args...))
	}
}
