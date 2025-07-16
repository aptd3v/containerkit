package update_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/update"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	uint100 = uint64(100)
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.UpdateConfig
		setFn    update.SetUpdateConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.UpdateConfig{},
			setFn:    update.Failf("test error %s", "foo"),
			field:    "UpdateConfig",
			wantErr:  true,
			message:  "Failf error format setter",
			expected: nil,
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.Fail(errors.New("test error")),
			field:    "UpdateConfig",
			wantErr:  true,
			message:  "Fail error setter",
			expected: nil,
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.WithOrder("start-first"),
			field:    "Order",
			wantErr:  false,
			message:  "WithOrder ok",
			expected: "start-first",
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.WithMaxFailureRatio(0.5),
			field:    "MaxFailureRatio",
			wantErr:  false,
			message:  "WithMaxFailureRatio ok",
			expected: float32(0.5),
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.WithMonitor(5),
			field:    "Monitor",
			wantErr:  false,
			message:  "WithMonitor ok",
			expected: types.Duration(time.Duration(5) * time.Second),
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.Failf("test error %s", "foo"),
			field:    "UpdateConfig",
			wantErr:  true,
			message:  "Failf error format setter",
			expected: nil,
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.Fail(errors.New("test error")),
			field:    "UpdateConfig",
			wantErr:  true,
			message:  "Fail error setter",
			expected: nil,
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.WithParallelism(100),
			field:    "Parallelism",
			wantErr:  false,
			message:  "WithParallelism ok",
			expected: &uint100,
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.WithDelay(5),
			field:    "Delay",
			wantErr:  false,
			message:  "WithDelay ok",
			expected: types.Duration(time.Duration(5) * time.Second),
		},
		{
			config:   &types.UpdateConfig{},
			setFn:    update.WithFailureAction("rollback"),
			field:    "FailureAction",
			wantErr:  false,
			message:  "WithFailureAction ok",
			expected: "rollback",
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
