process.env.CONTAINERKIT_MEMO = "0";

import assert from "node:assert/strict";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { Compose, Service } from "./generated.js";
import { ContainerKitError, isCKLoginRequiredError, isCKNotFoundError, isCKUnavailableError, start } from "./live/index.js";
import { runRoot, root, useEffect, useMemo, useSession, useState, type RootOptions } from "./live/root.js";
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

test("onExit rejects any value other than down", () => {
	assert.throws(() => root({ onExit: "stop" } as unknown as RootOptions, () => null), /onExit must be "down"/);
});

oneAtATime("onExit down takes down the project root applied", async () => {
	const mem = memoryTransport();
	const live = root({ onExit: "down" }, function App() {
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.equal(ops(mem.calls, "composeUp"), 1);
		process.emit("SIGINT");
		for (let i = 0; i < 20 && !mem.closed; i++) {
			await new Promise((resolve) => setImmediate(resolve));
		}
		const up = mem.calls.find((plan) => plan.actions?.[0]?.op === "composeUp");
		const down = mem.calls.filter((plan) => plan.actions?.[0]?.op === "composeDown");
		assert.equal(down.length, 1);
		assert.equal(down[0]?.yaml, up?.yaml);
		assert.equal(mem.closed, true);
		process.emit("SIGINT");
		await new Promise((resolve) => setImmediate(resolve));
		assert.equal(ops(mem.calls, "composeDown"), 1);
	} finally {
		await running.close();
	}
});

oneAtATime("onExit down waits for an apply that is still running", async () => {
	let releaseUp: () => void = () => undefined;
	const gate = new Promise<void>((resolve) => {
		releaseUp = resolve;
	});
	const calls: PlanDocument[] = [];
	let closed = false;
	const transport: SessionTransport = {
		async call(plan) {
			calls.push(plan);
			if (plan.actions?.[0]?.op === "composeUp") {
				await gate;
			}
			return { ok: true };
		},
		on() {
			return () => undefined;
		},
		async close() {
			closed = true;
		},
	};
	const live = root({ onExit: "down" }, function App() {
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const pending = runRoot(live, transport);
	await new Promise((resolve) => setImmediate(resolve));
	assert.equal(ops(calls, "composeUp"), 1);
	process.emit("SIGINT");
	process.emit("SIGTERM");
	await new Promise((resolve) => setImmediate(resolve));
	assert.equal(ops(calls, "composeDown"), 0);
	assert.equal(closed, false);
	releaseUp();
	const running = await pending;
	try {
		for (let i = 0; i < 20 && !closed; i++) {
			await new Promise((resolve) => setImmediate(resolve));
		}
		assert.equal(ops(calls, "composeDown"), 1);
		assert.equal(calls.find((plan) => plan.actions?.[0]?.op === "composeDown")?.yaml, calls.find((plan) => plan.actions?.[0]?.op === "composeUp")?.yaml);
		assert.equal(closed, true);
	} finally {
		await running.close();
	}
});

oneAtATime("onExit down with no project closes the session", async () => {
	const mem = memoryTransport();
	const live = root({ onExit: "down" }, () => null);
	const running = await runRoot(live, mem.transport);
	try {
		process.emit("SIGTERM");
		for (let i = 0; i < 20 && !mem.closed; i++) {
			await new Promise((resolve) => setImmediate(resolve));
		}
		assert.equal(ops(mem.calls, "composeDown"), 0);
		assert.equal(mem.closed, true);
	} finally {
		await running.close();
	}
});

oneAtATime("close leaves the project up when onExit is down", async () => {
	const mem = memoryTransport();
	const live = root({ onExit: "down" }, function App() {
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	await running.close();
	assert.equal(ops(mem.calls, "composeUp"), 1);
	assert.equal(ops(mem.calls, "composeDown"), 0);
	assert.equal(mem.closed, true);
});

oneAtATime("a failed down still closes the session", async () => {
	const errors: unknown[] = [];
	const calls: PlanDocument[] = [];
	let closed = false;
	const transport: SessionTransport = {
		async call(plan) {
			calls.push(plan);
			if (plan.actions?.[0]?.op === "composeDown") {
				throw new Error("down failed");
			}
			return { ok: true };
		},
		on() {
			return () => undefined;
		},
		async close() {
			closed = true;
		},
	};
	const live = root({ onExit: "down", onError: (err) => errors.push(err) }, function App() {
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const running = await runRoot(live, transport);
	try {
		process.emit("SIGINT");
		for (let i = 0; i < 20 && !closed; i++) {
			await new Promise((resolve) => setImmediate(resolve));
		}
		assert.equal(ops(calls, "composeDown"), 1);
		assert.equal(errors.length, 1);
		assert.equal(closed, true);
	} finally {
		await running.close();
	}
});

oneAtATime("an async root runs once and applies what it returns", async () => {
	const mem = memoryTransport();
	let calls = 0;
	const live = root({ fromEnv: true }, async function App(session) {
		calls += 1;
		await session.pullImage("alpine:latest");
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.equal(calls, 1);
		assert.equal(ops(mem.calls, "pullImage"), 1);
		assert.equal(mem.calls.find((plan) => plan.actions?.[0]?.op === "pullImage")?.actions?.[0]?.ref, "alpine:latest");
		assert.equal(ops(mem.calls, "composeUp"), 1);
		assert.deepEqual(mem.calls.find((plan) => plan.actions?.[0]?.op === "composeUp")?.actions?.[0]?.options, { detach: true });
		assert.match(mem.calls.find((plan) => plan.actions?.[0]?.op === "composeUp")?.yaml ?? "", /name: app/);
	} finally {
		await running.close();
	}
});

oneAtATime("an async root does not apply when the callback fails", async () => {
	const mem = memoryTransport();
	const errors: unknown[] = [];
	const live = root({ onError: (err) => errors.push(err) }, async function App(session) {
		await session.pullImage("missing");
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.equal(ops(mem.calls, "pullImage"), 1);
		assert.equal(ops(mem.calls, "composeUp"), 0);
		assert.equal(isCKNotFoundError(errors[0]), true);
		assert.equal(mem.closed, false);
	} finally {
		await running.close();
	}
});

oneAtATime("a failed once callback exits when the file is docker.js", async () => {
	const mem = memoryTransport();
	const prevArgv = process.argv[1];
	const prevExit = process.exit;
	let code: number | undefined;
	process.exit = ((status?: number) => {
		code = status ?? 0;
		throw new Error(`exit ${status}`);
	}) as typeof process.exit;
	const errors: unknown[] = [];
	try {
		const live = root({ onError: (err) => errors.push(err) }, async function App(session) {
			process.argv[1] = "/tmp/docker.js";
			await session.pullImage("missing");
			return (
				<Compose name="app">
					<Service name="api" image="alpine:latest" />
				</Compose>
			);
		});
		await assert.rejects(runRoot(live, mem.transport), /exit 1/);
		assert.equal(code, 1);
		assert.equal(ops(mem.calls, "composeUp"), 0);
		assert.equal(mem.closed, true);
		assert.equal(isCKNotFoundError(errors[0]), true);
		process.argv[1] = prevArgv;
		const again = root({}, () => null);
		const running = await runRoot(again, memoryTransport().transport);
		await running.close();
	} finally {
		process.exit = prevExit;
		process.argv[1] = prevArgv;
	}
});

oneAtATime("a failed render stays open when the file is docker.js", async () => {
	const mem = memoryTransport();
	const prevArgv = process.argv[1];
	const prevExit = process.exit;
	let code: number | undefined;
	process.exit = ((status?: number) => {
		code = status ?? 0;
		throw new Error(`exit ${status}`);
	}) as typeof process.exit;
	const errors: unknown[] = [];
	try {
		const live = root({ onError: (err) => errors.push(err) }, () => {
			process.argv[1] = "/tmp/docker.js";
			throw new Error("render broke");
		});
		const running = await runRoot(live, mem.transport);
		assert.equal(code, undefined);
		assert.equal((errors[0] as Error).message, "render broke");
		assert.equal(mem.closed, false);
		await running.close();
	} finally {
		process.exit = prevExit;
		process.argv[1] = prevArgv;
	}
});

oneAtATime("onUp runs after a waited apply", async () => {
	const mem = memoryTransport();
	const seen: string[] = [];
	const live = root({ wait: true, onUp: () => seen.push("up") }, async function App() {
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	try {
		assert.deepEqual(seen, ["up"]);
		assert.deepEqual(mem.calls.find((plan) => plan.actions?.[0]?.op === "composeUp")?.actions?.[0]?.options, {
			detach: true,
			wait: true,
		});
	} finally {
		await running.close();
	}
});

oneAtATime("a session callback does not render again", async () => {
	const mem = memoryTransport();
	let calls = 0;
	let setTick: (value: number) => void = () => undefined;
	const live = root({}, function App(session) {
		calls += 1;
		void session;
		const [, set] = useState(0);
		setTick = set;
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const running = await runRoot(live, mem.transport);
	try {
		setTick(1);
		await new Promise((resolve) => setImmediate(resolve));
		assert.equal(calls, 1);
		assert.equal(ops(mem.calls, "composeUp"), 1);
	} finally {
		await running.close();
	}
});

oneAtATime("onExit during an async pull closes without applying", async () => {
	let releasePull: () => void = () => undefined;
	const gate = new Promise<void>((resolve) => {
		releasePull = resolve;
	});
	const calls: PlanDocument[] = [];
	let closed = false;
	const transport: SessionTransport = {
		async call(plan) {
			calls.push(plan);
			if (plan.actions?.[0]?.op === "pullImage") {
				await gate;
			}
			return { ok: true };
		},
		on() {
			return () => undefined;
		},
		async close() {
			closed = true;
		},
	};
	const live = root({ onExit: "down" }, async function App(session) {
		await session.pullImage("alpine:latest");
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const pending = runRoot(live, transport);
	await new Promise((resolve) => setImmediate(resolve));
	assert.equal(ops(calls, "pullImage"), 1);
	process.emit("SIGINT");
	await new Promise((resolve) => setImmediate(resolve));
	assert.equal(ops(calls, "composeUp"), 0);
	assert.equal(closed, false);
	releasePull();
	const running = await pending;
	try {
		for (let i = 0; i < 20 && !closed; i++) {
			await new Promise((resolve) => setImmediate(resolve));
		}
		assert.equal(ops(calls, "composeUp"), 0);
		assert.equal(ops(calls, "composeDown"), 0);
		assert.equal(closed, true);
	} finally {
		await running.close();
	}
});

oneAtATime("onExit during an async apply takes the project down", async () => {
	let releaseUp: () => void = () => undefined;
	const gate = new Promise<void>((resolve) => {
		releaseUp = resolve;
	});
	const calls: PlanDocument[] = [];
	let closed = false;
	const transport: SessionTransport = {
		async call(plan) {
			calls.push(plan);
			if (plan.actions?.[0]?.op === "composeUp") {
				await gate;
			}
			return { ok: true };
		},
		on() {
			return () => undefined;
		},
		async close() {
			closed = true;
		},
	};
	const live = root({ onExit: "down" }, async function App() {
		return (
			<Compose name="app">
				<Service name="api" image="alpine:latest" />
			</Compose>
		);
	});
	const pending = runRoot(live, transport);
	await new Promise((resolve) => setImmediate(resolve));
	assert.equal(ops(calls, "composeUp"), 1);
	process.emit("SIGINT");
	await new Promise((resolve) => setImmediate(resolve));
	assert.equal(ops(calls, "composeDown"), 0);
	releaseUp();
	const running = await pending;
	try {
		for (let i = 0; i < 20 && !closed; i++) {
			await new Promise((resolve) => setImmediate(resolve));
		}
		assert.equal(ops(calls, "composeDown"), 1);
		assert.equal(
			calls.find((plan) => plan.actions?.[0]?.op === "composeDown")?.yaml,
			calls.find((plan) => plan.actions?.[0]?.op === "composeUp")?.yaml,
		);
		assert.equal(closed, true);
	} finally {
		await running.close();
	}
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

type ExactUnknown<T> = [unknown] extends [T] ? ([T] extends [unknown] ? true : false) : false;

function sessionMethodTypes(app: Session) {
	const tree = Compose({ name: "app" });
	const pullIsUnknown: ExactUnknown<Awaited<ReturnType<Session["pullImage"]>>> = true;
	const buildIsUnknown: ExactUnknown<Awaited<ReturnType<Session["composeBuild"]>>> = true;
	const logsAreUnknown: ExactUnknown<Awaited<ReturnType<Session["logs"]>>> = true;
	const attachIsUnknown: ExactUnknown<Awaited<ReturnType<Session["attachContainer"]>>> = true;
	const statsAreUnknown: ExactUnknown<Awaited<ReturnType<Session["statsContainer"]>>> = true;
	const copyIsUnknown: ExactUnknown<Awaited<ReturnType<Session["copyFrom"]>>> = true;
	const saveIsUnknown: ExactUnknown<Awaited<ReturnType<Session["saveImages"]>>> = true;
	const typed = {
		listImages: false as ExactUnknown<Awaited<ReturnType<Session["listImages"]>>>,
		inspectImage: false as ExactUnknown<Awaited<ReturnType<Session["inspectImage"]>>>,
		composePs: false as ExactUnknown<Awaited<ReturnType<Session["composePs"]>>>,
		composeExec: false as ExactUnknown<Awaited<ReturnType<Session["composeExec"]>>>,
		composeRun: false as ExactUnknown<Awaited<ReturnType<Session["composeRun"]>>>,
		composeWait: false as ExactUnknown<Awaited<ReturnType<Session["composeWait"]>>>,
		composePort: false as ExactUnknown<Awaited<ReturnType<Session["composePort"]>>>,
		composeViz: false as ExactUnknown<Awaited<ReturnType<Session["composeViz"]>>>,
		load: false as ExactUnknown<Awaited<ReturnType<Session["load"]>>>,
		create: false as ExactUnknown<Awaited<ReturnType<Session["create"]>>>,
		createExec: false as ExactUnknown<Awaited<ReturnType<Session["createExec"]>>>,
		createNetwork: false as ExactUnknown<Awaited<ReturnType<Session["createNetwork"]>>>,
		ping: false as ExactUnknown<Awaited<ReturnType<Session["ping"]>>>,
		listContainers: false as ExactUnknown<Awaited<ReturnType<Session["listContainers"]>>>,
		inspect: false as ExactUnknown<Awaited<ReturnType<Session["inspect"]>>>,
		listNetworks: false as ExactUnknown<Awaited<ReturnType<Session["listNetworks"]>>>,
		inspectNetwork: false as ExactUnknown<Awaited<ReturnType<Session["inspectNetwork"]>>>,
		listEngineVolumes: false as ExactUnknown<Awaited<ReturnType<Session["listEngineVolumes"]>>>,
		createVolume: false as ExactUnknown<Awaited<ReturnType<Session["createVolume"]>>>,
		searchImages: false as ExactUnknown<Awaited<ReturnType<Session["searchImages"]>>>,
		historyImage: false as ExactUnknown<Awaited<ReturnType<Session["historyImage"]>>>,
		topContainer: false as ExactUnknown<Awaited<ReturnType<Session["topContainer"]>>>,
		statPath: false as ExactUnknown<Awaited<ReturnType<Session["statPath"]>>>,
		composeList: false as ExactUnknown<Awaited<ReturnType<Session["composeList"]>>>,
		composeImages: false as ExactUnknown<Awaited<ReturnType<Session["composeImages"]>>>,
		composeVolumeList: false as ExactUnknown<Awaited<ReturnType<Session["composeVolumeList"]>>>,
		pruneImages: false as ExactUnknown<Awaited<ReturnType<Session["pruneImages"]>>>,
		pruneContainers: false as ExactUnknown<Awaited<ReturnType<Session["pruneContainers"]>>>,
		pruneNetworks: false as ExactUnknown<Awaited<ReturnType<Session["pruneNetworks"]>>>,
		pruneVolumes: false as ExactUnknown<Awaited<ReturnType<Session["pruneVolumes"]>>>,
		login: false as ExactUnknown<Awaited<ReturnType<Session["login"]>>>,
		diskUsage: false as ExactUnknown<Awaited<ReturnType<Session["diskUsage"]>>>,
		removeImage: false as ExactUnknown<Awaited<ReturnType<Session["removeImage"]>>>,
		listCheckpoints: false as ExactUnknown<Awaited<ReturnType<Session["listCheckpoints"]>>>,
		listConfigs: false as ExactUnknown<Awaited<ReturnType<Session["listConfigs"]>>>,
		inspectConfig: false as ExactUnknown<Awaited<ReturnType<Session["inspectConfig"]>>>,
		listSecrets: false as ExactUnknown<Awaited<ReturnType<Session["listSecrets"]>>>,
		inspectSecret: false as ExactUnknown<Awaited<ReturnType<Session["inspectSecret"]>>>,
		listPlugins: false as ExactUnknown<Awaited<ReturnType<Session["listPlugins"]>>>,
		inspectPlugin: false as ExactUnknown<Awaited<ReturnType<Session["inspectPlugin"]>>>,
		listNodes: false as ExactUnknown<Awaited<ReturnType<Session["listNodes"]>>>,
		inspectNode: false as ExactUnknown<Awaited<ReturnType<Session["inspectNode"]>>>,
		listSwarmServices: false as ExactUnknown<Awaited<ReturnType<Session["listSwarmServices"]>>>,
		inspectSwarmService: false as ExactUnknown<Awaited<ReturnType<Session["inspectSwarmService"]>>>,
		listTasks: false as ExactUnknown<Awaited<ReturnType<Session["listTasks"]>>>,
		inspectTask: false as ExactUnknown<Awaited<ReturnType<Session["inspectTask"]>>>,
		inspectSwarm: false as ExactUnknown<Awaited<ReturnType<Session["inspectSwarm"]>>>,
	};
	void pullIsUnknown;
	void buildIsUnknown;
	void logsAreUnknown;
	void attachIsUnknown;
	void statsAreUnknown;
	void copyIsUnknown;
	void saveIsUnknown;
	void typed;
	return Promise.all([
		app.pullImage("alpine:latest"),
		app.pullImage("alpine:latest", { platform: "linux/arm64" }),
		app.login({ serverAddress: "ghcr.io" }),
		app.composeUp(tree, { detach: true }),
		app.composeBuild(tree),
		app.containerRun({ image: "alpine:latest", command: ["echo", "ok"] }),
		app.listImages({ all: true }).then((images) => images.Items?.[0]?.RepoTags?.[0]),
		app.inspectImage("alpine:latest").then((image) => image.Id + image.Os),
		app.composePs(tree).then((rows) => {
			const row = rows?.[0];
			return row ? row.Service + row.State : "";
		}),
		app.listContainers({ all: true }).then((containers) => containers.Items?.[0]?.State),
		app.inspect("id").then((found) => found.Container.Id),
		app.createExec("id", { command: ["sh"] }).then((exec) => exec.ID),
		app.composePort(tree, "api", 80).then((published) => published[0]),
		app.ping().then((pong) => pong.APIVersion),
		app.pruneImages().then((report) => report.Report.SpaceReclaimed),
		app.searchImages("alpine").then((found) => found.Items?.[0]?.name),
		app.login().then((auth) => auth.Auth.Status),
		app.diskUsage().then((usage) => usage.Images.TotalSize),
		app.removeImage("app:dev").then((removed) => removed.Items?.[0]?.Deleted),
		app.listConfigs().then((configs) => configs.Items?.[0]?.ID),
		app.inspectConfig("id").then((found) => found.Config.Spec.Name),
		app.listPlugins().then((plugins) => plugins.Items?.[0]?.Name),
		app.inspectPlugin("name").then((found) => found.Plugin.Enabled),
		app.listNodes().then((nodes) => nodes.Items?.[0]?.Status?.State),
		app.inspectNode("id").then((found) => found.Node.Spec?.Role),
		app.listSecrets().then((secrets) => secrets.Items?.[0]?.ID),
		app.inspectSecret("id").then((found) => found.Secret.ID),
		app.listSwarmServices().then((services) => services.Items?.[0]?.Spec?.Name),
		app.inspectSwarmService("id").then((found) => found.Service.ID),
		app.listTasks().then((tasks) => tasks.Items?.[0]?.DesiredState),
		app.inspectTask("id").then((found) => found.Task.Status?.State),
		app.listCheckpoints("id").then((points) => points.Items?.[0]?.Name),
		app.inspectSwarm().then((swarm) => swarm.Swarm.JoinTokens.Worker),
	]);
}
void sessionMethodTypes;
