#!/usr/bin/env bash
# Shared cluster constants and helpers for the 4-node swarm e2e environment.
# shellcheck disable=SC2034

set -euo pipefail

SWARM_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CACHE_DIR="${CONTAINERKIT_SWARM_CACHE:-$SWARM_ROOT/.cache}"
RUN_DIR="${CONTAINERKIT_SWARM_RUN:-$SWARM_ROOT/.run}"
COMPOSE_PROJECT="${CONTAINERKIT_SWARM_PROJECT:-containerkit-swarm}"
CLUSTER_SUBNET="${CONTAINERKIT_SWARM_SUBNET:-10.0.0.0/24}"
CLUSTER_MCAST="${CONTAINERKIT_SWARM_MCAST:-230.0.0.1:12340}"
NODE_RAM_MB="${CONTAINERKIT_SWARM_RAM_MB:-768}"
NODE_CPUS="${CONTAINERKIT_SWARM_CPUS:-1}"
NODE_DISK_GB="${CONTAINERKIT_SWARM_DISK_GB:-8}"
BOOT_TIMEOUT_SEC="${CONTAINERKIT_SWARM_BOOT_TIMEOUT:-600}"

NODE_NAMES=(mgr w1 w2 w3)
NODE_IPS=(10.0.0.11 10.0.0.12 10.0.0.13 10.0.0.14)
NODE_SSH_PORTS=(2201 2202 2203 2204)
NODE_DOCKER_PORTS=(23751 23752 23753 23754)
NODE_MAC_WAN=(52:54:00:12:00:01 52:54:00:12:00:02 52:54:00:12:00:03 52:54:00:12:00:04)
NODE_MAC_LAN=(52:54:00:12:01:01 52:54:00:12:01:02 52:54:00:12:01:03 52:54:00:12:01:04)

HOST_BIND="${CONTAINERKIT_SWARM_BIND:-127.0.0.1}"

log() { printf '%s\n' "$*" >&2; }
die() { log "error: $*"; exit 1; }

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

host_arch() {
	local m
	m="$(uname -m)"
	case "$m" in
	arm64 | aarch64) printf 'aarch64\n' ;;
	x86_64 | amd64) printf 'x86_64\n' ;;
	*) die "unsupported architecture: $m" ;;
	esac
}

qemu_bin() {
	local arch="$1"
	case "$arch" in
	aarch64) printf 'qemu-system-aarch64\n' ;;
	x86_64) printf 'qemu-system-x86_64\n' ;;
	esac
}

qemu_available() {
	local arch bin
	arch="$(host_arch)"
	bin="$(qemu_bin "$arch")"
	command -v "$bin" >/dev/null 2>&1 && command -v qemu-img >/dev/null 2>&1
}

docker_available() {
	command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1
}

detect_backend() {
	if [[ -n "${CONTAINERKIT_SWARM_BACKEND:-}" ]]; then
		printf '%s\n' "$CONTAINERKIT_SWARM_BACKEND"
		return 0
	fi
	if qemu_available; then
		printf 'qemu\n'
		return 0
	fi
	if docker_available; then
		printf 'dind\n'
		return 0
	fi
	die "no swarm backend: install qemu (brew install qemu) or start Docker Desktop"
}

qemu_accel() {
	local os bin
	os="$(uname -s)"
	bin="$(qemu_bin "$(host_arch)")"
	if [[ "$os" == Darwin ]] && "$bin" -accel help 2>/dev/null | grep -q hvf; then
		printf 'hvf\n'
		return 0
	fi
	if [[ "$os" == Linux ]] && [[ -w /dev/kvm ]] && "$bin" -accel help 2>/dev/null | grep -q kvm; then
		printf 'kvm\n'
		return 0
	fi
	printf 'tcg\n'
}

cloud_image_url() {
	local arch="$1"
	if [[ -n "${CONTAINERKIT_SWARM_IMAGE_URL:-}" ]]; then
		printf '%s\n' "$CONTAINERKIT_SWARM_IMAGE_URL"
		return 0
	fi
	case "$arch" in
	aarch64)
		printf 'https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-arm64.qcow2\n'
		;;
	x86_64)
		printf 'https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2\n'
		;;
	esac
}

ensure_dirs() {
	mkdir -p "$CACHE_DIR" "$RUN_DIR"
}

write_cluster_env() {
	local backend="$1"
	local envf="$RUN_DIR/env"
	local nodes="" ips="" i
	for i in "${!NODE_NAMES[@]}"; do
		if [[ -n "$nodes" ]]; then
			nodes+=","
			ips+=","
		fi
		nodes+="tcp://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}"
		ips+="${NODE_IPS[$i]}"
	done
	cat >"$envf" <<EOF
CONTAINERKIT_SWARM_BACKEND=${backend}
CONTAINERKIT_SWARM_MANAGER=tcp://${HOST_BIND}:${NODE_DOCKER_PORTS[0]}
CONTAINERKIT_SWARM_NODES=${nodes}
CONTAINERKIT_SWARM_ADVERTISE=${NODE_IPS[0]}
CONTAINERKIT_SWARM_NODE_IPS=${ips}
CONTAINERKIT_SWARM_LISTEN=0.0.0.0:2377
EOF
	log "wrote $envf"
}

cluster_reachable() {
	local i=0
	for _ in "${NODE_NAMES[@]}"; do
		if ! wait_http_ok "http://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}/_ping" 2; then
			return 1
		fi
		i=$((i + 1))
	done
	return 0
}

wait_http_ok() {
	local url="$1"
	local timeout="${2:-$BOOT_TIMEOUT_SEC}"
	local elapsed=0
	while ((elapsed < timeout)); do
		if command -v curl >/dev/null 2>&1; then
			if curl -fsS --max-time 1 "$url" >/dev/null 2>&1; then
				return 0
			fi
		elif command -v nc >/dev/null 2>&1; then
			local host port
			host="${url#http://}"
			host="${host%%/*}"
			port="${host##*:}"
			host="${host%%:*}"
			if nc -z -G 1 "$host" "$port" >/dev/null 2>&1; then
				return 0
			fi
		fi
		sleep 2
		elapsed=$((elapsed + 2))
	done
	return 1
}

make_cidata_iso() {
	local src="$1"
	local iso="$2"
	rm -f "$iso" "${iso}.cdr"
	if command -v genisoimage >/dev/null 2>&1; then
		genisoimage -output "$iso" -volid cidata -joliet -rock -quiet "$src"
	elif command -v mkisofs >/dev/null 2>&1; then
		mkisofs -output "$iso" -volid cidata -joliet -rock "$src" >/dev/null
	elif command -v xorriso >/dev/null 2>&1; then
		xorriso -as mkisofs -V cidata -o "$iso" "$src" >/dev/null 2>&1
	elif command -v hdiutil >/dev/null 2>&1; then
		hdiutil makehybrid -iso -joliet -default-volume-name CIDATA -o "$iso" "$src" >/dev/null
		if [[ ! -f "$iso" && -f "${iso}.cdr" ]]; then
			mv "${iso}.cdr" "$iso"
		fi
		if [[ ! -f "$iso" && -f "${iso}.iso" ]]; then
			mv "${iso}.iso" "$iso"
		fi
	else
		die "need genisoimage, mkisofs, xorriso, or hdiutil to build cloud-init seed"
	fi
	[[ -f "$iso" ]] || die "failed to create cidata iso at $iso"
}

find_firmware() {
	local arch="$1"
	local code="" vars=""
	local candidates=()
	case "$arch" in
	aarch64)
		candidates=(
			/opt/homebrew/share/qemu/edk2-aarch64-code.fd
			/usr/local/share/qemu/edk2-aarch64-code.fd
			/usr/share/AAVMF/AAVMF_CODE.fd
			/usr/share/qemu/edk2-aarch64-code.fd
		)
		;;
	x86_64)
		candidates=(
			/opt/homebrew/share/qemu/edk2-x86_64-code.fd
			/usr/local/share/qemu/edk2-x86_64-code.fd
			/usr/share/OVMF/OVMF_CODE.fd
			/usr/share/qemu/edk2-x86_64-code.fd
		)
		;;
	esac
	local f
	for f in "${candidates[@]}"; do
		if [[ -f "$f" ]]; then
			code="$f"
			break
		fi
	done
	[[ -n "$code" ]] || die "UEFI firmware not found for $arch (install qemu)"
	case "$arch" in
	aarch64)
		for f in \
			/opt/homebrew/share/qemu/edk2-arm-vars.fd \
			/usr/local/share/qemu/edk2-arm-vars.fd \
			/usr/share/AAVMF/AAVMF_VARS.fd \
			/usr/share/qemu/edk2-arm-vars.fd; do
			if [[ -f "$f" ]]; then
				vars="$f"
				break
			fi
		done
		;;
	x86_64)
		for f in \
			/opt/homebrew/share/qemu/edk2-i386-vars.fd \
			/usr/local/share/qemu/edk2-i386-vars.fd \
			/usr/share/OVMF/OVMF_VARS.fd \
			/usr/share/qemu/edk2-i386-vars.fd; do
			if [[ -f "$f" ]]; then
				vars="$f"
				break
			fi
		done
		;;
	esac
	printf '%s\t%s\n' "$code" "$vars"
}

ensure_ssh_key() {
	local key="$RUN_DIR/id_ed25519"
	if [[ ! -f "$key" ]]; then
		need_cmd ssh-keygen
		ssh-keygen -t ed25519 -N '' -f "$key" >/dev/null
	fi
	cat "${key}.pub"
}

running_backend() {
	if [[ -f "$RUN_DIR/backend" ]]; then
		cat "$RUN_DIR/backend"
		return 0
	fi
	if [[ -f "$RUN_DIR/env" ]]; then
		# shellcheck disable=SC1091
		. "$RUN_DIR/env"
		printf '%s\n' "${CONTAINERKIT_SWARM_BACKEND:-}"
		return 0
	fi
	printf '\n'
}
