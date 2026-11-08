package kit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewContainerAndValidate(t *testing.T) {
	t.Parallel()
	c := NewContainer()
	require.Equal(t, "", c.Name)
	require.NoError(t, c.Validate())
}

func TestContainer_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		c       *Container
		wantErr string
	}{
		{name: "nil receiver", c: nil, wantErr: "container is nil"},
		{name: "empty errors ok", c: NewContainer(), wantErr: ""},
		{name: "nil errors slice ok", c: &Container{Errors: nil}, wantErr: ""},
		{name: "joined errors", c: &Container{Errors: []error{errors.New("x")}}, wantErr: "container config has errors"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.c.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}
