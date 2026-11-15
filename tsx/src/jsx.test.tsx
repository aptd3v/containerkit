/** @jsxImportSource containerkit */
import assert from "node:assert/strict";
import { test } from "node:test";
import {
	Build,
	Compose,
	Container,
	Deploy,
	Env,
	Label,
	Health,
	Job,
	Network,
	Publish,
	Service,
	Volume,
} from "./generated.js";
import type { DockerNode } from "./docker-node.js";
import { jsx } from "./jsx-runtime.js";
import { BuildArg } from "./env.js";
import { render } from "./render.js";

function _invalidNestingTypeCheck() {
	const _props: import("./generated.js").ComposeProps = {
		// @ts-expect-error Publish is not valid inside Compose
		children: Publish({ containerPort: 80 }),
	};
	void _props;
	// @ts-expect-error Publish is not valid inside Compose
	void jsx(Compose, {
		name: "bad",
		children: Publish({ containerPort: 80 }),
	});
}
void _invalidNestingTypeCheck;

function _htmlIsNotIntrinsic() {
	// @ts-expect-error HTML tags are not containerkit elements
	void (<div />);
}
void _htmlIsNotIntrinsic;

function _documentedJsx(): DockerNode<Compose<{ name?: string }>> {
	return (
		<Compose name="app">
			<Service name="api" image="alpine:latest" />
		</Compose>
	);
}
void _documentedJsx;

function _rootComponent(): DockerNode {
	return (
		<Compose name="app">
			<Service name="api" image="alpine:latest" />
		</Compose>
	);
}
void _rootComponent;

function _roots(): void {
	const compose: DockerNode = Compose({ name: "app" });
	const service: DockerNode = Service({ name: "api", image: "alpine:latest" });
	const job: DockerNode = Job({ name: "migrate" });
	const container: DockerNode = Container({ image: "alpine:latest" });
	void compose;
	void service;
	void job;
	void container;
	// @ts-expect-error Network is not a Compose, Service, Job, or Container root
	const network: DockerNode = Network({ name: "front" });
	void network;
}
void _roots;

function _composeNode(): DockerNode<Compose<{ name?: string }>> {
	const tree = Compose({
		name: "app",
		children: Service({ name: "api", image: "alpine:latest" }),
	});
	const _name: string | undefined = tree.props.name;
	const _kind: "compose" = tree.$$kind;
	void _name;
	void _kind;
	// @ts-expect-error image is not a Compose prop
	void tree.props.image;
	return tree;
}
void _composeNode;

function _jsxComposeNode(): DockerNode<Compose<{ name?: string }>> {
	const tree = jsx(Compose, {
		name: "app",
		children: jsx(Service, { name: "api", image: "alpine:latest" }),
	});
	const _name: string | undefined = tree.props.name;
	const _kind: "compose" = tree.$$kind;
	void _name;
	void _kind;
	return tree;
}
void _jsxComposeNode;

function _serviceIsNotCompose(): DockerNode<Compose<{ name?: string }>> {
	// @ts-expect-error Service is not a Compose node
	return Service({ name: "api", image: "alpine:latest" });
}
void _serviceIsNotCompose;

function _jsxServiceIsNotCompose(): DockerNode<Compose<{ name?: string }>> {
	// @ts-expect-error Service is not a Compose node
	return jsx(Service, { name: "api", image: "alpine:latest" });
}
void _jsxServiceIsNotCompose;

function _serviceProps(): DockerNode<Service<{ name?: string; image?: string }>> {
	const tree: DockerNode<Service<{ name?: string; image?: string }>> = Service({
		name: "api",
		image: "alpine:latest",
	});
	const _image: string | undefined = tree.props.image;
	void _image;
	// @ts-expect-error this Service props type has no command
	void tree.props.command;
	return tree;
}
void _serviceProps;

function _composeTagRejectsImage() {
	// @ts-expect-error image is not a Compose prop
	void (<Compose name="app" image="nope" />);
}
void _composeTagRejectsImage;

test("control flow && omits false branches", () => {
	const prod = false;
	const yaml = render(
		<Compose name="app">
			<Service name="api" image="alpine:latest">
				{prod && <Deploy replicas={3} />}
				<Env name="A" value="1" />
			</Service>
		</Compose>,
	);
	assert.equal(yaml.includes("deploy:"), false);
	assert.match(yaml, /A: "1"|A: 1/);
});

test("map emits repeated Env children", () => {
	const vars = [
		["FOO", "1"],
		["BAR", "2"],
	] as const;
	const yaml = render(
		<Compose name="app">
			<Service name="api" image="alpine:latest">
				{vars.map(([name, value]) => (
					<Env name={name} value={value} />
				))}
			</Service>
		</Compose>,
	);
	assert.match(yaml, /FOO:/);
	assert.match(yaml, /BAR:/);
});

test("function components compose as children", () => {
	function API(props: { tag: string }) {
		return (
			<Service name="api" image={`alpine:${props.tag}`}>
				<Publish hostPort={8080} containerPort={80} />
			</Service>
		);
	}
	const yaml = render(
		<Compose name="app">
			<API tag="3.20" />
		</Compose>,
	);
	assert.match(yaml, /alpine:3.20/);
	assert.match(yaml, /published: "8080"/);
});

test("Health and Volume nest under Service", () => {
	const yaml = render(
		<Compose name="app">
			<Service name="api" image="alpine:latest">
				<Volume source="data" target="/data" />
				<Health test={["CMD", "true"]} interval="5s" timeout="3s" retries={3} />
				<Build source="." dockerfilePath="Dockerfile" />
			</Service>
		</Compose>,
	);
	assert.match(yaml, /healthcheck:/);
	assert.match(yaml, /interval: "?5s"?/);
	assert.match(yaml, /target: \/data/);
	assert.match(yaml, /context: \./);
});

test("example App tree renders a compose file", () => {
	const prod = true;
	const yaml = render(
		<Compose name="app">
			<Network name="front" driver="bridge" />
			<Volume name="data" />
			<Service name="api" image="alpine:latest" command={["sleep", "infinity"]}>
				<Publish hostPort={8080} containerPort={80} protocol="tcp" />
				<Env name="APP" value="1" />
				<Volume source="data" target="/data" />
				<Health test={["CMD", "true"]} interval="5s" timeout="3s" retries={3} />
				{prod && <Deploy replicas={1} />}
				<Build source="." dockerfilePath="Dockerfile" />
				<Network name="front" />
			</Service>
		</Compose>,
	);
	assert.match(yaml, /^name: app$/m);
	assert.match(yaml, /published: "8080"/);
	assert.match(yaml, /healthcheck:/);
	assert.match(yaml, /deploy:/);
	assert.match(yaml, /replicas: 1/);
	assert.match(yaml, /driver: bridge/);
});

test("env children and value are the same node", () => {
	const url = "postgres://harbor:harbor@db:5432/harbor";
	const fromProp = <Env name="DATABASE_URL" value={url} />;
	const fromChild = <Env name="DATABASE_URL">{url}</Env>;
	assert.deepEqual(fromChild, fromProp);
	assert.deepEqual(fromProp, {
		$$kind: "env",
		props: { name: "DATABASE_URL", value: url },
		children: [],
	});
	assert.deepEqual(<Env name="EMPTY" value="" />, <Env name="EMPTY">{""}</Env>);
	assert.equal((<Env name="FLAG">false</Env>).props.value, "false");
	assert.equal((<Env name="COUNT">0</Env>).props.value, "0");
	assert.equal((<Env name="PAD">  x  </Env>).props.value, "  x  ");
	assert.equal((<Env name="SPACE"> </Env>).props.value, " ");

	const yaml = render(
		<Compose name="app">
			<Service name="api" image="alpine:latest">
				<Env name="FLAG">false</Env>
				<Env name="COUNT">0</Env>
				<Env name="PAD">  x  </Env>
				<Env name="EMPTY" value="" />
			</Service>
		</Compose>,
	);
	assert.match(yaml, /FLAG: "false"/);
	assert.match(yaml, /COUNT: "0"/);
	assert.match(yaml, /PAD: "  x  "/);
	assert.match(yaml, /EMPTY: ""/);

	assert.throws(() => Env({ name: "FOO" } as never), /requires a value/);
	assert.throws(() => Env({ name: "FOO", value: "a", children: "b" } as never), /not both/);
	assert.throws(() => Env({ name: "FOO", children: 0 } as never), /single string/);
	assert.throws(() => Env({ name: "FOO", children: ["a", "b"] } as never), /single string/);
	assert.throws(() => Env({ name: "FOO", children: false } as never), /requires a value/);
});

test("label and build arg children match the value prop", () => {
	assert.deepEqual(<Label name="com.harbor.role">notes</Label>, <Label name="com.harbor.role" value="notes" />);
	assert.equal((<Label name="FLAG">false</Label>).props.value, "false");
	assert.equal((<Label name="COUNT">0</Label>).props.value, "0");
	assert.equal((<Label name="PAD">  x  </Label>).props.value, "  x  ");

	const fromProp = <BuildArg name="APP_ENV" value="production" />;
	const fromChild = <BuildArg name="APP_ENV">production</BuildArg>;
	assert.deepEqual(fromChild, fromProp);
	assert.deepEqual(fromChild, <Env name="APP_ENV" value="production" />);
	assert.equal(fromChild.$$kind, "env");

	assert.throws(() => Label({ name: "k", value: "a", children: "b" } as never), /not both/);
	assert.throws(() => Label({ name: "k", children: ["a", "b"] } as never), /single string/);
	assert.throws(() => BuildArg({ name: "k", value: "a", children: "b" } as never), /not both/);
	assert.throws(() => Label({ name: "k" } as never), /requires a value/);
});

function _labelRequiresOneString() {
	// @ts-expect-error value and children are both set
	void (<Label name="k" value="a">b</Label>);
}
void _labelRequiresOneString;

function _envRequiresOneString() {
	// @ts-expect-error a name alone is not a value
	void (<Env name="FOO" />);
	// @ts-expect-error value and children are both set
	void (<Env name="FOO" value="a">b</Env>);
}
void _envRequiresOneString;
