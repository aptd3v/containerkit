// Package cstrto provides the options for the container start options
package cstrto

import "github.com/docker/docker/api/types/container"

// SetContainerStartOption is a function that sets the container start options.
type SetContainerStartOption func(*container.StartOptions) error

// WithCheckpointID sets the checkpoint ID for the container start options.
func WithCheckpointID(checkpointID string) SetContainerStartOption {
	return func(op *container.StartOptions) error {
		op.CheckpointID = checkpointID
		return nil
	}
}

// WithCheckpointDir sets the checkpoint directory for the container start options.
func WithCheckpointDir(checkpointDir string) SetContainerStartOption {
	return func(op *container.StartOptions) error {
		op.CheckpointDir = checkpointDir
		return nil
	}
}
