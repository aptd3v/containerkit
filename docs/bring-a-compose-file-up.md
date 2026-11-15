# Bring a compose file up

One [Connect](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Connect) gives you Engine and Compose. Load YAML, or define the file in code, then `Up`.

## From a file

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func upFile(ctx context.Context) error {
    cli, err := Connect(Host("unix:///var/run/docker.sock"))
    if err != nil {
        return err
    }
    defer cli.Close()

    loaded, err := cli.Load(ctx, "compose.yaml", Profile("dev"))
    if err != nil {
        return err
    }
    return loaded.Up(ctx, Detach(false))
}
```

[ImportFile](import-and-export.md) parses YAML with no daemon. Use it to inspect or `Write` a file. `cli.Load` talks to Compose, including remote includes and OCI.

## From code

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func upCode(ctx context.Context, cli *Client) error {
    compose := NewCompose("app").WithService(
        NewService(
            Name("api"),
            Image("alpine:latest"),
            Publish(HostPort(8080), ContainerPort(80), Protocol(TCP)),
        ),
    )
    return cli.Up(ctx, compose, Detach(true))
}
```

`Up` validates. Call [Validate](errors.md) yourself when you want the error before the daemon round-trip. [Host](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Host) is the daemon address. The host network namespace is `NetworkHost`.
