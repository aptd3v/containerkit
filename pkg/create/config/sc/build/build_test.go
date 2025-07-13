package build_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aptd3v/go-contain/pkg/create/config/sc/build"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/build/ulimit"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/secrets/secretservice"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	somestring = "test"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *types.BuildConfig
		setFn    build.SetBuildConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &types.BuildConfig{},
			setFn:    build.Failf("test error %s", "foo"),
			field:    "",
			wantErr:  true,
			message:  "Failf error",
			expected: nil,
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail error",
			expected: nil,
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithPrivileged(),
			field:    "Privileged",
			wantErr:  false,
			message:  "WithPrivileged ok",
			expected: true,
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithUlimit(""),
			field:    "Ulimits",
			wantErr:  true,
			message:  "WithUlimit empty name and no setters",
			expected: (map[string]*types.UlimitsConfig)(nil),
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithUlimit("foo", ulimit.Fail(errors.New("test error"))),
			field:    "Ulimits",
			wantErr:  true,
			message:  "WithUlimit error setter",
			expected: (map[string]*types.UlimitsConfig)(nil),
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithUlimit("foo"),
			field:    "Ulimits",
			wantErr:  false,
			message:  "WithUlimit no setters",
			expected: (map[string]*types.UlimitsConfig)(nil),
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithUlimit("foo", nil, nil),
			field:    "Ulimits",
			wantErr:  false,
			message:  "WithUlimit nil setters",
			expected: map[string]*types.UlimitsConfig{"foo": {}},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithUlimit("foo", ulimit.WithSoft(1000), ulimit.WithHard(2000)),
			field:    "Ulimits",
			wantErr:  false,
			message:  "WithUlimit ok",
			expected: map[string]*types.UlimitsConfig{"foo": {Soft: 1000, Hard: 2000}},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithTags("foo", "bar"),
			field:    "Tags",
			wantErr:  false,
			message:  "WithTags ok",
			expected: types.StringList{"foo", "bar"},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithSecret(nil, nil, nil),
			field:    "Secrets",
			wantErr:  false,
			message:  "WithSecret nil setters",
			expected: []types.ServiceSecretConfig{{}},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithSecret(),
			field:    "Secrets",
			wantErr:  false,
			message:  "WithSecret no setters",
			expected: ([]types.ServiceSecretConfig)(nil),
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithSecret(secretservice.Fail(errors.New("test error"))),
			field:    "Secrets",
			wantErr:  true,
			message:  "WithSecret error setters",
			expected: ([]types.ServiceSecretConfig)(nil),
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithTarget("foo"),
			field:    "Target",
			wantErr:  false,
			message:  "WithTarget ok",
			expected: "foo",
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithNetwork("foo"),
			field:    "Network",
			wantErr:  false,
			message:  "WithNetwork ok",
			expected: "foo",
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithPlatforms("foo", "bar"),
			field:    "Platforms",
			wantErr:  false,
			message:  "WithPlatforms ok",
			expected: types.StringList{"foo", "bar"},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithIsolation("foo"),
			field:    "Isolation",
			wantErr:  false,
			message:  "WithIsolation ok",
			expected: "foo",
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithExtraHosts("foo", "192.168.1.1", "192.168.1.2"),
			field:    "ExtraHosts",
			wantErr:  false,
			message:  "WithExtraHosts ok",
			expected: types.HostsList{"foo": []string{"192.168.1.1", "192.168.1.2"}},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithPull(),
			field:    "Pull",
			wantErr:  false,
			message:  "WithPull ok",
			expected: true,
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithAdditionalContexts("foo", "bar"),
			field:    "AdditionalContexts",
			wantErr:  false,
			message:  "WithAdditionalContexts ok",
			expected: types.Mapping{"foo": "bar"},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithNoCache(),
			field:    "NoCache",
			wantErr:  false,
			message:  "WithNoCache ok",
			expected: true,
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithCacheTo("foo"),
			field:    "CacheTo",
			wantErr:  false,
			message:  "WithCacheTo ok",
			expected: types.StringList{"foo"},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithCacheFrom("foo"),
			field:    "CacheFrom",
			wantErr:  false,
			message:  "WithCacheFrom ok",
			expected: types.StringList{"foo"},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithLabels("foo", "bar"),
			field:    "Labels",
			wantErr:  false,
			message:  "WithLabels ok",
			expected: types.Labels{"foo": "bar"},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithSSHKey("key", "/path/to/key"),
			field:    "SSH",
			wantErr:  false,
			message:  "WithSSHKey ok",
			expected: types.SSHConfig{types.SSHKey{ID: "key=/path/to/key"}},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithArgs("foo", "test"),
			field:    "Args",
			wantErr:  false,
			message:  "WithArgs ok",
			expected: types.MappingWithEquals{"foo": &somestring},
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithDockerfileInline("FROM alpine"),
			field:    "DockerfileInline",
			wantErr:  false,
			message:  "WithDockerfileInline ok",
			expected: "FROM alpine",
		},
		{
			config:   &types.BuildConfig{},
			setFn:    build.WithContext("foo"),
			field:    "Context",
			wantErr:  false,
			message:  "WithContext ok",
			expected: "foo",
		},

		{
			config:   &types.BuildConfig{},
			setFn:    build.WithDockerfile("foo"),
			field:    "Dockerfile",
			wantErr:  false,
			message:  "WithDockerfile ok",
			expected: "foo",
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
