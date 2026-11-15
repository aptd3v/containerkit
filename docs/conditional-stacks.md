# Conditional stacks

[logic.Group](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/logic#Group), [logic.When](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/logic#When), and [logic.Switch](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/logic#Switch) pick setters before they are applied. [AddService](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#AddService), [AddNetwork](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#AddNetwork), and [AddVolume](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#AddVolume) are the compose options those groups carry. [Compose.WithService](define-a-service.md) remains the call when you already have a `*Service`.

```go
import (
    . "github.com/aptd3v/containerkit/pkg/kit"
    "github.com/aptd3v/containerkit/pkg/logic"
)

prod := true

compose := NewCompose("app", logic.Group(
    AddNetwork(Name("front"), Driver("bridge")),
    logic.When(prod, AddVolume(Name("data"))),
    AddService(
        Name("api"),
        Image("alpine:latest"),
        logic.When(prod, Restart(Policy(UnlessStopped))),
    ),
))
```

`logic` stays a named import. The setters inside it come from the dot import.
