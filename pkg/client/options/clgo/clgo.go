// package clgo provides options for the container logs.
package clgo

import "github.com/docker/docker/api/types/container"

// SetContainerLogsOption is a function that sets a parameter for the container logs.
type SetContainerLogsOption func(*container.LogsOptions) error

// WithShowStdout sets the showStdout parameter for the container logs.
func WithShowStdout() SetContainerLogsOption {
	return func(o *container.LogsOptions) error {
		o.ShowStdout = true
		return nil
	}
}

// WithShowStderr sets the showStderr parameter for the container logs.
func WithShowStderr() SetContainerLogsOption {
	return func(o *container.LogsOptions) error {
		o.ShowStderr = true
		return nil
	}
}

// WithSince sets the since parameter for the container logs.
func WithSince(since string) SetContainerLogsOption {
	return func(o *container.LogsOptions) error {
		o.Since = since
		return nil
	}
}

// WithFollow sets the follow parameter for the container logs.
func WithFollow() SetContainerLogsOption {
	return func(o *container.LogsOptions) error {
		o.Follow = true
		return nil
	}
}

// WithTail sets the tail parameter for the container logs.
func WithTail(tail string) SetContainerLogsOption {
	return func(o *container.LogsOptions) error {
		o.Tail = tail
		return nil
	}
}

// WithTimestamps sets the timestamps parameter for the container logs.
func WithTimestamps() SetContainerLogsOption {
	return func(o *container.LogsOptions) error {
		o.Timestamps = true
		return nil
	}
}

// WithDetails sets the details parameter for the container logs.
func WithDetails() SetContainerLogsOption {
	return func(o *container.LogsOptions) error {
		o.Details = true
		return nil
	}
}

// WithUntil sets the until parameter for the container logs.
func WithUntil(until string) SetContainerLogsOption {
	return func(o *container.LogsOptions) error {
		o.Until = until
		return nil
	}
}
