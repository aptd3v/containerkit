import { Build, Compose, Network, Publish, Service } from "containerkit";
import { Command, Dockerfile, From, Run, Workdir } from "containerkit/dockerfile";
import { root, runIfMain } from "containerkit/live";
import { netDriver, proto } from "containerkit/vocab";

// Build, up, ps, stop, start, restart, exec, images, pause, run, then down.
// The Go program is examples/compose_commands. compose top is a Go method;
// the session lists the same containers with composePs.

const appService = "app";
const webService = "web";
const redisService = "redis";

const tree = (
	<Compose name="compose-commands-example">
		<Network name="backend" driver={netDriver.Bridge} />
		<Service name={appService}>
			<Build>
				<Dockerfile>
					<From image="alpine" tag="latest" />
					<Workdir path="/app" />
					<Run cmd='echo "Hello, World!"' />
					<Command executable="tail" args={["-f", "/dev/null"]} />
				</Dockerfile>
			</Build>
			<Network name="backend" />
		</Service>
		<Service name={webService} image="nginx:alpine" command={["nginx", "-g", "daemon off;"]}>
			<Publish protocol={proto.TCP} hostIP="0.0.0.0" hostPort={9080} containerPort={80} />
			<Network name="backend" />
		</Service>
		<Service name={redisService} image="redis:7-alpine" command={["redis-server"]}>
			<Network name="backend" />
		</Service>
	</Compose>
);

export { tree };

const live = root({ fromEnv: true, onExit: "down" }, async function App(session) {
	console.error("Building app service...");
	await session.composeBuild(tree, { services: appService });

	console.error("Starting stack...");
	await session.composeUp(tree, { detach: true, removeOrphans: true });

	console.error("--- docker compose ps -a ---");
	console.log(JSON.stringify(await session.composePs(tree, { all: true }), null, 2));

	console.error("Stopping web...");
	await session.composeStop(tree, { services: webService, timeout: "5s" });
	console.log(JSON.stringify(await session.composePs(tree, { all: true }), null, 2));

	console.error("Starting web...");
	await session.composeStart(tree, { services: webService });

	console.error("Restarting redis...");
	await session.composeRestart(tree, { services: redisService, noDeps: true, timeout: "5s" });

	console.error("--- docker compose exec ---");
	const execCode = await session.composeExec(tree, appService, {
		command: ["sh", "-c", "echo hello from compose exec"],
		tty: false,
	});
	console.log(`exec exit ${execCode}`);

	console.error("--- docker compose images ---");
	console.log(JSON.stringify(await session.composeImages(tree), null, 2));

	console.error("--- docker compose pause / unpause (web) ---");
	await session.composePause(tree, { services: webService });
	await session.composeUnpause(tree, { services: webService });

	console.error("--- docker compose run ---");
	const runCode = await session.composeRun(tree, appService, {
		command: ["echo", "hello from compose run"],
		tty: false,
		autoRemove: true,
		noDeps: true,
	});
	console.log(`run exit ${runCode}`);

	console.error("Stack running. The session takes it down on exit.");
	return tree;
});

export default live;

await runIfMain(live, import.meta.url);
