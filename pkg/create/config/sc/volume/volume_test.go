package volume_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/volume"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.VolumeConfig
		setFn    volume.SetVolumeProjectConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.VolumeConfig{},
			setFn:    volume.Failf("test %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &types.VolumeConfig{},
			setFn:    volume.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &types.VolumeConfig{},
			setFn:    volume.WithLabel("test", "test"),
			field:    "Labels",
			wantErr:  false,
			message:  "WithLabel ok",
			expected: types.Labels{"test": "test"},
		},
		{
			config:   &types.VolumeConfig{},
			setFn:    volume.WithDriver("test"),
			field:    "Driver",
			wantErr:  false,
			message:  "WithDriver ok",
			expected: "test",
		},
		{
			config:   &types.VolumeConfig{},
			setFn:    volume.WithDriverOptions("test", "test"),
			field:    "DriverOpts",
			wantErr:  false,
			message:  "WithDriverOptions ok",
			expected: types.Options{"test": "test"},
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
