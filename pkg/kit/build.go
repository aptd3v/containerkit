package kit

import "github.com/compose-spec/compose-go/v2/types"

// Build is a fluent Compose service build spec.
// Pass it to Service.Build. A later Build call is applied after earlier ones.
type Build struct {
	spec BuildSpec
}

// NewBuild starts an empty build spec.
func NewBuild() Build {
	return Build{}
}

// NewBuildFrom starts a build spec from an existing BuildSpec value.
func NewBuildFrom(spec BuildSpec) Build {
	return Build{spec: spec}
}

// Apply writes the build spec onto a Compose service.
func (b Build) Apply(cfg *types.ServiceConfig) error {
	return b.spec.Apply(cfg)
}

// Context sets the build context directory.
func (b Build) Context(path string) Build {
	b.spec.Context = path
	return b
}

// DockerfilePath sets the Dockerfile path inside the context.
func (b Build) DockerfilePath(path string) Build {
	b.spec.Dockerfile = path
	return b
}

// Dockerfile sets the Dockerfile text from a DockerFile builder.
// A nil file is ignored. When validation fails, WithService reports the error
// and does not apply the spec.
func (b Build) Dockerfile(file DockerFile) Build {
	if file == nil {
		return b
	}
	b.spec.DockerfileInline = file.String()
	if err := file.Validate(); err != nil {
		b.spec.err = err
	}
	return b
}

// Arg sets one build arg.
func (b Build) Arg(key, value string) Build {
	if b.spec.Args == nil {
		b.spec.Args = map[string]string{}
	}
	b.spec.Args[key] = value
	return b
}

// Args merges build args into the spec.
func (b Build) Args(args map[string]string) Build {
	for key, value := range args {
		b = b.Arg(key, value)
	}
	return b
}

// Label sets one build label.
func (b Build) Label(key, value string) Build {
	if b.spec.Labels == nil {
		b.spec.Labels = map[string]string{}
	}
	b.spec.Labels[key] = value
	return b
}

// Labels merges build labels into the spec.
func (b Build) Labels(labels map[string]string) Build {
	for key, value := range labels {
		b = b.Label(key, value)
	}
	return b
}

// Tag appends image tags applied to the built image.
func (b Build) Tag(tags ...string) Build {
	b.spec.Tags = append(b.spec.Tags, tags...)
	return b
}

// Target sets the Dockerfile target stage.
func (b Build) Target(name string) Build {
	b.spec.Target = name
	return b
}

// Network sets the network mode used during the build.
func (b Build) Network(name string) Build {
	b.spec.Network = name
	return b
}

// NoCache disables the build cache.
func (b Build) NoCache() Build {
	b.spec.NoCache = true
	return b
}

// Pull pulls newer versions of the base image.
func (b Build) Pull() Build {
	b.spec.Pull = true
	return b
}

// Privileged enables privileged mode for the build.
func (b Build) Privileged() Build {
	b.spec.Privileged = true
	return b
}

// Platform appends target platforms.
func (b Build) Platform(platforms ...string) Build {
	b.spec.Platforms = append(b.spec.Platforms, platforms...)
	return b
}

// CacheFrom appends cache import sources.
func (b Build) CacheFrom(sources ...string) Build {
	b.spec.CacheFrom = append(b.spec.CacheFrom, sources...)
	return b
}

// CacheTo appends cache export destinations.
func (b Build) CacheTo(destinations ...string) Build {
	b.spec.CacheTo = append(b.spec.CacheTo, destinations...)
	return b
}

// SSH appends an SSH agent socket or key available to the build.
func (b Build) SSH(id, path string) Build {
	b.spec.SSH = append(b.spec.SSH, BuildSSH{ID: id, Path: path})
	return b
}

// AdditionalContext sets one named additional build context.
func (b Build) AdditionalContext(name, value string) Build {
	if b.spec.AdditionalContexts == nil {
		b.spec.AdditionalContexts = map[string]string{}
	}
	b.spec.AdditionalContexts[name] = value
	return b
}

// AdditionalContexts merges named additional build contexts.
func (b Build) AdditionalContexts(contexts map[string]string) Build {
	for name, value := range contexts {
		b = b.AdditionalContext(name, value)
	}
	return b
}

// ExtraHost appends an extra host entry in "name:ip" form.
func (b Build) ExtraHost(host string) Build {
	b.spec.ExtraHosts = append(b.spec.ExtraHosts, host)
	return b
}

// Isolation sets the build isolation mode.
func (b Build) Isolation(mode string) Build {
	b.spec.Isolation = mode
	return b
}

// ShmSize sets /dev/shm size in bytes.
func (b Build) ShmSize(size int64) Build {
	b.spec.ShmSize = size
	return b
}

// Entitlement appends build entitlements.
func (b Build) Entitlement(names ...string) Build {
	b.spec.Entitlements = append(b.spec.Entitlements, names...)
	return b
}

// Ulimit sets one build ulimit.
func (b Build) Ulimit(name string, soft, hard int) Build {
	if b.spec.Ulimits == nil {
		b.spec.Ulimits = map[string]BuildUlimit{}
	}
	b.spec.Ulimits[name] = BuildUlimit{Soft: soft, Hard: hard}
	return b
}

// Secret appends a build-time secret reference.
func (b Build) Secret(secret BuildSecret) Build {
	b.spec.Secrets = append(b.spec.Secrets, secret)
	return b
}

// Provenance sets the provenance attestation mode.
func (b Build) Provenance(mode string) Build {
	b.spec.Provenance = mode
	return b
}

// SBOM sets the SBOM attestation mode.
func (b Build) SBOM(mode string) Build {
	b.spec.SBOM = mode
	return b
}

// NoCacheFilter appends cache filters that skip the cache.
func (b Build) NoCacheFilter(filters ...string) Build {
	b.spec.NoCacheFilter = append(b.spec.NoCacheFilter, filters...)
	return b
}
