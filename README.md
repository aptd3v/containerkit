# containerkit

[![Go Reference](https://pkg.go.dev/badge/github.com/aptd3v/containerkit.svg)](https://pkg.go.dev/github.com/aptd3v/containerkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/aptd3v/containerkit)](https://goreportcard.com/report/github.com/aptd3v/containerkit)
[![Go Version](https://img.shields.io/github/go-mod/go-version/aptd3v/containerkit)](https://go.dev/dl/)
[![License](https://img.shields.io/github/license/aptd3v/containerkit)](./LICENSE)

Define Docker stacks in Go. The same `*containerkit.Container` is a Compose service or an Engine create config. Compose runs in-process through `github.com/docker/compose/v5` and does not need a `docker compose` binary.

## Demo

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func main() {
	project := containerkit.NewProject("hello-world")
	project.WithService("hello",
		containerkit.NewContainer().
			Image("alpine:latest").
			Command("echo", "Hello from containerkit!"),
	)

	if err := project.Validate(); err != nil {
		log.Fatal(err)
	}

	app := containerkit.NewCompose(project)
	if err := app.Up(context.Background(), &containerkit.Up{Writer: os.Stdout}); err != nil {
		log.Fatal(err)
	}
}
```

The `echo` container exits when the command finishes.

## Getting started

```bash
go get github.com/aptd3v/containerkit@latest
```

- **Go** 1.26.8+
- **Docker Engine**
- Linux, macOS, or Windows

## Features

### Two ways to run the same container

| Piece | Import | Role |
|---|---|---|
| `Project`, `*Container` | `github.com/aptd3v/containerkit/pkg/containerkit` | Stack and container config. Fluent methods set image, env, ports, mounts, and the rest. |
| `NewCompose`, `Up`, `Down`, … | same package | In-process Compose runner. Each command takes an option struct. |
| Engine client | `github.com/aptd3v/containerkit/pkg/client` | Docker Engine SDK wrapper. Each call takes a per-op struct (`Remove`, `Exec`, `ImageBuild`, …). |

Compose fits a multi-service stack: networks, volumes, profiles, and `Up` / `Down` / `Logs`. The Engine client fits one container, image builds, and an interactive TTY. A hybrid program uses the client to build or exec, then Compose to run the stack. See [examples/nginx](./examples/nginx), [examples/image_inline](./examples/image_inline), and [examples/mongo_replica](./examples/mongo_replica).

### Compose stack

```go
project := containerkit.NewProject("my-app")
project.
	WithNetwork("backend").
	WithVolume("data").
	WithService("api",
		containerkit.NewContainer().
			Image("alpine:latest").
			Command("sleep", "inf").
			PortBinding(containerkit.PortBinding{Protocol: containerkit.TCP, HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "8080"}).
			RWNamedVolumeMount("data", "/data").
			Endpoint("backend"),
		containerkit.DependsOn("db"),
	)

if err := project.Validate(); err != nil {
	log.Fatal(err)
}

app := containerkit.NewCompose(project)
if err := app.Up(ctx, &containerkit.Up{
	Detach:        true,
	RemoveOrphans: true,
	Writer:        os.Stdout,
}); err != nil {
	log.Fatal(err)
}
defer app.Down(ctx, &containerkit.Down{RemoveVolumes: true, Writer: os.Stdout})
```

`Export` writes the project as Compose YAML:

```go
if err := project.Export("./docker-compose.yaml", 0644); err != nil {
	log.Fatal(err)
}
```

### Engine client

`NewContainer` joins every name part with a hyphen (`"my", "api"` becomes `my-api`). Pass a name for `ContainerCreate`. A Compose service can omit it. Construction options stay funcs on `client`.

```go
cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
if err != nil {
	log.Fatal(err)
}

ctr := containerkit.NewContainer("my-api")
ctr.Image("alpine:latest").Command("sleep", "inf")

resp, err := cli.ContainerCreate(ctx, ctr)
if err != nil {
	log.Fatal(err)
}
_ = cli.ContainerRemove(ctx, resp.ID, &client.Remove{Force: true, Volumes: true})
```

Pass that same `*containerkit.Container` to `WithService` when the stack should run under Compose.

### Dockerfile

`NewDockerFile` builds a Dockerfile in Go. `Export` writes the file. `WithInline` returns a `BuildSpec` for `WithService`.

```go
df := containerkit.NewDockerFile().
	FromAs("alpine:latest", "build").
	Run("echo ok > /out/app").
	From("alpine", "latest").
	Copy("/out/app", "/app", containerkit.CopyOpts{From: "build"})

if err := df.Export("Dockerfile", 0644); err != nil {
	log.Fatal(err)
}

project.WithService("app",
	containerkit.NewContainer().Image("my-image:tag"),
	df.WithInline(),
)
```

`NewLocalBuildContext` packs that Dockerfile into a tar for `client.ImageBuild` and honors `.dockerignore`. See [examples/dockerfile](./examples/dockerfile) and [examples/image_inline](./examples/image_inline).

### Rules

- Setters append to `Container.Errors` and to the project. Call `Validate` before `Up`, `Export`, or `ContainerCreate`. Those three calls run `Validate` themselves.
- A nil `Up` or `Down` uses defaults. The writer is `os.Stdout`.
- `Profiles` on a service is Compose’s `profiles:` key. If the project has any profiled service, set `Profiles` on the command (`Up.Profiles`, `Down.Profiles`, and the rest). Compose returns an error when that list is empty.

```go
project.WithService("worker",
	containerkit.NewContainer().Image("alpine:latest"),
	containerkit.Profiles("full"),
)
_ = app.Up(ctx, &containerkit.Up{Profiles: []string{"full"}, Writer: os.Stdout})
```

- Pass `DependsOn`, `Profiles`, and `BuildSpec` as extra arguments to `WithService`. They configure the Compose service.
- Sizes are a `ByteSize`: `MemoryLimit(100 * containerkit.MiB)`.
- Branch with an ordinary `if`.

```go
c := containerkit.NewContainer().Image("node:latest")
if runtime.GOOS == "linux" {
	c.RWNamedVolumeMount("data", "/app")
} else {
	c.VolumeBinds("./:/app/:rw")
}
```

### API surface

Full signatures are on [pkg.go.dev](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/containerkit).

**Container.** Image, env, command, ports, mounts, health, restart, resources, capabilities, logging, and platform. You can also mix in Docker SDK structs. See [examples/structs](./examples/structs).

**Project.** `WithService`, `WithNetwork`, `WithVolume`, `WithSecret`, `WithJob`, and `WithModel`.

**Compose commands.**

| Group | Commands |
|---|---|
| Lifecycle | `Up`, `Down`, `Start`, `Stop`, `Restart`, `Kill`, `Rm`, `Run`, `Scale` |
| Observe | `Logs`, `Ps`, `Events`, `Top`, `Stats`, `Wait` |
| Files and build | `Build`, `Pull`, `Push`, `Config`, `Export`, `Watch` |

**Engine client.** Containers, images, networks, volumes, swarm, and plugins, plus interactive TTY in `pkg/client/terminal`. See [pkg.go.dev/pkg/client](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/client).

### Examples

Run from the repo root.

| Path | Path type | What it shows |
|---|---|---|
| [examples/simple](./examples/simple) | Compose | Minimal `NewProject` → `Up` |
| [examples/wordpress](./examples/wordpress) | Compose | Multi-service, OS-conditional mounts, logs, signal `Down` |
| [examples/profiles](./examples/profiles) | Compose | `Profiles` mirrored on `Up` / `Kill` |
| [examples/events](./examples/events) | Compose | Events channel + healthcheck |
| [examples/compose_commands](./examples/compose_commands) | Compose | `Build`, `Ps`, `Start`, `Stop`, `Restart`, `Exec`, `Config`, `Images`, `Top`, `Pause`, `Run` |
| [examples/nginx](./examples/nginx) | Hybrid | `client.ImageBuild` then compose `Up` |
| [examples/image_inline](./examples/image_inline) | Hybrid | Inline Dockerfile; Engine inspect skips a rebuild when the tag exists |
| [examples/dockerfile](./examples/dockerfile) | Client | Multi-stage `NewDockerFile`, `Export`, `ImageBuild`, `ContainerCreate` |
| [examples/mongo_replica](./examples/mongo_replica) | Hybrid | Compose `Up` with `Wait`, then client exec to init a replica set |
| [examples/terminal](./examples/terminal) | Client | Interactive TTY exec + resize |
| [examples/structs](./examples/structs) | Config | Mix Docker SDK structs with fluent methods |
| [examples/supabase](./examples/supabase) | Compose | Full multi-profile stack. See its [README](./examples/supabase/README.md) |

```bash
go run ./examples/simple/
go run ./examples/supabase/                    # minimal
go run ./examples/supabase/ -profile full
go run ./examples/supabase/ -resource-limits
```

## Migrating from go-contain

This project was renamed from `go-contain`.

| Old | New |
|---|---|
| `github.com/aptd3v/go-contain` | `github.com/aptd3v/containerkit` |
| `pkg/create`, `cc.WithImage`, … | `pkg/containerkit`, fluent `.Image()`, … |
| `pkg/compose`, `up.WithRemoveOrphans()` | `containerkit.NewCompose` + `&containerkit.Up{…}` |
| `pkg/client/options/**` | per-op structs on `pkg/client` (`Remove`, `Exec`, …) |

## Contributing

Issues and pull requests are welcome: [issues](https://github.com/aptd3v/containerkit/issues) · [discussions](https://github.com/aptd3v/containerkit/discussions)

## License

MIT. See [LICENSE](./LICENSE).
