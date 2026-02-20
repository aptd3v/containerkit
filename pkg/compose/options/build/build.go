// Package build provides options for the compose build command
package build

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// WithBuildArg adds build-time variables (key=value), can be used multiple times
func WithBuildArg(args ...string) compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.BuildArg = append(opt.BuildArg, args...)
		return nil
	}
}

// WithBuilder sets the builder to use
func WithBuilder(builder string) compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Builder = builder
		return nil
	}
}

// WithCheck checks build configuration
func WithCheck() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Check = true
		return nil
	}
}

// WithMemory sets memory limit for the build container (e.g. "2G")
func WithMemory(memory string) compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Memory = memory
		return nil
	}
}

// WithNoCache disables cache when building the image
func WithNoCache() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.NoCache = true
		return nil
	}
}

// WithPrint prints equivalent bake file
func WithPrint() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Print = true
		return nil
	}
}

// WithProvenance adds a provenance attestation
func WithProvenance() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Provenance = true
		return nil
	}
}

// WithPull always attempt to pull a newer version of the image
func WithPull() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Pull = true
		return nil
	}
}

// WithPush pushes service images
func WithPush() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Push = true
		return nil
	}
}

// WithQuiet suppresses build output
func WithQuiet() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Quiet = true
		return nil
	}
}

// WithSBOM adds an SBOM attestation
func WithSBOM() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.SBOM = true
		return nil
	}
}

// WithSSH sets SSH authentications (e.g. "default" or "key=path"), can be used multiple times
func WithSSH(ssh ...string) compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.SSH = append(opt.SSH, ssh...)
		return nil
	}
}

// WithDependencies also builds dependencies (transitively)
func WithDependencies() compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.WithDependencies = true
		return nil
	}
}

// WithWriter sets the writer for stdout/stderr
func WithWriter(writer io.Writer) compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithProfiles sets the profiles to activate
func WithProfiles(profiles ...string) compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.Profiles = profiles
		return nil
	}
}

// WithServiceNames sets the optional service names (positional args)
func WithServiceNames(names ...string) compose.SetComposeBuildOption {
	return func(opt *compose.ComposeBuildOptions) error {
		opt.ServiceNames = append(opt.ServiceNames, names...)
		return nil
	}
}
