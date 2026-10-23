package containerkit

import (
	"reflect"
	"testing"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
)

func TestPlatformConfigSetters(t *testing.T) {
	tests := []struct {
		config   *ocispec.Platform
		setFn    func(*Container)
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config: &ocispec.Platform{},
			setFn: func(c *Container) {
				c.Architecture("amd64")
			},
			field:    "Architecture",
			wantErr:  false,
			message:  "WithArchitecture ok",
			expected: "amd64",
		},
		{
			config: &ocispec.Platform{},
			setFn: func(c *Container) {
				c.OS("linux")
			},
			field:    "OS",
			wantErr:  false,
			message:  "WithOS ok",
			expected: "linux",
		},
		{
			config: &ocispec.Platform{},
			setFn: func(c *Container) {
				c.OSVersion("1.0.0")
			},
			field:    "OSVersion",
			wantErr:  false,
			message:  "WithOSVersion ok",
			expected: "1.0.0",
		},
		{
			config: &ocispec.Platform{},
			setFn: func(c *Container) {
				c.OSFeatures("foo", "bar")
			},
			field:    "OSFeatures",
			wantErr:  false,
			message:  "WithOSFeatures ok",
			expected: []string{"foo", "bar"},
		},
		{
			config: &ocispec.Platform{},
			setFn: func(c *Container) {
				c.Variant("v1")
			},
			field:    "Variant",
			wantErr:  false,
			message:  "WithVariant ok",
			expected: "v1",
		},
	}
	for _, test := range tests {
		c := NewContainer("t")
		c.Config.Platform = test.config
		test.setFn(c)
		if test.wantErr {
			assert.NotEmpty(t, c.Errors)
			assert.True(t, errdefs.IsPlatformConfigError(c.Errors[0]), "expected container config error")
		} else {
			assert.Empty(t, c.Errors)
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}
