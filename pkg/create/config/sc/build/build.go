// Package build provides functions to set the build config for a service
package build

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/build/ulimit"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/secrets/secretservice"
	"github.com/compose-spec/compose-go/v2/types"
)

// SetBuildConfig is a function that sets the build config for a service
type SetBuildConfig func(opt *types.BuildConfig) error

// WithDockerfile sets the Dockerfile for the service
// parameters:
//   - path: the path to the Dockerfile
//   - context: the context for the build
func WithDockerfile(path string, context string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		opt.Dockerfile = path
		opt.Context = context
		return nil
	}
}

// WithDockerfileInline sets the Dockerfile inline for the service
// parameters:
//   - context: the context for the build
//   - inline: the inline Dockerfile
func WithDockerfileInline(context string, inline string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		opt.DockerfileInline = inline
		opt.Context = context
		return nil
	}
}

// WithArgs appends the args for the build
// parameters:
//   - key: the key for the arg
//   - value: the value for the arg
func WithArgs(key string, value string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.Args == nil {
			opt.Args = make(types.MappingWithEquals)
		}
		opt.Args[key] = &value
		return nil
	}
}

// WithSSHKey sets the SSH key for the build
// parameters:
//   - key: the key of the SSH key
//   - path: the path to the SSH key
//
// defines SSH authentications that the image
// builder should use during image build (e.g., cloning private repository).
func WithSSHKey(key string, path string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.SSH == nil {
			opt.SSH = make(types.SSHConfig, 0)
		}
		opt.SSH = append(opt.SSH, types.SSHKey{
			ID: fmt.Sprintf("%s=%s", key, path),
		})
		return nil
	}
}

// WithLabels appends the labels for the build
// parameters:
//   - key: the key for the label
//   - value: the value for the label
func WithLabels(key string, value string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.Labels == nil {
			opt.Labels = make(types.Labels)
		}
		opt.Labels[key] = value
		return nil
	}
}

// WithCacheFrom appends the cache from for the build
// parameters:
//   - cacheFrom: the cache from for the build
func WithCacheFrom(cacheFrom ...string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.CacheFrom == nil {
			opt.CacheFrom = make(types.StringList, 0)
		}
		opt.CacheFrom = append(opt.CacheFrom, cacheFrom...)
		return nil
	}
}

// WithCacheTo appends the cache to for the build
// parameters:
//   - cacheTo: the cache to for the build
func WithCacheTo(cacheTo ...string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.CacheTo == nil {
			opt.CacheTo = make(types.StringList, 0)
		}
		opt.CacheTo = append(opt.CacheTo, cacheTo...)
		return nil
	}
}

// WithNoCache sets the no cache for the build
// parameters:
//   - noCache: the no cache for the build
func WithNoCache() SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		opt.NoCache = true
		return nil
	}
}

// WithAdditionalContexts appends the additional contexts for the build
// parameters:
//   - key: the key for the additional context
//   - value: the value for the additional context
func WithAdditionalContexts(key string, value string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.AdditionalContexts == nil {
			opt.AdditionalContexts = make(types.Mapping)
		}
		opt.AdditionalContexts[key] = value
		return nil
	}
}

// WithPull sets the pull to true for the build
//
// pull requires the image builder to pull referenced images
// (FROM Dockerfile directive), even if those are already available in the local image store.
func WithPull() SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		opt.Pull = true
		return nil
	}
}

// WithExtraHosts appends the extra hosts for the build
// parameters:
//   - key: the key for the extra host
//   - value: the values for the extra host
//
// extra_hosts adds hostname mappings at build-time. Use the same syntax as extra_hosts.
//
//	extra_hosts:
//	 - "somehost=162.242.195.82"
//	 - "otherhost=50.31.209.229"
//	 - "myhostv6=::1"
func WithExtraHosts(key string, value ...string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.ExtraHosts == nil {
			opt.ExtraHosts = make(types.HostsList, 0)
		}
		opt.ExtraHosts[key] = value
		return nil
	}
}

// WithIsolation sets the isolation for the build
// parameters:
//   - isolation: the isolation for the build
//
// isolation specifies a build’s container isolation technology. Like isolation, supported values are platform specific.
func WithIsolation(isolation string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		opt.Isolation = isolation
		return nil
	}
}

// WithNetwork sets the network for the build
// parameters:
//   - network: the network for the build
//
// Set the network containers connect to for the RUN instructions during build.
func WithNetwork(network string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		opt.Network = network
		return nil
	}
}

// WithTarget sets the target for the build
// parameters:
//   - target: the target for the build
//
// target defines the stage to build as defined inside a multi-stage Dockerfile.
func WithTarget(target string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		opt.Target = target
		return nil
	}
}

// WithSecret appends a secret to the build
// parameters:
//   - setters: the setters for the secret
//
// secrets specifies secrets to expose to the build.
func WithSecret(setters ...secretservice.SetSecretServiceConfig) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.Secrets == nil {
			opt.Secrets = make([]types.ServiceSecretConfig, 0)
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
		opt.Secrets = append(opt.Secrets, secret)
		return nil
	}
}

// WithTags appends the tags for the build
// parameters:
//   - tags: the tags for the build
func WithTags(tags ...string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.Tags == nil {
			opt.Tags = make(types.StringList, 0)
		}
		opt.Tags = append(opt.Tags, tags...)
		return nil
	}
}

// WithUlimit appends a ulimit to the build
// parameters:
//   - name: the name for the ulimit
//   - setters: the setters for the ulimit
//
// ulimits overrides the default ulimits for a container.
// It's specified either as WithSingle for a single limit or WithSoft and WithHard for soft/hard limits.
func WithUlimit(name string, setters ...ulimit.SetUlimitConfig) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.Ulimits == nil {
			opt.Ulimits = make(map[string]*types.UlimitsConfig)
		}
		ulimit := types.UlimitsConfig{}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(&ulimit); err != nil {
				return err
			}
		}
		opt.Ulimits[name] = &ulimit
		return nil
	}
}

// WithPlatforms appends the platforms for the build
// parameters:
//   - platforms: the platforms for the build
func WithPlatforms(platforms ...string) SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		if opt.Platforms == nil {
			opt.Platforms = make(types.StringList, 0)
		}
		opt.Platforms = append(opt.Platforms, platforms...)
		return nil
	}
}

// WithPrivileged sets the privileged for the build
// parameters:
//   - privileged: the privileged for the build
//
// privileged configures the service image to build with
// elevated privileges. Support and actual impacts are platform specific.
func WithPrivileged() SetBuildConfig {
	return func(opt *types.BuildConfig) error {
		opt.Privileged = true
		return nil
	}
}
