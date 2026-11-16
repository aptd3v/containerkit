package kit

import (
	"context"
	"errors"
	"net/http"

	composeerrdefs "github.com/compose-spec/compose-go/v2/errdefs"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/errdefs/pkg/errhttp"
	composeapi "github.com/docker/compose/v5/pkg/api"
	engineclient "github.com/moby/moby/client"
)

// Engine errdefs sentinels. [errors.Is] matches both kit wrappers and raw
// Engine client errors.
var (
	ErrUnknown            = cerrdefs.ErrUnknown
	ErrInvalidArgument    = cerrdefs.ErrInvalidArgument
	ErrNotFound           = cerrdefs.ErrNotFound
	ErrAlreadyExists      = cerrdefs.ErrAlreadyExists
	ErrPermissionDenied   = cerrdefs.ErrPermissionDenied
	ErrResourceExhausted  = cerrdefs.ErrResourceExhausted
	ErrFailedPrecondition = cerrdefs.ErrFailedPrecondition
	ErrConflict           = cerrdefs.ErrConflict
	ErrNotModified        = cerrdefs.ErrNotModified
	ErrAborted            = cerrdefs.ErrAborted
	ErrOutOfRange         = cerrdefs.ErrOutOfRange
	ErrNotImplemented     = cerrdefs.ErrNotImplemented
	ErrInternal           = cerrdefs.ErrInternal
	ErrUnavailable        = cerrdefs.ErrUnavailable
	ErrDataLoss           = cerrdefs.ErrDataLoss
	ErrUnauthorized       = cerrdefs.ErrUnauthenticated
	ErrRedirect           = engineclient.ErrRedirect
)

// Compose SDK sentinels that have no Engine errdefs equivalent.
var (
	ErrCanceled        = composeapi.ErrCanceled
	ErrUnsupportedFlag = composeapi.ErrUnsupportedFlag
	ErrParsingFailed   = composeapi.ErrParsingFailed
	// Compose v5 dropped these sentinels; keep them so errors.Is still works
	// for callers and for wrapped registry-login failures.
	ErrLoginFailed          = errors.New("login failed")
	ErrLoginRequired        = errors.New("login required")
	ErrWrongContextType     = errors.New("wrong context type")
	ErrInvalidProject       = composeerrdefs.ErrInvalid
	ErrUnsupportedAttribute = composeerrdefs.ErrUnsupported
	ErrIncompatible         = composeerrdefs.ErrIncompatible
	ErrDisabled             = composeerrdefs.ErrDisabled
)

func (e *Error) classify() {
	if e == nil || e.err == nil || e.class != nil {
		return
	}
	err := e.err
	if _, ok := errors.AsType[InvalidTemplateError](err); ok {
		e.class = ErrInvalidArgument
		e.hint = "compose interpolation template is invalid"
		e.status = http.StatusBadRequest
		return
	}
	var miss MissingRequiredError
	if errors.As(err, &miss) {
		e.class = ErrInvalidArgument
		e.hint = "required compose variable is unset"
		e.status = http.StatusBadRequest
		if miss.Variable != "" {
			e.Path = miss.Variable
		}
		return
	}
	var js *JSONError
	if errors.As(err, &js) && js != nil {
		e.status = js.Code
		e.hint = "engine json stream error"
		if native := errhttp.ToNative(js.Code); native != nil {
			e.class = native
		} else {
			e.class = ErrUnknown
		}
		return
	}
	var exit *ExitError
	if errors.As(err, &exit) && exit != nil {
		e.class = ErrInternal
		e.hint = "container wait returned an error payload"
		if e.Path == "" || e.Path == "." {
			e.Path = "wait"
		}
		return
	}
	if engineclient.IsErrConnectionFailed(err) {
		e.class = ErrUnavailable
		e.hint = "cannot reach the docker daemon"
		e.status = http.StatusServiceUnavailable
		return
	}
	switch {
	case errors.Is(err, errRequired), errors.Is(err, errNilOption),
		errors.Is(err, errInvalid), errors.Is(err, errEmpty),
		errors.Is(err, errSecretRedacted):
		e.class = ErrInvalidArgument
		e.status = http.StatusBadRequest
	case errors.Is(err, errUnsupported):
		e.class = ErrUnsupportedAttribute
		e.hint = "value cannot be mapped across engine and compose"
	case errors.Is(err, ErrRedirect):
		e.class = ErrInvalidArgument
		e.hint = "unexpected http redirect from the daemon"
		e.status = http.StatusBadRequest
	case cerrdefs.IsNotFound(err), composeapi.IsNotFoundError(err), composeerrdefs.IsNotFoundError(err):
		e.class = ErrNotFound
		e.hint = "resource does not exist"
	case cerrdefs.IsAlreadyExists(err), composeapi.IsAlreadyExistsError(err):
		e.class = ErrAlreadyExists
		e.hint = "resource already exists"
	case cerrdefs.IsConflict(err):
		e.class = ErrConflict
		e.hint = "resource is in a conflicting state"
	case cerrdefs.IsUnauthorized(err):
		e.class = ErrUnauthorized
		e.hint = "registry or daemon rejected credentials"
	case cerrdefs.IsPermissionDenied(err), composeapi.IsForbiddenError(err):
		e.class = ErrPermissionDenied
		e.hint = "operation is forbidden"
	case cerrdefs.IsInvalidArgument(err), composeerrdefs.IsInvalidError(err):
		e.class = ErrInvalidArgument
		e.hint = "request or compose file is invalid"
	case cerrdefs.IsNotImplemented(err), composeapi.IsErrNotImplemented(err):
		e.class = ErrNotImplemented
		e.hint = "daemon or backend does not implement this action"
	case cerrdefs.IsUnavailable(err):
		e.class = ErrUnavailable
		e.hint = "daemon or registry is unavailable"
	case cerrdefs.IsResourceExhausted(err):
		e.class = ErrResourceExhausted
		e.hint = "daemon resource limits were exceeded"
	case cerrdefs.IsFailedPrecondition(err):
		e.class = ErrFailedPrecondition
		e.hint = "a required precondition is not met"
	case cerrdefs.IsNotModified(err):
		e.class = ErrNotModified
	case cerrdefs.IsAborted(err):
		e.class = ErrAborted
	case cerrdefs.IsOutOfRange(err):
		e.class = ErrOutOfRange
	case cerrdefs.IsInternal(err):
		e.class = ErrInternal
		e.hint = "daemon internal error"
	case cerrdefs.IsDataLoss(err):
		e.class = ErrDataLoss
	case cerrdefs.IsUnknown(err), composeapi.IsUnknownError(err):
		e.class = ErrUnknown
	case cerrdefs.IsCanceled(err), composeapi.IsErrCanceled(err), errors.Is(err, context.Canceled):
		e.class = ErrCanceled
		e.hint = "operation was canceled"
	case errors.Is(err, ErrLoginFailed):
		e.class = ErrLoginFailed
		e.hint = "registry login failed"
	case errors.Is(err, ErrLoginRequired):
		e.class = ErrLoginRequired
		e.hint = "registry login is required"
	case composeapi.IsErrUnsupportedFlag(err):
		e.class = ErrUnsupportedFlag
	case composeapi.IsErrParsingFailed(err):
		e.class = ErrParsingFailed
		e.hint = "compose parsing failed"
	case errors.Is(err, ErrWrongContextType):
		e.class = ErrWrongContextType
	case composeerrdefs.IsUnsupportedError(err):
		e.class = ErrUnsupportedAttribute
		e.hint = "compose file uses an unsupported attribute"
	case composeerrdefs.IsIncompatibleError(err):
		e.class = ErrIncompatible
		e.hint = "compose file uses an incompatible attribute"
	case errors.Is(err, ErrDisabled):
		e.class = ErrDisabled
		e.hint = "compose resource is disabled"
	}
	if e.class != nil && e.status == 0 {
		e.status = errhttp.ToHTTP(e.class)
	}
}

// ClassOf returns the SDK error class for err, or nil if unclassified.
func ClassOf(err error) error {
	var e *Error
	if errors.As(err, &e) && e.class != nil {
		return e.class
	}
	probe := &Error{err: err}
	probe.classify()
	return probe.class
}

func IsUnknown(err error) bool { return errors.Is(err, ErrUnknown) || cerrdefs.IsUnknown(err) }
func IsInvalidArgument(err error) bool {
	return errors.Is(err, ErrInvalidArgument) || cerrdefs.IsInvalidArgument(err)
}
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || cerrdefs.IsNotFound(err) ||
		composeapi.IsNotFoundError(err) || composeerrdefs.IsNotFoundError(err)
}
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists) || cerrdefs.IsAlreadyExists(err) || composeapi.IsAlreadyExistsError(err)
}
func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied) || cerrdefs.IsPermissionDenied(err) || composeapi.IsForbiddenError(err)
}
func IsResourceExhausted(err error) bool {
	return errors.Is(err, ErrResourceExhausted) || cerrdefs.IsResourceExhausted(err)
}
func IsFailedPrecondition(err error) bool {
	return errors.Is(err, ErrFailedPrecondition) || cerrdefs.IsFailedPrecondition(err)
}
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) || cerrdefs.IsConflict(err) }
func IsNotModified(err error) bool {
	return errors.Is(err, ErrNotModified) || cerrdefs.IsNotModified(err)
}
func IsAborted(err error) bool    { return errors.Is(err, ErrAborted) || cerrdefs.IsAborted(err) }
func IsOutOfRange(err error) bool { return errors.Is(err, ErrOutOfRange) || cerrdefs.IsOutOfRange(err) }
func IsNotImplemented(err error) bool {
	return errors.Is(err, ErrNotImplemented) || cerrdefs.IsNotImplemented(err) || composeapi.IsErrNotImplemented(err)
}
func IsInternal(err error) bool { return errors.Is(err, ErrInternal) || cerrdefs.IsInternal(err) }
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable) || cerrdefs.IsUnavailable(err)
}
func IsDataLoss(err error) bool { return errors.Is(err, ErrDataLoss) || cerrdefs.IsDataLoss(err) }
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) || cerrdefs.IsUnauthorized(err)
}
func IsCanceled(err error) bool {
	return errors.Is(err, ErrCanceled) || cerrdefs.IsCanceled(err) || composeapi.IsErrCanceled(err)
}
func IsConnectionFailed(err error) bool { return engineclient.IsErrConnectionFailed(err) }
func IsRedirect(err error) bool         { return errors.Is(err, ErrRedirect) }
func IsLoginFailed(err error) bool      { return errors.Is(err, ErrLoginFailed) }
func IsLoginRequired(err error) bool    { return errors.Is(err, ErrLoginRequired) }
func IsUnsupportedFlag(err error) bool {
	return errors.Is(err, ErrUnsupportedFlag) || composeapi.IsErrUnsupportedFlag(err)
}
func IsParsingFailed(err error) bool {
	return errors.Is(err, ErrParsingFailed) || composeapi.IsErrParsingFailed(err)
}
func IsWrongContextType(err error) bool { return errors.Is(err, ErrWrongContextType) }
func IsInvalidProject(err error) bool {
	return errors.Is(err, ErrInvalidProject) || composeerrdefs.IsInvalidError(err)
}
func IsUnsupportedAttribute(err error) bool {
	return errors.Is(err, ErrUnsupportedAttribute) || composeerrdefs.IsUnsupportedError(err)
}
func IsIncompatible(err error) bool {
	return errors.Is(err, ErrIncompatible) || composeerrdefs.IsIncompatibleError(err)
}
func IsDisabled(err error) bool { return errors.Is(err, ErrDisabled) }
