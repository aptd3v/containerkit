package pool_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/network/pool"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.IPAMPool
		setFn    pool.SetIpamPoolProjectConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.IPAMPool{},
			setFn:    pool.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf ok",
			expected: nil,
		},
		{
			config:   &types.IPAMPool{},
			setFn:    pool.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &types.IPAMPool{},
			setFn:    pool.WithAuxiliaryAddresses("test", "test"),
			field:    "AuxiliaryAddresses",
			wantErr:  false,
			message:  "WithAuxiliaryAddresses ok",
			expected: types.Mapping{"test": "test"},
		},
		{
			config:   &types.IPAMPool{},
			setFn:    pool.WithSubnet("10.0.0.0/24"),
			field:    "Subnet",
			wantErr:  false,
			message:  "WithSubnet ok",
			expected: "10.0.0.0/24",
		},
		{
			config:   &types.IPAMPool{},
			setFn:    pool.WithGateway("10.0.0.1"),
			field:    "Gateway",
			wantErr:  false,
			message:  "WithGateway ok",
			expected: "10.0.0.1",
		},
		{
			config:   &types.IPAMPool{},
			setFn:    pool.WithIpRange("10.0.0.0/24"),
			field:    "IPRange",
			wantErr:  false,
			message:  "WithIpRange ok",
			expected: "10.0.0.0/24",
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
