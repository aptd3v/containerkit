import assert from "node:assert/strict";
import { test } from "node:test";
import { Compose, Env, Health, Service } from "./generated.js";
import { composeUp, listImages, pullImage } from "./generated-actions.js";
import { connect, encode, run, setInvoke, type Client } from "./plan.js";
import type {
	ConnectOptions,
	EngineImageInspectOptions,
	KillOptions,
	PublishOptions,
	PullImageOptions,
	UpOptions,
	UpdateOptions,
	VolumeOptions,
} from "./generated-options.js";
import type { ServiceProps } from "./generated.js";
import { oci, order, selinux, signal } from "./vocab.js";

function _scopeTypes(cli: Client) {
	const connectOpts: ConnectOptions = { fromEnv: true, host: "unix:///var/run/docker.sock", timeout: "10s" };
	const upOpts: UpOptions = { detach: true, timeout: "30s" };
	void connect(connectOpts);
	void composeUp(cli, Compose({ name: "app" }), upOpts);
	const _health = Health({ timeout: "5s" });
	void _health;
	const _badConnect: ConnectOptions = {
		fromEnv: true,
		// @ts-expect-error detach is an Up setter, not a client setter
		detach: true,
	};
	void _badConnect;
	const _badUp: UpOptions = {
		detach: true,
		// @ts-expect-error host is a client setter, not an Up setter
		host: "unix:///var/run/docker.sock",
	};
	void _badUp;
}
void _scopeTypes;

type Absent<T, K extends string> = K extends keyof T ? false : true;

function _vocabOptions() {
	const kill: KillOptions = { stopSignal: signal.SIGTERM };
	const published: PublishOptions = { oci: oci.V1_1 };
	const updated: UpdateOptions = { order: order.StartFirst };
	const volume: VolumeOptions = { seLinux: selinux.Shared };
	const inspect: EngineImageInspectOptions = {
		imagePlatform: { architecture: "amd64", os: "linux", variant: "v8" },
	};
	const badKill: KillOptions = {
		// @ts-expect-error stopSignal is a signal name
		stopSignal: "nope",
	};
	const fromIsAbsent: Absent<ServiceProps, "fromContainer"> = true;
	const logIsAbsent: Absent<PullImageOptions, "log"> = true;
	const consumeIsAbsent: Absent<UpOptions, "consumeLogs"> = true;
	const outputIsAbsent: Absent<UpOptions, "outputStream"> = true;
	void kill;
	void published;
	void updated;
	void volume;
	void inspect;
	void badKill;
	void fromIsAbsent;
	void logIsAbsent;
	void consumeIsAbsent;
	void outputIsAbsent;
}
void _vocabOptions;

async function _factoryTakesCli({ cli }: { cli: Client }): Promise<import("./docker-node.js").DockerNode> {
	await run(listImages(cli, { all: true }));
	return Compose({ name: "from-images" });
}
void _factoryTakesCli;

test("connect holds client setters and is not a DockerNode", () => {
	const cli = connect({ fromEnv: true, timeout: "10s", host: "unix:///var/run/docker.sock" });
	assert.equal(cli.$$kind, "client");
	assert.equal(cli.options.fromEnv, true);
	assert.equal(cli.options.timeout, "10s");
});

test("composeUp encodes yaml plus UpOptions without putting them on Compose", () => {
	const cli = connect({ fromEnv: true });
	const doc = composeUp(
		cli,
		Compose({
			name: "app",
			children: Service({
				name: "api",
				image: "alpine:latest",
				children: [Env({ name: "APP", value: "1" }), Health({ timeout: "3s", test: ["CMD", "true"] })],
			}),
		}),
		{ detach: true, timeout: "30s" },
	);
	assert.equal(doc.connect?.fromEnv, true);
	assert.match(doc.yaml ?? "", /name: app/);
	assert.match(doc.yaml ?? "", /timeout: "?3s"?/);
	assert.equal(doc.actions?.[0]?.op, "composeUp");
	const opts = doc.actions?.[0]?.options as UpOptions;
	assert.equal(opts.detach, true);
	assert.equal(opts.timeout, "30s");
	assert.equal(JSON.parse(encode(doc)).actions[0].op, "composeUp");
});

test("pullImage and listImages encode engine verbs", () => {
	const cli = connect({ fromEnv: true });
	const pull = pullImage(cli, "alpine:latest", { all: true });
	assert.equal(pull.actions?.[0]?.op, "pullImage");
	assert.equal(pull.actions?.[0]?.ref, "alpine:latest");
	const listed = listImages(cli, { all: true, filter: { term: "reference", values: ["alpine*"] } });
	assert.equal(listed.actions?.[0]?.op, "listImages");
});

test("run uses setInvoke when provided", async () => {
	const cli = connect({ fromEnv: true });
	const doc = listImages(cli, { all: true });
	setInvoke(async (_op, payload) => payload);
	try {
		const out = await run(doc);
		assert.equal((out as { actions: { op: string }[] }).actions[0].op, "listImages");
	} finally {
		setInvoke(undefined);
	}
});

test("run without invoke fails when the binary is missing", async () => {
	const cli = connect({ fromEnv: true });
	const doc = listImages(cli, { all: true });
	const prev = process.env.CONTAINERKIT_BIN;
	process.env.CONTAINERKIT_BIN = "/no-such-containerkit";
	setInvoke(undefined);
	try {
		await run(doc);
		assert.ok(false, "expected run() to throw");
	} catch (err) {
		assert.match(String(err), /CONTAINERKIT_BIN|setInvoke/);
	} finally {
		if (prev === undefined) {
			delete process.env.CONTAINERKIT_BIN;
		} else {
			process.env.CONTAINERKIT_BIN = prev;
		}
	}
});
