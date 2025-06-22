// package ceao provides options for the container exec attach.
package ceao

import "github.com/docker/docker/api/types/container"

// SetContainerExecAttachOption is a function that sets a parameter for the container exec attach.
type SetContainerExecAttachOption func(*container.ExecAttachOptions) error

// WithDetach sets the detach flag for the exec attach options.
// ExecAttach will first check if it's detached
func WithDetach() SetContainerExecAttachOption {
	return func(o *container.ExecAttachOptions) error {
		o.Detach = true
		return nil
	}
}

// WithTty sets the tty flag for the exec attach options.
// Check if there's a tty
func WithTty() SetContainerExecAttachOption {
	return func(o *container.ExecAttachOptions) error {
		o.Tty = true
		return nil
	}
}

// WithConsoleSize sets the console size for the exec attach options.
// Terminal size [height, width], unused if Tty == false
func WithConsoleSize(width, height uint) SetContainerExecAttachOption {
	return func(o *container.ExecAttachOptions) error {
		o.ConsoleSize = &[2]uint{width, height}
		return nil
	}
}
