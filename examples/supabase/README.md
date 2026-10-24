# Supabase example

One-command local Supabase stack on [containerkit](https://github.com/aptd3v/containerkit): in-process Compose (no `docker compose` CLI), profiles, embedded volume files, and live events.

## Requirements

- Go 1.26+
- Docker Engine

## Run (from repo root)

```bash
go run ./examples/supabase/                    # minimal stack
go run ./examples/supabase/ -profile full      # full stack
go run ./examples/supabase/ -resource-limits   # memory/CPU limits on db, kong, studio
go run ./examples/supabase/ -profile full -resource-limits
go run ./examples/supabase/ -volumes-path /path/to/volumes
```

First `Up` can take several minutes (Postgres init start period is 120s; wait timeout is 300s).

## Open

| What | URL | Auth |
| --- | --- | --- |
| Studio | http://localhost:3000 | none (Studio talks to Kong/meta with generated keys) |
| API (Kong) | http://localhost:8000 | Kong dashboard basic auth `kong` / `kong` |
| Kong HTTPS | https://localhost:8443 | same |
| Postgres (full profile) | `localhost:5432` | `postgres` / `postgres` |
| Pooler (full profile) | `localhost:6543` | tenant `default` |

JWT secret, anon key, and service-role key are generated on every process start and injected into the containers. Studio does not need you to paste them. For a client SDK, read `SUPABASE_ANON_KEY` from the `studio` or `kong` container env.

Reusing `./volumes/db/data` across runs with a new JWT secret will break Auth. Wipe volumes (below) if login or JWTs fail.

## Flags

| Flag | Description |
| --- | --- |
| `-profile` | `minimal` (default) or `full`. |
| `-resource-limits` | Memory/CPU limits: db 2GiB / 2 CPU; kong and studio 512MiB / 1 CPU each. |
| `-volumes-path` | Host directory for bind-mounted config. Default: `./volumes` or `SUPABASE_VOLUMES_PATH`. |

Environment (optional): `SUPABASE_PROFILE`, `SUPABASE_RESOURCE_LIMITS=1`, `SUPABASE_VOLUMES_PATH`.

## Profiles

`Up` / `Logs` / `Events` / `Kill` / `Down` always pass `minimal`. `-profile full` appends `full`.

| Profile | Services |
| --- | --- |
| `minimal` | db, analytics, auth, rest, **meta**, kong, studio |
| `full` | vector, realtime, storage, imgproxy, functions, supavisor |

`meta` (postgres-meta) stays on `minimal` because Studio calls `http://meta:8080`. A stack without it fails Studio with `ENOTFOUND meta`.

## What it does

1. **Bootstrap** — writes embedded Kong, SQL, Vector, pooler, and edge-function files under the volumes path. No network fetch.
2. **Validate / export** — `project.Validate()`, then `Export` to `./docker-compose.yaml` at the repo root.
3. **Up** — detached Compose `Up` with `RemoveOrphans` for the selected profiles.
4. **Events + logs** — compose events as cyan JSON on stdout, then `Logs` follow.
5. **Ctrl+C** — `Kill` (SIGKILL) then `Down` (orphans and named volumes). Bind-mount files under `./volumes` stay on disk.

## Clean run

Named volumes (`db-config`, `deno-cache`) are removed on Ctrl+C. Postgres data is a bind mount and is not.

```bash
rm -rf volumes
go run ./examples/supabase/
```

## containerkit features used

- **Profiles** — `Service.Profiles("minimal")` / `"full"` mirrored on `Up.Profiles` (and Logs/Events/Kill/Down).
- **DependsOn** — Studio waits for `meta`; most services wait for healthy `db` / `analytics`.
- **Conditional limits** — ordinary `if` plus `kit.Deploy{Limits: &types.Resources{...}}`.
- **Health** — `kit.NewHealth().StartPeriod(120)` on db.
- **Compose API** — `Up`, `Logs`, `Events`, `Kill`, `Down` with profile fields on option structs.
- **Validation** — `project.Validate()` before up.
