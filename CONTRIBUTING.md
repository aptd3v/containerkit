# Contributing

Go 1.26.11 or newer, from `go.mod`. Node 20 or newer for the TypeScript package. Python 3 for `scripts/version.py`, `scripts/pins.py`, and `scripts/ci/kitcover.py`. Docker is required for the Engine job. The `//go:build e2e` suites stay optional.

## Check

From the repository root, the same commands CI runs:

```bash
python3 scripts/version.py --check
python3 scripts/pins.py --check
go test -race -shuffle=on ./...
go test -coverprofile=kit.out ./pkg/kit
python3 scripts/ci/kitcover.py --floor 50 kit.out
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

`go test ./...` does not dial a daemon. The Engine job does. The suites under `e2e/` run with `-tags e2e` against a live Engine and stay optional. Each package comment says which cluster it expects.

From `tsx/`:

```bash
npm audit --package-lock-only --include=dev --audit-level=high
```

`high` and `critical` fail **Advisory**. `--include=dev` keeps dev dependencies in the audit on a production-configured runner. Dependabot opens the update pull requests. It is not this gate. The Advisory job runs `govulncheck` with the Go version in `go.mod`. A newer local toolchain can report standard-library findings that job does not.

The Engine job pins Docker Engine server `29.9.3`, prints `docker version`, and runs `scripts/ci/engine.sh`. That printed server version is the compatibility claim. A second Engine version is out of scope. `tsx/docs/client.md` may still pin `apiVersion` or negotiate. On Ubuntu the job installs that server with `scripts/ci/pin-engine.sh` when the daemon is not already `29.9.3`.

```bash
go run ./tools/crossbuild
cd tsx
npm ci
npm run build
cd ..
ENGINE_VERSION=29.9.3 bash scripts/ci/engine.sh
```

`dist/containerkit` is the host binary `tools/crossbuild` wrote. The script deletes `alpine:3.22`, runs `pullImage`, and checks the image exists before `composeUp` of `scripts/ci/ck-health.yaml`. `ready` sleeps 3 seconds, then serves. `check` reads that URL once at startup and exits if the read fails. `compose ps` must list both services. Exec of `echo` must return `0`. Logs must contain `ck-check-marker`. Docker must show the project container and the project network before `composeDown`, and neither after, including stopped containers. `tsx/testdata/plan.json` is validated and is not that stack. When that file is one detached `composeUp` for `app` / `api` on `alpine:latest`, the script executes it. The copy in this tree is the edge-case plan, so the script validates it and executes `tsx/testdata/plan-run.json` instead, confirms `app` / `api` with `docker ps -a`, then removes project `app`. `tsx/ci/docker.js` is the live entry. `CONTAINERKIT_BIN` is the absolute path of `dist/containerkit`. The script waits for `ck-live-up`, sends SIGTERM to Node, and fails if a project container or the project network is still there. Fallback cleanup of `ck-live` runs only after that check.

Taskflow coverage in CI is 100% of statements. That gate is separate from `pkg/kit`:

```bash
go test -coverprofile=taskflow.out ./pkg/taskflow
go tool cover -func=taskflow.out
```

## pkg/kit coverage

The Go test job writes `kit.out` and uploads it as `pkg-kit-cover`. The floor is 50% of statements in the hand-written `pkg/kit` files. `pkg/kit/tsx_plan_gen.go` stays in the profile and is not part of the floor. That file is the generated decoder. `scripts/ci/kitcover.py` is the check.

`tsx/testdata/plan.json` is an edge-case sample read by `TestImportPlan` and `cmd/containerkit/run_test.go`. It exercises `pullImage`, `listImages`, `listContainers`, `inspectImage`, `inspect`, `load`, `composePs`, `composePort`, `composeKill`, `composeUp`, `composeDown`, `ping`, and `composeVolumeList`. The generated decoder dispatches many more operations. Citing that file alone overstates decoder coverage. `tsx/testdata/plan-run.json` is the one-action `composeUp` the Engine job executes. It is not the decoder either. Other decoder paths are the inline plans and `decode*` calls in `pkg/kit/tsx_plan_test.go` (`TestDecodeConnectAndUpOptions`, `TestDecodeVocabOptions`, `TestDecodeRejectsBadTimeout`, `TestDockerfileTSXPlan`, `TestPlanValidateRejectsBadNumbers`, `TestDrainedStreamPlansReturnOK`, `TestExecDispatchesKitVerbs`). `tsx/testdata/flow.json` is the taskflow document, covered by the taskflow gate.

The TypeScript package:

```bash
cd tsx
npm ci
npm test
npm run test:taskflow
```

`npm test` typechecks the library, the tests, and the React coexist fixture, then runs the compiled tests.

## Upstream pins

`go.mod` is the source of truth for the Moby API, the Moby client, compose-go, and Compose. `pkg/kit/doc.go` and the blob URLs in `tools/gencatalog/main.go` name the same versions. The historical pin line in `CHANGELOG.md` is not part of the check.

```bash
python3 scripts/pins.py --check
```

Every pull request runs that command as **Upstream pin agreement**.

**Upstream pin freshness** is a separate workflow. It runs Monday at 15:00 UTC and can be started by hand. It is not a pull-request check. `python3 scripts/pins.py --freshness` compares the four modules with the latest tags. When one is behind, the job opens or updates the issue "Upstream module pins are behind" and fails. When the four pins match the latest tags, that issue is closed.

```bash
python3 scripts/upstream.py
python3 scripts/upstream.py compose-go
python3 scripts/upstream.py compose
python3 scripts/upstream.py engine
```

`scripts/upstream.py` prints the release notes, commits, and file diff from the selected pin to the latest tag. `compose-go` is compose-go. `compose` is the Compose SDK. `engine` is the Engine API and the Engine client. `--base vX.Y.Z` compares that tag with the latest tag of one module.

As of 2026-11-16, the Moby API v1.57.0, the Moby client v0.6.3, compose-go v2.17.0, and Compose v5.7.1 match their latest tags.

## Generate

`pkg/kit` carries the generate directives. From the repository root:

```bash
go generate ./pkg/kit
```

`gencatalog` rewrites `internal/` from the modules in `go.mod`. `gentsx` rewrites `tsx/src/generated.ts`, `tsx/src/generated-options.ts`, and `tsx/src/generated-actions.ts`. Commit the result in the same change as the module bump. `internal/catalogtest` fails when an alias renames an upstream type.

## Version

`VERSION` is the source of truth. `scripts/version.py` copies it into `tsx/package.json`, `tsx/package-lock.json`, `cmd/containerkit/main.go`, and `npm/@aptd3v/containerkit-*/package.json`.

```bash
python3 scripts/version.py --check
python3 scripts/version.py patch
```

`patch`, `minor`, and `major` refuse to run when those copies already disagree.

## What stays local

`.gitignore` already excludes `playground/`, `dist/`, `tsx/node_modules/`, `tsx/dist/`, the platform binaries under `npm/@aptd3v/containerkit-*/bin/`, and `.containerkit/`. `go run ./tools/crossbuild` fills `dist/` and those `bin/` directories. Do not commit them.

Publishing is the tag workflow in `.github/workflows/release.yml`. The tag must equal `VERSION` with a leading `v`. The workflow authenticates to npm with GitHub's OIDC token.
