package resource_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource/device"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.Resource
		setFn    resource.SetResourceConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.Resource{},
			setFn:    resource.Fail(errors.New("test error")),
			field:    "Resource",
			wantErr:  true,
			message:  "Fail error setter",
			expected: nil,
		},
		{
			config:   &types.Resource{},
			setFn:    resource.Failf("test error %s", "foo"),
			field:    "Resource",
			wantErr:  true,
			message:  "Failf error format setter",
			expected: nil,
		},
		{
			config:  &types.Resource{},
			setFn:   resource.WithGenericResource("test", 1000),
			field:   "GenericResources",
			wantErr: false,
			message: "WithGenericResource ok",
			expected: []types.GenericResource{
				{
					DiscreteResourceSpec: &types.DiscreteGenericResource{
						Kind:  "test",
						Value: 1000,
					},
				},
			},
		},
		{
			config:   &types.Resource{},
			setFn:    resource.WithNanoCPUs(1000),
			field:    "NanoCPUs",
			wantErr:  false,
			message:  "WithNanoCPUs ok",
			expected: types.NanoCPUs(1000),
		},
		{
			config:   &types.Resource{},
			setFn:    resource.WithMemoryBytes(1000),
			field:    "MemoryBytes",
			wantErr:  false,
			message:  "WithMemoryBytes ok",
			expected: types.UnitBytes(1000),
		},
		{
			config:   &types.Resource{},
			setFn:    resource.WithPids(1000),
			field:    "Pids",
			wantErr:  false,
			message:  "WithPids ok",
			expected: int64(1000),
		},
		{
			config:   &types.Resource{},
			setFn:    resource.WithDevice(),
			field:    "Devices",
			wantErr:  false,
			message:  "WithDevice no setters",
			expected: [](types.DeviceRequest)(nil),
		},
		{
			config:   &types.Resource{},
			setFn:    resource.WithDevice(nil, nil, nil),
			field:    "Devices",
			wantErr:  false,
			message:  "WithDevice nil setters",
			expected: []types.DeviceRequest{{}},
		},
		{
			config:   &types.Resource{},
			setFn:    resource.WithDevice(device.WithCount(1000)),
			field:    "Devices",
			wantErr:  false,
			message:  "WithDevice ok",
			expected: []types.DeviceRequest{{Count: 1000}},
		},
		{
			config:   &types.Resource{},
			setFn:    resource.WithDevice(device.Fail(errors.New("test error"))),
			field:    "Devices",
			wantErr:  true,
			message:  "WithDevice device error setter",
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
