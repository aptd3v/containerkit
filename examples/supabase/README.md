# Supabase example

One-execute Supabase stack using [go-contain](https://github.com/aptd3v/go-contain): programmatic Compose with profiles, resource limits, embedded config, and real-time events.

## Requirements

- Go 1.21+
- Docker and Docker Compose

## Run (from repo root)

```bash
go run ./examples/supabase/                    # minimal stack, no resource limits
go run ./examples/supabase/ -profile full     # full stack
go run ./examples/supabase/ -resource-limits  # minimal with memory/CPU limits on db, kong, studio
go run ./examples/supabase/ -profile full -resource-limits
go run ./examples/supabase/ -volumes-path /path/to/volumes
```

## Flags


| Flag               | Description                                                                                                                                                  |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `-profile`         | `minimal` (default) or `full`. Minimal = db, analytics, kong, studio, auth, rest. Full adds vector, realtime, storage, imgproxy, meta, functions, supavisor. |
| `-resource-limits` | Apply memory/CPU limits to db (2GiB, 2 CPU), kong and studio (512MiB, 1 CPU each).                                                                           |
| `-volumes-path`    | Directory for volume files. Default: `./volumes` or `SUPABASE_VOLUMES_PATH`.                                                                                 |


Environment variables (optional): `SUPABASE_PROFILE`, `SUPABASE_RESOURCE_LIMITS=1`, `SUPABASE_VOLUMES_PATH`.

## What it does

1. **Bootstrap**: Writes embedded config (Kong, DB SQL, Vector, pooler, edge function) under the volumes path. No network fetch; all files are in the binary.
2. **Up**: Starts the stack with the chosen profile. DB has a 120s health start period for init.
3. **Events**: Streams compose events as JSON (cyan) to stdout alongside container logs.
4. **Ctrl+C**: Sends SIGKILL to containers, then runs `down` (remove orphans, volumes).

## Clean run

To reset DB and config and start fresh:

```bash
rm -rf volumes
go run ./examples/supabase/ -profile minimal
```

## go-contain features used

- **Profiles** — `sc.WithProfiles("minimal")` / `"full"` and `up.WithProfiles(...)` for run modes.
- **Conditional resource limits** — `tools.WhenTrue(enableResourceLimits, sc.WithDeploy(deploy.WithResourceLimits(...)))` on db, kong, studio.
- **Health and deploy** — `health.WithStartPeriod`, `resource.WithMemoryBytes`, `resource.WithNanoCPUs`.
- **Compose API** — `Up`, `Logs`, `Events`, `Kill`, `Down` with profile and option setters.
- **Validation** — `project.Validate()` before up.

