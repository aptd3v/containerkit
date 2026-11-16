# Restart and depends_on

[Restart](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Restart) takes [Policy](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Policy). `OnFailure` also takes [MaxRetries](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#MaxRetries). The policies are `NoRestart`, `OnFailure`, `Always`, and `UnlessStopped`.

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

[DependsOn](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#DependsOn) needs [Name](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Name). [Condition](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Condition) is `Started`, `Healthy`, or `CompletedSuccessfully`. Omit it and the condition is `Started`. Call `DependsOn` again to add another dependency. It does not replace the previous one.

[RestartOnFailure](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#RestartOnFailure) is the `depends_on` restart flag. It is not the service restart policy.
