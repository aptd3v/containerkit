// The repo typechecks the bare name "containerkit" through tsconfig paths.
// tsc leaves that specifier in the emitted JavaScript. Node resolves the
// package that is actually installed, which is @aptd3v/containerkit.

export function mapSpecifier(specifier) {
	if (specifier === "containerkit" || specifier.startsWith("containerkit/")) {
		return `@aptd3v/containerkit${specifier.slice("containerkit".length)}`;
	}
	return specifier;
}

export async function resolve(specifier, context, nextResolve) {
	return nextResolve(mapSpecifier(specifier), context);
}
