// Package logs is the package for the compose logs options
package logs

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// WithTail is a function that sets the tail flag
func WithTail(tail int) compose.SetComposeLogsOption {
	return func(opt *compose.ComposeLogsOptions) error {
		opt.Tail = &tail
		return nil
	}
}

func WithFollow() compose.SetComposeLogsOption {
	return func(opt *compose.ComposeLogsOptions) error {
		opt.Follow = true
		return nil
	}
}

func WithNoLogPrefix() compose.SetComposeLogsOption {
	return func(opt *compose.ComposeLogsOptions) error {
		opt.NoLogPrefix = true
		return nil
	}
}

// WithWriter sets the writer for the compose logs command stdout and stderr
//
// if writer is nil, it will use os.Stdout as a fallback
func WithWriter(writer io.Writer) compose.SetComposeLogsOption {
	return func(opt *compose.ComposeLogsOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithProfiles sets the profiles for the compose logs command.
func WithProfiles(profiles ...string) compose.SetComposeLogsOption {
	return func(opt *compose.ComposeLogsOptions) error {
		if opt.Profiles == nil {
			opt.Profiles = make([]string, 0, len(profiles))
		}
		opt.Profiles = append(opt.Profiles, profiles...)
		return nil
	}
}
