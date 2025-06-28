package compose

import (
	"errors"
	"fmt"
)

var (
	ErrComposeFlagError = fmt.Errorf("compose flag error")
	ErrComposeExecError = fmt.Errorf("compose exec error")
)

type ComposeFlagError struct {
	Flag    string
	Message string
}

func (e *ComposeFlagError) Unwrap() error {
	return ErrComposeFlagError
}

func (e *ComposeFlagError) Error() string {
	return fmt.Sprintf("compose flag error: %s: %s", e.Flag, e.Message)
}

func NewComposeFlagError(flag, message string) *ComposeFlagError {
	return &ComposeFlagError{
		Flag:    flag,
		Message: message,
	}
}
func IsComposeFlagError(err error) bool {
	return errors.Is(err, ErrComposeFlagError)
}

type ComposeExecError struct {
	Message string
}

func (e *ComposeExecError) Unwrap() error {
	return ErrComposeExecError
}
func (e *ComposeExecError) Error() string {
	return fmt.Sprintf("compose exec error: %s", e.Message)
}

func NewComposeExecError(err error) *ComposeExecError {
	return &ComposeExecError{
		Message: err.Error(),
	}
}

func IsComposeExecError(err error) bool {
	return errors.Is(err, ErrComposeExecError)
}
