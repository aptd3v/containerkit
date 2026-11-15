/** @jsxImportSource containerkit */
import { Compose, Service, type DockerNode } from "containerkit";

// This file sits in a project whose jsxImportSource is react.
// The pragma keeps <Compose> on containerkit even though @types/react is installed.
export function stack(): DockerNode {
	return (
		<Compose name="app">
			<Service name="api" image="alpine:latest" />
		</Compose>
	);
}

export function htmlIsRejected(): DockerNode {
	// @ts-expect-error HTML tags belong to React, not containerkit
	return <div />;
}
