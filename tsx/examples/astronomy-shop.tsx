import { writeFileSync } from "node:fs";
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
	type Child,
} from "containerkit";
import { Dockerfile, From, Label, Run, exportDockerfile } from "containerkit/dockerfile";
import { isCKNotFoundError, root, runIfMain } from "containerkit/live";
import { depend, netDriver, proto, restartPolicy, vol, waitFor } from "containerkit/vocab";

// OpenTelemetry Astronomy Shop, the same project as examples/astronomy_shop:
// every service, its environment, healthcheck, bind mount, and the kit-probe
// that curls the storefront after the stack is up.
//
// ASTRONOMY_SHOP_PROFILE selects shop, full, and observability. The default is
// shop,observability. Volumes are the files in examples/astronomy_shop/embed.
//
//   cd tsx && npm run build && ./node_modules/.bin/tsc -p examples/tsconfig.json
//   CONTAINERKIT_BIN=../dist/containerkit node dist-examples/astronomy-shop.js

const here = dirname(fileURLToPath(import.meta.url));
const embed = resolve(here, "../../examples/astronomy_shop/embed");
const probeImage = "astronomy-shop-kit-probe:local";
const networkName = "astronomy-shop";

const shop = "shop";
const full = "full";
const observability = "observability";

const imageName = "ghcr.io/open-telemetry/demo";
const imageVersion = "3.1.0";

const cfg = {
	collectorImage:
		"ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib:0.160.0",
	flagdImage: "ghcr.io/open-feature/flagd:v0.16.0",
	postgresImage: "postgres:18.6",
	valkeyImage: "ghcr.io/valkey-io/valkey:9.1.2-alpine3.24",
	grafanaImage: "grafana/grafana:13.2.1",
	jaegerImage: "quay.io/jaegertracing/jaeger:2.21.0",
	prometheusImage: "quay.io/prometheus/prometheus:v3.13.3",
	envPlatform: "local",
	adPort: "9555",
	adPrometheusPort: "9465",
	cartPort: "7070",
	checkoutPort: "5050",
	currencyPort: "7001",
	emailPort: "6060",
	paymentPort: "50051",
	productCatalogPort: "3550",
	quotePort: "8090",
	recommendationPort: "9001",
	shippingPort: "50050",
	frontendPort: "8080",
	envoyPort: "8080",
	envoyAdminPort: "10000",
	imageProviderPort: "8081",
	locustWebPort: "8089",
	flagdPort: "8013",
	flagdOFREPPort: "8016",
	flagdUIPort: "4000",
	valkeyPort: "6379",
	postgresPort: "5432",
	collectorGRPCPort: "4317",
	collectorHTTPPort: "4318",
	grafanaPort: "3000",
	jaegerUIPort: "16686",
	jaegerGRPCPort: "4317",
	prometheusPort: "9090",
	kafkaPort: "9092",
	postgresPassword: "changeit",
	postgresAstronomyPassword: "astronomy_password",
	postgresMonitoringPassword: "monitoring_password",
	otelServiceNamespace: "opentelemetry-demo",
	ipv6Enabled: "false",
};

function demo(name: string): string {
	return `${imageName}:${imageVersion}-${name}`;
}

function asset(path: string): string {
	return join(embed, path);
}

function parseProfiles(raw: string): string[] {
	if (raw.trim() === "") {
		return [shop, observability];
	}
	const seen = new Set<string>();
	const out: string[] = [];
	for (const part of raw.split(",")) {
		const profile = part.trim();
		if ((profile === shop || profile === full || profile === observability) && !seen.has(profile)) {
			seen.add(profile);
			out.push(profile);
		}
	}
	if (!seen.has(shop)) {
		out.unshift(shop);
	}
	return out.length === 0 ? [shop, observability] : out;
}

const profiles = parseProfiles(process.env.ASTRONOMY_SHOP_PROFILE ?? "");
const withObs = profiles.includes(observability);
const withFull = profiles.includes(full);

function resourceAttributes(criticality: string): string {
	const base = `service.namespace=${cfg.otelServiceNamespace},service.version=${imageVersion}`;
	return criticality === "" ? base : `${base},service.criticality=${criticality}`;
}

function otlpGRPC(): string {
	return `http://otel-collector:${cfg.collectorGRPCPort}`;
}

function otlpHTTP(): string {
	return `http://otel-collector:${cfg.collectorHTTPPort}`;
}

function commonOTEL(service: string, criticality: string, endpoint: string): Record<string, string> {
	return {
		OTEL_SERVICE_NAME: service,
		OTEL_RESOURCE_ATTRIBUTES: resourceAttributes(criticality),
		OTEL_EXPORTER_OTLP_ENDPOINT: endpoint,
		OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE: "cumulative",
		OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION: "explicit_bucket_histogram",
	};
}

function merge(...parts: Record<string, string>[]): Record<string, string> {
	return Object.assign({}, ...parts);
}

function Envs(props: { vars: Record<string, string> }) {
	return (
		<>
			{Object.entries(props.vars).map(([name, value]) => (
				<Env key={name} name={name} value={value} />
			))}
		</>
	);
}

function Ports(props: { ports: string[] }) {
	return (
		<>
			{props.ports.map((port) => (
				<Expose key={port} port={Number(port)} proto={proto.TCP} />
			))}
		</>
	);
}

function Bind(props: { source: string; target: string; readOnly?: boolean }) {
	return (
		<Volume
			kind={vol.Bind}
			source={props.source}
			target={props.target}
			readOnly={props.readOnly}
			seLinux="z"
		/>
	);
}

function Shell(props: { cmd: string; start: string; interval: string; timeout: string; retries: number }) {
	return (
		<Health
			test={["CMD-SHELL", props.cmd]}
			startPeriod={props.start}
			interval={props.interval}
			timeout={props.timeout}
			retries={props.retries}
		/>
	);
}

function Cmd(props: { argv: string[]; start: string; interval: string; timeout: string; retries: number }) {
	return (
		<Health
			test={["CMD", ...props.argv]}
			startPeriod={props.start}
			interval={props.interval}
			timeout={props.timeout}
			retries={props.retries}
		/>
	);
}

function Svc(props: {
	name: string;
	image: string;
	profile: string;
	command?: string[];
	user?: string;
	children?: Child;
}) {
	// composeUp has no profile flag. A selected profile is omitted from the
	// service so Compose starts it. An unselected profile stays on the service
	// and Compose skips it.
	const active = profiles.includes(props.profile);
	return (
		<Service
			name={props.name}
			image={props.image}
			profile={active ? undefined : props.profile}
			command={props.command}
			user={props.user}
		>
			<Restart policy={restartPolicy.UnlessStopped} />
			<Network name={networkName} />
			{props.children}
		</Service>
	);
}

function grpcHealth(port: string) {
	return <Cmd argv={["/bin/grpc_health_probe", `-addr=127.0.0.1:${port}`]} start="90s" interval="5s" timeout="10s" retries={30} />;
}

function bashTCP(port: string, start: string) {
	return <Shell cmd={`bash -c 'echo > /dev/tcp/127.0.0.1/${port}'`} start={start} interval="5s" timeout="5s" retries={20} />;
}

const probeDockerfile = (
	<Dockerfile>
		<From image="alpine" tag="3.20" />
		<Run cmd="apk add --no-cache curl" />
		<Label name="example" value="astronomy-shop-kit-probe" />
	</Dockerfile>
);

export function shopTree() {
	const collectorCmd = ["--config=/etc/otelcol-config.yml"];
	if (withObs) {
		collectorCmd.push("--config=/etc/otelcol-config-observability.yml");
	}
	collectorCmd.push("--config=/etc/otelcol-config-extras.yml");

	const checkoutEnv: Record<string, string> = {
		CHECKOUT_PORT: cfg.checkoutPort,
		FLAGD_HOST: "flagd",
		FLAGD_PORT: cfg.flagdPort,
		CART_ADDR: `cart:${cfg.cartPort}`,
		CURRENCY_ADDR: `currency:${cfg.currencyPort}`,
		EMAIL_ADDR: `http://email:${cfg.emailPort}`,
		PAYMENT_ADDR: `payment:${cfg.paymentPort}`,
		PRODUCT_CATALOG_ADDR: `product-catalog:${cfg.productCatalogPort}`,
		SHIPPING_ADDR: `http://shipping:${cfg.shippingPort}`,
		GOMEMLIMIT: "64MiB",
		OTEL_EXPORTER_OTLP_PROTOCOL: "http/protobuf",
		OTEL_GO_X_OBSERVABILITY: "true",
	};
	if (withFull) {
		checkoutEnv.KAFKA_ADDR = `kafka:${cfg.kafkaPort}`;
		checkoutEnv.KAFKA_TOPIC = "orders";
	}

	return (
		<Compose name="astronomy-shop">
			<Network name={networkName} driver={netDriver.Bridge} attachable enableIPv6={false} />
			<Svc
				name="astronomy-db"
				image={cfg.postgresImage}
				profile={shop}
				command={["postgres", "-c", "shared_preload_libraries=pg_stat_statements"]}
			>
				<Envs vars={{ POSTGRES_PASSWORD: cfg.postgresPassword }} />
				<Shell cmd="pg_isready -U astronomy_user -d astronomy_db" start="30s" interval="5s" timeout="10s" retries={20} />
				<Bind source={asset("postgresql/init.sql")} target="/docker-entrypoint-initdb.d/init.sql" readOnly />
			</Svc>
			<Svc name="valkey-cart" image={cfg.valkeyImage} profile={shop} user="valkey">
				<Shell cmd="valkey-cli ping" start="5s" interval="5s" timeout="3s" retries={10} />
			</Svc>
			<Svc name="otel-collector" image={cfg.collectorImage} profile={shop} user="0:0" command={collectorCmd}>
				<Envs
					vars={{
						OTEL_COLLECTOR_HOST: "0.0.0.0",
						OTEL_COLLECTOR_PORT_GRPC: cfg.collectorGRPCPort,
						OTEL_COLLECTOR_PORT_HTTP: cfg.collectorHTTPPort,
						FRONTEND_PROXY_ADDR: `frontend-proxy:${cfg.envoyPort}`,
						IMAGE_PROVIDER_HOST: "image-provider",
						IMAGE_PROVIDER_PORT: cfg.imageProviderPort,
						POSTGRES_HOST: "astronomy-db",
						POSTGRES_PORT: cfg.postgresPort,
						POSTGRES_MONITORING_PASSWORD: cfg.postgresMonitoringPassword,
						AD_PROMETHEUS_PORT: cfg.adPrometheusPort,
						GOMEMLIMIT: "160MiB",
					}}
				/>
				<Bind source={asset("otel-collector/otelcol-config.yml")} target="/etc/otelcol-config.yml" readOnly />
				<Bind source={asset("otel-collector/otelcol-config-extras.yml")} target="/etc/otelcol-config-extras.yml" readOnly />
				{withObs && (
					<Bind
						source={asset("otel-collector/otelcol-config-observability.yml")}
						target="/etc/otelcol-config-observability.yml"
						readOnly
					/>
				)}
				<Ports ports={[cfg.collectorGRPCPort, cfg.collectorHTTPPort]} />
			</Svc>
			<Svc name="flagd" image={cfg.flagdImage} profile={shop} command={["start", "--uri", "file:./etc/flagd/demo.flagd.json"]}>
				<Envs
					vars={merge(commonOTEL("flagd", "low", otlpHTTP()), {
						OTEL_EXPORTER_OTLP_PROTOCOL: "http/protobuf",
						GOMEMLIMIT: "60MiB",
					})}
				/>
				<Bind source={asset("flagd")} target="/etc/flagd" readOnly />
				<Ports ports={[cfg.flagdPort, cfg.flagdOFREPPort]} />
			</Svc>
			<Svc name="flagd-ui" image={demo("flagd-ui")} profile={shop}>
				<DependsOn name="otel-collector" />
				<DependsOn name="flagd" />
				<Envs
					vars={merge(commonOTEL("flagd-ui", "low", otlpHTTP()), {
						FLAGD_UI_PORT: cfg.flagdUIPort,
						SECRET_KEY_BASE: "yYrECL4qbNwleYInGJYvVnSkwJuSQJ4ijPTx5tirGUXrbznFIBFVJdPl5t6O9ASw",
						PHX_HOST: "localhost",
					})}
				/>
				{bashTCP(cfg.flagdUIPort, "10s")}
				<Bind source={asset("flagd")} target="/app/data" />
				<Ports ports={[cfg.flagdUIPort]} />
			</Svc>
			<Svc name="ad" image={demo("ad")} profile={shop}>
				<DependsOn name="otel-collector" />
				<DependsOn name="flagd" />
				<Envs
					vars={merge(commonOTEL("ad", "medium", otlpHTTP()), {
						AD_PORT: cfg.adPort,
						AD_PROMETHEUS_PORT: cfg.adPrometheusPort,
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						OTEL_LOGS_EXPORTER: "otlp",
						OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION: "latest",
					})}
				/>
				{bashTCP(cfg.adPort, "20s")}
				<Ports ports={[cfg.adPort]} />
			</Svc>
			<Svc name="cart" image={demo("cart")} profile={shop}>
				<DependsOn name="valkey-cart" condition={depend.Healthy} />
				<DependsOn name="otel-collector" />
				<DependsOn name="flagd" />
				<Envs
					vars={merge(commonOTEL("cart", "high", otlpGRPC()), {
						CART_PORT: cfg.cartPort,
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						VALKEY_ADDR: `valkey-cart:${cfg.valkeyPort}`,
						ASPNETCORE_HTTP_PORTS: cfg.cartPort,
					})}
				/>
				<Ports ports={[cfg.cartPort]} />
			</Svc>
			<Svc name="currency" image={demo("currency")} profile={shop}>
				<DependsOn name="otel-collector" />
				<Envs
					vars={merge(commonOTEL("currency", "high", otlpGRPC()), {
						CURRENCY_PORT: cfg.currencyPort,
						IPV6_ENABLED: cfg.ipv6Enabled,
						VERSION: imageVersion,
					})}
				/>
				<Shell cmd={`nc -z 127.0.0.1 ${cfg.currencyPort}`} start="10s" interval="5s" timeout="5s" retries={20} />
				<Ports ports={[cfg.currencyPort]} />
			</Svc>
			<Svc name="email" image={demo("email")} profile={shop}>
				<DependsOn name="otel-collector" />
				<Envs
					vars={merge(commonOTEL("email", "medium", otlpHTTP()), {
						APP_ENV: "production",
						EMAIL_PORT: cfg.emailPort,
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
					})}
				/>
				<Cmd
					argv={["ruby", "-e", `require 'socket'; TCPSocket.new('127.0.0.1', ${cfg.emailPort}).close`]}
					start="10s"
					interval="5s"
					timeout="5s"
					retries={20}
				/>
				<Ports ports={[cfg.emailPort]} />
			</Svc>
			<Svc name="payment" image={demo("payment")} profile={shop}>
				<DependsOn name="otel-collector" />
				<DependsOn name="flagd" />
				<Envs
					vars={merge(commonOTEL("payment", "critical", otlpGRPC()), {
						IPV6_ENABLED: cfg.ipv6Enabled,
						PAYMENT_PORT: cfg.paymentPort,
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						OTEL_EXPORTER_OTLP_PROTOCOL: "grpc",
					})}
				/>
				<Cmd
					argv={[
						"/nodejs/bin/node",
						"-e",
						`const net=require('net');const c=new net.Socket(); c.setTimeout(2000); c.connect(${cfg.paymentPort},'127.0.0.1',()=>{c.destroy();process.exit(0)}); c.on('error',()=>process.exit(1)); c.on('timeout',()=>process.exit(1))`,
					]}
					start="10s"
					interval="5s"
					timeout="5s"
					retries={20}
				/>
				<Ports ports={[cfg.paymentPort]} />
			</Svc>
			<Svc name="product-catalog" image={demo("product-catalog")} profile={shop}>
				<DependsOn name="otel-collector" />
				<DependsOn name="flagd" />
				<DependsOn name="astronomy-db" condition={depend.Healthy} />
				<Envs
					vars={merge(commonOTEL("product-catalog", "high", otlpGRPC()), {
						PRODUCT_CATALOG_PORT: cfg.productCatalogPort,
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						GOMEMLIMIT: "64MiB",
						OTEL_GO_X_OBSERVABILITY: "true",
						OTEL_CONFIG_FILE: "/otel-config.yml",
						OTEL_SEMCONV_STABILITY_OPT_IN: "database",
						DB_CONNECTION_STRING: `postgres://astronomy_user:${cfg.postgresAstronomyPassword}@astronomy-db/astronomy_db?sslmode=disable`,
					})}
				/>
				<Cmd
					argv={["/bin/grpc_health_probe", `-addr=127.0.0.1:${cfg.productCatalogPort}`]}
					start="90s"
					interval="5s"
					timeout="10s"
					retries={36}
				/>
				<Bind source={asset("otel-config.yml")} target="/otel-config.yml" readOnly />
				<Ports ports={[cfg.productCatalogPort]} />
			</Svc>
			<Svc name="quote" image={demo("quote")} profile={shop}>
				<DependsOn name="otel-collector" />
				<Envs
					vars={merge(commonOTEL("quote", "low", otlpHTTP()), {
						IPV6_ENABLED: cfg.ipv6Enabled,
						QUOTE_PORT: cfg.quotePort,
						OTEL_PHP_AUTOLOAD_ENABLED: "true",
						OTEL_PHP_INTERNAL_METRICS_ENABLED: "true",
					})}
				/>
				<Cmd
					argv={["php", "-r", `fsockopen('127.0.0.1', ${cfg.quotePort}) or die('fail');`]}
					start="10s"
					interval="5s"
					timeout="5s"
					retries={20}
				/>
				<Ports ports={[cfg.quotePort]} />
			</Svc>
			<Svc name="recommendation" image={demo("recommendation")} profile={shop}>
				<DependsOn name="product-catalog" />
				<DependsOn name="otel-collector" />
				<DependsOn name="flagd" />
				<Envs
					vars={merge(commonOTEL("recommendation", "medium", otlpHTTP()), {
						RECOMMENDATION_PORT: cfg.recommendationPort,
						PRODUCT_CATALOG_ADDR: `product-catalog:${cfg.productCatalogPort}`,
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION: "latest",
					})}
				/>
				<Shell cmd={`nc -z 127.0.0.1 ${cfg.recommendationPort}`} start="20s" interval="5s" timeout="5s" retries={20} />
				<Ports ports={[cfg.recommendationPort]} />
			</Svc>
			<Svc name="shipping" image={demo("shipping")} profile={shop}>
				<DependsOn name="otel-collector" />
				<DependsOn name="flagd" />
				<Envs
					vars={merge(commonOTEL("shipping", "high", otlpHTTP()), {
						IPV6_ENABLED: cfg.ipv6Enabled,
						SHIPPING_PORT: cfg.shippingPort,
						QUOTE_ADDR: `http://quote:${cfg.quotePort}`,
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						OTEL_EXPORTER_OTLP_PROTOCOL: "http/protobuf",
					})}
				/>
				<Cmd argv={["/app/healthcheck"]} start="10s" interval="5s" timeout="5s" retries={20} />
				<Ports ports={[cfg.shippingPort]} />
			</Svc>
			<Svc name="image-provider" image={demo("image-provider")} profile={shop}>
				<DependsOn name="otel-collector" />
				<Envs
					vars={{
						IMAGE_PROVIDER_PORT: cfg.imageProviderPort,
						OTEL_COLLECTOR_HOST: "otel-collector",
						OTEL_COLLECTOR_PORT_GRPC: cfg.collectorGRPCPort,
						OTEL_RESOURCE_ATTRIBUTES: resourceAttributes("low"),
						OTEL_SERVICE_NAME: "image-provider",
					}}
				/>
				<Cmd
					argv={["wget", "--quiet", "--tries=1", "--spider", `http://127.0.0.1:${cfg.imageProviderPort}/status`]}
					start="10s"
					interval="5s"
					timeout="5s"
					retries={20}
				/>
				<Ports ports={[cfg.imageProviderPort]} />
			</Svc>
			<Svc name="checkout" image={demo("checkout")} profile={shop}>
				<DependsOn name="cart" />
				<DependsOn name="currency" condition={depend.Healthy} />
				<DependsOn name="email" condition={depend.Healthy} />
				<DependsOn name="payment" condition={depend.Healthy} />
				<DependsOn name="product-catalog" />
				<DependsOn name="shipping" />
				<DependsOn name="otel-collector" />
				<DependsOn name="flagd" />
				{withFull && <DependsOn name="kafka" condition={depend.Healthy} />}
				<Envs vars={merge(commonOTEL("checkout", "critical", otlpHTTP()), checkoutEnv)} />
				{grpcHealth(cfg.checkoutPort)}
				<Ports ports={[cfg.checkoutPort]} />
			</Svc>
			<Svc name="frontend" image={demo("frontend")} profile={shop}>
				<DependsOn name="ad" condition={depend.Healthy} />
				<DependsOn name="cart" />
				<DependsOn name="checkout" />
				<DependsOn name="currency" condition={depend.Healthy} />
				<DependsOn name="product-catalog" />
				<DependsOn name="quote" condition={depend.Healthy} />
				<DependsOn name="recommendation" condition={depend.Healthy} />
				<DependsOn name="shipping" />
				<DependsOn name="otel-collector" />
				<DependsOn name="image-provider" condition={depend.Healthy} />
				<DependsOn name="flagd" />
				<Envs
					vars={merge(commonOTEL("frontend", "critical", otlpGRPC()), {
						PORT: cfg.frontendPort,
						FRONTEND_ADDR: `frontend:${cfg.frontendPort}`,
						AD_ADDR: `ad:${cfg.adPort}`,
						CART_ADDR: `cart:${cfg.cartPort}`,
						CHECKOUT_ADDR: `checkout:${cfg.checkoutPort}`,
						CURRENCY_ADDR: `currency:${cfg.currencyPort}`,
						PRODUCT_CATALOG_ADDR: `product-catalog:${cfg.productCatalogPort}`,
						RECOMMENDATION_ADDR: `recommendation:${cfg.recommendationPort}`,
						SHIPPING_ADDR: `http://shipping:${cfg.shippingPort}`,
						ENV_PLATFORM: cfg.envPlatform,
						PUBLIC_OTEL_EXPORTER_OTLP_TRACES_ENDPOINT: `http://localhost:${cfg.envoyPort}/otlp-http/v1/traces`,
						WEB_OTEL_SERVICE_NAME: "frontend-web",
						OTEL_COLLECTOR_HOST: "otel-collector",
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
					})}
				/>
				<Cmd
					argv={[
						"/nodejs/bin/node",
						"-e",
						`require('net').connect(${cfg.frontendPort},require('os').hostname(),function(){process.exit(0)}).on('error',function(){process.exit(1)})`,
					]}
					start="60s"
					interval="10s"
					timeout="10s"
					retries={5}
				/>
				<Ports ports={[cfg.frontendPort]} />
			</Svc>
			<Svc name="load-generator" image={demo("load-generator")} profile={shop}>
				<DependsOn name="frontend" condition={depend.Healthy} />
				<DependsOn name="flagd" />
				<Envs
					vars={merge(commonOTEL("load-generator", "low", otlpGRPC()), {
						LOCUST_WEB_PORT: cfg.locustWebPort,
						LOCUST_USERS: "5",
						LOCUST_HOST: `http://frontend-proxy:${cfg.envoyPort}`,
						LOCUST_HEADLESS: "false",
						LOCUST_AUTOSTART: "true",
						LOCUST_BROWSER_TRAFFIC_ENABLED: "true",
						LOCUST_HTTP_USER_WEIGHT: "9",
						LOCUST_BROWSER_USER_WEIGHT: "1",
						LOCUST_WEB_HOST: "0.0.0.0",
						OTEL_PYTHON_SDK_INTERNAL_METRICS_ENABLED: "true",
						PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION: "python",
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						FLAGD_OFREP_PORT: cfg.flagdOFREPPort,
					})}
				/>
				<Cmd
					argv={[
						"python3",
						"-c",
						`import socket; s=socket.create_connection(('127.0.0.1',${cfg.locustWebPort}),2); s.close()`,
					]}
					start="30s"
					interval="5s"
					timeout="5s"
					retries={20}
				/>
				<Publish protocol={proto.TCP} hostPort={Number(cfg.locustWebPort)} containerPort={Number(cfg.locustWebPort)} />
			</Svc>
			<Svc name="frontend-proxy" image={demo("frontend-proxy")} profile={shop}>
				<DependsOn name="frontend" condition={depend.Healthy} />
				<DependsOn name="load-generator" />
				<DependsOn name="flagd-ui" condition={depend.Healthy} />
				<Envs
					vars={{
						FRONTEND_PORT: cfg.frontendPort,
						FRONTEND_HOST: "frontend",
						LOCUST_WEB_HOST: "load-generator",
						LOCUST_WEB_PORT: cfg.locustWebPort,
						GRAFANA_PORT: cfg.grafanaPort,
						GRAFANA_HOST: "grafana",
						JAEGER_UI_PORT: cfg.jaegerUIPort,
						JAEGER_HOST: "jaeger",
						OPAMP_SERVER_HOST: "opamp-server",
						OPAMP_SERVER_UI_PORT: "4321",
						OTEL_COLLECTOR_HOST: "otel-collector",
						IMAGE_PROVIDER_HOST: "image-provider",
						IMAGE_PROVIDER_PORT: cfg.imageProviderPort,
						OTEL_COLLECTOR_PORT_GRPC: cfg.collectorGRPCPort,
						OTEL_COLLECTOR_PORT_HTTP: cfg.collectorHTTPPort,
						OTEL_RESOURCE_ATTRIBUTES: resourceAttributes("critical"),
						OTEL_SERVICE_NAME: "frontend-proxy",
						OTEL_SERVICE_NAMESPACE: cfg.otelServiceNamespace,
						ENVOY_PORT: cfg.envoyPort,
						ENVOY_ADDR: "0.0.0.0",
						ENVOY_ADMIN_PORT: cfg.envoyAdminPort,
						FIREPIT_HOST: "firepit",
						FIREPIT_PORT: "8080",
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						FLAGD_UI_HOST: "flagd-ui",
						FLAGD_UI_PORT: cfg.flagdUIPort,
						TELEMETRY_DOCS_HOST: "telemetry-docs",
						TELEMETRY_DOCS_PORT: "8000",
						CHATBOT_HOST: "chatbot",
						CHATBOT_PORT: "7860",
					}}
				/>
				<Cmd argv={["bash", "-c", `echo > /dev/tcp/127.0.0.1/${cfg.envoyPort}`]} start="10s" interval="5s" timeout="5s" retries={20} />
				<Publish protocol={proto.TCP} hostPort={Number(cfg.envoyPort)} containerPort={Number(cfg.envoyPort)} />
			</Svc>
			<Svc name="kafka" image={demo("kafka")} profile={full}>
				<Envs
					vars={merge(commonOTEL("kafka", "low", otlpHTTP()), {
						KAFKA_ADVERTISED_LISTENERS: "PLAINTEXT://kafka:9092",
						KAFKA_LISTENERS: "PLAINTEXT://kafka:9092,CONTROLLER://kafka:9093",
						KAFKA_CONTROLLER_QUORUM_VOTERS: "1@kafka:9093",
						KAFKA_HEAP_OPTS: "-Xmx400m -Xms400m",
						OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION: "latest",
					})}
				/>
				<Shell cmd="nc -z kafka 9092" start="10s" interval="5s" timeout="10s" retries={10} />
				<Ports ports={[cfg.kafkaPort]} />
			</Svc>
			<Svc name="accounting" image={demo("accounting")} profile={full}>
				<DependsOn name="otel-collector" />
				<DependsOn name="kafka" condition={depend.Healthy} />
				<Envs
					vars={merge(commonOTEL("accounting", "low", otlpHTTP()), {
						KAFKA_ADDR: `kafka:${cfg.kafkaPort}`,
						KAFKA_TOPIC: "orders",
						DB_CONNECTION_STRING: `Host=astronomy-db;Username=astronomy_user;Password=${cfg.postgresAstronomyPassword};Database=astronomy_db`,
						OTEL_DOTNET_AUTO_TRACES_ENTITYFRAMEWORKCORE_INSTRUMENTATION_ENABLED: "false",
					})}
				/>
			</Svc>
			<Svc name="fraud-detection" image={demo("fraud-detection")} profile={full}>
				<DependsOn name="otel-collector" />
				<DependsOn name="kafka" condition={depend.Healthy} />
				<Envs
					vars={merge(commonOTEL("fraud-detection", "low", otlpHTTP()), {
						FLAGD_HOST: "flagd",
						FLAGD_PORT: cfg.flagdPort,
						KAFKA_ADDR: `kafka:${cfg.kafkaPort}`,
						KAFKA_TOPIC: "orders",
						OTEL_INSTRUMENTATION_KAFKA_EXPERIMENTAL_SPAN_ATTRIBUTES: "true",
						OTEL_INSTRUMENTATION_MESSAGING_EXPERIMENTAL_RECEIVE_TELEMETRY_ENABLED: "true",
						OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION: "latest",
					})}
				/>
			</Svc>
			<Svc name="jaeger" image={cfg.jaegerImage} profile={observability} command={["--config=file:/etc/jaeger/config.yml"]}>
				<Envs
					vars={{
						JAEGER_HOST: "0.0.0.0",
						JAEGER_GRPC_PORT: cfg.jaegerGRPCPort,
						PROMETHEUS_ADDR: `prometheus:${cfg.prometheusPort}`,
						OTEL_COLLECTOR_HOST: "otel-collector",
						OTEL_COLLECTOR_PORT_HTTP: cfg.collectorHTTPPort,
						MEMORY_MAX_TRACES: "25000",
					}}
				/>
				<Bind source={asset("jaeger/config.yml")} target="/etc/jaeger/config.yml" readOnly />
				<Bind source={asset("jaeger/ui-config.json")} target="/etc/jaeger/ui-config.json" readOnly />
				<Publish protocol={proto.TCP} hostPort={Number(cfg.jaegerUIPort)} containerPort={Number(cfg.jaegerUIPort)} />
			</Svc>
			<Svc
				name="prometheus"
				image={cfg.prometheusImage}
				profile={observability}
				command={[
					"--web.console.templates=/etc/prometheus/consoles",
					"--web.console.libraries=/etc/prometheus/console_libraries",
					"--storage.tsdb.retention.time=7d",
					"--config.file=/etc/prometheus/prometheus-config.yaml",
					"--storage.tsdb.path=/prometheus",
					"--web.enable-lifecycle",
					"--web.route-prefix=/",
					"--web.enable-otlp-receiver",
					"--enable-feature=exemplar-storage",
				]}
			>
				<Bind source={asset("prometheus/prometheus-config.yaml")} target="/etc/prometheus/prometheus-config.yaml" readOnly />
				<Publish protocol={proto.TCP} hostPort={Number(cfg.prometheusPort)} containerPort={Number(cfg.prometheusPort)} />
			</Svc>
			<Svc name="grafana" image={cfg.grafanaImage} profile={observability}>
				<DependsOn name="prometheus" />
				<DependsOn name="jaeger" />
				<Bind source={asset("grafana/grafana.ini")} target="/etc/grafana/grafana.ini" readOnly />
				<Bind source={asset("grafana/provisioning")} target="/etc/grafana/provisioning" readOnly />
				<Publish protocol={proto.TCP} hostPort={Number(cfg.grafanaPort)} containerPort={Number(cfg.grafanaPort)} />
			</Svc>
		</Compose>
	);
}

export const tree = shopTree();

function printURLs() {
	console.error("");
	console.error("Astronomy Shop is up:");
	console.error(`  Store:          http://127.0.0.1:${cfg.envoyPort}/`);
	console.error(`  Load generator: http://127.0.0.1:${cfg.locustWebPort}/`);
	if (withObs) {
		console.error(`  Grafana:        http://127.0.0.1:${cfg.envoyPort}/grafana/`);
		console.error(`  Jaeger:         http://127.0.0.1:${cfg.jaegerUIPort}/jaeger/ui`);
		console.error(`  Prometheus:     http://127.0.0.1:${cfg.prometheusPort}/`);
	}
	console.error("");
}

const probeScript = `curl -fsS --max-time 20 http://frontend-proxy:${cfg.envoyPort}/ >/dev/null || curl -fsS --max-time 20 http://host.docker.internal:${cfg.envoyPort}/ >/dev/null && echo OK`;

const live = root({ fromEnv: true, onExit: "down" }, async function App(session) {
	const exportPath = "astronomy-shop.compose.yaml";
	writeFileSync(exportPath, exportYAML(tree));
	console.error(`exported compose project to ${exportPath}`);
	let exists = true;
	try {
		await session.inspectImage(probeImage);
	} catch (err) {
		if (!isCKNotFoundError(err)) {
			throw err;
		}
		exists = false;
	}
	if (!exists) {
		await session.buildImage({
			image: probeImage,
			parseDockerfile: exportDockerfile(probeDockerfile),
		});
	}
	await session.composeUp(tree, {
		detach: true,
		wait: true,
		timeout: "900s",
		removeOrphans: true,
	});
	let probeError: unknown;
	for (let attempt = 1; attempt <= 5; attempt++) {
		try {
			await session.run({
				name: "kit-probe",
				image: probeImage,
				command: ["sh", "-c", probeScript],
				network: { name: networkName },
				extraHost: { hostname: "host.docker.internal", ip: "host-gateway" },
			});
			await session.wait("kit-probe", { waitFor: waitFor.NextExit });
			probeError = undefined;
			break;
		} catch (err) {
			probeError = err;
			try {
				await session.remove("kit-probe", { force: true });
			} catch {
				// The probe never got a container to remove.
			}
		}
	}
	if (probeError) {
		throw probeError;
	}
	printURLs();
	return tree;
});

export default live;

await runIfMain(live, import.meta.url);
