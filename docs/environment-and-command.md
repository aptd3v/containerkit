# Environment and command

[Command](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Command), [Entrypoint](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Entrypoint), [Env](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Env), and [Label](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Label) apply to a container, a service, and a job. [User](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#User) is the process user. [WorkingDir](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#WorkingDir) is the directory inside the container.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("api"),
    Image("alpine:latest"),
    User("65534:65534"),
    WorkingDir("/srv"),
    Entrypoint("httpd"),
    Command("-f", "-p", "80", "-h", "/srv"),
    Env("APP", "prod"),
    Label("app", "api"),
    StopSignal(SIGTERM),
)
```

`Env` inside [Build](build-and-dockerfile.md) is a build-arg. `WorkingDir` is the directory inside the container. [ProjectDirectory](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#ProjectDirectory) is the compose project directory on a load. The kit re-exports `SIGTERM` and `SIGKILL`. The rest of the signal names are in `pkg/signal`.
