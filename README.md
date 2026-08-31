# containerkit

[![Go Reference](https://pkg.go.dev/badge/github.com/aptd3v/containerkit.svg)](https://pkg.go.dev/github.com/aptd3v/containerkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/aptd3v/containerkit)](https://goreportcard.com/report/github.com/aptd3v/containerkit)
[![Go Version](https://img.shields.io/github/go-mod/go-version/aptd3v/containerkit)](https://go.dev/dl/)
[![License](https://img.shields.io/github/license/aptd3v/containerkit)](./LICENSE)

Define Docker stacks in Go. The same `*containerkit.Container` feeds either the Compose CLI or the Docker Engine SDK.

```bash
go get github.com/aptd3v/containerkit@latest
```

## Two packages

| Package | Import | Role |
|---|---|---|
| Compose + fluent config | `github.com/aptd3v/containerkit/pkg/containerkit` | `Project`, `Container`, `NewCompose`, option structs (`Up`, `Down`, …) |
| Engine client | `github.com/aptd3v/containerkit/pkg/client` | Thin Docker SDK wrapper; per-op structs (`Remove`, `Exec`, `ImageBuild`, …) |

## Quick start

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

	app := containerkit.NewCompose(project)
	if err := app.Up(context.Background(), &containerkit.Up{Writer: os.Stdout}); err != nil {
		log.Fatal(err)
	}
}
```

Nil compose options mean defaults (writer is `os.Stdout`).

## Compose stack

```go
project := containerkit.NewProject("my-app")
project.
	WithNetwork("backend").
	WithVolume("data").
	WithService("api",
		containerkit.NewContainer().
			Image("alpine:latest").
			Command("sleep", "inf").
			PortBindings("tcp", "0.0.0.0", "8080", "8080").
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

Export YAML when you want a file:

```go
if err := project.Export("./docker-compose.yaml", 0644); err != nil {
	log.Fatal(err)
}
```

## Engine client

A container name is required for `ContainerCreate`. Construction options stay funcs on `client`.

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

The same `*containerkit.Container` can be passed to the Engine client or added as a Compose service.

## Fluent configuration

Methods live on `*containerkit.Container`. Nested objects (health, mount, build) and Compose CLI flags are structs. Service extras (`DependsOn`, `Profiles`, `BuildSpec`) are extra args to `WithService`.

```go
project.WithService("api",
	containerkit.NewContainer("my-api").
		Imagef("ubuntu:%s", tag).
		PortBindings("tcp", "0.0.0.0", "8080", "80").
		Endpoint("my-network").
		Architecture("amd64"),
	containerkit.DependsOn("db"),
	containerkit.Profiles("full"),
	containerkit.BuildSpec{Context: ".", Dockerfile: "Dockerfile"},
)
```

You can also mix in Docker SDK structs. See [examples/structs](./examples/structs).

## Conditionals

Prefer ordinary Go `if` inside factories. `WhenTrue` / `Group` / `Apply` / `OnlyIf` remain for first-class fragments.

```go
c := containerkit.NewContainer().Image("node:latest")
if runtime.GOOS == "linux" {
	c.RWNamedVolumeMount("data", "/app")
} else {
	c.VolumeBinds("./:/app/:rw")
}
```

## Profiles

`containerkit.Profiles` on a service is Compose’s `profiles:` key. To start those services, pass the same names on the command (`Up.Profiles`, `Down.Profiles`, `Kill.Profiles`, …). If the project has profiled services and the command omits `Profiles`, compose returns an error.

```go
project.WithService("worker",
	containerkit.NewContainer().Image("alpine:latest"),
	containerkit.Profiles("full"),
)
_ = app.Up(ctx, &containerkit.Up{Profiles: []string{"full"}, Writer: os.Stdout})
```

## Examples

Run from the repo root.

| Path | Path type | What it shows |
|---|---|---|
| [examples/simple](./examples/simple) | Compose | Minimal `NewProject` → `Up` |
| [examples/wordpress](./examples/wordpress) | Compose | Multi-service, OS-conditional mounts, logs, signal `Down` |
| [examples/profiles](./examples/profiles) | Compose | `Profiles` mirrored on `Up` / `Kill` |
| [examples/events](./examples/events) | Compose | Events channel + healthcheck |
| [examples/compose_commands](./examples/compose_commands) | Compose | `Build`, `Ps`, `Start`, `Stop`, `Restart`, `Exec`, … |
| [examples/nginx](./examples/nginx) | Hybrid | `client.ImageBuild` then compose `Up` |
| [examples/image_inline](./examples/image_inline) | Hybrid | Inline Dockerfile `BuildSpec` |
| [examples/mongo_replica](./examples/mongo_replica) | Hybrid | Compose `Up` + client exec to init a replica set |
| [examples/terminal](./examples/terminal) | Client | Interactive TTY exec + resize |
| [examples/structs](./examples/structs) | Config | Mix Docker SDK structs with fluent methods |
| [examples/supabase](./examples/supabase) | Compose | Full multi-profile stack |

```bash
go run ./examples/simple/
go run ./examples/supabase/                    # minimal
go run ./examples/supabase/ -profile full
go run ./examples/supabase/ -resource-limits
```

## Prerequisites

- **Go** 1.24+
- **Docker** 28.2+ with Compose v2
- Linux, macOS, or Windows

## Migrating from go-contain

This project was renamed from `go-contain`. The public API is two packages; the old option-func packages are gone.

| Old | New |
|---|---|
| `github.com/aptd3v/go-contain` | `github.com/aptd3v/containerkit` |
| `pkg/create`, `cc.WithImage`, … | `pkg/containerkit`, fluent `.Image()`, … |
| `pkg/compose`, `up.WithRemoveOrphans()` | `containerkit.NewCompose` + `&containerkit.Up{…}` |
| `pkg/client/options/**` | per-op structs on `pkg/client` (`Remove`, `Exec`, …) |

## License

MIT. See [LICENSE](./LICENSE).

## Contributing

Issues and pull requests are welcome: [issues](https://github.com/aptd3v/containerkit/issues) · [discussions](https://github.com/aptd3v/containerkit/discussions)
