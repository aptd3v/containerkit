#!/usr/bin/env bash
# Report whether this host can provision the registry e2e environment.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

arch="$(host_arch)"
os="$(uname -s)"
backend="none"
ok=0

printf 'host\t%s/%s\n' "$os" "$arch"

if qemu_available; then
	bin="$(qemu_bin "$arch")"
	accel="$(qemu_accel)"
	printf 'qemu\t%s (%s accel)\n' "$(command -v "$bin")" "$accel"
	if [[ "$accel" == tcg ]]; then
		printf 'warn\tQEMU will use TCG; first boot will be slow\n'
	fi
	backend=qemu
	ok=1
else
	printf 'qemu\tmissing (install with: brew install qemu)\n'
fi

if command -v docker >/dev/null 2>&1; then
	if docker_available; then
		printf 'docker\t%s (daemon up)\n' "$(command -v docker)"
		if [[ "$backend" == none ]]; then
			backend=dind
		fi
		ok=1
	else
		printf 'docker\t%s (daemon not reachable)\n' "$(command -v docker)"
	fi
else
	printf 'docker\tmissing\n'
fi

if command -v htpasswd >/dev/null 2>&1; then
	printf 'htpasswd\tok\n'
else
	printf 'htpasswd\tmissing (needed to write bcrypt registry credentials)\n'
	ok=0
fi

if command -v hdiutil >/dev/null 2>&1 || command -v genisoimage >/dev/null 2>&1 || command -v mkisofs >/dev/null 2>&1 || command -v xorriso >/dev/null 2>&1; then
	printf 'cidata\tok\n'
else
	printf 'cidata\tmissing iso tool (hdiutil/genisoimage)\n'
	if [[ "$backend" == qemu ]]; then
		ok=0
	fi
fi

printf 'engine\t%s\n' "tcp://${HOST_BIND}:${DOCKER_PORT}"
printf 'registry\t%s\n' "http://${HOST_BIND}:${REGISTRY_PORT}/v2/"
printf 'backend\t%s\n' "$backend"

if [[ "$ok" -eq 1 ]]; then
	printf 'preflight\tok\n'
	exit 0
fi
printf 'preflight\tneed qemu (or a running Docker daemon) and htpasswd\n'
exit 1
