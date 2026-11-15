// A scalar component preserves the string it receives from JSX.
// A multiline literal may already have been normalized by the compiler.

import { element } from "./node.js";

export type EnvInput =
	| { name?: string; value: string; children?: never }
	| { name?: string; value?: never; children: string };

const oneString = (tag: string) => `containerkit/tsx: <${tag}> accepts a single string`;

type Raw = Record<string, unknown>;

// resolveEnv folds a child string into value. The string is the one JSX passed in.
export function resolveEnv(props: EnvInput | null | undefined): { name?: string; value: string } {
	return resolveNamedValue(props, "env");
}

export function resolveNamedValue(
	props: EnvInput | null | undefined,
	tag: string,
): { name?: string; value: string } {
	return resolveScalar(props, "value", { tag, required: true }) as { name?: string; value: string };
}

// resolveScalar writes one string child into key. An empty string is present.
// required false leaves the key off the result when both sources are absent.
export function resolveScalar(
	props: object | null | undefined,
	key: string,
	options: { tag: string; required: boolean },
): Record<string, unknown> {
	const raw = (props ?? {}) as Raw;
	const current = raw[key];
	const child = takeString(raw.children, options.tag);
	if (current !== undefined && typeof current !== "string") {
		throw new Error(oneString(options.tag));
	}
	const hasValue = typeof current === "string";
	const hasChild = child !== undefined;
	if (hasValue && hasChild) {
		throw new Error(`containerkit/tsx: <${options.tag}> accepts ${key} or children, not both`);
	}
	const rest: Record<string, unknown> = {};
	for (const [name, value] of Object.entries(raw)) {
		if (name === "children" || name === key) {
			continue;
		}
		rest[name] = value;
	}
	if (hasValue) {
		return { ...rest, [key]: current };
	}
	if (hasChild) {
		return { ...rest, [key]: child };
	}
	if (options.required) {
		throw new Error(missing(options.tag, key));
	}
	return rest;
}

// BuildArg is the build-arg spelling of an env node. Build renders that node as an arg.
export function BuildArg(props: EnvInput | null | undefined) {
	return element("env", resolveEnv(props));
}

function missing(tag: string, key: string): string {
	if (key === "value") {
		return `containerkit/tsx: <${tag}> requires a value. An empty string is value=""`;
	}
	return `containerkit/tsx: <${tag}> requires a ${key}`;
}

// takeString returns one string, undefined when the child is absent, or throws.
function takeString(child: unknown, tag: string): string | undefined {
	const parts: unknown[] = [];
	const walk = (value: unknown): void => {
		if (value == null || value === false || value === true) {
			return;
		}
		if (Array.isArray(value)) {
			for (const item of value) {
				walk(item);
			}
			return;
		}
		parts.push(value);
	};
	walk(child);
	if (parts.length === 0) {
		return undefined;
	}
	if (parts.length === 1 && typeof parts[0] === "string") {
		return parts[0];
	}
	throw new Error(oneString(tag));
}
