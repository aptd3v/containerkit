import type { PlanOp } from "./generated-actions.js";
import type { ConnectOptions } from "./generated-options.js";
import type { FlowStep } from "./taskflow.js";

// Client is a session token. Client setters live on connect(), never on DockerNode.
export type Client = {
	readonly $$kind: "client";
	readonly options: ConnectOptions;
};

export type PlanAction = {
	op: PlanOp;
	options?: object;
	ref?: string;
	service?: string;
	id?: string;
	path?: string;
	[key: string]: unknown;
};

export type PlanDocument = {
	connect?: ConnectOptions;
	yaml?: string;
	actions?: PlanAction[];
	flow?: {
		name: string;
		steps: FlowStep[];
	};
};

export function connect(options?: ConnectOptions): Client {
	return { $$kind: "client", options: options ?? {} };
}

export function encode(doc: PlanDocument): string {
	return JSON.stringify(doc, null, 2) + "\n";
}

export type Invoke = (op: string, payload: unknown) => Promise<unknown>;

let invokeFn: Invoke | undefined;

export function setInvoke(fn: Invoke | undefined): void {
	invokeFn = fn;
}

export async function run(doc: PlanDocument): Promise<unknown> {
	if (invokeFn) {
		return invokeFn("exec", doc);
	}
	const { runWithBinary } = await import("./runner.js");
	return runWithBinary(doc);
}
