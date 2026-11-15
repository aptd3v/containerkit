#!/usr/bin/env bash
# Force-remove every container, user-defined network, volume, and image,
# then run docker system prune -a --volumes.
#
# This deletes running containers, named volume data, images, and build cache
# on the current Docker context. Default networks (bridge, host, none) stay.
#
# Usage: scripts/docker-wipe.sh [--yes]
# Without --yes, the script prints the counts and waits for the word "wipe".
set -euo pipefail

if ! command -v docker >/dev/null 2>&1; then
	echo "docker is not on PATH" >&2
	exit 1
fi

confirm=0
if [[ "${1:-}" == "--yes" ]]; then
	confirm=1
elif [[ -n "${1:-}" ]]; then
	echo "usage: scripts/docker-wipe.sh [--yes]" >&2
	exit 2
fi

containers=$(docker ps -aq)
networks=$(docker network ls --format '{{.Name}}' | grep -v -E '^(bridge|host|none)$' || true)
volumes=$(docker volume ls -q)
images=$(docker images -aq | sort -u)

count() {
	if [[ -z "${1:-}" ]]; then
		echo 0
	else
		# shellcheck disable=SC2086
		echo $1 | wc -w | tr -d ' '
	fi
}

echo "This will permanently delete on the current Docker context:"
echo "  containers: $(count "$containers") (running and stopped)"
echo "  networks:   $(count "$networks") (user-defined; bridge, host, and none stay)"
echo "  volumes:    $(count "$volumes") (named and anonymous data)"
echo "  images:     $(count "$images")"
echo "  then:       docker system prune -af --volumes (stopped leftovers, unused networks, unused images, build cache, unused volumes)"

if [[ "$confirm" -ne 1 ]]; then
	if [[ ! -t 0 ]]; then
		echo "refusing to wipe without --yes when stdin is not a terminal" >&2
		exit 1
	fi
	printf 'Type wipe to continue: '
	read -r answer
	if [[ "$answer" != "wipe" ]]; then
		echo "aborted"
		exit 1
	fi
fi

if [[ -n "$containers" ]]; then
	# shellcheck disable=SC2086
	docker rm -f $containers
fi

if [[ -n "$networks" ]]; then
	# shellcheck disable=SC2086
	docker network rm -f $networks
fi

if [[ -n "$volumes" ]]; then
	# shellcheck disable=SC2086
	docker volume rm -f $volumes
fi

if [[ -n "$images" ]]; then
	# shellcheck disable=SC2086
	docker rmi -f $images
fi

docker system prune -af --volumes

echo "docker wipe finished"
