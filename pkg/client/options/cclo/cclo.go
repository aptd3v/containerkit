// package cclo provides options for the container checkpoint list.
package cclo

import "github.com/docker/docker/api/types/checkpoint"

// SetContainerListOption is a function that sets a parameter for the container list.
type SetContainerCheckpointListOption func(*checkpoint.ListOptions) error

// WithCheckpointName sets the checkpoint name.
func WithCheckpointDir(dir string) SetContainerCheckpointListOption {
	return func(o *checkpoint.ListOptions) error {
		o.CheckpointDir = dir
		return nil
	}
}
