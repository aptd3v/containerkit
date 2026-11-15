#!/usr/bin/env bash
# Print backend, forwarded Docker API ports, and ping each node.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

backend="$(running_backend)"
printf 'backend\t%s\n' "${backend:-unknown}"
if [[ -f "$RUN_DIR/env" ]]; then
	# shellcheck disable=SC1091
	. "$RUN_DIR/env"
	printf 'manager\t%s\n' "${CONTAINERKIT_SWARM_MANAGER:-}"
fi

i=0
for name in "${NODE_NAMES[@]}"; do
	url="http://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}/_ping"
	if wait_http_ok "$url" 2; then
		printf 'node\t%s\t%s\tok\n' "$name" "tcp://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}"
	else
		printf 'node\t%s\t%s\tdown\n' "$name" "tcp://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}"
	fi
	i=$((i + 1))
done
