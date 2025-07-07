package health

import (
	"fmt"
	"time"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/docker/docker/api/types/container"
)

type SetHealthcheckConfig func(opt *container.HealthConfig) error

// WithkStartPeriod sets the start period for the health check
// parameters:
//   - startPeriod: the start period for the health check in seconds
func WithStartPeriod(startPeriod int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if startPeriod < 0 {
			return create.NewContainerConfigError("healthcheck", "start_period must be greater than 0")
		}
		opt.StartPeriod = time.Duration(startPeriod) * time.Second
		return nil
	}
}

// WithTimeout sets the timeout for the health check
// parameters:
//   - timeout: the timeout for the health check in seconds
func WithTimeout(timeout int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if timeout < 0 {
			return create.NewContainerConfigError("healthcheck", "timeout must be greater than 0")
		}
		opt.Timeout = time.Duration(timeout) * time.Second
		return nil
	}
}

// WithInterval sets the interval for the health check
// parameters:
//   - interval: the interval for the health check in seconds
func WithInterval(interval int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if interval < 0 {
			return create.NewContainerConfigError("healthcheck", "interval must be greater than or equal to 0")
		}
		opt.Interval = time.Duration(interval) * time.Second
		return nil
	}
}

// WithRetries sets the number of retries for the health check
// parameters:
//   - retries: the number of retries for the health check
func WithRetries(retries int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if retries < 0 {
			return create.NewContainerConfigError("healthcheck", "retries must be greater than or equal to 0")
		}
		opt.Retries = retries
		return nil
	}
}

// WithTest appends the test for the health check
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
		return create.NewContainerConfigError("healthcheck", err.Error())
	}
}

// Failf is a function that returns an error with a formatted string
//
// note: this is useful for when you want to fail the health check
// and append the error to the container config error collection
func Failf(stringFormat string, args ...interface{}) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		return create.NewContainerConfigError("healthcheck", fmt.Sprintf(stringFormat, args...))
	}
}
