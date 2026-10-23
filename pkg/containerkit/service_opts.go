package containerkit

import (
	"fmt"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

// DependsOn marks a service as depending on another (condition: started).
func DependsOn(service string) SetServiceConfig {
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

// DependsOnHealthy marks a service as depending on another becoming healthy.
func DependsOnHealthy(service string) SetServiceConfig {
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

// Profiles assigns Compose profiles to a service.
func Profiles(profiles ...string) SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if config.Profiles == nil {
			config.Profiles = make([]string, 0)
		}
		config.Profiles = append(config.Profiles, profiles...)
		return nil
	}
}

// EnvFile adds a required env file to a service.
func EnvFile(path string) SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		config.EnvFiles = append(config.EnvFiles, types.EnvFile{
			Path:     path,
			Required: true,
		})
		return nil
	}
}

// NoAttach sets attach to false for a service.
func NoAttach() SetServiceConfig {
	attach := false
	return func(config *types.ServiceConfig) error {
		config.Attach = &attach
		return nil
	}
}

// Annotation sets a service annotation.
func Annotation(key, value string) SetServiceConfig {
	return func(config *types.ServiceConfig) error {
		if config.Annotations == nil {
			config.Annotations = make(types.Mapping)
		}
		config.Annotations[key] = value
		return nil
	}
}

// Develop sets Compose watch/develop config.
func Develop(action WatchAction, watchPath, target string, ignorePaths ...string) SetServiceConfig {
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

// PreStart adds a run-to-completion hook that runs before the service starts.
func PreStart(ctr *Container, perReplica bool) SetServiceConfig {
	return func(service *types.ServiceConfig) error {
		if ctr == nil {
			return errdefs.NewServiceConfigError("pre_start", "container is nil")
		}
		if err := ctr.Validate(); err != nil {
			return err
		}
		spec, specErr := serviceSpec(ctr.Name, ctr)
		if specErr != nil {
			return specErr
		}
		service.PreStart = append(service.PreStart, types.PreStartHook{
			ContainerSpec: spec.ContainerSpec,
			PerReplica:    perReplica,
		})
		return nil
	}
}

// ServiceModel binds a project model to a service.
func ServiceModel(name, endpointVar, modelVar string) SetServiceConfig {
	return func(service *types.ServiceConfig) error {
		if name == "" {
			return errdefs.NewServiceConfigError("model", "model name is empty")
		}
		if service.Models == nil {
			service.Models = map[string]*types.ServiceModelConfig{}
		}
		service.Models[name] = &types.ServiceModelConfig{
			EndpointVariable: endpointVar,
			ModelVariable:    modelVar,
		}
		return nil
	}
}

func applyServiceExtra(serv *types.ServiceConfig, extra any) error {
	switch e := extra.(type) {
	case nil:
		return nil
	case SetServiceConfig:
		return e(serv)
	case BuildSpec:
		if e.err != nil {
			return e.err
		}
		return e.apply(serv)
	case *BuildSpec:
		if e == nil {
			return nil
		}
		if e.err != nil {
			return e.err
		}
		return e.apply(serv)
	case Deploy:
		e.apply(serv)
		return nil
	case *Deploy:
		if e != nil {
			e.apply(serv)
		}
		return nil
	case ServiceSecret:
		e.apply(serv)
		return nil
	default:
		return errdefs.NewServiceConfigError("service", fmt.Sprintf("unsupported service extra type %T", extra))
	}
}
