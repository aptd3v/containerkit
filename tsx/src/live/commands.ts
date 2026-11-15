import * as generated from "../generated-actions.js";
import { buildDockerfile } from "../index.js";
import type { Client, PlanDocument } from "../plan.js";
import { Flow } from "../taskflow.js";

type Command = (cli: Client, ...args: never[]) => PlanDocument;

export const commandByName: Record<string, Command> = {};

for (const [name, value] of Object.entries(generated)) {
	if (typeof value === "function") {
		commandByName[name] = value as Command;
	}
}
commandByName.buildDockerfile = buildDockerfile;
commandByName.flow = Flow as Command;

type DropClient<F> = F extends (cli: Client, ...args: infer A) => PlanDocument
	? (...args: A) => Promise<unknown>
	: never;

export type SessionCommands = {
	[K in keyof typeof generated]: DropClient<(typeof generated)[K]>;
} & {
	buildDockerfile: DropClient<typeof buildDockerfile>;
	flow: DropClient<typeof Flow>;
};
