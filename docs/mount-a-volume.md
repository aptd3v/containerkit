# Mount a volume

[Volume](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Volume) is a bind, a named volume, or tmpfs on a container, a service, and a job. [ReadOnly](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#ReadOnly) inside `Volume` is that mount. At the top level it is the container root filesystem.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

svc := NewService(
    Name("worker"),
    Image("alpine:latest"),
    Volume(
        Kind(Bind),
        Source("/host/data"),
        Target("/data"),
        ReadOnly(true),
    ),
    Volume(
        Kind(NamedVolume),
        Name("cache"),
        Target("/cache"),
    ),
    Volume(
        Kind(Tmpfs),
        Target("/tmp"),
    ),
)
```

A top-level compose `volumes:` entry is [Compose.WithVolume](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Compose.WithVolume), not this setter. [VolumesFrom](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#VolumesFrom) mounts another container's volumes. [VolumeDriver](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#VolumeDriver) sets the engine volume driver name.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"

compose := NewCompose("app",
    AddVolume(Name("cache"), Driver("local")),
).WithService(NewService(
    Name("worker"),
    Image("alpine:latest"),
    Volume(Name("cache"), Target("/cache")),
))
```
