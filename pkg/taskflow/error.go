package taskflow

import (
	"errors"
	"fmt"
)

var (
	// ErrNoClient is returned when a Container or Build step runs without a client.
	ErrNoClient = errors.New("docker client is nil")
	// ErrAlreadyRun is returned when Run is called a second time on the same Flow.
	ErrAlreadyRun = errors.New("already run")
)

// TaskError is a named step failure. Unwrap returns Err.
type TaskError struct {
	Task string
	Err  error
}

func (e *TaskError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return fmt.Sprintf("task %s", e.Task)
	}
	return fmt.Sprintf("task %s: %v", e.Task, e.Err)
}

func (e *TaskError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func taskErr(name string, err error) error {
	if err == nil {
		return nil
	}
	var te *TaskError
	if errors.As(err, &te) {
		return err
	}
	return &TaskError{Task: name, Err: err}
}
