package kit

import (
	"errors"
	"fmt"

	templatetypes "github.com/aptd3v/containerkit/internal/compose/template/types"
	commontypes "github.com/aptd3v/containerkit/internal/engine/common/types"
	containertypes "github.com/aptd3v/containerkit/internal/engine/container/types"
	jsonstreamtypes "github.com/aptd3v/containerkit/internal/engine/jsonstream/types"
)

// Engine and Compose error payloads, aliased so callers never import those SDKs.
type (
	JSONError            = jsonstreamtypes.Error
	ErrorResponse        = commontypes.ErrorResponse
	WaitExitError        = containertypes.WaitExitError
	WaitResponse         = containertypes.WaitResponse
	InvalidTemplateError = templatetypes.InvalidTemplateError
	MissingRequiredError = templatetypes.MissingRequiredError
)

// ExitError is a [WaitResponse] error payload as a Go error.
type ExitError struct {
	StatusCode int64
	WaitExitError
}

func (e *ExitError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("container exited with status %d", e.StatusCode)
}

// ErrorFromResponse wraps an Engine [ErrorResponse] message.
func ErrorFromResponse(r ErrorResponse) error {
	if r.Message == "" {
		return nil
	}
	return cfgErr("engine", "", "ErrorResponse", errors.New(r.Message))
}

// ErrorFromWait wraps a [WaitResponse] error payload as [ExitError].
func ErrorFromWait(resp WaitResponse) error {
	if resp.Error == nil || resp.Error.Message == "" {
		return nil
	}
	return cfgErr("container", "wait", "Wait", &ExitError{
		StatusCode:    resp.StatusCode,
		WaitExitError: *resp.Error,
	})
}

func waitFailed(err error) ContainerWaitResult {
	errC := make(chan error, 1)
	resultC := make(chan WaitResponse)
	errC <- err
	return ContainerWaitResult{Result: resultC, Error: errC}
}
