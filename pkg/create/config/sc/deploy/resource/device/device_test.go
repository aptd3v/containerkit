package device_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource/device"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.DeviceRequest
		setFn    device.SetDeviceConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.DeviceRequest{},
			setFn:    device.WithDriver("test"),
			field:    "Driver",
			wantErr:  false,
			message:  "WithDriver ok",
			expected: "test",
		},
		{
			config:   &types.DeviceRequest{},
			setFn:    device.WithIDs("test"),
			field:    "IDs",
			wantErr:  false,
			message:  "WithIDs ok",
			expected: []string{"test"},
		},
		{
			config:   &types.DeviceRequest{},
			setFn:    device.WithCount(1000),
			field:    "Count",
			wantErr:  false,
			message:  "WithCount ok",
			expected: types.DeviceCount(1000),
		},
		{
			config:   &types.DeviceRequest{},
			setFn:    device.WithCapabilities("test"),
			field:    "Capabilities",
			wantErr:  false,
			message:  "WithCapabilities ok",
			expected: []string{"test"},
		},
		{
			config:   &types.DeviceRequest{},
			setFn:    device.Fail(errors.New("test error")),
			field:    "DeviceRequest",
			wantErr:  true,
			message:  "Fail error setter",
			expected: nil,
		},
		{
			config:   &types.DeviceRequest{},
			setFn:    device.Failf("test error %s", "foo"),
			field:    "DeviceRequest",
			wantErr:  true,
			message:  "Failf error format setter",
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
