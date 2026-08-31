package client

import (
	"errors"
	"fmt"
	"testing"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/stretchr/testify/require"
)

func TestIsErrNotFound(t *testing.T) {
	require.False(t, IsErrNotFound(nil))
	require.False(t, IsErrNotFound(errors.New("no such container: missing")))
	require.True(t, IsErrNotFound(cerrdefs.ErrNotFound))
	require.True(t, IsErrNotFound(fmt.Errorf("inspect: %w", cerrdefs.ErrNotFound)))
}
