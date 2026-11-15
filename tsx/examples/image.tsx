import {
	Arg,
	Command,
	Comment,
	Copy,
	Dockerfile,
	Entrypoint,
	Env,
	Expose,
	From,
	HealthCheck,
	Label,
	Run,
	RunLayer,
	StopSignal,
	User,
	Volume,
	Workdir,
	exportDockerfile,
	type DockerfileNode,
} from "containerkit/dockerfile";

type Mode = "production" | "development";

type ImageConfig = {
	name: string;
	version: string;
	mode: Mode;
	node: string;
	platform?: string;
	ports: readonly string[];
	buildPackages: readonly string[];
	sourceExclude: readonly string[];
	dataDirs: readonly string[];
};

const manifests = ["package.json", "pnpm-lock.yaml"] as const;

function configFromEnv(env: NodeJS.ProcessEnv): ImageConfig {
	const mode: Mode = env.APP_ENV === "production" ? "production" : "development";
	const ports = (env.PORTS ?? "8080").split(",").map((port) => port.trim()).filter((port) => port !== "");
	if (ports.length === 0) {
		throw new Error("PORTS must name at least one port");
	}
	const dev = mode === "development";
	return {
		name: env.APP_NAME?.trim() || "ledger",
		version: env.APP_VERSION?.trim() || "0.8.1",
		mode,
		node: env.NODE_VERSION?.trim() || "22",
		platform: env.TARGET_PLATFORM?.trim() || undefined,
		ports,
		buildPackages: dev ? ["git", "python3", "make", "procps"] : ["git", "python3", "make"],
		sourceExclude: dev ? ["node_modules", "dist"] : ["node_modules", "dist", "**/*.test.ts", "**/*.spec.ts"],
		dataDirs: dev ? ["/var/lib/ledger", "/var/log/ledger"] : ["/var/lib/ledger"],
	};
}

function LedgerImage(props: { config: ImageConfig }): DockerfileNode {
	const { config } = props;
	const dev = config.mode === "development";
	return (
		<Dockerfile>
			<Arg name="APP_ENV" default={config.mode} />
			<Arg name="APP_VERSION" default={config.version} />
			<Toolchain config={config} />
			<Dependencies />
			<BuildStage config={config} />
			{dev && <TestStage />}
			<ReleaseStage />
			<Runtime config={config} />
		</Dockerfile>
	);
}

function Toolchain(props: { config: ImageConfig }) {
	const { config } = props;
	const packages = config.buildPackages.join(" ");
	return (
		<>
			<Comment text={`toolchain: node ${config.node} and OS packages`} />
			<From image="node" tag={`${config.node}-bookworm`} as="toolchain" platform={config.platform} />
			<Workdir>/src</Workdir>
			<Run
				cmd={`apt-get update && apt-get install -y --no-install-recommends ${packages} && rm -rf /var/lib/apt/lists/*`}
			/>
		</>
	);
}

function Dependencies() {
	return (
		<>
			<Comment text="deps: lockfile install, kept apart from the source copy" />
			<From image="toolchain" as="deps" />
			<Workdir>/src</Workdir>
			{manifests.map((file) => (
				<Copy key={file} src={file} dest={`/src/${file}`} />
			))}
			<Env name="PNPM_HOME">/pnpm</Env>
			<Env name="PATH">/pnpm:$PATH</Env>
			<Run cmd="corepack enable" />
			<RunLayer cmd="corepack prepare pnpm@9.15.0 --activate" />
			<Run cmd="pnpm install --frozen-lockfile" mount="type=cache,id=pnpm,target=/pnpm/store" />
		</>
	);
}

function BuildStage(props: { config: ImageConfig }) {
	const { config } = props;
	const dev = config.mode === "development";
	return (
		<>
			<Comment text="build: source on its own layer above the install" />
			<From image="deps" as="build" />
			<Workdir>/src</Workdir>
			<Copy src="." dest="/src" exclude={config.sourceExclude} />
			<RunLayer cmd="pnpm run build" />
			{dev && <RunLayer cmd="pnpm run build:types" />}
		</>
	);
}

function TestStage() {
	return (
		<>
			<Comment text="test: present only for a development export" />
			<From image="build" as="test" />
			<Workdir>/src</Workdir>
			<RunLayer cmd="pnpm test" />
		</>
	);
}

function ReleaseStage() {
	return (
		<>
			<Comment text="release: production dependencies only" />
			<From image="build" as="release" />
			<Workdir>/src</Workdir>
			<RunLayer cmd="pnpm prune --prod" />
		</>
	);
}

function Runtime(props: { config: ImageConfig }): DockerfileNode {
	const { config } = props;
	const dev = config.mode === "development";
	const dirs = config.dataDirs.join(" ");
	const labels = {
		"org.opencontainers.image.title": config.name,
		"org.opencontainers.image.version": config.version,
		"com.containerkit.env": config.mode,
	};
	return (
		<>
			<Comment text={`${config.name} runtime (${config.mode})`} />
			<From image="node" tag={`${config.node}-bookworm-slim`} as="runtime" platform={config.platform} />
			<Arg name="APP_VERSION" />
			<Run cmd="apt-get update && apt-get install -y --no-install-recommends tini ca-certificates && rm -rf /var/lib/apt/lists/*" />
			{dev && <RunLayer cmd="apt-get update && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/*" />}
			<RunLayer cmd={`mkdir -p ${dirs} && chown -R node:node ${dirs}`} />
			<Copy from="release" src="/src/dist" dest="/app/dist" chown="node:node" link />
			<Copy from="release" src="/src/node_modules" dest="/app/node_modules" chown="node:node" link />
			<Copy from="release" src="/src/package.json" dest="/app/package.json" chown="node:node" link />
			<User>node</User>
			<Workdir>/app</Workdir>
			<Env name="NODE_ENV" value={config.mode} />
			<Env name="APP_NAME" value={config.name} />
			<Env name="APP_VERSION" value={config.version} />
			{Object.entries(labels).map(([name, value]) => (
				<Label key={name} name={name} value={value} />
			))}
			{config.ports.map((port) => (
				<Expose key={port} port={port} />
			))}
			<Volume paths={config.dataDirs} />
			<StopSignal signal="SIGTERM" />
			<HealthCheck
				test={["CMD", "node", "-e", `fetch('http://127.0.0.1:${config.ports[0]}/healthz').then((r)=>process.exit(r.ok?0:1))`]}
				interval="30s"
				timeout="3s"
				startPeriod={dev ? "5s" : "15s"}
				retries={3}
			/>
			<Entrypoint executable="/usr/bin/tini" args={["--"]} />
			<Command executable="node" args={["/app/dist/server.js"]} />
		</>
	);
}

process.stdout.write(exportDockerfile(<LedgerImage config={configFromEnv(process.env)} />));
