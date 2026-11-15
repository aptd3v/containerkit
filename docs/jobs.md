# Jobs

[NewJob](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NewJob) is a compose job: the same container setters as a service, plus build, secrets, and depends_on. [Compose.WithJob](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.WithJob) adds it to the file.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

migrate := NewJob(
    Name("migrate"),
    Image("alpine:latest"),
    Command("echo", "migrated"),
    DependsOn(Name("db"), Condition(CompletedSuccessfully)),
)

compose := NewCompose("app").
    WithJob(migrate).
    WithService(NewService(Name("api"), Image("alpine:latest")))
```

A job does not take [Publish](publish-a-port.md) or [Deploy](resources-and-deploy.md). Those stay on services.
