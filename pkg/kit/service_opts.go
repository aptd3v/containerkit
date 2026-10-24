package kit

import (
	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

// DependsOn marks the service as depending on another (condition: started).
func (s *Service) DependsOn(name string) *Service {
	return s.add(setServiceConfig(func(config *types.ServiceConfig) error {
		if config.DependsOn == nil {
			config.DependsOn = make(types.DependsOnConfig, 0)
		}
		config.DependsOn[name] = types.ServiceDependency{
			Condition: "service_started",
			Restart:   true,
			Required:  true,
		}
		return nil
	}))
}

// DependsOnHealthy marks the service as depending on another becoming healthy.
func (s *Service) DependsOnHealthy(name string) *Service {
	return s.add(setServiceConfig(func(config *types.ServiceConfig) error {
		if config.DependsOn == nil {
			config.DependsOn = make(types.DependsOnConfig, 0)
		}
		config.DependsOn[name] = types.ServiceDependency{
			Condition: "service_healthy",
			Restart:   true,
			Required:  true,
		}
		return nil
	}))
}

// Profiles assigns Compose profiles to the service.
func (s *Service) Profiles(profiles ...string) *Service {
	return s.add(setServiceConfig(func(config *types.ServiceConfig) error {
		if config.Profiles == nil {
			config.Profiles = make([]string, 0)
		}
		config.Profiles = append(config.Profiles, profiles...)
		return nil
	}))
}

// EnvFile adds a required env file to the service.
func (s *Service) EnvFile(path string) *Service {
	return s.add(setServiceConfig(func(config *types.ServiceConfig) error {
		config.EnvFiles = append(config.EnvFiles, types.EnvFile{
			Path:     path,
			Required: true,
		})
		return nil
	}))
}

// NoAttach sets attach to false for the service.
func (s *Service) NoAttach() *Service {
	attach := false
	return s.add(setServiceConfig(func(config *types.ServiceConfig) error {
		config.Attach = &attach
		return nil
	}))
}

// Annotation sets a service annotation.
func (s *Service) Annotation(key, value string) *Service {
	return s.add(setServiceConfig(func(config *types.ServiceConfig) error {
		if config.Annotations == nil {
			config.Annotations = make(types.Mapping)
		}
		config.Annotations[key] = value
		return nil
	}))
}

// Develop sets Compose watch/develop config.
func (s *Service) Develop(watch Develop) *Service {
	return s.add(setServiceConfig(func(config *types.ServiceConfig) error {
		config.Develop = &types.DevelopConfig{
			Watch: []types.Trigger{
				{
					Path:   watch.Path,
					Action: types.WatchAction(watch.Action),
					Target: watch.Target,
					Ignore: watch.Ignore,
				},
			},
		}
		return nil
	}))
}

// PreStart adds a run-to-completion hook that runs before the service starts.
func (s *Service) PreStart(ctr *Container, perReplica bool) *Service {
	return s.add(setServiceConfig(func(service *types.ServiceConfig) error {
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
	}))
}

// Model binds a project model to the service.
func (s *Service) Model(name, endpointVar, modelVar string) *Service {
	return s.add(setServiceConfig(func(service *types.ServiceConfig) error {
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
	}))
}

// Build sets the service build spec. A later Build call is applied after earlier ones.
func (s *Service) Build(build Build) *Service {
	return s.add(build)
}

// Dockerfile sets the service's inline Dockerfile directly without needing NewBuild.
func (s *Service) Dockerfile(file DockerFile) *Service {
	return s.add(NewBuild().Dockerfile(file))
}

// Deploy sets Compose deploy configuration.
func (s *Service) Deploy(deploy Deploy) *Service {
	return s.add(deploy)
}

// Secret adds a service-level secret reference.
func (s *Service) Secret(secret ServiceSecret) *Service {
	return s.add(secret)
}
