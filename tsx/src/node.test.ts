import assert from "node:assert/strict";
import { test } from "node:test";
import { Build, Compose, Env, Health, Publish, Restart, Service } from "./generated.js";
import { Command, CommandShell, Dockerfile, From, Run, Syntax } from "./docker-file.js";
import { Fragment, jsx } from "./jsx-runtime.js";
import { element, flatten, isNode, isReactElement } from "./node.js";
import { render } from "./render.js";

test("flatten drops false, null, and boolean", () => {
	const kids = flatten([false, null, undefined, true, element("env", { name: "A", value: "1" })]);
	assert.equal(kids.length, 1);
	assert.equal(kids[0].$$kind, "env");
});

test("flatten unwraps fragments and nested arrays", () => {
	const inner = element("env", { name: "B", value: "2" });
	const frag = element("fragment", { children: inner });
	const kids = flatten([[frag], [inner]]);
	assert.equal(kids.length, 2);
	assert.equal(kids[0].$$kind, "env");
	assert.equal(kids[1].$$kind, "env");
});

test("flatten rejects scalar strings", () => {
	assert.throws(() => flatten("nope"), /invalid child/);
});

test("React elements are rejected instead of rendered", () => {
	const el = { $$typeof: Symbol.for("react.element"), type: "div", props: { children: "hi" } };
	assert.equal(isReactElement(el), true);
	assert.throws(() => flatten(el), /compiled with React/);
	assert.throws(() => render(el as never), /compiled with React/);
});

test("jsx calls function components", () => {
	function Box(props: Record<string, unknown>) {
		return element("service", props);
	}
	const n = jsx(Box, { name: "api", image: "alpine" });
	assert.equal(n.$$kind, "service");
	assert.equal(n.props.name, "api");
	assert.ok(isNode(n));
});

test("a Dockerfile Run is rejected when the Service tree is built", () => {
	assert.throws(
		() =>
			// @ts-expect-error dfRun is not a Service child
			jsx(Service, {
				name: "api",
				image: "alpine",
				children: Run({ cmd: "echo" }),
			}),
		/<dfRun> is not valid inside <service>/,
	);
	assert.throws(
		() =>
			Service({
				name: "api",
				image: "alpine",
				children: Fragment({ children: Run({ cmd: "echo" }) }),
			}),
		/<dfRun> is not valid inside <service>/,
	);
});

test("Run is rejected under Build and kept under Dockerfile", () => {
	assert.throws(
		() =>
			Build({
				// @ts-expect-error dfRun is not a Build child
				children: Run({ cmd: "echo" }),
			}),
		/<dfRun> is not valid inside <build>/,
	);
	const file = Dockerfile({
		children: [From({ image: "nginx", tag: "alpine" }), Run({ cmd: "echo" })],
	});
	assert.deepEqual(
		file.children.map((child) => child.$$kind),
		["dfFrom", "dfRun"],
	);
});

test("Health, Publish, and Build stay Service children", () => {
	const node = jsx(Service, {
		name: "api",
		image: "alpine",
		children: [
			Health({ test: ["CMD", "true"] }),
			Publish({ hostPort: 80, containerPort: 80 }),
			Build({
				children: Dockerfile({
					children: [From({ image: "nginx", tag: "alpine" }), Run({ cmd: "echo" })],
				}),
			}),
		],
	});
	assert.deepEqual(
		node.children.map((child) => child.$$kind),
		["health", "publish", "build"],
	);
	assert.equal(node.children[2].children[0].$$kind, "dockerfile");
});

test("a hand-built node still fails when it is rendered", () => {
	const bad = {
		$$kind: "service",
		props: { name: "api" },
		children: [{ $$kind: "dfRun", props: { cmd: "echo" }, children: [] }],
	};
	assert.throws(() => render(bad), /<dfRun> is not valid inside <service>/);
});

test("a leaf, a Dockerfile, and Compose reject a child when the tree is built", () => {
	assert.throws(
		() =>
			Health({
				test: ["CMD", "true"],
				// @ts-expect-error env is not a Health child
				children: Env({ name: "A", value: "1" }),
			}),
		/<env> is not valid inside <health>/,
	);
	assert.throws(
		() =>
			Publish({
				hostPort: 80,
				containerPort: 80,
				// @ts-expect-error env is not a Publish child
				children: Env({ name: "A", value: "1" }),
			}),
		/<env> is not valid inside <publish>/,
	);
	assert.throws(
		() =>
			Dockerfile({
				children: Health({ test: ["CMD", "true"] }),
			}),
		/<health> is not valid inside <dockerfile>/,
	);
	assert.throws(
		() =>
			// @ts-expect-error dfRun is not a Compose child
			Compose({ name: "ck", children: Run({ cmd: "echo" }) }),
		/<dfRun> is not valid inside <compose>/,
	);
	assert.throws(
		() =>
			Run({
				cmd: "echo",
				// @ts-expect-error dfFrom is not a Run child
				children: From({ image: "alpine" }),
			}),
		/<dfFrom> is not valid inside <dfRun>/,
	);
});

test("a child Go stores once cannot be repeated", () => {
	const health = () => Health({ test: ["CMD", "true"] });
	assert.throws(
		() => Service({ name: "api", image: "alpine", children: [health(), health()] }),
		/<health> is already set inside <service>/,
	);
	assert.throws(
		() =>
			Service({
				name: "api",
				image: "alpine",
				children: [Restart({ policy: "no" }), Restart({ policy: "always" })],
			}),
		/<restart> is already set inside <service>/,
	);
	assert.throws(
		() =>
			Build({
				children: [
					Dockerfile({ children: From({ image: "alpine" }) }),
					Dockerfile({ children: From({ image: "nginx" }) }),
				],
			}),
		/Build accepts one Dockerfile/,
	);
	assert.throws(
		() =>
			Build({
				dockerfilePath: "Dockerfile",
				children: Dockerfile({ children: From({ image: "alpine" }) }),
			}),
		/Build accepts one Dockerfile/,
	);
	const publish = () => Publish({ hostPort: 80, containerPort: 80 });
	const repeated = Service({ name: "api", image: "alpine", children: [publish(), publish()] });
	assert.deepEqual(
		repeated.children.map((child) => child.$$kind),
		["publish", "publish"],
	);
});

test("Dockerfile directives and CMD are set once", () => {
	assert.throws(
		() =>
			Dockerfile({
				children: [Syntax({ value: "docker/dockerfile:1" }), Syntax({ value: "docker/dockerfile:1.4" })],
			}),
		/syntax directive is already set/,
	);
	assert.throws(
		() =>
			Dockerfile({
				children: [Command({ executable: "true" }), CommandShell({ executable: "false" })],
			}),
		/command has already been set/,
	);
	const stages = Dockerfile({
		children: [From({ image: "golang:1.22" }), From({ image: "alpine" })],
	});
	assert.deepEqual(
		stages.children.map((child) => child.$$kind),
		["dfFrom", "dfFrom"],
	);
});

test("a hand-built leaf child and a repeated restart fail at render", () => {
	const health = {
		$$kind: "service",
		props: { name: "api", image: "alpine" },
		children: [
			{
				$$kind: "health",
				props: { test: ["CMD", "true"] },
				children: [{ $$kind: "env", props: { name: "A", value: "1" }, children: [] }],
			},
		],
	};
	assert.throws(() => render(health), /<env> is not valid inside <health>/);
	const restarted = {
		$$kind: "service",
		props: { name: "api", image: "alpine" },
		children: [
			{ $$kind: "restart", props: { policy: "no" }, children: [] },
			{ $$kind: "restart", props: { policy: "always" }, children: [] },
		],
	};
	assert.throws(() => render(restarted), /<restart> is already set inside <service>/);
});

test("Fragment is kind fragment", () => {
	const n = Fragment({ children: element("service", { name: "a" }) });
	assert.equal(n.$$kind, "fragment");
	assert.equal(n.children[0].$$kind, "service");
});
