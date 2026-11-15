# Supabase example

A self-hosted [Supabase](https://supabase.com/docs/guides/self-hosting/docker) stack defined in Go with containerkit. `stack.go` builds the project with `logic.Group`, `logic.Switch`, `logic.When`, and `logic.Each`: the gateway, Postgres, pooler, storage, and logs choices are option groups, and the database init files are one mount list. `main.go` embeds `docker/`, generates keys, and runs an attached `Up` so container logs print as each service starts. Ctrl+C then runs `Kill` / `Down`.

`docker/` holds the upstream SQL, gateway config, and scripts from `self-hosted/v0.8.4`. The program writes that tree to a temp directory because the stack bind-mounts those files. The temp directory is outside this repo, so Docker is not asked to mount the project path.

## Requirements

- Go 1.27+
- Docker Engine
- `openssl` (key generation)
- `node` 16+ or Docker (asymmetric API keys)

## Run (from repo root)

```bash
go run ./examples/supabase/
```

```bash
go run ./examples/supabase/ -gateway kong -postgres 17 -pooler supavisor -storage file -logs
```

| Flag | Default | Values |
| --- | --- | --- |
| `-gateway` | `envoy` | `envoy`, `kong`, `caddy`, `nginx` |
| `-postgres` | `17` | `15`, `17` |
| `-pooler` | `supavisor` | `supavisor`, `pgbouncer` |
| `-storage` | `file` | `file`, `s3`, `rustfs` |
| `-logs` | off | Logflare and Vector |

Ctrl+C kills the containers, removes Compose volumes, and deletes the temp directory.

## Open

| What | URL | Auth |
| --- | --- | --- |
| Studio and API | http://localhost:8000 | dashboard user `supabase` and the generated `DASHBOARD_PASSWORD` |
| Postgres | localhost:5432 | generated `POSTGRES_PASSWORD`, through Supavisor |

Caddy and nginx listen on 80 and 443 instead, and PgBouncer listens on `POOLER_PROXY_PORT_TRANSACTION`. The generated keys are in the temp `.env`.
