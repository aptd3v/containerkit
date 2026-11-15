# Networks

[Network](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Network) attaches a service, job, or container to a named network. [Compose.WithNetwork](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.WithNetwork) declares that network on the project.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

compose := NewCompose("app",
    AddNetwork(Name("front"), Driver("bridge")),
).WithService(NewService(
    Name("api"),
    Image("alpine:latest"),
    Network(
        Name("front"),
        Alias("api"),
        IPv4("10.0.0.8"),
    ),
))
```

[NetworkMode](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NetworkMode) is `bridge`, `host`, or `none`. The host namespace constant is `NetworkHost`. [Host](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Host) is the daemon address on [Connect](bring-a-compose-file-up.md), not this mode.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("api"),
    Image("alpine:latest"),
    NetworkMode(NetworkHost),
)
```

[DNS](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#DNS) adds a resolver. [NetworkDisabled](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NetworkDisabled) sets `network_mode: none`.
