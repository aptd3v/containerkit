package kit

import (
	"context"
	"testing"
)

// Compile-time proof: one Name/Image vocabulary, no ContainerName/ServiceName,
// no Name[Container] type arguments. Applicability is the option interface.
var (
	_ ContainerOption           = Name("worker")
	_ ServiceOption             = Name("worker")
	_ ComposeOption             = Name("app")
	_ ContainerOption           = Image("alpine:latest")
	_ ServiceOption             = Image("alpine:latest")
	_ UpOption                  = Detach(true)
	_ UpOption                  = Quiet(true)
	_ UpOption                  = Timeout(0)
	_ DownOption                = Volumes(true)
	_ EngineCreateOption        = Name("x")
	_ RemoveOption              = Force(true)
	_ EngineStopOption          = Timeout(0)
	_ EnginePullOption          = All(true)
	_ EngineBuildOption         = Image("app:latest")
	_ BuildOption               = (*Dockerfile)(nil)
	_ EngineBuildOption         = (*Dockerfile)(nil)
	_ DockerFile                = (*Dockerfile)(nil)
	_ EngineNetworkOption       = Driver("bridge")
	_ EngineVolumeOption        = Name("data")
	_ GenerateOption            = Name("app")
	_ LoadProjectOption         = Name("app")
	_ LoadProjectOption         = File("compose.yaml")
	_ LoadProjectOption         = AllResources(true)
	_ LoadProjectOption         = Offline(true)
	_ SecretOption              = FromFile("./secret")
	_ ComposeFileOption         = FromFile("./nginx.conf")
	_ BuildOption               = DockerfilePath("Dockerfile")
	_ EngineBuildOption         = DockerfilePath("Dockerfile")
	_ ExportOption              = OutputFile("out.tar")
	_ ClientOption              = FromEnv()
	_ ClientOption              = Host("unix:///var/run/docker.sock")
	_ ClientOption              = Timeout(0)
	_ ClientOption              = NegotiateAPIVersion()
	_ ClientOption              = TraceProvider(nil)
	_ EngineContainerListOption = All(true)
	_ EngineContainerListOption = Filter("status", "running")
	_ ServiceOption             = ExternalLink("db")
	_ ServiceOption             = Develop(WatchFiles(Path("./src"), Action(Sync)))
	_ BuildOption               = Entitlement("network.host")
	_ DeployOption              = Reserve(Memory(1), CPUs(0.1))
	_ DeployOption              = Update(Parallelism(1), Delay(0))
	_ ComposeNetOption          = EnableIPv4(true)
	_ EngineWaitOption          = WaitFor(NotRunning)
	_ TLSOption                 = CA("ca.pem")
	_ PortOption                = Protocol(TCP)
	_ PortOption                = HostPort(8080)
	_ DependOption              = Name("db")
	_ ComposeOption             = AddService(Name("api"), Image("alpine:latest"))
	_ ServiceOption             = (*Container)(nil)

	_ = func() {
		_ = NewContainer(
			Name("worker"),
			Image("alpine:latest"),
		)
		_ = NewService(
			Name("worker"),
			Image("alpine:latest"),
		)
		_ = NewCompose("app")
		_ = NewCompose(
			"app",
			AddService(
				Name("api"),
				Image("alpine:latest"),
				Publish(
					Protocol(TCP),
					HostIP("0.0.0.0"),
					HostPort(8080),
					ContainerPort(80),
				),
			),
		)
		myService := NewService(
			NewContainer(Name("foo-server"), Image("alpine:latest")),
		)
		_ = NewCompose("app").WithService(myService)
		var compose *Compose
		_ = compose.WithService
		_ = compose.Up
		_ = func(c *Compose, ctx context.Context, api Composer) error {
			return c.Up(ctx, api, Detach(true), Quiet(true), Timeout(0))
		}
		_ = func(c *Container, ctx context.Context, api Engine) (Created, error) {
			return c.Create(ctx, api, Name("x"))
		}
		_ = func() (Engine, error) {
			return NewEngine(FromEnv(), Host("unix:///var/run/docker.sock"), Timeout(0), NegotiateAPIVersion(), TraceProvider(nil))
		}
		_ = func(ctx context.Context, api Engine) (ContainerListResult, error) {
			return ListContainers(ctx, api, All(true), Filter("status", "running"), IncludeSize(true), Limit(10))
		}
		df := NewDockerfile().From("alpine:latest").Run("echo ok")
		_ = NewService(Name("web"), Build(df))
		_ = func(ctx context.Context, api Engine) error {
			return BuildImage(ctx, api, df, Image("web:dev"))
		}
	}
)

func TestInferenceProofCompilesAndRuns(t *testing.T) {
	t.Parallel()
	c := NewContainer(Name("worker"), Image("alpine:latest"))
	if c.Name() != "worker" || c.Image() != "alpine:latest" {
		t.Fatalf("container = %+v", c)
	}
	s := NewService(Name("worker"), Image("alpine:latest"))
	if s.Name() != "worker" || s.Image() != "alpine:latest" {
		t.Fatalf("service = %+v", s)
	}
	p := NewCompose("app")
	if p.Name() != "app" {
		t.Fatalf("compose = %+v", p)
	}
	myService := NewService(NewContainer(Name("foo-server"), Image("alpine:latest")))
	chained := NewCompose("app").WithService(myService)
	if chained.Name() != "app" || chained.Services()[0].Name() != "foo-server" {
		t.Fatalf("fluent = %+v", chained)
	}
}
