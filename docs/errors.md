# Errors

Constructors return the object and keep configuration errors on it. [Validate](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.Validate) reports them. Runtime calls (`Up`, `Create`, `ExportYAML`, `Write`) validate before they talk to the daemon.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

compose := NewCompose("app").WithService(NewService(Name("api")))
if err := compose.Validate(); err != nil {
    return err
}
```

A daemon failure is an [Error](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Error). Its class is one of `ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrPermissionDenied`, `ErrAlreadyExists`, `ErrUnavailable`, `ErrNotImplemented`, `ErrCanceled`, `ErrInvalidArgument`, and the other sentinels. [IsNotFound](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#IsNotFound) matches a classified not-found without importing Engine or Compose.

```go
import (
    "errors"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

if IsNotFound(err) {
    return nil
}
if errors.Is(err, ErrConflict) {
    return err
}
var kitErr *Error
if errors.As(err, &kitErr) {
    _ = kitErr.Path
    _ = kitErr.Hint()
    _ = kitErr.StatusCode()
}
```

Secret values are redacted on `Error`. `Unwrap` still carries the original cause.
