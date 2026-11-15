package kit

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	clienttypes "github.com/aptd3v/containerkit/internal/client/types"
)

func TestUnifiedVolumeAndNetworkMapToBothSDKs(t *testing.T) {
	t.Parallel()
	opts := []ContainerOption{
		Name("web"),
		Image("nginx:latest"),
		Entrypoint("/docker-entrypoint.sh"),
		DomainName("local"),
		StopSignal("SIGTERM"),
		DNS("8.8.8.8"),
		ExtraHost("db", "10.0.0.2"),
		Volume(Kind(Bind), Source("/host/data"), Target("/data"), ReadOnly(true)),
		Volume(Name("cache"), Target("/cache")),
		Network(Name("frontend"), Alias("www"), IPv4("10.0.0.10")),
		NetworkMode("bridge"),
		CPUs(1.5),
		CPUShares(512),
		ShmSize(64 << 20),
		Logging(Driver("json-file"), DriverOpt("max-size", "10m")),
		Ulimit(Name("nofile"), Soft(1024), Hard(2048)),
		Device(Source("/dev/null"), Target("/dev/null"), Perms("rwm")),
		Sysctl("net.core.somaxconn", "1024"),
		SecurityOpt("no-new-privileges:true"),
		GroupAdd("audio"),
		IPC("shareable"),
		Expose(53, UDP),
	}
	c := NewContainer(opts...)
	cfg, host, netcfg, err := c.EngineConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Entrypoint[0] != "/docker-entrypoint.sh" || cfg.Domainname != "local" {
		t.Fatalf("config identity = %+v", cfg)
	}
	if len(host.Mounts) != 2 || !host.Mounts[0].ReadOnly {
		t.Fatalf("mounts = %+v", host.Mounts)
	}
	if host.NanoCPUs != 1.5e9 || host.CPUShares != 512 {
		t.Fatalf("cpu = %d %d", host.NanoCPUs, host.CPUShares)
	}
	if host.LogConfig.Type != "json-file" {
		t.Fatalf("log = %+v", host.LogConfig)
	}
	if netcfg.EndpointsConfig["frontend"] == nil {
		t.Fatalf("endpoints = %+v", netcfg.EndpointsConfig)
	}
	if host.NetworkMode != "bridge" {
		t.Fatalf("network mode = %s", host.NetworkMode)
	}

	s := NewService(
		Name("web"),
		Image("nginx:latest"),
		Entrypoint("/docker-entrypoint.sh"),
		DomainName("local"),
		StopSignal("SIGTERM"),
		DNS("8.8.8.8"),
		ExtraHost("db", "10.0.0.2"),
		Volume(Kind(Bind), Source("/host/data"), Target("/data"), ReadOnly(true)),
		Volume(Name("cache"), Target("/cache")),
		Network(Name("frontend"), Alias("www"), IPv4("10.0.0.10")),
		NetworkMode("bridge"),
		CPUs(1.5),
		CPUShares(512),
		ShmSize(64<<20),
		Logging(Driver("json-file"), DriverOpt("max-size", "10m")),
		Ulimit(Name("nofile"), Soft(1024), Hard(2048)),
		Device(Source("/dev/null"), Target("/dev/null"), Perms("rwm")),
		Sysctl("net.core.somaxconn", "1024"),
		SecurityOpt("no-new-privileges:true"),
		GroupAdd("audio"),
		IPC("shareable"),
		Expose(53, UDP),
		Secret(Name("dbpass"), Target("/run/secrets/dbpass")),
	)
	sc, err := s.ComposeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if sc.Entrypoint[0] != "/docker-entrypoint.sh" {
		t.Fatalf("compose entrypoint = %v", sc.Entrypoint)
	}
	if len(sc.Volumes) != 2 || !sc.Volumes[0].ReadOnly || sc.Volumes[0].Target != "/data" {
		t.Fatalf("compose volumes = %+v", sc.Volumes)
	}
	if sc.Networks["frontend"] == nil || sc.Networks["frontend"].Ipv4Address != "10.0.0.10" {
		t.Fatalf("compose networks = %+v", sc.Networks)
	}
	if sc.CPUS != 1.5 {
		t.Fatalf("compose cpus = %v", sc.CPUS)
	}
	if len(sc.Secrets) != 1 || sc.Secrets[0].Source != "dbpass" {
		t.Fatalf("secrets = %+v", sc.Secrets)
	}
}

func TestComposeResourcesAndBuild(t *testing.T) {
	t.Parallel()
	p := NewCompose(
		"shop",
		AddNetwork(Name("front"), Driver("bridge"), Internal(false), Subnet("10.0.0.0/24")),
		AddVolume(Name("dbdata"), Driver("local")),
		AddSecret(Name("dbpass"), FromFile("./dbpass.txt")),
		AddFile(Name("nginx"), FromFile("./nginx.conf")),
		AddService(
			Name("api"),
			Build(Source("."), DockerfilePath("Dockerfile"), Target("prod"), Env("VERSION", "1"), Privileged(false)),
			PullPolicy(PullMissing),
			Deploy(Replicas(2), Memory(1<<30), CPUs(0.5)),
			Volume(Name("dbdata"), Target("/var/lib/db")),
			Network(Name("front"), Alias("api")),
			Profile("full"),
		),
		AddJob(
			Name("migrate"),
			Image("migrate:latest"),
			Command("up"),
		),
	)
	yml, err := p.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	text := string(yml)
	for _, want := range []string{
		"name: shop",
		"api:",
		"migrate:",
		"context: .",
		"dockerfile: Dockerfile",
		"target: prod",
		"dbdata:",
		"front:",
		"dbpass:",
		"nginx:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("export missing %q in:\n%s", want, text)
		}
	}
}

func TestReadOnlyVocabularyAtBothScopes(t *testing.T) {
	t.Parallel()
	c := NewContainer(
		Image("alpine:latest"),
		ReadOnly(true),
		Volume(Source("/etc/hosts"), Target("/mnt/hosts"), Kind(Bind), ReadOnly(true)),
	)
	_, host, _, err := c.EngineConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !host.ReadonlyRootfs {
		t.Fatal("rootfs should be read-only")
	}
	if len(host.Mounts) != 1 || !host.Mounts[0].ReadOnly {
		t.Fatalf("mount readonly = %+v", host.Mounts)
	}
}

type fakeEngine struct {
	created       clienttypes.ContainerCreateOptions
	started       string
	stopped       clienttypes.ContainerStopOptions
	removed       clienttypes.ContainerRemoveOptions
	pulled        clienttypes.ImagePullOptions
	built         clienttypes.ImageBuildOptions
	network       clienttypes.NetworkCreateOptions
	volume        clienttypes.VolumeCreateOptions
	netName       string
	pullRef       string
	listed        ContainerListOptions
	pullResp      ImagePullResponse
	pullErr       error
	onPull        func()
	buildBody     io.ReadCloser
	buildErr      error
	onBuild       func()
	pluginCreate  PluginCreateOptions
	pluginInstall PluginInstallOptions
	pluginEnable  PluginEnableOptions
	pluginUpgrade PluginUpgradeOptions
	pluginSet     PluginSetOptions
	pluginList    PluginListOptions
	pluginRemove  PluginRemoveOptions
	pluginDisable PluginDisableOptions
	pluginPush    PluginPushOptions
	ping          PingOptions
	diskUsage     DiskUsageOptions
	login         RegistryLoginOptions
	events        EventsListOptions
	clientVer     string
	daemonHost    string
	serverVer     ServerVersionResult
	serviceCreate ServiceCreateOptions
	serviceUpdate ServiceUpdateOptions
	secretCreate  SecretCreateOptions
	configCreate  ConfigCreateOptions
	nodeUpdate    NodeUpdateOptions
	volumeUpdate  VolumeUpdateOptions
}

func (f *fakeEngine) ContainerCreate(_ context.Context, options clienttypes.ContainerCreateOptions) (clienttypes.ContainerCreateResult, error) {
	f.created = options
	return clienttypes.ContainerCreateResult{ID: "abc"}, nil
}
func (f *fakeEngine) ContainerStart(_ context.Context, id string, _ clienttypes.ContainerStartOptions) (clienttypes.ContainerStartResult, error) {
	f.started = id
	return clienttypes.ContainerStartResult{}, nil
}
func (f *fakeEngine) ContainerStop(_ context.Context, _ string, options clienttypes.ContainerStopOptions) (clienttypes.ContainerStopResult, error) {
	f.stopped = options
	return clienttypes.ContainerStopResult{}, nil
}
func (f *fakeEngine) ContainerRemove(_ context.Context, _ string, options clienttypes.ContainerRemoveOptions) (clienttypes.ContainerRemoveResult, error) {
	f.removed = options
	return clienttypes.ContainerRemoveResult{}, nil
}
func (f *fakeEngine) ContainerRestart(context.Context, string, clienttypes.ContainerRestartOptions) (clienttypes.ContainerRestartResult, error) {
	return clienttypes.ContainerRestartResult{}, nil
}
func (f *fakeEngine) ContainerKill(context.Context, string, clienttypes.ContainerKillOptions) (clienttypes.ContainerKillResult, error) {
	return clienttypes.ContainerKillResult{}, nil
}
func (f *fakeEngine) ContainerPause(context.Context, string, clienttypes.ContainerPauseOptions) (clienttypes.ContainerPauseResult, error) {
	return clienttypes.ContainerPauseResult{}, nil
}
func (f *fakeEngine) ContainerUnpause(context.Context, string, clienttypes.ContainerUnpauseOptions) (clienttypes.ContainerUnpauseResult, error) {
	return clienttypes.ContainerUnpauseResult{}, nil
}
func (f *fakeEngine) ContainerInspect(context.Context, string, clienttypes.ContainerInspectOptions) (clienttypes.ContainerInspectResult, error) {
	return clienttypes.ContainerInspectResult{}, nil
}
func (f *fakeEngine) ContainerLogs(context.Context, string, clienttypes.ContainerLogsOptions) (clienttypes.ContainerLogsResult, error) {
	return nil, nil
}
func (f *fakeEngine) ContainerWait(context.Context, string, clienttypes.ContainerWaitOptions) clienttypes.ContainerWaitResult {
	var r clienttypes.ContainerWaitResult
	return r
}
func (f *fakeEngine) ImagePull(_ context.Context, ref string, options clienttypes.ImagePullOptions) (clienttypes.ImagePullResponse, error) {
	f.pullRef = ref
	f.pulled = options
	if f.onPull != nil {
		f.onPull()
	}
	if f.pullErr != nil {
		return nil, f.pullErr
	}
	return f.pullResp, nil
}
func (f *fakeEngine) ImageBuild(_ context.Context, _ io.Reader, options clienttypes.ImageBuildOptions) (clienttypes.ImageBuildResult, error) {
	f.built = options
	if f.onBuild != nil {
		f.onBuild()
	}
	if f.buildErr != nil {
		return clienttypes.ImageBuildResult{}, f.buildErr
	}
	return clienttypes.ImageBuildResult{Body: f.buildBody}, nil
}
func (f *fakeEngine) NetworkCreate(_ context.Context, name string, options clienttypes.NetworkCreateOptions) (clienttypes.NetworkCreateResult, error) {
	f.netName = name
	f.network = options
	return clienttypes.NetworkCreateResult{}, nil
}
func (f *fakeEngine) VolumeCreate(_ context.Context, options clienttypes.VolumeCreateOptions) (clienttypes.VolumeCreateResult, error) {
	f.volume = options
	return clienttypes.VolumeCreateResult{}, nil
}
func (f *fakeEngine) Close() error { return nil }

var _ Engine = (*fakeEngine)(nil)

func TestContainerCreateAndStartRuntime(t *testing.T) {
	t.Parallel()
	c := NewContainer(Name("worker"), Image("alpine:latest"))
	api := &fakeEngine{}
	created, err := c.Create(context.Background(), api)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "abc" || api.created.Name != "worker" || api.created.Config.Image != "alpine:latest" {
		t.Fatalf("created = %+v cfg=%+v", created, api.created)
	}
	if err := c.Start(context.Background(), api, created.ID); err != nil {
		t.Fatal(err)
	}
	if api.started != "abc" {
		t.Fatalf("started = %s", api.started)
	}
}

func TestComposeUpDownRuntime(t *testing.T) {
	t.Parallel()
	p := NewCompose("app", AddService(Name("api"), Image("alpine:latest")))
	api := &fakeComposer{}
	if err := p.Up(context.Background(), api, Detach(true), Quiet(true)); err != nil {
		t.Fatal(err)
	}
	if api.up == nil || api.up.Name != "app" {
		t.Fatalf("up compose = %+v", api.up)
	}
	if err := p.Down(context.Background(), api, Volumes(true), RemoveOrphans(true)); err != nil {
		t.Fatal(err)
	}
	if api.down != "app" {
		t.Fatalf("down = %s", api.down)
	}
}

func TestListContainersFluentOptions(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	if _, err := ListContainers(context.Background(), api, All(true), Filter("status", "running"), IncludeSize(true), Limit(5)); err != nil {
		t.Fatal(err)
	}
	if !api.listed.All || !api.listed.Size || api.listed.Limit != 5 {
		t.Fatalf("list opts = %+v", api.listed)
	}
	if !api.listed.Filters["status"]["running"] {
		t.Fatalf("filters = %+v", api.listed.Filters)
	}
}

func TestComposeRemainingFieldsExport(t *testing.T) {
	t.Parallel()
	p := NewCompose(
		"full",
		AddNetwork(Name("net"), Driver("bridge"), EnableIPv4(true), EnableIPv6(false)),
		AddModel(Name("llm"), Image("ai/mistral"), ContextSize(4096), RuntimeFlag("--verbose")),
		AddService(
			Name("api"),
			Image("alpine:latest"),
			CPURealtimePeriod(1000),
			CPURealtimeRuntime(500),
			PullRefreshAfter("24h"),
			ExternalLink("legacy:db"),
			Model(Name("llm"), EndpointVar("MODEL_URL"), ModelVar("MODEL")),
			Develop(WatchFiles(Path("./src"), Action(Sync), Target("/app"), Ignore("*.tmp"))),
			Build(
				Source("."),
				DockerfilePath("Dockerfile"),
				Entitlement("network.host"),
				Provenance("mode=min"),
				SBOM("mode=min"),
				SSH("default", ""),
				CacheFrom("type=registry"),
				CacheTo("type=inline"),
				NoCacheFilter("mod"),
				AdditionalContext("src", "./src"),
				Secret(Name("npm"), Target("/run/secrets/npm")),
				Ulimit(Name("nofile"), Soft(1024), Hard(2048)),
				Platform("linux/amd64"),
				Isolation("default"),
				ExtraHost("db", "10.0.0.2"),
				NetworkMode("host"),
			),
			Deploy(
				Replicas(2),
				Memory(1<<30),
				CPUs(0.5),
				Reserve(Memory(64<<20), CPUs(0.1)),
				Update(Parallelism(1), Delay(0), FailureAction("pause"), Order("start-first")),
				Rollback(Parallelism(1), FailureAction("continue")),
				Restart(Policy(OnFailure), MaxRetries(3), Delay(0), Window(0)),
				Placement(Constraint("node.role==worker"), Spread("node.labels.zone"), MaxReplicas(2)),
				EndpointMode("vip"),
			),
		),
	)
	yml, err := p.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	text := string(yml)
	for _, want := range []string{
		"cpu_rt_period: 1000",
		"cpu_rt_runtime: 500",
		"pull_refresh_after: 24h",
		"legacy:db",
		"llm:",
		"endpoint_var: MODEL_URL",
		"./src",
		"action: sync",
		"network.host",
		"mode=min",
		"cache_from:",
		"additional_contexts:",
		"nofile:",
		"linux/amd64",
		"reservations:",
		"update_config:",
		"rollback_config:",
		"restart_policy:",
		"on-failure",
		"node.role==worker",
		"endpoint_mode: vip",
		"enable_ipv4: true",
		"enable_ipv6: false",
		"ai/mistral",
		"context_size: 4096",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("export missing %q in:\n%s", want, text)
		}
	}
	svc, err := p.Services()[0].ComposeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if svc.CPURTPeriod != 1000 || svc.CPURTRuntime != 500 {
		t.Fatalf("cpu rt = %d %d", svc.CPURTPeriod, svc.CPURTRuntime)
	}
	if svc.Models["llm"] == nil || svc.Models["llm"].EndpointVariable != "MODEL_URL" {
		t.Fatalf("models = %+v", svc.Models)
	}
	if svc.Develop == nil || len(svc.Develop.Watch) != 1 || svc.Develop.Watch[0].Path != "./src" {
		t.Fatalf("develop = %+v", svc.Develop)
	}
	if svc.Build == nil || len(svc.Build.Entitlements) != 1 || svc.Build.Provenance == "" || len(svc.Build.SSH) != 1 {
		t.Fatalf("build = %+v", svc.Build)
	}
	if svc.Deploy == nil || svc.Deploy.Resources.Reservations == nil || svc.Deploy.UpdateConfig == nil || svc.Deploy.RestartPolicy == nil {
		t.Fatalf("deploy = %+v", svc.Deploy)
	}
}

func TestComposeFileDocsExport(t *testing.T) {
	t.Parallel()
	p := NewCompose(
		"docs",
		AddInclude(Path("commons.yml"), ProjectDirectory("."), EnvFile(".env")),
		AddNetwork(
			Name("appnet"),
			Driver("bridge"),
			Subnet("172.16.0.0/24"),
			Gateway("172.16.0.1"),
			IPRange("172.16.0.128/25"),
			AuxAddress("host1", "172.16.0.5"),
			IPAMDriver("default"),
			IPAMOpt("foo", "bar"),
		),
		AddSecret(Name("token"), FromFile("./token"), TemplateDriver("golang")),
		AddService(
			Name("api"),
			Image("alpine:latest"),
			TmpfsMount("/run"),
			PullPolicy(PullDaily),
			EnvFile(".env", Required(false), Format("raw")),
			Extends("base"),
			ExtendsFile("common.yml"),
			Provider("aws"),
			ProviderOption("region", "us-east-1"),
			CredentialConfig("mycred"),
			PostStart(Command("echo", "up"), User("root"), Privileged(true), WorkingDir("/tmp"), Env("HOOK", "1")),
			PreStop(Command("echo", "down")),
			PreStart(Image("alpine:latest"), Command("true"), PerReplica(true)),
			DependsOn(Name("db"), Required(false)),
			Health(Test("CMD", "true"), StartInterval(5*time.Second), StartPeriod(10*time.Second)),
			Publish(Name("http"), ContainerPort(80), AppProtocol("http"), Mode("host"), Published("8080-8081")),
			Network(Name("appnet"), LinkLocal("169.254.0.1"), Priority(100), GatewayPriority(10), InterfaceName("eth1"), UseServiceNameAlias(false)),
			Volume(Source("./data"), Target("/data"), Kind(Bind), SELinux("z"), CreateHostPath(false), Recursive("disabled")),
			Volume(Target("/tmpfs"), Kind(Tmpfs), Size(64<<20), FileMode(0o755)),
			Secret(Name("token"), Target("/run/secrets/token"), UID("1000"), GID("1000"), FileMode(0o400)),
			Develop(WatchFiles(
				Path("./src"),
				Action(ExecSync),
				Target("/app"),
				InitialSync(true),
				Command("go", "test"),
				User("root"),
				Privileged(true),
				WorkingDir("/src"),
				Env("X", "1"),
			)),
			Deploy(
				Mode("replicated-job"),
				PidsLimit(100),
				GPU(Capabilities("gpu"), Count(1)),
				Discrete("SSD", 3),
				Reserve(PidsLimit(10), GPU(Capabilities("gpu"), Count(1))),
			),
		),
		AddService(Name("db"), Image("postgres:16")),
	)
	yml, err := p.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	text := string(yml)
	for _, want := range []string{
		"include:",
		"commons.yml",
		"project_directory:",
		"start_interval:",
		"app_protocol: http",
		"mode: host",
		"8080-8081",
		"link_local_ips:",
		"gw_priority: 10",
		"interface_name: eth1",
		"use_service_name_alias: false",
		"ip_range:",
		"aux_addresses:",
		"tmpfs:",
		"/run",
		"pull_policy: daily",
		"format: raw",
		"required: false",
		"extends:",
		"file: common.yml",
		"provider:",
		"region:",
		"credential_spec:",
		"config: mycred",
		"post_start:",
		"pre_stop:",
		"pre_start:",
		"per_replica: true",
		"initial_sync: true",
		"selinux: z",
		"create_host_path: false",
		"recursive: disabled",
		"template_driver: golang",
		"replicated-job",
		"pids: 100",
		"discrete_resource_spec:",
		"SSD",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("export missing %q in:\n%s", want, text)
		}
	}
	svc, err := p.Services()[0].ComposeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if svc.HealthCheck == nil || svc.HealthCheck.StartInterval == nil {
		t.Fatalf("health start_interval = %+v", svc.HealthCheck)
	}
	if len(svc.Ports) != 1 || svc.Ports[0].Name != "http" || svc.Ports[0].AppProtocol != "http" || svc.Ports[0].Mode != "host" {
		t.Fatalf("ports = %+v", svc.Ports)
	}
	if len(svc.Tmpfs) != 1 || svc.Tmpfs[0] != "/run" {
		t.Fatalf("tmpfs = %+v", svc.Tmpfs)
	}
	net := svc.Networks["appnet"]
	if net == nil || len(net.LinkLocalIPs) != 1 || net.GatewayPriority != 10 || net.UseServiceNameAlias == nil || *net.UseServiceNameAlias {
		t.Fatalf("network = %+v", net)
	}
	if svc.Extends == nil || svc.Extends.File != "common.yml" {
		t.Fatalf("extends = %+v", svc.Extends)
	}
	if svc.Provider == nil || svc.Provider.Options["region"][0] != "us-east-1" {
		t.Fatalf("provider = %+v", svc.Provider)
	}
	if svc.CredentialSpec == nil || svc.CredentialSpec.Config != "mycred" {
		t.Fatalf("credential_spec = %+v", svc.CredentialSpec)
	}
	if len(svc.PostStart) != 1 || svc.PostStart[0].User != "root" || !svc.PostStart[0].Privileged {
		t.Fatalf("post_start = %+v", svc.PostStart)
	}
	if svc.Develop == nil || !svc.Develop.Watch[0].InitialSync || len(svc.Develop.Watch[0].Exec.Command) == 0 {
		t.Fatalf("develop = %+v", svc.Develop)
	}
	if svc.Deploy == nil || svc.Deploy.Resources.Limits == nil || svc.Deploy.Resources.Limits.Pids != 100 {
		t.Fatalf("deploy limits = %+v", svc.Deploy)
	}
	if svc.Deploy.Resources.Reservations == nil || svc.Deploy.Resources.Reservations.Pids != 10 {
		t.Fatalf("deploy reservations = %+v", svc.Deploy.Resources.Reservations)
	}
	if len(svc.EnvFiles) != 1 || svc.EnvFiles[0].Format != "raw" || bool(svc.EnvFiles[0].Required) {
		t.Fatalf("env_files = %+v", svc.EnvFiles)
	}
	dep := svc.DependsOn["db"]
	if dep.Required {
		t.Fatalf("depends_on required = %+v", dep)
	}
	if len(svc.Secrets) != 1 || svc.Secrets[0].UID != "1000" || svc.Secrets[0].Mode == nil {
		t.Fatalf("secrets = %+v", svc.Secrets)
	}
	if len(svc.PreStart) != 1 || !svc.PreStart[0].PerReplica {
		t.Fatalf("pre_start = %+v", svc.PreStart)
	}
}
