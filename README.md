# containerkit

[![Go Reference](https://pkg.go.dev/badge/github.com/aptd3v/containerkit.svg)](https://pkg.go.dev/github.com/aptd3v/containerkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/aptd3v/containerkit)](https://goreportcard.com/report/github.com/aptd3v/containerkit)
[![Go Version](https://img.shields.io/github/go-mod/go-version/aptd3v/containerkit)](https://go.dev/dl/)
[![License](https://img.shields.io/github/license/aptd3v/containerkit)](./LICENSE)

Define Docker stacks in Go. The same `*kit.Container` is a Compose service or an Engine create config. Compose runs in-process through `github.com/docker/compose/v5` and does not need a `docker compose` binary.

## Demo

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func main() {
	project := kit.NewProject("hello-world")
	project.WithService(kit.NewService("hello",
		kit.NewContainer().
			Image("alpine:latest").
			Command("echo", "Hello from containerkit!"),
	))

	if err := project.Validate(); err != nil {
		log.Fatal(err)
	}

	app := kit.NewCompose(project)
	if err := app.Up(context.Background(), &kit.Up{Writer: os.Stdout}); err != nil {
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
| `Project`, `*Container` | `github.com/aptd3v/containerkit/pkg/kit` | Stack and container config. Fluent methods set image, env, ports, mounts, and the rest. |
| SDK value types | `github.com/aptd3v/containerkit/pkg/kit/types` | `ByteSize`, `PortBinding`, `Capability`, host modes, mount aliases. |
| `NewCompose`, `Up`, `Down`, … | `pkg/kit` | In-process Compose runner. Each command takes an option struct. |
| Engine client | `github.com/aptd3v/containerkit/pkg/client` | Docker Engine SDK wrapper. Each call takes a per-op struct (`Remove`, `Exec`, `ImageBuild`, …). |

Compose fits a multi-service stack: networks, volumes, profiles, and `Up` / `Down` / `Logs`. The Engine client fits one container, image builds, and an interactive TTY. A hybrid program uses the client to build or exec, then Compose to run the stack. See [examples/nginx](./examples/nginx), [examples/image_inline](./examples/image_inline), and [examples/mongo_replica](./examples/mongo_replica).

### Compose stack

```go
project := kit.NewProject("my-app")
project.
	WithNetwork("backend").
	WithVolume("data").
	WithService(kit.NewService("api",
		kit.NewContainer().
			Image("alpine:latest").
			Command("sleep", "inf").
			PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "8080"}).
			RWNamedVolumeMount("data", "/data").
			Endpoint("backend"),
	).DependsOn("db"))

if err := project.Validate(); err != nil {
	log.Fatal(err)
}

app := kit.NewCompose(project)
if err := app.Up(ctx, &kit.Up{
	Detach:        true,
	RemoveOrphans: true,
	Writer:        os.Stdout,
}); err != nil {
	log.Fatal(err)
}
defer app.Down(ctx, &kit.Down{RemoveVolumes: true, Writer: os.Stdout})
```

`Export` writes the project as Compose YAML. `Print` writes that YAML to stdout. `WriteTo` writes it to any `io.Writer`.

```go
if err := project.Export("./docker-compose.yaml", 0644); err != nil {
	log.Fatal(err)
}
_ = project.Print()
_, _ = project.WriteTo(w)
```

### Engine client

`NewContainer` joins every name part with a hyphen (`"my", "api"` becomes `my-api`). Pass a name for `ContainerCreate`. A Compose service can omit it. Construction options stay funcs on `client`.

```go
cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
if err != nil {
	log.Fatal(err)
}

ctr := kit.NewContainer("my-api")
ctr.Image("alpine:latest").Command("sleep", "inf")

resp, err := cli.ContainerCreate(ctx, ctr)
if err != nil {
	log.Fatal(err)
}
_ = cli.ContainerRemove(ctx, resp.ID, &client.Remove{Force: true, RemoveVolumes: true})
```

Pass that same `*kit.Container` to `NewService` when the stack should run under Compose.

### Task graph

`pkg/taskflow` is optional. It orders the same `*kit.Container`, `kit.Build`, and `*kit.Compose` values as a DAG. Import it only when you want that; `pkg/kit` and `pkg/client` do not depend on it.

```go
flow := taskflow.New("ci", cli)
build := flow.Build("build",
	kit.NewBuild().Dockerfile(df),
	&client.ImageBuild{Tags: []string{"app:ci"}},
)
up := flow.Compose("up-db", db, &kit.Up{Detach: true, Wait: true})
test := flow.Container("test", kit.NewContainer().Image("app:ci"), nil)

test.DependsOn(up.DependsOn(build))

defer flow.Remove(ctx)
err := flow.Run(ctx, &taskflow.Run{Writer: os.Stdout})
```

A failed step skips its successors. Full guide: [`pkg/taskflow`](./pkg/taskflow). Demo: [examples/taskflow](./examples/taskflow).

### Dockerfile

`NewDockerFile` writes a Dockerfile in Go. `ParseDockerfile` wraps a string and `ReadDockerfile` loads a file. `Export` writes the file to disk. A `Build` carries the Dockerfile plus the build context: pass it to `Service.Build` for Compose, or call `Reader` for the tar `client.ImageBuild` accepts.

```go
df := kit.NewDockerFile().
	FromAs("alpine:latest", "build").
	Run("echo ok > /out/app").
	From("alpine:latest").
	Copy("/out/app", "/app", kit.CopyOpts{From: "build"})

if err := df.Export("Dockerfile", 0644); err != nil {
	log.Fatal(err)
}

build := kit.NewBuild().Context(".").Dockerfile(df)

project.WithService(kit.NewService("app",
	kit.NewContainer().Image("my-image:tag"),
).Build(build))

r, err := build.Reader()
if err != nil {
	log.Fatal(err)
}
resp, err := cli.ImageBuild(ctx, r, &client.ImageBuild{Tags: []string{"my-image:tag"}})
```

`Reader` packs the context directory, writes the generated Dockerfile over any `Dockerfile` already there, and honors `.dockerignore`. An empty context packs only the Dockerfile. When only `Context` or `DockerfilePath` is set, `Reader` reads that file from disk. See [examples/dockerfile](./examples/dockerfile) and [examples/image_inline](./examples/image_inline).

### Rules

- Setters append to `Container.Errors` and to the project. Call `Validate` before `Up`, `Export`, or `ContainerCreate`. Those three calls run `Validate` themselves.
- A nil `Up` or `Down` uses defaults. The writer is `os.Stdout`.
- `Service.Profiles` is Compose’s `profiles:` key. If the project has any profiled service, set `Profiles` on the command (`Up.Profiles`, `Down.Profiles`, and the rest). Compose returns an error when that list is empty.

```go
project.WithService(kit.NewService("worker",
	kit.NewContainer().Image("alpine:latest"),
).Profiles("full"))
_ = app.Up(ctx, &kit.Up{Profiles: []string{"full"}, Writer: os.Stdout})
```

- Compose-only settings are methods on `*Service`: `DependsOn`, `Profiles`, `Build`, `Deploy`. `*Container` stays the process config.
- Sizes are a `types.ByteSize`: `MemoryLimit(100 * types.MiB)`.
- Branch with an ordinary `if`.

```go
c := kit.NewContainer().Image("node:latest")
if runtime.GOOS == "linux" {
	c.RWNamedVolumeMount("data", "/app")
} else {
	c.VolumeBinds("./:/app/:rw")
}
```

### API surface

Full signatures are on [pkg.go.dev](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit).

**Container.** Image, env, command, ports, mounts, health, restart, resources, capabilities, logging, and platform. You can also mix in Docker SDK structs. See [examples/structs](./examples/structs).

**Service.** `NewService(name, container)`, then `DependsOn`, `DependsOnHealthy`, `Profiles`, `Build`, `Deploy`, `Secret`, and the other Compose-only methods. Pass the service to `Project.WithService`.

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
| [examples/taskflow](./examples/taskflow) | Task graph | `pkg/taskflow`: build, Compose `Up`, parallel test containers, retry, `Down` |
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
| `pkg/create`, `cc.WithImage`, … | `pkg/kit`, fluent `.Image()`, … |
| `pkg/compose`, `up.WithRemoveOrphans()` | `kit.NewCompose` + `&kit.Up{…}` |
| `pkg/client/options/**` | per-op structs on `pkg/client` (`Remove`, `Exec`, …) |

## Contributing

Issues and pull requests are welcome: [issues](https://github.com/aptd3v/containerkit/issues) · [discussions](https://github.com/aptd3v/containerkit/discussions)

## License

MIT. See [LICENSE](./LICENSE).
