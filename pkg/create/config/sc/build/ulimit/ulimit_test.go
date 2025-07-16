package ulimit_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/build/ulimit"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.UlimitsConfig
		setFn    ulimit.SetUlimitConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.UlimitsConfig{},
			setFn:    ulimit.WithHard(1000),
			field:    "Hard",
			wantErr:  false,
			message:  "WithHard ok",
			expected: 1000,
		},
		{
			config:   &types.UlimitsConfig{},
			setFn:    ulimit.WithSingle(1000),
			field:    "Single",
			wantErr:  false,
			message:  "WithSingle ok",
			expected: 1000,
		},
		{
			config:   &types.UlimitsConfig{},
			setFn:    ulimit.WithSoft(1000),
			field:    "Soft",
			wantErr:  false,
			message:  "WithSoft ok",
			expected: 1000,
		},
		{
			config:   &types.UlimitsConfig{},
			setFn:    ulimit.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf error",
			expected: nil,
		},
		{
			config:   &types.UlimitsConfig{},
			setFn:    ulimit.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail error",
			expected: nil,
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
