package compose

import (
	"errors"
	"fmt"
)

var (
	ErrComposeError     = fmt.Errorf("compose error")
	ErrComposeFlagError = fmt.Errorf("compose flag error")
	ErrComposeUpError   = fmt.Errorf("compose up error")
	ErrComposeDownError = fmt.Errorf("compose down error")
	ErrComposeLogsError = fmt.Errorf("compose logs error")
)

// ComposeFlagError is the error for the compose flag
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

// ComposeError is the error for the compose command
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

// ComposeUpError is the error for the compose up command
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

// ComposeDownError is the error for the compose down command
type ComposeDownError struct {
	Message string
}

func (e *ComposeDownError) Unwrap() error {
	return ErrComposeDownError
}

func (e *ComposeDownError) Error() string {
	return fmt.Sprintf("compose down error: %s", e.Message)
}

func NewComposeDownError(err error) *ComposeDownError {
	return &ComposeDownError{
		Message: err.Error(),
	}
}

func IsComposeDownError(err error) bool {
	return errors.Is(err, ErrComposeDownError)
}

// ComposeLogsError is the error for the compose logs command
type ComposeLogsError struct {
	Message string
}

func (e *ComposeLogsError) Unwrap() error {
	return ErrComposeLogsError
}

func (e *ComposeLogsError) Error() string {
	return fmt.Sprintf("compose logs error: %s", e.Message)
}

func NewComposeLogsError(err error) *ComposeLogsError {
	return &ComposeLogsError{
		Message: err.Error(),
	}
}

func IsComposeLogsError(err error) bool {
	return errors.Is(err, ErrComposeLogsError)
}
