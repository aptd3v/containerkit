/** @jsxImportSource containerkit */
import assert from "node:assert/strict";
import { test } from "node:test";
import { Build, Compose, Service, connect, encode, buildDockerfile } from "./index.js";
import {
	Add,
	Arg,
	Command,
	Comment,
	Copy,
	Dockerfile,
	Entrypoint,
	Env,
	Expose,
	From,
	HealthCheck,
	Label,
	Onbuild,
	Run,
	RunArgs,
	StopSignal,
	User,
	Volume,
	Workdir,
	exportDockerfile,
} from "./docker-file.js";
import { render } from "./render.js";

test("Dockerfile TSX matches the kit builder", () => {
	const text = exportDockerfile(
		<Dockerfile>
			<From image="alpine" tag="latest" />
			<From image="golang:1.22" as="build" />
			<Arg name="VERSION" />
			<Arg name="APP" default="web" />
			<Env name="FOO" value="bar" />
			<Copy src="a.txt" dest="b.txt" />
			<Copy src="file with space.txt" dest="/dest with space" />
			<Add src="src" dest="dest" />
			<Add src="src file" dest="dest dir" />
			<Entrypoint executable="/bin/sh" args={["-c"]} />
			<Expose port="80" />
			<Label name="k" value="v" />
			<Onbuild command="RUN echo hi" />
			<Workdir path="/app" />
			<StopSignal signal="SIGTERM" />
			<User name="root" />
			<Comment text="note" />
			<Volume paths={["/data", "/logs"]} />
			<HealthCheck test={["CMD", "true"]} timeout="1s" interval="1s" retries={1} />
			<Run cmd="apk add --no-cache curl" />
			<Run cmd="echo still-run" />
			<RunArgs args={["--no-cache", "ok"]} />
			<Command executable="sleep" args={["infinity"]} />
		</Dockerfile>,
	);
	for (const want of [
		"FROM alpine:latest",
		"FROM golang:1.22 AS build",
		`COPY ["file with space.txt", "/dest with space"]`,
		`ADD ["src file", "dest dir"]`,
		"RUN apk add --no-cache curl && \\",
		`CMD ["sleep", "infinity"]`,
		`ENTRYPOINT ["/bin/sh", "-c"]`,
		`VOLUME ["/data", "/logs"]`,
		"--interval=1s",
		"--timeout=1s",
		"--retries=1",
		`CMD ["true"]`,
	]) {
		assert.ok(text.includes(want), `missing ${want} in:\n${text}`);
	}
	assert.equal(text.includes("--start-period="), false);
});

test("Build nests Dockerfile as dockerfile_inline", () => {
	const yaml = render(
		<Compose name="app">
			<Service name="api">
				<Build>
					<Dockerfile>
						<From image="alpine" tag="latest" />
						<Run cmd="echo ok" />
					</Dockerfile>
				</Build>
			</Service>
		</Compose>,
	);
	assert.match(yaml, /dockerfile_inline:/);
	assert.match(yaml, /FROM alpine:latest/);
	assert.match(yaml, /RUN echo ok/);
});

test("buildDockerfile encodes parseDockerfile for Go", () => {
	const plan = JSON.parse(
		encode(
			buildDockerfile(
				connect({ fromEnv: true }),
				<Dockerfile>
					<From image="alpine" tag="latest" />
					<Command executable="true" />
				</Dockerfile>,
				{ image: "app:dev" },
			),
		),
	) as { actions: { op: string; options: { parseDockerfile: string; image: string } }[] };
	assert.equal(plan.actions[0].op, "buildImage");
	assert.equal(plan.actions[0].options.image, "app:dev");
	assert.match(plan.actions[0].options.parseDockerfile, /^FROM alpine:latest\nCMD \["true"\]\n$/);
});

test("dockerfile scalar children match props", () => {
	assert.deepEqual(<Env name="FOO">bar</Env>, <Env name="FOO" value="bar" />);
	assert.deepEqual(<Label name="k">v</Label>, <Label name="k" value="v" />);
	assert.deepEqual(<Workdir>/app</Workdir>, <Workdir path="/app" />);
	assert.deepEqual(<User>1000:1000</User>, <User name="1000:1000" />);
	assert.equal((<Env name="FLAG">false</Env>).props.value, "false");
	assert.equal((<Env name="COUNT">0</Env>).props.value, "0");
	assert.equal((<Env name="PAD">  x  </Env>).props.value, "  x  ");

	const bare = <Arg name="APP_ENV" />;
	const emptyProp = <Arg name="APP_ENV" default="" />;
	const emptyChild = <Arg name="APP_ENV">{""}</Arg>;
	assert.deepEqual(emptyChild, emptyProp);
	assert.notDeepEqual(bare, emptyProp);
	assert.equal("default" in bare.props, false);
	assert.equal(emptyProp.props.default, "");
	assert.deepEqual(<Arg name="APP_ENV">production</Arg>, <Arg name="APP_ENV" default="production" />);

	const text = (node: Parameters<typeof exportDockerfile>[0]) => exportDockerfile(node);
	assert.equal(
		text(<Dockerfile><Env name="FOO">bar</Env></Dockerfile>),
		text(<Dockerfile><Env name="FOO" value="bar" /></Dockerfile>),
	);
	assert.equal(text(<Dockerfile><Arg name="APP_ENV" /></Dockerfile>), "ARG APP_ENV\n");
	assert.equal(text(<Dockerfile><Arg name="APP_ENV" default="" /></Dockerfile>), "ARG APP_ENV=\n");
	assert.equal(text(<Dockerfile><Arg name="APP_ENV">{""}</Arg></Dockerfile>), "ARG APP_ENV=\n");
	assert.equal(text(<Dockerfile><Arg name="APP_ENV">production</Arg></Dockerfile>), "ARG APP_ENV=production\n");

	assert.throws(() => Env({ name: "FOO", value: "a", children: "b" } as never), /not both/);
	assert.throws(() => Workdir({ path: "/app", children: "/src" } as never), /not both/);
	assert.throws(() => User({} as never), /requires a name/);
	assert.throws(() => Env({ name: "FOO", children: ["a", "b"] } as never), /single string/);
});

test("a second CMD is rejected", () => {
	assert.throws(
		() =>
			exportDockerfile(
				<Dockerfile>
					<From image="alpine" tag="latest" />
					<Command executable="true" />
					<Command executable="false" />
				</Dockerfile>,
			),
		/command has already been set/,
	);
});
