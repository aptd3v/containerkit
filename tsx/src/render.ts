import { exportDockerfile } from "./docker-file.js";
import { checkChildren, flatten, isReactElement, reactCompiledJSX, type Node } from "./node.js";
import { toYAML } from "./yaml.js";

const servicePropMap: Record<string, string> = {
	workingDir: "working_dir",
	readOnly: "read_only",
	domainName: "domainname",
	macAddress: "mac_address",
	stopSignal: "stop_signal",
	stopTimeout: "stop_grace_period",
	networkMode: "network_mode",
	volumeDriver: "volume_driver",
	volumesFrom: "volumes_from",
	groupAdd: "group_add",
	securityOpt: "security_opt",
	capAdd: "cap_add",
	capDrop: "cap_drop",
	dnsOption: "dns_opt",
	dnsSearch: "dns_search",
	oomKillDisable: "oom_kill_disable",
	oomScoreAdj: "oom_score_adj",
	pidsLimit: "pids_limit",
	shmSize: "shm_size",
	memoryReservation: "mem_reservation",
	memorySwap: "memswap_limit",
	memorySwappiness: "mem_swappiness",
	cpuShares: "cpu_shares",
	cpuPeriod: "cpu_period",
	cpuQuota: "cpu_quota",
	cpuSet: "cpuset",
	cpuSetMems: "cpuset_mems",
	cpuCount: "cpu_count",
	cpuPercent: "cpu_percent",
	cpuRealtimePeriod: "cpu_rt_period",
	cpuRealtimeRuntime: "cpu_rt_runtime",
	cgroupParent: "cgroup_parent",
	cgroupNS: "cgroupns_mode",
	userNS: "userns_mode",
	pullPolicy: "pull_policy",
	pullRefreshAfter: "pull_refresh_after",
	containerName: "container_name",
	tmpfsMount: "tmpfs",
	memory: "mem_limit",
	interactive: "stdin_open",
	init: "init",
	cpus: "cpus",
	networkDisabled: "network_mode",
};

const skipServiceProps = new Set(["name"]);

export function render(node: Node): string {
	if (isReactElement(node)) {
		throw new Error(reactCompiledJSX);
	}
	return toYAML(toDocument(node));
}

// exportYAML prints a Compose file. It does not call the daemon and it does not
// restart a session. The node must be Compose, and it must have a name.
export function exportYAML(node: Node): string {
	if (isReactElement(node)) {
		throw new Error(reactCompiledJSX);
	}
	if (node.$$kind !== "compose") {
		throw new Error(`containerkit: exportYAML requires a Compose node, got <${node.$$kind}>`);
	}
	const name = node.props.name;
	if (typeof name !== "string" || name.trim() === "") {
		throw new Error("containerkit: Compose requires name");
	}
	return render(node);
}

function toDocument(node: Node): Record<string, unknown> {
	const root = projectRoot(node);
	switch (root.$$kind) {
		case "compose":
			return renderCompose(root);
		case "service":
			return wrapNamed("services", root, renderService(root));
		case "job":
			return wrapNamed("jobs", root, renderService(root));
		case "container":
			return wrapNamed("services", root, renderService(root));
		default:
			throw new Error(
				`containerkit/tsx: root must be Compose, Service, Job, or Container, got <${root.$$kind}>`,
			);
	}
}

/** The Compose, Service, Job, or Container a root callback applies. A fragment must hold exactly one. */
export function projectRoot(node: Node): Node {
	const root = unwrapRoot(node);
	switch (root.$$kind) {
		case "compose":
		case "service":
		case "job":
		case "container":
			return root;
		default:
			throw new Error(
				`containerkit/tsx: root must be Compose, Service, Job, or Container, got <${root.$$kind}>`,
			);
	}
}

function unwrapRoot(node: Node): Node {
	if (node.$$kind !== "fragment") {
		return node;
	}
	const kids = node.children.filter((c) => c.$$kind !== "fragment");
	if (kids.length === 1) {
		return unwrapRoot(kids[0]);
	}
	throw new Error("containerkit/tsx: fragment root must have a single Compose, Service, Job, or Container");
}

function wrapNamed(section: string, node: Node, body: Record<string, unknown>): Record<string, unknown> {
	const name = resourceName(node, "app");
	return { name, [section]: { [resourceName(node, "service")]: body } };
}

function renderCompose(node: Node): Record<string, unknown> {
	assertChildren(node);
	const doc: Record<string, unknown> = {};
	const name = str(node.props.name);
	if (name) {
		doc.name = name;
	}
	const services: Record<string, unknown> = {};
	const jobs: Record<string, unknown> = {};
	const networks: Record<string, unknown> = {};
	const volumes: Record<string, unknown> = {};
	const secrets: Record<string, unknown> = {};
	const configs: Record<string, unknown> = {};
	const models: Record<string, unknown> = {};
	const includes: unknown[] = [];

	for (const child of node.children) {
		switch (child.$$kind) {
			case "service":
			case "container":
				services[resourceName(child, "service")] = renderService(child);
				break;
			case "job":
				jobs[resourceName(child, "job")] = renderService(child);
				break;
			case "network":
				networks[resourceName(child, "network")] = renderProjectNetwork(child);
				break;
			case "volume":
				volumes[resourceName(child, "volume")] = renderProjectVolume(child);
				break;
			case "secret":
				secrets[resourceName(child, "secret")] = renderProjectSecret(child);
				break;
			case "fileConfig":
				configs[resourceName(child, "config")] = renderProjectSecret(child);
				break;
			case "model":
				models[resourceName(child, "model")] = objectFrom(child, { source: "model" });
				break;
			case "include":
				includes.push(renderInclude(child));
				break;
			default:
				throwInvalid(node, child);
		}
	}

	if (Object.keys(services).length) {
		doc.services = services;
	}
	if (Object.keys(jobs).length) {
		doc.jobs = jobs;
	}
	if (Object.keys(networks).length) {
		doc.networks = networks;
	}
	if (Object.keys(volumes).length) {
		doc.volumes = volumes;
	}
	if (Object.keys(secrets).length) {
		doc.secrets = secrets;
	}
	if (Object.keys(configs).length) {
		doc.configs = configs;
	}
	if (Object.keys(models).length) {
		doc.models = models;
	}
	if (includes.length) {
		doc.include = includes;
	}
	return doc;
}

function renderService(node: Node): Record<string, unknown> {
	assertChildren(node);
	const svc: Record<string, unknown> = {};
	assignServiceProps(svc, node.props);

	const ports: unknown[] = [];
	const expose: string[] = [];
	const environment: Record<string, unknown> = {};
	const labels: Record<string, unknown> = {};
	const volumes: unknown[] = [];
	const networks: Record<string, unknown> = {};
	const extraHosts: Record<string, unknown> = {};
	const sysctls: Record<string, unknown> = {};
	const annotations: Record<string, unknown> = {};
	const storageOpt: Record<string, unknown> = {};
	const envFiles: unknown[] = [];
	const secrets: unknown[] = [];
	const configs: unknown[] = [];
	const devices: unknown[] = [];
	const gpus: unknown[] = [];
	const ulimits: Record<string, unknown> = {};
	const dependsOn: Record<string, unknown> = {};
	const models: Record<string, unknown> = {};

	const applyChild = (child: Node): void => {
		switch (child.$$kind) {
			case "container":
				assignServiceProps(svc, child.props);
				assertChildren(child);
				for (const nested of child.children) {
					applyChild(nested);
				}
				break;
			case "publish":
				ports.push(renderPublish(child));
				break;
			case "expose":
				expose.push(formatExpose(child));
				break;
			case "env":
				putKV(environment, child, "environment");
				break;
			case "label":
				putKV(labels, child, "labels");
				break;
			case "volume":
				volumes.push(renderMount(child));
				break;
			case "network":
				networks[resourceName(child, "network")] = renderServiceNetwork(child);
				break;
			case "health":
				svc.healthcheck = renderHealth(child);
				break;
			case "deploy":
				svc.deploy = renderDeploy(child);
				break;
			case "build":
				svc.build = renderBuild(child);
				break;
			case "logging":
				svc.logging = renderLogging(child);
				break;
			case "dependsOn":
				dependsOn[resourceName(child, "service")] = renderDepends(child);
				break;
			case "restart":
				svc.restart = renderRestart(child);
				break;
			case "extraHost":
				if (typeof child.props.hostname === "string") {
					extraHosts[child.props.hostname] = child.props.ip;
				}
				break;
			case "sysctl":
				putKV(sysctls, child, "sysctls");
				break;
			case "annotation":
				putKV(annotations, child, "annotations");
				break;
			case "storageOpt":
				putKV(storageOpt, child, "storage_opt");
				break;
			case "envFile":
				envFiles.push(renderEnvFile(child));
				break;
			case "secret":
				secrets.push(renderServiceSecret(child));
				break;
			case "fileConfig":
				configs.push(renderServiceSecret(child));
				break;
			case "device":
				devices.push(renderDevice(child));
				break;
			case "gpu":
				gpus.push(renderGPU(child));
				break;
			case "ulimit":
				ulimits[resourceName(child, "ulimit")] = renderUlimit(child);
				break;
			case "model":
				models[resourceName(child, "model")] = objectFrom(child, {});
				break;
			case "develop":
				svc.develop = renderDevelop(child);
				break;
			case "postStart":
			case "preStart":
			case "preStop": {
				const hooks = (svc[child.$$kind] as unknown[]) ?? [];
				hooks.push(objectFrom(child, {}));
				svc[camelToSnake(child.$$kind)] = hooks;
				break;
			}
			case "providerOption": {
				if (!isPlain(svc.provider)) {
					svc.provider = { options: {} };
				}
				const provider = svc.provider as Record<string, unknown>;
				if (!isPlain(provider.options)) {
					provider.options = {};
				}
				putKV(provider.options as Record<string, unknown>, child, "provider");
				break;
			}
			default:
				throwInvalid(node, child);
		}
	};

	for (const child of node.children) {
		applyChild(child);
	}

	if (ports.length) {
		svc.ports = ports;
	}
	if (expose.length) {
		svc.expose = expose;
	}
	if (Object.keys(environment).length) {
		svc.environment = environment;
	}
	if (Object.keys(labels).length) {
		svc.labels = labels;
	}
	if (volumes.length) {
		svc.volumes = volumes;
	}
	if (Object.keys(networks).length) {
		svc.networks = networks;
	}
	if (Object.keys(extraHosts).length) {
		svc.extra_hosts = extraHosts;
	}
	if (Object.keys(sysctls).length) {
		svc.sysctls = sysctls;
	}
	if (Object.keys(annotations).length) {
		svc.annotations = annotations;
	}
	if (Object.keys(storageOpt).length) {
		svc.storage_opt = storageOpt;
	}
	if (envFiles.length) {
		svc.env_file = envFiles;
	}
	if (secrets.length) {
		svc.secrets = secrets;
	}
	if (configs.length) {
		svc.configs = configs;
	}
	if (devices.length) {
		svc.devices = devices;
	}
	if (gpus.length) {
		svc.gpus = gpus;
	}
	if (Object.keys(ulimits).length) {
		svc.ulimits = ulimits;
	}
	if (Object.keys(dependsOn).length) {
		svc.depends_on = dependsOn;
	}
	if (Object.keys(models).length) {
		svc.models = models;
	}
	return svc;
}

function assignServiceProps(svc: Record<string, unknown>, props: Record<string, unknown>): void {
	for (const [k, v] of Object.entries(props)) {
		if (v === undefined || skipServiceProps.has(k)) {
			continue;
		}
		if (k === "networkDisabled") {
			if (v === true && svc.network_mode === undefined) {
				svc.network_mode = "none";
			}
			continue;
		}
		if (k === "securityOpt" && typeof v === "string" && v !== "") {
			svc.security_opt = [v];
			continue;
		}
		if (k === "profile") {
			if (typeof v === "string" && v !== "") {
				const cur = svc.profiles;
				const list = Array.isArray(cur) ? cur.slice() : [];
				list.push(v);
				svc.profiles = list;
			}
			continue;
		}
		const key = servicePropMap[k] ?? camelToSnake(k);
		svc[key] = normalizeScalar(k, v);
	}
}

function formatExpose(node: Node): string {
	const protoRaw = node.props.proto;
	const proto = Array.isArray(protoRaw) ? protoRaw[0] : protoRaw;
	const p = typeof proto === "string" && proto !== "" ? proto : "tcp";
	return `${node.props.port}/${p}`;
}

function renderPublish(node: Node): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	if (node.props.containerPort !== undefined) {
		out.target = node.props.containerPort;
	}
	if (node.props.hostPort !== undefined) {
		out.published = String(node.props.hostPort);
	} else if (node.props.published !== undefined) {
		out.published = String(node.props.published);
	}
	if (node.props.protocol !== undefined) {
		out.protocol = node.props.protocol;
	}
	if (node.props.hostIP !== undefined) {
		out.host_ip = node.props.hostIP;
	}
	if (node.props.mode !== undefined) {
		out.mode = node.props.mode;
	}
	if (node.props.name !== undefined) {
		out.name = node.props.name;
	}
	if (node.props.appProtocol !== undefined) {
		out.app_protocol = node.props.appProtocol;
	}
	return out;
}

function renderMount(node: Node): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	const kind = str(node.props.kind) ?? inferVolumeKind(node);
	if (kind) {
		out.type = kind;
	}
	if (node.props.source !== undefined) {
		out.source = node.props.source;
	} else if (kind !== "tmpfs" && node.props.name !== undefined) {
		out.source = node.props.name;
	}
	if (node.props.target !== undefined) {
		out.target = node.props.target;
	}
	if (node.props.readOnly === true) {
		out.read_only = true;
	}
	if (node.props.size !== undefined) {
		out.tmpfs = { size: node.props.size };
	}
	if (node.props.subPath !== undefined || node.props.noCopy === true) {
		out.volume = omitEmpty({
			subpath: node.props.subPath,
			nocopy: node.props.noCopy === true ? true : undefined,
		});
	}
	if (
		node.props.propagation !== undefined ||
		node.props.seLinux !== undefined ||
		node.props.createHostPath !== undefined ||
		node.props.recursive !== undefined
	) {
		out.bind = omitEmpty({
			propagation: node.props.propagation,
			selinux: node.props.seLinux,
			create_host_path: node.props.createHostPath,
			recursive: node.props.recursive,
		});
	}
	return out;
}

function inferVolumeKind(node: Node): string {
	const source = str(node.props.source);
	if (node.props.size !== undefined && !source) {
		return "tmpfs";
	}
	if (source && (source.startsWith("/") || source.startsWith(".") || source.startsWith("\\"))) {
		return "bind";
	}
	return "volume";
}

function renderHealth(node: Node): Record<string, unknown> {
	if (node.props.disabled === true) {
		return { disable: true };
	}
	const out: Record<string, unknown> = {};
	if (node.props.test !== undefined) {
		out.test = Array.isArray(node.props.test) ? node.props.test : [node.props.test];
	}
	const interval = duration(node.props.interval);
	if (interval) {
		out.interval = interval;
	}
	const timeout = duration(node.props.timeout);
	if (timeout) {
		out.timeout = timeout;
	}
	const startPeriod = duration(node.props.startPeriod);
	if (startPeriod) {
		out.start_period = startPeriod;
	}
	const startInterval = duration(node.props.startInterval);
	if (startInterval) {
		out.start_interval = startInterval;
	}
	if (typeof node.props.retries === "number") {
		out.retries = node.props.retries;
	}
	return out;
}

function renderDeploy(node: Node): Record<string, unknown> {
	assertChildren(node);
	const out: Record<string, unknown> = {};
	if (node.props.mode !== undefined) {
		out.mode = node.props.mode;
	}
	if (typeof node.props.replicas === "number") {
		out.replicas = node.props.replicas;
	}
	if (node.props.endpointMode !== undefined) {
		out.endpoint_mode = node.props.endpointMode;
	}
	const limits = omitEmpty({
		memory: node.props.memory,
		cpus: node.props.cpus,
		pids: node.props.pidsLimit,
	});
	const labels: Record<string, unknown> = {};
	let reservations: Record<string, unknown> | undefined;
	for (const child of node.children) {
		switch (child.$$kind) {
			case "label":
				putKV(labels, child, "labels");
				break;
			case "reserve":
				reservations = omitEmpty({
					memory: child.props.memory,
					cpus: child.props.cpus,
					pids: child.props.pidsLimit,
				});
				break;
			case "restart":
				out.restart_policy = objectFrom(child, {
					maxRetries: "max_attempts",
				});
				break;
			case "update":
				out.update_config = objectFrom(child, {});
				break;
			case "rollback":
				out.rollback_config = objectFrom(child, {});
				break;
			case "placement":
				out.placement = objectFrom(child, { maxReplicas: "max_replicas" });
				break;
			case "gpu":
			case "discrete":
				break;
			default:
				throwInvalid(node, child);
		}
	}
	if (Object.keys(limits).length || reservations) {
		out.resources = omitEmpty({
			limits: Object.keys(limits).length ? limits : undefined,
			reservations,
		});
	}
	if (Object.keys(labels).length) {
		out.labels = labels;
	}
	return out;
}

function renderBuild(node: Node): Record<string, unknown> {
	assertChildren(node);
	const out = objectFrom(node, {
		source: "context",
		dockerfilePath: "dockerfile",
		networkMode: "network",
		noCache: "no_cache",
		noCacheFilter: "no_cache_filter",
		cacheFrom: "cache_from",
		cacheTo: "cache_to",
		shmSize: "shm_size",
		tag: "tags",
	});
	const args: Record<string, unknown> = {};
	const labels: Record<string, unknown> = {};
	const extraHosts: Record<string, unknown> = {};
	const ssh: unknown[] = [];
	const additional: Record<string, unknown> = {};
	for (const child of node.children) {
		switch (child.$$kind) {
			case "dockerfile":
				if (out.dockerfile_inline !== undefined || node.props.dockerfilePath || node.props.parseDockerfile || node.props.readDockerfile) {
					throw new Error("containerkit/tsx: Build accepts one Dockerfile, and not alongside dockerfilePath, parseDockerfile, or readDockerfile");
				}
				out.dockerfile_inline = exportDockerfile(child);
				break;
			case "env":
				putKV(args, child, "args");
				break;
			case "label":
				putKV(labels, child, "labels");
				break;
			case "extraHost":
				if (typeof child.props.hostname === "string") {
					extraHosts[child.props.hostname] = child.props.ip;
				}
				break;
			case "ssh":
				ssh.push(omitEmpty({ id: child.props.id, path: child.props.path }));
				break;
			case "additionalContext":
				putKV(additional, child, "additional_contexts");
				break;
			case "secret":
			case "ulimit":
				break;
			default:
				throwInvalid(node, child);
		}
	}
	if (Object.keys(args).length) {
		out.args = args;
	}
	if (Object.keys(labels).length) {
		out.labels = labels;
	}
	if (Object.keys(extraHosts).length) {
		out.extra_hosts = extraHosts;
	}
	if (ssh.length) {
		out.ssh = ssh;
	}
	if (Object.keys(additional).length) {
		out.additional_contexts = additional;
	}
	return out;
}

function renderLogging(node: Node): Record<string, unknown> {
	assertChildren(node);
	const out: Record<string, unknown> = {};
	if (node.props.driver !== undefined) {
		out.driver = node.props.driver;
	}
	const options: Record<string, unknown> = {};
	for (const child of node.children) {
		if (child.$$kind === "driverOpt") {
			putKV(options, child, "options");
		} else {
			throwInvalid(node, child);
		}
	}
	if (Object.keys(options).length) {
		out.options = options;
	}
	return out;
}

function renderDepends(node: Node): Record<string, unknown> {
	return omitEmpty({
		condition: node.props.condition ?? "service_started",
		restart: node.props.restartOnFailure,
		required: node.props.required,
	});
}

function renderRestart(node: Node): string {
	const policy = str(node.props.policy) ?? "no";
	if (policy === "on-failure" && typeof node.props.maxRetries === "number") {
		return `on-failure:${node.props.maxRetries}`;
	}
	return policy;
}

function renderProjectNetwork(node: Node): Record<string, unknown> {
	assertChildren(node);
	const out = objectFrom(node, {
		enableIPv4: "enable_ipv4",
		enableIPv6: "enable_ipv6",
		ipamDriver: "driver",
	});
	delete out.name;
	delete out.alias;
	delete out.ipv4;
	delete out.ipv6;
	delete out.mac_address;
	delete out.priority;
	delete out.gateway_priority;
	delete out.interface_name;
	delete out.link_local;
	delete out.use_service_name_alias;
	const labels: Record<string, unknown> = {};
	const ipamOpts: Record<string, unknown> = {};
	const driverOpts: Record<string, unknown> = {};
	const ipamConfig: Record<string, unknown> = {};
	if (node.props.subnet) {
		ipamConfig.subnet = node.props.subnet;
	}
	if (node.props.gateway) {
		ipamConfig.gateway = node.props.gateway;
	}
	if (node.props.ipRange) {
		ipamConfig.ip_range = node.props.ipRange;
	}
	for (const child of node.children) {
		switch (child.$$kind) {
			case "label":
				putKV(labels, child, "labels");
				break;
			case "driverOpt":
				putKV(driverOpts, child, "driver_opts");
				break;
			case "ipamOpt":
				putKV(ipamOpts, child, "ipam");
				break;
			case "auxAddress":
				break;
			default:
				throwInvalid(node, child);
		}
	}
	if (Object.keys(labels).length) {
		out.labels = labels;
	}
	if (Object.keys(driverOpts).length) {
		out.driver_opts = driverOpts;
	}
	if (Object.keys(ipamConfig).length || Object.keys(ipamOpts).length || node.props.ipamDriver) {
		out.ipam = omitEmpty({
			driver: node.props.ipamDriver,
			config: Object.keys(ipamConfig).length ? [ipamConfig] : undefined,
			options: Object.keys(ipamOpts).length ? ipamOpts : undefined,
		});
	}
	if (Object.keys(out).length === 0) {
		return {};
	}
	return out;
}

function renderProjectVolume(node: Node): Record<string, unknown> {
	assertChildren(node);
	const out = omitEmpty({
		driver: node.props.driver,
		external: node.props.external === true ? true : undefined,
	});
	const labels: Record<string, unknown> = {};
	const driverOpts: Record<string, unknown> = {};
	for (const child of node.children) {
		switch (child.$$kind) {
			case "label":
				putKV(labels, child, "labels");
				break;
			case "driverOpt":
				putKV(driverOpts, child, "driver_opts");
				break;
			default:
				throwInvalid(node, child);
		}
	}
	if (Object.keys(labels).length) {
		out.labels = labels;
	}
	if (Object.keys(driverOpts).length) {
		out.driver_opts = driverOpts;
	}
	return out;
}

function renderProjectSecret(node: Node): Record<string, unknown> {
	assertChildren(node);
	const out = objectFrom(node, { fromFile: "file", fileMode: "mode" });
	delete out.name;
	delete out.target;
	const labels: Record<string, unknown> = {};
	for (const child of node.children) {
		if (child.$$kind === "label") {
			putKV(labels, child, "labels");
		} else {
			throwInvalid(node, child);
		}
	}
	if (Object.keys(labels).length) {
		out.labels = labels;
	}
	return out;
}

function renderServiceNetwork(node: Node): Record<string, unknown> | null {
	assertChildren(node);
	const aliases = asStringList(node.props.alias);
	const out = omitEmpty({
		aliases: aliases.length ? aliases : undefined,
		ipv4_address: node.props.ipv4,
		ipv6_address: node.props.ipv6,
		mac_address: node.props.macAddress,
		priority: node.props.priority,
		gateway_priority: node.props.gatewayPriority,
		interface_name: node.props.interfaceName,
		link_local_ips: asStringList(node.props.linkLocal),
	});
	return Object.keys(out).length ? out : {};
}

function renderServiceSecret(node: Node): Record<string, unknown> | string {
	const name = resourceName(node, "secret");
	if (node.props.target === undefined && node.props.uid === undefined && node.props.gid === undefined) {
		return name;
	}
	return omitEmpty({
		source: node.props.source ?? name,
		target: node.props.target,
		uid: node.props.uid,
		gid: node.props.gid,
		mode: node.props.fileMode,
	});
}

function renderEnvFile(node: Node): Record<string, unknown> | string {
	const path = str(node.props.path) ?? "";
	if (node.props.required === undefined && node.props.format === undefined) {
		return path;
	}
	return omitEmpty({
		path,
		required: node.props.required,
		format: node.props.format,
	});
}

function renderInclude(node: Node): Record<string, unknown> | string {
	assertChildren(node);
	const path = str(node.props.path) ?? "";
	const envFiles = node.children
		.filter((c) => c.$$kind === "envFile")
		.map((c) => str(c.props.path))
		.filter((p): p is string => !!p);
	if (!node.props.projectDirectory && envFiles.length === 0) {
		return path;
	}
	return omitEmpty({
		path,
		project_directory: node.props.projectDirectory,
		env_file: envFiles.length ? envFiles : undefined,
	});
}

function renderDevice(node: Node): Record<string, unknown> {
	return omitEmpty({
		source: node.props.source,
		target: node.props.target,
		permissions: node.props.perms,
	});
}

function renderGPU(node: Node): Record<string, unknown> {
	assertChildren(node);
	const out = objectFrom(node, {});
	const driverOpts: Record<string, unknown> = {};
	for (const child of node.children) {
		if (child.$$kind === "driverOpt") {
			putKV(driverOpts, child, "options");
		} else {
			throwInvalid(node, child);
		}
	}
	if (Object.keys(driverOpts).length) {
		out.driver_opts = driverOpts;
	}
	return out;
}

function renderUlimit(node: Node): Record<string, unknown> | number {
	if (node.props.soft === undefined && typeof node.props.hard === "number") {
		return node.props.hard as number;
	}
	if (typeof node.props.soft === "number" && node.props.hard === undefined) {
		return node.props.soft as number;
	}
	return omitEmpty({
		soft: node.props.soft,
		hard: node.props.hard,
	});
}

function renderDevelop(node: Node): Record<string, unknown> {
	assertChildren(node);
	const watch = node.children
		.filter((c) => c.$$kind === "watchTrigger" || c.$$kind === "watchFiles")
		.map((c) => objectFrom(c, {}));
	return omitEmpty({ watch: watch.length ? watch : undefined });
}

function objectFrom(node: Node, remap: Record<string, string>): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	for (const [k, v] of Object.entries(node.props)) {
		if (v === undefined || k === "children") {
			continue;
		}
		const key = remap[k] ?? camelToSnake(k);
		out[key] = normalizeScalar(k, v);
	}
	return out;
}

function putKV(dst: Record<string, unknown>, node: Node, what: string): void {
	const name = str(node.props.name);
	if (!name) {
		throw new Error(`containerkit/tsx: <${node.$$kind}> in ${what} requires name`);
	}
	dst[name] = node.props.value ?? "";
}

function resourceName(node: Node, fallback: string): string {
	const n = str(node.props.name);
	if (n) {
		return n;
	}
	const image = str(node.props.image);
	if (image) {
		const base = image.split("/").pop() ?? image;
		return base.split(":")[0] || fallback;
	}
	const source = str(node.props.source);
	if (source) {
		return source;
	}
	const path = str(node.props.path);
	if (path) {
		return path;
	}
	return fallback;
}

function assertChildren(parent: Node): void {
	const flat = flatten(parent.children);
	if (parent.$$kind !== "fragment") {
		checkChildren(parent.$$kind, flat, parent.props, true);
	}
	for (const child of flat) {
		assertChildren(child);
	}
}

function throwInvalid(parent: Node, child: Node): never {
	throw new Error(`containerkit/tsx: <${child.$$kind}> is not valid inside <${parent.$$kind}>`);
}

function str(v: unknown): string | undefined {
	return typeof v === "string" && v !== "" ? v : undefined;
}

function asStringList(v: unknown): string[] {
	if (typeof v === "string" && v !== "") {
		return [v];
	}
	if (Array.isArray(v)) {
		return v.filter((x): x is string => typeof x === "string");
	}
	return [];
}

function duration(v: unknown): string | undefined {
	if (v === undefined || v === null) {
		return undefined;
	}
	if (typeof v === "string") {
		return v;
	}
	if (typeof v === "number") {
		if (v === 0) {
			return "0s";
		}
		if (Number.isInteger(v) && Math.abs(v) >= 1e6) {
			if (v % 1e9 === 0) {
				return `${v / 1e9}s`;
			}
			if (v % 1e6 === 0) {
				return `${v / 1e6}ms`;
			}
			return `${v}ns`;
		}
		return `${v}s`;
	}
	return undefined;
}

function normalizeScalar(_k: string, v: unknown): unknown {
	return v;
}

function camelToSnake(s: string): string {
	return s.replace(/[A-Z]/g, (ch) => `_${ch.toLowerCase()}`);
}

function omitEmpty(obj: Record<string, unknown>): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	for (const [k, v] of Object.entries(obj)) {
		if (v === undefined || v === null) {
			continue;
		}
		if (Array.isArray(v) && v.length === 0) {
			continue;
		}
		if (isPlain(v) && Object.keys(v).length === 0) {
			continue;
		}
		out[k] = v;
	}
	return out;
}

function isPlain(v: unknown): v is Record<string, unknown> {
	return v !== null && typeof v === "object" && !Array.isArray(v);
}
