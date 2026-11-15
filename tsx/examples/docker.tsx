import { Build, Compose, Health, Publish, Service } from "containerkit";
import { Dockerfile, From, Run } from "containerkit/dockerfile";
import { isCKUnavailableError, root } from "containerkit/live";
import { proto } from "containerkit/vocab";

// A live session that runs once. The file is docker.tsx so the compiled entry
// is docker.js, which is one of the names root() starts by itself.
//
//   cd tsx
//   npm run live
//
// Docker must be running. The script builds cmd/containerkit and points
// CONTAINERKIT_BIN at it. composeBuild makes the image from the Dockerfile
// in the tree. root then brings that project up. onUp prints the address
// after the healthcheck passes. onExit: "down" takes the project down on
// Ctrl+C. LIVE_PORT chooses the host port (default 18080).

const port = Number(process.env.LIVE_PORT ?? "18080");

export default root(
	{
		fromEnv: true,
		onExit: "down",
		wait: true,
		onUp() {
			console.error(`ck-live http://127.0.0.1:${port}`);
		},
		onError(err) {
			const message = err instanceof Error ? err.message : String(err);
			console.error(message);
			if (isCKUnavailableError(err)) {
				console.error("Docker is not reachable. The session will close.");
			}
		},
	},
	async function App(session) {
		const tree = (
			<Compose name="ck-live">
				<Service name="api" image="ck-live:dev">
					<Build>
						<Dockerfile>
							<From image="nginx" tag="alpine" />
							<Run cmd="printf ok > /usr/share/nginx/html/index.html" />
						</Dockerfile>
					</Build>
					<Publish hostPort={port} containerPort={80} protocol={proto.TCP} />
					<Health
						test={["CMD", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1/"]}
						interval="1s"
						timeout="3s"
						retries={3}
						startPeriod="5s"
					/>
				</Service>
			</Compose>
		);
		await session.composeBuild(tree);
		return tree;
	},
);
