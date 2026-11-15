import {
	Compose,
	DependsOn,
	Env,
	Health,
	Publish,
	Restart,
	Service,
	Volume,
	connect,
	encode,
	type Client,
	type DockerNode,
	type PlanDocument,
} from "containerkit";
import { Command, Dockerfile, From, Label, Run, exportDockerfile } from "containerkit/dockerfile";
import { Always, Build, ComposeStep, Container, Flow, If, Tag } from "containerkit/taskflow";
import { depend, proto, restartPolicy } from "containerkit/vocab";

// A plan action list runs each call after the previous one returns. This
// release is not a list:
//
//   build the image ──┐
//   start Postgres  ──┘  both can start now
//   lint               waits for the image, then overlaps Postgres startup; it never opens a database
//   unit               waits for the image, overlaps lint and migrate
//   migrate            waits for the image and for Postgres to be healthy
//   tag                waits for lint, unit, and migrate, and is skipped if any failed
//   down               waits only for migrate, so unit or lint may still be running.
//                      Those checks do not use Postgres. down still runs when migrate failed.
//
// Compose DependsOn only orders services inside one project. It does not build
// the image from this Dockerfile, retry migrate, skip the tag, or take the
// database down on the failure path. Those edges are the flow. The Dockerfile,
// the Compose file, the vocabulary, and connect() are the same values a
// one-shot plan would carry.
//
// Stdout is the plan document. `containerkit exec -` runs it. RELEASE=1
// selects the release tag; any other value keeps a non-release tag.

const imageCI = process.env.IMAGE?.trim() || "ck-taskflow-ci:latest";
const imageRelease = process.env.RELEASE_TAG?.trim() || "ck-taskflow-ci:release";
const imageKeep = process.env.KEEP_TAG?.trim() || "ck-taskflow-ci:keep";

function AppImage() {
	return (
		<Dockerfile>
			<From image="alpine" tag="latest" />
			<Run cmd="printf 'ok\\n' > /app" />
			<Label name="com.containerkit.flow" value="ci" />
			<Command executable="cat" args={["/app"]} />
		</Dockerfile>
	);
}

function Stack(): DockerNode<Compose> {
	return (
		<Compose name="ck-taskflow">
			<Volume name="pgdata" />
			<Service name="db" image="postgres:16-alpine" containerName="ck-taskflow-db">
				<Env name="POSTGRES_PASSWORD">ci</Env>
				<Volume source="pgdata" target="/var/lib/postgresql/data" />
				<Publish hostPort={5432} containerPort={5432} protocol={proto.TCP} />
				<Restart policy={restartPolicy.UnlessStopped} />
				<Health test={["CMD", "pg_isready", "-U", "postgres"]} interval="2s" timeout="3s" retries={20} />
			</Service>
			<Service name="migrate" image={imageCI} command={["cat", "/app"]}>
				<DependsOn name="db" condition={depend.Healthy} />
				<Restart policy={restartPolicy.No} />
			</Service>
		</Compose>
	);
}

function pipeline(cli: Client): PlanDocument {
	const stack = Stack();
	const release = process.env.RELEASE === "1";
	return Flow(cli, {
		name: "ci",
		steps: [
			Build({
				name: "build",
				parseDockerfile: exportDockerfile(<AppImage />),
				tag: imageCI,
			}),
			ComposeStep({
				name: "up-db",
				node: stack,
				op: "composeUp",
				options: { detach: true, wait: true, services: ["db"] },
			}),
			Container({
				name: "unit",
				image: imageCI,
				command: ["cat", "/app"],
				dependsOn: ["build"],
				priority: "high",
			}),
			Container({
				name: "lint",
				image: imageCI,
				command: ["/bin/true"],
				dependsOn: ["build"],
				priority: "low",
			}),
			ComposeStep({
				name: "migrate",
				node: stack,
				op: "composeRun",
				service: "migrate",
				dependsOn: ["build", "up-db"],
				retry: 2,
			}),
			If({
				name: "release",
				when: release,
				dependsOn: ["unit", "lint", "migrate"],
				then: [Tag({ name: "tag", source: imageCI, target: imageRelease })],
				else: [Tag({ name: "keep", source: imageCI, target: imageKeep })],
			}),
			// unit and lint do not use Postgres, so they are not predecessors.
			// down can start while either check is still running.
			Always(
				ComposeStep({
					name: "down",
					node: stack,
					op: "composeDown",
					options: { volumes: true },
					dependsOn: ["migrate"],
				}),
			),
		],
	});
}

const doc = pipeline(connect({ fromEnv: true }));
console.error(
	"build overlaps up-db. lint waits for the image, then overlaps Postgres. migrate retries. a failed check skips the tag. down waits only for migrate.",
);
console.log(encode(doc));
