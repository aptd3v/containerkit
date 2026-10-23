package containerkit

import (
	"reflect"
	"testing"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

var boolFalse = false

func TestServiceConfigSetters(t *testing.T) {
	tests := []struct {
		config   *types.ServiceConfig
		setFn    func(*types.ServiceConfig) error
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.ServiceConfig{},
			setFn:    Profiles("foo", "bar"),
			field:    "Profiles",
			wantErr:  false,
			message:  "WithProfiles ok",
			expected: []string{"foo", "bar"},
		},
		{
			config:  &types.ServiceConfig{},
			setFn:   EnvFile("foo"),
			field:   "EnvFiles",
			wantErr: false,
			message: "WithEnvFile ok",
			expected: []types.EnvFile{{
				Path:     "foo",
				Required: true,
			}},
		},
		{
			config:  &types.ServiceConfig{},
			setFn:   DependsOnHealthy("foo"),
			field:   "DependsOn",
			wantErr: false,
			message: "WithDependsOnHealthy ok",
			expected: types.DependsOnConfig{"foo": types.ServiceDependency{
				Condition: "service_healthy",
				Restart:   true,
				Required:  true,
			}},
		},
		{
			config:  &types.ServiceConfig{},
			setFn:   DependsOn("foo"),
			field:   "DependsOn",
			wantErr: false,
			message: "WithDependsOn ok",
			expected: types.DependsOnConfig{"foo": types.ServiceDependency{
				Condition: "service_started",
				Restart:   true,
				Required:  true,
			}},
		},
		{
			config:  &types.ServiceConfig{},
			setFn:   Develop(WatchActionSyncRestart, "foo", "bar"),
			field:   "Develop",
			wantErr: false,
			message: "WithDevelop ok",
			expected: &types.DevelopConfig{
				Watch: []types.Trigger{
					{
						Path:   "foo",
						Action: types.WatchActionSyncRestart,
						Target: "bar",
					},
				},
			},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    Annotation("foo", "bar"),
			field:    "Annotations",
			wantErr:  false,
			message:  "WithAnnotation ok",
			expected: types.Mapping{"foo": "bar"},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    NoAttach(),
			field:    "Attach",
			wantErr:  false,
			message:  "WithNoAttach ok",
			expected: &boolFalse,
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

func TestWithEnvFileWhenEnvironmentIsSet(t *testing.T) {
	config := &types.ServiceConfig{}
	config.Environment = types.NewMappingWithEquals([]string{"A=1"})
	assert.NoError(t, EnvFile("bar")(config))
	assert.Equal(t, []types.EnvFile{{Path: "bar", Required: true}}, config.EnvFiles)
}
