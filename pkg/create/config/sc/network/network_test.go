package network_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/network"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/network/pool"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	boolTrue = true
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.NetworkConfig
		setFn    network.SetNetworkProjectConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.NetworkConfig{},
			setFn:    network.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf ok",
			expected: nil,
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithLabel("test", "test"),
			field:    "Labels",
			wantErr:  false,
			message:  "WithLabel ok",
			expected: types.Labels{"test": "test"},
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithEnableIPv6(),
			field:    "EnableIPv6",
			wantErr:  false,
			message:  "WithEnableIPv6 ok",
			expected: &boolTrue,
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithAttachable(),
			field:    "Attachable",
			wantErr:  false,
			message:  "WithAttachable ok",
			expected: true,
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithInternal(),
			field:    "Internal",
			wantErr:  false,
			message:  "WithInternal ok",
			expected: true,
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithIpamPool(pool.Fail(errors.New("test error"))),
			field:    "Ipam",
			wantErr:  true,
			message:  "WithIpamPool error setter",
			expected: nil,
		},
		{
			config:  &types.NetworkConfig{},
			setFn:   network.WithIpamPool(nil, nil, nil),
			field:   "Ipam",
			wantErr: false,
			message: "WithIpamPool nil setters",
			expected: types.IPAMConfig{
				Config: []*types.IPAMPool{{}},
			},
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithIpamPool(),
			field:    "Ipam",
			wantErr:  false,
			message:  "WithIpamPool empty setters",
			expected: types.IPAMConfig{},
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithIpamPool(pool.WithGateway("192.168.1.1")),
			field:    "Ipam",
			wantErr:  false,
			message:  "WithIpamPool ok",
			expected: types.IPAMConfig{Config: []*types.IPAMPool{{Gateway: "192.168.1.1"}}},
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithIpamDriver("test"),
			field:    "Ipam",
			wantErr:  false,
			message:  "WithIpamDriver ok",
			expected: types.IPAMConfig{Driver: "test"},
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithDriverOptions("test", "test"),
			field:    "DriverOpts",
			wantErr:  false,
			message:  "WithDriverOptions ok",
			expected: types.Options{"test": "test"},
		},
		{
			config:   &types.NetworkConfig{},
			setFn:    network.WithDriver("test"),
			field:    "Driver",
			wantErr:  false,
			message:  "WithDriver ok",
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
