# Healthcheck

[Health](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Health) takes [Test](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Test), [Interval](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Interval), [Timeout](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Timeout), [Retries](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Retries), and [StartPeriod](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#StartPeriod). `Test` is only the command. Extra strings are arguments.

```go
import (
    "time"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

svc := NewService(
    Name("api"),
    Image("alpine:latest"),
    Health(
        Test("CMD", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1/"),
        Interval(5*time.Second),
        Timeout(2*time.Second),
        Retries(3),
        StartPeriod(10*time.Second),
    ),
)
```

[Disabled](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Disabled) turns the healthcheck off. The same setter, passed to plugin install, leaves the plugin disabled. `Timeout` outside `Health` is a compose wait, an engine stop, or a client deadline.
