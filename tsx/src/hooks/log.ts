import { useEffect, useSession } from "../live/root.js";
import type { SessionLogEvent } from "../live/session.js";

/**
 * Subscribe to session log and status events after the render settles.
 *
 * @remarks
 * The handler does not receive error events. Subscribe to those with `session.on("error")`.
 * The subscription is removed before the next run and when the session closes.
 * Call it from a render callback. An async once callback does not run effects.
 *
 * @param onEvent - Called with each log or status event
 * @throws Error when called outside a render callback
 *
 * @example
 * ```tsx
 * function App() {
 *   useLiveLog((event) => {
 *     console.log(event.line);
 *   });
 *   return <Service name="web" image="nginx:alpine" />;
 * }
 * ```
 *
 * @see {@link useServiceState}
 */
export function useLiveLog(onEvent: (event: SessionLogEvent) => void): void {
	const session = useSession();
	useEffect(() => {
		const offLog = session.on("log", onEvent);
		const offStatus = session.on("status", onEvent);
		return () => {
			offLog();
			offStatus();
		};
	}, [session, onEvent]);
}
