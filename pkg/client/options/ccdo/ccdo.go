// package ccdo provides options for the container checkpoint delete.
package ccdo

import "github.com/docker/docker/api/types/checkpoint"

// SetContainerCheckpointDeleteOption is a function that sets a parameter for the container checkpoint delete.
type SetContainerCheckpointDeleteOption func(*checkpoint.DeleteOptions) error

// WithCheckpointDir sets the checkpoint dir.
func WithCheckpointDir(dir string) SetContainerCheckpointDeleteOption {
	return func(o *checkpoint.DeleteOptions) error {
		o.CheckpointDir = dir
		return nil
	}
}

// WithCheckpointID sets the checkpoint id.
func WithCheckpointID(id string) SetContainerCheckpointDeleteOption {
	return func(o *checkpoint.DeleteOptions) error {
		o.CheckpointID = id
		return nil
	}
}
