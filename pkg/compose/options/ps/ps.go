// Package ps provides options for the compose ps command
package ps

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// WithAll shows all stopped containers (including those created by run)
func WithAll() compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.All = true
		return nil
	}
}

// WithFilter adds a filter (e.g. "status=running")
func WithFilter(keyValue string) compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.Filter = append(opt.Filter, keyValue)
		return nil
	}
}

// WithFormat sets the output format (e.g. "table", "json")
func WithFormat(format string) compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.Format = format
		return nil
	}
}

// WithNoTrunc disables truncation of output
func WithNoTrunc() compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.NoTrunc = true
		return nil
	}
}

// WithOrphans sets whether to include orphaned services (false = --no-orphans)
func WithOrphans(include bool) compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.Orphans = &include
		return nil
	}
}

// WithQuiet only display container IDs
func WithQuiet() compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.Quiet = true
		return nil
	}
}

// WithServices display services
func WithServices() compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.Services = true
		return nil
	}
}

// WithStatus filters by status (e.g. "running", "exited")
func WithStatus(status string) compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.Status = status
		return nil
	}
}

// WithWriter sets the writer for stdout/stderr
func WithWriter(writer io.Writer) compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithProfiles sets the profiles to activate
func WithProfiles(profiles ...string) compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.Profiles = profiles
		return nil
	}
}

// WithServiceNames sets the optional service names (positional args)
func WithServiceNames(names ...string) compose.SetComposePsOption {
	return func(opt *compose.ComposePsOptions) error {
		opt.ServiceNames = append(opt.ServiceNames, names...)
		return nil
	}
}
