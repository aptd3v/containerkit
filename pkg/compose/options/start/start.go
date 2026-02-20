// Package start provides options for the compose start command
package start

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// WithWait waits for services to be running
func WithWait() compose.SetComposeStartOption {
	return func(opt *compose.ComposeStartOptions) error {
		opt.Wait = true
		return nil
	}
}

// WithWaitTimeout sets the maximum duration in seconds to wait
func WithWaitTimeout(seconds int) compose.SetComposeStartOption {
	return func(opt *compose.ComposeStartOptions) error {
		opt.WaitTimeout = &seconds
		return nil
	}
}

// WithWriter sets the writer for stdout/stderr
func WithWriter(writer io.Writer) compose.SetComposeStartOption {
	return func(opt *compose.ComposeStartOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithProfiles sets the profiles to activate
func WithProfiles(profiles ...string) compose.SetComposeStartOption {
	return func(opt *compose.ComposeStartOptions) error {
		opt.Profiles = profiles
		return nil
	}
}

// WithServiceNames sets the optional service names (positional args)
func WithServiceNames(names ...string) compose.SetComposeStartOption {
	return func(opt *compose.ComposeStartOptions) error {
		opt.ServiceNames = append(opt.ServiceNames, names...)
		return nil
	}
}
