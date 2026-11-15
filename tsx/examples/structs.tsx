import { Compose, DriverOpt, Env, Expose, Health, Logging, Network, Publish, Service } from "containerkit";
import { exportYAML } from "containerkit/live";
import { logDriver, proto } from "containerkit/vocab";

// The Alpine service from examples/structs, as a Compose tree.
// exportYAML is the TSX counterpart of Compose.ExportYAML. Nothing talks to the daemon.

const arch = process.arch === "x64" ? "amd64" : process.arch;

export const tree = (
	<Compose name="my-project">
		<Network name="my-network" driver="bridge" />
		<Service name="my-alpine-service" image="alpine" command={["tail", "-f", "/dev/null"]} platform={`linux/${arch}`}>
			<Env name="ENV1" value="value1" />
			<Env name="ENV2" value="value2" />
			<Publish protocol={proto.TCP} hostIP="0.0.0.0" hostPort={8080} containerPort={8080} />
			<Expose port={8080} proto={proto.TCP} />
			<Health test={["CMD", "curl", "-f", "http://localhost:8080"]} interval="10s" timeout="5s" retries={3} />
			<Logging driver={logDriver.JSONFile}>
				<DriverOpt name="max-file" value="3" />
				<DriverOpt name="max-size" value="10m" />
			</Logging>
			<Network name="my-network" alias="my-alpine-service" />
		</Service>
	</Compose>
);

console.log(exportYAML(tree));
