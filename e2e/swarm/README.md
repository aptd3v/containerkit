# Swarm e2e cluster

Four lightweight nodes, each running Docker Engine, used to e2e-test `pkg/kit` Swarm APIs (`InitSwarm`, `JoinSwarm`, `ListNodes`, `CreateSwarmService`).

## Layout

| Node | Role (after tests) | LAN IP     | Docker API (host)        |
| ---- | ------------------ | ---------- | ------------------------ |
| mgr  | manager            | 10.0.0.11  | `tcp://127.0.0.1:23751`  |
| w1   | worker             | 10.0.0.12  | `tcp://127.0.0.1:23752`  |
| w2   | worker             | 10.0.0.13  | `tcp://127.0.0.1:23753`  |
| w3   | worker             | 10.0.0.14  | `tcp://127.0.0.1:23754`  |

Bring-up starts Engine only. The e2e test initializes the swarm through kit.

## Backends

1. **QEMU** (preferred): 4 cloud VMs, ~768MB RAM each, user-mode hostfwd for SSH/API plus a multicast Ethernet fabric for node-to-node swarm traffic.
2. **DinD** (similar, used when QEMU is missing): 4 privileged `docker:27.5.1-dind` containers on a `10.0.0.0/24` compose network.

```bash
./e2e/swarm/preflight.sh          # qemu vs docker vs missing tools
./e2e/swarm/up.sh                 # start 4 nodes
./e2e/swarm/status.sh             # ping forwarded APIs
go test -tags e2e -count=1 ./e2e/swarm
./e2e/swarm/down.sh               # stop the 4 nodes this project started
./e2e/swarm/down.sh --wipe        # also drop node disks (qemu) or compose volumes (dind)
```

Force a backend with `CONTAINERKIT_SWARM_BACKEND=qemu` or `dind`.

Install QEMU on macOS with `brew install qemu`. First QEMU boot downloads a Debian cloud image and installs Docker inside each VM (several minutes).

Default `go test ./...` does not run these tests (`//go:build e2e`).
