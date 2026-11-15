# Import and export

[ImportYAML](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#ImportYAML) and [ImportFile](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#ImportFile) parse Compose with no daemon. [Env](environment-and-command.md) on import interpolates `${VAR}`. [File](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#File) adds another YAML path. [Name](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Name) overrides the project name.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

compose := ImportYAML(yml, Name("app"), Env("TAG", "latest"))
local := ImportFile("compose.yaml", File("compose.override.yaml"), Profile("dev"))
if err := local.Validate(); err != nil {
    return err
}
```

[Compose.ExportYAML](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.ExportYAML) returns the bytes. [Compose.Write](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.Write) sends them to an [Export](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Export) writer. Both validate first.

```go
import (
    "os"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func writeCompose(compose *Compose) error {
    yml, err := compose.ExportYAML()
    if err != nil {
        return err
    }
    _ = yml
    return compose.Write(Export{Writer: os.Stdout})
}
```

`os.Stdout` needs a normal `os` import next to the dot import. Bringing that file up on a daemon is [Load](bring-a-compose-file-up.md) or [Client.Up](bring-a-compose-file-up.md). A TSX plan JSON is [ImportPlan](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#ImportPlan), then `Run`.

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func runPlan(ctx context.Context, json []byte) error {
    return ImportPlan(json).Run(ctx)
}
```
