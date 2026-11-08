// Package kit is a Go SDK for defining and running Docker containers.
//
// The same *Container is a Compose service config or an Engine create config.
// Projects run in-process through github.com/docker/compose/v5 (no docker compose
// binary). Pair kit with github.com/aptd3v/containerkit/pkg/client for Engine
// operations, and optionally github.com/aptd3v/containerkit/pkg/taskflow for a
// DAG over the same values. The HTTP daemon and canvas UI are not part of this
// published module.
package kit
