#!/usr/bin/env bash
# Boot 4 QEMU VMs with Docker Engine listening on forwarded TCP ports.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck disable=SC1091
. "$ROOT/common.sh"

arch="$(host_arch)"
bin="$(qemu_bin "$arch")"
need_cmd "$bin"
need_cmd qemu-img
need_cmd curl
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

render_user_data() {
	local hostname="$1"
	cat <<EOF
#cloud-config
hostname: ${hostname}
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
runcmd:
  - curl -fsSL https://get.docker.com | sh
  - mkdir -p /etc/docker /etc/systemd/system/docker.service.d
  - printf '%s\\n' '{"hosts":["unix:///var/run/docker.sock","tcp://0.0.0.0:2375"]}' > /etc/docker/daemon.json
  - printf '%s\\n' '[Service]' 'ExecStart=' 'ExecStart=/usr/bin/dockerd' > /etc/systemd/system/docker.service.d/override.conf
  - systemctl daemon-reload
  - systemctl enable --now qemu-guest-agent
  - systemctl enable docker
  - systemctl restart docker
final_message: "containerkit swarm node ready"
EOF
}

render_network_config() {
	local wan="$1"
	local lan="$2"
	local ip="$3"
	cat <<EOF
network:
  version: 2
  ethernets:
    wan:
      match:
        macaddress: "${wan}"
      dhcp4: true
      dhcp4-overrides:
        route-metric: 100
    lan:
      match:
        macaddress: "${lan}"
      addresses:
        - ${ip}/24
      dhcp4: false
EOF
}

launch_node() {
	local idx="$1"
	local name="${NODE_NAMES[$idx]}"
	local ip="${NODE_IPS[$idx]}"
	local ssh_port="${NODE_SSH_PORTS[$idx]}"
	local docker_port="${NODE_DOCKER_PORTS[$idx]}"
	local mac_wan="${NODE_MAC_WAN[$idx]}"
	local mac_lan="${NODE_MAC_LAN[$idx]}"
	local disk="$RUN_DIR/${name}.qcow2"
	local seed_dir="$RUN_DIR/${name}-seed"
	local seed_iso="$RUN_DIR/${name}-cidata.iso"
	local vars="$RUN_DIR/${name}-vars.fd"
	local logf="$RUN_DIR/${name}.serial"
	local pidf="$RUN_DIR/${name}.pid"

	if [[ -f "$pidf" ]] && kill -0 "$(cat "$pidf")" 2>/dev/null; then
		log "$name already running (pid $(cat "$pidf"))"
		return 0
	fi

	if [[ ! -f "$disk" ]]; then
		qemu-img create -f qcow2 -F qcow2 -b "$base_img" "$disk" "${NODE_DISK_GB}G" >/dev/null
	fi
	if [[ -n "$fw_vars" && ! -f "$vars" ]]; then
		cp "$fw_vars" "$vars"
	fi

	rm -rf "$seed_dir"
	mkdir -p "$seed_dir"
	render_user_data "$name" >"$seed_dir/user-data"
	printf 'instance-id: containerkit-%s\nlocal-hostname: %s\n' "$name" "$name" >"$seed_dir/meta-data"
	render_network_config "$mac_wan" "$mac_lan" "$ip" >"$seed_dir/network-config"
	make_cidata_iso "$seed_dir" "$seed_iso"

	local netdev_lan=(
		-netdev "socket,id=lan0,mcast=${CLUSTER_MCAST}"
		-device "virtio-net-pci,netdev=lan0,mac=${mac_lan}"
	)
	if "$bin" -netdev help 2>&1 | grep -q vmnet-shared; then
		# Apple vmnet gives VMs a shared LAN; keep mcast as the stable 10.0.0.0/24 fabric.
		true
	fi

	local drive_vars=()
	if [[ -f "$vars" ]]; then
		drive_vars=(-drive "if=pflash,format=raw,file=${vars}")
	fi

	log "starting $name (ssh ${HOST_BIND}:${ssh_port}, docker ${HOST_BIND}:${docker_port}, lan ${ip})"
	"$bin" \
		-name "$name" \
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
		-netdev "user,id=wan0,hostfwd=tcp:${HOST_BIND}:${ssh_port}-:22,hostfwd=tcp:${HOST_BIND}:${docker_port}-:2375" \
		-device "virtio-net-pci,netdev=wan0,mac=${mac_wan}" \
		"${netdev_lan[@]}" \
		-pidfile "$pidf" \
		-daemonize
}

i=0
for _ in "${NODE_NAMES[@]}"; do
	launch_node "$i"
	i=$((i + 1))
done

write_cluster_env qemu

log "waiting for Docker Engine on all 4 nodes (cloud-init installs Docker; up to ${BOOT_TIMEOUT_SEC}s)"
failed=0
i=0
for name in "${NODE_NAMES[@]}"; do
	url="http://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}/_ping"
	if wait_http_ok "$url" "$BOOT_TIMEOUT_SEC"; then
		log "$name docker ready at tcp://${HOST_BIND}:${NODE_DOCKER_PORTS[$i]}"
	else
		log "$name did not become ready; last serial lines:"
		tail -n 40 "$RUN_DIR/${name}.serial" 2>/dev/null || true
		failed=1
	fi
	i=$((i + 1))
done
[[ "$failed" -eq 0 ]] || die "one or more QEMU nodes failed to start Docker"
log "qemu cluster is up; run: go test -tags e2e -count=1 ./e2e/swarm"
