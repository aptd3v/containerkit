#!/usr/bin/env bash
# Bring up 4 lightweight nodes with Docker Engine (QEMU VMs preferred, else DinD).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

ensure_dirs
if cluster_reachable; then
	backend="$(running_backend)"
	if [[ -z "$backend" ]]; then
		backend=dind
		if qemu_available && ! docker_available; then
			backend=qemu
		fi
	fi
	printf '%s\n' "$backend" >"$RUN_DIR/backend"
	write_cluster_env "$backend"
	log "cluster already reachable via ${backend}; not starting a second fabric"
	exit 0
fi

backend="$(detect_backend)"
log "bringing up 4-node swarm cluster via ${backend}"
printf '%s\n' "$backend" >"$RUN_DIR/backend"
exec "$ROOT/${backend}/up.sh"
