import * as moduleHooks from "node:module";
import { mapSpecifier } from "./containerkit-loader.mjs";

const hooks = {
	resolve(specifier, context, nextResolve) {
		return nextResolve(mapSpecifier(specifier), context);
	},
};

if (typeof moduleHooks.registerHooks === "function") {
	moduleHooks.registerHooks(hooks);
} else {
	moduleHooks.register("./containerkit-loader.mjs", import.meta.url);
}
