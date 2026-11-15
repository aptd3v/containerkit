import { element } from "./node.js";
export function jsx(type, props, _key) {
    if (typeof type === "function") {
        return type(props ?? {});
    }
    return element(String(type), props);
}
export const jsxs = jsx;
export const jsxDEV = jsx;
export function Fragment(props) {
    return element("fragment", props);
}
