import { Compose, DependsOn, Service, type ComposeNode } from "containerkit";
import { root, runIfMain } from "containerkit/live";
import { signal } from "containerkit/vocab";

// Profiles hide services until that profile is selected. message-chain echoes
// one line per service. kill-me stays up until SIGKILL.
// The Go program is examples/profiles. composeUp has no profile flag, so each
// call uses the tree that already contains only that profile.

const paragraph = `containerkit
is a Go library that provides a programmatic and composable interface
for defining, running, and managing Docker containers and Compose projects.
It abstracts both the Docker SDK and Docker Compose into a unified API.
In this example, we'll use profiles and demonstrate how to start a service
with a specific profile.
`;

const selectedProfile = "message-chain";
const lines = paragraph.split("\n");

export const project = (
	<Compose name="my-project">
		<Service name="never-service" image="alpine:latest" command={["echo", "you wont see me"]} profile="never" />
		<Service
			name="never-ever-service"
			image="alpine:latest"
			command={["echo", "you wont see me part II"]}
			profile="never-ever"
		/>
		<Service name="kill-me-service" image="alpine:latest" command={["tail", "-f", "/dev/null"]} profile="kill-me" />
		{lines.map((line, i) => (
			<Service
				key={i}
				name={`service${i}`}
				image="alpine:latest"
				command={["echo", line]}
				profile={selectedProfile}
			>
				{i > 0 && <DependsOn name={`service${i - 1}`} />}
			</Service>
		))}
	</Compose>
);

function chain(): ComposeNode {
	return (
		<Compose name="my-project">
			{lines.map((line, i) => (
				<Service key={i} name={`service${i}`} image="alpine:latest" command={["echo", line]}>
					{i > 0 && <DependsOn name={`service${i - 1}`} />}
				</Service>
			))}
		</Compose>
	);
}

const killMe = (
	<Compose name="my-project">
		<Service name="kill-me-service" image="alpine:latest" command={["tail", "-f", "/dev/null"]} />
	</Compose>
);

const live = root({ fromEnv: true }, async function App(session) {
	await session.composeUp(chain(), { removeOrphans: true });
	await session.composeUp(killMe, { detach: true, removeOrphans: true });
	await session.composeKill(killMe, { stopSignal: signal.SIGKILL, removeOrphans: true });
	return project;
});

export default live;

await runIfMain(live, import.meta.url);
