package client

import (
	"bytes"
	"context"
	"io"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestWaitCondition(t *testing.T) {
	_ = WaitConditionNotRunning
	_ = WaitConditionNextExit
	_ = WaitConditionRemoved
}

func TestContainerOptsApply(t *testing.T) {
	require.False(t, (*List)(nil).apply().Size)
	l := (&List{Size: true, All: true, Latest: true, Since: "a", Before: "b", Limit: 3, Filters: make(Filters).Add("id", "x")}).apply()
	require.True(t, l.All)
	require.Equal(t, 3, l.Limit)
	require.True(t, l.Filters["id"]["x"])
	require.True(t, l.Filters["since"]["a"])
	require.True(t, l.Filters["before"]["b"])
	require.Equal(t, 1, (&List{Latest: true}).apply().Limit)

	require.Equal(t, "", (*Start)(nil).apply().CheckpointID)
	require.Equal(t, "c", (&Start{CheckpointID: "c", CheckpointDir: "d"}).apply().CheckpointID)

	to := 5
	require.Nil(t, (*Stop)(nil).apply().Timeout)
	require.Equal(t, "SIGTERM", (&Stop{Timeout: &to, Signal: "SIGTERM"}).apply().Signal)

	require.False(t, (*Remove)(nil).apply().Force)
	require.True(t, (&Remove{Force: true, RemoveVolumes: true, RemoveLinks: true}).apply().Force)

	require.False(t, (*Logs)(nil).apply().Follow)
	lg := (&Logs{ShowStdout: true, ShowStderr: true, Since: "1", Until: "2", Timestamps: true, Follow: true, Tail: "10", Details: true}).apply()
	require.True(t, lg.Follow)

	require.Empty(t, (*Exec)(nil).apply().Cmd)
	ex := (&Exec{User: "root", Privileged: true, TTY: true, ConsoleSize: ConsoleSize{Width: 80, Height: 24}, AttachStdin: true, AttachStderr: true, AttachStdout: true, DetachKeys: "ctrl-c", Env: []string{"A=1"}, WorkingDir: "/", Cmd: []string{"true"}}).apply()
	require.Equal(t, []string{"true"}, ex.Cmd)
	require.Equal(t, uint(80), ex.ConsoleSize.Width)
	require.Equal(t, uint(24), ex.ConsoleSize.Height)

	require.False(t, (*ExecStart)(nil).apply().TTY)
	es := (&ExecStart{Detach: true, TTY: true, ConsoleSize: ConsoleSize{Width: 1, Height: 2}}).apply()
	require.Equal(t, uint(1), es.ConsoleSize.Width)

	require.Equal(t, uint(0), (*Resize)(nil).apply().Width)
	require.Equal(t, uint(10), (&Resize{Width: 10, Height: 20}).apply().Width)

	require.False(t, (*ExecAttach)(nil).apply().TTY)
	ea := (&ExecAttach{TTY: true, ConsoleSize: ConsoleSize{Width: 3, Height: 4}}).apply()
	require.Equal(t, uint(3), ea.ConsoleSize.Width)

	require.False(t, (*Attach)(nil).apply().Stream)
	at := (&Attach{Stream: true, Stdin: true, Stdout: true, Stderr: true, DetachKeys: "ctrl-p", Logs: true}).apply()
	require.True(t, at.Stream)

	require.Empty(t, (*Prune)(nil).apply().Filters)
	require.Len(t, (&Prune{Filters: make(Filters).Add("label", "x")}).apply().Filters, 1)

	require.Equal(t, "", (*Commit)(nil).apply().Reference)
	ctr := kit.NewContainer("c").Image("alpine")
	cfg, err := convert[container.Config](ctr.Config.Container)
	require.NoError(t, err)
	cm := (&Commit{Reference: "t:1", Comment: "c", Author: "a", Changes: []string{"CMD true"}, NoPause: false, Config: cfg}).apply()
	require.Equal(t, "t:1", cm.Reference)
	require.NotNil(t, cm.Config)
	require.False(t, cm.NoPause)

	require.Nil(t, (*Update)(nil).apply().Resources)
	up := (&Update{
		Resources:     &container.Resources{CPUShares: 10, Memory: 1},
		RestartPolicy: &container.RestartPolicy{Name: container.RestartPolicyOnFailure, MaximumRetryCount: 2},
	}).apply()
	require.Equal(t, int64(10), up.Resources.CPUShares)
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
	require.False(t, (*ImagePull)(nil).apply().All)
	p := (&ImagePull{All: true, PrivilegeFunc: func(context.Context) (string, error) { return "", nil }, RegistryAuth: "auth", Platforms: []ocispec.Platform{{OS: "linux", Architecture: "amd64"}}}).apply()
	require.Equal(t, "auth", p.RegistryAuth)
	require.NotEmpty(t, p.Platforms)

	c := (&ImageCreate{RegistryAuth: "auth", Platforms: []ocispec.Platform{{OS: "linux"}}}).apply()
	require.Equal(t, "auth", c.RegistryAuth)
	require.Equal(t, "linux", c.Platforms[0].OS)

	require.False(t, (*ImageList)(nil).apply().All)
	require.False(t, (*Ping)(nil).apply().NegotiateAPIVersion)
	require.True(t, (&Ping{NegotiateAPIVersion: true, ForceNegotiate: true}).apply().ForceNegotiate)
	listed := (&ImageList{All: true, SharedSize: true, Identity: true, Manifests: true, Filters: make(Filters).Add("dangling", "true")}).apply()
	require.True(t, listed.All)
	require.True(t, listed.Identity)
	require.True(t, listed.Manifests)

	require.Empty(t, (*ImageBuild)(nil).apply().Tags)
	b := &ImageBuild{}
	b.Tags = []string{"x"}
	require.Equal(t, []string{"x"}, b.apply().Tags)

	require.False(t, (*ImageRemove)(nil).apply().Force)
	require.True(t, (&ImageRemove{Force: true, PruneChildren: true, Platforms: []ocispec.Platform{{OS: "linux"}}}).apply().Force)

	s := (*ImageSearch)(nil).apply()
	require.Equal(t, 0, s.Limit)
	s = (&ImageSearch{Limit: 5, Filters: make(Filters).Add("is-official", "true"), RegistryAuth: "auth", PrivilegeFunc: func(context.Context) (string, error) { return "", nil }}).apply()
	require.Equal(t, 5, s.Limit)
	require.Equal(t, "auth", s.RegistryAuth)

	require.Empty(t, (*ImagePrune)(nil).apply().Filters)
	require.Len(t, (&ImagePrune{Filters: make(Filters).Add("dangling", "true")}).apply().Filters, 1)

	push := (&ImagePush{All: true, RegistryAuth: "auth", Platform: &ocispec.Platform{OS: "linux", Architecture: "amd64"}}).apply()
	require.Equal(t, "auth", push.RegistryAuth)
	require.True(t, push.All)
	require.Equal(t, "amd64", push.Platform.Architecture)
	require.Empty(t, (*ImagePush)(nil).apply().RegistryAuth)

	inspect := (&ImageInspect{Manifests: true, Platform: &ocispec.Platform{OS: "linux"}}).options()
	require.Len(t, inspect, 2)
	require.Empty(t, (*ImageInspect)(nil).options())
}

func TestNetworkVolumeSwarmOptsApply(t *testing.T) {
	require.Equal(t, "", (*NetworkCreate)(nil).apply().Driver)
	v4 := true
	subnet := netip.MustParsePrefix("10.0.0.0/24")
	ipRange := netip.MustParsePrefix("10.0.0.0/28")
	gw := netip.MustParseAddr("10.0.0.1")
	aux := netip.MustParseAddr("10.0.0.2")
	nc := (&NetworkCreate{
		Driver: "bridge", Scope: "local", EnableIPv4: &v4, EnableIPv6: &v4,
		Internal: true, Attachable: true, Ingress: false, ConfigOnly: false,
		Options: map[string]string{"a": "b"}, Labels: map[string]string{"l": "1"},
		IPAM: &network.IPAM{
			Driver: "default", Options: map[string]string{"o": "1"},
			Config: []network.IPAMConfig{{Subnet: subnet, IPRange: ipRange, Gateway: gw, AuxAddress: map[string]netip.Addr{"h": aux}}},
		},
	}).apply()
	require.NotNil(t, nc.IPAM)
	require.Len(t, nc.IPAM.Config, 1)

	require.Equal(t, "", (*NetworkConnect)(nil).apply().Container)
	require.Equal(t, "c", (&NetworkConnect{Container: "c", EndpointConfig: &network.EndpointSettings{}}).apply().Container)

	require.False(t, (*NetworkInspect)(nil).apply().Verbose)
	require.True(t, (&NetworkInspect{Scope: "local", Verbose: true}).apply().Verbose)
	require.Empty(t, (*NetworkList)(nil).apply().Filters)
	require.Len(t, (&NetworkList{Filters: make(Filters).Add("name", "n")}).apply().Filters, 1)
	require.Empty(t, (*NetworkPrune)(nil).apply().Filters)
	require.Len(t, (&NetworkPrune{Filters: make(Filters).Add("label", "x")}).apply().Filters, 1)

	require.Equal(t, "", (*VolumeCreate)(nil).apply().Name)
	require.Equal(t, "v", (&VolumeCreate{Name: "v", Driver: "local", DriverOpts: map[string]string{"a": "1"}, Labels: map[string]string{"l": "1"}, ClusterVolumeSpec: &volume.ClusterVolumeSpec{}}).apply().Name)
	require.Empty(t, (*VolumeList)(nil).apply().Filters)
	require.Len(t, (&VolumeList{Filters: make(Filters).Add("name", "v")}).apply().Filters, 1)
	require.Nil(t, (*VolumeUpdate)(nil).apply().Spec)
	require.NotNil(t, (&VolumeUpdate{Spec: &volume.ClusterVolumeSpec{}}).apply().Spec)
	require.Empty(t, (*VolumePrune)(nil).apply().Filters)
	require.False(t, (*VolumePrune)(nil).apply().All)
	require.Len(t, (&VolumePrune{All: true, Filters: make(Filters).Add("label", "x")}).apply().Filters, 1)
	require.True(t, (&VolumePrune{All: true}).apply().All)

	require.Equal(t, "", (*SwarmInit)(nil).apply().ListenAddr)
	init := (&SwarmInit{
		ListenAddr: "0.0.0.0:2377", AdvertiseAddr: "127.0.0.1", DataPathAddr: "127.0.0.1",
		DataPathPort: 4789, ForceNewCluster: true, Spec: swarm.Spec{}, AutoLockManagers: true,
		Availability: swarm.NodeAvailabilityActive, DefaultAddrPool: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}, SubnetSize: 24,
	}).apply()
	require.True(t, init.ForceNewCluster)
	require.Len(t, init.DefaultAddrPool, 1)
	require.Equal(t, "", (*SwarmJoin)(nil).apply().JoinToken)
	j := (&SwarmJoin{ListenAddr: "0.0.0.0:2377", AdvertiseAddr: "1.1.1.1", DataPathAddr: "1.1.1.1", RemoteAddrs: []string{"2.2.2.2:2377"}, JoinToken: "tok", Availability: swarm.NodeAvailabilityActive}).apply()
	require.Equal(t, "tok", j.JoinToken)
}

func TestBuildCachePruneOptsApply(t *testing.T) {
	require.False(t, (*BuildCachePrune)(nil).apply().All)
	require.Empty(t, (*BuildCachePrune)(nil).apply().Filters)
	pruned := (&BuildCachePrune{
		All: true, ReservedSpace: 10, MaxUsedSpace: 20, MinFreeSpace: 30,
		Filters: make(Filters).Add("until", "24h"),
	}).apply()
	require.True(t, pruned.All)
	require.Equal(t, int64(10), pruned.ReservedSpace)
	require.Equal(t, int64(20), pruned.MaxUsedSpace)
	require.Equal(t, int64(30), pruned.MinFreeSpace)
	require.Len(t, pruned.Filters, 1)
}

func TestPluginOptsApply(t *testing.T) {
	require.Empty(t, (*PluginList)(nil).apply().Filters)
	require.Len(t, (&PluginList{Filters: make(Filters).Add("enabled", "true")}).apply().Filters, 1)

	installed := (&PluginInstall{
		Disabled: true, AcceptAllPermissions: true, RemoteRef: "example/plugin",
		Args: []string{"DEBUG=1"}, RegistryAuth: "auth",
	}).apply()
	require.True(t, installed.Disabled)
	require.True(t, installed.AcceptAllPermissions)
	require.Equal(t, "example/plugin", installed.RemoteRef)
	require.Equal(t, []string{"DEBUG=1"}, installed.Args)
	require.Equal(t, "auth", installed.RegistryAuth)
	require.Empty(t, (*PluginInstall)(nil).apply().RegistryAuth)

	require.Equal(t, "auth", (&PluginUpgrade{RegistryAuth: "auth"}).apply().RegistryAuth)
	require.Equal(t, "auth", (&PluginPush{RegistryAuth: "auth"}).apply().RegistryAuth)
	require.Empty(t, (*PluginPush)(nil).apply().RegistryAuth)

	require.Equal(t, "repo", (&PluginCreate{RepoName: "repo"}).apply().RepoName)
	require.Equal(t, 30, (&PluginEnable{Timeout: 30}).apply().Timeout)
	require.True(t, (&PluginDisable{Force: true}).apply().Force)
	require.False(t, (*PluginRemove)(nil).apply().Force)
	require.Equal(t, []string{"A=1"}, (&PluginSet{Args: []string{"A=1"}}).apply().Args)
}

func TestSwarmObjectOptsApply(t *testing.T) {
	require.Empty(t, (*NodeList)(nil).apply().Filters)
	require.Len(t, (&NodeList{Filters: make(Filters).Add("role", "manager")}).apply().Filters, 1)
	require.True(t, (&NodeRemove{Force: true}).apply().Force)
	require.False(t, (*NodeRemove)(nil).apply().Force)
	require.Equal(t, uint64(3), (&NodeUpdate{Version: swarm.Version{Index: 3}}).apply().Version.Index)

	created := (&ServiceCreate{EncodedRegistryAuth: "auth", QueryRegistry: true}).apply()
	require.Equal(t, "auth", created.EncodedRegistryAuth)
	require.True(t, created.QueryRegistry)
	require.Empty(t, (*ServiceCreate)(nil).apply().EncodedRegistryAuth)

	require.True(t, (&ServiceInspect{InsertDefaults: true}).apply().InsertDefaults)
	listed := (&ServiceList{Status: true, Filters: make(Filters).Add("name", "web")}).apply()
	require.True(t, listed.Status)
	require.Len(t, listed.Filters, 1)

	updated := (&ServiceUpdate{Rollback: "previous", RegistryAuthFrom: "spec", EncodedRegistryAuth: "auth"}).apply()
	require.Equal(t, "previous", updated.Rollback)
	require.Equal(t, swarm.RegistryAuthSource("spec"), updated.RegistryAuthFrom)
	require.Equal(t, "auth", updated.EncodedRegistryAuth)

	require.True(t, (&ServiceLogs{Follow: true, Tail: "10"}).apply().Follow)
	require.True(t, (&TaskLogs{Details: true}).apply().Details)
	require.Equal(t, "tok", (&SecretCreate{Spec: swarm.SecretSpec{Annotations: swarm.Annotations{Name: "tok"}}}).apply().Spec.Annotations.Name)
	require.Equal(t, "cfg", (&ConfigCreate{Spec: swarm.ConfigSpec{Annotations: swarm.Annotations{Name: "cfg"}}}).apply().Spec.Annotations.Name)
	require.True(t, (&SwarmUpdate{RotateManagerToken: true, RotateWorkerToken: true, RotateManagerUnlockKey: true}).apply().RotateManagerToken)
	require.Equal(t, "key", (&SwarmUnlock{Key: "key"}).apply().Key)
	require.Empty(t, (*SecretList)(nil).apply().Filters)
	require.Equal(t, uint64(1), (&SecretUpdate{Version: swarm.Version{Index: 1}}).apply().Version.Index)
	require.Equal(t, uint64(2), (&ConfigUpdate{Version: swarm.Version{Index: 2}}).apply().Version.Index)
}

func TestContainerConfigConverts(t *testing.T) {
	ctr := kit.NewContainer("web").Image("alpine")
	cfg, err := convert[container.Config](ctr.Config.Container)
	require.NoError(t, err)
	require.Equal(t, "alpine", cfg.Image)
	_, err = convert[container.HostConfig](ctr.Config.Host)
	require.NoError(t, err)
	_, err = convert[network.NetworkingConfig](ctr.Config.Network)
	require.NoError(t, err)
}

func TestNewClientConstruction(t *testing.T) {
	cli, err := NewClient(nil, FromEnv(), WithTimeout(5*time.Second), WithUserAgent("containerkit-test"), WithHTTPHeaders(map[string]string{"X-Test": "1"}), WithAPIVersion(""), WithAPIVersionNegotiation(), WithAPIVersionFromEnv(), WithHostFromEnv(), WithTLSClientConfigFromEnv(), WithScheme("http"))
	require.NoError(t, err)
	require.NotNil(t, cli.Unwrap())

	_, err = NewClient(WithHost("unix:///var/run/docker.sock"), WithDialContext(nil))
	_ = err

	_, err = NewClient(WithTLSClientConfig("/no/ca", "/no/cert", "/no/key"))
	require.Error(t, err)

	_, err = NewClient(WithTraceProvider(nil), WithTraceOptions())
	require.NoError(t, err)
	_ = bytes.Buffer{}
	_ = io.Discard
}
