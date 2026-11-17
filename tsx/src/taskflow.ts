import type { ComposeStepSpec } from "./generated-actions.js";
import type { BuildImageOptions, ContainerOptions, EngineImageTagOptions, UpOptions } from "./generated-options.js";
import type { Node } from "./node.js";
import type { Client, PlanDocument } from "./plan.js";
import { render } from "./render.js";

export type Priority = "high" | "normal" | "low";

// FlowStep is one node in a flow document. Edges are step names.
export type FlowStep = {
	name: string;
	op: string;
	yaml?: string;
	options?: object;
	dependsOn?: string[];
	precede?: string[];
	then?: FlowStep[];
	else?: FlowStep[];
	when?: boolean;
	retry?: number;
	skipIfPresent?: boolean;
	always?: boolean;
	priority?: Priority;
	steps?: FlowStep[];
	service?: string;
	id?: string;
	path?: string;
	ref?: string;
	repository?: string;
	port?: number;
};

type Edges = {
	name: string;
	dependsOn?: string[];
	precede?: string[];
	retry?: number;
	always?: boolean;
	priority?: Priority;
};

function withEdges(step: FlowStep, spec: Edges): FlowStep {
	if (spec.dependsOn) {
		step.dependsOn = spec.dependsOn;
	}
	if (spec.precede) {
		step.precede = spec.precede;
	}
	if (spec.retry !== undefined) {
		step.retry = spec.retry;
	}
	if (spec.always) {
		step.always = spec.always;
	}
	if (spec.priority) {
		step.priority = spec.priority;
	}
	return step;
}

// stepsOf flattens a JSX-style child into the step list Flow stores.
export function stepsOf(children: FlowStep | FlowStep[] | undefined): FlowStep[] {
	if (!children) {
		return [];
	}
	if (Array.isArray(children)) {
		return children;
	}
	return [children];
}

// Flow encodes a task graph on a plan document. Go runs it when the flow field is set.
export function Flow(
	cli: Client,
	spec: { name?: string; volumes?: string[]; steps?: FlowStep[]; children?: FlowStep | FlowStep[] },
): PlanDocument {
	const steps = spec.steps ?? stepsOf(spec.children);
	const flow: NonNullable<PlanDocument["flow"]> = {
		name: spec.name ?? "taskflow",
		steps,
	};
	if (spec.volumes) {
		flow.volumes = spec.volumes;
	}
	return {
		connect: cli.options,
		flow,
	};
}

// volumeSource is the mount source Import rewrites to a run-private volume name.
export function volumeSource(logical: string): string {
	return `volume:${logical}`;
}

// Service brings a Compose project up for the life of the declaring flow or subflow.
export function Service(spec: Edges & { node: Node } & UpOptions): FlowStep {
	const { name, dependsOn, precede, retry, always, priority, node, ...options } = spec;
	return withEdges(
		{ name, op: "service", yaml: render(node), options },
		{ name, dependsOn, precede, retry, always, priority },
	);
}

// Build is an image build step. Options are the buildImage bag.
// skipIfPresent asks Go to inspect tags and skip the build when every tag exists.
export function Build(spec: Edges & BuildImageOptions & { skipIfPresent?: boolean }): FlowStep {
	const { name, dependsOn, precede, retry, always, priority, skipIfPresent, ...options } = spec;
	const step = withEdges({ name, op: "buildImage", options }, { name, dependsOn, precede, retry, always, priority });
	if (skipIfPresent) {
		step.skipIfPresent = true;
	}
	return step;
}

// Container is a one-shot container step. Options are the container bag.
export function Container(spec: Edges & ContainerOptions): FlowStep {
	const { name, dependsOn, precede, retry, always, priority, ...options } = spec;
	return withEdges({ name, op: "containerRun", options }, { name, dependsOn, precede, retry, always, priority });
}

// ComposeStep renders a Compose node to YAML and stores a compose op.
// The op selects the options type and the required fields.
// An inline options object is checked against that type.
// A predeclared options value is checked structurally, so extra properties can remain when its type overlaps.
export function ComposeStep(spec: Edges & { node: Node } & ComposeStepSpec): FlowStep {
	const step = withEdges(
		{ name: spec.name, op: spec.op, yaml: render(spec.node), options: spec.options },
		spec,
	);
	if ("service" in spec && spec.service) {
		step.service = spec.service;
	}
	if ("port" in spec && spec.port) {
		step.port = spec.port;
	}
	if ("repository" in spec && spec.repository) {
		step.repository = spec.repository;
	}
	return step;
}

// If is a boolean chosen while the module evaluates. then and else are step arrays.
export function If(spec: Edges & { when: boolean; then?: FlowStep[]; else?: FlowStep[] }): FlowStep {
	return withEdges({ name: spec.name, op: "if", when: spec.when, then: spec.then, else: spec.else }, spec);
}

// Subflow nests steps. They are baked in before run.
export function Subflow(spec: Edges & { steps: FlowStep[] }): FlowStep {
	return withEdges({ name: spec.name, op: "subflow", steps: spec.steps }, spec);
}

// Tag retags an image. image sets both source and target when those are omitted.
export function Tag(spec: Edges & EngineImageTagOptions & { image?: string }): FlowStep {
	const source = spec.source ?? spec.image;
	const target = spec.target ?? spec.image;
	return withEdges({ name: spec.name, op: "tagImage", options: { source, target } }, spec);
}

// Always marks a step to run after a failed predecessor.
export function Always(step: FlowStep): FlowStep {
	return { ...step, always: true };
}

// Retry reruns a failing step n more times.
export function Retry(n: number, step: FlowStep): FlowStep {
	return { ...step, retry: n };
}
