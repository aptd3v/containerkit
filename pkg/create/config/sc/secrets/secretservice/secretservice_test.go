package secretservice_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/secrets/secretservice"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.ServiceSecretConfig
		setFn    secretservice.SetSecretServiceConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.ServiceSecretConfig{},
			setFn:    secretservice.WithMode(0664),
			field:    "Mode",
			wantErr:  false,
			message:  "WithMode ok",
			expected: "0664",
		},
		{
			config:   &types.ServiceSecretConfig{},
			setFn:    secretservice.WithGID("test"),
			field:    "GID",
			wantErr:  false,
			message:  "WithGID ok",
			expected: "test",
		},
		{
			config:   &types.ServiceSecretConfig{},
			setFn:    secretservice.WithUID("test"),
			field:    "UID",
			wantErr:  false,
			message:  "WithUID ok",
			expected: "test",
		},
		{
			config:   &types.ServiceSecretConfig{},
			setFn:    secretservice.WithTarget("test"),
			field:    "Target",
			wantErr:  false,
			message:  "WithTarget ok",
			expected: "test",
		},
		{
			config:   &types.ServiceSecretConfig{},
			setFn:    secretservice.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf ok",
			expected: nil,
		},
		{
			config:   &types.ServiceSecretConfig{},
			setFn:    secretservice.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &types.ServiceSecretConfig{},
			setFn:    secretservice.WithSource("test"),
			field:    "Source",
			wantErr:  false,
			message:  "WithSource ok",
			expected: "test",
		},
	}

	for _, test := range tests {
		err := test.setFn(test.config)
		if test.wantErr {
			assert.Error(t, err)
			assert.True(t, errdefs.IsServiceConfigError(err), "expected service config error")
		} else {
			assert.NoError(t, err)

			if test.field == "Mode" {
				// testing mode is difficult because it is a pointer to a types.FileMode
				// and internally it is a pointed to a variable within the setters closure
				// so this is a way to confirm its set correctly
				assert.Equal(t, test.expected, test.config.Mode.String(), test.message)
				continue
			}
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}
