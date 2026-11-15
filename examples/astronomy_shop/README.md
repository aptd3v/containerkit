# Astronomy Shop (OpenTelemetry demo)

Flagship [containerkit](https://github.com/aptd3v/containerkit) showcase: the CNCF [OpenTelemetry Astronomy Shop](https://github.com/open-telemetry/opentelemetry-demo) defined entirely in Go. It pulls published `ghcr.io/open-telemetry/demo` images, uses Compose profiles and health waits, runs a taskflow `Build -> Up -> probe` DAG, and builds a small **kit-probe** image with Dockerfile-as-Go (no vendored upstream `src/`).

## Requirements

- Go 1.26+
- Docker Engine
- Works on **macOS** and **Linux** (native or Docker Desktop).
- RAM: about 3 GB for `shop`, about 6 GB for `shop,observability` (default). Add more for `full` (Kafka).


On some **Linux** nested or containerized Docker hosts, same-bridge traffic is
dropped when `net.bridge.bridge-nf-call-iptables=1`.
If Linux services stay unhealthy, set:

```bash
sudo sysctl -w net.bridge.bridge-nf-call-iptables=0 net.bridge.bridge-nf-call-ip6tables=0
```

## Run (from repo root)

```bash
# Store + Grafana/Jaeger/Prometheus (default)
go run ./examples/astronomy_shop/

# Store only
go run ./examples/astronomy_shop/ -profile shop

# Store + Kafka accounting / fraud path
go run ./examples/astronomy_shop/ -profile shop,full,observability

# Long-running: Events + Logs until Ctrl+C (then Kill/Down)
go run ./examples/astronomy_shop/ -follow
```

Default mode validates, exports compose YAML, builds `astronomy-shop-kit-probe:local` if missing, runs Compose `Up` with `Wait` (up to 15 minutes), probes the storefront, prints URLs, then stays up until Ctrl+C (then `Down`). `-follow` is the same lifecycle plus compose events and logs.

## Open

| What | URL | Profile |
| --- | --- | --- |
| Store (Envoy) | http://localhost:8080/ | shop |
| Load generator (Locust) | http://localhost:8089/ | shop |
| Grafana | http://localhost:8080/grafana/ | observability |
| Jaeger UI | http://localhost:16686/jaeger/ui | observability |
| Prometheus | http://localhost:9090/ | observability |

Images are pinned to `IMAGE_VERSION=3.1.0` (`ghcr.io/open-telemetry/demo:3.1.0-*`).

## Flags

| Flag | Description |
| --- | --- |
| `-profile` | Comma-separated: `shop`, `full`, `observability`. Default `shop,observability`. `shop` is always included. |
| `-follow` | After `Up`, also stream compose events and logs until Ctrl+C. |
| `-volumes-path` | Host directory for bind-mounted configs. Defaults to a `volumes` dir next to the example source (portable across working directories), or `ASTRONOMY_SHOP_VOLUMES_PATH`. |
| `-export` | Write resolved compose YAML (default `./astronomy-shop.compose.yaml`; empty skips). |

## Profiles

`Up` / `Logs` / `Events` / `Kill` / `Down` always receive the active profile list (same rule as the Supabase example).

| Profile | Services (approx.) | Intent |
| --- | --- | --- |
| `shop` | ad, cart, checkout, currency, email, frontend, frontend-proxy, image-provider, load-generator, payment, product-catalog, quote, recommendation, shipping, flagd, flagd-ui, astronomy-db, valkey-cart, otel-collector | Store (~3 GB) |
| `full` | + kafka, accounting, fraud-detection | Kafka event path |
| `observability` | + jaeger, grafana, prometheus | Traces and dashboards (~6 GB with shop) |

OpenSearch and OpAMP from upstream `compose.observability.yaml` are omitted to keep embeds and RAM lighter. Jaeger, Grafana, and Prometheus cover the traces path.

## What it does

1. **Bootstrap**: writes embedded flagd, Postgres init, collector, Jaeger, Grafana, and Prometheus configs under the volumes path (offline).
2. **Validate / export**: `project.Validate()`, then optional `Export`.
3. **taskflow** (default): `Build(kit-probe)` -> `Compose(Up{Detach, Wait, Profiles})` -> `Container(kit-probe)` with `Retry` -> print URLs -> Ctrl+C -> `Down`.
4. **`-follow`**: `Up` -> Events + Logs -> Ctrl+C -> `Kill` / `Down` (no kit-probe).
5. **Teardown**: profiled `Down` plus `flow.Remove` for the probe container.

## containerkit features used

- **Profiles**: `Service.Profiles(...)` mirrored on every Compose op.
- **DependsOn / DependsOnHealthy**: store DAG matches upstream health gates.
- **Dockerfile-as-Go**: `kit.NewDockerFile` + `flow.Build` for kit-probe.
- **taskflow**: ordered Build -> Up -> probe with retry and Dump.
- **Up.Wait**: blocks until healthy before probing.
- **Compose API**: Up, Logs, Events, Kill, Down.

## Clean run

If a previous run created an IPv6-enabled `astronomy-shop` network, remove it so Compose recreates it IPv4-only:

```bash
rm -rf examples/astronomy_shop/volumes astronomy-shop.compose.yaml
docker network rm astronomy-shop 2>/dev/null || true
go run ./examples/astronomy_shop/ -follow
```
