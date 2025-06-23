// Package sc provides functions to set the service config
package sc

import (
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/compose-spec/compose-go/types"
)

type WatchAction string

const (
	WatchActionSync        WatchAction = "sync"
	WatchActionRebuild     WatchAction = "rebuild"
	WatchActionSyncRestart WatchAction = "sync+restart"
)

// WithAnnotation sets an annotation for the service
func WithAnnotation(key, value string) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if config.Annotations == nil {
			config.Annotations = make(types.Mapping)
		}
		config.Annotations[key] = value
		return nil
	}
}

// WithDevelop sets the develop config for the service
// parameters:
//   - action: the action to take when the watch path changes
//   - watchPath: the path to watch
//   - target: the target to sync to
//   - ignorePaths: the paths to ignore
//
// note: this only works with the --watch flag in the compose cli
func WithDevelop(action WatchAction, watchPath string, target string, ignorePaths ...string) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		config.Develop = &types.DevelopConfig{
			Watch: []types.Trigger{
				{
					Path:   watchPath,
					Action: types.WatchAction(action),
					Target: target,
					Ignore: ignorePaths,
				},
			},
		}
		return nil
	}
}

// WithDependsOn appends the depends on config for the service
// parameters:
//   - service: the service to depend on
//   - condition: the condition to depend on
//   - restart: whether to restart the service if the dependency is not met
//   - required: whether the service is required to start
func WithDependsOn(service string) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if config.DependsOn == nil {
			config.DependsOn = make(types.DependsOnConfig, 0)
		}
		config.DependsOn[service] = types.ServiceDependency{
			Condition: "service_started",
			Restart:   true,
			Required:  true,
		}
		return nil
	}
}
