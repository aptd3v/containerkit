import { Compose, Service } from "containerkit";
import { root, runIfMain } from "containerkit/live";

// One Alpine container that prints hello and exits.
// The Go program is examples/simple.

const tree = (
	<Compose name="simple">
		<Service name="hello" image="alpine:latest" command={["echo", "hello"]} />
	</Compose>
);

export { tree };

const live = root({ fromEnv: true }, async function App(session) {
	await session.composeUp(tree);
	return tree;
});

export default live;

await runIfMain(live, import.meta.url);
