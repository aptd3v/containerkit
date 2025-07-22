package compose

import (
	"errors"
	"fmt"
)

// Errors for the compose command and its setters.
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

// NewComposeFlagError makes a new ComposeFlagError with the given flag and message.
func NewComposeFlagError(flag, message string) *ComposeFlagError {
	return &ComposeFlagError{
		Flag:    flag,
		Message: message,
	}
}

// IsComposeFlagError checks if the error is a ComposeFlagError.
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

// NewComposeKillError creates a new ComposeKillError with the given error message.
func NewComposeKillError(err error) *ComposeKillError {
	return &ComposeKillError{
		Message: err.Error(),
	}
}

// IsComposeKillError checks if the error is a ComposeKillError.
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

// NewComposeEventsError creates a new ComposeEventsError with the given error message.
func NewComposeEventsError(err error) *ComposeEventsError {
	return &ComposeEventsError{
		Message: err.Error(),
	}
}

// IsComposeEventsError checks if the error is a ComposeEventsError.
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

// IsComposeError checks if the error is a ComposeError.
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

// NewComposeUpError creates a new ComposeUpError with the given error message.
func NewComposeUpError(err error) *ComposeUpError {
	return &ComposeUpError{
		Message: err.Error(),
	}
}

// IsComposeUpError checks if the error is a ComposeUpError.
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

// NewComposeDownError creates a new ComposeDownError with the given error message.
func NewComposeDownError(err error) *ComposeDownError {
	return &ComposeDownError{
		Message: err.Error(),
	}
}

// IsComposeDownError checks if the error is a ComposeDownError.
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

// NewComposeLogsError creates a new ComposeLogsError with the given error message.
func NewComposeLogsError(err error) *ComposeLogsError {
	return &ComposeLogsError{
		Message: err.Error(),
	}
}

// IsComposeLogsError checks if the error is a ComposeLogsError.
func IsComposeLogsError(err error) bool {
	return errors.Is(err, ErrComposeLogsError)
}
