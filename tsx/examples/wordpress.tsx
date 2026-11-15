import { existsSync, writeFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
	Compose,
	DependsOn,
	Env,
	Expose,
	Health,
	Network,
	Publish,
	Restart,
	Service,
	Volume,
	exportYAML,
} from "containerkit";
import { root, runIfMain } from "containerkit/live";
import { depend, netDriver, proto, restartPolicy, vol } from "containerkit/vocab";

// Three WordPress containers behind HAProxy, with MySQL and, off Windows, Portainer.
// The Go program is examples/wordpress. This writes the same haproxy.cfg and
// compose file, then brings the project up.

const numWordPress = 3;
const services = Array.from({ length: numWordPress }, (_, i) => `wordpress-example-${i + 1}`);
const linux = process.platform === "linux";

function haproxy(backends: string[]): string {
	const lines = [
		"global",
		"	log stdout format raw local0",
		"defaults",
		"	log		global",
		"	mode	http",
		"	option	httplog",
		"	option	dontlognull",
		"	timeout connect 5000",
		"	timeout client  50000",
		"	timeout server  50000",
		"frontend http_front",
		"	bind *:80",
		"	default_backend wordpress_back",
		"backend wordpress_back",
		"	balance roundrobin",
		...backends.map((backend, i) => `	server wp${i + 1} ${backend}:80 check`),
		"",
	];
	return lines.join("\n");
}

export const haproxyConfig = haproxy(services);

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const wordpressDir = join(repo, "examples/wordpress");
const haproxyPath = join(wordpressDir, "haproxy.cfg");

function dockerSock(): string {
	const uid = process.getuid?.();
	if (uid !== undefined) {
		const rootless = `/var/run/user/${uid}/docker.sock`;
		if (existsSync(rootless)) {
			return rootless;
		}
	}
	return "/var/run/docker.sock";
}

const tree = (
	<Compose name={`gocontain-wp-scale-${numWordPress}`}>
		<Volume name="wordpress-data" />
		<Volume name="database-data" />
		<Volume name="portainer-data" />
		<Network name="wordpress-network" driver={netDriver.Bridge} />
		<Service name="database-example" image="mysql:8.0">
			<Env name="MYSQL_DATABASE" value="exampledb" />
			<Env name="MYSQL_PASSWORD" value="examplepass" />
			<Env name="MYSQL_USER" value="exampleuser" />
			<Env name="MYSQL_RANDOM_ROOT_PASSWORD" value="1" />
			<Health
				test={["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-pexamplepass"]}
				startPeriod="5s"
				interval="10s"
				timeout="20s"
				retries={3}
			/>
			{linux ? (
				<Volume source="database-data" target="/var/lib/mysql/" />
			) : (
				<Volume kind={vol.Bind} source={join(wordpressDir, "database/")} target="/var/lib/mysql/" />
			)}
			<Network name="wordpress-network" />
		</Service>
		{services.map((name, i) => (
			<Service key={name} name={name} image="wordpress:latest">
				<DependsOn name="database-example" condition={depend.Healthy} />
				{i > 0 && <DependsOn name={services[i - 1]!} />}
				<Env name="WORDPRESS_DB_HOST" value="database-example" />
				<Env name="WORDPRESS_DB_USER" value="exampleuser" />
				<Env name="WORDPRESS_DB_PASSWORD" value="examplepass" />
				<Env name="WORDPRESS_DB_NAME" value="exampledb" />
				<Expose port={80} proto={proto.TCP} />
				<Health
					test={["CMD", "curl", "-f", "http://localhost/wp-login.php"]}
					startPeriod="5s"
					interval="10s"
					timeout="20s"
					retries={3}
				/>
				<Restart policy={restartPolicy.UnlessStopped} />
				{linux ? (
					<Volume source="wordpress-data" target="/var/www/html" />
				) : (
					<Volume kind={vol.Bind} source={join(wordpressDir, "src")} target="/var/www/html" />
				)}
				<Network name="wordpress-network" />
			</Service>
		))}
		<Service name="proxy-container" image="haproxy:latest" command={["-f", "/usr/local/etc/haproxy/haproxy.cfg"]}>
			{services.map((name) => (
				<DependsOn key={name} name={name} condition={depend.Healthy} />
			))}
			<Publish protocol={proto.TCP} hostIP="0.0.0.0" hostPort={80} containerPort={80} />
			<Volume
				kind={vol.Bind}
				source={haproxyPath}
				target="/usr/local/etc/haproxy/haproxy.cfg"
				readOnly
			/>
			<Network name="wordpress-network" />
		</Service>
		{process.platform !== "win32" && (
			<Service name="portainer-container" image="portainer/portainer-ce:latest">
				<Publish protocol={proto.TCP} hostIP="0.0.0.0" hostPort={9000} containerPort={9000} />
				<Volume source="portainer-data" target="/data" />
				<Volume kind={vol.Bind} source={dockerSock()} target="/var/run/docker.sock" />
				<Network name="wordpress-network" />
			</Service>
		)}
	</Compose>
);

export { tree };

const live = root({ fromEnv: true, onExit: "down" }, async function App(session) {
	writeFileSync(haproxyPath, haproxyConfig);
	writeFileSync(join(wordpressDir, "docker-compose.yaml"), exportYAML(tree));
	await session.composeUp(tree, { detach: true, removeOrphans: true, timeout: "3s" });
	return tree;
});

export default live;

await runIfMain(live, import.meta.url);
