# containerkit

[![Go Reference](https://pkg.go.dev/badge/github.com/aptd3v/containerkit.svg)](https://pkg.go.dev/github.com/aptd3v/containerkit)
[![Go Version](https://img.shields.io/github/go-mod/go-version/aptd3v/containerkit)](https://go.dev/dl/)
[![License](https://img.shields.io/github/license/aptd3v/containerkit)](./LICENSE)

Containerkit is the Docker stack as a program. Define a stack in Go or TSX, then use one client to pull an image, bring up the Compose project, and inspect its services. Both languages share a Docker vocabulary, and the client covers both the Engine and Compose SDKs.

Export the project as Compose YAML, or the image definition as a Dockerfile. For live development, a `docker.ts` keeps the session open and reapplies the project only when the rendered YAML changes.

## A project is a component

```tsx
import { Compose, Service, Publish, Volume, Health, Env, Restart, exportYAML } from "containerkit";
import { proto, restartPolicy } from "containerkit/vocab";

const prod = process.env.NODE_ENV === "production";

export default function App() {
  return (
    <Compose name="app">
      <Volume name="data" />
      <Service
        name="api"
        image="alpine:latest"
        command={["sh", "-c", "mkdir -p /www && echo ok > /www/index.html && exec httpd -f -p 80 -h /www"]}
      >
        <Publish hostPort={8080} containerPort={80} protocol={proto.TCP} />
        <Volume source="data" target="/www" />
        <Restart policy={restartPolicy.UnlessStopped} />
        <Health test={["CMD", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1/"]} interval="5s" retries={3} />
        <Env name="APP" value={prod ? "prod" : "dev"} />
      </Service>
    </Compose>
  );
}

export const yaml = exportYAML(<App />);
```

Containerkit generates its bindings from pinned Moby and Compose SDK versions. Updating the pins and regenerating picks up new upstream fields.

`false` disappears. `.map()` repeats a child. A function that returns a node is a component, so a gateway, a database, and a pooler are props on one stack. [`examples/supabase`](examples/supabase) is that stack in Go. [`tsx/examples/supabase.tsx`](tsx/examples/supabase.tsx) is the same project in TSX: Envoy or Kong, Postgres 15 or 17, Supavisor or PgBouncer, file or S3, logs on or off, printed as Compose on stdout.

A Dockerfile is the other tree, in `containerkit/dockerfile`. `<Run>` joins into one layer. `<RunLayer>` starts the next. `exportDockerfile` prints the file. [`tsx/examples/image.tsx`](tsx/examples/image.tsx) is a multi-stage build whose stages and layers follow the environment.

[`tsx/examples/taskflow.tsx`](tsx/examples/taskflow.tsx) orders those pieces as a graph. The image build overlaps Postgres coming up. Lint waits for that image, then overlaps the database. Migrate can be tried again. A failed check skips the tag. Down waits only for migrate, because unit and lint do not use Postgres, and the project is still taken down when migrate fails. Stdout is the plan document.

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

## It stays open

`docker.ts` default-exports `root()`. The session stays up for the life of the process. A later render applies the project only when the rendered YAML changes. Exporting the file does not stop it. [`tsx/examples/docker.tsx`](tsx/examples/docker.tsx) is a running copy: from `tsx/`, `npm run live` builds a small image, serves `http://127.0.0.1:18080`, and `onExit: "down"` takes the project down on Ctrl+C.

```tsx
import { Compose, Service, Publish, Health } from "containerkit";
import { proto } from "containerkit/vocab";
import { root } from "containerkit/live";

export default root({ fromEnv: true }, function App() {
  return (
    <Compose name="app">
      <Service
        name="api"
        image="alpine:latest"
        command={["sh", "-c", "mkdir -p /www && echo ok > /www/index.html && exec httpd -f -p 80 -h /www"]}
      >
        <Publish hostPort={8080} containerPort={80} protocol={proto.TCP} />
        <Health test={["CMD", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1/"]} interval="5s" retries={3} />
      </Service>
    </Compose>
  );
});
```

Commands are methods on that session. `await app.pullImage("alpine:latest")` runs, then the next line runs. `useMemo` writes the result to `.containerkit/memo` and skips the call the next time the same inputs come back.

## The same words in Go

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

`Connect` is the daemon. `Run` starts a container. `Up` applies a project. `ImportFile` reads a compose file with no daemon. `Name`, `Image`, `Env`, `Memory`, and `ReadOnly` mean the same thing on a container, a service, and a job. The compiler rejects a name on a scope that does not have it.

A few words only look the same:

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

Go, from `go.mod` (1.26.8 or newer):

```bash
go get github.com/aptd3v/containerkit
```

Node 20 or newer, once `v0.8.1` is on npm:

```bash
npm install containerkit
```

[CHANGELOG.md](CHANGELOG.md) is what 0.8.1 contains.

Read the Go kit in order from [docs/README.md](docs/README.md): define a service, then run it. Read TypeScript from [tsx/docs/README.md](tsx/docs/README.md): trees, the live session, the client, commands, then Dockerfiles.

The TypeScript package lives in [`tsx/`](tsx/README.md). A folder that extends `containerkit/tsconfig.jsx.json` compiles `<Compose>` without a per-file pragma. React and containerkit can share a repository. They cannot share a file.

A live session needs the `containerkit` binary on `PATH`, or `CONTAINERKIT_BIN` pointing at it. [The runner](docs/runner-binary.md) is `validate`, `exec`, and `session`, and `go run ./tools/crossbuild` builds it for linux, darwin, and windows.

- [For coding agents](llms.txt). Emit containerkit instead of a Compose file plus a script.
- [Contributing](CONTRIBUTING.md). Tests, `go generate`, and `VERSION`.

## License

MIT. See [LICENSE](LICENSE).
