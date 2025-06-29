package compose

import (
	"errors"
	"fmt"
)

var (
	ErrComposeError     = fmt.Errorf("compose error")
	ErrComposeFlagError = fmt.Errorf("compose flag error")
	ErrComposeUpError   = fmt.Errorf("compose up error")
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

type ComposeError struct {
	Message string
}

func (e *ComposeError) Unwrap() error {
	return ErrComposeError
}
func (e *ComposeError) Error() string {
	return fmt.Sprintf("compose exec error: %s", e.Message)
}

func NewComposeError(err error) *ComposeError {
	return &ComposeError{
		Message: err.Error(),
	}
}

func IsComposeError(err error) bool {
	return errors.Is(err, ErrComposeError)
}

type ComposeUpError struct {
	Message string
}

func (e *ComposeUpError) Unwrap() error {
	return ErrComposeUpError
}

func (e *ComposeUpError) Error() string {
	return fmt.Sprintf("compose up error: %s", e.Message)
}

func NewComposeUpError(err error) *ComposeUpError {
	return &ComposeUpError{
		Message: err.Error(),
	}
}

func IsComposeUpError(err error) bool {
	return errors.Is(err, ErrComposeUpError)
}
