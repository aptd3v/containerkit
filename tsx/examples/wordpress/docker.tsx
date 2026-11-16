/** @jsxImportSource containerkit */
import { Compose, DependsOn, Env, Health, Network, Publish, Service, Volume } from "containerkit";
import { useShutdown } from "containerkit/hooks";
import { root } from "containerkit/live";
import { depend, netDriver, proto } from "containerkit/vocab";

// WordPress and MySQL on one network. The database volume is removed with the
// project on Ctrl+C. Passwords here are for a local example.

const db = "wordpress";
const user = "wordpress";
const password = "wordpress";
const rootPassword = "example";

function App() {
	useShutdown({ removeOrphans: true, volumes: true });
	return (
		<Compose name="wordpress">
			<Network name="wordpress" driver={netDriver.Bridge} />
			<Volume name="db" />
			<Service name="db" image="mysql:8.4">
				<Env name="MYSQL_DATABASE" value={db} />
				<Env name="MYSQL_USER" value={user} />
				<Env name="MYSQL_PASSWORD" value={password} />
				<Env name="MYSQL_ROOT_PASSWORD" value={rootPassword} />
				<Volume source="db" target="/var/lib/mysql" />
				<Network name="wordpress" />
				<Health
					test={["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-uroot", `-p${rootPassword}`]}
					interval="5s"
					timeout="5s"
					retries={20}
				/>
			</Service>
			<Service name="wordpress" image="wordpress:6-apache">
				<DependsOn name="db" condition={depend.Healthy} />
				<Env name="WORDPRESS_DB_HOST" value="db" />
				<Env name="WORDPRESS_DB_NAME" value={db} />
				<Env name="WORDPRESS_DB_USER" value={user} />
				<Env name="WORDPRESS_DB_PASSWORD" value={password} />
				<Publish protocol={proto.TCP} hostIP="127.0.0.1" hostPort={8080} containerPort={80} />
				<Network name="wordpress" />
				<Health
					test={["CMD-SHELL", "bash -c '</dev/tcp/127.0.0.1/80'"]}
					interval="10s"
					timeout="5s"
					retries={12}
				/>
			</Service>
		</Compose>
	);
}

export default root(
	{
		fromEnv: true,
		onExit: "down",
		wait: true,
		onUp: () => console.log("WordPress at http://127.0.0.1:8080"),
	},
	App,
);
