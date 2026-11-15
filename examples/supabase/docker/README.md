# Supabase files for the containerkit example

This directory is the data the Go example mounts: database init SQL, Envoy and Kong config, Caddy and nginx templates, pooler and Vector config, edge functions, and the key-generation scripts. It comes from the upstream [self-hosted Docker tree](https://supabase.com/docs/guides/self-hosting/docker) at `self-hosted/v0.8.4`. The stack Compose files are gone. `dev/` and `tests/` still have upstream overlays; the example does not load them.

The stack is not defined here. [`../stack.go`](../stack.go) builds it with containerkit: images, health checks, `DependsOn`, published ports, bind mounts, named volumes, network aliases, memory limits, and one-shot bucket jobs. [`../main.go`](../main.go) embeds this directory, writes it to a temp dir Docker can mount, generates keys into `.env`, and runs an attached `Up` so container logs print as each service starts. Ctrl+C then runs `Kill` and `Down`.

## Run

From the repository root:

```bash
go run ./examples/supabase/
```

```bash
go run ./examples/supabase/ -gateway kong -postgres 15 -pooler pgbouncer -storage s3 -logs
```

Ctrl+C kills the containers, removes the Compose volumes, and deletes the temp directory.

| Flag | Default | Replaces |
| --- | --- | --- |
| `-gateway` | `envoy` | `envoy`, `kong`, `caddy`, or `nginx` |
| `-postgres` | `17` | `15` or `17` |
| `-pooler` | `supavisor` | `supavisor` or `pgbouncer` |
| `-storage` | `file` | `file`, `s3` (MinIO), or `rustfs` |
| `-logs` | off | Logflare and Vector |

`-gateway caddy` and `-gateway nginx` publish 80/443 and leave the API gateway unpublished, because the proxy is the listener. `-gateway kong` publishes the HTTP port and 8443. Envoy and Kong both join the project network as `envoy` and `kong`, so the rest of the stack resolves either name.

Studio and the API are served by the gateway at `http://localhost:8000` (or `API_GW_HTTP_PORT`). The dashboard user is `supabase`. Postgres is published by Supavisor on `POSTGRES_PORT` (5432). PgBouncer publishes `POOLER_PROXY_PORT_TRANSACTION` instead. Passwords and API keys are written to the temp `.env` by `utils/generate-keys.sh` and `utils/add-new-auth-keys.sh`.

## Services

Studio, Auth, PostgREST, Realtime, Storage, imgproxy, postgres-meta, PostgreSQL, Edge Runtime, and a pooler always start. Logflare, Vector, MinIO, and RustFS start only for the matching flag. Each init SQL file is mounted on its own path under `/docker-entrypoint-initdb.d`, so the image's own migrations stay in place.
