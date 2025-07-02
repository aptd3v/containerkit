// Package remove provides options for image remove.
package remove

import (
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/docker/docker/api/types/image"
	p "github.com/opencontainers/image-spec/specs-go/v1"
)

// SetImageRemoveOption is a function that sets the image remove options.
type SetImageRemoveOption func(*image.RemoveOptions) error

// WithForce sets the force flag for the image remove.
func WithForce() SetImageRemoveOption {
	return func(o *image.RemoveOptions) error {
		o.Force = true
		return nil
	}
}

// WithPruneChildren sets the prune children flag for the image remove.
func WithPruneChildren() SetImageRemoveOption {
	return func(o *image.RemoveOptions) error {
		o.PruneChildren = true
		return nil
	}
}

// WithPlatform appends the platform for the image remove.
func WithPlatform(setters ...create.SetPlatformConfig) SetImageRemoveOption {
	return func(o *image.RemoveOptions) error {
		pc := p.Platform{}
		for _, setter := range setters {
			if err := setter(&pc); err != nil {
				return err
			}
		}
		if o.Platforms == nil {
			o.Platforms = make([]p.Platform, 0)
		}
		o.Platforms = append(o.Platforms, pc)
		return nil
	}
}
