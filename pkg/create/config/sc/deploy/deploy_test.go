package deploy_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/update"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	intValue  = 5
	uintValue = uint64(5)
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.DeployConfig
		setFn    deploy.SetDeployConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.Failf("test error %s", "foo"),
			field:    "DeployConfig",
			wantErr:  true,
			message:  "Failf error format setter",
			expected: types.DeployConfig{},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.Fail(errors.New("test error")),
			field:    "DeployConfig",
			wantErr:  true,
			message:  "Fail error setter",
			expected: types.DeployConfig{},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithResourceReservations(resource.Fail(errors.New("test error"))),
			field:    "Resources",
			wantErr:  true,
			message:  "WithResourceReservations error setter",
			expected: types.Resources{},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithResourceReservations(resource.Failf("test error %s", "foo")),
			field:    "Resources",
			wantErr:  true,
			message:  "WithResourceReservations error format setter",
			expected: types.Resources{},
		},
		{
			config:  &types.DeployConfig{},
			setFn:   deploy.WithResourceReservations(nil, nil, nil),
			field:   "Resources",
			wantErr: false,
			message: "WithResourceReservations nil setters",
			expected: types.Resources{
				Reservations: &types.Resource{},
			},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithResourceReservations(),
			field:    "Resources",
			wantErr:  false,
			message:  "WithResourceReservations no setters",
			expected: types.Resources{},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithResourceLimits(resource.Failf("test error %s", "foo")),
			field:    "Resources",
			wantErr:  true,
			message:  "WithResourceLimits error format setter",
			expected: types.Resources{},
		},
		{
			config:  &types.DeployConfig{},
			setFn:   deploy.WithResourceLimits(resource.Fail(errors.New("test error"))),
			field:   "Resources",
			wantErr: true,
			message: "WithResourceLimits error setter",
			expected: types.Resources{
				Limits: &types.Resource{},
			},
		},
		{
			config:  &types.DeployConfig{},
			setFn:   deploy.WithResourceLimits(nil, nil, nil),
			field:   "Resources",
			wantErr: false,
			message: "WithResourceLimits nil setters",
			expected: types.Resources{
				Limits: &types.Resource{},
			},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithResourceLimits(),
			field:    "Resources",
			wantErr:  false,
			message:  "WithResourceLimits no setters",
			expected: types.Resources{},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithUpdateConfig(update.Fail(errors.New("test error"))),
			field:    "UpdateConfig",
			wantErr:  true,
			message:  "WithUpdateConfig error setter",
			expected: &types.UpdateConfig{},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithUpdateConfig(nil),
			field:    "UpdateConfig",
			wantErr:  false,
			message:  "WithUpdateConfig nil setters",
			expected: &types.UpdateConfig{},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithUpdateConfig(),
			field:    "UpdateConfig",
			wantErr:  false,
			message:  "WithUpdateConfig no setters",
			expected: (*types.UpdateConfig)(nil),
		},
		{
			config:  &types.DeployConfig{},
			setFn:   deploy.WithUpdateConfig(update.WithParallelism(5)),
			field:   "UpdateConfig",
			wantErr: false,
			message: "WithUpdateConfig ok",
			expected: &types.UpdateConfig{
				Parallelism: &uintValue,
			},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithLabel("foo", "bar"),
			field:    "Labels",
			wantErr:  false,
			message:  "WithLabel ok",
			expected: types.Labels{"foo": "bar"},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithReplicas(5),
			field:    "Replicas",
			wantErr:  false,
			message:  "WithReplicas ok",
			expected: &intValue,
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithMode("foo"),
			field:    "Mode",
			wantErr:  false,
			message:  "WithMode ok",
			expected: "foo",
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithRollbackConfig(update.Fail(errors.New("test error"))),
			field:    "RollbackConfig",
			wantErr:  true,
			message:  "WithRollbackConfig error setter",
			expected: &types.UpdateConfig{},
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithRollbackConfig(),
			field:    "RollbackConfig",
			wantErr:  false,
			message:  "WithRollbackConfig no setters",
			expected: (*types.UpdateConfig)(nil),
		},
		{
			config:   &types.DeployConfig{},
			setFn:    deploy.WithRollbackConfig(nil, nil, nil),
			field:    "RollbackConfig",
			wantErr:  false,
			message:  "WithRollbackConfig nil setters",
			expected: &types.UpdateConfig{},
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
