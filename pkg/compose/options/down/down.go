// Package down is the package for the compose down options
package down

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

type RemoveImageFlag string

const (
	RemoveAll  RemoveImageFlag = "all"
	RemoveNone RemoveImageFlag = "none"
)

// WithRemoveOrphans is a function that sets the remove orphans flag
func WithRemoveOrphans() compose.SetComposeDownOption {
	return func(opt *compose.ComposeDownOptions) error {
		opt.RemoveOrphans = true
		return nil
	}
}

// WithTimeout is a function that sets the timeout flag in seconds
func WithTimeout(timeout int) compose.SetComposeDownOption {
	return func(opt *compose.ComposeDownOptions) error {
		opt.Timeout = &timeout
		return nil
	}
}

// WithRemoveImage is a function that sets the remove image flag
func WithRemoveImage(flag RemoveImageFlag) compose.SetComposeDownOption {
	return func(opt *compose.ComposeDownOptions) error {
		str := string(flag)
		opt.RemoveImage = &str
		return nil
	}
}

// WithRemoveVolumes is a function that sets the remove volumes flag
//
// Remove named volumes declared in the "volumes" section of the Compose file and anonymous volumes attached to containers
func WithRemoveVolumes() compose.SetComposeDownOption {
	return func(opt *compose.ComposeDownOptions) error {
		opt.RemoveVolumes = true
		return nil
	}
}

// WithWriter sets the writer for the compose down command stdout and stderr
//
// if writer is nil, it will use os.Stdout as a fallback
func WithWriter(writer io.Writer) compose.SetComposeDownOption {
	return func(opt *compose.ComposeDownOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithProfiles is a function that sets the profiles for the compose down command
func WithProfiles(profiles ...string) compose.SetComposeDownOption {
	return func(opt *compose.ComposeDownOptions) error {
		opt.Profiles = profiles
		return nil
	}
}
