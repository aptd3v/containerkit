import { Compose, Health, Publish, Service } from "containerkit";
import { root, runIfMain } from "containerkit/live";
import { proto } from "containerkit/vocab";

// Stream compose events while nginx is started and stopped three times.
// The Go program is examples/events.

const serviceName = "nginx";

const tree = (
	<Compose name="events-project">
		<Service name={serviceName} image="nginx:latest">
			<Publish protocol={proto.TCP} hostIP="0.0.0.0" hostPort={8080} containerPort={80} />
			<Health
				test={["CMD-SHELL", "curl -f http://localhost:80 || exit 1"]}
				interval="2s"
				timeout="5s"
				retries={3}
			/>
		</Service>
	</Compose>
);

export { tree };

const live = root({ fromEnv: true }, async function App(session) {
	const events = session.composeEvents(tree, { services: serviceName });
	for (let i = 0; i < 3; i++) {
		console.error(`Cycle ${i + 1}: Starting container...`);
		await session.composeUp(tree, { detach: true });
		console.error(`Cycle ${i + 1}: Stopping container...`);
		await session.composeDown(tree, { removeOrphans: true });
	}
	await events;
	return tree;
});

export default live;

await runIfMain(live, import.meta.url);
