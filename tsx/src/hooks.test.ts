import assert from "node:assert/strict";
import { test } from "node:test";
import { Compose, Service } from "./generated.js";
import { useEnv, useLiveLog, useServiceState, useShutdown } from "./hooks/index.js";
import { ensureImage, hostPlatform, stopAndRemove } from "./live/helpers.js";
import { ContainerKitError } from "./live/error.js";
import { root, runRoot } from "./live/root.js";
import type { LiveEvent, Session, SessionTransport } from "./live/session.js";
import type { PlanDocument } from "./plan.js";

const tree = Compose({
	name: "hooks",
	children: [Service({ name: "api", image: "alpine:latest" })],
});

let lane: Promise<void> = Promise.resolve();

function oneAtATime(name: string, fn: () => Promise<void>): void {
	test(name, async () => {
		const previous = lane;
		let release: () => void = () => undefined;
		lane = new Promise<void>((resolve) => {
			release = resolve;
		});
		await previous;
		try {
			await fn();
		} finally {
			release();
		}
	});
}

oneAtATime("useShutdown takes the project down once", async () => {
	const calls: PlanDocument[] = [];
	const transport: SessionTransport = {
		async call(plan) {
			calls.push(plan);
			return { ok: true };
		},
		on() {
			return () => undefined;
		},
		async close() {
			return undefined;
		},
	};
	let release: () => void = () => undefined;
	const gate = new Promise<void>((resolve) => {
		release = resolve;
	});
	let dispose: () => void = () => undefined;
	const live = root({}, async function App() {
		dispose = useShutdown(tree, { removeOrphans: true, volumes: true });
		await gate;
		return null;
	});
	const exit = process.exit;
	let exits = 0;
	process.exit = (() => {
		exits += 1;
	}) as typeof process.exit;
	const pending = runRoot(live, transport);
	try {
		await new Promise((resolve) => setImmediate(resolve));
		process.emit("SIGINT");
		process.emit("SIGINT");
		await new Promise((resolve) => setImmediate(resolve));
		const downs = calls.filter((plan) => plan.actions?.[0]?.op === "composeDown");
		assert.equal(downs.length, 1);
		const options = downs[0]?.actions?.[0]?.options as { volumes?: boolean; removeOrphans?: boolean } | undefined;
		assert.equal(options?.volumes, true);
		assert.equal(options?.removeOrphans, true);
		assert.equal(exits, 1);
	} finally {
		dispose();
		process.exit = exit;
		release();
		const running = await pending;
		await running.close();
	}
});

oneAtATime("useShutdown downs the project root applied", async () => {
	const calls: PlanDocument[] = [];
	const transport: SessionTransport = {
		async call(plan) {
			calls.push(plan);
			return { ok: true };
		},
		on() {
			return () => undefined;
		},
		async close() {
			return undefined;
		},
	};
	let dispose: () => void = () => undefined;
	const live = root({}, function App() {
		dispose = useShutdown({ removeOrphans: true, volumes: true });
		return tree;
	});
	const exit = process.exit;
	process.exit = (() => undefined) as typeof process.exit;
	try {
		const running = await runRoot(live, transport);
		process.emit("SIGINT");
		await new Promise((resolve) => setImmediate(resolve));
		const downs = calls.filter((plan) => plan.actions?.[0]?.op === "composeDown");
		assert.equal(downs.length, 1);
		assert.match(String(downs[0]?.yaml), /api/);
		const options = downs[0]?.actions?.[0]?.options as { volumes?: boolean; removeOrphans?: boolean } | undefined;
		assert.equal(options?.volumes, true);
		assert.equal(options?.removeOrphans, true);
		await running.close();
	} finally {
		dispose();
		process.exit = exit;
	}
});

oneAtATime("useLiveLog hears log and status events", async () => {
	const seen: LiveEvent[] = [];
	let emit: (event: LiveEvent) => void = () => undefined;
	const transport: SessionTransport = {
		async call() {
			return { ok: true };
		},
		on(kind, handler) {
			if (kind === "event") {
				emit = handler;
			}
			return () => undefined;
		},
		async close() {
			return undefined;
		},
	};
	const live = root({}, function App() {
		useLiveLog((event) => {
			seen.push(event);
		});
		return null;
	});
	const running = await runRoot(live, transport);
	try {
		emit({ type: "event", kind: "log", line: "hello" });
		emit({ type: "event", kind: "status", status: "working" });
		emit({ type: "event", kind: "error", message: "no" });
		assert.deepEqual(
			seen.map((event) => event.kind),
			["log", "status"],
		);
	} finally {
		await running.close();
	}
});

oneAtATime("useEnv reads the variable once", async () => {
	const previous = process.env.CONTAINERKIT_HOOK_ENV;
	process.env.CONTAINERKIT_HOOK_ENV = "edge";
	const seen: string[] = [];
	const transport: SessionTransport = {
		async call() {
			return { ok: true };
		},
		on() {
			return () => undefined;
		},
		async close() {
			return undefined;
		},
	};
	const live = root({}, function App() {
		seen.push(useEnv("CONTAINERKIT_HOOK_ENV", "fallback"));
		seen.push(useEnv("CONTAINERKIT_HOOK_MISSING", "fallback"));
		return null;
	});
	const running = await runRoot(live, transport);
	try {
		assert.deepEqual(seen, ["edge", "fallback"]);
	} finally {
		if (previous === undefined) {
			delete process.env.CONTAINERKIT_HOOK_ENV;
		} else {
			process.env.CONTAINERKIT_HOOK_ENV = previous;
		}
		await running.close();
	}
});

oneAtATime("useServiceState keeps the latest status for one service", async () => {
	const states: Array<LiveEvent | undefined> = [];
	let emit: (event: LiveEvent) => void = () => undefined;
	const transport: SessionTransport = {
		async call() {
			return { ok: true };
		},
		on(kind, handler) {
			if (kind === "event") {
				emit = handler;
			}
			return () => undefined;
		},
		async close() {
			return undefined;
		},
	};
	const live = root({}, function App() {
		states.push(useServiceState("nginx"));
		return null;
	});
	const running = await runRoot(live, transport);
	try {
		emit({ type: "event", kind: "status", source: "api", status: "working" });
		emit({ type: "event", kind: "status", source: "Container nginx", status: "done", text: "Healthy" });
		for (let i = 0; i < 5 && states.at(-1)?.source !== "Container nginx"; i++) {
			await new Promise((resolve) => setImmediate(resolve));
		}
		assert.equal(states[0], undefined);
		assert.equal(states.at(-1)?.status, "done");
		assert.equal(states.at(-1)?.text, "Healthy");
	} finally {
		await running.close();
	}
});

test("ensureImage pulls only a missing image", async () => {
	let pulled = 0;
	const missing = {
		async inspectImage() {
			throw new ContainerKitError({ type: "error", class: "notFound", message: "missing" });
		},
		async pullImage() {
			pulled += 1;
		},
	} as unknown as Session;
	await ensureImage(missing, "alpine:latest", "linux/amd64");
	assert.equal(pulled, 1);
	const present = {
		async inspectImage() {
			return {};
		},
		async pullImage() {
			pulled += 1;
		},
	} as unknown as Session;
	await ensureImage(present, "alpine:latest");
	assert.equal(pulled, 1);
});

test("stopAndRemove ignores a container that is already gone", async () => {
	const calls: string[] = [];
	const session = {
		async stop(id: string) {
			calls.push(`stop ${id}`);
			throw new ContainerKitError({ type: "error", class: "notFound", message: "missing" });
		},
		async remove(id: string) {
			calls.push(`remove ${id}`);
		},
	} as unknown as Session;
	await stopAndRemove(session, "web");
	assert.deepEqual(calls, ["stop web", "remove web"]);
});

test("hostPlatform names this machine", () => {
	assert.match(hostPlatform(), /^linux\/.+/);
});
