// package execstart provides options for the container exec start.
package execstart

import "github.com/docker/docker/api/types/container"

// SetContainerExecStartOption is a function that sets a parameter for the container exec start.
type SetContainerExecStartOption func(*container.ExecStartOptions) error

// WithDetach sets the detach flag for the exec start options.
// ExecStart will first check if it's detached
func WithDetach() SetContainerExecStartOption {

	return func(o *container.ExecStartOptions) error {
		o.Detach = true
		return nil
	}
}

// WithTty sets the tty flag for the exec start options.
// Check if there's a tty
func WithTty() SetContainerExecStartOption {

	return func(o *container.ExecStartOptions) error {
		o.Tty = true
		return nil
	}
}

// WithConsoleSize sets the console size for the exec start options.
// Terminal size [height, width], unused if Tty == false
func WithConsoleSize(width, height uint) SetContainerExecStartOption {
	return func(o *container.ExecStartOptions) error {
		o.ConsoleSize = &[2]uint{width, height}
		return nil
	}
}
