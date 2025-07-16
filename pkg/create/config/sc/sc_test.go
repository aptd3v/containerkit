package sc_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/sc"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/build"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/secrets/secretservice"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	boolFalse = false
	int10     = 10
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.ServiceConfig
		setFn    create.SetServiceConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf error setters",
			expected: (*types.ServiceConfig)(nil),
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail error setters",
			expected: (*types.ServiceConfig)(nil),
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithSecret(secretservice.Fail(errors.New("test error"))),
			field:    "Secrets",
			wantErr:  true,
			message:  "WithSecret error setters",
			expected: ([]types.ServiceSecretConfig)(nil),
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithSecret(nil, nil, nil),
			field:    "Secrets",
			wantErr:  false,
			message:  "WithSecret nil setters",
			expected: []types.ServiceSecretConfig{{}},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithSecret(),
			field:    "Secrets",
			wantErr:  false,
			message:  "WithSecret no setters",
			expected: ([]types.ServiceSecretConfig)(nil),
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithBuild(),
			field:    "Build",
			wantErr:  false,
			message:  "WithBuild no setters",
			expected: (*types.BuildConfig)(nil),
		},
		{
			config:  &types.ServiceConfig{},
			setFn:   sc.WithBuild(build.WithDockerfile("foo")),
			field:   "Build",
			wantErr: false,
			message: "WithBuild ok",
			expected: &types.BuildConfig{
				Dockerfile: "foo",
			},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithBuild(nil, nil),
			field:    "Build",
			wantErr:  false,
			message:  "WithBuild nil setters",
			expected: &types.BuildConfig{},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithBuild(build.Fail(errors.New("test error"))),
			field:    "Build",
			wantErr:  true,
			message:  "WithBuild error setters",
			expected: &types.BuildConfig{},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithProfiles("foo", "bar"),
			field:    "Profiles",
			wantErr:  false,
			message:  "WithProfiles ok",
			expected: []string{"foo", "bar"},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithDeploy(deploy.Fail(errors.New("test error"))),
			field:    "Deploy",
			wantErr:  true,
			message:  "WithDeploy error setters",
			expected: &types.DeployConfig{},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithDeploy(nil, nil, nil),
			field:    "Deploy",
			wantErr:  false,
			message:  "WithDeploy nil setters",
			expected: &types.DeployConfig{},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithDeploy(),
			field:    "Deploy",
			wantErr:  false,
			message:  "WithDeploy no setters",
			expected: (*types.DeployConfig)(nil),
		},
		{
			config:  &types.ServiceConfig{},
			setFn:   sc.WithDeploy(deploy.WithReplicas(10)),
			field:   "Deploy",
			wantErr: false,
			message: "WithDeploy ok",
			expected: &types.DeployConfig{
				Replicas: &int10,
			},
		},
		{
			config:  &types.ServiceConfig{},
			setFn:   sc.WithEnvFile("foo"),
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
			setFn:   sc.WithDependsOnHealthy("foo"),
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
			setFn:   sc.WithDependsOn("foo"),
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
			setFn:   sc.WithDevelop(sc.WatchActionSyncRestart, "foo", "bar"),
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
			setFn:    sc.WithAnnotation("foo", "bar"),
			field:    "Annotations",
			wantErr:  false,
			message:  "WithAnnotation ok",
			expected: types.Mapping{"foo": "bar"},
		},
		{
			config:   &types.ServiceConfig{},
			setFn:    sc.WithNoAttach(),
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
