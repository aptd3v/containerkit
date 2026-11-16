# Logging

[Logging](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Logging) is the container log driver: Engine `LogConfig`, Compose `logging`. [Driver](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Driver) is the name. `JSONFile` is the json-file driver. [DriverOpt](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#DriverOpt) is one option.

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

[NewLogger](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#NewLogger) is the progress stream for pull, build, and compose. [Log](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Log) passes it to those calls. [Color](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Color), [Progress](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Progress), and [OutputStream](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#OutputStream) configure it. With no options it writes to stderr and colors when that destination is a terminal.

```go
import (
    "os"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

log := NewLogger(OutputStream(os.Stdout), Color(false), Progress("plain"))
```

`Log` is that logger. `Logging` is the container's log driver.

A log line uses the same columns as a status row: kind, glyph, id, operation, then the text. `NO_COLOR` turns color off. With no `Color` option, color follows the terminal.

On a compose `Up` that is not detached, `Log` is also the container log consumer. Service stdout is printed through it. `Detach(true)` keeps the logger on progress and does not attach. An `Up` with neither `Detach(true)` nor `Log` still runs the containers, and their output stays in the log driver.
