package nc_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/nc"
	"github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint"
	"github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint/ipam"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *network.NetworkingConfig
		setFn    create.SetNetworkConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &network.NetworkingConfig{},
			setFn:    nc.Failf("test error %s", "foo"),
			field:    "EndpointsConfig",
			wantErr:  true,
			message:  "Failf ok",
			expected: nil,
		},
		{
			config:   &network.NetworkingConfig{},
			setFn:    nc.Fail(errors.New("test error")),
			field:    "EndpointsConfig",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config: &network.NetworkingConfig{},
			setFn: nc.WithEndpoint("test",
				endpoint.WithIPAMConfig(
					ipam.WithIPv4Address("192.168.1.1/24"),
				),
			),
			field:   "EndpointsConfig",
			wantErr: false,
			message: "WithEndpoint ok",
			expected: map[string]*network.EndpointSettings{
				"test": {
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: "192.168.1.1/24",
					},
				},
			},
		},
		{
			config: &network.NetworkingConfig{},
			setFn: nc.WithEndpoint("test",
				endpoint.WithIPAMConfig(
					ipam.WithIPv4Address("192.168.1.1/24"),
				),
				endpoint.Fail(errors.New("test error")),
			),
			field:   "EndpointsConfig",
			wantErr: true,
			message: "WithEndpoint error",
			expected: map[string]*network.EndpointSettings{
				"test": {
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: "192.168.1.1/24",
					},
				},
			},
		},
	}
	for _, test := range tests {
		err := test.setFn(test.config)
		if test.wantErr {
			assert.Error(t, err)
			assert.True(t, errdefs.IsNetworkConfigError(err), "expected network config error")
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}
