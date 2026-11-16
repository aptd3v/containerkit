import type { DockerNode } from "../docker-node.js";
import type { DownOptions } from "../generated-options.js";
import { isNode } from "../node.js";
import { isContainerKitError } from "../live/error.js";
import { forgetShutdown, rememberShutdown, shutdownTarget, useSession } from "../live/root.js";
import type { Session } from "../live/session.js";

const installed = new WeakMap<Session, () => void>();

/**
 * Take the applied project down on SIGINT and SIGTERM, then exit.
 *
 * @remarks
 * Call it from a render callback, or before the first `await` in an async root.
 * The project is the one `root` last applied, so a later render that returns a larger tree is what gets taken down.
 * Pass a project only when this callback applies that tree itself, before `root` has applied a return value.
 * When `onExit` is `"down"`, this stores the options on that shutdown instead of adding a second listener.
 * A second signal while the down is running is ignored.
 * An already-closed session (`unavailable`) is not printed. Any other failure is printed and the process exits with status 1.
 *
 * @param project - Tree to take down when `root` has not applied one yet
 * @param options - Passed through to `session.composeDown`. As the only argument, these are the down options.
 * @returns A function that removes the signal listeners, or clears the options when `onExit` is `"down"`
 * @throws Error when called outside a root callback
 *
 * @example
 * ```tsx
 * export default root({ fromEnv: true, onExit: "down" }, function App() {
 *   useShutdown({ removeOrphans: true, volumes: true });
 *   return (
 *     <Compose name="app">
 *       <Service name="web" image="nginx:alpine" />
 *     </Compose>
 *   );
 * });
 * ```
 *
 * @see {@link useSession}
 */
export function useShutdown(options?: DownOptions): () => void;
export function useShutdown(project: DockerNode, options?: DownOptions): () => void;
export function useShutdown(projectOrOptions?: DockerNode | DownOptions, options?: DownOptions): () => void {
	const session = useSession();
	const project = isNode(projectOrOptions) ? projectOrOptions : undefined;
	const down = isNode(projectOrOptions) ? options : projectOrOptions;
	rememberShutdown(session, project, down);
	if (shutdownTarget(session).onExit === "down") {
		return () => forgetShutdown(session);
	}
	const existing = installed.get(session);
	if (existing) {
		return existing;
	}
	let stopping = false;
	const stop = () => {
		if (stopping) {
			return;
		}
		stopping = true;
		const target = shutdownTarget(session);
		const done = target.project ? session.composeDown(target.project, target.options) : Promise.resolve();
		void done
			.catch((err: unknown) => {
				if (isContainerKitError(err, "unavailable")) {
					return;
				}
				console.error(err instanceof Error ? err.message : err);
				process.exitCode = 1;
			})
			.finally(() => {
				process.exit();
			});
	};
	process.on("SIGINT", stop);
	process.on("SIGTERM", stop);
	const dispose = () => {
		process.off("SIGINT", stop);
		process.off("SIGTERM", stop);
		installed.delete(session);
		forgetShutdown(session);
	};
	installed.set(session, dispose);
	return dispose;
}
