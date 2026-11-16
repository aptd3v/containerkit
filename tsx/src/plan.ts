import type { PlanOp } from "./generated-actions.js";
import type { ConnectOptions } from "./generated-options.js";
import type { LiveEvent } from "./live/session.js";
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

/** How a one-shot run reports progress. `off` writes none. */
export type RunProgress = "auto" | "plain" | "tty" | "quiet" | "json" | "off";

export type PlanDocument = {
	connect?: ConnectOptions;
	yaml?: string;
	actions?: PlanAction[];
	flow?: {
		name: string;
		steps: FlowStep[];
	};
	/** Progress for `containerkit exec`. Empty is the same as `auto` on that command. */
	progress?: RunProgress;
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

/** Settings for {@link run}. */
export type RunOptions = {
	/**
	 * Progress written by the runner. The default is `auto`: a terminal redraws,
	 * and a pipe prints one line per event. `json` writes one event record per line.
	 * `off` writes none.
	 */
	progress?: RunProgress;
	/** Called for each JSON progress event. Those lines are not also printed. */
	onEvent?: (event: LiveEvent) => void;
};

export async function run(doc: PlanDocument, options?: RunOptions): Promise<unknown> {
	const progress = options?.progress ?? doc.progress ?? "auto";
	const payload: PlanDocument = { ...doc, progress };
	if (invokeFn) {
		return invokeFn("exec", payload);
	}
	const { runWithBinary } = await import("./runner.js");
	return runWithBinary(payload, { progress, onEvent: options?.onEvent });
}
