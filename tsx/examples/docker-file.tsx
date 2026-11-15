import {
	Comment,
	Copy,
	Dockerfile,
	Entrypoint,
	Expose,
	From,
	HealthCheck,
	Label,
	Run,
	Syntax,
	User,
	Workdir,
	exportDockerfile,
} from "containerkit/dockerfile";
import { root, runIfMain } from "containerkit/live";
import { proto } from "containerkit/vocab";

// Multi-stage Dockerfile, built and served on port 8088.
// The Go program is examples/dockerfile.

const imageRef = "dockerfile-example:latest";

const dockerfile = (
	<Dockerfile>
		<Syntax value="docker/dockerfile:1" />
		<Comment text="stage a static page, then serve it with busybox httpd" />
		<From image="alpine:latest" as="build" />
		<Workdir path="/out" />
		<Run cmd="printf 'Hello from containerkit\n' > index.html" />
		<Run cmd="chmod 644 index.html" />
		<From image="alpine" tag="latest" />
		<Run cmd="apk add --no-cache busybox-extras" />
		<Copy src="/out/index.html" dest="/www/index.html" from="build" />
		<Label name="example" value="dockerfile" />
		<Expose port="8080" />
		<User name="nobody" />
		<HealthCheck
			test={["CMD-SHELL", "wget -qO- http://127.0.0.1:8080/ >/dev/null"]}
			interval="5s"
			timeout="2s"
			retries={3}
		/>
		<Entrypoint executable="/usr/sbin/httpd" args={["-f", "-p", "8080", "-h", "/www"]} />
	</Dockerfile>
);

const text = exportDockerfile(dockerfile);
console.log(text);

export { dockerfile, text };

const live = root({ fromEnv: true }, async function App(session) {
	await session.buildImage({
		image: imageRef,
		parseDockerfile: text,
		force: true,
	});
	await session.run({
		name: "dockerfile-example",
		image: imageRef,
		publish: {
			protocol: proto.TCP,
			hostIP: "0.0.0.0",
			hostPort: 8088,
			containerPort: 8080,
		},
	});
	return null;
});

export default live;

await runIfMain(live, import.meta.url);
