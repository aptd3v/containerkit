# Run a container

[Client.Run](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Client.Run) creates and starts a container on the client's Engine. [Run](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Run) dials [Connect](bring-a-compose-file-up.md), runs, and closes.

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func start(ctx context.Context) error {
    return Run(ctx,
        Name("web"),
        Image("nginx:latest"),
        Publish(HostPort(8080), ContainerPort(80), Protocol(TCP)),
    )
}
```

[Container.Create](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Container.Create) returns the id when you need it before start.

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func create(ctx context.Context, cli *Client) (string, error) {
    worker := NewContainer(Name("worker"), Image("alpine:latest"))
    created, err := worker.Create(ctx, cli.Engine())
    if err != nil {
        return "", err
    }
    if err := worker.Start(ctx, cli.Engine(), created.ID); err != nil {
        return "", err
    }
    return created.ID, nil
}
```

[PullImage](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#PullImage) and [BuildImage](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#BuildImage) take the Engine from the same client. A [Logger](logging.md) is the progress stream.

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func images(ctx context.Context, cli *Client, log *Logger) error {
    if _, err := PullImage(ctx, cli.Engine(), "alpine:latest", Log(log)); err != nil {
        return err
    }
    _, err := BuildImage(ctx, cli.Engine(),
        NewDockerfile().From("alpine", "latest"),
        Image("app:dev"),
        Log(log),
    )
    return err
}
```
