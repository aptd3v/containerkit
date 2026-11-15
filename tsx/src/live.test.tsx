process.env.CONTAINERKIT_MEMO = "0";

import assert from "node:assert/strict";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { Compose, Service } from "./generated.js";
import { ContainerKitError, isCKLoginRequiredError, isCKNotFoundError, isCKUnavailableError, start } from "./live/index.js";
import { runRoot, root, useEffect, useMemo, useSession, useState } from "./live/root.js";
import { openSession, type Session, type SessionTransport } from "./live/session.js";
import type { PlanDocument } from "./plan.js";

type Memory = {
	calls: PlanDocument[];
	closed: boolean;
	transport: SessionTransport;
};

function memoryTransport(opts?: { login?: boolean }): Memory {
	const calls: PlanDocument[] = [];
	const mem: Memory = {
		calls,
		closed: false,
		transport: {
			async call(plan) {
				calls.push(plan);
				const action = plan.actions?.[0];
				if (opts?.login && action?.op === "pullImage" && calls.filter((c) => c.actions?.[0]?.op === "pullImage").length === 1) {
					throw new ContainerKitError({ type: "error", class: "loginRequired", message: "login required" });
				}
				if (action?.op === "pullImage" && action.ref === "missing") {
					throw new ContainerKitError({ type: "error", class: "notFound", message: "missing" });
				}
				if (action?.op === "pullImage" && action.ref === "down") {
					throw new ContainerKitError({ type: "error", class: "unavailable", message: "cannot reach the docker daemon" });
				}
				return { ok: true };
			},
			on() {
				return () => undefined;
			},
			async close() {
				mem.closed = true;
			},
		},
	};
	return mem;
}

function ops(calls: PlanDocument[], op: string): number {
	return calls.filter((plan) => plan.actions?.[0]?.op === op).length;
}

async function pullBase(app: Session, ref: string) {
	try {
		return await app.pullImage(ref);
	} catch (err) {
		if (!isCKLoginRequiredError(err)) {
			throw err;
		}
		await app.login({ serverAddress: "ghcr.io" });
		return app.pullImage(ref);
	}
}

let serial: Promise<void> = Promise.resolve();

function oneAtATime(name: string, fn: () => Promise<void>) {
	test(name, () => {
		const run = serial.then(fn, fn);
		serial = run.then(() => undefined, () => undefined);
		return run;
	});
}

oneAtATime("state restarts the stack and useMemo skips the pull", async () => {
	const mem = memoryTransport();
	let setReplicas: (value: number) => void = () => undefined;
	const live = root({ fromEnv: true }, function App() {
		const app = useSession();
		const [replicas, set] = useState(1);
		setReplicas = set;
		useMemo(() => app.pullImage("alpine:latest"), ["alpine:latest"]);
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" replicas={replicas} />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.equal(ops(mem.calls, "pullImage"), 1);
		assert.equal(mem.calls.find((plan) => plan.actions?.[0]?.op === "pullImage")?.actions?.[0]?.ref, "alpine:latest");
		assert.equal(mem.calls.find((plan) => plan.actions?.[0]?.op === "pullImage")?.connect?.fromEnv, true);
		assert.equal(ops(mem.calls, "composeUp"), 1);
		setReplicas(2);
		await new Promise((resolve) => setImmediate(resolve));
		assert.equal(ops(mem.calls, "pullImage"), 1);
		assert.equal(ops(mem.calls, "composeUp"), 2);
	} finally {
		await running.close();
	}
});

oneAtATime("a changed memo dep pulls again", async () => {
	const mem = memoryTransport();
	let setRef: (value: string) => void = () => undefined;
	const live = root({ fromEnv: true }, function App() {
		const app = useSession();
		const [ref, set] = useState("alpine:latest");
		setRef = set;
		useMemo(() => pullBase(app, ref), [ref]);
		return (
			<Compose name="app">
				<Service name="api" image={ref} />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.equal(ops(mem.calls, "pullImage"), 1);
		setRef("alpine:3");
		await new Promise((resolve) => setImmediate(resolve));
		assert.equal(ops(mem.calls, "pullImage"), 2);
		assert.equal(ops(mem.calls, "composeUp"), 2);
	} finally {
		await running.close();
	}
});

oneAtATime("loginRequired is handled inside the memoized client function", async () => {
	const mem = memoryTransport({ login: true });
	const live = root({ fromEnv: true }, function App() {
		const app = useSession();
		useMemo(() => pullBase(app, "alpine:latest"), ["alpine:latest"]);
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.equal(ops(mem.calls, "pullImage"), 2);
		assert.equal(ops(mem.calls, "login"), 1);
		assert.equal(ops(mem.calls, "composeUp"), 1);
	} finally {
		await running.close();
	}
});

oneAtATime("a rejected memo is not cached and does not restart", async () => {
	const mem = memoryTransport();
	const errors: unknown[] = [];
	let tries = 0;
	let setTick: (value: number) => void = () => undefined;
	const live = root({ onError: (err) => errors.push(err) }, function App() {
		const app = useSession();
		const [tick, set] = useState(0);
		setTick = set;
		useMemo(() => {
			tries += 1;
			void tick;
			return app.pullImage("missing");
		}, ["missing"]);
		return null;
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.equal(tries, 1);
		assert.equal(ops(mem.calls, "composeUp"), 0);
		assert.equal(isCKNotFoundError(errors[0]), true);
		setTick(1);
		await new Promise((resolve) => setImmediate(resolve));
		assert.equal(tries, 2);
		assert.equal(ops(mem.calls, "composeUp"), 0);
	} finally {
		await running.close();
	}
});

function pullLatest(app: Session) {
	return app.pullImage("alpine:latest");
}

oneAtATime("memo is read back from disk for the next root", async () => {
	const dir = mkdtempSync(join(tmpdir(), "containerkit-memo-"));
	const first = memoryTransport();
	const live = root({ cacheDir: dir }, function App() {
		const app = useSession();
		const deps = { a: "alpine:latest", b: 1 };
		useMemo(() => pullLatest(app), [deps]);
		return null;
	});
	const running = await runRoot(live, first.transport);
	await running.close();
	assert.equal(ops(first.calls, "pullImage"), 1);

	const second = memoryTransport();
	let setOrder: (value: number) => void = () => undefined;
	const again = root({ cacheDir: dir }, function App() {
		const app = useSession();
		const [order, set] = useState(0);
		setOrder = set;
		const deps = order === 0 ? { a: "alpine:latest", b: 1 } : { b: 1, a: "alpine:latest" };
		useMemo(() => pullLatest(app), [deps]);
		return null;
	});
	const restarted = await runRoot(again, second.transport);
	try {
		assert.equal(ops(second.calls, "pullImage"), 0);
		setOrder(1);
		await new Promise((resolve) => setImmediate(resolve));
		assert.equal(ops(second.calls, "pullImage"), 0);
	} finally {
		await restarted.close();
	}
});

oneAtATime("unavailable closes the session", async () => {
	const mem = memoryTransport();
	const live = root({}, function App() {
		const app = useSession();
		useMemo(() => app.pullImage("down"), ["down"]);
		return null;
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.equal(mem.closed, true);
		assert.equal(isCKUnavailableError(new ContainerKitError({ type: "error", class: "unavailable", message: "down" })), true);
	} finally {
		await running.close();
	}
});

oneAtATime("effects clean up when the session closes", async () => {
	let mounted = 0;
	let cleaned = 0;
	const mem = memoryTransport();
	const live = root({}, function App() {
		useEffect(() => {
			mounted += 1;
			return () => {
				cleaned += 1;
			};
		}, []);
		return null;
	});
	const running = await runRoot(live, mem.transport);
	assert.equal(mounted, 1);
	assert.equal(cleaned, 0);
	await running.close();
	assert.equal(cleaned, 1);
});

test("hooks throw outside the root", () => {
	assert.throws(() => useSession(), /docker\.ts root/);
	assert.throws(() => useState(1), /docker\.ts root/);
	assert.throws(() => useMemo(() => 1, []), /docker\.ts root/);
	assert.throws(() => useEffect(() => undefined, []), /docker\.ts root/);
});

oneAtATime("a second root throws until the first closes", async () => {
	const mem = memoryTransport();
	const first = root({}, () => null);
	assert.throws(() => root({}, () => null), /second root/);
	const running = await runRoot(first, mem.transport);
	await running.close();
	const second = root({}, () => null);
	const again = await runRoot(second, mem.transport);
	await again.close();
});

test("start requires docker.ts", async () => {
	const dir = mkdtempSync(join(tmpdir(), "containerkit-live-"));
	await assert.rejects(start(join(dir, "app.ts")), /docker\.ts/);
	await assert.rejects(start(join(dir, "docker.ts")), /required/);
});

test("session client reads a result frame", async () => {
	const fake = new URL("../testdata/fake-containerkit.mjs", import.meta.url);
	const transport = openSession({ bin: process.execPath, args: [fake.pathname] });
	const value = (await transport.call({ actions: [{ op: "ping" }] })) as { ok: boolean; op: string };
	assert.equal(value.ok, true);
	assert.equal(value.op, "ping");
	await transport.close();
});

test("isCKNotFoundError rejects other errors", () => {
	assert.equal(isCKNotFoundError(new Error("nope")), false);
	assert.equal(isCKNotFoundError(new ContainerKitError({ type: "error", class: "unavailable", message: "down" })), false);
});

function sessionMethodTypes(app: Session) {
	const tree = Compose({ name: "app" });
	return Promise.all([
		app.pullImage("alpine:latest"),
		app.pullImage("alpine:latest", { platform: "linux/arm64" }),
		app.login({ serverAddress: "ghcr.io" }),
		app.composeUp(tree, { detach: true }),
		app.ping(),
		app.containerRun({ image: "alpine:latest", command: ["echo", "ok"] }),
	]);
}
void sessionMethodTypes;
