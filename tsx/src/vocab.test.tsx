/** @jsxImportSource containerkit */
import assert from "node:assert/strict";
import { test } from "node:test";
import { Compose, Publish, Restart, Service, Volume } from "./generated.js";
import { render } from "./render.js";
import type * as vocabPath from "containerkit/vocab";
import {
	cap,
	cascade,
	depend,
	df,
	ipc,
	netmode,
	proto,
	pullPolicy,
	restartPolicy,
	selinux,
	signal,
	vocab,
	vol,
	waitFor,
} from "./vocab.js";

type _VocabPath = typeof vocabPath.cap;
const _vocabPath: _VocabPath = cap;
void _vocabPath;

test("vocabulary matches the values props and options expect", () => {
	assert.equal(cap.NET_ADMIN, "NET_ADMIN");
	assert.equal(cap.ALL, "ALL");
	assert.equal(proto.TCP, "tcp");
	assert.equal(proto.UDP, "udp");
	assert.equal(restartPolicy.Always, "always");
	assert.equal(restartPolicy.OnFailure, "on-failure");
	assert.equal(pullPolicy.IfNotPresent, "if_not_present");
	assert.equal(pullPolicy.every("24h"), "every_24h");
	assert.equal(netmode.Bridge, "bridge");
	assert.equal(netmode.container("abc"), "container:abc");
	assert.equal(netmode.service("api"), "service:api");
	assert.equal(ipc.container("abc"), "container:abc");
	assert.equal(depend.Healthy, "service_healthy");
	assert.equal(vol.Pipe, "npipe");
	assert.equal(selinux.Shared, "z");
	assert.equal(selinux.Private, "Z");
	assert.equal(signal.SIGTERM, "SIGTERM");
	assert.equal(df.BuildCache, "build-cache");
	assert.equal(waitFor.NotRunning, "not-running");
	assert.equal(cascade.Ignore, 0);
	assert.equal(cascade.Stop, 1);
	assert.equal(cascade.Fail, 2);
	assert.equal(vocab.cap, cap);
	assert.equal(vocab.restartPolicy, restartPolicy);
});

test("service props accept vocabulary values", () => {
	const yaml = render(
		<Compose name="app">
			<Service
				name="api"
				image="alpine:latest"
				capAdd={cap.NET_ADMIN}
				capDrop={cap.MKNOD}
				networkMode={netmode.Bridge}
				stopSignal={signal.SIGTERM}
			>
				<Publish containerPort={80} protocol={proto.TCP} />
				<Restart policy={restartPolicy.Always} />
				<Volume source="/data" target="/data" kind={vol.Bind} />
			</Service>
		</Compose>,
	);
	assert.match(yaml, /cap_add: NET_ADMIN/);
	assert.match(yaml, /cap_drop: MKNOD/);
	assert.match(yaml, /network_mode: bridge/);
	assert.match(yaml, /stop_signal: SIGTERM/);
	assert.match(yaml, /protocol: tcp/);
	assert.match(yaml, /restart: always/);
	assert.match(yaml, /type: bind/);
});
