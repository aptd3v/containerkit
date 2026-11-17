# Changelog

## 0.8.4

A taskflow run keeps its containers, edges, and volumes together. Pull, compose, and container logs are one event.

- A cycle, a shared container name, and a failed apply are errors. A subflow cycle is checked only among the tasks that subflow created, and that subflow stamps only those tasks. `Run`, `List`, and `Remove` no longer replace a nil context with `context.Background`.
- `SkipIfPresent` on a build inspects every tag. All present finishes done with reason `already_present` and does not build. A missing tag builds once. `SkipIfPresent` on any other step fails that step once.
- `Service` stays up until every other task in the declaring flow or subflow has finished, then downs once without a volume wipe. A canceled run still downs. Two services in one scope tear down without waiting on each other.
- `Volume` allocates a physical name when it is called. `Run` creates that volume labeled with the flow and the run. `Remove` deletes it only after this run's containers are gone, and only when the run label matches.
- `@aptd3v/containerkit/taskflow` `Flow` accepts `volumes`. `volumeSource` emits the mount source Import rewrites. `Service` emits `op: "service"`. `ComposeStep` uses the op to select the options type and the required fields.
- `Progress` `json` writes one JSON object per log or status event. `quiet` drops `working` events. `NewFrameLogger` delivers the same event to a callback.
- Pins verified 2026-11-16: `github.com/moby/moby/api` v1.57.0, `github.com/moby/moby/client` v0.6.3, `github.com/compose-spec/compose-go/v2` v2.17.0, `github.com/docker/compose/v5` v5.7.1. Go is 1.26.11.

## 0.8.3

The tag workflow publishes with GitHub's OIDC token.

## 0.8.2

The npm package is `@aptd3v/containerkit`. The install includes the runner. The `containerkit` command selects the binary for this machine.

## 0.8.1

`pkg/kit` is one vocabulary for Engine and Compose. The same names exist in TSX. The license is Apache 2.0.

- `Connect` dials Engine and Compose together. `NewEngine` and `NewComposer` remain when a program needs one SDK.
- `NewContainer`, `NewService`, and `NewCompose` take the same setters. A `*Container` is a `ServiceOption`. `AsService` copies a container onto a Compose service. `Validate` names each Engine-only field that has no Compose mapping: `AutoRemove`, `PublishAll`, `ContainerIDFile`, `Shell`, `OnBuild`, `MaskedPath`, `ReadonlyPath`.
- `ImportFile` reads compose YAML with no daemon. `ImportPlan` and `ExecJSON` run the document the TSX package emits. `Client.Load` loads a file with a daemon.
- Vocabulary packages hold the long tail of finite values: `pkg/proto`, `pkg/cap`, `pkg/restart`, and the rest. `TCP`, `Always`, `Bridge`, and the other common values stay in `pkg/kit`.
- `pkg/logic` is `When`, `Switch`, and `Group`. The kit does not import it.
- `pkg/taskflow` is an optional DAG. Importing it registers `ExecFlow` and `ValidateFlow`. A program that never imports it never links `go-taskflow`.
- `pkg/terminal` copies a hijacked stream. `AttachTerminal` and `ExecTerminal` open it.
- Runtime failures are `*kit.Error`. `errors.Is` and `IsNotFound` match the SDK sentinels.
- `internal/` catalogs the pinned upstream types under their original names. `go generate` in `pkg/kit` runs `gencatalog` and `gentsx`.
- The TypeScript package is `tsx/`. Components and session methods are generated from the kit. `exportYAML` and `exportDockerfile` do not dial. `docker.ts`, `docker.tsx`, or `docker.js` default-exports `root()` and keeps a session open. A callback with no parameters applies the project when the rendered YAML changes. An async callback, or one that takes the session, runs once. `onExit: "down"` takes the last project down on SIGINT and SIGTERM. `onUp` runs after a successful apply, and `wait: true` waits until containers are running and healthy. An attached compose up follows container logs as session events. `detach: true` does not.
- Closed vocabularies on TSX props are unions of those values: protocol, restart, capabilities, signals, progress, and the rest of `@aptd3v/containerkit/vocab`. `pullPolicy`, `networkMode`, `ipc`, and `pid` also accept the prefix the helpers return. `mode` and `driver` stay strings. `StopSignal` on a Dockerfile takes a POSIX signal name.
- Session results are typed for `login`, `diskUsage`, `removeImage`, and the config, secret, plugin, node, service, task, checkpoint, and swarm lists and inspects. `pullImage`, `pushImage`, `buildImage`, `composeBuild`, `logs`, `attachContainer`, `statsContainer`, `copyFrom`, `saveImages`, and `run()` stay untyped.
- `@aptd3v/containerkit/hooks` is `useEnv`, `useLiveLog`, `useServiceState`, and `useShutdown`.
- `@aptd3v/containerkit/eslint` marks an inline child the parent does not accept. `tsc` still accepts that child. A child held in a variable stays invisible to the rule.
- `cmd/containerkit` is `version`, `validate`, `exec`, `session`, and `terminal`. A `v*` tag that matches `VERSION` publishes the six platform archives and the npm packages. `NOTICE` lists the Apache modules linked into the binary. See [the runner](docs/runner-binary.md).
- The TypeScript programs you run are nginx, WordPress and MySQL, a MongoDB replica set, and an interactive shell, under `tsx/examples`. `examples/supabase` and `examples/taskflow` stay Go programs.

Pins verified 2026-11-12: `github.com/moby/moby/api` v1.56.1, `github.com/moby/moby/client` v0.6.2, `github.com/compose-spec/compose-go/v2` v2.16.2, `github.com/docker/compose/v5` v5.6.1.
