import { useState } from "../live/root.js";

/**
 * Read one process environment variable into state.
 *
 * @remarks
 * The value is read once, on the first render. A later change to `process.env` is not seen.
 * An unset variable uses `fallback`. A variable set to an empty string stays empty.
 * Call it from a render callback.
 *
 * @param name - Environment variable name, such as `IMAGE`
 * @param fallback - Value used when the variable is unset. Defaults to `""`
 * @returns The variable's value, or `fallback`
 * @throws Error when called outside a render callback
 *
 * @example
 * ```tsx
 * function App() {
 *   const image = useEnv("IMAGE", "nginx:alpine");
 *   return <Service name="web" image={image} />;
 * }
 * ```
 */
export function useEnv(name: string, fallback = ""): string {
	const [value] = useState(() => process.env[name] ?? fallback);
	return value;
}
