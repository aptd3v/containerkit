# Resources and deploy

[Memory](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Memory) on a service is `mem_limit`. The same setter inside [Deploy](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Deploy) is `deploy.resources.limits.memory`.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("api"),
    Image("alpine:latest"),
    Memory(512 << 20),
    Deploy(
        Mode(Replicated),
        Replicas(3),
        Memory(1 << 30),
    ),
    Profile("prod"),
)
```

[Replicas](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Replicas) inside `Deploy` is the swarm replica count. On the service itself it is compose `scale`. [Mode](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Mode) here is `Replicated` or `Global`. `svcmode.ReplicatedJob` and `svcmode.GlobalJob` are the job modes. Inside [Publish](publish-a-port.md) it is `PortIngress` or `PortHost`.

[ShmSize](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#ShmSize) is `/dev/shm`. [StopSignal](environment-and-command.md) takes a signal name. The kit re-exports `SIGTERM` and `SIGKILL`. The rest are in `pkg/signal`.
