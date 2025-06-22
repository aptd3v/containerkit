// package ibo provides options for the image build.
package ibo

import (
	"io"

	"github.com/aptd3v/go-contain/pkg/client/auth"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/registry"
)

type Isolation string

// Isolation modes for containers
const (
	IsolationEmpty   Isolation = ""        // IsolationEmpty is unspecified (same behavior as default)
	IsolationDefault Isolation = "default" // IsolationDefault is the default isolation mode on current daemon
	IsolationProcess Isolation = "process" // IsolationProcess is process isolation mode
	IsolationHyperV  Isolation = "hyperv"  // IsolationHyperV is HyperV isolation mode
)

// BuilderVersion sets the version of underlying builder to use
type BuilderVersion string

const (
	// BuilderV1 is the first generation builder in docker daemon
	BuilderV1 BuilderVersion = "1"
	// BuilderBuildKit is builder based on moby/buildkit project
	BuilderBuildKit BuilderVersion = "2"
)

// SetImageBuildOption is a function that sets a parameter for the image build.
type SetImageBuildOption func(*build.ImageBuildOptions) error

// WithTags appends the tags for the image build.
func WithTags(tags ...string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		if o.Tags == nil {
			o.Tags = make([]string, 0, len(tags))
		}
		o.Tags = append(o.Tags, tags...)
		return nil
	}
}

// WithSuppressOutput sets the suppress output flag for the image build.
func WithSuppressOutput() SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.SuppressOutput = true
		return nil
	}
}
func WithRemoteContext(remoteContext string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.RemoteContext = remoteContext
		return nil
	}
}

// WithNoCache sets the no cache flag for the image build.
func WithNoCache() SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.NoCache = true
		return nil
	}
}

// WithRemove sets the remove flag for the image build.
func WithRemove() SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Remove = true
		return nil
	}
}

// WithForceRemove sets the force remove flag for the image build.
func WithForceRemove() SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.ForceRemove = true
		return nil
	}
}

// WithPullParent sets the pull parent flag for the image build.
func WithPullParent() SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.PullParent = true
		return nil
	}
}

// WithIsolation sets the isolation mode for the image build.
func WithIsolation(isolation Isolation) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Isolation = container.Isolation(isolation)
		return nil
	}
}

// WithCPUSetCPUs sets the CPU set CPUs for the image build.
func WithCPUSetCPUs(cpus string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.CPUSetCPUs = cpus
		return nil
	}
}

// WithCPUSetMems sets the CPU set mems for the image build.
func WithCPUSetMems(mems string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.CPUSetMems = mems
		return nil
	}
}

// WithCPUShares sets the CPU shares for the image build.
func WithCPUShares(shares int64) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.CPUShares = shares
		return nil
	}
}

// WithMemory sets the memory for the image build.
func WithMemory(memory int64) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Memory = memory
		return nil
	}
}

// WithMemorySwap sets the memory swap for the image build.
func WithMemorySwap(memorySwap int64) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.MemorySwap = memorySwap
		return nil
	}
}

// WithCgroupParent sets the cgroup parent for the image build.
func WithCgroupParent(cgroupParent string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.CgroupParent = cgroupParent
		return nil
	}
}

// WithNetworkMode sets the network mode for the image build.
func WithNetworkMode(networkMode string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.NetworkMode = networkMode
		return nil
	}
}

// WithShmSize sets the shm size for the image build.
func WithShmSize(shmSize int64) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.ShmSize = shmSize
		return nil
	}
}

// WithDockerfile sets the dockerfile for the image build.
func WithDockerfile(dockerfile string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Dockerfile = dockerfile
		return nil
	}
}

// WithUlimits appends the ulimits for the image build.
func WithUlimits(name string, soft, hard int64) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		if o.Ulimits == nil {
			o.Ulimits = make([]*container.Ulimit, 0, 1)
		}
		o.Ulimits = append(o.Ulimits, &container.Ulimit{Name: name, Soft: soft, Hard: hard})
		return nil
	}
}

// WithBuildArgs appends the build args for the image build.
func WithBuildArgs(key string, value string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		if o.BuildArgs == nil {
			o.BuildArgs = make(map[string]*string)
		}
		o.BuildArgs[key] = &value
		return nil
	}
}

// WithAuthConfig appends the auth config for the image build.
func WithAuthConfig(name string, setters ...auth.SetRegistryAuthConfigOption) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		authConfig := registry.AuthConfig{}
		for _, setter := range setters {
			if err := setter(&authConfig); err != nil {
				return err
			}
		}
		if o.AuthConfigs == nil {
			o.AuthConfigs = make(map[string]registry.AuthConfig)
		}
		o.AuthConfigs[name] = authConfig
		return nil
	}
}

// WithContext sets the context for the image build.
func WithContext(context io.Reader) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Context = context
		return nil
	}
}

// WithLabel appends the label for the image build.
func WithLabel(key, value string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		if o.Labels == nil {
			o.Labels = make(map[string]string)
		}
		o.Labels[key] = value
		return nil
	}
}

// WithSquash sets the squash flag for the image build.
func WithSquash() SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Squash = true
		return nil
	}
}

// WithCacheFrom appends the cache from for the image build.
func WithCacheFrom(cacheFrom ...string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		if o.CacheFrom == nil {
			o.CacheFrom = make([]string, 0, len(cacheFrom))
		}
		o.CacheFrom = append(o.CacheFrom, cacheFrom...)
		return nil
	}
}

// WithSecurityOpt appends the security opt for the image build.
func WithSecurityOpt(securityOpt ...string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		if o.SecurityOpt == nil {
			o.SecurityOpt = make([]string, 0, len(securityOpt))
		}
		o.SecurityOpt = append(o.SecurityOpt, securityOpt...)
		return nil
	}
}

// WithExtraHosts appends the extra hosts for the image build.
func WithExtraHosts(extraHosts ...string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		if o.ExtraHosts == nil {
			o.ExtraHosts = make([]string, 0, len(extraHosts))
		}
		o.ExtraHosts = append(o.ExtraHosts, extraHosts...)
		return nil
	}
}

// WithTarget sets the target for the image build.
func WithTarget(target string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Target = target
		return nil
	}
}

// WithSessionID sets the session id for the image build.
func WithSessionID(sessionID string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.SessionID = sessionID
		return nil
	}
}

// WithPlatform sets the platform for the image build.
func WithPlatform(platform string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Platform = platform
		return nil
	}
}

// WithBuilderVersion sets the builder version for the image build.
func WithBuilderVersion(builderVersion string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.Version = build.BuilderVersion(builderVersion)
		return nil
	}
}

// WithBuildID sets the build id for the image build.
func WithBuildID(buildID string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		o.BuildID = buildID
		return nil
	}
}

// WithOutputs appends the outputs for the image build.
func WithOutputs(oType string, attrs map[string]string) SetImageBuildOption {
	return func(o *build.ImageBuildOptions) error {
		if o.Outputs == nil {
			o.Outputs = make([]build.ImageBuildOutput, 0)
		}
		o.Outputs = append(o.Outputs, build.ImageBuildOutput{Type: oType, Attrs: attrs})
		return nil
	}
}
