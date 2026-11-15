import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import { Compose, Container, Env, Expose, Publish, Service } from "./generated.js";
import { jsx } from "./jsx-runtime.js";
import { type Node } from "./node.js";
import { exportYAML, render } from "./render.js";

const here = dirname(fileURLToPath(import.meta.url));

test("render emits compose-spec YAML for a service tree", () => {
	const yaml = render(
		jsx(Compose, {
			name: "demo",
			children: jsx(Service, {
				name: "web",
				image: "nginx:latest",
				children: [
					jsx(Publish, { hostPort: 8080, containerPort: 80, protocol: "tcp" }),
					jsx(Env, { name: "FOO", value: "bar" }),
				],
			}),
		}),
	);
	assert.match(yaml, /^name: demo$/m);
	assert.match(yaml, /web:/);
	assert.match(yaml, /image: "?nginx:latest"?/);
	assert.match(yaml, /published: "8080"/);
	assert.match(yaml, /target: 80/);
	assert.match(yaml, /FOO: bar/);
});

test("standalone Container wraps as a single-service compose file", () => {
	const yaml = render(jsx(Container, { name: "worker", image: "alpine:latest" }));
	assert.match(yaml, /services:/);
	assert.match(yaml, /worker:/);
	assert.match(yaml, /image: "?alpine:latest"?/);
});

test("profile and expose use compose-spec fields", () => {
	const yaml = render(
		jsx(Compose, {
			name: "demo",
			children: jsx(Service, {
				name: "api",
				image: "nginx:latest",
				profile: "minimal",
				children: jsx(Expose, { port: 8000, proto: "tcp" }),
			}),
		}),
	);
	assert.match(yaml, /profiles:\n\s+- minimal/);
	assert.match(yaml, /expose:\n\s+- 8000\/tcp/);
	assert.equal(yaml.includes("\n  profile:"), false);
});

test("securityOpt renders as a compose list", () => {
	const yaml = render(
		jsx(Compose, {
			name: "demo",
			children: jsx(Service, {
				name: "vector",
				image: "alpine:latest",
				securityOpt: "label=disable",
			}),
		}),
	);
	assert.match(yaml, /security_opt:\n\s+- label=disable/);
});

test("invalid nesting throws at render", () => {
	const bad: Node = {
		$$kind: "compose",
		props: { name: "bad" },
		children: [{ $$kind: "publish", props: { containerPort: 80 }, children: [] }],
	};
	assert.throws(() => render(bad), /<publish> is not valid inside <compose>/);
});

test("exportYAML prints a named Compose and does not accept another root", () => {
	const tree = jsx(Compose, {
		name: "app",
		children: jsx(Service, { name: "api", image: "alpine:latest" }),
	});
	const yaml = exportYAML(tree);
	assert.match(yaml, /^name: app$/m);
	assert.match(yaml, /image: "?alpine:latest"?/);
	assert.throws(() => exportYAML(jsx(Service, { name: "api", image: "alpine:latest" })), /Compose node/);
	assert.throws(() => exportYAML(jsx(Container, { image: "alpine:latest" })), /Compose node/);
	assert.throws(() => exportYAML(jsx(Compose, {})), /requires name/);
});

test("example fixture matches App.tsx render shape", () => {
	const fixture = readFileSync(resolve(here, "..", "testdata", "example.yaml"), "utf8");
	assert.match(fixture, /name: app/);
	assert.match(fixture, /alpine:latest/);
	assert.match(fixture, /published: "8080"/);
	assert.match(fixture, /healthcheck:/);
	assert.match(fixture, /networks:/);
	assert.match(fixture, /volumes:/);
});
