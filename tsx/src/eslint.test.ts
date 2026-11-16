import assert from "node:assert/strict";
import { test } from "node:test";
import { Linter } from "eslint";
import { parser } from "typescript-eslint";
import { composeComponents, dockerfileComponents } from "./generated.js";
import { recommended } from "./eslint.js";

const linter = new Linter({ configType: "flat" });

function messages(code: string): string[] {
	const result = linter.verify(code, [
		{
			files: ["**/*.{jsx,tsx}"],
			languageOptions: { parser, parserOptions: { ecmaFeatures: { jsx: true }, sourceType: "module" } },
		},
		...recommended,
	], { filename: "infra/app.tsx" });
	return result.map((item) => item.message);
}

test("component names come from the generator", () => {
	assert.equal(composeComponents.Service, "service");
	assert.equal(composeComponents.Health, "health");
	assert.equal(composeComponents.Dockerfile, "dockerfile");
	assert.equal(dockerfileComponents.Run, "dfRun");
	assert.equal(dockerfileComponents.Env, "dfEnv");
	assert.equal(dockerfileComponents.HealthCheck, "dfHealth");
});

test("an inline tag in the wrong parent is marked", () => {
	assert.deepEqual(messages(`
		import { Compose, Service, Health, Publish } from "containerkit";
		import { Dockerfile, From, Run } from "containerkit/dockerfile";
		export function App() {
			return (
				<Compose name="ck">
					<Service name="api" image="alpine">
						<Health test={["CMD", "true"]} />
						<Publish hostPort={80} containerPort={80} />
						<Dockerfile>
							<From image="nginx" tag="alpine" />
							<Run cmd="echo" />
						</Dockerfile>
					</Service>
				</Compose>
			);
		}
	`), [
		"<Dockerfile> is not valid inside <Service>",
	]);
});

test("the placement table flags the same nests the builder throws for", () => {
	assert.deepEqual(messages(`
		import { Compose, Service, Health, Publish, Build, Env } from "containerkit";
		import { Run, From, Dockerfile, Syntax, Command, CommandShell } from "containerkit/dockerfile";
		const flag = true;
		export function App() {
			return (
				<Compose name="ck">
					<Run cmd="echo" />
					<Service name="api" image="alpine">
						<Health test={["CMD", "true"]} />
						<Health test={["CMD", "true"]} />
						<Publish hostPort={1} containerPort={1} />
						<Publish hostPort={2} containerPort={2} />
						<Health test={["CMD", "true"]}>
							<Env name="A" value="1" />
						</Health>
						<Build dockerfilePath="Dockerfile">
							<Dockerfile>
								<From image="alpine" />
								<From image="nginx" />
								<Syntax value="docker/dockerfile:1" />
								<Syntax value="docker/dockerfile:1.4" />
								<Command executable="true" />
								<CommandShell executable="false" />
							</Dockerfile>
						</Build>
						{flag && <Run cmd="echo" />}
						<Service name="nested" image="alpine">
							<></>
							<Health test={["CMD", "true"]} />
							<Health test={["CMD", "true"]} />
						</Service>
					</Service>
				</Compose>
			);
		}
	`), [
		"<Run> is not valid inside <Compose>",
		"<Health> is already set inside <Service>",
		"<Health> is already set inside <Service>",
		"<Env> is not valid inside <Health>",
		"Build accepts one Dockerfile, and not alongside dockerfilePath, parseDockerfile, or readDockerfile",
		"syntax directive is already set",
		"command has already been set",
		"<Run> is not valid inside <Service>",
		"<Service> is not valid inside <Service>",
		"<Health> is already set inside <Service>",
	]);
});

test("a rename, a namespace, and a dockerfile Env are distinguished", () => {
	assert.deepEqual(messages(`
		import { Compose as App, Service } from "containerkit";
		import * as df from "containerkit/dockerfile";
		import { Env } from "containerkit/dockerfile";
		export function Demo() {
			return (
				<App name="ck">
					<df.Run cmd="echo" />
					<Service name="api" image="alpine">
						<Env name="A" value="1" />
					</Service>
				</App>
			);
		}
	`), [
		"<df.Run> is not valid inside <App>",
		"<Env> is not valid inside <Service>",
	]);
});

test("a child kept in a variable is invisible to the rule", () => {
	assert.deepEqual(messages(`
		import { Compose } from "containerkit";
		import { Run } from "containerkit/dockerfile";
		const cmd = <Run cmd="echo" />;
		export function App() {
			return <Compose name="ck">{cmd}</Compose>;
		}
	`), []);
});
