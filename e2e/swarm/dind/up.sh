#!/usr/bin/env bash
# Bring up 4 privileged Docker-in-Docker nodes as a QEMU-similar swarm fabric.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

need_cmd docker
docker info >/dev/null || die "docker daemon is not reachable"
ensure_dirs

log "starting 4 dind nodes via compose project ${COMPOSE_PROJECT}"
docker compose -p "$COMPOSE_PROJECT" -f "$ROOT/dind/compose.yaml" up -d --wait --wait-timeout "$BOOT_TIMEOUT_SEC"

write_cluster_env dind

log "waiting for Docker API on forwarded ports"
failed=0
i=0
for name in "${NODE_NAMES[@]}"; do
	url="http://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}/_ping"
	if wait_http_ok "$url" 60; then
		log "$name docker ready at tcp://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}"
	else
		log "$name API not reachable at $url"
		failed=1
	fi
	i=$((i + 1))
done
[[ "$failed" -eq 0 ]] || die "one or more dind nodes failed to expose the Docker API"
log "dind cluster is up; run: go test -tags e2e -count=1 ./e2e/swarm"
