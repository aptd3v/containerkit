# Logging

[Logging](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Logging) is the container log driver (Engine `LogConfig`, Compose `logging`). [Driver](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Driver) is the name. `JSONFile` is the json-file driver. [DriverOpt](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#DriverOpt) is one option.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("api"),
    Image("alpine:latest"),
    Logging(
        Driver(JSONFile),
        DriverOpt("max-size", "10m"),
    ),
)
```

[NewLogger](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NewLogger) is the progress stream for pull, build, and compose. [Log](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Log) passes it to those calls. [Color](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Color), [Progress](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Progress), and [OutputStream](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#OutputStream) configure it. With no options it writes to stderr and colors when the destination is a terminal.

```go
import (
    "os"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

log := NewLogger(OutputStream(os.Stderr), Color(false), Progress("plain"))
```

`Log` is that logger. `Logging` is the container's log driver.
