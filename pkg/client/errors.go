package client

import (
	cerrdefs "github.com/containerd/errdefs"
)

// IsErrNotFound reports whether err is a Docker Engine "not found" error
// (container, image, network, volume, and similar). It unwraps wrapped errors.
func IsErrNotFound(err error) bool {
	return cerrdefs.IsNotFound(err)
}

// IsErrConflict reports whether err is a Docker Engine conflict (HTTP 409),
// such as a name already in use or a container that is still running.
func IsErrConflict(err error) bool {
	return cerrdefs.IsConflict(err)
}

// IsErrInvalidArgument reports whether err is a Docker Engine invalid-argument
// error (HTTP 400).
func IsErrInvalidArgument(err error) bool {
	return cerrdefs.IsInvalidArgument(err)
}

// IsErrUnauthorized reports whether err is a Docker Engine authentication
// error (HTTP 401), including registry login failures.
func IsErrUnauthorized(err error) bool {
	return cerrdefs.IsUnauthorized(err)
}

// IsErrPermissionDenied reports whether err is a Docker Engine forbidden
// error (HTTP 403).
func IsErrPermissionDenied(err error) bool {
	return cerrdefs.IsPermissionDenied(err)
}

// IsErrUnavailable reports whether err is a Docker Engine unavailable error
// (HTTP 503), such as a daemon that cannot be reached.
func IsErrUnavailable(err error) bool {
	return cerrdefs.IsUnavailable(err)
}

// IsErrNotImplemented reports whether err is a Docker Engine not-implemented
// error (HTTP 501).
func IsErrNotImplemented(err error) bool {
	return cerrdefs.IsNotImplemented(err)
}

// IsErrNotModified reports whether err is a Docker Engine not-modified error
// (HTTP 304).
func IsErrNotModified(err error) bool {
	return cerrdefs.IsNotModified(err)
}

// IsErrInternal reports whether err is a Docker Engine internal error (HTTP 500).
func IsErrInternal(err error) bool {
	return cerrdefs.IsInternal(err)
}

// IsErrUnknown reports whether err is an unclassified Docker Engine error.
func IsErrUnknown(err error) bool {
	return cerrdefs.IsUnknown(err)
}

// IsErrDataLoss reports whether err is a Docker Engine data-loss error.
func IsErrDataLoss(err error) bool {
	return cerrdefs.IsDataLoss(err)
}

// IsErrCanceled reports whether err is from a canceled context. It unwraps
// wrapped errors.
func IsErrCanceled(err error) bool {
	return cerrdefs.IsCanceled(err)
}

// IsErrDeadlineExceeded reports whether err is from a context deadline. It
// unwraps wrapped errors.
func IsErrDeadlineExceeded(err error) bool {
	return cerrdefs.IsDeadlineExceeded(err)
}
