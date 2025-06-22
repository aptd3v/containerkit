// Package cstpo provides the options for the container stop options
package cstpo

import "github.com/docker/docker/api/types/container"

// SetContainerStopOption is a function that sets the container stop options.
type SetContainerStopOption func(*container.StopOptions) error

// WithTimeout sets the timeout for the container stop options.
/*
Timeout (optional) is the timeout (in seconds) to wait for the container to stop gracefully before forcibly terminating it with SIGKILL.

- Use nil to use the default timeout (10 seconds).

- Use '-1' to wait indefinitely.

- Use '0' to not wait for the container to exit gracefully, and immediately proceeds to forcibly terminating the container.

- Other positive values are used as timeout (in seconds).
*/
func WithTimeout(timeout int) SetContainerStopOption {
	return func(op *container.StopOptions) error {
		op.Timeout = &timeout
		return nil
	}
}

// WithSignal sets the signal for the container stop options.
func WithSignal(signal string) SetContainerStopOption {
	return func(op *container.StopOptions) error {
		op.Signal = signal
		return nil
	}
}
