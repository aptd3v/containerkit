# Environment and command

[Command](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Command), [Entrypoint](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Entrypoint), [Env](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Env), and [Label](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Label) apply to a container, a service, and a job. [User](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#User) and [WorkingDir](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#WorkingDir) are the process user and the container working directory.

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

`Env` inside [Build](build-and-dockerfile.md) is a build-arg. `WorkingDir` is not [ProjectDirectory](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#ProjectDirectory); that one is the compose project directory on a load.
