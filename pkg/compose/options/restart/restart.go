// Package restart provides options for the compose restart command
package restart

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// WithNoDeps does not restart dependent services
func WithNoDeps() compose.SetComposeRestartOption {
	return func(opt *compose.ComposeRestartOptions) error {
		opt.NoDeps = true
		return nil
	}
}

// WithTimeout sets the shutdown timeout in seconds
func WithTimeout(seconds int) compose.SetComposeRestartOption {
	return func(opt *compose.ComposeRestartOptions) error {
		opt.Timeout = &seconds
		return nil
	}
}

// WithWriter sets the writer for stdout/stderr
func WithWriter(writer io.Writer) compose.SetComposeRestartOption {
	return func(opt *compose.ComposeRestartOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithProfiles sets the profiles to activate
func WithProfiles(profiles ...string) compose.SetComposeRestartOption {
	return func(opt *compose.ComposeRestartOptions) error {
		opt.Profiles = profiles
		return nil
	}
}

// WithServiceNames sets the optional service names (positional args)
func WithServiceNames(names ...string) compose.SetComposeRestartOption {
	return func(opt *compose.ComposeRestartOptions) error {
		opt.ServiceNames = append(opt.ServiceNames, names...)
		return nil
	}
}
