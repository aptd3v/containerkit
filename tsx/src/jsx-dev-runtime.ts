export { jsx, jsx as jsxs, jsx as jsxDEV, Fragment } from "./jsx-runtime.js";
export type { Node, Child } from "./jsx-runtime.js";

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
