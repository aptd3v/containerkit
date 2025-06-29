// Package sc provides functions to set the service config
package sc

import (
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/build"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/secrets/secretservice"
	"github.com/compose-spec/compose-go/v2/types"
)

// WatchAction is the action to take when the watch path changes
type WatchAction string

const (
	WatchActionSync        WatchAction = "sync"
	WatchActionRebuild     WatchAction = "rebuild"
	WatchActionSyncRestart WatchAction = "sync+restart"
)

// WithNoAttach sets the attach option to false for the service
func WithNoAttach() create.SetServiceConfig {
	attach := false
	return func(config *types.ServiceConfig) error {
		config.Attach = &attach
		return nil
	}
}

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

// WithDependsOn appends the depends on config for the service
// parameters:
//   - service: the service to depend on
func WithDependsOnHealthy(service string) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if config.DependsOn == nil {
			config.DependsOn = make(types.DependsOnConfig, 0)
		}
		config.DependsOn[service] = types.ServiceDependency{
			Condition: "service_healthy",
			Restart:   true,
			Required:  true,
		}
		return nil
	}
}

// WithEnvFile appends the env file paths for the service
// parameters:
//   - path: the path to the env file
func WithEnvFile(path string) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if config.Environment == nil {
			config.EnvFiles = make([]types.EnvFile, 0)
		}
		config.EnvFiles = append(config.EnvFiles, types.EnvFile{
			Path:     path,
			Required: true,
		})
		return nil
	}
}

// WithDeploy sets the deploy config for the service
// parameters:
//   - setters: the setters for the deploy config
func WithDeploy(setters ...deploy.SetDeployConfig) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if len(setters) == 0 {
			return nil
		}
		if config.Deploy == nil {
			config.Deploy = &types.DeployConfig{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(config.Deploy); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithProfiles appends the profiles for the service
// parameters:
//   - profiles: the profiles to append
func WithProfiles(profiles ...string) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if config.Profiles == nil {
			config.Profiles = make([]string, 0)
		}
		config.Profiles = append(config.Profiles, profiles...)
		return nil
	}
}

// WithBuild sets the build config for the service
// parameters:
//   - setters: the setters for the build config
func WithBuild(setters ...build.SetBuildConfig) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if len(setters) == 0 {
			return nil
		}
		if config.Build == nil {
			config.Build = &types.BuildConfig{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(config.Build); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithSecret appends a secret to the service
// parameters:
//   - setters: the setters for the secret
//
// secrets specifies secrets to expose to the service.
func WithSecret(setters ...secretservice.SetSecretServiceConfig) create.SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if len(setters) == 0 {
			return nil
		}
		if config.Secrets == nil {
			config.Secrets = make([]types.ServiceSecretConfig, 0)
		}
		secret := types.ServiceSecretConfig{}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(&secret); err != nil {
				return err
			}
		}
		config.Secrets = append(config.Secrets, secret)
		return nil
	}
}
