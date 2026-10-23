package containerkit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewContainerAndValidate(t *testing.T) {
	c := NewContainer()
	require.Equal(t, "", c.Name)
	require.NoError(t, c.Validate())

	c.Errors = nil
	require.NoError(t, c.Validate())

	c.Errors = []error{errors.New("x")}
	require.Error(t, c.Validate())
}
