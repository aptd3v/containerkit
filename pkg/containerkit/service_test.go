package containerkit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

func TestProjectMarshalFullConfig(t *testing.T) {
	mode := int64(0444)
	replicas := 2
	c := NewContainer("web-1").Image("nginx:latest").Command("nginx", "-g", "daemon off;").Env("FOO", "bar").Tty().ExposedPort("tcp", "80").HealthCheck(Health{Test: []string{"CMD", "true"}, Timeout: 1, Interval: 1, Retries: 1}).Entrypoint("/docker-entrypoint.sh").StdinOpen().StopSignal("SIGQUIT").WorkingDir("/usr/share/nginx/html").Label("app", "web").DomainName("web.test").HostName("web").User("nginx").StopTimeout(5).BlkioWeight(200).BlkioDeviceReadBps("/dev/zero", 1000*B).BlkioDeviceWriteBps("/dev/zero", 1000*B).BlkioDeviceReadIOps("/dev/zero", 10).BlkioDeviceWriteIOps("/dev/zero", 10).AddedCapabilities(NET_BIND_SERVICE).DroppedCapabilities(MKNOD).CgroupParent("system.slice").DeviceCgroupRules("c 1:3 rwm").CPUCount(1).CPUPercent(20).CPUPeriod(100*time.Millisecond).CPUQuota(50*time.Millisecond).CPUShares(256).CpusetCpus(mustCpuset("0")).CPURealtimeRuntime(0).CPURealtimePeriod(0).PidMode(PidHost).MemoryReservation(8*MiB).MemorySwap(64*MiB).MemoryLimit(32*MiB).ShmSize(MiB).DNSLookups("8.8.8.8").DNSSearches("test").DNSOptions("ndots:2").ExtraHost("host.docker.internal", "host-gateway").OomScoreAdj(50).OomKillDisable().AddedDevice(Device{HostPath: "/dev/null", ContainerPath: "/dev/null", Read: true, Write: true, Mknod: true}).AddedGroups("www-data").Init().IpcMode(IpcShareable).Isolation(IsolationDefault).LogDriver("json-file", map[string]string{"max-size": "1m"}).VolumesFrom("other", false).VolumeBinds("/tmp:/opt/bind:ro,z").RWNamedVolumeMount("data", "/var/lib/data").TmpfsMount("/tmpfs", 1024, 0755).PortBinding(PortBinding{Protocol: TCP, HostIP: "127.0.0.1", HostPort: "8080", ContainerPort: "80"}).Privileged().ReadOnlyRootfs().RestartPolicyUnlessStopped().Runtime("runc").SecurityOpts("no-new-privileges:true").Sysctls("net.ipv4.ip_forward", "0").Tmpfs("/run", "rw").Ulimits(Ulimit{Name: "nofile", Soft: 1024, Hard: 2048}).UserNSMode(UserNSHost).UTSMode(UTSHost).MemorySwappiness(40).PidsLimit(50).Architecture("amd64").Endpoint("front", EndpointOpts{Aliases: []string{"web"}, IPv4: "10.1.0.10"})

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
	require.Contains(t, string(raw), "host.docker.internal=host-gateway")
	require.NotNil(t, p.Unwrap())

	svc, err := p.GetService("web")
	require.NoError(t, err)
	require.Equal(t, "web", svc.Name)
	require.Empty(t, svc.ContainerName, "deploy set should clear container name")
	require.Equal(t, []string{"host-gateway"}, svc.ExtraHosts["host.docker.internal"])

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
	bad := NewContainer().Image("alpine")
	bad.Errors = append(bad.Errors, errorsNew("cc"))
	p3.WithService("a", bad)
	require.Error(t, p3.Validate())

	p4 := NewProject("p4")
	p4.WithService("a", NewContainer().Image("alpine"), 42)
	require.Error(t, p4.Validate())

	pHosts := NewProject("hosts")
	pHosts.WithService("web", NewContainer().Image("alpine").ExtraHost("no-ip", ""))
	require.Error(t, pHosts.Validate())

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

func TestWithNetworkOneSpec(t *testing.T) {
	ok := NewProject("n")
	ok.WithService("web", NewContainer().Image("alpine"))
	ok.WithNetwork("front", ProjectNetwork{Driver: "bridge", Internal: true})
	require.NoError(t, ok.Validate())
	require.Equal(t, "bridge", ok.Unwrap().Networks["front"].Driver)
	require.True(t, ok.Unwrap().Networks["front"].Internal)

	bad := NewProject("n")
	bad.WithService("web", NewContainer().Image("alpine"))
	bad.WithNetwork("front", ProjectNetwork{Driver: "bridge"}, ProjectNetwork{Internal: true})
	require.Error(t, bad.Validate())
}

func TestImplicitDefaultNetwork(t *testing.T) {
	t.Run("service without networks uses project default", func(t *testing.T) {
		p := NewProject("nginx-example")
		p.WithService("nginx", NewContainer().
			Image("nginx:latest").
			PortBinding(PortBinding{Protocol: TCP, HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "80"}),
		)
		require.NoError(t, p.Validate())

		svc, err := p.GetService("nginx")
		require.NoError(t, err)
		require.Contains(t, svc.Networks, "default")
		require.Contains(t, p.Unwrap().Networks, "default")
		require.Equal(t, "nginx-example_default", p.Unwrap().Networks["default"].Name)
		require.Len(t, svc.Ports, 1)
		require.Equal(t, uint32(80), svc.Ports[0].Target)
		require.Equal(t, "8080", svc.Ports[0].Published)
	})

	t.Run("network mode skips default", func(t *testing.T) {
		p := NewProject("isolated")
		p.WithService("none", NewContainer().Image("alpine").DisabledNetwork())
		require.NoError(t, p.Validate())

		svc, err := p.GetService("none")
		require.NoError(t, err)
		require.Equal(t, "none", svc.NetworkMode)
		require.NotContains(t, svc.Networks, "default")
		require.NotContains(t, p.Unwrap().Networks, "default")
	})

	t.Run("explicit networks skip implicit default", func(t *testing.T) {
		p := NewProject("custom")
		p.WithNetwork("front")
		p.WithService("web", NewContainer().Image("nginx").Endpoint("front"))
		require.NoError(t, p.Validate())

		svc, err := p.GetService("web")
		require.NoError(t, err)
		require.Contains(t, svc.Networks, "front")
		require.NotContains(t, svc.Networks, "default")
		require.NotContains(t, p.Unwrap().Networks, "default")
	})

	t.Run("existing default network is kept", func(t *testing.T) {
		p := NewProject("named")
		p.WithNetwork("default")
		p.WithService("web", NewContainer().Image("nginx"))
		require.NoError(t, p.Validate())

		require.Equal(t, "default", p.Unwrap().Networks["default"].Name)
		svc, err := p.GetService("web")
		require.NoError(t, err)
		require.Contains(t, svc.Networks, "default")
	})
}

func errorsNew(s string) error { return errdefs.NewContainerConfigError("x", s) }

func TestConvertersNilAndEdge(t *testing.T) {
	require.Nil(t, convertExposedPorts(nil))
	require.Nil(t, convertHealthCheck(nil))
	require.Nil(t, convertLogging(nil))
	require.Nil(t, convertLogging(&container.LogConfig{}))
	require.Empty(t, convertDevices(nil))
	require.Empty(t, convertVolumesFrom(nil))

	hosts, err := convertExtraHosts(nil)
	require.NoError(t, err)
	require.Nil(t, hosts)
	hosts, err = convertExtraHosts([]string{})
	require.NoError(t, err)
	require.Nil(t, hosts)
	hosts, err = convertExtraHosts([]string{"host.docker.internal:host-gateway", "foo=127.0.0.1"})
	require.NoError(t, err)
	require.Equal(t, []string{"host-gateway"}, hosts["host.docker.internal"])
	require.Equal(t, []string{"127.0.0.1"}, hosts["foo"])
	_, err = convertExtraHosts([]string{"missing-ip"})
	require.Error(t, err)

	hc := &container.HostConfig{}
	require.Nil(t, convertBlkioConfig(hc))

	vols, tmpfsFromMounts := convertVolumes(nil, &container.HostConfig{
		Binds: []string{"no-colon", "src:notslash", "data:/var/lib:rw", "/src:/dst:ro,rshared,cached", "/src:/other:rshared", "/tmp:/opt:z"},
	})
	require.Empty(t, tmpfsFromMounts)
	require.Len(t, vols, 4)
	require.Equal(t, "volume", vols[0].Type)
	require.Equal(t, "data", vols[0].Source)
	require.Equal(t, "/var/lib", vols[0].Target)
	require.False(t, vols[0].ReadOnly)
	require.Equal(t, "bind", vols[1].Type)
	require.True(t, vols[1].ReadOnly)
	require.Equal(t, "cached", vols[1].Consistency)
	require.Equal(t, "rshared", vols[1].Bind.Propagation)
	require.Equal(t, "/dst", vols[1].Target)
	require.False(t, vols[2].ReadOnly)
	require.Equal(t, "rshared", vols[2].Bind.Propagation)
	require.Equal(t, "z", vols[3].Bind.SELinux)

	ports := convertPortsBindings(map[nat.Port][]nat.PortBinding{
		nat.Port("80/tcp"): {{HostIP: "0.0.0.0", HostPort: "8080"}},
	})
	require.Len(t, ports, 1)

	nets := convertNetworks(&network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{
		"n": {
			GwPriority: 42,
			IPAMConfig: &network.EndpointIPAMConfig{IPv4Address: "10.0.0.2"},
		},
	}})
	require.Equal(t, "10.0.0.2", nets["n"].Ipv4Address)
	require.Equal(t, 42, nets["n"].GatewayPriority)
	require.Zero(t, nets["n"].Priority)

	require.Equal(t, []string{"/run:rw"}, []string(convertTmpfs(map[string]string{"/run": "rw"})))
	ul := convertUlimits([]*container.Ulimit{nil, {Name: "nofile", Soft: 1, Hard: 2}})
	require.Equal(t, 1, ul["nofile"].Soft)

	hc.BlkioWeightDevice = []*blkiodev.WeightDevice{{Path: "/dev/zero", Weight: 10}}
	require.NotNil(t, convertBlkioConfig(hc))
}

func TestMacAddressFromEndpoints(t *testing.T) {
	require.Empty(t, macAddressFromEndpoints(nil))
	require.Empty(t, macAddressFromEndpoints(&network.NetworkingConfig{}))
	require.Equal(t, "02:42:ac:11:00:02", macAddressFromEndpoints(&network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"a": {MacAddress: "02:42:ac:11:00:02"},
			"b": nil,
			"c": {},
		},
	}))
	require.Empty(t, macAddressFromEndpoints(&network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"a": {MacAddress: "02:42:ac:11:00:02"},
			"b": {MacAddress: "02:42:ac:11:00:03"},
		},
	}))
}

func TestWithServiceMapsContainerFields(t *testing.T) {
	c := NewContainer("api").Image("alpine").Volume("/cache").Volume("/data").DisabledNetwork().StorageOpt("size", "20G").VolumeDriver("local").DeviceRequest(DeviceRequest{Driver: "nvidia", Count: -1, DeviceIDs: []string{"0"}, Capabilities: []string{"gpu", "compute", "utility"}, Options: map[string]string{"opt": "1"}}).RestartPolicy(RestartPolicyOnFailure, 3).OS("linux").Architecture("arm64").Variant("v8").DisabledHealthCheck().RWNamedVolumeMountWithLabel("data", "/var/lib/data", "foo", "bar").RWNamedVolumeSubPath("data", "/var/lib/sub", "nested").NonRecursiveBindMount("/host", "/mnt", true).HostBindMountRecursiveRO("/host/ro", "/mnt/ro").TmpfsMountUIDGID("/run", 4096, "1000", "1000").Endpoint("back", EndpointOpts{DriverOptions: map[string]string{"com.docker.network.endpoint.ifname": "eth1"}, MacAddress: "02:42:ac:11:00:02"})
	c.Config.Host.NanoCPUs = 1_500_000_000
	c.Config.Host.Links = []string{"db:database"}

	p := NewProject("mapped")
	p.WithService("api", c)
	require.NoError(t, p.Validate())
	svc, err := p.GetService("api")
	require.NoError(t, err)

	require.Equal(t, float32(1.5), svc.CPUS)
	require.Equal(t, "on-failure:3", svc.Restart)
	require.Equal(t, "local", svc.VolumeDriver)
	require.Equal(t, map[string]string{"size": "20G"}, svc.StorageOpt)
	require.Equal(t, []string{"db:database"}, svc.Links)
	require.Equal(t, "none", svc.NetworkMode)
	require.Equal(t, "linux/arm64/v8", svc.Platform)
	require.True(t, svc.HealthCheck.Disable)
	require.Empty(t, svc.HealthCheck.Test)
	require.Equal(t, []types.DeviceRequest{{
		Driver:       "nvidia",
		Count:        -1,
		IDs:          []string{"0"},
		Capabilities: []string{"gpu", "compute", "utility"},
		Options:      types.Mapping{"opt": "1"},
	}}, svc.Gpus)
	require.Equal(t, "eth1", svc.Networks["back"].DriverOpts["com.docker.network.endpoint.ifname"])
	require.Equal(t, "02:42:ac:11:00:02", svc.Networks["back"].MacAddress)
	require.Equal(t, "02:42:ac:11:00:02", svc.MacAddress)

	require.Equal(t, "/run:size=4096,uid=1000,gid=1000", string(svc.Tmpfs[0]))

	byTarget := map[string]types.ServiceVolumeConfig{}
	for _, vol := range svc.Volumes {
		byTarget[vol.Target] = vol
	}
	require.Equal(t, types.VolumeTypeVolume, byTarget["/cache"].Type)
	require.Empty(t, byTarget["/cache"].Source)
	require.Equal(t, types.VolumeTypeVolume, byTarget["/data"].Type)
	require.Equal(t, "bar", byTarget["/var/lib/data"].Volume.Labels["foo"])
	require.Equal(t, "nested", byTarget["/var/lib/sub"].Volume.Subpath)
	require.True(t, byTarget["/mnt"].ReadOnly)
	require.Equal(t, "disabled", byTarget["/mnt"].Bind.Recursive)
	require.Equal(t, "readonly", byTarget["/mnt/ro"].Bind.Recursive)

	healthy := NewContainer().Image("alpine").HealthCheck(Health{
		Test: []string{"CMD", "true"}, Timeout: 1, Interval: 1, Retries: 1,
	})
	healthy.Config.Container.Healthcheck.StartInterval = 2 * time.Second
	p.WithService("healthy", healthy)
	got, err := p.GetService("healthy")
	require.NoError(t, err)
	require.NotNil(t, got.HealthCheck.StartInterval)
	require.Equal(t, types.Duration(2*time.Second), *got.HealthCheck.StartInterval)

	imageVols, extra := convertVolumes(nil, &container.HostConfig{
		Mounts: []mount.Mount{{
			Type:         mount.TypeImage,
			Source:       "alpine",
			Target:       "/img",
			ImageOptions: &mount.ImageOptions{Subpath: "bin"},
		}, {
			Type:     mount.TypeBind,
			Source:   "/src",
			Target:   "/dst",
			ReadOnly: true,
			BindOptions: &mount.BindOptions{
				ReadOnlyNonRecursive: true,
			},
		}},
	})
	require.Empty(t, extra)
	require.Equal(t, "bin", imageVols[0].Image.SubPath)
	require.Equal(t, "writable", imageVols[1].Bind.Recursive)
}

func TestFirstNonEmptyAndApplySpecs(t *testing.T) {
	require.Equal(t, "a", firstNonEmpty("a", "b"))
	require.Equal(t, "b", firstNonEmpty("", "b"))

	cfg := &types.ServiceConfig{}
	require.NoError(t, BuildSpec{}.apply(cfg))
	require.NotNil(t, cfg.Build)

	Deploy{}.apply(cfg)
	require.NotNil(t, cfg.Deploy)

	var nilBuild *BuildSpec
	var nilDeploy *Deploy
	require.NoError(t, applyServiceExtra(cfg, nilBuild))
	require.NoError(t, applyServiceExtra(cfg, nilDeploy))
	require.NoError(t, applyServiceExtra(cfg, nil))

	mode := int64(0644)
	manual := true
	p := NewProject("spec")
	p.WithModel("chat", Model{Model: "ai/llama", ContextSize: 4096, RuntimeFlags: []string{"--verbose"}})
	p.WithJob("nightly", NewContainer().Image("alpine").Command("true"), Job{
		Profiles: []string{"batch"},
		Manual:   &manual,
		Schedule: []JobSchedule{{Cron: "0 2 * * *", Timezone: "UTC"}},
	})
	p.WithService("api", NewContainer().Image("alpine"),
		PreStart(NewContainer("init").Image("alpine").Command("true"), true),
		ServiceModel("chat", "MODEL_URL", "MODEL_NAME"),
		BuildSpec{
			SSH:                []BuildSSH{{ID: "default", Path: "/tmp/sock"}},
			AdditionalContexts: map[string]string{"src": "https://example.test"},
			ExtraHosts:         []string{"db:10.0.0.1"},
			Isolation:          "default",
			ShmSize:            1024,
			Entitlements:       []string{"network.host"},
			Ulimits:            map[string]BuildUlimit{"nofile": {Soft: 10, Hard: 20}},
			Secrets:            []BuildSecret{{Source: "token", Target: "token", Mode: &mode}},
			Provenance:         "mode=min",
			SBOM:               "true",
			NoCacheFilter:      []string{"apk"},
		},
	)
	require.NoError(t, p.Validate())
	raw, err := p.Marshal()
	require.NoError(t, err)
	text := string(raw)
	require.Contains(t, text, "ai/llama")
	require.Contains(t, text, "nightly")
	require.Contains(t, text, "0 2 * * *")
	require.Contains(t, text, "per_replica: true")
	require.Contains(t, text, "endpoint_var: MODEL_URL")
	require.Contains(t, text, "provenance: mode=min")
	require.Contains(t, text, "no_cache_filter")
}
