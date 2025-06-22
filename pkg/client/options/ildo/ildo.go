// Package ildo provides options for image load.
package ildo

import (
	"io"

	"github.com/aptd3v/go-contain/pkg/create"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// ImageLoadOptions is the image load options.
type ImageLoadOptions struct {
	Input     io.Reader
	Quiet     bool
	Platforms []v1.Platform
}

// SetImageLoadOption is a function that sets the image load options.
type SetImageLoadOption func(*ImageLoadOptions) error

// WithInput sets the input for the image load.
func WithInput(input io.Reader) SetImageLoadOption {
	return func(o *ImageLoadOptions) error {
		o.Input = input
		return nil
	}
}

// WithQuiet sets the quiet flag for the image load.
func WithQuiet(quiet bool) SetImageLoadOption {
	return func(o *ImageLoadOptions) error {
		o.Quiet = quiet
		return nil
	}
}

// WithPlatforms sets the platforms for the image load.
func WithPlatforms(setters ...create.SetPlatformConfig) SetImageLoadOption {
	return func(o *ImageLoadOptions) error {
		pc := create.PlatformConfig{}
		platforms := make([]v1.Platform, 0)
		for _, setter := range setters {
			if setter != nil {
				if err := setter(&pc); err != nil {
					return err
				}
				platforms = append(platforms, *pc.Platform)
			}
		}
		if o.Platforms == nil {
			o.Platforms = make([]v1.Platform, 0)
		}
		o.Platforms = append(o.Platforms, platforms...)
		return nil
	}
}
