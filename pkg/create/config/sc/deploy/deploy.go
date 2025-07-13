// Package deploy provides functions to set the deploy configuration for a service
package deploy

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/update"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

type SetDeployConfig func(opt *types.DeployConfig) error

// WithMode sets the deploy mode
// parameters:
//   - mode: the deploy mode
func WithMode(mode string) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		opt.Mode = mode
		return nil
	}
}

// WithReplicas sets the number of replicas for the service
// parameters:
//   - replicas: the number of replicas for the service
func WithReplicas(replicas int) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		opt.Replicas = &replicas
		return nil
	}
}

// WithLabel appends a label for the service
// parameters:
//   - key: the key of the label
//   - value: the value of the label
func WithLabel(key, value string) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		if opt.Labels == nil {
			opt.Labels = make(map[string]string)
		}
		opt.Labels[key] = value
		return nil
	}
}

// WithUpdateConfig sets the update configuration for the service
// parameters:
//   - setters: the setters for the update configuration
func WithUpdateConfig(setters ...update.SetUpdateConfig) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		if len(setters) == 0 {
			return nil
		}
		if opt.UpdateConfig == nil {
			opt.UpdateConfig = &types.UpdateConfig{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(opt.UpdateConfig); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithRollbackConfig sets the rollback configuration for the service
// parameters:
//   - setters: the setters for the rollback configuration
func WithRollbackConfig(setters ...update.SetUpdateConfig) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		if len(setters) == 0 {
			return nil
		}
		if opt.RollbackConfig == nil {
			opt.RollbackConfig = &types.UpdateConfig{}
		}

		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(opt.RollbackConfig); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithResourceLimits sets the resource limits for the service
// parameters:
//   - setters: the setters for the resource limits
func WithResourceLimits(setters ...resource.SetResourceConfig) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		if len(setters) == 0 {
			return nil
		}
		if opt.Resources.Limits == nil {
			opt.Resources.Limits = &types.Resource{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(opt.Resources.Limits); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithResourceReservations sets the resource reservations for the service
// parameters:
//   - setters: the setters for the resource reservations
func WithResourceReservations(setters ...resource.SetResourceConfig) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		if len(setters) == 0 {
			return nil
		}
		if opt.Resources.Reservations == nil {
			opt.Resources.Reservations = &types.Resource{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(opt.Resources.Reservations); err != nil {
				return err
			}
		}
		return nil
	}
}

// Fail is a function that returns a setter that always returns the given error
//
// note: this is useful for when you want to fail the deploy config
// and append the error to the service config error collection
func Fail(err error) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		return errdefs.NewServiceConfigError("deploy", err.Error())
	}
}

// Failf is a function that returns a setter that always returns the given error
//
// note: this is useful for when you want to fail the deploy config
// and append the error to the service config error collection
func Failf(stringFormat string, args ...any) SetDeployConfig {
	return func(opt *types.DeployConfig) error {
		return errdefs.NewServiceConfigError("deploy", fmt.Sprintf(stringFormat, args...))
	}
}
