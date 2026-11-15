# Publish a port

[Publish](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Publish) maps a host port to a container port on a container, a service, and a job. [Expose](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Expose) records the container port with no host binding.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("api"),
    Image("alpine:latest"),
    Publish(
        Protocol(TCP),
        HostIP("0.0.0.0"),
        HostPort(8080),
        ContainerPort(80),
        Mode(PortHost),
    ),
    Expose(8081, TCP),
)
```

`Publish` is a port mapping. Publishing the compose application as an OCI artifact is [Compose.PublishOCI](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.PublishOCI). `Mode` inside `Publish` is `PortIngress` or `PortHost`. The same setter inside [Deploy](resources-and-deploy.md) is `Replicated` or `Global`, and `svcmode.ReplicatedJob` or `svcmode.GlobalJob` for a job.
