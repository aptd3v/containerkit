// package cao provides options for the container attach.
package cao

import "github.com/docker/docker/api/types/container"

// SetContainerAttachOption is a function that sets a parameter for the container attach.
type SetContainerAttachOption func(*container.AttachOptions) error

// WithStream sets the stream flag for the container attach options.
func WithStream() SetContainerAttachOption {
	return func(o *container.AttachOptions) error {
		o.Stream = true
		return nil
	}
}

// WithStdin sets the stdin flag for the container attach options.
func WithStdin() SetContainerAttachOption {
	return func(o *container.AttachOptions) error {
		o.Stdin = true
		return nil
	}
}

// WithStdout sets the stdout flag for the container attach options.
func WithStdout() SetContainerAttachOption {
	return func(o *container.AttachOptions) error {
		o.Stdout = true
		return nil
	}
}

// WithStderr sets the stderr flag for the container attach options.
func WithStderr() SetContainerAttachOption {
	return func(o *container.AttachOptions) error {
		o.Stderr = true
		return nil
	}
}

// WithDetachKeys sets the detach keys for the container attach options.
func WithDetachKeys(detachKeys string) SetContainerAttachOption {
	return func(o *container.AttachOptions) error {
		o.DetachKeys = detachKeys
		return nil
	}
}

// WithLogs sets the logs flag for the container attach options.
func WithLogs() SetContainerAttachOption {
	return func(o *container.AttachOptions) error {
		o.Logs = true
		return nil
	}
}
