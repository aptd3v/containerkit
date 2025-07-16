package mount_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/aptd3v/go-contain/pkg/tools"
	dockerMount "github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *dockerMount.Mount
		setFn    mount.SetMountConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &dockerMount.Mount{},
			setFn:    mount.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf ok",
			expected: nil,
		},
		{
			config:   &dockerMount.Mount{},
			setFn:    mount.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail ok",
			expected: nil,
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithBindCreateMountpoint(),
			field:   "BindOptions",
			wantErr: false,
			message: "WithBindCreateMountpoint ok",
			expected: &dockerMount.BindOptions{
				CreateMountpoint: true,
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithBindReadOnlyForceRecursive(),
			field:   "BindOptions",
			wantErr: false,
			message: "WithBindReadOnlyForceRecursive ok",
			expected: &dockerMount.BindOptions{
				ReadOnlyForceRecursive: true,
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithBindReadOnlyNonRecursive(),
			field:   "BindOptions",
			wantErr: false,
			message: "WithBindReadOnlyNonRecursive ok",
			expected: &dockerMount.BindOptions{
				ReadOnlyNonRecursive: true,
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithBindNonRecursive(),
			field:   "BindOptions",
			wantErr: false,
			message: "WithBindNonRecursive ok",
			expected: &dockerMount.BindOptions{
				NonRecursive: true,
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithBindPropagation(mount.PropagationShared),
			field:   "BindOptions",
			wantErr: false,
			message: "WithBindPropagation ok",
			expected: &dockerMount.BindOptions{
				Propagation: "shared",
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithVolumeDriver("driver", "rw,noatime", "/dev/sda1"),
			field:   "VolumeOptions",
			wantErr: false,
			message: "WithVolumeDriver ok",
			expected: &dockerMount.VolumeOptions{
				DriverConfig: &dockerMount.Driver{
					Name: "driver",
					Options: map[string]string{
						"o":      "rw,noatime",
						"device": "/dev/sda1",
					},
				},
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithVolumeSubPath("/some/subpath"),
			field:   "VolumeOptions",
			wantErr: false,
			message: "WithVolumeSubPath ok",
			expected: &dockerMount.VolumeOptions{
				Subpath: "/some/subpath",
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithVolumeLabel("label", "value"),
			field:   "VolumeOptions",
			wantErr: false,
			message: "WithVolumeLabel ok",
			expected: &dockerMount.VolumeOptions{
				Labels: map[string]string{
					"label": "value",
				},
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithVolumeNoCopy(),
			field:   "VolumeOptions",
			wantErr: false,
			message: "WithVolumeNoCopy ok",
			expected: &dockerMount.VolumeOptions{
				NoCopy: true,
			},
		},
		{
			config: &dockerMount.Mount{},
			setFn: tools.Group(
				mount.WithTmpfsKeyValue("uid", "1000"),
				mount.WithTmpfsKeyValue("gid", "1000"),
			),
			field:   "TmpfsOptions",
			wantErr: false,
			message: "WithTmpfsKeyValue ok",
			expected: &dockerMount.TmpfsOptions{
				Options: [][]string{
					{"uid", "1000"},
					{"gid", "1000"},
				},
			},
		},
		{
			config: &dockerMount.Mount{},
			setFn: tools.Group(
				mount.WithTmpfsFlag("exec"),
				mount.WithTmpfsFlag("foo"),
				mount.WithTmpfsFlag("bar"),
			),
			field:   "TmpfsOptions",
			wantErr: false,
			message: "WithTmpfsFlag ok",
			expected: &dockerMount.TmpfsOptions{
				Options: [][]string{
					{"exec"},
					{"foo"},
					{"bar"},
				},
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithTmpfsMode(0755),
			field:   "TmpfsOptions",
			wantErr: false,
			message: "WithTmpfsMode ok",
			expected: &dockerMount.TmpfsOptions{
				Mode: 0755,
			},
		},
		{
			config:  &dockerMount.Mount{},
			setFn:   mount.WithTmpfsSizeBytes(1024),
			field:   "TmpfsOptions",
			wantErr: false,
			message: "WithTmpfsSizeBytes ok",
			expected: &dockerMount.TmpfsOptions{
				SizeBytes: 1024,
			},
		},
		{
			config:   &dockerMount.Mount{},
			setFn:    mount.WithConsistency(mount.ConsistencyDefault),
			field:    "Consistency",
			wantErr:  false,
			message:  "WithConsistency ok",
			expected: dockerMount.ConsistencyDefault,
		},
		{
			config:   &dockerMount.Mount{},
			setFn:    mount.WithReadWrite(),
			field:    "ReadOnly",
			wantErr:  false,
			message:  "WithReadWrite ok",
			expected: false,
		},
		{
			config:   &dockerMount.Mount{},
			setFn:    mount.WithReadOnly(),
			field:    "ReadOnly",
			wantErr:  false,
			message:  "WithReadOnly ok",
			expected: true,
		},
		{
			config:   &dockerMount.Mount{},
			setFn:    mount.WithTarget("/tmp/test"),
			field:    "Target",
			wantErr:  false,
			message:  "WithTarget ok",
			expected: "/tmp/test",
		},
		{
			config:   &dockerMount.Mount{},
			setFn:    mount.WithType(mount.MountTypeBind),
			field:    "Type",
			wantErr:  false,
			message:  "WithType ok",
			expected: dockerMount.TypeBind,
		},
		{
			config:   &dockerMount.Mount{},
			setFn:    mount.WithSource("/tmp/test"),
			field:    "Source",
			wantErr:  false,
			message:  "WithSource ok",
			expected: "/tmp/test",
		},
	}

	for _, test := range tests {
		err := test.setFn(test.config)
		if test.wantErr {
			assert.Error(t, err)
			assert.True(t, errdefs.IsHostConfigError(err), "expected container config error")
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, reflect.ValueOf(*test.config).FieldByName(test.field).Interface(), test.message)
		}
	}
}
