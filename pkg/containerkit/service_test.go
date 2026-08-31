package containerkit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

func TestProjectMarshalFullConfig(t *testing.T) {
	mode := int64(0444)
	replicas := 2
	c := NewContainer("web-1").
		Image("nginx:latest").
		Command("nginx", "-g", "daemon off;").
		Env("FOO", "bar").
		Tty().
		ExposedPort("tcp", "80").
		HealthCheck(Health{Test: []string{"CMD", "true"}, Timeout: 1, Interval: 1, Retries: 1}).
		Entrypoint("/docker-entrypoint.sh").
		StdinOpen().
		StopSignal("SIGQUIT").
		WorkingDir("/usr/share/nginx/html").
		Label("app", "web").
		DomainName("web.test").
		HostName("web").
		User("nginx").
		StopTimeout(5).
		BlkioWeight(200).
		BlkioDeviceReadBps("/dev/zero", 1000).
		BlkioDeviceWriteBps("/dev/zero", 1000).
		BlkioDeviceReadIOps("/dev/zero", 10).
		BlkioDeviceWriteIOps("/dev/zero", 10).
		AddedCapabilities(NET_BIND_SERVICE).
		DroppedCapabilities(MKNOD).
		CgroupParent("system.slice").
		DeviceCgroupRules("c 1:3 rwm").
		CPUCount(1).
		CPUPercent(20).
		CPUPeriod(100000).
		CPUQuota(50000).
		CPUShares(256).
		CpusetCpus("0").
		CPURealtimeRuntime(0).
		CPURealtimePeriod(0).
		PidMode("host").
		MemoryReservation(8*1024*1024).
		MemorySwap(64*1024*1024).
		MemoryLimit(32*1024*1024).
		ShmSize(1024*1024).
		DNSLookups("8.8.8.8").
		DNSSearches("test").
		DNSOptions("ndots:2").
		OomScoreAdj(50).
		OomKillDisable().
		AddedDevice("/dev/null", "/dev/null", "rwm").
		AddedGroups("www-data").
		Init().
		IpcMode("shareable").
		Isolation("default").
		LogDriver("json-file", map[string]string{"max-size": "1m"}).
		VolumesFrom("other:rw").
		VolumeBinds("/tmp:/opt/bind:ro,z").
		RWNamedVolumeMount("data", "/var/lib/data").
		TmpfsMount("/tmpfs", 1024, 0755).
		PortBindings("tcp", "127.0.0.1", "8080", "80").
		Privileged().
		ReadOnlyRootfs().
		RestartPolicyUnlessStopped().
		Runtime("runc").
		SecurityOpts("no-new-privileges:true").
		Sysctls("net.ipv4.ip_forward", "0").
		Tmpfs("/run", "rw").
		Ulimits("nofile", 1024, 2048).
		UserNSMode("host").
		UTSMode("host").
		MemorySwappiness(40).
		PidsLimit(50).
		Architecture("amd64").
		Endpoint("front", EndpointOpts{Aliases: []string{"web"}, IPv4: "10.1.0.10"})

	p := NewProject("full")
	p.WithNetwork("front", ProjectNetwork{
		Driver:        "bridge",
		DriverOptions: map[string]string{"com.docker.network.bridge.name": "br-full"},
		Internal:      true,
		Attachable:    true,
		EnableIPv6:    true,
		Labels:        map[string]string{"n": "1"},
		IPAMDriver:    "default",
		IPAMPools: []IPAMPool{{
			Subnet:             "10.1.0.0/24",
			Gateway:            "10.1.0.1",
			IPRange:            "10.1.0.0/28",
			AuxiliaryAddresses: map[string]string{"host": "10.1.0.2"},
		}},
	})
	p.WithVolume("data", ProjectVolume{
		Driver:        "local",
		DriverOptions: map[string]string{"type": "tmpfs"},
		Labels:        map[string]string{"v": "1"},
	})
	p.WithSecret("token", ProjectSecret{
		Name:           "token",
		File:           "./token.txt",
		Content:        "secret",
		Environment:    "TOKEN",
		External:       false,
		Driver:         "file",
		DriverOptions:  map[string]string{"opt": "1"},
		TemplateDriver: "golang",
	})
	p.WithService("web", c,
		nil,
		DependsOn("db"),
		DependsOnHealthy("db"),
		Profiles("full"),
		EnvFile(".env"),
		NoAttach(),
		Annotation("com.example", "yes"),
		Develop(WatchActionSync, "./src", "/app", "tmp"),
		BuildSpec{
			Context:          ".",
			Dockerfile:       "Dockerfile",
			DockerfileInline: "FROM alpine\n",
			Args:             map[string]string{"A": "1"},
			Labels:           map[string]string{"b": "2"},
			Tags:             []string{"web:dev"},
			Target:           "runtime",
			Network:          "host",
			NoCache:          true,
			Pull:             true,
			Privileged:       true,
			Platforms:        []string{"linux/amd64"},
			CacheFrom:        []string{"type=local,src=/tmp/c"},
			CacheTo:          []string{"type=inline"},
		},
		&BuildSpec{Context: "./app"},
		Deploy{
			Mode:     "replicated",
			Replicas: &replicas,
			Labels:   map[string]string{"d": "1"},
			Limits:   &Resources{NanoCPUs: 1e9, MemoryBytes: 64 * 1024 * 1024, Pids: 10},
			Reservations: &Resources{
				NanoCPUs:    5e8,
				MemoryBytes: 32 * 1024 * 1024,
			},
		},
		ServiceSecret{Source: "token", Target: "/run/secrets/token", UID: "0", GID: "0", Mode: &mode},
	)
	require.NoError(t, p.Validate())

	raw, err := p.Marshal()
	require.NoError(t, err)
	require.Contains(t, string(raw), "image:")
	require.NotNil(t, p.Unwrap())

	svc, err := p.GetService("web")
	require.NoError(t, err)
	require.Equal(t, "web", svc.Name)
	require.Empty(t, svc.ContainerName, "deploy set should clear container name")

	n := 0
	require.NoError(t, p.ForEachService(func(name string, service *types.ServiceConfig) error {
		n++
		require.Equal(t, "web", name)
		return nil
	}))
	require.Equal(t, 1, n)
	require.Error(t, p.ForEachService(func(string, *types.ServiceConfig) error {
		return errdefs.NewProjectConfigError("web", "stop")
	}))

	path := filepath.Join(t.TempDir(), "compose.yml")
	require.NoError(t, p.Export(path, 0644))
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestProjectValidationAndErrors(t *testing.T) {
	empty := NewProject("empty")
	require.Error(t, empty.Validate())
	_, err := empty.Marshal()
	require.Error(t, err)
	require.Error(t, empty.Export(filepath.Join(t.TempDir(), "x.yml"), 0644))

	p := NewProject("p")
	p.WithService("bad", NewContainer("bad"))
	require.Error(t, p.Validate())

	ok := NewContainer().Image("alpine")
	p2 := NewProject("p2")
	p2.WithService("a", ok)
	p2.WithService("a", ok)
	require.Error(t, p2.Validate())

	p3 := NewProject("p3")
	p3.WithService("a", NewContainer().Image("alpine").WithContainerConfig(func(*container.Config) error {
		return errorsNew("cc")
	}))
	require.Error(t, p3.Validate())

	p4 := NewProject("p4")
	p4.WithService("a", NewContainer().Image("alpine"), 42)
	require.Error(t, p4.Validate())

	require.Error(t, p4.ForEachService(nil))
	_, err = p4.GetService("missing")
	require.Error(t, err)
	require.True(t, errdefs.IsProjectConfigError(err))

	p5 := NewProject("p5")
	p5.wrapped.Services = nil
	require.Error(t, p5.ForEachService(func(string, *types.ServiceConfig) error { return nil }))
	_, err = p5.GetService("x")
	require.Error(t, err)

	p6 := NewProject("p6")
	p6.wrapped.Volumes = nil
	p6.wrapped.Secrets = nil
	p6.wrapped.Networks = nil
	p6.WithVolume("v")
	p6.WithSecret("s")
	p6.WithNetwork("n")
	require.Contains(t, p6.wrapped.Volumes, "v")
}

func errorsNew(s string) error { return errdefs.NewContainerConfigError("x", s) }

func TestConvertersNilAndEdge(t *testing.T) {
	require.Nil(t, convertExposedPorts(nil))
	require.Nil(t, convertHealthCheck(nil))
	require.Nil(t, convertLogging(nil))
	require.Nil(t, convertLogging(&container.LogConfig{}))
	require.Empty(t, convertDevices(nil))
	require.Empty(t, convertVolumesFrom(nil))

	hc := &container.HostConfig{}
	require.Nil(t, convertBlkioConfig(hc))

	vols := convertVolumes(&container.HostConfig{
		Binds: []string{"no-colon", "src:notslash"},
	})
	require.Empty(t, vols)

	ports := convertPortsBindings(map[nat.Port][]nat.PortBinding{
		nat.Port("80/tcp"): {{HostIP: "0.0.0.0", HostPort: "8080"}},
	})
	require.Len(t, ports, 1)

	nets := convertNetworks(&network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{
		"n": {IPAMConfig: &network.EndpointIPAMConfig{IPv4Address: "10.0.0.2"}},
	}})
	require.Equal(t, "10.0.0.2", nets["n"].Ipv4Address)

	require.Equal(t, []string{"/run:rw"}, []string(convertTmpfs(map[string]string{"/run": "rw"})))
	ul := convertUlimits([]*container.Ulimit{nil, {Name: "nofile", Soft: 1, Hard: 2}})
	require.Equal(t, 1, ul["nofile"].Soft)

	hc.BlkioWeightDevice = []*blkiodev.WeightDevice{{Path: "/dev/zero", Weight: 10}}
	require.NotNil(t, convertBlkioConfig(hc))
}

func TestFirstNonEmptyAndApplySpecs(t *testing.T) {
	require.Equal(t, "a", firstNonEmpty("a", "b"))
	require.Equal(t, "b", firstNonEmpty("", "b"))

	cfg := &types.ServiceConfig{}
	BuildSpec{}.apply(cfg)
	require.NotNil(t, cfg.Build)

	Deploy{}.apply(cfg)
	require.NotNil(t, cfg.Deploy)

	var nilBuild *BuildSpec
	var nilDeploy *Deploy
	require.NoError(t, applyServiceExtra(cfg, nilBuild))
	require.NoError(t, applyServiceExtra(cfg, nilDeploy))
	require.NoError(t, applyServiceExtra(cfg, nil))
}
