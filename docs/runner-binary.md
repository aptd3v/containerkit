# Runner binary

`containerkit` (`cmd/containerkit`) runs a plan JSON or a compose YAML against a Docker daemon. Authoring TSX still needs Node; executing a plan does not need the Go toolchain once this binary is on `PATH`.

Pushing a tag `vX.Y.Z` that matches `VERSION` runs `.github/workflows/release.yml`. That workflow tests, builds the six archives, publishes `containerkit` and `@containerkit/<os>-<arch>` to npm, then opens the GitHub release. The repository secret `NPM_TOKEN` is the npm auth token. A local build does not publish.

## Build all platforms

From the repo root (needs Go, no extra modules):

```bash
go run ./tools/crossbuild
```

Writes `dist/`:

| File | Notes |
| --- | --- |
| `containerkit_<os>_<arch>.tar.gz` | linux and darwin, amd64 and arm64 |
| `containerkit_windows_<arch>.zip` | amd64 and arm64 |
| `checksums.txt` | SHA-256 of the six archives |
| `containerkit` | host `GOOS/GOARCH` binary (`containerkit.exe` on Windows) |

Archives are named `containerkit_<os>_<arch>`. The file inside each archive is `containerkit` (`containerkit.exe` on Windows).

`CGO_ENABLED=0`. Builds without ldflags stamp `cmd/containerkit`'s `main.version`. `scripts/version.py` keeps that fallback, `VERSION`, and the TSX package on the same semver (`patch`, `minor`, `major`, or an explicit version). `tools/crossbuild` reads `VERSION` unless `CONTAINERKIT_VERSION` is set, and stamps `git rev-parse HEAD` unless `CONTAINERKIT_COMMIT` is set.

GoReleaser uses the same matrix. A snapshot stays on this machine:

```bash
goreleaser release --snapshot --clean --skip=publish
```

`node tsx/scripts/publish-optional.mjs check` confirms `VERSION`, `tsx/package.json`, and the six platform binaries agree. `publish` sends those packages to npm and needs `NODE_AUTH_TOKEN` or `NPM_TOKEN`. The tag workflow is the path that publishes.

## Commands

```bash
containerkit version
containerkit validate tsx/testdata/plan.json
containerkit validate tsx/testdata/example.yaml
containerkit exec plan.json
echo "$json" | containerkit exec -
containerkit session
```

`validate` calls `ImportPlan.Validate` or `ImportYAML.Validate` and does not dial the daemon. `exec` does: JSON is a TSX plan (`encode(composeUp(...))`); YAML is treated as compose (connect from env, then `composeUp`).

`session` keeps one process open until stdin reaches EOF or the process is signaled. Stdin and stdout are JSON, one object per line. A blank line is ignored. Diagnostics go to stderr.

A call:

```json
{"id": 1, "type": "exec", "plan": {}}
```

`plan` is the same document `exec` accepts. The reply repeats `id`.

Success:

```json
{"id": 1, "type": "result", "ok": true, "value": {}}
```

Failure:

```json
{"id": 1, "type": "error", "class": "notFound", "message": "no such container"}
```

`class` is a stable string such as `notFound`, `conflict`, or `invalidArgument`. `status`, `hint`, `receiver`, `path`, and `exitCode` are present only when the error has them.

Progress is an event, not a result. `kind` is `log` or `status`.

```json
{"type": "event", "kind": "log", "source": "api", "stream": "stdout", "line": "ready"}
```

The first plan dials the daemon. A later plan reuses that client while its `connect` object is unchanged. `containerkit/live` is the TypeScript side of this pipe.

## From npm

`npm install containerkit` installs the TypeScript library and the one platform package that matches the machine, for example `@containerkit/darwin-arm64`. The other five stay optional and are skipped. The library's `bin` entry is a small script. npm links it as `node_modules/.bin/containerkit`, and that script runs the Go binary from the platform package. `npm run` puts `.bin` on `PATH`, so a package that depends on containerkit can call `containerkit` from its scripts.

The packages in this repo depend on those platform folders with `file:`. `prepack` rewrites those paths to the release version so a published install fetches them from the registry. `go run ./tools/crossbuild` copies each binary into `npm/@containerkit/<os>-<arch>/bin/`.

## From TypeScript

```ts
import { binaryInvoke, run, setInvoke, composeUp, connect } from "containerkit";

process.env.CONTAINERKIT_BIN = "./dist/containerkit";
await run(composeUp(connect({ fromEnv: true }), tree, { detach: true }));

setInvoke(binaryInvoke("./dist/containerkit"));
```

`run()` uses `setInvoke` when set, otherwise spawns `CONTAINERKIT_BIN` or `containerkit` on `PATH`. `encode()` stays the no-I/O path.
