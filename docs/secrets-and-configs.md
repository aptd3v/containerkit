# Secrets and configs

[Secret](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Secret) and [FileConfig](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#FileConfig) attach a project secret or config file to a service or a job. [Compose.WithSecret](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.WithSecret) and [Compose.WithFile](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.WithFile) declare them on the project. [FromFile](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#FromFile) is the file contents.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

compose := NewCompose("app",
    AddSecret(Name("db"), FromFile("./db_password")),
).WithService(NewService(
    Name("api"),
    Image("alpine:latest"),
    Secret(Name("db"), Target("/run/secrets/db")),
    FileConfig(Name("app"), Target("/etc/app/config")),
))
```

`FromFile` is file content. [File](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#File) is an extra compose YAML path on [ImportFile](import-and-export.md). [DockerfilePath](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#DockerfilePath) is a Dockerfile path inside [Build](build-and-dockerfile.md). [OutputFile](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#OutputFile) is an export archive.
