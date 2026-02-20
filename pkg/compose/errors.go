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
	ErrComposeKillError    = fmt.Errorf("compose kill error")
	ErrComposeEventsError  = fmt.Errorf("compose events error")
	ErrComposePsError      = fmt.Errorf("compose ps error")
	ErrComposeStartError   = fmt.Errorf("compose start error")
	ErrComposeStopError    = fmt.Errorf("compose stop error")
	ErrComposeRestartError = fmt.Errorf("compose restart error")
	ErrComposeBuildError  = fmt.Errorf("compose build error")
	ErrComposePullError   = fmt.Errorf("compose pull error")
	ErrComposeExecError   = fmt.Errorf("compose exec error")
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

// ComposePsError is the error for the compose ps command
type ComposePsError struct {
	Message string
}

func (e *ComposePsError) Unwrap() error { return ErrComposePsError }
func (e *ComposePsError) Error() string { return fmt.Sprintf("compose ps error: %s", e.Message) }

func NewComposePsError(err error) *ComposePsError {
	return &ComposePsError{Message: err.Error()}
}
func IsComposePsError(err error) bool { return errors.Is(err, ErrComposePsError) }

// ComposeStartError is the error for the compose start command
type ComposeStartError struct {
	Message string
}

func (e *ComposeStartError) Unwrap() error { return ErrComposeStartError }
func (e *ComposeStartError) Error() string { return fmt.Sprintf("compose start error: %s", e.Message) }

func NewComposeStartError(err error) *ComposeStartError {
	return &ComposeStartError{Message: err.Error()}
}
func IsComposeStartError(err error) bool { return errors.Is(err, ErrComposeStartError) }

// ComposeStopError is the error for the compose stop command
type ComposeStopError struct {
	Message string
}

func (e *ComposeStopError) Unwrap() error { return ErrComposeStopError }
func (e *ComposeStopError) Error() string { return fmt.Sprintf("compose stop error: %s", e.Message) }

func NewComposeStopError(err error) *ComposeStopError {
	return &ComposeStopError{Message: err.Error()}
}
func IsComposeStopError(err error) bool { return errors.Is(err, ErrComposeStopError) }

// ComposeRestartError is the error for the compose restart command
type ComposeRestartError struct {
	Message string
}

func (e *ComposeRestartError) Unwrap() error { return ErrComposeRestartError }
func (e *ComposeRestartError) Error() string { return fmt.Sprintf("compose restart error: %s", e.Message) }

func NewComposeRestartError(err error) *ComposeRestartError {
	return &ComposeRestartError{Message: err.Error()}
}
func IsComposeRestartError(err error) bool { return errors.Is(err, ErrComposeRestartError) }

// ComposeBuildError is the error for the compose build command
type ComposeBuildError struct {
	Message string
}

func (e *ComposeBuildError) Unwrap() error { return ErrComposeBuildError }
func (e *ComposeBuildError) Error() string { return fmt.Sprintf("compose build error: %s", e.Message) }

func NewComposeBuildError(err error) *ComposeBuildError {
	return &ComposeBuildError{Message: err.Error()}
}
func IsComposeBuildError(err error) bool { return errors.Is(err, ErrComposeBuildError) }

// ComposePullError is the error for the compose pull command
type ComposePullError struct {
	Message string
}

func (e *ComposePullError) Unwrap() error { return ErrComposePullError }
func (e *ComposePullError) Error() string { return fmt.Sprintf("compose pull error: %s", e.Message) }

func NewComposePullError(err error) *ComposePullError {
	return &ComposePullError{Message: err.Error()}
}
func IsComposePullError(err error) bool { return errors.Is(err, ErrComposePullError) }

// ComposeExecError is the error for the compose exec command
type ComposeExecError struct {
	Message string
}

func (e *ComposeExecError) Unwrap() error { return ErrComposeExecError }
func (e *ComposeExecError) Error() string { return fmt.Sprintf("compose exec error: %s", e.Message) }

func NewComposeExecError(err error) *ComposeExecError {
	return &ComposeExecError{Message: err.Error()}
}
func IsComposeExecError(err error) bool { return errors.Is(err, ErrComposeExecError) }
