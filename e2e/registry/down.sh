#!/usr/bin/env bash
# Tear down the registry e2e environment started by up.sh.
# Stops only containerkit-registry QEMU PIDs or the containerkit-registry Compose project.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

wipe=0
if [[ "${1:-}" == "--wipe" ]]; then
	wipe=1
fi

backend="$(running_backend)"
if [[ -z "$backend" ]]; then
	if qemu_available && [[ -f "$RUN_DIR/registry.pid" ]]; then
		backend=qemu
	elif docker_available && docker compose -p "$COMPOSE_PROJECT" ps -q >/dev/null 2>&1; then
		backend=dind
	else
		log "nothing to tear down"
		exit 0
	fi
fi

log "tearing down ${backend} registry"
case "$backend" in
qemu)
	pidf="$RUN_DIR/registry.pid"
	if [[ -f "$pidf" ]]; then
		pid="$(cat "$pidf" 2>/dev/null || true)"
		if [[ -n "${pid:-}" ]] && kill -0 "$pid" 2>/dev/null; then
			log "stopping qemu pid $pid"
			kill -TERM "$pid" 2>/dev/null || true
			sleep 1
			if kill -0 "$pid" 2>/dev/null; then
				kill -KILL "$pid" 2>/dev/null || true
			fi
		fi
		rm -f "$pidf"
	fi
	if [[ "$wipe" -eq 1 ]]; then
		log "removing registry VM disk and seeds in $RUN_DIR (cached cloud image kept)"
		rm -rf "$RUN_DIR"
	fi
	;;
dind)
	need_cmd docker
	args=(compose -p "$COMPOSE_PROJECT" -f "$ROOT/dind/compose.yaml" down --remove-orphans)
	if [[ "$wipe" -eq 1 ]]; then
		args+=(--volumes)
		log "compose down --volumes removes containerkit-registry project containers, network, and named volumes"
	else
		log "compose down removes containerkit-registry project containers and network; volumes kept"
	fi
	docker "${args[@]}"
	;;
*)
	die "unknown backend: $backend"
	;;
esac

rm -f "$RUN_DIR/backend"
log "registry down"
