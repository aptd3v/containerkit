#!/usr/bin/env bash
# Boot one QEMU VM with Docker Engine and an htpasswd-authenticated registry.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

arch="$(host_arch)"
bin="$(qemu_bin "$arch")"
need_cmd "$bin"
need_cmd qemu-img
need_cmd curl
need_cmd htpasswd
ensure_dirs
ensure_ssh_key >/dev/null
pubkey="$(ensure_ssh_key)"
accel="$(qemu_accel)"
if [[ "$accel" == tcg ]]; then
	log "warning: using TCG emulation; first boot will take several minutes"
fi

img_url="$(cloud_image_url "$arch")"
base_img="$CACHE_DIR/$(basename "$img_url")"
if [[ ! -f "$base_img" ]]; then
	log "downloading cloud image $img_url"
	curl -L --fail --retry 3 --continue-at - -o "${base_img}.part" "$img_url"
	mv "${base_img}.part" "$base_img"
fi

fw="$(find_firmware "$arch")"
fw_code="${fw%%$'\t'*}"
fw_vars="${fw#*$'\t'}"
[[ -n "$fw_code" ]] || die "no UEFI firmware"
if [[ "$arch" == aarch64 && -z "$fw_vars" ]]; then
	die "aarch64 UEFI vars template not found (install qemu)"
fi
base_img="$(cd "$(dirname "$base_img")" && pwd)/$(basename "$base_img")"

machine_args=()
cpu_args=()
case "$arch" in
aarch64)
	machine_args=(-machine virt)
	if [[ "$accel" == hvf || "$accel" == kvm ]]; then
		cpu_args=(-cpu host)
	else
		cpu_args=(-cpu cortex-a72)
	fi
	;;
x86_64)
	machine_args=(-machine q35)
	if [[ "$accel" == hvf || "$accel" == kvm ]]; then
		cpu_args=(-cpu host)
	else
		cpu_args=(-cpu qemu64)
	fi
	;;
esac

write_htpasswd "$RUN_DIR/htpasswd"
htpasswd_line="$(tr -d '\r\n' <"$RUN_DIR/htpasswd")"

pidf="$RUN_DIR/registry.pid"
disk="$RUN_DIR/registry.qcow2"
seed_dir="$RUN_DIR/registry-seed"
seed_iso="$RUN_DIR/registry-cidata.iso"
vars="$RUN_DIR/registry-vars.fd"
logf="$RUN_DIR/registry.serial"

if [[ -f "$pidf" ]] && kill -0 "$(cat "$pidf")" 2>/dev/null; then
	log "registry VM already running (pid $(cat "$pidf"))"
else
	if [[ ! -f "$disk" ]]; then
		qemu-img create -f qcow2 -F qcow2 -b "$base_img" "$disk" "${NODE_DISK_GB}G" >/dev/null
	fi
	if [[ -n "$fw_vars" && ! -f "$vars" ]]; then
		cp "$fw_vars" "$vars"
	fi

	rm -rf "$seed_dir"
	mkdir -p "$seed_dir"
	cat >"$seed_dir/user-data" <<EOF
#cloud-config
hostname: registry
manage_etc_hosts: true
users:
  - name: debian
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: [sudo]
    shell: /bin/bash
    lock_passwd: false
    plain_text_passwd: debian
    ssh_authorized_keys:
      - ${pubkey}
ssh_pwauth: true
package_update: true
packages:
  - curl
  - ca-certificates
  - qemu-guest-agent
write_files:
  - path: /opt/registry/auth/htpasswd
    permissions: '0644'
    content: |
      ${htpasswd_line}
runcmd:
  - curl -fsSL https://get.docker.com | sh
  - mkdir -p /etc/docker /etc/systemd/system/docker.service.d
  - printf '%s\n' '{"hosts":["unix:///var/run/docker.sock","tcp://0.0.0.0:2375"],"insecure-registries":["127.0.0.1:5000","localhost:5000"]}' > /etc/docker/daemon.json
  - printf '%s\n' '[Service]' 'ExecStart=' 'ExecStart=/usr/bin/dockerd' > /etc/systemd/system/docker.service.d/override.conf
  - systemctl daemon-reload
  - systemctl enable --now qemu-guest-agent
  - systemctl enable docker
  - systemctl restart docker
  - /bin/bash -c 'for i in \$(seq 1 90); do docker info >/dev/null 2>&1 && break; sleep 2; done; docker rm -f containerkit-registry >/dev/null 2>&1 || true; docker run -d --name containerkit-registry --restart always -p 5000:5000 -v /opt/registry/auth:/auth -e REGISTRY_AUTH=htpasswd -e REGISTRY_AUTH_HTPASSWD_REALM=Registry -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd -e REGISTRY_STORAGE_DELETE_ENABLED=true ${REGISTRY_IMAGE}'
final_message: "containerkit registry node ready"
EOF
	printf 'instance-id: containerkit-registry\nlocal-hostname: registry\n' >"$seed_dir/meta-data"
	make_cidata_iso "$seed_dir" "$seed_iso"

	drive_vars=()
	if [[ -f "$vars" ]]; then
		drive_vars=(-drive "if=pflash,format=raw,file=${vars}")
	fi

	log "starting registry VM (ssh ${HOST_BIND}:${SSH_PORT}, docker ${HOST_BIND}:${DOCKER_PORT}, registry ${HOST_BIND}:${REGISTRY_PORT})"
	"$bin" \
		-name containerkit-registry \
		"${machine_args[@]}" \
		-accel "$accel" \
		"${cpu_args[@]}" \
		-smp "$NODE_CPUS" \
		-m "$NODE_RAM_MB" \
		-display none \
		-serial "file:${logf}" \
		-drive "if=pflash,format=raw,readonly=on,file=${fw_code}" \
		"${drive_vars[@]}" \
		-drive "if=virtio,format=qcow2,file=${disk}" \
		-drive "if=virtio,format=raw,file=${seed_iso}" \
		-netdev "user,id=wan0,hostfwd=tcp:${HOST_BIND}:${SSH_PORT}-:22,hostfwd=tcp:${HOST_BIND}:${DOCKER_PORT}-:2375,hostfwd=tcp:${HOST_BIND}:${REGISTRY_PORT}-:5000" \
		-device "virtio-net-pci,netdev=wan0,mac=${VM_MAC}" \
		-pidfile "$pidf" \
		-daemonize
fi

write_registry_env qemu "tcp://${HOST_BIND}:${DOCKER_PORT}" "127.0.0.1:5000"

log "waiting for Docker Engine (cloud-init installs Docker; up to ${BOOT_TIMEOUT_SEC}s)"
if wait_http_ok "http://${HOST_BIND}:${DOCKER_PORT}/_ping" "$BOOT_TIMEOUT_SEC"; then
	log "docker ready at tcp://${HOST_BIND}:${DOCKER_PORT}"
else
	log "engine did not become ready; last serial lines:"
	tail -n 80 "$logf" 2>/dev/null || true
	die "QEMU registry VM failed to start Docker"
fi

log "waiting for registry auth challenge on http://${HOST_BIND}:${REGISTRY_PORT}/v2/"
if wait_registry_auth "http://${HOST_BIND}:${REGISTRY_PORT}/v2/" "$BOOT_TIMEOUT_SEC"; then
	log "registry requires authentication"
else
	log "registry did not challenge; last serial lines:"
	tail -n 80 "$logf" 2>/dev/null || true
	die "authenticated registry did not become ready"
fi

log "qemu registry is up; run: go test -tags e2e -count=1 ./e2e/registry"
