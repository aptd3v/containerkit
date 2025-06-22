// Package crmo provides the options for the container remove options
package crmo

import "github.com/docker/docker/api/types/container"

// SetContainerRemoveOption is a function that sets the container remove options.
type SetContainerRemoveOption func(*container.RemoveOptions) error

// WithForce sets the force flag for the container remove options.
func WithForce() SetContainerRemoveOption {
	return func(op *container.RemoveOptions) error {
		op.Force = true
		return nil
	}
}

// WithVolumes sets the volumes flag for the container remove options.
func WithVolumes() SetContainerRemoveOption {
	return func(op *container.RemoveOptions) error {
		op.RemoveVolumes = true
		return nil
	}
}

// WithLinks sets the links flag for the container remove options.
func WithLinks() SetContainerRemoveOption {
	return func(op *container.RemoveOptions) error {
		op.RemoveLinks = true
		return nil
	}
}
