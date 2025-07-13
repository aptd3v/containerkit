package endpoint_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint"
	"github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint/ipam"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *network.EndpointSettings
		setFn    endpoint.SetEndpointConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &network.EndpointSettings{},
			setFn:    endpoint.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf ok",
			expected: nil,
		},
		{
			config:   &network.EndpointSettings{},
			setFn:    endpoint.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &network.EndpointSettings{},
			setFn:    endpoint.WithDriverOptions("foo", "bar"),
			field:    "DriverOpts",
			wantErr:  false,
			message:  "WithDriverOptions ok",
			expected: map[string]string{"foo": "bar"},
		},
		{
			config:   &network.EndpointSettings{},
			setFn:    endpoint.WithMacAddress("02:42:ac:11:00:02"),
			field:    "MacAddress",
			wantErr:  false,
			message:  "WithMacAddress ok",
			expected: "02:42:ac:11:00:02",
		},
		{
			config:   &network.EndpointSettings{},
			setFn:    endpoint.WithAliases("test"),
			field:    "Aliases",
			wantErr:  false,
			message:  "WithAliases ok",
			expected: []string{"test"},
		},
		{
			config:   &network.EndpointSettings{},
			setFn:    endpoint.WithLinks("test"),
			field:    "Links",
			wantErr:  false,
			message:  "WithLinks ok",
			expected: []string{"test"},
		},
		{
			config:  &network.EndpointSettings{},
			setFn:   endpoint.WithIPAMConfig(ipam.WithIPv4Address("192.168.1.1/24")),
			field:   "IPAMConfig",
			wantErr: false,
			message: "WithIPAMConfig ok",
			expected: &network.EndpointIPAMConfig{
				IPv4Address: "192.168.1.1/24",
			},
		},
		{
			config:   &network.EndpointSettings{},
			setFn:    endpoint.WithIPAMConfig(ipam.Fail(errors.New("test error"))),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
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
