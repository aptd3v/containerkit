import { isCKNotFoundError, root, runIfMain } from "containerkit/live";

// Create an Alpine container and start an exec shell in it.
// The Go program is examples/terminal. That program copies the local terminal.
// startExec runs the exec the session already created.

const name = "exec-example";
const image = "alpine:latest";

const live = root({ fromEnv: true }, async function App(session) {
	try {
		await session.inspectImage(image);
	} catch (err) {
		if (!isCKNotFoundError(err)) {
			throw err;
		}
		await session.pullImage(image, { platform: `linux/${process.arch === "x64" ? "amd64" : process.arch}` });
	}

	await session.run({
		name,
		image,
		command: ["tail", "-f", "/dev/null"],
	});

	try {
		const exec = await session.createExec(name, {
			command: ["/bin/sh"],
			tty: true,
			attachStdin: true,
			attachStdout: true,
			attachStderr: true,
			consoleSize: { height: 24, width: 80 },
		});
		await session.startExec(exec.ID, { tty: true });
	} finally {
		await session.stop(name);
		await session.remove(name, { force: true });
	}
	return null;
});

export default live;

await runIfMain(live, import.meta.url);
