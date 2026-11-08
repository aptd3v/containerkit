package kit

import (
	"reflect"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

var (
	nilMapString      map[string]string
	nilMapStruct      map[string]struct{}
	nilStringSlice    []string
	nilDockerStrSlice strslice.StrSlice
	nilStopTimeout    *int
)

func TestContainerConfigSetters(t *testing.T) {
	stopTimeout := 10
	tests := []struct {
		config   *container.Config
		setFn    func(*Container)
		field    string
		expected any
		wantErr  bool
		message  string
	}{
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Env("TEST", "test")
			},
			field:    "Env",
			expected: []string{"TEST=test"},
			wantErr:  false,
			message:  "Env ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Env("", "")
			},
			field:    "Env",
			expected: make([]string, 0),
			wantErr:  true,
			message:  "Env error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.EnvMap(map[string]string{"TEST": "test"})
			},
			field:    "Env",
			expected: []string{"TEST=test"},
			wantErr:  false,
			message:  "EnvMap ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.EnvMap(map[string]string{"": ""})
			},
			field:    "Env",
			expected: make([]string, 0),
			wantErr:  true,
			message:  "EnvMap error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.ExposedPort("tcp", "80")
			},
			field:    "ExposedPorts",
			expected: nat.PortSet{nat.Port("80/tcp"): {}},
			wantErr:  false,
			message:  "ExposedPort ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.ExposedPort("tcp", "300000")
			},
			field:    "ExposedPorts",
			expected: nat.PortSet{},
			wantErr:  true,
			message:  "ExposedPort error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.HostName("test")
			},
			field:    "Hostname",
			expected: "test",
			wantErr:  false,
			message:  "HostName ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.HostName("")
			},
			field:    "Hostname",
			expected: "",
			wantErr:  true,
			message:  "HostName error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.DomainName("test")
			},
			field:    "Domainname",
			expected: "test",
			wantErr:  false,
			message:  "DomainName ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.DomainName("")
			},
			field:    "Domainname",
			expected: "",
			wantErr:  true,
			message:  "DomainName error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Image("test")
			},
			field:    "Image",
			expected: "test",
			wantErr:  false,
			message:  "Image ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Image("")
			},
			field:    "Image",
			expected: "",
			wantErr:  true,
			message:  "Image error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Imagef("test-%s", "test")
			},
			field:    "Image",
			expected: "test-test",
			wantErr:  false,
			message:  "Imagef ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Imagef("%s", "")
			},
			field:    "Image",
			expected: "",
			wantErr:  true,
			message:  "Imagef error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Command("test", "test2", "test3")
			},
			field:    "Cmd",
			expected: strslice.StrSlice([]string{"test", "test2", "test3"}),
			wantErr:  false,
			message:  "Command ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Command()
			},
			field:    "Cmd",
			expected: strslice.StrSlice(strslice.StrSlice(nil)),
			wantErr:  true,
			message:  "Command error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.User("test")
			},
			field:    "User",
			expected: "test",
			wantErr:  false,
			message:  "User ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.User("")
			},
			field:    "User",
			expected: "",
			wantErr:  true,
			message:  "User error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.AttachedStdin()
			},
			field:    "AttachStdin",
			expected: true,
			wantErr:  false,
			message:  "AttachedStdin ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.AttachedStdout()
			},
			field:    "AttachStdout",
			expected: true,
			wantErr:  false,
			message:  "AttachedStdout ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.AttachedStderr()
			},
			field:    "AttachStderr",
			expected: true,
			wantErr:  false,
			message:  "AttachedStderr ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Tty()
			},
			field:    "Tty",
			expected: true,
			wantErr:  false,
			message:  "Tty ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StdinOpen()
			},
			field:    "OpenStdin",
			expected: true,
			wantErr:  false,
			message:  "StdinOpen ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StdinOpen()
			},
			field:    "OpenStdin",
			expected: true,
			wantErr:  false,
			message:  "StdinOpen ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StdinOnce()
			},
			field:    "StdinOnce",
			expected: true,
			wantErr:  false,
			message:  "StdinOnce ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.EscapedArgs()
			},
			field:    "ArgsEscaped",
			expected: true,
			wantErr:  false,
			message:  "EscapedArgs ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Volume("test")
			},
			field:    "Volumes",
			expected: map[string]struct{}{"test": {}},
			wantErr:  false,
			message:  "Volume ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Volume("")
			},
			field:    "Volumes",
			expected: nilMapStruct,
			wantErr:  true,
			message:  "Volume error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.WorkingDir("test")
			},
			field:    "WorkingDir",
			expected: "test",
			wantErr:  false,
			message:  "WorkingDir ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.WorkingDir("")
			},
			field:    "WorkingDir",
			expected: "",
			wantErr:  true,
			message:  "WorkingDir error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.DisabledNetwork()
			},
			field:    "NetworkDisabled",
			expected: true,
			wantErr:  false,
			message:  "DisabledNetwork ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.OnBuild("test", "test2", "test3")
			},
			field:    "OnBuild",
			expected: []string{"test", "test2", "test3"},
			wantErr:  false,
			message:  "OnBuild ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.OnBuild()
			},
			field:    "OnBuild",
			expected: nilStringSlice,
			wantErr:  true,
			message:  "OnBuild error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Label("test", "test")
			},
			field:    "Labels",
			expected: map[string]string{"test": "test"},
			wantErr:  false,
			message:  "Label ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Label("", "")
			},
			field:    "Labels",
			expected: nilMapString,
			wantErr:  true,
			message:  "Label error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StopSignal("test")
			},
			field:    "StopSignal",
			expected: "test",
			wantErr:  false,
			message:  "StopSignal ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StopSignal("")
			},
			field:    "StopSignal",
			expected: "",
			wantErr:  true,
			message:  "StopSignal error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Entrypoint("test", "test2", "test3")
			},
			field:    "Entrypoint",
			expected: strslice.StrSlice{"test", "test2", "test3"},
			wantErr:  false,
			message:  "Entrypoint ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Entrypoint()
			},
			field:    "Entrypoint",
			expected: nilDockerStrSlice,
			wantErr:  true,
			message:  "Entrypoint error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Shell("test", "test2", "test3")
			},
			field:    "Shell",
			expected: strslice.StrSlice{"test", "test2", "test3"},
			wantErr:  false,
			message:  "Shell ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Shell()
			},
			field:    "Shell",
			expected: nilDockerStrSlice,
			wantErr:  true,
			message:  "Shell error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StopTimeout(10)
			},
			field:    "StopTimeout",
			expected: &stopTimeout,
			wantErr:  false,
			message:  "StopTimeout ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StopTimeout(0)
			},
			field:    "StopTimeout",
			expected: nilStopTimeout,
			wantErr:  true,
			message:  "StopTimeout error",
		},
	}

	for _, test := range tests {
		c := NewContainer("t")
		c.Config.Container = test.config
		test.setFn(c)
		if test.wantErr {
			assert.NotEmpty(t, c.Errors)
			assert.True(t, errdefs.IsContainerConfigError(c.Errors[0]), "expected container config error")
		} else {
			assert.Empty(t, c.Errors)
		}
		if test.field != "" {
			assert.Equal(t,
				test.expected,
				reflect.ValueOf(*test.config).FieldByName(test.field).Interface(),
				test.message,
			)
		}
	}
}
