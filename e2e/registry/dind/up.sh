#!/usr/bin/env bash
# Bring up DinD Engine plus an htpasswd-authenticated registry (QEMU-similar).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

need_cmd docker
need_cmd htpasswd
docker info >/dev/null || die "docker daemon is not reachable"
ensure_dirs
write_htpasswd "$RUN_DIR/htpasswd"

log "starting registry + dind via compose project ${COMPOSE_PROJECT}"
docker compose -p "$COMPOSE_PROJECT" -f "$ROOT/dind/compose.yaml" up -d --wait --wait-timeout "$BOOT_TIMEOUT_SEC"

write_registry_env dind "tcp://${HOST_BIND}:${DOCKER_PORT}" "registry:5000"

log "waiting for Docker API on tcp://${HOST_BIND}:${DOCKER_PORT}"
if wait_http_ok "http://${HOST_BIND}:${DOCKER_PORT}/_ping" 60; then
	log "docker ready at tcp://${HOST_BIND}:${DOCKER_PORT}"
else
	die "dind engine API not reachable"
fi

log "waiting for registry auth challenge"
if wait_registry_auth "http://${HOST_BIND}:${REGISTRY_PORT}/v2/" 60; then
	log "registry requires authentication"
else
	die "authenticated registry did not become ready"
fi

log "dind registry is up; run: go test -tags e2e -count=1 ./e2e/registry"
