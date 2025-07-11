package hc_test

import (
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
	"github.com/docker/docker/api/types/container"
	mountType "github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *container.HostConfig
		setFn    create.SetHostConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config: &container.HostConfig{},
			setFn: hc.WithMountPoint(
				mount.WithSource("/tmp/test"),
				mount.WithTarget("/tmp/test"),
				mount.WithType("bind"),
				mount.WithReadOnly(),
			),
			field:   "Mounts",
			wantErr: false,
			message: "WithMountPoint ok",
			expected: []mountType.Mount{
				{
					Source:   "/tmp/test",
					Target:   "/tmp/test",
					Type:     "bind",
					ReadOnly: true,
				},
			},
		},
		//TODO add more tests for WithMountPoint
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithMemoryLimit(1024),
			field:    "Memory",
			wantErr:  false,
			message:  "WithMemoryLimit ok",
			expected: int64(1024),
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithAutoRemove(),
			field:    "AutoRemove",
			wantErr:  false,
			message:  "WithAutoRemove ok",
			expected: true,
		},
		{
			config:  &container.HostConfig{},
			setFn:   hc.WithPortBindings("tcp", "0.0.0.0", "8080", "8080"),
			field:   "PortBindings",
			wantErr: false,
			message: "WithPortBindings ok",
			expected: nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
			},
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "0.0.0.0", "808012", ""),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings error",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("tcp", "0.0.0.0", "", ""),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings empty host and container port",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithPortBindings("", "", "", ""),
			field:    "PortBindings",
			wantErr:  true,
			message:  "WithPortBindings empty values",
			expected: nil,
		},
		{
			config:   &container.HostConfig{},
			setFn:    hc.WithDNSLookups("1.1.1.1", "8.8.8.8"),
			field:    "DNS",
			wantErr:  false,
			message:  "WithDNSLookups ok",
			expected: []string{"1.1.1.1", "8.8.8.8"},
		},
	}

	for _, test := range tests {
		err := test.setFn(test.config)
		if test.wantErr {
			assert.Error(t, err)
			assert.True(t, create.IsHostConfigError(err), "expected container config error")
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}
