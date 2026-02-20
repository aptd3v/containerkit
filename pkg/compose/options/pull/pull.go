// Package pull provides options for the compose pull command
package pull

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// WithIgnoreBuildable ignores images that can be built
func WithIgnoreBuildable() compose.SetComposePullOption {
	return func(opt *compose.ComposePullOptions) error {
		opt.IgnoreBuildable = true
		return nil
	}
}

// WithIgnorePullFailures pulls what it can and ignores images with pull failures
func WithIgnorePullFailures() compose.SetComposePullOption {
	return func(opt *compose.ComposePullOptions) error {
		opt.IgnorePullFailures = true
		return nil
	}
}

// WithIncludeDeps also pulls services declared as dependencies
func WithIncludeDeps() compose.SetComposePullOption {
	return func(opt *compose.ComposePullOptions) error {
		opt.IncludeDeps = true
		return nil
	}
}

// WithPolicy sets pull policy (e.g. "missing", "always", "never")
func WithPolicy(policy string) compose.SetComposePullOption {
	return func(opt *compose.ComposePullOptions) error {
		opt.Policy = policy
		return nil
	}
}

// WithQuiet pulls without printing progress information
func WithQuiet() compose.SetComposePullOption {
	return func(opt *compose.ComposePullOptions) error {
		opt.Quiet = true
		return nil
	}
}

// WithWriter sets the writer for stdout/stderr
func WithWriter(writer io.Writer) compose.SetComposePullOption {
	return func(opt *compose.ComposePullOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithProfiles sets the profiles to activate
func WithProfiles(profiles ...string) compose.SetComposePullOption {
	return func(opt *compose.ComposePullOptions) error {
		opt.Profiles = profiles
		return nil
	}
}

// WithServiceNames sets the optional service names (positional args)
func WithServiceNames(names ...string) compose.SetComposePullOption {
	return func(opt *compose.ComposePullOptions) error {
		opt.ServiceNames = append(opt.ServiceNames, names...)
		return nil
	}
}
