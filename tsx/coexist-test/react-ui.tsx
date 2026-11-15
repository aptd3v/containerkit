import { Compose } from "containerkit";

// Project jsxImportSource is react. <div> stays a React element.
export function Page() {
	return <div className="app">hello</div>;
}

export function composeIsNotUi() {
	// @ts-expect-error Compose returns a containerkit node, not a React element
	return <Compose name="app" />;
}
