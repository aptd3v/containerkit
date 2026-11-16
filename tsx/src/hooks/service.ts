import { useEffect, useSession, useState } from "../live/root.js";
import type { LiveEvent } from "../live/session.js";

/**
 * Latest status event for one service.
 *
 * @remarks
 * Call it from a render callback. A matching event stores the event and renders again.
 * A status event matches when `source` or `parent` is `name`, or `source` ends with ` name`,
 * so `Container web` matches `web` and `web-proxy` does not.
 * Error events are not included. Listen for those with `session.on("error")`.
 *
 * @param name - Service name, such as `web`
 * @returns The latest matching status event, or `undefined` until one arrives
 * @throws Error when called outside a render callback
 *
 * @example
 * ```tsx
 * function App() {
 *   const web = useServiceState("web");
 *   console.log(web?.line);
 *   return <Service name="web" image="nginx:alpine" />;
 * }
 * ```
 *
 * @see {@link useLiveLog}
 */
export function useServiceState(name: string): LiveEvent | undefined {
	const session = useSession();
	const [event, setEvent] = useState<LiveEvent | undefined>(undefined);
	useEffect(() => {
		return session.on("status", (next) => {
			if (forService(next, name)) {
				setEvent(next);
			}
		});
	}, [session, name]);
	return event;
}

function forService(event: LiveEvent, name: string): boolean {
	if (event.source === name || event.parent === name) {
		return true;
	}
	return event.source?.endsWith(` ${name}`) === true;
}
