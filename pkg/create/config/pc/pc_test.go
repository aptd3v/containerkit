package pc_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/pc"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *ocispec.Platform
		setFn    create.SetPlatformConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &ocispec.Platform{},
			setFn:    pc.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:   &ocispec.Platform{},
			setFn:    pc.WithArchitecture("amd64"),
			field:    "Architecture",
			wantErr:  false,
			message:  "WithArchitecture ok",
			expected: "amd64",
		},
		{
			config:   &ocispec.Platform{},
			setFn:    pc.WithOS("linux"),
			field:    "OS",
			wantErr:  false,
			message:  "WithOS ok",
			expected: "linux",
		},
		{
			config:   &ocispec.Platform{},
			setFn:    pc.WithOSVersion("1.0.0"),
			field:    "OSVersion",
			wantErr:  false,
			message:  "WithOSVersion ok",
			expected: "1.0.0",
		},
		{
			config:   &ocispec.Platform{},
			setFn:    pc.WithOSFeatures("foo", "bar"),
			field:    "OSFeatures",
			wantErr:  false,
			message:  "WithOSFeatures ok",
			expected: []string{"foo", "bar"},
		},
		{
			config:   &ocispec.Platform{},
			setFn:    pc.WithVariant("v1"),
			field:    "Variant",
			wantErr:  false,
			message:  "WithVariant ok",
			expected: "v1",
		},
		{
			config:   &ocispec.Platform{},
			setFn:    pc.Fail(errors.New("test error")),
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
			assert.True(t, errdefs.IsPlatformConfigError(err), "expected container config error")
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}
