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

`Progress` is `auto`, `plain`, `tty`, `quiet`, or `json`. `auto` uses the live region when the destination is a terminal and one line per event otherwise. `quiet` drops `working` events. `json` writes one JSON object per event and does not draw a live region. The object is the session event: `type`, `kind`, `group`, `status`, `source`, `stream`, `line`, and `display`. `text`, `detail`, `parent`, `current`, and `total` are present when set.

```go
log := NewLogger(OutputStream(os.Stdout), Progress("json"))
```

`NewFrameLogger` is the same event as a callback. It does not write the transcript. A program that wants both the callback and a rendering logger writes the line from the callback.

`Log` on a call and `Logger.Context` see the same Compose resource rows as `Log` passed to `Connect`. The client keeps one bridge. `Start` reads the logger on the call context. Container rows (`created`, `started`, `healthy`) go there. `Connect(..., Log(log))` is the fallback when the call has no logger. One Compose operation at a time per client.

On a compose `Up` that is not detached, `Log` is also the container log consumer. Service stdout is printed through it. `Detach(true)` keeps the logger on progress and does not attach. An `Up` with neither `Detach(true)` nor `Log` still runs the containers, and their output stays in the log driver.

`PullImage`, `BuildImage`, `PushImage`, and `LoadImage` read the Engine JSON progress stream into the logger. With a logger, `LoadImage` consumes that body and the returned reader is at EOF. Without a logger, the reader is the progress body. `Quiet` on load is the daemon flag. It is not implied by `Log`.

`Container.Logs` still returns the Engine stream. When a logger is present, reading it emits `KindLog` lines (`stdout` and `stderr`) and `Read` returns the original bytes, headers included. A stream that does not start with a multiplex header is treated as raw stdout. `Compose.Logs` with a nil consumer uses `Log` or the context logger, and it is an error when neither is set. Another consumer is teed with that logger. Passing the logger itself does not deliver the line twice.

Attach, exec, stats, and `SaveImages` stay raw streams.
