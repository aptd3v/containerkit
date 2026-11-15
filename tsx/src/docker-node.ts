import type { Compose, Container, Job, Service } from "./generated.js";

// DockerNode<Compose<{ name?: string }>> is a compose tree whose props are that
// object. Service, Job, and Container are the other roots. With no type
// argument, DockerNode is any of those four.
export type DockerNode<T extends Compose | Service | Job | Container = Compose | Service | Job | Container> = T;
