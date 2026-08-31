package cc

import (
	"fmt"

	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/cc/health"
	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/docker/docker/api/types/container"
)

// WithHealthCheck sets the health check for the container
// parameters:
//   - healthCheckFns: the health check functions to set
func WithHealthCheck(setters ...health.SetHealthcheckConfig) func(*container.Config) error {
	return func(opt *container.Config) error {
		if opt.Healthcheck == nil {
			opt.Healthcheck = &container.HealthConfig{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(opt.Healthcheck); err != nil {
				return errdefs.NewContainerConfigError("healthcheck", fmt.Sprintf("failed to set health check: %s", err))
			}
		}
		if len(opt.Healthcheck.Test) == 0 {
			// docker ignores healthcheck if test is empty, so we set it to NONE as default
			opt.Healthcheck.Test = []string{"NONE"}
			return nil
		}
		return nil
	}
}

// WithDisabledHealthCheck disables the health check by setting it to NONE.
func WithDisabledHealthCheck() func(*container.Config) error {
	return WithHealthCheck(
		health.WithTest("NONE"),
	)
}
