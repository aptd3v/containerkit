import type { Compose, Container, Job, Service } from "./generated.js";

/**
 * A Compose, Service, Job, or Container tree.
 *
 * @remarks
 * This is the value JSX produces and the value `root` applies.
 * A static IP belongs on a `Network` child. `ipv4` and `ipv6` are not props of this node.
 * `macAddress` on a Service, Job, or Container is that container's MAC.
 * `macAddress` on a `Network` child is the MAC of that one endpoint.
 * The `Network` under `Compose` sets `subnet` and `gateway`. The `Network` inside a service joins it.
 *
 * @template T - `Compose`, `Service`, `Job`, or `Container`. Omitted, the type is any of those four.
 *
 * @example
 * ```tsx
 * const tree: DockerNode = (
 *   <Compose name="app">
 *     <Network name="backend" driver="bridge" subnet="10.0.0.0/24" gateway="10.0.0.1" />
 *     <Service name="web" image="nginx:alpine" hostname="web">
 *       <Network name="backend" ipv4="10.0.0.10" macAddress="02:42:0a:00:00:0a" />
 *     </Service>
 *   </Compose>
 * );
 * ```
 */
export type DockerNode<T extends Compose | Service | Job | Container = Compose | Service | Job | Container> = T;
