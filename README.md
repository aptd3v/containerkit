# containerkit

[![Go Reference](https://pkg.go.dev/badge/github.com/aptd3v/containerkit.svg)](https://pkg.go.dev/github.com/aptd3v/containerkit)
[![Go Version](https://img.shields.io/github/go-mod/go-version/aptd3v/containerkit)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-Apache_2.0-blue)](./LICENSE)

Containerkit is the Docker stack as a Go program. Define a project with setters, then use one client to pull an image, bring the Compose project up, and inspect its services. The client covers the Engine and Compose SDKs. `Name`, `Image`, `Env`, `Memory`, and `ReadOnly` mean the same thing on a container, a service, and a job. The compiler rejects a setter on a scope that does not have it.

Export the project as Compose YAML, or the image definition as a Dockerfile.

TypeScript docs, install, and the live session are in [tsx/README.md](tsx/README.md).

## Bring a project up

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

cli, err := Connect()
if err != nil {
    return err
}
defer cli.Close()

app := NewCompose("app").WithService(
    NewService(
        Name("api"),
        Image("alpine:latest"),
        Command("sh", "-c", "mkdir -p /www && echo ok > /www/index.html && exec httpd -f -p 80 -h /www"),
        Publish(Protocol(TCP), HostPort(8080), ContainerPort(80)),
    ),
)
if err := cli.Up(ctx, app); err != nil {
    return err
}

loaded, err := cli.Load(ctx, "compose.yaml")
if err != nil {
    return err
}
return loaded.Up(ctx)
```

`Connect` is the daemon. With no options it reads the host the `docker` command already uses. `Run` starts a container. `Up` applies a project. `ImportFile` reads a compose file with no daemon.

[`examples/supabase`](examples/supabase) is a large stack written this way. Smaller programs are [`examples/nginx`](examples/nginx), [`examples/wordpress`](examples/wordpress), [`examples/mongo_replica`](examples/mongo_replica), and [`examples/terminal`](examples/terminal).

## A container becomes a service

`NewContainer` and `NewService` write the same spec. `EngineConfig` turns that spec into the Engine SDK's `Config`, `HostConfig`, and `NetworkingConfig`. `AsService` copies the spec onto a Compose service, and `ComposeConfig` turns it into compose-go's `ServiceConfig`. `NewService(FromContainer(worker))` and passing the container itself do the same copy.

```go
worker := NewContainer(
    Name("api"),
    Image("alpine:latest"),
    Command("sh", "-c", "mkdir -p /www && echo ok > /www/index.html && exec httpd -f -p 80 -h /www"),
    Publish(Protocol(TCP), HostPort(8080), ContainerPort(80)),
)
svc := worker.AsService()
```

`AutoRemove`, `PublishAll`, `ContainerIDFile`, `Shell`, `OnBuild`, `MaskedPath`, and `ReadonlyPath` exist only on the Engine create call. They stay on the container. `Validate` on the service reports each one as an unsupported mapping.

## Names that only look the same

| Word | Means |
| --- | --- |
| `File` | A compose YAML path |
| `FromFile` | Secret or config content |
| `DockerfilePath` | The Dockerfile on a build |
| `OutputFile` | An archive written by export |
| `Publish` | A port mapping |
| `PublishOCI` | The project published as an OCI artifact |
| `Log` | The progress logger |
| `Logging` | The container log driver |
| `WorkingDir` | The directory inside the container |
| `ProjectDirectory` | The compose project directory on disk |

`Test` is the healthcheck command. `Interval`, `Timeout`, and `Retries` are the timing around it.

## Current with upstream

The kit is generated from the modules it wraps: `github.com/moby/moby/api`, `github.com/moby/moby/client`, `github.com/compose-spec/compose-go/v2`, and `github.com/docker/compose/v5`, plus the image spec, OCI, and reference types beside them. `gencatalog` copies each exported type into `internal/` under its original name. `gentsx` writes the TSX components and the session methods from that kit. Bump the modules, run `go generate`, and the tree moves with Docker.

## Start

Go, from `go.mod` (1.26.11 or newer):

```bash
go get github.com/aptd3v/containerkit
```

[CHANGELOG.md](CHANGELOG.md) is what 0.8.4 contains.

Read the kit in order from [docs/README.md](docs/README.md): define a service, then run it.

Node install and the TypeScript API are in [tsx/README.md](tsx/README.md).

[The runner](docs/runner-binary.md) is `validate`, `exec`, and `session`. `go run ./tools/crossbuild` builds it for linux, darwin, and windows. The TypeScript package ships that binary.

- [For coding agents](AGENTS.md). Emit containerkit instead of a Compose file plus a script.
- [Contributing](CONTRIBUTING.md). Tests, `go generate`, and `VERSION`.

## License

Apache 2.0. See [LICENSE](LICENSE).
