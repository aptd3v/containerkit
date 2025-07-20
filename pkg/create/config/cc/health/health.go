package health

import (
	"fmt"
	"time"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/docker/docker/api/types/container"
)

type SetHealthcheckConfig func(opt *container.HealthConfig) error

// WithStartPeriod sets the start period for the health check.
//
// Accepts either:
//   - int: interpreted as seconds
//   - string: a valid time.ParseDuration string (e.g. "10s", "1m")
//
// A duration of 0 is allowed and disables the start delay.
// Negative values will return an error.
func WithStartPeriod[T ~int | string](startPeriod T) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		var duration time.Duration
		var err error
		switch v := any(startPeriod).(type) {
		case int:
			duration = time.Duration(v) * time.Second
		case string:
			duration, err = time.ParseDuration(v)
			if err != nil {
				return errdefs.NewContainerConfigError("healthcheck", fmt.Sprintf("error parsing start period: %s", err))
			}
		}
		if duration < 0 {
			return errdefs.NewContainerConfigError("healthcheck", "start period must be non-negative")
		}
		opt.StartPeriod = duration
		return nil
	}
}

// WithTimeout sets the timeout for the health check
//
// Accepts either:
//   - int: interpreted as seconds
//   - string: a valid time.ParseDuration string (e.g. "10s", "1m")
//
// A duration of 0 is allowed and disables the timeout.
// Negative values will return an error.
func WithTimeout[T ~int | string](timeout T) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		var duration time.Duration
		var err error
		switch v := any(timeout).(type) {
		case int:
			duration = time.Duration(v) * time.Second
		case string:
			duration, err = time.ParseDuration(v)
			if err != nil {
				return errdefs.NewContainerConfigError("healthcheck", fmt.Sprintf("error parsing timeout: %s", err))
			}
		}
		if duration < 0 {
			return errdefs.NewContainerConfigError("healthcheck", "timeout must be non-negative")
		}
		opt.Timeout = duration
		return nil
	}
}

// WithInterval sets the interval for the health check
//
// Accepts either:
//   - int: interpreted as seconds
//   - string: a valid time.ParseDuration string (e.g. "10s", "1m")
//
// A duration of 0 is allowed and disables the interval.
// Negative values will return an error.
func WithInterval[T ~int | string](interval T) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		var duration time.Duration
		var err error
		switch v := any(interval).(type) {
		case int:
			duration = time.Duration(v) * time.Second
		case string:
			duration, err = time.ParseDuration(v)
			if err != nil {
				return errdefs.NewContainerConfigError("healthcheck", fmt.Sprintf("error parsing interval: %s", err))
			}
		}
		if duration < 0 {
			return errdefs.NewContainerConfigError("healthcheck", "interval must be non-negative")
		}
		opt.Interval = duration
		return nil
	}
}

// WithRetries sets the number of retries for the health check
// parameters:
//   - retries: the number of retries for the health check
func WithRetries(retries int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if retries < 0 {
			return errdefs.NewContainerConfigError("healthcheck", "retries must be greater than or equal to 0")
		}
		opt.Retries = retries
		return nil
	}
}

// WithTest appends the test slice for the health check
// parameters:
//   - test: the test for the health check
func WithTest(test ...string) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if len(opt.Test) == 0 {
			opt.Test = []string{}
		}
		opt.Test = append(opt.Test, test...)
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the health check
// and append the error to the container config error collection
func Fail(err error) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		return errdefs.NewContainerConfigError("healthcheck", err.Error())
	}
}

// Failf is a function that returns an error with a formatted string
//
// note: this is useful for when you want to fail the health check
// and append the error to the container config error collection
func Failf(stringFormat string, args ...any) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		return errdefs.NewContainerConfigError("healthcheck", fmt.Sprintf(stringFormat, args...))
	}
}
