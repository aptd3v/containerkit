import assert from "node:assert/strict";
import { test } from "node:test";
import { element, flatten, isNode, isReactElement } from "./node.js";
import { render } from "./render.js";
import { Fragment, jsx } from "./jsx-runtime.js";

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

test("Fragment is kind fragment", () => {
	const n = Fragment({ children: element("service", { name: "a" }) });
	assert.equal(n.$$kind, "fragment");
	assert.equal(n.children[0].$$kind, "service");
});
