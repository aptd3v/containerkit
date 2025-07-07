package cc_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

var nilMapString map[string]string
var nilMapStruct map[string]struct{}
var nilStringSlice []string
var nilDockerStrSlice strslice.StrSlice
var nilStopTimeout *int

func TestAssignments(t *testing.T) {
	stopTimeout := 10
	tests := []struct {
		config   *container.Config
		setFn    create.SetContainerConfig
		field    string
		expected any
		wantErr  bool
		message  string
	}{
		{
			config:   &container.Config{},
			setFn:    cc.WithEnv("TEST", "test"),
			field:    "Env",
			expected: []string{"TEST=test"},
			wantErr:  false,
			message:  "WithEnv ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithEnv("", ""),
			field:    "Env",
			expected: make([]string, 0),
			wantErr:  true,
			message:  "WithEnv error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithEnvMap(map[string]string{"TEST": "test"}),
			field:    "Env",
			expected: []string{"TEST=test"},
			wantErr:  false,
			message:  "WithEnvMap ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithEnvMap(map[string]string{"": ""}),
			field:    "Env",
			expected: make([]string, 0),
			wantErr:  true,
			message:  "WithEnvMap error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithExposedPort("tcp", "80"),
			field:    "ExposedPorts",
			expected: nat.PortSet{nat.Port("80/tcp"): {}},
			wantErr:  false,
			message:  "WithExposedPort ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithExposedPort("tcp", "300000"),
			field:    "ExposedPorts",
			expected: nat.PortSet{},
			wantErr:  true,
			message:  "WithExposedPort error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithHostName("test"),
			field:    "Hostname",
			expected: "test",
			wantErr:  false,
			message:  "WithHostName ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithHostName(""),
			field:    "Hostname",
			expected: "",
			wantErr:  true,
			message:  "WithHostName error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithDomainName("test"),
			field:    "Domainname",
			expected: "test",
			wantErr:  false,
			message:  "WithDomainName ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithDomainName(""),
			field:    "Domainname",
			expected: "",
			wantErr:  true,
			message:  "WithDomainName error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithImage("test"),
			field:    "Image",
			expected: "test",
			wantErr:  false,
			message:  "WithImage ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithImage(""),
			field:    "Image",
			expected: "",
			wantErr:  true,
			message:  "WithImage error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithImagef("test-%s", "test"),
			field:    "Image",
			expected: "test-test",
			wantErr:  false,
			message:  "WithImagef ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithImagef("%s", ""),
			field:    "Image",
			expected: "",
			wantErr:  true,
			message:  "WithImagef error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithCommand("test", "test2", "test3"),
			field:    "Cmd",
			expected: strslice.StrSlice([]string{"test", "test2", "test3"}),
			wantErr:  false,
			message:  "WithCommand ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithCommand(),
			field:    "Cmd",
			expected: strslice.StrSlice(strslice.StrSlice(nil)),
			wantErr:  true,
			message:  "WithCommand error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithUser("test"),
			field:    "User",
			expected: "test",
			wantErr:  false,
			message:  "WithUser ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithUser(""),
			field:    "User",
			expected: "",
			wantErr:  true,
			message:  "WithUser error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithAttachedStdin(),
			field:    "AttachStdin",
			expected: true,
			wantErr:  false,
			message:  "WithAttachedStdin ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithAttachedStdout(),
			field:    "AttachStdout",
			expected: true,
			wantErr:  false,
			message:  "WithAttachedStdout ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithAttachedStderr(),
			field:    "AttachStderr",
			expected: true,
			wantErr:  false,
			message:  "WithAttachedStderr ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithTty(),
			field:    "Tty",
			expected: true,
			wantErr:  false,
			message:  "WithTty ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithStdinOpen(),
			field:    "OpenStdin",
			expected: true,
			wantErr:  false,
			message:  "WithStdinOpen ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithStdinOpen(),
			field:    "OpenStdin",
			expected: true,
			wantErr:  false,
			message:  "WithStdinOpen ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithStdinOnce(),
			field:    "StdinOnce",
			expected: true,
			wantErr:  false,
			message:  "WithStdinOnce ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithEscapedArgs(),
			field:    "ArgsEscaped",
			expected: true,
			wantErr:  false,
			message:  "WithEscapedArgs ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithVolume("test"),
			field:    "Volumes",
			expected: map[string]struct{}{"test": {}},
			wantErr:  false,
			message:  "WithVolume ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithVolume(""),
			field:    "Volumes",
			expected: nilMapStruct,
			wantErr:  true,
			message:  "WithVolume error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithWorkingDir("test"),
			field:    "WorkingDir",
			expected: "test",
			wantErr:  false,
			message:  "WithWorkingDir ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithWorkingDir(""),
			field:    "WorkingDir",
			expected: "",
			wantErr:  true,
			message:  "WithWorkingDir error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithDisabledNetwork(),
			field:    "NetworkDisabled",
			expected: true,
			wantErr:  false,
			message:  "WithDisabledNetwork ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithOnBuild("test", "test2", "test3"),
			field:    "OnBuild",
			expected: []string{"test", "test2", "test3"},
			wantErr:  false,
			message:  "WithOnBuild ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithOnBuild(),
			field:    "OnBuild",
			expected: nilStringSlice,
			wantErr:  true,
			message:  "WithOnBuild error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithLabel("test", "test"),
			field:    "Labels",
			expected: map[string]string{"test": "test"},
			wantErr:  false,
			message:  "WithLabel ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithLabel("", ""),
			field:    "Labels",
			expected: nilMapString,
			wantErr:  true,
			message:  "WithLabel error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithStopSignal("test"),
			field:    "StopSignal",
			expected: "test",
			wantErr:  false,
			message:  "WithStopSignal ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithStopSignal(""),
			field:    "StopSignal",
			expected: "",
			wantErr:  true,
			message:  "WithStopSignal error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithEntrypoint("test", "test2", "test3"),
			field:    "Entrypoint",
			expected: strslice.StrSlice{"test", "test2", "test3"},
			wantErr:  false,
			message:  "WithEntrypoint ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithEntrypoint(),
			field:    "Entrypoint",
			expected: nilDockerStrSlice,
			wantErr:  true,
			message:  "WithEntrypoint error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithShell("test", "test2", "test3"),
			field:    "Shell",
			expected: strslice.StrSlice{"test", "test2", "test3"},
			wantErr:  false,
			message:  "WithShell ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithShell(),
			field:    "Shell",
			expected: nilDockerStrSlice,
			wantErr:  true,
			message:  "WithShell error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithStopTimeout(10),
			field:    "StopTimeout",
			expected: &stopTimeout,
			wantErr:  false,
			message:  "WithStopTimeout ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithStopTimeout(0),
			field:    "StopTimeout",
			expected: nilStopTimeout,
			wantErr:  true,
			message:  "WithStopTimeout error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithMacAddress("12:34:56:78:90:AB"),
			field:    "MacAddress",
			expected: "12:34:56:78:90:AB",
			wantErr:  false,
			message:  "WithMacAddress ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.WithMacAddress(""),
			field:    "MacAddress",
			expected: "",
			wantErr:  true,
			message:  "WithMacAddress error",
		},
		{
			config:   &container.Config{},
			setFn:    cc.Fail(errors.New("test")),
			field:    "",
			expected: nil,
			wantErr:  true,
			message:  "Fail ok",
		},
		{
			config:   &container.Config{},
			setFn:    cc.Failf("test-%s", "test"),
			field:    "",
			expected: nil,
			wantErr:  true,
			message:  "Failf ok",
		},
		{
			config: &container.Config{},
			setFn: cc.WithHealthCheck(
				nil,
				health.WithInterval(10),
				health.WithRetries(10),
				health.WithTimeout(10),
				health.WithStartPeriod(10),
				health.WithTest("CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"),
			),
			field: "Healthcheck",
			expected: &container.HealthConfig{
				Interval:    10 * time.Second,
				Retries:     10,
				StartPeriod: 10 * time.Second,
				Test:        []string{"CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"},
				Timeout:     10 * time.Second,
			},
			wantErr: false,
			message: "WithHealthCheck ok",
		},
		{
			config: &container.Config{},
			setFn: cc.WithHealthCheck(
				health.Fail(errors.New("test")),
			),
			field:    "Healthcheck",
			expected: &container.HealthConfig{},
			wantErr:  true,
			message:  "WithHealthCheck error",
		},
		{
			config: &container.Config{},
			setFn: cc.WithHealthCheck(
				health.WithTest(),
			),
			field: "Healthcheck",
			expected: &container.HealthConfig{
				Test: []string{"NONE"},
			},
			wantErr: false,
			message: "WithHealthCheck ok with empty test",
		},
	}

	for _, test := range tests {
		err := test.setFn(test.config)
		if test.wantErr {
			assert.Error(t, err)
			assert.True(t, create.IsContainerConfigError(err), "expected container config error")
		} else {
			assert.NoError(t, err)
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
