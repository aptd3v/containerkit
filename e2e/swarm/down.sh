#!/usr/bin/env bash
# Tear down the 4-node swarm cluster started by up.sh.
# Stops only containerkit-swarm QEMU PIDs or the containerkit-swarm Compose project.
# Does not prune host Docker images, volumes, or unrelated containers.
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
	if qemu_available && compgen -G "$RUN_DIR"/*.pid >/dev/null 2>&1; then
		backend=qemu
	elif docker_available && docker compose -p "$COMPOSE_PROJECT" ps -q >/dev/null 2>&1; then
		backend=dind
	else
		log "nothing to tear down"
		exit 0
	fi
fi

log "tearing down ${backend} cluster"
case "$backend" in
qemu)
	shopt -s nullglob
	for pidf in "$RUN_DIR"/*.pid; do
		pid="$(cat "$pidf" 2>/dev/null || true)"
		if [[ -n "${pid:-}" ]] && kill -0 "$pid" 2>/dev/null; then
			log "stopping qemu pid $pid"
			kill -TERM "$pid" 2>/dev/null || true
		fi
	done
	sleep 1
	for pidf in "$RUN_DIR"/*.pid; do
		pid="$(cat "$pidf" 2>/dev/null || true)"
		if [[ -n "${pid:-}" ]] && kill -0 "$pid" 2>/dev/null; then
			kill -KILL "$pid" 2>/dev/null || true
		fi
		rm -f "$pidf"
	done
	if [[ "$wipe" -eq 1 ]]; then
		log "removing node disks and seeds in $RUN_DIR (cached cloud image kept)"
		rm -rf "$RUN_DIR"
	fi
	;;
dind)
	need_cmd docker
	args=(compose -p "$COMPOSE_PROJECT" -f "$ROOT/dind/compose.yaml" down --remove-orphans)
	if [[ "$wipe" -eq 1 ]]; then
		args+=(--volumes)
		log "compose down --volumes removes containerkit-swarm project containers, network, and named volumes"
	else
		log "compose down removes containerkit-swarm project containers and network; volumes kept"
	fi
	docker "${args[@]}"
	;;
*)
	die "unknown backend: $backend"
	;;
esac

rm -f "$RUN_DIR/backend"
log "cluster down"
