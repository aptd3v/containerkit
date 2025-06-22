// Package clo provides the options for the container list options
package clo

import "github.com/docker/docker/api/types/container"

// SetContainerListOption is a function that sets the container list options.
type SetContainerListOption func(*container.ListOptions) error

// WithSize sets the container list options to include the size.
func WithSize() SetContainerListOption {
	return func(op *container.ListOptions) error {
		op.Size = true
		return nil
	}
}

// WithAll sets the container list options to include all containers.
func WithAll() SetContainerListOption {
	return func(op *container.ListOptions) error {
		op.All = true
		return nil
	}
}

// WithLatest sets the container list options to include the latest container.
func WithLatest() SetContainerListOption {
	return func(op *container.ListOptions) error {
		op.Latest = true
		return nil
	}
}

// WithSince sets the container list options to include the since container.
func WithSince(since string) SetContainerListOption {
	return func(op *container.ListOptions) error {
		op.Since = since
		return nil
	}
}

// WithBefore sets the container list options to include the before container.
func WithBefore(before string) SetContainerListOption {
	return func(op *container.ListOptions) error {
		op.Before = before
		return nil
	}
}

// WithLimit sets the container list options to include the limit.
func WithLimit(limit int) SetContainerListOption {
	return func(op *container.ListOptions) error {
		op.Limit = limit
		return nil
	}
}

// WithFilters sets the container list options to include the filters.
func WithFilters(key, value string) SetContainerListOption {
	return func(op *container.ListOptions) error {
		op.Filters.Add(key, value)
		return nil
	}
}
