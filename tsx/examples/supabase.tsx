import { existsSync, readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
	Compose,
	DependsOn,
	Env,
	Health,
	Network,
	Publish,
	Restart,
	Service,
	Volume,
	exportYAML,
	type Child,
} from "containerkit";
import { depend, netDriver, proto, restartPolicy, vol } from "containerkit/vocab";

type Gateway = "envoy" | "kong" | "caddy" | "nginx";
type Postgres = "15" | "17";
type Pooler = "supavisor" | "pgbouncer";
type Storage = "file" | "s3" | "rustfs";

type StackConfig = {
	dir: string;
	env: Record<string, string>;
	gateway: Gateway;
	postgres: Postgres;
	pooler: Pooler;
	storage: Storage;
	logs: boolean;
};

const sqlMounts = [
	["volumes/db/realtime.sql", "/docker-entrypoint-initdb.d/migrations/99-realtime.sql"],
	["volumes/db/webhooks.sql", "/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql"],
	["volumes/db/roles.sql", "/docker-entrypoint-initdb.d/init-scripts/99-roles.sql"],
	["volumes/db/jwt.sql", "/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql"],
	["volumes/db/data", "/var/lib/postgresql/data"],
	["volumes/db/_supabase.sql", "/docker-entrypoint-initdb.d/migrations/97-_supabase.sql"],
	["volumes/db/logs.sql", "/docker-entrypoint-initdb.d/migrations/99-logs.sql"],
	["volumes/db/pooler.sql", "/docker-entrypoint-initdb.d/migrations/99-pooler.sql"],
] as const;

function Supabase(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Compose name="supabase">
			<Network name="default" driver={netDriver.Bridge} />
			<ProjectVolumes cfg={cfg} />
			<Database cfg={cfg} />
			<Auth cfg={cfg} />
			<Rest cfg={cfg} />
			<Realtime cfg={cfg} />
			<Imgproxy cfg={cfg} />
			<Meta cfg={cfg} />
			<StorageService cfg={cfg} />
			<Studio cfg={cfg} />
			<Gateway cfg={cfg} />
			<Functions cfg={cfg} />
			{cfg.storage === "s3" && (
				<>
					<Minio cfg={cfg} />
					<MinioBucket cfg={cfg} />
				</>
			)}
			{cfg.storage === "rustfs" && (
				<>
					<Rustfs cfg={cfg} />
					<RustfsBucket cfg={cfg} />
				</>
			)}
			{cfg.pooler === "supavisor" ? <Supavisor cfg={cfg} /> : <PgBouncer cfg={cfg} />}
			{cfg.gateway === "caddy" && <Caddy cfg={cfg} />}
			{cfg.gateway === "nginx" && <Nginx cfg={cfg} />}
			{cfg.logs && (
				<>
					<Analytics cfg={cfg} />
					<Vector cfg={cfg} />
				</>
			)}
		</Compose>
	);
}

function ProjectVolumes(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<>
			<Volume name="db-config" />
			<Volume name="deno-cache" />
			{cfg.gateway === "caddy" && (
				<>
					<Volume name="caddy_data" />
					<Volume name="caddy_config" />
				</>
			)}
			{cfg.gateway === "nginx" && <Volume name="nginx_letsencrypt" />}
			{cfg.storage === "s3" && <Volume name="minio-data" />}
			{cfg.storage === "rustfs" && <Volume name="rustfs-data" />}
		</>
	);
}

function Database(props: { cfg: StackConfig }) {
	const { cfg } = props;
	const port = or(cfg, "POSTGRES_PORT", "5432");
	const db = or(cfg, "POSTGRES_DB", "postgres");
	return (
		<Keep
			name="db"
			container="supabase-db"
			image={cfg.postgres === "15" ? "supabase/postgres:15.8.1.085" : "supabase/postgres:17.6.1.136"}
			command={["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf", "-c", "log_min_messages=fatal"]}
		>
			<Probe test={["CMD", "pg_isready", "-U", "postgres", "-h", "localhost"]} interval="5s" timeout="5s" retries={10} />
			{sqlMounts.map(([src, target]) => (
				<Bind key={target} src={file(cfg, src)} target={target} />
			))}
			<Volume name="db-config" target="/etc/postgresql-custom" />
			<EnvMap
				values={{
					POSTGRES_HOST: "/var/run/postgresql",
					PGPORT: port,
					POSTGRES_PORT: port,
					PGPASSWORD: e(cfg, "POSTGRES_PASSWORD"),
					POSTGRES_PASSWORD: e(cfg, "POSTGRES_PASSWORD"),
					PGDATABASE: db,
					POSTGRES_DB: db,
					JWT_EXP: or(cfg, "JWT_EXPIRY", "3600"),
				}}
			/>
		</Keep>
	);
}

function Auth(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep name="auth" container="supabase-auth" image="supabase/gotrue:v2.197.0">
			<DependsOn name="db" condition={depend.Healthy} />
			<Probe
				test={["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9999/health"]}
				interval="5s"
				timeout="5s"
				retries={3}
			/>
			<EnvMap
				values={{
					GOTRUE_API_HOST: "0.0.0.0",
					GOTRUE_API_PORT: "9999",
					API_EXTERNAL_URL: e(cfg, "API_EXTERNAL_URL"),
					GOTRUE_DB_DRIVER: "postgres",
					GOTRUE_DB_DATABASE_URL: pg(cfg, "supabase_auth_admin", or(cfg, "POSTGRES_DB", "postgres")),
					GOTRUE_SITE_URL: e(cfg, "SITE_URL"),
					GOTRUE_URI_ALLOW_LIST: e(cfg, "ADDITIONAL_REDIRECT_URLS"),
					GOTRUE_DISABLE_SIGNUP: or(cfg, "DISABLE_SIGNUP", "false"),
					GOTRUE_JWT_ADMIN_ROLES: "service_role",
					GOTRUE_JWT_AUD: "authenticated",
					GOTRUE_JWT_DEFAULT_GROUP_NAME: "authenticated",
					GOTRUE_JWT_EXP: or(cfg, "JWT_EXPIRY", "3600"),
					GOTRUE_JWT_SECRET: e(cfg, "JWT_SECRET"),
					GOTRUE_JWT_ISSUER: e(cfg, "API_EXTERNAL_URL"),
					GOTRUE_EXTERNAL_EMAIL_ENABLED: or(cfg, "ENABLE_EMAIL_SIGNUP", "true"),
					GOTRUE_EXTERNAL_ANONYMOUS_USERS_ENABLED: or(cfg, "ENABLE_ANONYMOUS_USERS", "false"),
					GOTRUE_MAILER_AUTOCONFIRM: or(cfg, "ENABLE_EMAIL_AUTOCONFIRM", "false"),
					GOTRUE_SMTP_ADMIN_EMAIL: e(cfg, "SMTP_ADMIN_EMAIL"),
					GOTRUE_SMTP_HOST: e(cfg, "SMTP_HOST"),
					GOTRUE_SMTP_PORT: e(cfg, "SMTP_PORT"),
					GOTRUE_SMTP_USER: e(cfg, "SMTP_USER"),
					GOTRUE_SMTP_PASS: e(cfg, "SMTP_PASS"),
					GOTRUE_SMTP_SENDER_NAME: e(cfg, "SMTP_SENDER_NAME"),
					GOTRUE_MAILER_URLPATHS_INVITE: e(cfg, "MAILER_URLPATHS_INVITE"),
					GOTRUE_MAILER_URLPATHS_CONFIRMATION: e(cfg, "MAILER_URLPATHS_CONFIRMATION"),
					GOTRUE_MAILER_URLPATHS_RECOVERY: e(cfg, "MAILER_URLPATHS_RECOVERY"),
					GOTRUE_MAILER_URLPATHS_EMAIL_CHANGE: e(cfg, "MAILER_URLPATHS_EMAIL_CHANGE"),
					GOTRUE_EXTERNAL_PHONE_ENABLED: or(cfg, "ENABLE_PHONE_SIGNUP", "true"),
					GOTRUE_SMS_AUTOCONFIRM: or(cfg, "ENABLE_PHONE_AUTOCONFIRM", "true"),
				}}
			/>
			<EnvIf cfg={cfg} from="JWT_KEYS" name="GOTRUE_JWT_KEYS" />
		</Keep>
	);
}

function Rest(props: { cfg: StackConfig }) {
	const { cfg } = props;
	const secret = e(cfg, "JWT_JWKS") || e(cfg, "JWT_SECRET");
	return (
		<Keep name="rest" container="supabase-rest" image="postgrest/postgrest:v14.18" command={["postgrest"]}>
			<DependsOn name="db" condition={depend.Healthy} />
			<Probe test={["CMD", "postgrest", "--ready"]} interval="5s" timeout="5s" retries={3} />
			<EnvMap
				values={{
					PGRST_DB_URI: pg(cfg, "authenticator", or(cfg, "POSTGRES_DB", "postgres")),
					PGRST_DB_SCHEMAS: or(cfg, "PGRST_DB_SCHEMAS", "public,graphql_public"),
					PGRST_DB_MAX_ROWS: or(cfg, "PGRST_DB_MAX_ROWS", "1000"),
					PGRST_DB_EXTRA_SEARCH_PATH: or(cfg, "PGRST_DB_EXTRA_SEARCH_PATH", "public"),
					PGRST_DB_ANON_ROLE: "anon",
					PGRST_ADMIN_SERVER_PORT: "3001",
					PGRST_ADMIN_SERVER_HOST: "localhost",
					PGRST_JWT_SECRET: secret,
					PGRST_DB_USE_LEGACY_GUCS: "false",
					PGRST_APP_SETTINGS_JWT_EXP: or(cfg, "JWT_EXPIRY", "3600"),
				}}
			/>
		</Keep>
	);
}

function Realtime(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep name="realtime" container="realtime-dev.supabase-realtime" image="supabase/realtime:v2.148.3">
			<DependsOn name="db" condition={depend.Healthy} />
			<Probe
				test={[
					"CMD-SHELL",
					`curl -sSfL --head -o /dev/null -H "Authorization: Bearer ${e(cfg, "ANON_KEY")}" http://localhost:4000/api/tenants/realtime-dev/health`,
				]}
				interval="30s"
				timeout="5s"
				retries={3}
				start="10s"
			/>
			<EnvMap
				values={{
					PORT: "4000",
					DB_HOST: or(cfg, "POSTGRES_HOST", "db"),
					DB_PORT: or(cfg, "POSTGRES_PORT", "5432"),
					DB_USER: "supabase_admin",
					DB_PASSWORD: e(cfg, "POSTGRES_PASSWORD"),
					DB_NAME: or(cfg, "POSTGRES_DB", "postgres"),
					DB_AFTER_CONNECT_QUERY: "SET search_path TO _realtime",
					DB_ENC_KEY: or(cfg, "REALTIME_DB_ENC_KEY", "supabaserealtime"),
					API_JWT_SECRET: e(cfg, "JWT_SECRET"),
					SECRET_KEY_BASE: e(cfg, "SECRET_KEY_BASE"),
					METRICS_JWT_SECRET: e(cfg, "JWT_SECRET"),
					ERL_AFLAGS: "-proto_dist inet_tcp",
					DNS_NODES: "''",
					RLIMIT_NOFILE: "10000",
					APP_NAME: "realtime",
					SEED_SELF_HOST: "true",
					RUN_JANITOR: "true",
					DISABLE_HEALTHCHECK_LOGGING: "true",
				}}
			/>
			<EnvIf cfg={cfg} from="JWT_JWKS" name="API_JWT_JWKS" />
		</Keep>
	);
}

function StorageService(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep name="storage" container="supabase-storage" image="supabase/storage-api:v1.89.12">
			<DependsOn name="db" condition={depend.Healthy} />
			<DependsOn name="rest" condition={depend.Started} />
			<DependsOn name="imgproxy" condition={depend.Started} />
			<Probe
				test={["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://storage:5000/status"]}
				interval="5s"
				timeout="5s"
				retries={3}
				start="10s"
			/>
			<Bind src={file(cfg, "volumes/storage")} target="/var/lib/storage" />
			<EnvMap
				values={{
					ANON_KEY: e(cfg, "ANON_KEY"),
					SERVICE_KEY: e(cfg, "SERVICE_ROLE_KEY"),
					POSTGREST_URL: "http://rest:3000",
					AUTH_JWT_SECRET: e(cfg, "JWT_SECRET"),
					DATABASE_URL: pg(cfg, "supabase_storage_admin", or(cfg, "POSTGRES_DB", "postgres")),
					STORAGE_PUBLIC_URL: e(cfg, "SUPABASE_PUBLIC_URL"),
					REQUEST_ALLOW_X_FORWARDED_PATH: "true",
					FILE_SIZE_LIMIT: "52428800",
					GLOBAL_S3_BUCKET: or(cfg, "GLOBAL_S3_BUCKET", "stub"),
					FILE_STORAGE_BACKEND_PATH: "/var/lib/storage",
					TENANT_ID: or(cfg, "STORAGE_TENANT_ID", "stub"),
					REGION: or(cfg, "REGION", "stub"),
					ENABLE_IMAGE_TRANSFORMATION: "true",
					IMGPROXY_URL: "http://imgproxy:5001",
					S3_PROTOCOL_ACCESS_KEY_ID: e(cfg, "S3_PROTOCOL_ACCESS_KEY_ID"),
					S3_PROTOCOL_ACCESS_KEY_SECRET: e(cfg, "S3_PROTOCOL_ACCESS_KEY_SECRET"),
				}}
			/>
			<EnvIf cfg={cfg} from="JWT_JWKS" name="JWT_JWKS" />
			{cfg.storage === "s3" && (
				<>
					<DependsOn name="minio" condition={depend.Healthy} />
					<DependsOn name="minio-createbucket" condition={depend.CompletedSuccessfully} />
					<EnvMap
						values={{
							STORAGE_BACKEND: "s3",
							GLOBAL_S3_ENDPOINT: "http://minio:9000",
							GLOBAL_S3_PROTOCOL: "http",
							GLOBAL_S3_FORCE_PATH_STYLE: "true",
							AWS_ACCESS_KEY_ID: e(cfg, "MINIO_ROOT_USER"),
							AWS_SECRET_ACCESS_KEY: e(cfg, "MINIO_ROOT_PASSWORD"),
						}}
					/>
				</>
			)}
			{cfg.storage === "rustfs" && (
				<>
					<DependsOn name="rustfs" condition={depend.Healthy} />
					<DependsOn name="rustfs-createbucket" condition={depend.CompletedSuccessfully} />
					<EnvMap
						values={{
							STORAGE_BACKEND: "s3",
							GLOBAL_S3_ENDPOINT: "http://rustfs:9000",
							GLOBAL_S3_PROTOCOL: "http",
							GLOBAL_S3_FORCE_PATH_STYLE: "true",
							AWS_ACCESS_KEY_ID: e(cfg, "MINIO_ROOT_USER"),
							AWS_SECRET_ACCESS_KEY: e(cfg, "MINIO_ROOT_PASSWORD"),
						}}
					/>
				</>
			)}
			{cfg.storage === "file" && <Env name="STORAGE_BACKEND" value="file" />}
		</Keep>
	);
}

function Imgproxy(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep name="imgproxy" container="supabase-imgproxy" image="darthsim/imgproxy:v4.0.21">
			<Probe test={["CMD", "imgproxy", "health"]} interval="5s" timeout="5s" retries={3} />
			<Bind src={file(cfg, "volumes/storage")} target="/var/lib/storage" />
			<EnvMap
				values={{
					IMGPROXY_BIND: ":5001",
					IMGPROXY_LOCAL_FILESYSTEM_ROOT: "/",
					IMGPROXY_USE_ETAG: "true",
					IMGPROXY_AUTO_WEBP: or(cfg, "IMGPROXY_AUTO_WEBP", "true"),
					IMGPROXY_MAX_SRC_RESOLUTION: "16.8",
				}}
			/>
		</Keep>
	);
}

function Meta(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep name="meta" container="supabase-meta" image="supabase/postgres-meta:v0.100.0">
			<DependsOn name="db" condition={depend.Healthy} />
			<EnvMap
				values={{
					PG_META_PORT: "8080",
					PG_META_DB_HOST: or(cfg, "POSTGRES_HOST", "db"),
					PG_META_DB_PORT: or(cfg, "POSTGRES_PORT", "5432"),
					PG_META_DB_NAME: or(cfg, "POSTGRES_DB", "postgres"),
					PG_META_DB_USER: "postgres",
					PG_META_DB_PASSWORD: e(cfg, "POSTGRES_PASSWORD"),
					CRYPTO_KEY: e(cfg, "PG_META_CRYPTO_KEY"),
				}}
			/>
		</Keep>
	);
}

function Studio(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep name="studio" container="supabase-studio" image="supabase/studio:2026.11.02-sha-a05e7ec">
			<Probe
				test={[
					"CMD-SHELL",
					`node -e "fetch('http://localhost:3000/api/platform/profile').then((r) => {if (r.status !== 200) throw new Error(r.status)})"`,
				]}
				interval="5s"
				timeout="10s"
				retries={3}
				start="20s"
			/>
			<Bind src={file(cfg, "volumes/snippets")} target="/app/snippets" />
			<Bind src={file(cfg, "volumes/functions")} target="/app/edge-functions" readOnly />
			<EnvMap
				values={{
					HOSTNAME: "0.0.0.0",
					STUDIO_PG_META_URL: "http://meta:8080",
					POSTGRES_PORT: or(cfg, "POSTGRES_PORT", "5432"),
					POSTGRES_HOST: or(cfg, "POSTGRES_HOST", "db"),
					POSTGRES_DB: or(cfg, "POSTGRES_DB", "postgres"),
					POSTGRES_PASSWORD: e(cfg, "POSTGRES_PASSWORD"),
					POSTGRES_USER_READ_WRITE: "postgres",
					PG_META_CRYPTO_KEY: e(cfg, "PG_META_CRYPTO_KEY"),
					PGRST_DB_SCHEMAS: or(cfg, "PGRST_DB_SCHEMAS", "public,graphql_public"),
					PGRST_DB_MAX_ROWS: or(cfg, "PGRST_DB_MAX_ROWS", "1000"),
					PGRST_DB_EXTRA_SEARCH_PATH: or(cfg, "PGRST_DB_EXTRA_SEARCH_PATH", "public"),
					DEFAULT_ORGANIZATION_NAME: or(cfg, "STUDIO_DEFAULT_ORGANIZATION", "Default Organization"),
					DEFAULT_PROJECT_NAME: or(cfg, "STUDIO_DEFAULT_PROJECT", "Default Project"),
					OPENAI_API_KEY: e(cfg, "OPENAI_API_KEY"),
					SUPABASE_URL: "http://api-gw:8000",
					SUPABASE_PUBLIC_URL: e(cfg, "SUPABASE_PUBLIC_URL"),
					SUPABASE_ANON_KEY: e(cfg, "ANON_KEY"),
					SUPABASE_SERVICE_KEY: e(cfg, "SERVICE_ROLE_KEY"),
					AUTH_JWT_SECRET: e(cfg, "JWT_SECRET"),
					SUPABASE_PUBLISHABLE_KEY: e(cfg, "SUPABASE_PUBLISHABLE_KEY"),
					SUPABASE_SECRET_KEY: e(cfg, "SUPABASE_SECRET_KEY"),
					SNIPPETS_MANAGEMENT_FOLDER: "/app/snippets",
					EDGE_FUNCTIONS_MANAGEMENT_FOLDER: "/app/edge-functions",
				}}
			/>
			{cfg.logs ? (
				<>
					<DependsOn name="analytics" condition={depend.Healthy} />
					<Env name="LOGFLARE_PRIVATE_ACCESS_TOKEN" value={e(cfg, "LOGFLARE_PRIVATE_ACCESS_TOKEN")} />
					<Env name="LOGFLARE_URL" value="http://analytics:4000" />
					<Env name="ENABLED_FEATURES_LOGS_ALL" value="true" />
				</>
			) : (
				<Env name="ENABLED_FEATURES_LOGS_ALL" value="false" />
			)}
		</Keep>
	);
}

function Gateway(props: { cfg: StackConfig }) {
	return props.cfg.gateway === "kong" ? <Kong cfg={props.cfg} /> : <Envoy cfg={props.cfg} />;
}

function Envoy(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Service
			name="api-gw"
			containerName="supabase-envoy"
			image="envoyproxy/envoy:v1.39.6"
			entrypoint={["/bin/sh", "/docker-entrypoint.sh"]}
		>
			<Restart policy={restartPolicy.UnlessStopped} />
			<GatewayNet />
			<DependsOn name="studio" condition={depend.Healthy} />
			<Probe test={["CMD-SHELL", "timeout 1 bash -c '</dev/tcp/127.0.0.1/8000'"]} interval="10s" timeout="5s" retries={3} />
			<Bind src={file(cfg, "volumes/api/envoy/envoy.yaml")} target="/etc/envoy/envoy.yaml" readOnly />
			<Bind src={file(cfg, "volumes/api/envoy/cds.yaml")} target="/etc/envoy/cds.yaml" readOnly />
			<Bind src={file(cfg, "volumes/api/envoy/lds.template.yaml")} target="/etc/envoy/lds.template.yaml" readOnly />
			<Bind src={file(cfg, "volumes/api/envoy/docker-entrypoint.sh")} target="/docker-entrypoint.sh" readOnly />
			<EnvMap
				values={{
					ANON_KEY: e(cfg, "ANON_KEY"),
					SERVICE_ROLE_KEY: e(cfg, "SERVICE_ROLE_KEY"),
					SUPABASE_PUBLISHABLE_KEY: e(cfg, "SUPABASE_PUBLISHABLE_KEY"),
					SUPABASE_SECRET_KEY: e(cfg, "SUPABASE_SECRET_KEY"),
					ANON_KEY_ASYMMETRIC: e(cfg, "ANON_KEY_ASYMMETRIC"),
					SERVICE_ROLE_KEY_ASYMMETRIC: e(cfg, "SERVICE_ROLE_KEY_ASYMMETRIC"),
					SUPABASE_PUBLIC_URL: e(cfg, "SUPABASE_PUBLIC_URL"),
					DASHBOARD_USERNAME: or(cfg, "DASHBOARD_USERNAME", "supabase"),
					DASHBOARD_PASSWORD: e(cfg, "DASHBOARD_PASSWORD"),
				}}
			/>
			{cfg.gateway === "envoy" && <Tcp host={httpPort(cfg)} container={8000} />}
		</Service>
	);
}

function Kong(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Service
			name="api-gw"
			containerName="supabase-kong"
			image="kong/kong:3.9.3"
			entrypoint={["/bin/sh", "/home/kong/kong-entrypoint.sh"]}
		>
			<Restart policy={restartPolicy.UnlessStopped} />
			<GatewayNet />
			<DependsOn name="studio" condition={depend.Healthy} />
			<Probe test={["CMD", "kong", "health"]} interval="5s" timeout="5s" retries={5} />
			<Tcp host={httpPort(cfg)} container={8000} />
			<Tcp host={port(cfg, "KONG_HTTPS_PORT", 8443)} container={8443} />
			<Bind src={file(cfg, "volumes/api/kong.yml")} target="/home/kong/temp.yml" readOnly />
			<Bind src={file(cfg, "volumes/api/kong-entrypoint.sh")} target="/home/kong/kong-entrypoint.sh" readOnly />
			<EnvMap
				values={{
					KONG_DATABASE: "off",
					KONG_DECLARATIVE_CONFIG: "/usr/local/kong/kong.yml",
					KONG_ROUTER_FLAVOR: "expressions",
					KONG_DNS_ORDER: "LAST,A,CNAME",
					KONG_DNS_NOT_FOUND_TTL: "1",
					KONG_DNS_VALID_TTL: "5",
					KONG_PLUGINS: "request-transformer,cors,key-auth,acl,basic-auth,request-termination,ip-restriction,post-function",
					KONG_NGINX_PROXY_PROXY_BUFFER_SIZE: "160k",
					KONG_NGINX_PROXY_PROXY_BUFFERS: "64 160k",
					KONG_PROXY_ACCESS_LOG: "/dev/stdout combined",
					SUPABASE_ANON_KEY: e(cfg, "ANON_KEY"),
					SUPABASE_SERVICE_KEY: e(cfg, "SERVICE_ROLE_KEY"),
					SUPABASE_PUBLISHABLE_KEY: e(cfg, "SUPABASE_PUBLISHABLE_KEY"),
					SUPABASE_SECRET_KEY: e(cfg, "SUPABASE_SECRET_KEY"),
					ANON_KEY_ASYMMETRIC: e(cfg, "ANON_KEY_ASYMMETRIC"),
					SERVICE_ROLE_KEY_ASYMMETRIC: e(cfg, "SERVICE_ROLE_KEY_ASYMMETRIC"),
					DASHBOARD_USERNAME: or(cfg, "DASHBOARD_USERNAME", "supabase"),
					DASHBOARD_PASSWORD: e(cfg, "DASHBOARD_PASSWORD"),
				}}
			/>
		</Service>
	);
}

function Caddy(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep
			name="caddy"
			container="supabase-caddy"
			image="caddy:2"
			command={[
				"/bin/sh",
				"-c",
				`PROXY_AUTH_PASSWORD=$(caddy hash-password --plaintext "$PROXY_AUTH_PASSWORD") && caddy run --config /etc/caddy/Caddyfile --adapter caddyfile`,
			]}
		>
			<DependsOn name="api-gw" condition={depend.Healthy} />
			<DependsOn name="studio" condition={depend.Healthy} />
			<Tcp host={80} container={80} />
			<Tcp host={443} container={443} />
			<Publish protocol={proto.UDP} hostPort={443} containerPort={443} />
			<Bind src={file(cfg, "volumes/proxy/caddy")} target="/etc/caddy" />
			<Volume name="caddy_data" target="/data" />
			<Volume name="caddy_config" target="/config" />
			<EnvMap
				values={{
					PROXY_DOMAIN: e(cfg, "PROXY_DOMAIN"),
					PROXY_AUTH_USERNAME: or(cfg, "DASHBOARD_USERNAME", "supabase"),
					PROXY_AUTH_PASSWORD: e(cfg, "DASHBOARD_PASSWORD"),
				}}
			/>
		</Keep>
	);
}

function Nginx(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep
			name="nginx"
			container="supabase-nginx"
			image="jonasal/nginx-certbot:6.2.0-nginx1.31.8"
			command={[
				"/bin/bash",
				"-c",
				`printf '%s:%s\\n' "$PROXY_AUTH_USERNAME" "$(openssl passwd -apr1 "$PROXY_AUTH_PASSWORD")" > /etc/nginx/user_conf.d/dashboard-passwd && envsubst '$PROXY_DOMAIN' < /etc/nginx/supabase-nginx.conf.tpl > /etc/nginx/user_conf.d/nginx.conf && /scripts/start_nginx_certbot.sh`,
			]}
		>
			<DependsOn name="api-gw" condition={depend.Healthy} />
			<DependsOn name="studio" condition={depend.Healthy} />
			<Tcp host={80} container={80} />
			<Tcp host={443} container={443} />
			<Bind src={file(cfg, "volumes/proxy/nginx/supabase-nginx.conf.tpl")} target="/etc/nginx/supabase-nginx.conf.tpl" readOnly />
			<Volume name="nginx_letsencrypt" target="/etc/letsencrypt" />
			<EnvMap
				values={{
					PROXY_DOMAIN: e(cfg, "PROXY_DOMAIN"),
					CERTBOT_EMAIL: e(cfg, "CERTBOT_EMAIL"),
					PROXY_AUTH_USERNAME: or(cfg, "DASHBOARD_USERNAME", "supabase"),
					PROXY_AUTH_PASSWORD: e(cfg, "DASHBOARD_PASSWORD"),
				}}
			/>
		</Keep>
	);
}

function Functions(props: { cfg: StackConfig }) {
	const { cfg } = props;
	const host = or(cfg, "POSTGRES_HOST", "db");
	const portNo = or(cfg, "POSTGRES_PORT", "5432");
	const db = or(cfg, "POSTGRES_DB", "postgres");
	return (
		<Keep
			name="functions"
			container="supabase-edge-functions"
			image="supabase/edge-runtime:v1.77.9"
			memory={parseBytes(or(cfg, "FUNCTIONS_MEMORY_LIMIT", "2g"))}
			command={["start", "--user-worker-request-idle-timeout", "150000", "--main-service", "/home/deno/functions/main"]}
		>
			<DependsOn name="api-gw" condition={depend.Healthy} />
			<Probe test={["CMD-SHELL", "timeout 1 bash -c '</dev/tcp/127.0.0.1/9000'"]} interval="5s" timeout="5s" retries={3} />
			<Bind src={file(cfg, "volumes/functions")} target="/home/deno/functions" />
			<Volume name="deno-cache" target="/root/.cache/deno" />
			<EnvMap
				values={{
					JWT_SECRET: e(cfg, "JWT_SECRET"),
					SUPABASE_URL: "http://api-gw:8000",
					SUPABASE_PUBLIC_URL: e(cfg, "SUPABASE_PUBLIC_URL"),
					SUPABASE_ANON_KEY: e(cfg, "ANON_KEY"),
					SUPABASE_SERVICE_ROLE_KEY: e(cfg, "SERVICE_ROLE_KEY"),
					SUPABASE_PUBLISHABLE_KEYS: `{"default":"${e(cfg, "SUPABASE_PUBLISHABLE_KEY")}"}`,
					SUPABASE_SECRET_KEYS: `{"default":"${e(cfg, "SUPABASE_SECRET_KEY")}"}`,
					SUPABASE_DB_URL: `postgresql://postgres:${e(cfg, "POSTGRES_PASSWORD")}@${host}:${portNo}/${db}`,
					VERIFY_JWT: or(cfg, "FUNCTIONS_VERIFY_JWT", "false"),
				}}
			/>
			<EnvIf cfg={cfg} from="JWT_JWKS" name="SUPABASE_JWKS" />
		</Keep>
	);
}

function Supavisor(props: { cfg: StackConfig }) {
	const { cfg } = props;
	const host = or(cfg, "POSTGRES_HOST", "db");
	const portNo = or(cfg, "POSTGRES_PORT", "5432");
	return (
		<Keep
			name="supavisor"
			container="supabase-pooler"
			image="supabase/supavisor:2.9.13"
			command={["/bin/sh", "-c", `/app/bin/migrate && /app/bin/supavisor eval "$(cat /etc/pooler/pooler.exs)" && /app/bin/server`]}
		>
			<DependsOn name="db" condition={depend.Healthy} />
			<Tcp host={port(cfg, "POSTGRES_PORT", 5432)} container={5432} />
			<Tcp host={port(cfg, "POOLER_PROXY_PORT_TRANSACTION", 6543)} container={6543} />
			<Bind src={file(cfg, "volumes/pooler/pooler.exs")} target="/etc/pooler/pooler.exs" readOnly />
			<Probe
				test={["CMD", "curl", "-sSfL", "--head", "-o", "/dev/null", "http://127.0.0.1:4000/api/health"]}
				interval="10s"
				timeout="5s"
				retries={10}
				start="30s"
			/>
			<EnvMap
				values={{
					PORT: "4000",
					POSTGRES_PORT: portNo,
					POSTGRES_HOST: host,
					POSTGRES_DB: or(cfg, "POSTGRES_DB", "postgres"),
					POSTGRES_PASSWORD: e(cfg, "POSTGRES_PASSWORD"),
					DATABASE_URL: `ecto://supabase_admin:${e(cfg, "POSTGRES_PASSWORD")}@${host}:${portNo}/_supabase`,
					CLUSTER_POSTGRES: "true",
					SECRET_KEY_BASE: e(cfg, "SECRET_KEY_BASE"),
					VAULT_ENC_KEY: e(cfg, "VAULT_ENC_KEY"),
					API_JWT_SECRET: e(cfg, "JWT_SECRET"),
					METRICS_JWT_SECRET: e(cfg, "JWT_SECRET"),
					REGION: "local",
					ERL_AFLAGS: "-proto_dist inet_tcp",
					POOLER_TENANT_ID: e(cfg, "POOLER_TENANT_ID"),
					POOLER_DEFAULT_POOL_SIZE: or(cfg, "POOLER_DEFAULT_POOL_SIZE", "20"),
					POOLER_MAX_CLIENT_CONN: or(cfg, "POOLER_MAX_CLIENT_CONN", "100"),
					POOLER_POOL_MODE: "transaction",
					DB_POOL_SIZE: or(cfg, "POOLER_DB_POOL_SIZE", "5"),
				}}
			/>
		</Keep>
	);
}

function PgBouncer(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep name="pgbouncer" container="supabase-pgbouncer" image="edoburu/pgbouncer:v1.25.2-p0">
			<DependsOn name="db" condition={depend.Healthy} />
			<Tcp host={port(cfg, "POOLER_PROXY_PORT_TRANSACTION", 6543)} container={6432} />
			<Probe test={["CMD", "pg_isready", "-h", "localhost", "-p", "6432"]} interval="10s" timeout="5s" retries={3} start="5s" />
			<EnvMap
				values={{
					DB_USER: "pgbouncer",
					DB_PASSWORD: e(cfg, "POSTGRES_PASSWORD"),
					DB_HOST: or(cfg, "POSTGRES_HOST", "db"),
					DB_PORT: or(cfg, "POSTGRES_PORT", "5432"),
					DB_NAME: or(cfg, "POSTGRES_DB", "postgres"),
					LISTEN_PORT: "6432",
					AUTH_TYPE: "scram-sha-256",
					AUTH_QUERY: "SELECT * FROM pgbouncer.get_auth($1)",
					POOL_MODE: "transaction",
					DEFAULT_POOL_SIZE: or(cfg, "POOLER_DEFAULT_POOL_SIZE", "20"),
					MAX_CLIENT_CONN: or(cfg, "POOLER_MAX_CLIENT_CONN", "100"),
					ADMIN_USERS: "pgbouncer",
					STATS_USERS: "pgbouncer",
				}}
			/>
		</Keep>
	);
}

function Minio(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Keep name="minio" container="supabase-minio" image="cgr.dev/chainguard/minio" command={["server", "--console-address", ":9001", "/data"]}>
			<Probe test={["CMD", "mc", "ready", "local"]} interval="2s" timeout="10s" retries={5} />
			<Volume name="minio-data" target="/data" />
			<Env name="MINIO_ROOT_USER" value={e(cfg, "MINIO_ROOT_USER")} />
			<Env name="MINIO_ROOT_PASSWORD" value={e(cfg, "MINIO_ROOT_PASSWORD")} />
		</Keep>
	);
}

function MinioBucket(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Once
			name="minio-createbucket"
			container="supabase-minio-createbucket"
			image="cgr.dev/chainguard/minio-client:latest-dev"
			entrypoint={[
				"/bin/sh",
				"-ec",
				`mc alias set supa-minio http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" && mc mb --ignore-existing "supa-minio/$GLOBAL_S3_BUCKET"`,
			]}
		>
			<DependsOn name="minio" condition={depend.Healthy} />
			<EnvMap
				values={{
					MINIO_ROOT_USER: e(cfg, "MINIO_ROOT_USER"),
					MINIO_ROOT_PASSWORD: e(cfg, "MINIO_ROOT_PASSWORD"),
					GLOBAL_S3_BUCKET: or(cfg, "GLOBAL_S3_BUCKET", "stub"),
				}}
			/>
		</Once>
	);
}

function Rustfs(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Service name="rustfs" containerName="supabase-rustfs" image="rustfs/rustfs:1.0.4" securityOpt="no-new-privileges:true">
			<Restart policy={restartPolicy.UnlessStopped} />
			<Network name="default" />
			<Probe test={["CMD", "curl", "-f", "http://rustfs:9000/health"]} interval="2s" timeout="10s" retries={5} />
			<Volume name="rustfs-data" target="/data" />
			<EnvMap
				values={{
					RUSTFS_ACCESS_KEY: e(cfg, "MINIO_ROOT_USER"),
					RUSTFS_SECRET_KEY: e(cfg, "MINIO_ROOT_PASSWORD"),
					RUSTFS_CONSOLE_ADDRESS: "0.0.0.0:9001",
					RUSTFS_CORS_ALLOWED_ORIGINS: "*",
					RUSTFS_CONSOLE_CORS_ALLOWED_ORIGINS: "*",
				}}
			/>
		</Service>
	);
}

function RustfsBucket(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Once
			name="rustfs-createbucket"
			container="supabase-rustfs-createbucket"
			image="rustfs/rc"
			entrypoint={[
				"/bin/sh",
				"-ec",
				`rc alias set supa-rustfs http://rustfs:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" && rc mb --ignore-existing "supa-rustfs/$GLOBAL_S3_BUCKET"`,
			]}
		>
			<DependsOn name="rustfs" condition={depend.Healthy} />
			<EnvMap
				values={{
					MINIO_ROOT_USER: e(cfg, "MINIO_ROOT_USER"),
					MINIO_ROOT_PASSWORD: e(cfg, "MINIO_ROOT_PASSWORD"),
					GLOBAL_S3_BUCKET: or(cfg, "GLOBAL_S3_BUCKET", "stub"),
				}}
			/>
		</Once>
	);
}

function Analytics(props: { cfg: StackConfig }) {
	const { cfg } = props;
	const host = or(cfg, "POSTGRES_HOST", "db");
	const portNo = or(cfg, "POSTGRES_PORT", "5432");
	return (
		<Keep name="analytics" container="supabase-analytics" image="supabase/logflare:1.50.15">
			<DependsOn name="db" condition={depend.Healthy} />
			<Probe test={["CMD-SHELL", "curl -sSfL -o /dev/null http://localhost:4000/health"]} interval="5s" timeout="5s" retries={10} start="30s" />
			<EnvMap
				values={{
					LOGFLARE_NODE_HOST: "127.0.0.1",
					DB_USERNAME: "supabase_admin",
					DB_DATABASE: "_supabase",
					DB_HOSTNAME: host,
					DB_PORT: portNo,
					DB_PASSWORD: e(cfg, "POSTGRES_PASSWORD"),
					DB_SCHEMA: "_analytics",
					LOGFLARE_SINGLE_TENANT: "true",
					LOGFLARE_SUPABASE_MODE: "true",
					LOGFLARE_PUBLIC_ACCESS_TOKEN: e(cfg, "LOGFLARE_PUBLIC_ACCESS_TOKEN"),
					LOGFLARE_PRIVATE_ACCESS_TOKEN: e(cfg, "LOGFLARE_PRIVATE_ACCESS_TOKEN"),
					LOGFLARE_FEATURE_FLAG_OVERRIDE: "multibackend=true",
					POSTGRES_BACKEND_URL: `postgresql://supabase_admin:${e(cfg, "POSTGRES_PASSWORD")}@${host}:${portNo}/_supabase`,
					POSTGRES_BACKEND_SCHEMA: "_analytics",
				}}
			/>
		</Keep>
	);
}

function Vector(props: { cfg: StackConfig }) {
	const { cfg } = props;
	return (
		<Service
			name="vector"
			containerName="supabase-vector"
			image="timberio/vector:0.53.0-alpine"
			securityOpt="label=disable"
			command={["--config", "/etc/vector/vector.yml"]}
		>
			<Restart policy={restartPolicy.UnlessStopped} />
			<Network name="default" />
			<DependsOn name="analytics" condition={depend.Healthy} />
			<Probe test={["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://vector:9001/health"]} interval="5s" timeout="5s" retries={3} />
			<Bind src={file(cfg, "volumes/logs/vector.yml")} target="/etc/vector/vector.yml" readOnly />
			<Bind src={or(cfg, "DOCKER_SOCKET_LOCATION", "/var/run/docker.sock")} target="/var/run/docker.sock" readOnly />
			<Env name="LOGFLARE_PUBLIC_ACCESS_TOKEN" value={e(cfg, "LOGFLARE_PUBLIC_ACCESS_TOKEN")} />
		</Service>
	);
}

function Keep(props: {
	name: string;
	container: string;
	image?: string;
	command?: string[];
	entrypoint?: string[];
	memory?: number;
	securityOpt?: string;
	children?: Child;
}) {
	return (
		<Service
			name={props.name}
			containerName={props.container}
			image={props.image}
			command={props.command}
			entrypoint={props.entrypoint}
			memory={props.memory}
			securityOpt={props.securityOpt}
		>
			<Restart policy={restartPolicy.UnlessStopped} />
			<Network name="default" />
			{props.children}
		</Service>
	);
}

function Once(props: { name: string; container: string; image: string; entrypoint?: string[]; children?: Child }) {
	return (
		<Service name={props.name} containerName={props.container} image={props.image} entrypoint={props.entrypoint}>
			<Network name="default" />
			{props.children}
		</Service>
	);
}

function GatewayNet() {
	const net = Network({ name: "default" });
	net.props.alias = ["envoy", "kong"] as unknown as string;
	return net;
}

function Probe(props: { test: string[]; interval: string; timeout: string; retries: number; start?: string }) {
	return <Health test={props.test} interval={props.interval} timeout={props.timeout} retries={props.retries} startPeriod={props.start} />;
}

function Bind(props: { src: string; target: string; readOnly?: boolean }) {
	return <Volume kind={vol.Bind} source={props.src} target={props.target} readOnly={props.readOnly} />;
}

function Tcp(props: { host: number; container: number }) {
	return <Publish protocol={proto.TCP} hostPort={props.host} containerPort={props.container} />;
}

function EnvMap(props: { values: Record<string, string> }) {
	return (
		<>
			{Object.entries(props.values).map(([name, value]) => (
				<Env key={name} name={name} value={value} />
			))}
		</>
	);
}

function EnvIf(props: { cfg: StackConfig; from: string; name: string }) {
	const value = e(props.cfg, props.from);
	if (value === "") {
		return null;
	}
	return <Env name={props.name} value={value} />;
}

function parseBytes(value: string): number {
	const text = value.trim().toLowerCase();
	const match = /^(\d+(?:\.\d+)?)([gmk])?$/.exec(text);
	if (!match) {
		return 2 * 1024 * 1024 * 1024;
	}
	const n = Number(match[1]);
	const mult = match[2] === "g" ? 1024 * 1024 * 1024 : match[2] === "m" ? 1024 * 1024 : match[2] === "k" ? 1024 : 1;
	return Math.floor(n * mult);
}

function e(cfg: StackConfig, key: string): string {
	return (cfg.env[key] ?? "").trim();
}

function or(cfg: StackConfig, key: string, fallback: string): string {
	const value = e(cfg, key);
	return value === "" ? fallback : value;
}

function port(cfg: StackConfig, key: string, fallback: number): number {
	const value = e(cfg, key);
	if (value === "") {
		return fallback;
	}
	const n = Number(value);
	return Number.isInteger(n) && n >= 0 && n <= 65535 ? n : fallback;
}

function httpPort(cfg: StackConfig): number {
	if (e(cfg, "API_GW_HTTP_PORT") !== "") {
		return port(cfg, "API_GW_HTTP_PORT", 8000);
	}
	return port(cfg, "KONG_HTTP_PORT", 8000);
}

function pg(cfg: StackConfig, user: string, db: string): string {
	return `postgres://${user}:${e(cfg, "POSTGRES_PASSWORD")}@${or(cfg, "POSTGRES_HOST", "db")}:${or(cfg, "POSTGRES_PORT", "5432")}/${db}`;
}

function file(cfg: StackConfig, rel: string): string {
	return join(cfg.dir, rel);
}

function loadEnvFile(path: string): Record<string, string> {
	if (!existsSync(path)) {
		return {};
	}
	const out: Record<string, string> = {};
	for (const line of readFileSync(path, "utf8").split("\n")) {
		const trimmed = line.trim();
		if (trimmed === "" || trimmed.startsWith("#")) {
			continue;
		}
		const eq = trimmed.indexOf("=");
		if (eq <= 0) {
			continue;
		}
		const key = trimmed.slice(0, eq).trim();
		let value = trimmed.slice(eq + 1).trim();
		if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
			value = value.slice(1, -1);
		}
		out[key] = value;
	}
	return out;
}

function flag(name: string): string | undefined {
	const prefixed = process.argv.find((arg) => arg.startsWith(`--${name}=`));
	if (prefixed) {
		return prefixed.slice(name.length + 3);
	}
	const i = process.argv.indexOf(`--${name}`);
	const next = i >= 0 ? process.argv[i + 1] : undefined;
	if (next && !next.startsWith("--")) {
		return next;
	}
	return undefined;
}

function oneOf<T extends string>(name: string, value: string, allowed: readonly T[]): T {
	if ((allowed as readonly string[]).includes(value)) {
		return value as T;
	}
	throw new Error(`${name} ${JSON.stringify(value)} must be ${allowed.join(", ")}`);
}

function configFromArgs(): StackConfig {
	const compiledDir = dirname(fileURLToPath(import.meta.url));
	const dir = resolve(process.env.SUPABASE_DIR?.trim() || flag("dir") || resolve(compiledDir, "../../examples/supabase/docker"));
	const env = { ...loadEnvFile(join(dir, ".env.example")), ...loadEnvFile(join(dir, ".env")) };
	for (const [key, value] of Object.entries(process.env)) {
		if (value !== undefined && key in env) {
			env[key] = value;
		}
	}
	return {
		dir,
		env,
		gateway: oneOf("gateway", flag("gateway") ?? "envoy", ["envoy", "kong", "caddy", "nginx"] as const),
		postgres: oneOf("postgres", flag("postgres") ?? "17", ["15", "17"] as const),
		pooler: oneOf("pooler", flag("pooler") ?? "supavisor", ["supavisor", "pgbouncer"] as const),
		storage: oneOf("storage", flag("storage") ?? "file", ["file", "s3", "rustfs"] as const),
		logs: process.argv.includes("--logs"),
	};
}

process.stdout.write(exportYAML(<Supabase cfg={configFromArgs()} />));
