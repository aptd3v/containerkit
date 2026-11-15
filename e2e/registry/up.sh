#!/usr/bin/env bash
# Bring up a Docker Engine plus an authenticated registry (QEMU preferred, else DinD).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

ensure_dirs
if engine_reachable && wait_registry_auth "http://${HOST_BIND}:${REGISTRY_PORT}/v2/" 2; then
	backend="$(running_backend)"
	if [[ -z "$backend" ]]; then
		backend="$(detect_backend)"
	fi
	printf '%s\n' "$backend" >"$RUN_DIR/backend"
	if [[ "$backend" == qemu ]]; then
		write_registry_env qemu "tcp://${HOST_BIND}:${DOCKER_PORT}" "127.0.0.1:5000"
	else
		write_registry_env dind "tcp://${HOST_BIND}:${DOCKER_PORT}" "registry:5000"
	fi
	log "registry already reachable via ${backend}; not starting a second fabric"
	exit 0
fi

backend="$(detect_backend)"
log "bringing up authenticated registry via ${backend}"
printf '%s\n' "$backend" >"$RUN_DIR/backend"
exec "$ROOT/${backend}/up.sh"
