// Package kit is one vocabulary for Docker Engine and Compose.
//
// Three nouns:
//
//   - Define a [Container], [Service], or [Compose] file in code
//     ([NewContainer], [NewService], [NewCompose])
//   - Load compose YAML from disk ([ImportFile] locally, [Client.Load] with a daemon)
//   - Run it through one [Client] ([Connect], then [Client.Run] / [Client.Up] / [Loaded.Up])
//
// Import the kit package:
//
//	import "github.com/aptd3v/containerkit/pkg/kit"
//
//	cli, err := kit.Connect()
//	defer cli.Close()
//	err = cli.Run(ctx, kit.Name("web"), kit.Image("nginx"))
//
//	compose := kit.NewCompose("app").WithService(
//		kit.NewService(kit.Name("api"), kit.Image("alpine:latest")),
//	)
//	err = cli.Up(ctx, compose)
//
//	loaded, err := cli.Load(ctx, "compose.yaml")
//	err = loaded.Up(ctx)
//
// [Connect] is the usual constructor. [NewEngine] and [NewComposer] remain when
// a program needs only one SDK. [LoadProject] is the advanced loader [Client.Load]
// wraps.
//
// Build a compose file with [Compose.WithService], [Compose.WithNetwork],
// [Compose.WithVolume], [Compose.WithSecret], [Compose.WithFile], and
// [Compose.WithJob]. [AddService] exists so [logic.Group] can carry a service;
// it is not the public grammar.
//
// The same names apply at every scope they mean the same thing:
//
//	ReadOnly(true)                          // root filesystem
//	Volume(..., ReadOnly(true))             // one mount
//	Env("VERSION", "1")                     // container environment
//	Build(Source("."), Env("VERSION", "1")) // build-arg
//	Memory(1<<30)                           // mem_limit
//	Deploy(Memory(1<<30))                   // deploy.resources.limits.memory
//
// Meanings that only look the same have different names:
//
//   - [File] is a compose YAML path; secret/config content is [FromFile];
//     a Dockerfile path is [DockerfilePath]; an export archive is [OutputFile]
//   - [Publish] publishes a port; [Compose.PublishOCI] publishes an OCI artifact
//   - [All] includes stopped/list/prune candidates; [AllResources] is LoadProject
//   - [Log] is the progress [Logger]; container log drivers are [Logging]
//   - [WorkingDir] is a container working directory; [ProjectDirectory] is the
//     compose project directory on [LoadProject]
//
// [Test] is the healthcheck command only. Interval, timeout, and retries are
// [Interval], [Timeout], and [Retries].
//
// Constructors always return a value. [Compose.Up], [Container.Create],
// [Compose.ExportYAML], and [Compose.Write] validate before they run.
// [Container.Validate] / [Service.Validate] / [Compose.Validate] are for
// checking earlier. Common finite values ([TCP], [Always], [Bridge], [JSONFile],
// [PortHost], [Replicated]) live in this package; the long tail stays in
// pkg/cap, pkg/proto, and the other vocabulary packages.
//
// A [*Container] is a [ServiceOption], so NewService(NewContainer(Name("foo-server"), ...))
// inherits the container name, or generates one from the image.
//
//	compose := ImportYAML(yml, Name("app"), Env("TAG", "latest"))
//	local := ImportFile("compose.yaml", File("compose.override.yaml"))
//	log := NewLogger(Color(true), Progress("auto"))
//	err = compose.Up(ctx, cli.Composer(), Detach(true), Log(log), Timeout(10*time.Second))
//	created, err := container.Create(ctx, cli.Engine(), Name("worker"))
//	_, err = BuildImage(ctx, cli.Engine(), NewDockerfile().From("alpine:latest"), Image("app:dev"), Log(log))
//	_, err = PullImage(ctx, cli.Engine(), "alpine:latest", Log(log))
//
// Runtime failures are [Error] values classified onto [ErrNotFound],
// [ErrConflict], and the other SDK sentinels so [errors.Is] / [IsNotFound]
// work without importing Engine or Compose.
//
// Configuration targets the latest stable modules verified on 2026-11-12:
//
//   - github.com/moby/moby/api v1.56.1
//   - github.com/moby/moby/client v0.6.2
//   - github.com/compose-spec/compose-go/v2 v2.16.2
//   - github.com/docker/compose/v5 v5.6.1 (runtime API + in-process Composer)
//
// Engine AutoRemove, PublishAll, ContainerIDFile, Shell, OnBuild, MaskedPath,
// and ReadonlyPath have no Compose equivalent; [FromContainer] reports them.
package kit

//go:generate go run ../../tools/gencatalog
//go:generate go run ../../tools/gentsx
