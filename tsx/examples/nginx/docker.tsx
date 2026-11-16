/** @jsxImportSource containerkit */
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { Build, Compose, Health, Publish, Service } from "containerkit";
import { Command, Copy, Dockerfile, Expose, From } from "containerkit/dockerfile";
import { useShutdown } from "containerkit/hooks";
import { root } from "containerkit/live";
import { proto } from "containerkit/vocab";

// A published nginx site. The files live with the Go example, so the build
// context is that directory. Ctrl+C takes the project down.

const context = resolve(dirname(fileURLToPath(import.meta.url)), "../../../examples/nginx");

const dockerfile = (
	<Dockerfile>
		<From image="nginx" tag="alpine" />
		<Copy src="nginx.conf" dest="/etc/nginx/nginx.conf" />
		<Copy src="index.html" dest="/usr/share/nginx/html/index.html" />
		<Expose port="80" />
		<Command executable="nginx" args={["-g", "daemon off;"]} />
	</Dockerfile>
);

function App() {
	useShutdown({ removeOrphans: true });
	return (
		<Compose name="nginx-example">
			<Service name="nginx" image="nginx-example:latest">
				<Build source={context}>{dockerfile}</Build>
				<Publish protocol={proto.TCP} hostIP="127.0.0.1" hostPort={8080} containerPort={80} />
				<Health
					test={["CMD-SHELL", "wget -q -O /dev/null http://127.0.0.1/ || exit 1"]}
					interval="5s"
					timeout="3s"
					retries={5}
				/>
			</Service>
		</Compose>
	);
}

export default root(
	{
		fromEnv: true,
		onExit: "down",
		wait: true,
		onUp: () => console.log("nginx at http://127.0.0.1:8080"),
	},
	App,
);
