import { buildImage } from "./generated-actions.js";
import type { BuildImageOptions } from "./generated-options.js";
import { exportDockerfile, type DockerfileNode } from "./docker-file.js";
import type { Client, PlanDocument } from "./plan.js";

export { BuildArg } from "./env.js";
export { exportYAML, render } from "./render.js";
export { Dockerfile, exportDockerfile } from "./docker-file.js";
export type { DockerfileNode } from "./docker-file.js";
export { jsx, jsxs, jsxDEV, Fragment } from "./jsx-runtime.js";
export { flatten, element, isNode } from "./node.js";
export type { Child, MaybeChild, Node } from "./node.js";
export * from "./generated.js";
export * from "./generated-options.js";
export * from "./generated-actions.js";
export { connect, encode, run, setInvoke } from "./plan.js";
export { binaryInvoke, resolveBinary } from "./runner.js";
export type { BinaryInvokeOptions } from "./runner.js";
export type { Client, Invoke, PlanAction, PlanDocument, RunOptions, RunProgress } from "./plan.js";
export type { DockerNode } from "./docker-node.js";

/**
 * Build an image from a Dockerfile tree.
 *
 * @param cli - Connected client. `connect({ fromEnv: true })` reads the local Docker host.
 * @param file - `<Dockerfile>` tree from `containerkit/dockerfile`
 * @param options - Image build settings. `parseDockerfile` is filled from `file`.
 * @returns A plan. Pass it to `run`, or to `session.call` from a live session.
 *
 * @example
 * ```tsx
 * const plan = buildDockerfile(connect({ fromEnv: true }), (
 *   <Dockerfile>
 *     <From image="alpine" tag="latest" />
 *   </Dockerfile>
 * ), { image: "app:dev" });
 * ```
 *
 * @see {@link exportDockerfile}
 */
export function buildDockerfile(
	cli: Client,
	file: DockerfileNode,
	options?: Omit<BuildImageOptions, "parseDockerfile">,
): PlanDocument {
	return buildImage(cli, { ...options, parseDockerfile: exportDockerfile(file) });
}
