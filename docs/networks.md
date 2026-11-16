# Networks

[Network](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Network) attaches a container, a service, or a job to a named network. Declare that network with [AddNetwork](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#AddNetwork) on `NewCompose`, or with [Compose.WithNetwork](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.WithNetwork).

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

compose := NewCompose("app",
    AddNetwork(
        Name("front"),
        Driver("bridge"),
        Subnet("10.0.0.0/24"),
        Gateway("10.0.0.1"),
    ),
).WithService(NewService(
    Name("api"),
    Image("alpine:latest"),
    Hostname("api"),
    Network(
        Name("front"),
        Alias("api"),
        IPv4("10.0.0.8"),
        MacAddress("02:42:0a:00:00:08"),
    ),
))
```

`Subnet` and `Gateway` belong on the project network. `IPv4`, `IPv6`, and `MacAddress` inside `Network` are that endpoint, and the address has to fall inside the subnet. `MacAddress` on the service itself is the container MAC. `Hostname` is the name the container uses for itself.

[NetworkMode](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NetworkMode) selects a namespace. `NetworkHost` is the host namespace. [Host](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Host) is the daemon address on [Connect](bring-a-compose-file-up.md).

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("api"),
    Image("alpine:latest"),
    NetworkMode(NetworkHost),
)
```

[DNS](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#DNS) adds a resolver. [NetworkDisabled](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NetworkDisabled) sets `network_mode: none`.
