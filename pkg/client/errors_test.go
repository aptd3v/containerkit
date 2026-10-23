package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/stretchr/testify/require"
)

func TestEngineErrorClasses(t *testing.T) {
	plain := errors.New("no such container: missing")
	cases := []struct {
		name string
		fn   func(error) bool
		err  error
	}{
		{name: "not found", fn: IsErrNotFound, err: cerrdefs.ErrNotFound},
		{name: "conflict", fn: IsErrConflict, err: cerrdefs.ErrConflict},
		{name: "invalid argument", fn: IsErrInvalidArgument, err: cerrdefs.ErrInvalidArgument},
		{name: "unauthorized", fn: IsErrUnauthorized, err: cerrdefs.ErrUnauthenticated},
		{name: "permission denied", fn: IsErrPermissionDenied, err: cerrdefs.ErrPermissionDenied},
		{name: "unavailable", fn: IsErrUnavailable, err: cerrdefs.ErrUnavailable},
		{name: "not implemented", fn: IsErrNotImplemented, err: cerrdefs.ErrNotImplemented},
		{name: "not modified", fn: IsErrNotModified, err: cerrdefs.ErrNotModified},
		{name: "internal", fn: IsErrInternal, err: cerrdefs.ErrInternal},
		{name: "unknown", fn: IsErrUnknown, err: cerrdefs.ErrUnknown},
		{name: "data loss", fn: IsErrDataLoss, err: cerrdefs.ErrDataLoss},
		{name: "canceled", fn: IsErrCanceled, err: context.Canceled},
		{name: "deadline exceeded", fn: IsErrDeadlineExceeded, err: context.DeadlineExceeded},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.False(t, tc.fn(nil))
			require.False(t, tc.fn(plain))
			require.True(t, tc.fn(tc.err))
			require.True(t, tc.fn(fmt.Errorf("inspect: %w", tc.err)))
		})
	}
}
