# Bring a compose file up

[Connect](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Connect) dials Engine and Compose together. [FromEnv](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#FromEnv) uses the host the `docker` command already uses. [Host](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Host) is an explicit daemon address. The host network namespace is `NetworkHost`, not this setter.

[Client.Up](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Client.Up) and [Loaded.Up](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Loaded.Up) already hold the composer. [Compose.Up](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.Up) takes it as the second argument: `app.Up(ctx, cli.Composer(), ...)`.

## From a file

```go
import (
    "context"
    "time"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func upFile(ctx context.Context) error {
    cli, err := Connect(FromEnv())
    if err != nil {
        return err
    }
    defer cli.Close()

    loaded, err := cli.Load(ctx, "compose.yaml", Profile("dev"))
    if err != nil {
        return err
    }
    return loaded.Up(ctx, Detach(true), Timeout(2*time.Minute))
}
```

[ImportFile](import-and-export.md) parses YAML with no daemon. Use it to inspect or [Write](import-and-export.md) a file. `Client.Load` talks to Compose, including remote includes and OCI. `Profile` on `Load` keeps the services in that profile.

## From code

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func upCode(ctx context.Context, cli *Client) error {
    app := NewCompose("app").WithService(NewService(
        Name("api"),
        Image("alpine:latest"),
        Publish(Protocol(TCP), HostPort(8080), ContainerPort(80)),
    ))
    return app.Up(ctx, cli.Composer(), Detach(true), RemoveOrphans(true))
}
```

`Up` validates before it talks to the daemon. Call [Validate](errors.md) when you want that error with no round-trip.

`Timeout` on `Up` is how long Compose waits for a `depends_on` condition, and the stop grace used when a container is recreated. A service that waits for a database to become [healthy](healthcheck.md) needs minutes, not a few seconds. `Detach(true)` returns after start. Omit it, or pass `Detach(false)`, and a [Log](logging.md) follows container output until the containers exit.
