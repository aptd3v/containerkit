// package checkpointlist provides options for the container checkpoint list.
package checkpointlist

import "github.com/docker/docker/api/types/checkpoint"

// SetContainerCheckpointListOption is a function that sets a parameter for the container checkpoint list.
type SetContainerCheckpointListOption func(*checkpoint.ListOptions) error

// WithCheckpointName sets the checkpoint name.
func WithCheckpointDir(dir string) SetContainerCheckpointListOption {
	return func(o *checkpoint.ListOptions) error {
		o.CheckpointDir = dir
		return nil
	}
}
