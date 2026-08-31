package containerkit

import (
	"errors"
	"fmt"
)

// Errors for the compose command and its setters.
var (
	ErrComposeError        = fmt.Errorf("compose error")
	ErrComposeFlagError    = fmt.Errorf("compose flag error")
	ErrComposeUpError      = fmt.Errorf("compose up error")
	ErrComposeDownError    = fmt.Errorf("compose down error")
	ErrComposeLogsError    = fmt.Errorf("compose logs error")
	ErrComposeKillError    = fmt.Errorf("compose kill error")
	ErrComposeEventsError  = fmt.Errorf("compose events error")
	ErrComposePsError      = fmt.Errorf("compose ps error")
	ErrComposeStartError   = fmt.Errorf("compose start error")
	ErrComposeStopError    = fmt.Errorf("compose stop error")
	ErrComposeRestartError = fmt.Errorf("compose restart error")
	ErrComposeBuildError   = fmt.Errorf("compose build error")
	ErrComposePullError    = fmt.Errorf("compose pull error")
	ErrComposeExecError    = fmt.Errorf("compose exec error")
	ErrComposeAttachError  = fmt.Errorf("compose attach error")
	ErrComposeCommitError  = fmt.Errorf("compose commit error")
	ErrComposeConfigError  = fmt.Errorf("compose config error")
	ErrComposeCpError      = fmt.Errorf("compose cp error")
	ErrComposeCreateError  = fmt.Errorf("compose create error")
	ErrComposeExportError  = fmt.Errorf("compose export error")
	ErrComposeImagesError  = fmt.Errorf("compose images error")
	ErrComposeLsError      = fmt.Errorf("compose ls error")
	ErrComposePauseError   = fmt.Errorf("compose pause error")
	ErrComposePortError    = fmt.Errorf("compose port error")
	ErrComposePublishError = fmt.Errorf("compose publish error")
	ErrComposePushError    = fmt.Errorf("compose push error")
	ErrComposeRmError      = fmt.Errorf("compose rm error")
	ErrComposeRunError     = fmt.Errorf("compose run error")
	ErrComposeScaleError   = fmt.Errorf("compose scale error")
	ErrComposeStatsError   = fmt.Errorf("compose stats error")
	ErrComposeTopError     = fmt.Errorf("compose top error")
	ErrComposeUnpauseError = fmt.Errorf("compose unpause error")
	ErrComposeVersionError = fmt.Errorf("compose version error")
	ErrComposeVolumesError = fmt.Errorf("compose volumes error")
	ErrComposeWaitError    = fmt.Errorf("compose wait error")
	ErrComposeWatchError   = fmt.Errorf("compose watch error")
)

func causeMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func joinCause(sentinel, err error) error {
	return errors.Join(sentinel, err)
}

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
	err     error
}

func (e *ComposeKillError) Unwrap() error {
	return joinCause(ErrComposeKillError, e.err)
}

func (e *ComposeKillError) Error() string {
	return fmt.Sprintf("compose kill error: %s", e.Message)
}

// NewComposeKillError creates a new ComposeKillError wrapping err.
func NewComposeKillError(err error) *ComposeKillError {
	return &ComposeKillError{Message: causeMessage(err), err: err}
}

// IsComposeKillError checks if the error is a ComposeKillError.
func IsComposeKillError(err error) bool {
	return errors.Is(err, ErrComposeKillError)
}

// ComposeEventsError is the error for the compose events command
type ComposeEventsError struct {
	Message string
	err     error
}

func (e *ComposeEventsError) Unwrap() error {
	return joinCause(ErrComposeEventsError, e.err)
}

func (e *ComposeEventsError) Error() string {
	return fmt.Sprintf("compose events error: %s", e.Message)
}

// NewComposeEventsError creates a new ComposeEventsError wrapping err.
func NewComposeEventsError(err error) *ComposeEventsError {
	return &ComposeEventsError{Message: causeMessage(err), err: err}
}

// IsComposeEventsError checks if the error is a ComposeEventsError.
func IsComposeEventsError(err error) bool {
	return errors.Is(err, ErrComposeEventsError)
}

// ComposeError is the error for the compose command
type ComposeError struct {
	Message string
	err     error
}

func (e *ComposeError) Unwrap() error {
	return joinCause(ErrComposeError, e.err)
}

func (e *ComposeError) Error() string {
	return fmt.Sprintf("compose exec error: %s", e.Message)
}

func NewComposeError(err error) *ComposeError {
	return &ComposeError{Message: causeMessage(err), err: err}
}

// IsComposeError checks if the error is a ComposeError.
func IsComposeError(err error) bool {
	return errors.Is(err, ErrComposeError)
}

// ComposeUpError is the error for the compose up command
type ComposeUpError struct {
	Message string
	err     error
}

func (e *ComposeUpError) Unwrap() error {
	return joinCause(ErrComposeUpError, e.err)
}

func (e *ComposeUpError) Error() string {
	return fmt.Sprintf("compose up error: %s", e.Message)
}

// NewComposeUpError creates a new ComposeUpError wrapping err.
func NewComposeUpError(err error) *ComposeUpError {
	return &ComposeUpError{Message: causeMessage(err), err: err}
}

// IsComposeUpError checks if the error is a ComposeUpError.
func IsComposeUpError(err error) bool {
	return errors.Is(err, ErrComposeUpError)
}

// ComposeDownError is the error for the compose down command
type ComposeDownError struct {
	Message string
	err     error
}

func (e *ComposeDownError) Unwrap() error {
	return joinCause(ErrComposeDownError, e.err)
}

func (e *ComposeDownError) Error() string {
	return fmt.Sprintf("compose down error: %s", e.Message)
}

// NewComposeDownError creates a new ComposeDownError wrapping err.
func NewComposeDownError(err error) *ComposeDownError {
	return &ComposeDownError{Message: causeMessage(err), err: err}
}

// IsComposeDownError checks if the error is a ComposeDownError.
func IsComposeDownError(err error) bool {
	return errors.Is(err, ErrComposeDownError)
}

// ComposeLogsError is the error for the compose logs command
type ComposeLogsError struct {
	Message string
	err     error
}

func (e *ComposeLogsError) Unwrap() error {
	return joinCause(ErrComposeLogsError, e.err)
}

func (e *ComposeLogsError) Error() string {
	return fmt.Sprintf("compose logs error: %s", e.Message)
}

// NewComposeLogsError creates a new ComposeLogsError wrapping err.
func NewComposeLogsError(err error) *ComposeLogsError {
	return &ComposeLogsError{Message: causeMessage(err), err: err}
}

// IsComposeLogsError checks if the error is a ComposeLogsError.
func IsComposeLogsError(err error) bool {
	return errors.Is(err, ErrComposeLogsError)
}

// ComposePsError is the error for the compose ps command
type ComposePsError struct {
	Message string
	err     error
}

func (e *ComposePsError) Unwrap() error { return joinCause(ErrComposePsError, e.err) }
func (e *ComposePsError) Error() string { return fmt.Sprintf("compose ps error: %s", e.Message) }

func NewComposePsError(err error) *ComposePsError {
	return &ComposePsError{Message: causeMessage(err), err: err}
}
func IsComposePsError(err error) bool { return errors.Is(err, ErrComposePsError) }

// ComposeStartError is the error for the compose start command
type ComposeStartError struct {
	Message string
	err     error
}

func (e *ComposeStartError) Unwrap() error { return joinCause(ErrComposeStartError, e.err) }
func (e *ComposeStartError) Error() string { return fmt.Sprintf("compose start error: %s", e.Message) }

func NewComposeStartError(err error) *ComposeStartError {
	return &ComposeStartError{Message: causeMessage(err), err: err}
}
func IsComposeStartError(err error) bool { return errors.Is(err, ErrComposeStartError) }

// ComposeStopError is the error for the compose stop command
type ComposeStopError struct {
	Message string
	err     error
}

func (e *ComposeStopError) Unwrap() error { return joinCause(ErrComposeStopError, e.err) }
func (e *ComposeStopError) Error() string { return fmt.Sprintf("compose stop error: %s", e.Message) }

func NewComposeStopError(err error) *ComposeStopError {
	return &ComposeStopError{Message: causeMessage(err), err: err}
}
func IsComposeStopError(err error) bool { return errors.Is(err, ErrComposeStopError) }

// ComposeRestartError is the error for the compose restart command
type ComposeRestartError struct {
	Message string
	err     error
}

func (e *ComposeRestartError) Unwrap() error { return joinCause(ErrComposeRestartError, e.err) }
func (e *ComposeRestartError) Error() string {
	return fmt.Sprintf("compose restart error: %s", e.Message)
}

func NewComposeRestartError(err error) *ComposeRestartError {
	return &ComposeRestartError{Message: causeMessage(err), err: err}
}
func IsComposeRestartError(err error) bool { return errors.Is(err, ErrComposeRestartError) }

// ComposeBuildError is the error for the compose build command
type ComposeBuildError struct {
	Message string
	err     error
}

func (e *ComposeBuildError) Unwrap() error { return joinCause(ErrComposeBuildError, e.err) }
func (e *ComposeBuildError) Error() string { return fmt.Sprintf("compose build error: %s", e.Message) }

func NewComposeBuildError(err error) *ComposeBuildError {
	return &ComposeBuildError{Message: causeMessage(err), err: err}
}
func IsComposeBuildError(err error) bool { return errors.Is(err, ErrComposeBuildError) }

// ComposePullError is the error for the compose pull command
type ComposePullError struct {
	Message string
	err     error
}

func (e *ComposePullError) Unwrap() error { return joinCause(ErrComposePullError, e.err) }
func (e *ComposePullError) Error() string { return fmt.Sprintf("compose pull error: %s", e.Message) }

func NewComposePullError(err error) *ComposePullError {
	return &ComposePullError{Message: causeMessage(err), err: err}
}
func IsComposePullError(err error) bool { return errors.Is(err, ErrComposePullError) }

// ComposeExecError is the error for the compose exec command
type ComposeExecError struct {
	Message string
	err     error
}

func (e *ComposeExecError) Unwrap() error { return joinCause(ErrComposeExecError, e.err) }
func (e *ComposeExecError) Error() string { return fmt.Sprintf("compose exec error: %s", e.Message) }

func NewComposeExecError(err error) *ComposeExecError {
	return &ComposeExecError{Message: causeMessage(err), err: err}
}
func IsComposeExecError(err error) bool { return errors.Is(err, ErrComposeExecError) }

type ComposeAttachError struct {
	Message string
	err     error
}

func (e *ComposeAttachError) Unwrap() error { return joinCause(ErrComposeAttachError, e.err) }
func (e *ComposeAttachError) Error() string {
	return fmt.Sprintf("compose attach error: %s", e.Message)
}
func NewComposeAttachError(err error) *ComposeAttachError {
	return &ComposeAttachError{Message: causeMessage(err), err: err}
}
func IsComposeAttachError(err error) bool { return errors.Is(err, ErrComposeAttachError) }

type ComposeCommitError struct {
	Message string
	err     error
}

func (e *ComposeCommitError) Unwrap() error { return joinCause(ErrComposeCommitError, e.err) }
func (e *ComposeCommitError) Error() string {
	return fmt.Sprintf("compose commit error: %s", e.Message)
}
func NewComposeCommitError(err error) *ComposeCommitError {
	return &ComposeCommitError{Message: causeMessage(err), err: err}
}
func IsComposeCommitError(err error) bool { return errors.Is(err, ErrComposeCommitError) }

type ComposeConfigError struct {
	Message string
	err     error
}

func (e *ComposeConfigError) Unwrap() error { return joinCause(ErrComposeConfigError, e.err) }
func (e *ComposeConfigError) Error() string {
	return fmt.Sprintf("compose config error: %s", e.Message)
}
func NewComposeConfigError(err error) *ComposeConfigError {
	return &ComposeConfigError{Message: causeMessage(err), err: err}
}
func IsComposeConfigError(err error) bool { return errors.Is(err, ErrComposeConfigError) }

type ComposeCpError struct {
	Message string
	err     error
}

func (e *ComposeCpError) Unwrap() error { return joinCause(ErrComposeCpError, e.err) }
func (e *ComposeCpError) Error() string { return fmt.Sprintf("compose cp error: %s", e.Message) }
func NewComposeCpError(err error) *ComposeCpError {
	return &ComposeCpError{Message: causeMessage(err), err: err}
}
func IsComposeCpError(err error) bool { return errors.Is(err, ErrComposeCpError) }

type ComposeCreateError struct {
	Message string
	err     error
}

func (e *ComposeCreateError) Unwrap() error { return joinCause(ErrComposeCreateError, e.err) }
func (e *ComposeCreateError) Error() string {
	return fmt.Sprintf("compose create error: %s", e.Message)
}
func NewComposeCreateError(err error) *ComposeCreateError {
	return &ComposeCreateError{Message: causeMessage(err), err: err}
}
func IsComposeCreateError(err error) bool { return errors.Is(err, ErrComposeCreateError) }

type ComposeExportError struct {
	Message string
	err     error
}

func (e *ComposeExportError) Unwrap() error { return joinCause(ErrComposeExportError, e.err) }
func (e *ComposeExportError) Error() string {
	return fmt.Sprintf("compose export error: %s", e.Message)
}
func NewComposeExportError(err error) *ComposeExportError {
	return &ComposeExportError{Message: causeMessage(err), err: err}
}
func IsComposeExportError(err error) bool { return errors.Is(err, ErrComposeExportError) }

type ComposeImagesError struct {
	Message string
	err     error
}

func (e *ComposeImagesError) Unwrap() error { return joinCause(ErrComposeImagesError, e.err) }
func (e *ComposeImagesError) Error() string {
	return fmt.Sprintf("compose images error: %s", e.Message)
}
func NewComposeImagesError(err error) *ComposeImagesError {
	return &ComposeImagesError{Message: causeMessage(err), err: err}
}
func IsComposeImagesError(err error) bool { return errors.Is(err, ErrComposeImagesError) }

type ComposeLsError struct {
	Message string
	err     error
}

func (e *ComposeLsError) Unwrap() error { return joinCause(ErrComposeLsError, e.err) }
func (e *ComposeLsError) Error() string { return fmt.Sprintf("compose ls error: %s", e.Message) }
func NewComposeLsError(err error) *ComposeLsError {
	return &ComposeLsError{Message: causeMessage(err), err: err}
}
func IsComposeLsError(err error) bool { return errors.Is(err, ErrComposeLsError) }

type ComposePauseError struct {
	Message string
	err     error
}

func (e *ComposePauseError) Unwrap() error { return joinCause(ErrComposePauseError, e.err) }
func (e *ComposePauseError) Error() string { return fmt.Sprintf("compose pause error: %s", e.Message) }
func NewComposePauseError(err error) *ComposePauseError {
	return &ComposePauseError{Message: causeMessage(err), err: err}
}
func IsComposePauseError(err error) bool { return errors.Is(err, ErrComposePauseError) }

type ComposePortError struct {
	Message string
	err     error
}

func (e *ComposePortError) Unwrap() error { return joinCause(ErrComposePortError, e.err) }
func (e *ComposePortError) Error() string { return fmt.Sprintf("compose port error: %s", e.Message) }
func NewComposePortError(err error) *ComposePortError {
	return &ComposePortError{Message: causeMessage(err), err: err}
}
func IsComposePortError(err error) bool { return errors.Is(err, ErrComposePortError) }

type ComposePublishError struct {
	Message string
	err     error
}

func (e *ComposePublishError) Unwrap() error { return joinCause(ErrComposePublishError, e.err) }
func (e *ComposePublishError) Error() string {
	return fmt.Sprintf("compose publish error: %s", e.Message)
}
func NewComposePublishError(err error) *ComposePublishError {
	return &ComposePublishError{Message: causeMessage(err), err: err}
}
func IsComposePublishError(err error) bool { return errors.Is(err, ErrComposePublishError) }

type ComposePushError struct {
	Message string
	err     error
}

func (e *ComposePushError) Unwrap() error { return joinCause(ErrComposePushError, e.err) }
func (e *ComposePushError) Error() string { return fmt.Sprintf("compose push error: %s", e.Message) }
func NewComposePushError(err error) *ComposePushError {
	return &ComposePushError{Message: causeMessage(err), err: err}
}
func IsComposePushError(err error) bool { return errors.Is(err, ErrComposePushError) }

type ComposeRmError struct {
	Message string
	err     error
}

func (e *ComposeRmError) Unwrap() error { return joinCause(ErrComposeRmError, e.err) }
func (e *ComposeRmError) Error() string { return fmt.Sprintf("compose rm error: %s", e.Message) }
func NewComposeRmError(err error) *ComposeRmError {
	return &ComposeRmError{Message: causeMessage(err), err: err}
}
func IsComposeRmError(err error) bool { return errors.Is(err, ErrComposeRmError) }

type ComposeRunError struct {
	Message string
	err     error
}

func (e *ComposeRunError) Unwrap() error { return joinCause(ErrComposeRunError, e.err) }
func (e *ComposeRunError) Error() string { return fmt.Sprintf("compose run error: %s", e.Message) }
func NewComposeRunError(err error) *ComposeRunError {
	return &ComposeRunError{Message: causeMessage(err), err: err}
}
func IsComposeRunError(err error) bool { return errors.Is(err, ErrComposeRunError) }

type ComposeScaleError struct {
	Message string
	err     error
}

func (e *ComposeScaleError) Unwrap() error { return joinCause(ErrComposeScaleError, e.err) }
func (e *ComposeScaleError) Error() string { return fmt.Sprintf("compose scale error: %s", e.Message) }
func NewComposeScaleError(err error) *ComposeScaleError {
	return &ComposeScaleError{Message: causeMessage(err), err: err}
}
func IsComposeScaleError(err error) bool { return errors.Is(err, ErrComposeScaleError) }

type ComposeStatsError struct {
	Message string
	err     error
}

func (e *ComposeStatsError) Unwrap() error { return joinCause(ErrComposeStatsError, e.err) }
func (e *ComposeStatsError) Error() string { return fmt.Sprintf("compose stats error: %s", e.Message) }
func NewComposeStatsError(err error) *ComposeStatsError {
	return &ComposeStatsError{Message: causeMessage(err), err: err}
}
func IsComposeStatsError(err error) bool { return errors.Is(err, ErrComposeStatsError) }

type ComposeTopError struct {
	Message string
	err     error
}

func (e *ComposeTopError) Unwrap() error { return joinCause(ErrComposeTopError, e.err) }
func (e *ComposeTopError) Error() string { return fmt.Sprintf("compose top error: %s", e.Message) }
func NewComposeTopError(err error) *ComposeTopError {
	return &ComposeTopError{Message: causeMessage(err), err: err}
}
func IsComposeTopError(err error) bool { return errors.Is(err, ErrComposeTopError) }

type ComposeUnpauseError struct {
	Message string
	err     error
}

func (e *ComposeUnpauseError) Unwrap() error { return joinCause(ErrComposeUnpauseError, e.err) }
func (e *ComposeUnpauseError) Error() string {
	return fmt.Sprintf("compose unpause error: %s", e.Message)
}
func NewComposeUnpauseError(err error) *ComposeUnpauseError {
	return &ComposeUnpauseError{Message: causeMessage(err), err: err}
}
func IsComposeUnpauseError(err error) bool { return errors.Is(err, ErrComposeUnpauseError) }

type ComposeVersionError struct {
	Message string
	err     error
}

func (e *ComposeVersionError) Unwrap() error { return joinCause(ErrComposeVersionError, e.err) }
func (e *ComposeVersionError) Error() string {
	return fmt.Sprintf("compose version error: %s", e.Message)
}
func NewComposeVersionError(err error) *ComposeVersionError {
	return &ComposeVersionError{Message: causeMessage(err), err: err}
}
func IsComposeVersionError(err error) bool { return errors.Is(err, ErrComposeVersionError) }

type ComposeVolumesError struct {
	Message string
	err     error
}

func (e *ComposeVolumesError) Unwrap() error { return joinCause(ErrComposeVolumesError, e.err) }
func (e *ComposeVolumesError) Error() string {
	return fmt.Sprintf("compose volumes error: %s", e.Message)
}
func NewComposeVolumesError(err error) *ComposeVolumesError {
	return &ComposeVolumesError{Message: causeMessage(err), err: err}
}
func IsComposeVolumesError(err error) bool { return errors.Is(err, ErrComposeVolumesError) }

type ComposeWaitError struct {
	Message string
	err     error
}

func (e *ComposeWaitError) Unwrap() error { return joinCause(ErrComposeWaitError, e.err) }
func (e *ComposeWaitError) Error() string { return fmt.Sprintf("compose wait error: %s", e.Message) }
func NewComposeWaitError(err error) *ComposeWaitError {
	return &ComposeWaitError{Message: causeMessage(err), err: err}
}
func IsComposeWaitError(err error) bool { return errors.Is(err, ErrComposeWaitError) }

type ComposeWatchError struct {
	Message string
	err     error
}

func (e *ComposeWatchError) Unwrap() error { return joinCause(ErrComposeWatchError, e.err) }
func (e *ComposeWatchError) Error() string { return fmt.Sprintf("compose watch error: %s", e.Message) }
func NewComposeWatchError(err error) *ComposeWatchError {
	return &ComposeWatchError{Message: causeMessage(err), err: err}
}
func IsComposeWatchError(err error) bool { return errors.Is(err, ErrComposeWatchError) }
