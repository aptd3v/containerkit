// Live entry for the Engine job. The file name is docker.js so root()
// starts itself and onExit: "down" exits this process on SIGTERM.
// scripts/ci/engine.sh sets CONTAINERKIT_BIN to the absolute dist/containerkit
// from tools/crossbuild, waits for ck-live-up, then signals this process.

import { Compose, Service } from "../dist/index.js";
import { root } from "../dist/live/index.js";

export default root(
	{
		fromEnv: true,
		onExit: "down",
		onUp() {
			console.log("ck-live-up");
		},
	},
	function App() {
		return Compose({
			name: "ck-live",
			children: Service({
				name: "hold",
				image: "alpine:3.22",
				command: ["sleep", "infinity"],
			}),
		});
	},
);
