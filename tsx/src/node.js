const REACT_ELEMENT = Symbol.for("react.element");
export const reactCompiledJSX = "containerkit/tsx: this JSX was compiled with React. Keep React's jsxImportSource on UI files. A containerkit file needs `/** @jsxImportSource containerkit */`, or a tsconfig that extends `containerkit/tsconfig.jsx.json`. One file cannot mix <div> and <Compose>.";
export function isReactElement(v) {
    return typeof v === "object" && v !== null && v.$$typeof === REACT_ELEMENT;
}
export function flatten(child) {
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
        const n = child;
        if (n.$$kind === "fragment") {
            return n.children;
        }
        return [n];
    }
    throw new Error(`containerkit/tsx: invalid child ${typeof child}`);
}
export function element(kind, props) {
    const p = props ?? {};
    const { children, ...rest } = p;
    return {
        $$kind: kind,
        props: rest,
        children: flatten(children),
    };
}
export function isNode(v) {
    return typeof v === "object" && v !== null && "$$kind" in v;
}
