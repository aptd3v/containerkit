package compose

import (
	"errors"
	"fmt"
)

var (
	ErrComposeError       = fmt.Errorf("compose error")
	ErrComposeFlagError   = fmt.Errorf("compose flag error")
	ErrComposeUpError     = fmt.Errorf("compose up error")
	ErrComposeDownError   = fmt.Errorf("compose down error")
	ErrComposeLogsError   = fmt.Errorf("compose logs error")
	ErrComposeKillError   = fmt.Errorf("compose kill error")
	ErrComposeEventsError = fmt.Errorf("compose events error")
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

// ComposeKillError is the error for the compose kill command

type ComposeKillError struct {
	Message string
}

func (e *ComposeKillError) Unwrap() error {
	return ErrComposeKillError
}

func (e *ComposeKillError) Error() string {
	return fmt.Sprintf("compose kill error: %s", e.Message)
}

func NewComposeKillError(err error) *ComposeKillError {
	return &ComposeKillError{
		Message: err.Error(),
	}
}

func IsComposeKillError(err error) bool {
	return errors.Is(err, ErrComposeKillError)
}

// ComposeEventsError is the error for the compose events command

type ComposeEventsError struct {
	Message string
}

func (e *ComposeEventsError) Unwrap() error {
	return ErrComposeEventsError
}

func (e *ComposeEventsError) Error() string {
	return fmt.Sprintf("compose events error: %s", e.Message)
}

func NewComposeEventsError(err error) *ComposeEventsError {
	return &ComposeEventsError{
		Message: err.Error(),
	}
}

func IsComposeEventsError(err error) bool {
	return errors.Is(err, ErrComposeEventsError)
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
