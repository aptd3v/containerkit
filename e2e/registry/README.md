# Registry auth e2e

One Docker Engine plus a private `registry:2` that requires htpasswd credentials. Used to e2e-test `pkg/kit` `Login`, authenticated `PushImage`, and authenticated `PullImage`.

## Layout

| Role     | Address (Engine's view) | Host forward                 |
| -------- | ----------------------- | ---------------------------- |
| Engine   | unix + `tcp://0.0.0.0:2375` | `tcp://127.0.0.1:23755` |
| Registry | `127.0.0.1:5000` (QEMU) or `registry:5000` (DinD) | `http://127.0.0.1:50005/v2/` |

Default credentials: user `containerkit`, password `containerkit-e2e`.

## Backends

1. **QEMU** (preferred): one Debian cloud VM, Docker Engine, `registry:2` with htpasswd, insecure-registries for `127.0.0.1:5000`.
2. **DinD** (similar, used when QEMU is missing): privileged `docker:27.5.1-dind` plus `registry:2` on a compose network.

```bash
./e2e/registry/preflight.sh
./e2e/registry/up.sh
./e2e/registry/status.sh
go test -tags e2e -count=1 ./e2e/registry
./e2e/registry/down.sh
./e2e/registry/down.sh --wipe
```

Force a backend with `CONTAINERKIT_REGISTRY_BACKEND=qemu` or `dind`.

Install QEMU on macOS with `brew install qemu`. First QEMU boot downloads a Debian cloud image, installs Docker, and pulls `registry:2` (several minutes).

Default `go test ./...` does not run these tests (`//go:build e2e`).
