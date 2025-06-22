// package cctco provides options for the container copy to container.
package cctco

import "github.com/docker/docker/api/types/container"

// SetContainerCopyToContainerOption is a function that sets a parameter for the container copy to container.
type SetContainerCopyToContainerOption func(*container.CopyToContainerOptions) error

// WithAllowOverwriteDirWithFile sets the allow overwrite dir with file.
func WithAllowOverwriteDirWithFile() SetContainerCopyToContainerOption {
	return func(o *container.CopyToContainerOptions) error {
		o.AllowOverwriteDirWithFile = true
		return nil
	}
}
func WithCopyUIDGID() SetContainerCopyToContainerOption {
	return func(o *container.CopyToContainerOptions) error {
		o.CopyUIDGID = true
		return nil
	}
}
