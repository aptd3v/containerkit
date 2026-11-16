# Contributing

Go 1.26.8 or newer, from `go.mod`. Node 20 or newer for the TypeScript package. Python 3 for `scripts/version.py`. Docker is required only for the `e2e` build tag.

## Check

From the repository root, the same commands CI runs:

```bash
python3 scripts/version.py --check
go test -race -shuffle=on ./...
```

`go test ./...` does not dial a daemon. The suites under `e2e/` run with `-tags e2e` against a live Engine. Each package comment says which cluster it expects.

Taskflow coverage in CI is 100% of statements:

```bash
go test -coverprofile=taskflow.out ./pkg/taskflow
go tool cover -func=taskflow.out
```

The TypeScript package:

```bash
cd tsx
npm ci
npm test
npm run test:taskflow
```

`npm test` typechecks the library, the tests, and the React coexist fixture, then runs the compiled tests.

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

Publishing is the tag workflow in `.github/workflows/release.yml`. The tag must equal `VERSION` with a leading `v`. The repository secret `NPM_TOKEN` is what the workflow uses for npm.
