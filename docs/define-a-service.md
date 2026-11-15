# A service

[NewContainer](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NewContainer), [NewService](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NewService), and [NewCompose](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NewCompose) take the same setters. A `*Container` is a `ServiceOption`, so the service keeps the container's name and image.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

worker := NewContainer(
    Name("worker"),
    Image("alpine:latest"),
    Command("sleep", "infinity"),
)

api := NewService(worker)
compose := NewCompose("app").WithService(api)
if err := compose.Validate(); err != nil {
    return err
}
```

`NewService` without a container still takes the setters directly:

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("api"),
    Image("nginx:latest"),
)
```

[Container.AsService](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Container.AsService) is the same copy. Engine-only fields (`AutoRemove`, `PublishAll`, `ContainerIDFile`, `Shell`, `OnBuild`, `MaskedPath`, `ReadonlyPath`) stay on the container. `Validate` on the service reports each one.

Constructors always return a value. `Validate` is how you read the errors before a daemon call. `Up`, `Create`, `ExportYAML`, and `Write` validate on their own.
