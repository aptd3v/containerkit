import { Build, Compose, Develop, Health, Publish, Service, WatchFiles } from "containerkit";
import { Command, Copy, Dockerfile, Expose, From, exportDockerfile } from "containerkit/dockerfile";
import { root, runIfMain } from "containerkit/live";
import { proto, watchAction } from "containerkit/vocab";

// nginx-example:latest from examples/nginx, published on port 8080.
// The Go program is examples/nginx. BuildImage there packs ./examples/nginx
// as the context; composeBuild here uses the Dockerfile stored on the service.

const dockerfile = (
	<Dockerfile>
		<From image="nginx" tag="latest" />
		<Copy src="nginx.conf" dest="/etc/nginx/nginx.conf" />
		<Copy src="index.html" dest="/usr/share/nginx/html/index.html" />
		<Expose port="80" />
		<Command executable="nginx" args={["-g", "daemon off;"]} />
	</Dockerfile>
);

const tree = (
	<Compose name="nginx-example">
		<Service
			name="nginx"
			image="nginx-example:latest"
			command={["nginx", "-g", "daemon off;"]}
			memory={100 << 20}
		>
			<Build source="./examples/nginx">{dockerfile}</Build>
			<Publish protocol={proto.TCP} hostIP="0.0.0.0" hostPort={8080} containerPort={80} />
			<Health
				test={["CMD-SHELL", "curl -f http://localhost:80 || exit 1"]}
				interval="10s"
				timeout="5s"
				retries={3}
			/>
			<Develop>
				<WatchFiles action={watchAction.Sync} path="./src" target="/app" ignore="tmp" />
			</Develop>
		</Service>
	</Compose>
);

export { tree };

const live = root({ fromEnv: true }, async function App(session) {
	await session.buildImage({
		image: "nginx-example:latest",
		source: "./examples/nginx",
		parseDockerfile: exportDockerfile(dockerfile),
	});
	await session.composeBuild(tree);
	await session.composeUp(tree, { detach: true, removeOrphans: true });
	console.error("app started at http://localhost:8080");
	return tree;
});

export default live;

await runIfMain(live, import.meta.url);
