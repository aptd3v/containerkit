package containerkit

import (
	"fmt"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/sc"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/sc/network"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/sc/network/pool"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/sc/secrets/projectsecret"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/sc/volume"
	"github.com/compose-spec/compose-go/v2/types"
)

// DependsOn marks a service as depending on another (condition: started).
func DependsOn(service string) SetServiceConfig {
	return sc.WithDependsOn(service)
}

// DependsOnHealthy marks a service as depending on another becoming healthy.
func DependsOnHealthy(service string) SetServiceConfig {
	return sc.WithDependsOnHealthy(service)
}

// Profiles assigns Compose profiles to a service.
func Profiles(profiles ...string) SetServiceConfig {
	return sc.WithProfiles(profiles...)
}

// EnvFile adds a required env file to a service.
func EnvFile(path string) SetServiceConfig {
	return sc.WithEnvFile(path)
}

// NoAttach sets attach to false for a service.
func NoAttach() SetServiceConfig {
	return sc.WithNoAttach()
}

// Annotation sets a service annotation.
func Annotation(key, value string) SetServiceConfig {
	return sc.WithAnnotation(key, value)
}

// Develop sets Compose watch/develop config.
func Develop(action WatchAction, watchPath, target string, ignorePaths ...string) SetServiceConfig {
	return sc.WithDevelop(sc.WatchAction(action), watchPath, target, ignorePaths...)
}

func applyProjectNetwork(name string, spec ProjectNetwork) []network.SetNetworkProjectConfig {
	var setters []network.SetNetworkProjectConfig
	if spec.Driver != "" {
		setters = append(setters, network.WithDriver(spec.Driver))
	}
	for k, v := range spec.DriverOptions {
		setters = append(setters, network.WithDriverOptions(k, v))
	}
	if spec.Internal {
		setters = append(setters, network.WithInternal())
	}
	if spec.Attachable {
		setters = append(setters, network.WithAttachable())
	}
	if spec.EnableIPv6 {
		setters = append(setters, network.WithEnableIPv6())
	}
	for k, v := range spec.Labels {
		setters = append(setters, network.WithLabel(k, v))
	}
	if spec.IPAMDriver != "" {
		setters = append(setters, network.WithIpamDriver(spec.IPAMDriver))
	}
	for _, p := range spec.IPAMPools {
		var ps []pool.SetIpamPoolProjectConfig
		if p.Subnet != "" {
			ps = append(ps, pool.WithSubnet(p.Subnet))
		}
		if p.Gateway != "" {
			ps = append(ps, pool.WithGateway(p.Gateway))
		}
		if p.IPRange != "" {
			ps = append(ps, pool.WithIpRange(p.IPRange))
		}
		for k, v := range p.AuxiliaryAddresses {
			ps = append(ps, pool.WithAuxiliaryAddresses(k, v))
		}
		setters = append(setters, network.WithIpamPool(ps...))
	}
	_ = name
	return setters
}

func applyProjectVolume(spec ProjectVolume) []volume.SetVolumeProjectConfig {
	var setters []volume.SetVolumeProjectConfig
	if spec.Driver != "" {
		setters = append(setters, volume.WithDriver(spec.Driver))
	}
	for k, v := range spec.DriverOptions {
		setters = append(setters, volume.WithDriverOptions(k, v))
	}
	for k, v := range spec.Labels {
		setters = append(setters, volume.WithLabel(k, v))
	}
	return setters
}

func applyProjectSecret(spec ProjectSecret) []projectsecret.SetProjectSecretConfig {
	var setters []projectsecret.SetProjectSecretConfig
	if spec.Name != "" {
		setters = append(setters, projectsecret.WithName(spec.Name))
	}
	if spec.File != "" {
		setters = append(setters, projectsecret.WithFile(spec.File))
	}
	if spec.Content != "" {
		setters = append(setters, projectsecret.WithContent(spec.Content))
	}
	if spec.Environment != "" {
		setters = append(setters, projectsecret.WithEnvironment(spec.Environment))
	}
	if spec.External {
		setters = append(setters, projectsecret.WithExternal())
	}
	if spec.Driver != "" {
		setters = append(setters, projectsecret.WithDriver(spec.Driver))
	}
	for k, v := range spec.DriverOptions {
		setters = append(setters, projectsecret.WithDriverOptions(k, v))
	}
	if spec.TemplateDriver != "" {
		setters = append(setters, projectsecret.WithTemplateDriver(spec.TemplateDriver))
	}
	return setters
}

func applyServiceExtra(serv *types.ServiceConfig, extra any) error {
	switch e := extra.(type) {
	case nil:
		return nil
	case SetServiceConfig:
		return e(serv)
	case BuildSpec:
		e.apply(serv)
		return nil
	case *BuildSpec:
		if e != nil {
			e.apply(serv)
		}
		return nil
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
