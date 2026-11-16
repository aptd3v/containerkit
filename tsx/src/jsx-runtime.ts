import { element, type Child, type Node } from "./node.js";

export type { Node, Child };

type Component<P = Record<string, unknown>, R extends Node = Node> = (props: P) => R;

export function jsx<P, R extends Node>(
	type: Component<P, R>,
	props: P | null,
	_key?: string,
): R;
export function jsx(
	type: string,
	props: Record<string, unknown> | null,
	_key?: string,
): Node;
export function jsx(
	type: string | Component,
	props: Record<string, unknown> | null,
	_key?: string,
): Node {
	if (typeof type === "function") {
		return type(props ?? {});
	}
	return element(String(type), props);
}

export const jsxs = jsx;
export const jsxDEV = jsx;

export function Fragment(props: { children?: Child }): Node<"fragment"> {
	return element("fragment", props);
}

// This namespace is what the language server consults for jsxImportSource.
// If it is missing, or the import source does not resolve, TypeScript looks up
// a global namespace named React (jsx: react-jsx defaults the specifier to "react").
// Element is left unset on purpose. TypeScript types every tag as that one type,
// and a missing Element is any. One Element cannot be both a Compose result and a
// Service child, so the precise node type stays on the component function.
// Intrinsic elements are closed, so <div> is not valid.
export namespace JSX {
	export interface ElementChildrenAttribute {
		children: {};
	}
	export interface IntrinsicAttributes {
		key?: string | number;
	}
	export interface IntrinsicElements {
		[elemName: string]: never;
	}
}
