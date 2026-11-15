import assert from "node:assert/strict";
import { test } from "node:test";
import { Compose, Service } from "./generated.js";
import { commandByName } from "./live/commands.js";
import { connect } from "./plan.js";
import { Always, Build, ComposeStep, Container, Flow, If, Retry, Subflow, Tag, stepsOf } from "./taskflow.js";

test("flow document encodes the CI graph", () => {
	const cli = connect({ fromEnv: true });
	const db = Compose({ name: "app", children: Service({ name: "db", image: "postgres:16" }) });
	const doc = Flow(cli, {
		name: "ci",
		steps: [
			Build({ name: "build", image: "app:ci", source: ".", dependsOn: undefined }),
			ComposeStep({
				name: "up-db",
				node: db,
				op: "composeUp",
				options: { detach: true, wait: true },
				dependsOn: ["build"],
				service: "db",
				port: 5432,
				repository: "app",
			}),
			Container({
				name: "test",
				image: "app:ci",
				command: ["go", "test", "./..."],
				dependsOn: ["up-db"],
				retry: 1,
				priority: "high",
			}),
			If({
				name: "release",
				when: process.env.RELEASE === "1",
				dependsOn: ["test"],
				then: [Tag({ name: "tag", image: "app:ci", always: true, precede: ["test"] })],
				else: [Tag({ name: "keep", source: "app:ci", target: "app:ci" })],
			}),
			Subflow({ name: "inner", steps: [Build({ name: "again", image: "app:ci" })], priority: "low" }),
		],
	});
	assert.equal(doc.connect?.fromEnv, true);
	assert.equal(doc.flow?.name, "ci");
	const steps = doc.flow?.steps as Array<Record<string, unknown>>;
	assert.equal(steps[0].op, "buildImage");
	assert.deepEqual(steps[0].options, { image: "app:ci", source: "." });
	assert.equal(steps[1].op, "composeUp");
	assert.match(String(steps[1].yaml), /postgres:16/);
	assert.deepEqual(steps[1].options, { detach: true, wait: true });
	assert.deepEqual(steps[1].dependsOn, ["build"]);
	assert.equal(steps[2].op, "containerRun");
	assert.equal(steps[3].op, "if");
	assert.equal(steps[3].when, false);
	const then = steps[3].then as Array<Record<string, unknown>>;
	assert.equal(then[0].op, "tagImage");
	assert.deepEqual(then[0].options, { source: "app:ci", target: "app:ci" });
	const kept = Retry(2, Always(Build({ name: "retry", image: "app:ci" })));
	assert.equal(kept.retry, 2);
	assert.equal(kept.always, true);
});

test("flow accepts children and default name", () => {
	const cli = connect({});
	const one = Build({ name: "only", image: "app" });
	const doc = Flow(cli, { children: [one, Container({ name: "run", image: "app", precede: ["only"] })] });
	assert.equal(doc.flow?.name, "taskflow");
	assert.equal(doc.flow?.steps.length, 2);
	assert.deepEqual(stepsOf(undefined), []);
	assert.deepEqual(stepsOf(one), [one]);
	const named = Flow(cli, { name: "other", steps: [] });
	assert.equal(named.flow?.name, "other");
	assert.equal(typeof commandByName.flow, "function");
});
