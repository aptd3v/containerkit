import { Build, Compose, Service } from "containerkit";
import { Command, Dockerfile, From, Label, Run, Workdir, exportDockerfile } from "containerkit/dockerfile";
import { isCKNotFoundError, root, runIfMain } from "containerkit/live";

// Build hello-inline:v1 when the tag is missing, then bring the service up.
// A later run finds the tag and skips the build.
// The Go program is examples/image_inline.

const taggedImage = "hello-inline:v1";

const dockerfile = (
	<Dockerfile>
		<From image="alpine:latest" />
		<Label name="example" value="image-inline" />
		<Workdir path="/app" />
		<Run cmd="echo 'Hello, World!' > hello.txt" />
		<Command executable="sleep" args={["5s"]} />
	</Dockerfile>
);

export const tree = (
	<Compose name="image-inline">
		<Service name="hello-image-inline" image={taggedImage}>
			<Build>{dockerfile}</Build>
		</Service>
	</Compose>
);

console.log(exportDockerfile(dockerfile));

const live = root({ fromEnv: true, onExit: "down" }, async function App(session) {
	let exists = true;
	try {
		await session.inspectImage(taggedImage);
	} catch (err) {
		if (!isCKNotFoundError(err)) {
			throw err;
		}
		exists = false;
	}
	if (!exists) {
		await session.composeBuild(tree);
	}
	await session.composeUp(tree);
	return tree;
});

export default live;

await runIfMain(live, import.meta.url);
