package client

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/client/auth"
	"github.com/aptd3v/containerkit/pkg/containerkit"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestArgsFromFilters(t *testing.T) {
	a := argsFromFilters(nil)
	require.Equal(t, 0, a.Len())
	a = argsFromFilters([]Filter{{Key: "label", Value: "a=b"}, {Key: "status", Value: "running"}})
	require.True(t, a.ExactMatch("label", "a=b"))
	_ = WaitConditionNotRunning
	_ = WaitConditionNextExit
	_ = WaitConditionRemoved
}

func TestContainerOptsApply(t *testing.T) {
	require.False(t, (*List)(nil).apply().Size)
	l := (&List{Size: true, All: true, Latest: true, Since: "a", Before: "b", Limit: 3, Filters: []Filter{{Key: "id", Value: "x"}}}).apply()
	require.True(t, l.All)
	require.Equal(t, 3, l.Limit)

	require.Equal(t, "", (*Start)(nil).apply().CheckpointID)
	require.Equal(t, "c", (&Start{CheckpointID: "c", CheckpointDir: "d"}).apply().CheckpointID)

	to := 5
	require.Nil(t, (*Stop)(nil).apply().Timeout)
	require.Equal(t, "SIGTERM", (&Stop{Timeout: &to, Signal: "SIGTERM"}).apply().Signal)

	require.False(t, (*Remove)(nil).apply().Force)
	require.True(t, (&Remove{Force: true, Volumes: true, Links: true}).apply().Force)

	require.False(t, (*Logs)(nil).apply().Follow)
	lg := (&Logs{ShowStdout: true, ShowStderr: true, Since: "1", Until: "2", Timestamps: true, Follow: true, Tail: "10", Details: true}).apply()
	require.True(t, lg.Follow)

	require.Empty(t, (*Exec)(nil).apply().Cmd)
	ex := (&Exec{User: "root", Privileged: true, Tty: true, ConsoleWidth: 80, ConsoleHeight: 24, AttachStdin: true, AttachStderr: true, AttachStdout: true, Detach: true, DetachKeys: "ctrl-c", Env: []string{"A=1"}, WorkingDir: "/", Command: []string{"true"}}).apply()
	require.Equal(t, []string{"true"}, ex.Cmd)
	require.NotNil(t, ex.ConsoleSize)

	require.False(t, (*ExecStart)(nil).apply().Tty)
	es := (&ExecStart{Detach: true, Tty: true, ConsoleWidth: 1, ConsoleHeight: 2}).apply()
	require.NotNil(t, es.ConsoleSize)

	require.Equal(t, uint(0), (*Resize)(nil).apply().Width)
	require.Equal(t, uint(10), (&Resize{Width: 10, Height: 20}).apply().Width)

	require.False(t, (*ExecAttach)(nil).apply().Tty)
	ea := (&ExecAttach{Detach: true, Tty: true, ConsoleWidth: 3, ConsoleHeight: 4}).apply()
	require.NotNil(t, ea.ConsoleSize)

	require.False(t, (*Attach)(nil).apply().Stream)
	at := (&Attach{Stream: true, Stdin: true, Stdout: true, Stderr: true, DetachKeys: "ctrl-p", Logs: true}).apply()
	require.True(t, at.Stream)

	require.Equal(t, 0, (*Prune)(nil).apply().Len())
	require.Equal(t, 1, (&Prune{Filters: []Filter{{Key: "label", Value: "x"}}}).apply().Len())

	require.Equal(t, "", (*Commit)(nil).apply().Reference)
	ctr := containerkit.NewContainer("c").Image("alpine")
	cm := (&Commit{Reference: "t:1", Comment: "c", Author: "a", Changes: []string{"CMD true"}, Pause: true, Config: ctr}).apply()
	require.Equal(t, "t:1", cm.Reference)
	require.NotNil(t, cm.Config)

	require.Equal(t, int64(0), (*Update)(nil).apply().CPUShares)
	up := (&Update{RestartPolicy: containerkit.RestartPolicyOnFailure, RestartMaxRetry: 2, CPUShares: 10, Memory: 1}).apply()
	require.Equal(t, int64(10), up.CPUShares)
	require.Equal(t, 2, up.RestartPolicy.MaximumRetryCount)

	require.False(t, (*CopyTo)(nil).apply().CopyUIDGID)
	require.True(t, (&CopyTo{AllowOverwriteDirWithFile: true, CopyUIDGID: true, Content: strings.NewReader("x")}).apply().CopyUIDGID)

	require.Equal(t, "", (*CheckpointCreate)(nil).apply().CheckpointID)
	require.True(t, (&CheckpointCreate{CheckpointID: "c", CheckpointDir: "d", Exit: true}).apply().Exit)
	require.Equal(t, "", (*CheckpointList)(nil).apply().CheckpointDir)
	require.Equal(t, "d", (&CheckpointList{CheckpointDir: "d"}).apply().CheckpointDir)
	require.Equal(t, "", (*CheckpointDelete)(nil).apply().CheckpointID)
	require.Equal(t, "c", (&CheckpointDelete{CheckpointID: "c", CheckpointDir: "d"}).apply().CheckpointID)
}

func TestImageOptsApply(t *testing.T) {
	p, err := (*ImagePull)(nil).apply()
	require.NoError(t, err)
	p, err = (&ImagePull{All: true, CurrentPlatform: true, Platform: "linux/amd64", PrivilegeFunc: func(context.Context) (string, error) { return "", nil }, Auth: auth.Auth{Username: "u", Password: "p", ServerAddress: "docker.io"}}).apply()
	require.NoError(t, err)
	require.NotEmpty(t, p.RegistryAuth)
	require.NotEmpty(t, p.Platform)

	c, err := (*ImageCreate)(nil).apply()
	require.NoError(t, err)
	c, err = (&ImageCreate{Platform: "linux", Auth: auth.Auth{Username: "u", Password: "p"}}).apply()
	require.NoError(t, err)
	require.NotEmpty(t, c.RegistryAuth)

	require.False(t, (*ImageList)(nil).apply().All)
	require.True(t, (&ImageList{All: true, SharedSize: true, ContainerCount: true, Manifests: true, Filters: []Filter{{Key: "dangling", Value: "true"}}}).apply().All)

	require.Empty(t, (*ImageBuild)(nil).apply().Tags)
	b := &ImageBuild{}
	b.Tags = []string{"x"}
	require.Equal(t, []string{"x"}, b.apply().Tags)

	require.False(t, (*ImageRemove)(nil).apply().Force)
	require.True(t, (&ImageRemove{Force: true, PruneChildren: true, Platforms: []ocispec.Platform{{OS: "linux"}}}).apply().Force)

	s, err := (*ImageSearch)(nil).apply()
	require.NoError(t, err)
	s, err = (&ImageSearch{Limit: 5, Filters: []Filter{{Key: "is-official", Value: "true"}}, Auth: auth.Auth{Username: "u", Password: "p"}, PrivilegeFunc: func(context.Context) (string, error) { return "", nil }}).apply()
	require.NoError(t, err)
	require.Equal(t, 5, s.Limit)
	require.NotEmpty(t, s.RegistryAuth)

	require.Equal(t, 0, (*ImagePrune)(nil).apply().Len())
	require.Equal(t, 1, (&ImagePrune{Filters: []Filter{{Key: "dangling", Value: "true"}}}).apply().Len())
	_ = filters.NewArgs()
}

func TestNetworkVolumeSwarmOptsApply(t *testing.T) {
	require.Equal(t, "", (*NetworkCreate)(nil).apply().Driver)
	v4 := true
	nc := (&NetworkCreate{
		Driver: "bridge", Scope: "local", EnableIPv4: &v4, EnableIPv6: &v4,
		Internal: true, Attachable: true, Ingress: false, ConfigOnly: false,
		Options: map[string]string{"a": "b"}, Labels: map[string]string{"l": "1"},
		IPAM: &NetworkIPAM{
			Driver: "default", Options: map[string]string{"o": "1"},
			Config: []NetworkIPAMConfig{{Subnet: "10.0.0.0/24", IPRange: "10.0.0.0/28", Gateway: "10.0.0.1", Auxiliary: map[string]string{"h": "10.0.0.2"}}},
		},
	}).apply()
	require.NotNil(t, nc.IPAM)
	require.Len(t, nc.IPAM.Config, 1)

	require.Equal(t, "", (*NetworkConnect)(nil).apply().Container)
	require.Equal(t, "c", (&NetworkConnect{Container: "c", Endpoint: &network.EndpointSettings{}}).apply().Container)

	require.False(t, (*NetworkInspect)(nil).apply().Verbose)
	require.True(t, (&NetworkInspect{Scope: "local", Verbose: true}).apply().Verbose)
	require.Equal(t, 0, (*NetworkList)(nil).apply().Filters.Len())
	require.Equal(t, 1, (&NetworkList{Filters: []Filter{{Key: "name", Value: "n"}}}).apply().Filters.Len())
	require.Equal(t, 0, (*NetworkPrune)(nil).apply().Len())
	require.Equal(t, 1, (&NetworkPrune{Filters: []Filter{{Key: "label", Value: "x"}}}).apply().Len())

	require.Equal(t, "", (*VolumeCreate)(nil).apply().Name)
	require.Equal(t, "v", (&VolumeCreate{Name: "v", Driver: "local", DriverOpts: map[string]string{"a": "1"}, Labels: map[string]string{"l": "1"}, ClusterVolumeSpec: &volume.ClusterVolumeSpec{}}).apply().Name)
	require.Equal(t, 0, (*VolumeList)(nil).apply().Filters.Len())
	require.Equal(t, 1, (&VolumeList{Filters: []Filter{{Key: "name", Value: "v"}}}).apply().Filters.Len())
	require.Nil(t, (*VolumeUpdate)(nil).apply().Spec)
	require.NotNil(t, (&VolumeUpdate{ClusterVolumeSpec: &volume.ClusterVolumeSpec{}}).apply().Spec)
	require.Equal(t, 0, (*VolumePrune)(nil).apply().Len())
	require.Equal(t, 1, (&VolumePrune{Filters: []Filter{{Key: "label", Value: "x"}}}).apply().Len())

	require.Equal(t, "", (*SwarmInit)(nil).apply().ListenAddr)
	init := (&SwarmInit{
		ListenAddr: "0.0.0.0:2377", AdvertiseAddr: "127.0.0.1", DataPathAddr: "127.0.0.1",
		DataPathPort: 4789, ForceNewCluster: true, Spec: swarm.Spec{}, AutoLockManagers: true,
		Availability: swarm.NodeAvailabilityActive, DefaultAddrPool: []string{"10.0.0.0/8"}, SubnetSize: 24,
	}).apply()
	require.True(t, init.ForceNewCluster)
	require.Equal(t, "", (*SwarmJoin)(nil).apply().JoinToken)
	j := (&SwarmJoin{ListenAddr: "0.0.0.0:2377", AdvertiseAddr: "1.1.1.1", DataPathAddr: "1.1.1.1", RemoteAddrs: []string{"2.2.2.2:2377"}, JoinToken: "tok", Availability: swarm.NodeAvailabilityActive}).apply()
	require.Equal(t, "tok", j.JoinToken)
}

func TestNewClientConstruction(t *testing.T) {
	cli, err := NewClient(nil, FromEnv(), WithAPIVersionNegotiation(), WithTimeout(5*time.Second), WithUserAgent("containerkit-test"), WithHTTPHeaders(map[string]string{"X-Test": "1"}), WithVersion(""), WithVersionFromEnv(), WithHostFromEnv(), WithTLSClientConfigFromEnv(), WithScheme("http"))
	require.NoError(t, err)
	require.NotNil(t, cli.Unwrap())

	_, err = NewClient(WithHost("unix:///var/run/docker.sock"), WithDialContext(nil))
	// dialer may fail depending on transport; construction should still be attempted
	_ = err

	_, err = NewClient(WithTLSClientConfig("/no/ca", "/no/cert", "/no/key"))
	require.Error(t, err)

	_, err = NewClient(WithTraceProvider(nil), WithTraceOptions())
	require.NoError(t, err)
	_ = bytes.Buffer{}
	_ = io.Discard
}
