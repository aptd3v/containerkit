# Restart and depends_on

[Restart](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Restart) takes [Policy](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Policy) and, for `OnFailure`, [MaxRetries](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#MaxRetries). Policies are `NoRestart`, `OnFailure`, `Always`, and `UnlessStopped`.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

db := NewService(
    Name("db"),
    Image("postgres:16"),
    Restart(Policy(UnlessStopped)),
)

api := NewService(
    Name("api"),
    Image("alpine:latest"),
    Restart(Policy(OnFailure), MaxRetries(3)),
    DependsOn(
        Name("db"),
        Condition(Healthy),
    ),
)
```

[DependsOn](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#DependsOn) needs [Name](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Name). [Condition](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Condition) is `Started`, `Healthy`, or `CompletedSuccessfully`. [RestartOnFailure](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#RestartOnFailure) is the depends_on restart flag, not the service restart policy.
