package ipam_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint/ipam"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *network.EndpointIPAMConfig
		setFn    ipam.SetIPAMConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &network.EndpointIPAMConfig{},
			setFn:    ipam.WithIPv4Address("192.168.1.1/24"),
			field:    "IPv4Address",
			wantErr:  false,
			message:  "WithIPv4Address ok",
			expected: "192.168.1.1/24",
		},
		{
			config:   &network.EndpointIPAMConfig{},
			setFn:    ipam.WithIPv6Address("2001:db8::1/64"),
			field:    "IPv6Address",
			wantErr:  false,
			message:  "WithIPv6Address ok",
			expected: "2001:db8::1/64",
		},
		{
			config:   &network.EndpointIPAMConfig{},
			setFn:    ipam.WithLinkLocalIPs("fe80::1%eth0", "fe80::2%eth0"),
			field:    "LinkLocalIPs",
			wantErr:  false,
			message:  "WithLinkLocalIPs ok",
			expected: []string{"fe80::1%eth0", "fe80::2%eth0"},
		},
		{
			config:   &network.EndpointIPAMConfig{},
			setFn:    ipam.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &network.EndpointIPAMConfig{},
			setFn:    ipam.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf ok",
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
