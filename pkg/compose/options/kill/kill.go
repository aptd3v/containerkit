// Package kill is the package for the compose kill options
package kill

import (
	"io"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// SetComposeKillOption is the type for the compose kill options
type SetComposeKillOption func(*compose.ComposeKillOptions) error

// ComposeKillOptions is the type for the compose kill options
type ComposeKillOptions struct {
	Signal *string
}

// WithSignal is a function that sets the signal for the kill
func WithSignal(signal string) compose.SetComposeKillOption {
	return func(opt *compose.ComposeKillOptions) error {
		opt.Signal = &signal
		return nil
	}
}

// WithRemoveOrphans is a function that sets the remove orphans flag
func WithRemoveOrphans() compose.SetComposeKillOption {
	return func(opt *compose.ComposeKillOptions) error {
		opt.RemoveOrphans = true
		return nil
	}
}

// WithWriter is a function that sets the writer
func WithWriter(writer io.Writer) compose.SetComposeKillOption {
	return func(opt *compose.ComposeKillOptions) error {
		opt.Writer = writer
		return nil
	}
}

// WithProfiles is a function that sets the profiles
func WithProfiles(profiles ...string) compose.SetComposeKillOption {
	return func(opt *compose.ComposeKillOptions) error {
		if opt.Profiles == nil {
			opt.Profiles = make([]string, 0, len(profiles))
		}
		opt.Profiles = append(opt.Profiles, profiles...)
		return nil
	}
}
