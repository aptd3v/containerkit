package projectsecret_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/secrets/projectsecret"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.SecretConfig
		setFn    projectsecret.SetProjectSecretConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.WithTemplateDriver("test"),
			field:    "TemplateDriver",
			wantErr:  false,
			message:  "WithTemplateDriver ok",
			expected: "test",
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.WithDriverOptions("test", "test"),
			field:    "DriverOpts",
			wantErr:  false,
			message:  "WithDriverOptions ok",
			expected: map[string]string{"test": "test"},
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.WithDriver("test"),
			field:    "Driver",
			wantErr:  false,
			message:  "WithDriver ok",
			expected: "test",
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.WithExternal(),
			field:    "External",
			wantErr:  false,
			message:  "WithExternal ok",
			expected: types.External(true),
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.WithEnvironment("test"),
			field:    "Environment",
			wantErr:  false,
			message:  "WithEnvironment ok",
			expected: "test",
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.WithContent("test"),
			field:    "Content",
			wantErr:  false,
			message:  "WithContent ok",
			expected: "test",
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.WithName("test"),
			field:    "Name",
			wantErr:  false,
			message:  "WithName ok",
			expected: "test",
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf ok",
			expected: nil,
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &types.SecretConfig{},
			setFn:    projectsecret.WithFile("test"),
			field:    "File",
			wantErr:  false,
			message:  "WithFile ok",
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
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}
