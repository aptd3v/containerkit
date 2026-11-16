# Same word, different scope

One setter keeps its meaning at every scope that accepts it.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("api"),
    Image("alpine:latest"),
    ReadOnly(true),
    Volume(Kind(Bind), Source("/host"), Target("/data"), ReadOnly(true)),
    Env("VERSION", "1"),
    Memory(1<<30),
    Build(Source("."), Env("VERSION", "1")),
    Deploy(Memory(1<<30)),
)
```

`Env` on the service is the process environment. `Env` inside `Build` is a build-arg. `Memory` on the service is `mem_limit`. `Memory` inside `Deploy` is the deploy limit. `ReadOnly` on the service is the root filesystem. `ReadOnly` inside `Volume` is that mount.

Names that only look alike are different setters:

| You want | Setter |
| --- | --- |
| Compose YAML path on import | [File](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#File) |
| Secret or config bytes | [FromFile](secrets-and-configs.md) |
| Dockerfile path | [DockerfilePath](build-and-dockerfile.md) |
| Export archive | [OutputFile](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#OutputFile) |
| Port mapping | [Publish](publish-a-port.md) |
| OCI artifact of the project | [Compose.PublishOCI](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.PublishOCI) |
| Progress logger | [Log](logging.md) |
| Container log driver | [Logging](logging.md) |
| Directory inside the container | [WorkingDir](environment-and-command.md) |
| Compose project directory | [ProjectDirectory](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#ProjectDirectory) |
| Include stopped containers in a list | [All](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#All) |
| LoadProject resource filter | [AllResources](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#AllResources) |
| Daemon address | [Host](bring-a-compose-file-up.md) |
| Host network namespace | `NetworkHost` in [NetworkMode](networks.md) |
| One healthcheck probe | `Timeout` inside [Health](healthcheck.md) |
| Wait for `depends_on` | `Timeout` on [Up](bring-a-compose-file-up.md) |

`Test` is the healthcheck command. `Interval`, `Timeout`, and `Retries` beside it are the timing of that check.
