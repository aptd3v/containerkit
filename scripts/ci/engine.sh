#!/usr/bin/env bash
# Run dist/containerkit against one live Engine.
#
# The healthcheck project is scripts/ci/ck-health.yaml, not tsx/testdata/plan.json.
# pullImage is proved by deleting the test image first and inspecting it
# before composeUp. composeDown is proved with docker ps -a and docker network
# ls. The live section starts tsx/ci/docker.js, signals that Node process,
# and fails if project containers or the project network remain. Fallback
# cleanup of the live project runs only after those assertions.
#
# Usage: scripts/ci/engine.sh [all|live]
# ENGINE_VERSION, when set, must match docker's server version.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$root"

mode="${1:-all}"
case "$mode" in
all | live) ;;
*)
	echo "usage: scripts/ci/engine.sh [all|live]" >&2
	exit 2
	;;
esac

IMAGE="alpine:3.22"
HEALTH_PROJECT="ck-health"
LIVE_PROJECT="ck-live"
APP_PROJECT="app"
HEALTH_YAML="scripts/ci/ck-health.yaml"
PLAN_EDGE="tsx/testdata/plan.json"
PLAN_RUN="tsx/testdata/plan-run.json"

if [[ -n "${CONTAINERKIT_BIN:-}" ]]; then
	BIN="$CONTAINERKIT_BIN"
else
	BIN="$root/dist/containerkit"
fi
if [[ ! -x "$BIN" ]]; then
	echo "missing host binary $BIN (go run ./tools/crossbuild)" >&2
	exit 1
fi
BIN="$(cd "$(dirname "$BIN")" && pwd)/$(basename "$BIN")"
export CONTAINERKIT_BIN="$BIN"

if ! command -v docker >/dev/null 2>&1; then
	echo "docker is not on PATH" >&2
	exit 1
fi

echo "== Docker Engine =="
docker version
server="$(docker version --format '{{.Server.Version}}')"
if [[ -n "${ENGINE_VERSION:-}" && "$server" != "$ENGINE_VERSION" ]]; then
	echo "Engine server ${server} does not match pinned ${ENGINE_VERSION}" >&2
	exit 1
fi

ck() {
	local out
	out="$(mktemp)"
	if ! "$BIN" "$@" >"$out"; then
		echo "containerkit $* failed" >&2
		cat "$out" >&2
		rm -f "$out"
		exit 1
	fi
	cat "$out"
	rm -f "$out"
}

json_last() {
	python3 -c '
import json, sys
text = sys.stdin.read()
parsed = None
for line in reversed(text.splitlines()):
    line = line.strip()
    if not line:
        continue
    try:
        parsed = json.loads(line)
    except json.JSONDecodeError:
        continue
    else:
        break
if parsed is None:
    sys.stderr.write(repr(text[:500]) + "\n")
    sys.exit("containerkit wrote no JSON result")
json.dump(parsed, sys.stdout)
print()
'
}

project_containers() {
	docker ps -aq --filter "label=com.docker.compose.project=$1"
}

project_networks() {
	docker network ls -q --filter "label=com.docker.compose.project=$1"
}

show_project() {
	local project="$1"
	echo "containers for ${project}:"
	docker ps -a --filter "label=com.docker.compose.project=${project}"
	echo "networks for ${project}:"
	docker network ls --filter "label=com.docker.compose.project=${project}"
}

down_owned() {
	local project="$1"
	docker compose -p "$project" down --remove-orphans >/dev/null 2>&1 || true
}

require_absent_image() {
	docker image rm "$IMAGE" >/dev/null 2>&1 || true
	if docker image inspect "$IMAGE" >/dev/null 2>&1; then
		echo "test image ${IMAGE} is still present, so a no-op pullImage could pass" >&2
		docker ps -a --filter "ancestor=${IMAGE}" >&2 || true
		exit 1
	fi
}

write_pull_plan() {
	local dest="$1"
	python3 - "$dest" "$IMAGE" <<'PY'
import json, sys
dest, image = sys.argv[1:]
plan = {
    "connect": {"fromEnv": True},
    "actions": [{"op": "pullImage", "ref": image}],
}
with open(dest, "w") as fh:
    json.dump(plan, fh)
PY
}

write_health_plan() {
	local dest="$1" op="$2"
	python3 - "$dest" "$op" "$HEALTH_YAML" <<'PY'
import json, sys
dest, op, yaml_path = sys.argv[1:]
yaml = open(yaml_path).read()
if op == "up":
    actions = [{"op": "composeUp", "options": {"detach": True, "wait": True, "timeout": "30s"}}]
elif op == "ps":
    actions = [{"op": "composePs", "options": {"all": True}}]
elif op == "exec":
    actions = [{"op": "composeExec", "service": "check", "options": {"command": ["echo", "ck-exec-ok"]}}]
elif op == "down":
    actions = [{"op": "composeDown", "options": {"removeOrphans": True}}]
else:
    sys.exit(f"unknown health op {op}")
with open(dest, "w") as fh:
    json.dump({"connect": {"fromEnv": True}, "yaml": yaml, "actions": actions}, fh)
PY
}

plan_is_runnable() {
	python3 - "$1" <<'PY'
import json, sys
plan = json.load(open(sys.argv[1]))
actions = plan.get("actions") or []
yaml = plan.get("yaml") or ""
connect = plan.get("connect") or {}
ok = (
    len(actions) == 1
    and actions[0].get("op") == "composeUp"
    and (actions[0].get("options") or {}).get("detach") is True
    and "name: app\n" in yaml
    and "image: alpine:latest\n" in yaml
    and not connect.get("tls")
    and not connect.get("dryRun")
)
sys.exit(0 if ok else 1)
PY
}

cleanup_app_plan() {
	local yaml_file="$1" dest
	dest="$(mktemp)"
	python3 - "$dest" "$yaml_file" <<'PY'
import json, sys
dest, src = sys.argv[1:]
plan = json.load(open(src))
down = {
    "connect": {"fromEnv": True},
    "yaml": plan.get("yaml") or "",
    "actions": [{"op": "composeDown", "options": {"removeOrphans": True}}],
}
with open(dest, "w") as fh:
    json.dump(down, fh)
PY
	ck exec "$dest" >/dev/null
	rm -f "$dest"
}

health_proof() {
	local work ps_json exec_json containers networks logs
	echo "== healthcheck project ${HEALTH_PROJECT} =="
	grep -q "^name: ${HEALTH_PROJECT}$" "$HEALTH_YAML"
	grep -q "image: ${IMAGE}" "$HEALTH_YAML"
	python3 - "$HEALTH_YAML" <<'PY'
import sys
text = open(sys.argv[1]).read()
check = text.split("\n  check:\n", 1)[1]
if "while " in check or "until " in check:
    sys.exit("check polls; it must read readiness once and exit on failure")
PY
	down_owned "$HEALTH_PROJECT"
	require_absent_image
	work="$(mktemp -d)"
	write_pull_plan "$work/pull.json"
	echo "pullImage ${IMAGE}"
	ck exec "$work/pull.json" >/dev/null
	if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
		echo "pullImage returned without creating ${IMAGE}" >&2
		exit 1
	fi
	write_health_plan "$work/up.json" up
	write_health_plan "$work/ps.json" ps
	write_health_plan "$work/exec.json" exec
	write_health_plan "$work/down.json" down
	echo "composeUp ${HEALTH_PROJECT}"
	ck exec "$work/up.json" >/dev/null
	ps_json="$(ck exec "$work/ps.json" | json_last)"
	python3 -c '
import json, sys
rows = json.loads(sys.argv[1])
services = {row.get("Service") for row in rows}
missing = {"ready", "check"} - services
if missing:
    sys.exit(f"compose ps missing {sorted(missing)}: {sys.argv[1]}")
' "$ps_json"
	exec_json="$(ck exec "$work/exec.json" | json_last)"
	python3 -c '
import json, sys
if json.loads(sys.argv[1]) != 0:
    sys.exit(f"exec returned {sys.argv[1]}, want 0")
' "$exec_json"
	logs=""
	while IFS= read -r id; do
		logs="$id"
		break
	done < <(docker ps -aq --filter "label=com.docker.compose.project=${HEALTH_PROJECT}" --filter "label=com.docker.compose.service=check")
	if [[ -z "$logs" ]] || ! docker logs "$logs" 2>&1 | grep -q ck-check-marker; then
		echo "logs for ${HEALTH_PROJECT}/check do not contain ck-check-marker" >&2
		if [[ -n "$logs" ]]; then
			docker logs "$logs" >&2 || true
		fi
		exit 1
	fi
	containers="$(project_containers "$HEALTH_PROJECT")"
	networks="$(project_networks "$HEALTH_PROJECT")"
	if [[ -z "$containers" || -z "$networks" ]]; then
		echo "project container or network missing before compose down" >&2
		show_project "$HEALTH_PROJECT" >&2
		exit 1
	fi
	echo "composeDown ${HEALTH_PROJECT}"
	ck exec "$work/down.json" >/dev/null
	containers="$(project_containers "$HEALTH_PROJECT")"
	networks="$(project_networks "$HEALTH_PROJECT")"
	if [[ -n "$containers" || -n "$networks" ]]; then
		echo "project resources survived compose down:" >&2
		show_project "$HEALTH_PROJECT" >&2
		exit 1
	fi
	rm -rf "$work"
}

plan_proof() {
	local target
	echo "== plan fixture =="
	ck validate "$PLAN_EDGE"
	ck validate "$PLAN_RUN"
	if plan_is_runnable "$PLAN_EDGE"; then
		target="$PLAN_EDGE"
		echo "exec ${PLAN_EDGE}"
	else
		target="$PLAN_RUN"
		echo "validate-only ${PLAN_EDGE}; exec ${PLAN_RUN}"
	fi
	if [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
		down_owned "$APP_PROJECT"
	elif [[ -n "$(project_containers "$APP_PROJECT")" ]]; then
		echo "project ${APP_PROJECT} already has containers. Set CK_ENGINE_RESET_APP=1 to remove them first." >&2
		show_project "$APP_PROJECT" >&2
		exit 1
	fi
	if [[ "${CK_ENGINE_RESET_APP:-}" == "1" ]]; then
		down_owned "$APP_PROJECT"
	fi
	ck exec "$target" >/dev/null
	if [[ -z "$(docker ps -aq --filter "label=com.docker.compose.project=${APP_PROJECT}" --filter "label=com.docker.compose.service=api")" ]]; then
		echo "app/api was not created, including stopped containers" >&2
		show_project "$APP_PROJECT" >&2
		exit 1
	fi
	echo "cleanup project ${APP_PROJECT}"
	cleanup_app_plan "$target"
	if [[ -n "$(project_containers "$APP_PROJECT")" || -n "$(project_networks "$APP_PROJECT")" ]]; then
		echo "project ${APP_PROJECT} survived cleanup:" >&2
		show_project "$APP_PROJECT" >&2
		exit 1
	fi
}

live_proof() {
	local log err pid i containers networks code
	echo "== live teardown ${LIVE_PROJECT} =="
	if [[ ! -f tsx/dist/live/index.js ]]; then
		echo "missing tsx/dist (cd tsx && npm ci && npm run build)" >&2
		exit 1
	fi
	down_owned "$LIVE_PROJECT"
	log="$(mktemp)"
	err="$(mktemp)"
	node tsx/ci/docker.js >"$log" 2>"$err" &
	pid="$!"
	for i in $(seq 1 90); do
		if grep -q ck-live-up "$log"; then
			break
		fi
		if ! kill -0 "$pid" 2>/dev/null; then
			echo "node exited before onUp" >&2
			cat "$log" >&2 || true
			cat "$err" >&2 || true
			exit 1
		fi
		sleep 1
	done
	if ! grep -q ck-live-up "$log"; then
		echo "timed out waiting for ck-live-up" >&2
		cat "$log" >&2 || true
		cat "$err" >&2 || true
		kill -TERM "$pid" 2>/dev/null || true
		exit 1
	fi
	if [[ -z "$(project_containers "$LIVE_PROJECT")" ]]; then
		echo "project container missing after onUp" >&2
		show_project "$LIVE_PROJECT" >&2
		kill -TERM "$pid" 2>/dev/null || true
		exit 1
	fi
	kill -TERM "$pid"
	for i in $(seq 1 60); do
		if ! kill -0 "$pid" 2>/dev/null; then
			break
		fi
		sleep 1
	done
	if kill -0 "$pid" 2>/dev/null; then
		echo "node did not exit within 60s after SIGTERM" >&2
		kill -KILL "$pid" 2>/dev/null || true
		exit 1
	fi
	set +e
	wait "$pid"
	code=$?
	set -e
	containers="$(project_containers "$LIVE_PROJECT")"
	networks="$(project_networks "$LIVE_PROJECT")"
	if [[ -n "$containers" || -n "$networks" ]]; then
		echo "surviving project resources after SIGTERM (node exit ${code}):" >&2
		show_project "$LIVE_PROJECT" >&2
		exit 1
	fi
	if [[ "$code" -ne 0 ]]; then
		echo "node exited ${code}" >&2
		cat "$err" >&2 || true
		exit 1
	fi
	# Fallback cleanup runs only after the assertions above.
	down_owned "$LIVE_PROJECT"
	rm -f "$log" "$err"
}

case "$mode" in
all)
	health_proof
	plan_proof
	live_proof
	;;
live)
	live_proof
	;;
esac

echo "engine proof ok (${server})"
