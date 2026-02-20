// Package stop provides options for the compose stop command
package stop

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// WithTimeout sets the shutdown timeout in seconds
func WithTimeout(seconds int) compose.SetComposeStopOption {
	return func(opt *compose.ComposeStopOptions) error {
		opt.Timeout = &seconds
		return nil
	}
}

// WithWriter sets the writer for stdout/stderr
func WithWriter(writer io.Writer) compose.SetComposeStopOption {
	return func(opt *compose.ComposeStopOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithProfiles sets the profiles to activate
func WithProfiles(profiles ...string) compose.SetComposeStopOption {
	return func(opt *compose.ComposeStopOptions) error {
		opt.Profiles = profiles
		return nil
	}
}

// WithServiceNames sets the optional service names (positional args)
func WithServiceNames(names ...string) compose.SetComposeStopOption {
	return func(opt *compose.ComposeStopOptions) error {
		opt.ServiceNames = append(opt.ServiceNames, names...)
		return nil
	}
}
