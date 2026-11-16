import { allowedChildren, exclusiveChildProps, singleChildren } from "./generated.js";

export type MaybeChild<T> = T | false | null | undefined | boolean | readonly MaybeChild<T>[];

export type Child = MaybeChild<Node>;

export type Node<K extends string = string, P = Record<string, unknown>> = {
	$$kind: K;
	props: P;
	children: Node[];
};

const REACT_ELEMENT = Symbol.for("react.element");

export const reactCompiledJSX =
	"containerkit/tsx: this JSX was compiled with React. Keep React's jsxImportSource on UI files. A containerkit file needs `/** @jsxImportSource containerkit */`, or a tsconfig that extends `containerkit/tsconfig.jsx.json`. One file cannot mix <div> and <Compose>.";

export function isReactElement(v: unknown): boolean {
	return typeof v === "object" && v !== null && (v as { $$typeof?: unknown }).$$typeof === REACT_ELEMENT;
}

export function flatten(child: unknown): Node[] {
	if (child == null || child === false || child === true) {
		return [];
	}
	if (Array.isArray(child)) {
		return child.flatMap(flatten);
	}
	if (isReactElement(child)) {
		throw new Error(reactCompiledJSX);
	}
	if (typeof child === "object" && child !== null && "$$kind" in child) {
		const n = child as Node;
		if (n.$$kind === "fragment") {
			return n.children;
		}
		return [n];
	}
	throw new Error(`containerkit/tsx: invalid child ${typeof child}`);
}

export function element<K extends string, P extends Record<string, unknown>>(
	kind: K,
	props: (P & { children?: Child }) | null,
): Node<K, Omit<P, "children">> {
	const p = props ?? ({} as P & { children?: Child });
	const { children, ...rest } = p;
	const flat = flatten(children);
	checkChildren(kind, flat, rest, false);
	return {
		$$kind: kind,
		props: rest as Omit<P, "children">,
		children: flat,
	};
}

// checkChildren enforces the generated placement table.
// A missing kind is skipped here and rejected by render, which passes missingIsError.
// An empty list means the kind accepts no node children.
// singleChildren rejects a second child from the same group.
export function checkChildren(
	kind: string,
	children: readonly Node[],
	props: object | null | undefined,
	missingIsError: boolean,
): void {
	const allowed = allowedChildren[kind];
	if (!allowed) {
		if (!missingIsError) {
			return;
		}
		for (const child of children) {
			throw new Error(`containerkit/tsx: <${child.$$kind}> is not valid inside <${kind}>`);
		}
		return;
	}
	const groups = singleChildren[kind] ?? [];
	const seen = new Set<number>();
	for (const child of children) {
		if (!allowed.includes(child.$$kind)) {
			throw new Error(`containerkit/tsx: <${child.$$kind}> is not valid inside <${kind}>`);
		}
		for (let i = 0; i < groups.length; i++) {
			const group = groups[i];
			if (!group.kinds.includes(child.$$kind)) {
				continue;
			}
			if (seen.has(i)) {
				throw new Error(`containerkit/tsx: ${group.message}`);
			}
			seen.add(i);
			break;
		}
	}
	const exclusive = exclusiveChildProps[kind];
	if (!exclusive) {
		return;
	}
	const raw = (props ?? {}) as Record<string, unknown>;
	const hasChild = children.some((child) => child.$$kind === exclusive.child);
	const hasProp = exclusive.props.some((name) => raw[name] !== undefined && raw[name] !== "");
	if (hasChild && hasProp) {
		throw new Error(`containerkit/tsx: ${exclusive.message}`);
	}
}

export function isNode(v: unknown): v is Node {
	return typeof v === "object" && v !== null && "$$kind" in v;
}
