// package cero provides options for the container exec resize.
package cero

import "github.com/docker/docker/api/types/container"

// SetContainerExecResizeOption is a function that sets a parameter for the container exec resize.
type SetContainerExecResizeOption func(*container.ResizeOptions) error

// WithHeight sets the height for the container exec resize options.
func WithSize(width, height uint) SetContainerExecResizeOption {
	return func(o *container.ResizeOptions) error {
		o.Height = height
		o.Width = width
		return nil
	}
}
