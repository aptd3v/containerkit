import { isCKNotFoundError, root, runIfMain } from "containerkit/live";

// Create an Alpine container and attach an interactive shell to it.
// The Go program is examples/terminal. session.terminal runs that same
// StartExecTerminal path and blocks until the shell exits.

const name = "exec-example";
const image = "alpine:latest";

const live = root({ fromEnv: true }, async function App(session) {
	try {
		await session.inspectImage(image);
	} catch (err) {
		if (!isCKNotFoundError(err)) {
			throw err;
		}
		console.error("pulling alpine:latest");
		await session.pullImage(image, { platform: `linux/${process.arch === "x64" ? "amd64" : process.arch}` });
	}

	await session.run({
		name,
		image,
		command: ["tail", "-f", "/dev/null"],
	});

	try {
		await session.terminal(name);
		console.log("session closed running cleanup");
	} finally {
		try {
			await session.stop(name);
		} catch (err) {
			if (!isCKNotFoundError(err)) {
				console.error(err);
			}
		}
		try {
			await session.remove(name, { force: true });
		} catch (err) {
			if (!isCKNotFoundError(err)) {
				console.error(err);
			}
		}
		console.log("Cleanup completed");
	}
	return null;
});

export default live;

await runIfMain(live, import.meta.url);
