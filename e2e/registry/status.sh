#!/usr/bin/env bash
# Print backend, forwarded Engine API, and registry auth challenge.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

backend="$(running_backend)"
printf 'backend\t%s\n' "${backend:-unknown}"
if [[ -f "$RUN_DIR/env" ]]; then
	# shellcheck disable=SC1091
	. "$RUN_DIR/env"
	printf 'engine\t%s\n' "${CONTAINERKIT_REGISTRY_ENGINE:-}"
	printf 'registry\t%s\n' "${CONTAINERKIT_REGISTRY_HOST:-}"
fi

if wait_http_ok "http://${HOST_BIND}:${DOCKER_PORT}/_ping" 2; then
	printf 'engine-api\t%s\tok\n' "tcp://${HOST_BIND}:${DOCKER_PORT}"
else
	printf 'engine-api\t%s\tdown\n' "tcp://${HOST_BIND}:${DOCKER_PORT}"
fi

url="http://${HOST_BIND}:${REGISTRY_PORT}/v2/"
if wait_registry_auth "$url" 2; then
	printf 'registry\t%s\tauth-required\n' "$url"
else
	printf 'registry\t%s\tdown\n' "$url"
fi
