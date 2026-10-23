package containerkit

import (
	"reflect"
	"testing"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
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
			message:  "WithEnv ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Env("", "")
			},
			field:    "Env",
			expected: make([]string, 0),
			wantErr:  true,
			message:  "WithEnv error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.EnvMap(map[string]string{"TEST": "test"})
			},
			field:    "Env",
			expected: []string{"TEST=test"},
			wantErr:  false,
			message:  "WithEnvMap ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.EnvMap(map[string]string{"": ""})
			},
			field:    "Env",
			expected: make([]string, 0),
			wantErr:  true,
			message:  "WithEnvMap error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.ExposedPort("tcp", "80")
			},
			field:    "ExposedPorts",
			expected: nat.PortSet{nat.Port("80/tcp"): {}},
			wantErr:  false,
			message:  "WithExposedPort ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.ExposedPort("tcp", "300000")
			},
			field:    "ExposedPorts",
			expected: nat.PortSet{},
			wantErr:  true,
			message:  "WithExposedPort error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.HostName("test")
			},
			field:    "Hostname",
			expected: "test",
			wantErr:  false,
			message:  "WithHostName ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.HostName("")
			},
			field:    "Hostname",
			expected: "",
			wantErr:  true,
			message:  "WithHostName error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.DomainName("test")
			},
			field:    "Domainname",
			expected: "test",
			wantErr:  false,
			message:  "WithDomainName ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.DomainName("")
			},
			field:    "Domainname",
			expected: "",
			wantErr:  true,
			message:  "WithDomainName error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Image("test")
			},
			field:    "Image",
			expected: "test",
			wantErr:  false,
			message:  "WithImage ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Image("")
			},
			field:    "Image",
			expected: "",
			wantErr:  true,
			message:  "WithImage error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Imagef("test-%s", "test")
			},
			field:    "Image",
			expected: "test-test",
			wantErr:  false,
			message:  "WithImagef ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Imagef("%s", "")
			},
			field:    "Image",
			expected: "",
			wantErr:  true,
			message:  "WithImagef error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Command("test", "test2", "test3")
			},
			field:    "Cmd",
			expected: strslice.StrSlice([]string{"test", "test2", "test3"}),
			wantErr:  false,
			message:  "WithCommand ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Command()
			},
			field:    "Cmd",
			expected: strslice.StrSlice(strslice.StrSlice(nil)),
			wantErr:  true,
			message:  "WithCommand error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.User("test")
			},
			field:    "User",
			expected: "test",
			wantErr:  false,
			message:  "WithUser ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.User("")
			},
			field:    "User",
			expected: "",
			wantErr:  true,
			message:  "WithUser error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.AttachedStdin()
			},
			field:    "AttachStdin",
			expected: true,
			wantErr:  false,
			message:  "WithAttachedStdin ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.AttachedStdout()
			},
			field:    "AttachStdout",
			expected: true,
			wantErr:  false,
			message:  "WithAttachedStdout ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.AttachedStderr()
			},
			field:    "AttachStderr",
			expected: true,
			wantErr:  false,
			message:  "WithAttachedStderr ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Tty()
			},
			field:    "Tty",
			expected: true,
			wantErr:  false,
			message:  "WithTty ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StdinOpen()
			},
			field:    "OpenStdin",
			expected: true,
			wantErr:  false,
			message:  "WithStdinOpen ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StdinOpen()
			},
			field:    "OpenStdin",
			expected: true,
			wantErr:  false,
			message:  "WithStdinOpen ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StdinOnce()
			},
			field:    "StdinOnce",
			expected: true,
			wantErr:  false,
			message:  "WithStdinOnce ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.EscapedArgs()
			},
			field:    "ArgsEscaped",
			expected: true,
			wantErr:  false,
			message:  "WithEscapedArgs ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Volume("test")
			},
			field:    "Volumes",
			expected: map[string]struct{}{"test": {}},
			wantErr:  false,
			message:  "WithVolume ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Volume("")
			},
			field:    "Volumes",
			expected: nilMapStruct,
			wantErr:  true,
			message:  "WithVolume error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.WorkingDir("test")
			},
			field:    "WorkingDir",
			expected: "test",
			wantErr:  false,
			message:  "WithWorkingDir ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.WorkingDir("")
			},
			field:    "WorkingDir",
			expected: "",
			wantErr:  true,
			message:  "WithWorkingDir error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.DisabledNetwork()
			},
			field:    "NetworkDisabled",
			expected: true,
			wantErr:  false,
			message:  "WithDisabledNetwork ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.OnBuild("test", "test2", "test3")
			},
			field:    "OnBuild",
			expected: []string{"test", "test2", "test3"},
			wantErr:  false,
			message:  "WithOnBuild ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.OnBuild()
			},
			field:    "OnBuild",
			expected: nilStringSlice,
			wantErr:  true,
			message:  "WithOnBuild error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Label("test", "test")
			},
			field:    "Labels",
			expected: map[string]string{"test": "test"},
			wantErr:  false,
			message:  "WithLabel ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Label("", "")
			},
			field:    "Labels",
			expected: nilMapString,
			wantErr:  true,
			message:  "WithLabel error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StopSignal("test")
			},
			field:    "StopSignal",
			expected: "test",
			wantErr:  false,
			message:  "WithStopSignal ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StopSignal("")
			},
			field:    "StopSignal",
			expected: "",
			wantErr:  true,
			message:  "WithStopSignal error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Entrypoint("test", "test2", "test3")
			},
			field:    "Entrypoint",
			expected: strslice.StrSlice{"test", "test2", "test3"},
			wantErr:  false,
			message:  "WithEntrypoint ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Entrypoint()
			},
			field:    "Entrypoint",
			expected: nilDockerStrSlice,
			wantErr:  true,
			message:  "WithEntrypoint error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Shell("test", "test2", "test3")
			},
			field:    "Shell",
			expected: strslice.StrSlice{"test", "test2", "test3"},
			wantErr:  false,
			message:  "WithShell ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.Shell()
			},
			field:    "Shell",
			expected: nilDockerStrSlice,
			wantErr:  true,
			message:  "WithShell error",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StopTimeout(10)
			},
			field:    "StopTimeout",
			expected: &stopTimeout,
			wantErr:  false,
			message:  "WithStopTimeout ok",
		},
		{
			config: &container.Config{},
			setFn: func(c *Container) {
				c.StopTimeout(0)
			},
			field:    "StopTimeout",
			expected: nilStopTimeout,
			wantErr:  true,
			message:  "WithStopTimeout error",
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
