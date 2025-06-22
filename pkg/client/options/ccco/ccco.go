// package ccco provides options for the container checkpoint create.
package ccco

import "github.com/docker/docker/api/types/checkpoint"

// SetContainerCheckpointCreateOption is a function that sets a parameter for the container checkpoint create.
type SetContainerCheckpointCreateOption func(*checkpoint.CreateOptions) error

// WithCheckpointID sets the checkpoint id.
func WithCheckpointID(id string) SetContainerCheckpointCreateOption {
	return func(o *checkpoint.CreateOptions) error {
		o.CheckpointID = id
		return nil
	}
}

// WithCheckpointDir sets the checkpoint dir.
func WithCheckpointDir(dir string) SetContainerCheckpointCreateOption {
	return func(o *checkpoint.CreateOptions) error {
		o.CheckpointDir = dir
		return nil
	}
}

// WithExit sets the exit flag.
func WithExit() SetContainerCheckpointCreateOption {
	return func(o *checkpoint.CreateOptions) error {
		o.Exit = true
		return nil
	}
}
