// package cco provides options for the container commit.
package cco

import (
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/docker/docker/api/types/container"
)

// SetContainerCommitOption is a function that sets a parameter for the container commit.
type SetContainerCommitOption func(*container.CommitOptions) error

// WithReference sets the reference for the container commit options.
func WithReference(reference string) SetContainerCommitOption {
	return func(o *container.CommitOptions) error {
		o.Reference = reference
		return nil
	}
}

// WithComment sets the comment for the container commit options.
func WithComment(comment string) SetContainerCommitOption {
	return func(o *container.CommitOptions) error {
		o.Comment = comment
		return nil
	}
}

// WithAuthor sets the author for the container commit options.
func WithAuthor(author string) SetContainerCommitOption {
	return func(o *container.CommitOptions) error {
		o.Author = author
		return nil
	}
}

// WithChanges appends the changes for the container commit options.
func WithChanges(changes ...string) SetContainerCommitOption {
	return func(o *container.CommitOptions) error {
		if o.Changes == nil {
			o.Changes = []string{}
		}
		o.Changes = append(o.Changes, changes...)
		return nil
	}
}

// WithPause sets the pause flag for the container commit options.
func WithPause(pause bool) SetContainerCommitOption {
	return func(o *container.CommitOptions) error {
		o.Pause = pause
		return nil
	}
}

// WithConfig sets the config for the container commit options.
func WithConfig(setters ...create.SetContainerConfig) SetContainerCommitOption {
	config := create.ContainerConfig{
		Config: &container.Config{},
	}
	return func(o *container.CommitOptions) error {
		if o.Config == nil {
			o.Config = config.Config
		}
		for _, setter := range setters {
			if setter != nil {
				if err := setter(&config); err != nil {
					return err
				}
			}
		}
		return nil
	}
}
