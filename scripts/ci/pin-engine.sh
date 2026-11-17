#!/usr/bin/env bash
# Install Docker Engine ${ENGINE_VERSION} on Ubuntu and print docker version.
# The printed server version is the Engine job's compatibility claim.
# A second Engine version is out of scope.
set -euo pipefail

ENGINE_VERSION="${ENGINE_VERSION:-29.9.3}"

if ! command -v docker >/dev/null 2>&1; then
	echo "docker is not on PATH" >&2
	exit 1
fi

server="$(docker version --format '{{.Server.Version}}')"
if [[ "$server" == "$ENGINE_VERSION" ]]; then
	docker version
	exit 0
fi

if [[ "$(uname -s)" != "Linux" ]]; then
	echo "Engine server ${server} does not match pinned ${ENGINE_VERSION}." >&2
	echo "scripts/ci/pin-engine.sh installs that server on Ubuntu. This machine is $(uname -s)." >&2
	exit 1
fi

if [[ ! -r /etc/os-release ]]; then
	echo "missing /etc/os-release" >&2
	exit 1
fi
# shellcheck disable=SC1091
. /etc/os-release
if [[ "${ID:-}" != "ubuntu" ]]; then
	echo "pin-engine.sh installs Docker Engine on Ubuntu, not ${ID:-unknown}" >&2
	exit 1
fi
codename="${UBUNTU_CODENAME:-${VERSION_CODENAME:-}}"
if [[ -z "$codename" ]]; then
	echo "Ubuntu codename is empty" >&2
	exit 1
fi

if ! apt-cache madison docker-ce | grep -q "5:${ENGINE_VERSION}-"; then
	sudo install -m 0755 -d /etc/apt/keyrings
	if [[ ! -f /etc/apt/keyrings/docker.asc ]]; then
		sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
		sudo chmod a+r /etc/apt/keyrings/docker.asc
	fi
	echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${codename} stable" |
		sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
	sudo apt-get update
fi

# apt-cache can exit 100. Keep the listing and match it in a second step.
madison="$(apt-cache madison docker-ce || true)"
pkg="$(printf '%s\n' "$madison" | awk -v v="5:${ENGINE_VERSION}-" '$3 ~ v { print $3; exit }')"
if [[ -z "$pkg" ]]; then
	echo "docker-ce ${ENGINE_VERSION} is not in the apt repo for ${codename}" >&2
	printf '%s\n' "$madison" | head >&2 || true
	exit 1
fi

sudo apt-get install -y --allow-downgrades "docker-ce=${pkg}" "docker-ce-cli=${pkg}"
sudo systemctl restart docker
for _ in $(seq 1 30); do
	if docker info >/dev/null 2>&1; then
		break
	fi
	sleep 1
done

server="$(docker version --format '{{.Server.Version}}')"
docker version
if [[ "$server" != "$ENGINE_VERSION" ]]; then
	echo "Engine server ${server} does not match pinned ${ENGINE_VERSION}" >&2
	exit 1
fi
