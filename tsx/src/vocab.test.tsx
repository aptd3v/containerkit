/** @jsxImportSource containerkit */
import assert from "node:assert/strict";
import { test } from "node:test";
import type { StopSignalProps } from "./docker-file.js";
import { Compose, Publish, Restart, Service, Volume } from "./generated.js";
import type { DependsOnProps, DeployProps, RollbackProps, ServiceProps, VolumeProps, WatchFilesProps } from "./generated.js";
import type { ComposeBuildOptions, ConnectOptions, EngineWaitOptions } from "./generated-options.js";
import { render } from "./render.js";
import type * as vocabPath from "containerkit/vocab";
import {
	cap,
	cascade,
	cgroupns,
	depend,
	df,
	endpoint,
	failact,
	ipc,
	isolation,
	netmode,
	pidns,
	propagate,
	progress,
	proto,
	pullPolicy,
	recursive,
	restartPolicy,
	scheme,
	selinux,
	signal,
	userns,
	uts,
	vocab,
	vol,
	waitFor,
	watchAction,
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

function _closedVocab() {
	const service: ServiceProps = {
		isolation: isolation.HyperV,
		cgroupNS: cgroupns.Private,
		networkMode: netmode.container("abc"),
	};
	const mount: VolumeProps = {
		kind: vol.Tmpfs,
		propagation: propagate.RPrivate,
		recursive: recursive.ReadOnly,
	};
	const edge: DependsOnProps = { condition: depend.CompletedSuccessfully };
	const deploy: DeployProps = { endpointMode: endpoint.DNSRR };
	const rollback: RollbackProps = { failureAction: failact.Rollback };
	const stop: StopSignalProps = { signal: signal.SIGTERM };
	const badKind: VolumeProps = {
		// @ts-expect-error kind is bind, volume, tmpfs, npipe, cluster, or image
		kind: "directory",
	};
	const badSignal: StopSignalProps = {
		// @ts-expect-error signal is a POSIX signal name
		signal: "nope",
	};
	const namespaces: ServiceProps = {
		userNS: userns.Host,
		uts: uts.Host,
		pullPolicy: pullPolicy.every("24h"),
		networkMode: netmode.service("api"),
		ipc: ipc.container("abc"),
		pid: pidns.container("abc"),
	};
	const badPull: ServiceProps = {
		// @ts-expect-error pullPolicy is a known policy or every_<period>
		pullPolicy: "sometimes",
	};
	const badNet: ServiceProps = {
		// @ts-expect-error networkMode is bridge, host, none, default, container:<id>, or service:<name>
		networkMode: "bridg",
	};
	const badIPC: ServiceProps = {
		// @ts-expect-error ipc is none, private, shareable, host, or container:<id>
		ipc: "shared",
	};
	const badPID: ServiceProps = {
		// @ts-expect-error pid is host or container:<id>
		pid: "private",
	};
	const watch: WatchFilesProps = { action: watchAction.SyncRestart };
	const client: ConnectOptions = { scheme: scheme.Unix };
	const build: ComposeBuildOptions = { progress: progress.Plain };
	const waiting: EngineWaitOptions = { waitFor: waitFor.Removed };
	const deployMode: DeployProps = { mode: "replicated" };
	const badProgress: ComposeBuildOptions = {
		// @ts-expect-error progress is auto, plain, tty, quiet, or json
		progress: "verbose",
	};
	const badScheme: ConnectOptions = {
		// @ts-expect-error scheme is unix, npipe, tcp, http, https, ssh, or fd
		scheme: "ws",
	};
	const badAction: WatchFilesProps = {
		// @ts-expect-error action is sync, rebuild, sync+restart, sync+exec, or restart
		action: "restart-now",
	};
	const badWait: EngineWaitOptions = {
		// @ts-expect-error waitFor is not-running, next-exit, or removed
		waitFor: "healthy",
	};
	const badUserNS: ServiceProps = {
		// @ts-expect-error userNS is host
		userNS: "private",
	};
	void service;
	void mount;
	void edge;
	void deploy;
	void rollback;
	void stop;
	void badKind;
	void badSignal;
	void namespaces;
	void watch;
	void client;
	void build;
	void waiting;
	void deployMode;
	void badProgress;
	void badScheme;
	void badAction;
	void badWait;
	void badUserNS;
	void badPull;
	void badNet;
	void badIPC;
	void badPID;
}
void _closedVocab;
