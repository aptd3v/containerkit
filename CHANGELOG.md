# Changelog

## 0.8.1

One vocabulary for the Engine SDK and the Compose SDK, in Go and in TSX. `v0.7.2` programs that import `pkg/client` or the old kit setters do not compile against this release.

- `pkg/kit` defines a container, a service, and a Compose project with the same setters. `Connect` dials both SDKs. `AsService` copies a container onto a service. `Validate` reports Engine-only fields that have no Compose mapping.
- `pkg/client` is gone. Hijacked terminal I/O lives in `pkg/terminal`.
- Finite Docker values live in small packages (`pkg/proto`, `pkg/cap`, `pkg/restart`, and the rest). The common values stay on the kit.
- `internal/` catalogs the pinned Moby, Compose, OCI, and reference types under their original names. `gencatalog` and `gentsx` regenerate the catalog and the TSX components.
- The TypeScript package renders Compose YAML and Dockerfiles, and a live `docker.ts` applies the project only when the rendered YAML changes.
- `cmd/containerkit` validates a plan without a daemon, execs one, or keeps a session open. Tagging `v0.8.1` publishes the six platform archives to GitHub and the npm packages.
- `pkg/taskflow` is an optional DAG. Importing it registers the kit hooks.
- The Astronomy Shop and the small demos are gone. `examples/supabase` and `examples/taskflow` are the programs that show the kit.

Install the module with `go get github.com/aptd3v/containerkit@v0.8.1`. Install the TypeScript package with `npm install containerkit`.
