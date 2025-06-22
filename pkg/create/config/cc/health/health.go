package health

import (
	"time"

	"github.com/docker/docker/api/types/container"
)

type SetHealthcheckConfig func(opt *container.HealthConfig) error

// WithHealthCheckStartPeriod sets the start period for the health check
// parameters:
//   - startPeriod: the start period for the health check in seconds
func WithHealthCheckStartPeriod(startPeriod int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		opt.StartPeriod = time.Duration(startPeriod) * time.Second
		return nil
	}
}

// WithHealthCheckTimeout sets the timeout for the health check
// parameters:
//   - timeout: the timeout for the health check in seconds
func WithHealthCheckTimeout(timeout int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if opt == nil {
			opt = &container.HealthConfig{}
		}
		opt.Timeout = time.Duration(timeout) * time.Second
		return nil
	}
}

// WithHealthCheckInterval sets the interval for the health check
// parameters:
//   - interval: the interval for the health check in seconds
func WithHealthCheckInterval(interval int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if opt == nil {
			opt = &container.HealthConfig{}
		}
		opt.Interval = time.Duration(interval) * time.Second
		return nil
	}
}

// WithHealthCheckRetries sets the number of retries for the health check
// parameters:
//   - retries: the number of retries for the health check
func WithHealthCheckRetries(retries int) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if opt == nil {
			opt = &container.HealthConfig{}
		}
		opt.Retries = retries
		return nil
	}
}

// WithHealthCheckTest appends the test for the health check
// parameters:
//   - test: the test for the health check
func WithHealthCheckTest(test ...string) SetHealthcheckConfig {
	return func(opt *container.HealthConfig) error {
		if opt == nil {
			opt = &container.HealthConfig{}
		}
		if len(opt.Test) == 0 {
			opt.Test = []string{}
		}
		opt.Test = append(opt.Test, test...)
		return nil
	}
}
