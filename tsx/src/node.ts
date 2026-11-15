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
	return {
		$$kind: kind,
		props: rest as Omit<P, "children">,
		children: flatten(children),
	};
}

export function isNode(v: unknown): v is Node {
	return typeof v === "object" && v !== null && "$$kind" in v;
}
