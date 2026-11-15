package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/logic"
)

// stackConfig is the self-hosted Supabase project, expressed with containerkit
// instead of the docker-compose*.yml overrides. Each field replaces one of
// those files: gateway (envoy, kong, caddy, nginx), postgres (15 or 17),
// pooler (supavisor or pgbouncer), storage (file, s3, rustfs), and logs.
type stackConfig struct {
	dir      string
	env      map[string]string
	gateway  string
	postgres string
	pooler   string
	storage  string
	logs     bool
}

func setupStack(cfg stackConfig) (*Compose, error) {
	if err := cfg.valid(); err != nil {
		return nil, err
	}
	if cfg.env == nil {
		cfg.env = map[string]string{}
	}
	p := NewCompose("supabase", logic.Group(
		AddNetwork(Name("default"), Driver("bridge")),
		AddVolume(Name("db-config")),
		AddVolume(Name("deno-cache")),
		logic.Switch(
			logic.Case(cfg.gateway == "caddy", AddVolume(Name("caddy_data")), AddVolume(Name("caddy_config"))),
			logic.Case(cfg.gateway == "nginx", AddVolume(Name("nginx_letsencrypt"))),
		),
		logic.Switch(
			logic.Case(cfg.storage == "s3", AddVolume(Name("minio-data"))),
			logic.Case(cfg.storage == "rustfs", AddVolume(Name("rustfs-data"))),
		),
		logic.When(cfg.logs, cfg.analytics()),
		cfg.db(),
		cfg.auth(),
		cfg.rest(),
		cfg.realtime(),
		cfg.imgproxy(),
		cfg.meta(),
		cfg.storageService(),
		cfg.studio(),
		cfg.gatewayService(),
		cfg.functions(),
		logic.Switch(
			logic.Case(cfg.storage == "s3", cfg.minio(), cfg.minioBucket()),
			logic.Case(cfg.storage == "rustfs", cfg.rustfs(), cfg.rustfsBucket()),
		),
		logic.Switch(
			logic.Case(cfg.pooler == "supavisor", cfg.supavisor()),
			logic.Case(cfg.pooler == "pgbouncer", cfg.pgbouncer()),
		),
		logic.Switch(
			logic.Case(cfg.gateway == "caddy", cfg.caddy()),
			logic.Case(cfg.gateway == "nginx", cfg.nginx()),
		),
		logic.When(cfg.logs, cfg.vector()),
	))
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (c stackConfig) valid() error {
	switch c.gateway {
	case "envoy", "kong", "caddy", "nginx":
	default:
		return fmt.Errorf("gateway %q must be envoy, kong, caddy, or nginx", c.gateway)
	}
	switch c.postgres {
	case "15", "17":
	default:
		return fmt.Errorf("postgres %q must be 15 or 17", c.postgres)
	}
	switch c.pooler {
	case "supavisor", "pgbouncer":
	default:
		return fmt.Errorf("pooler %q must be supavisor or pgbouncer", c.pooler)
	}
	switch c.storage {
	case "file", "s3", "rustfs":
	default:
		return fmt.Errorf("storage %q must be file, s3, or rustfs", c.storage)
	}
	return nil
}

func (c stackConfig) e(key string) string {
	return strings.TrimSpace(c.env[key])
}

func (c stackConfig) or(key, fallback string) string {
	if v := c.e(key); v != "" {
		return v
	}
	return fallback
}

func (c stackConfig) port(key string, fallback uint16) uint16 {
	v := c.e(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseUint(v, 10, 16)
	if err != nil {
		return fallback
	}
	return uint16(n)
}

func (c stackConfig) httpPort() uint16 {
	if c.e("API_GW_HTTP_PORT") != "" {
		return c.port("API_GW_HTTP_PORT", 8000)
	}
	return c.port("KONG_HTTP_PORT", 8000)
}

func (c stackConfig) file(rel string) string {
	return filepath.Join(c.dir, filepath.FromSlash(rel))
}

func (c stackConfig) pg(user, db string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user, c.e("POSTGRES_PASSWORD"), c.or("POSTGRES_HOST", "db"), c.or("POSTGRES_PORT", "5432"), db)
}

func (c stackConfig) envIf(key, name string) Grouped {
	return logic.When(c.e(key) != "", Env(name, c.e(key)))
}

type sqlMount struct{ src, dst string }

func lan() NetworkOpt { return Network(Name("default")) }

func gatewayNet() NetworkOpt {
	return Network(Name("default"), Alias("envoy"), Alias("kong"))
}

func keep(name, container, image string, opts ...ServiceOption) AddServiceOpt {
	head := []ServiceOption{Name(name), ContainerName(container), Restart(Policy(UnlessStopped)), lan()}
	if image != "" {
		head = append(head, Image(image))
	}
	return AddService(append(head, opts...)...)
}

func once(name, container, image string, opts ...ServiceOption) AddServiceOpt {
	head := []ServiceOption{Name(name), ContainerName(container), lan()}
	if image != "" {
		head = append(head, Image(image))
	}
	return AddService(append(head, opts...)...)
}

func probe(test []string, every, timeout time.Duration, retries int, start time.Duration) HealthOpt {
	return Health(
		Test(test...),
		Interval(every),
		Timeout(timeout),
		Retries(retries),
		logic.When(start > 0, StartPeriod(start)),
	)
}

func bind(src, target string, ro bool) VolumeOpt {
	return Volume(
		Kind(Bind),
		Source(src),
		Target(target),
		logic.When(ro, ReadOnly(true)),
	)
}

func tcp(host, container uint16) PublishOpt {
	return Publish(Protocol(TCP), HostPort(host), ContainerPort(container))
}

func (c stackConfig) db() AddServiceOpt {
	port := c.or("POSTGRES_PORT", "5432")
	db := c.or("POSTGRES_DB", "postgres")
	return keep("db", "supabase-db", "",
		logic.WhenElse(c.postgres == "15",
			Image("supabase/postgres:15.8.1.085"),
			Image("supabase/postgres:17.6.1.136"),
		),
		probe([]string{"CMD", "pg_isready", "-U", "postgres", "-h", "localhost"}, 5*time.Second, 5*time.Second, 10, 0),
		logic.Each([]sqlMount{
			{"volumes/db/realtime.sql", "/docker-entrypoint-initdb.d/migrations/99-realtime.sql"},
			{"volumes/db/webhooks.sql", "/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql"},
			{"volumes/db/roles.sql", "/docker-entrypoint-initdb.d/init-scripts/99-roles.sql"},
			{"volumes/db/jwt.sql", "/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql"},
			{"volumes/db/data", "/var/lib/postgresql/data"},
			{"volumes/db/_supabase.sql", "/docker-entrypoint-initdb.d/migrations/97-_supabase.sql"},
			{"volumes/db/logs.sql", "/docker-entrypoint-initdb.d/migrations/99-logs.sql"},
			{"volumes/db/pooler.sql", "/docker-entrypoint-initdb.d/migrations/99-pooler.sql"},
		}, func(_ int, m sqlMount) any {
			return bind(c.file(m.src), m.dst, false)
		}),
		Volume(Name("db-config"), Target("/etc/postgresql-custom")),
		Env("POSTGRES_HOST", "/var/run/postgresql"),
		Env("PGPORT", port),
		Env("POSTGRES_PORT", port),
		Env("PGPASSWORD", c.e("POSTGRES_PASSWORD")),
		Env("POSTGRES_PASSWORD", c.e("POSTGRES_PASSWORD")),
		Env("PGDATABASE", db),
		Env("POSTGRES_DB", db),
		Env("JWT_EXP", c.or("JWT_EXPIRY", "3600")),
		Command("postgres", "-c", "config_file=/etc/postgresql/postgresql.conf", "-c", "log_min_messages=fatal"),
	)
}

func (c stackConfig) auth() AddServiceOpt {
	opts := []ServiceOption{
		DependsOn(Name("db"), Condition(Healthy)),
		probe([]string{"CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9999/health"}, 5*time.Second, 5*time.Second, 3, 0),
		Env("GOTRUE_API_HOST", "0.0.0.0"),
		Env("GOTRUE_API_PORT", "9999"),
		Env("API_EXTERNAL_URL", c.e("API_EXTERNAL_URL")),
		Env("GOTRUE_DB_DRIVER", "postgres"),
		Env("GOTRUE_DB_DATABASE_URL", c.pg("supabase_auth_admin", c.or("POSTGRES_DB", "postgres"))),
		Env("GOTRUE_SITE_URL", c.e("SITE_URL")),
		Env("GOTRUE_URI_ALLOW_LIST", c.e("ADDITIONAL_REDIRECT_URLS")),
		Env("GOTRUE_DISABLE_SIGNUP", c.or("DISABLE_SIGNUP", "false")),
		Env("GOTRUE_JWT_ADMIN_ROLES", "service_role"),
		Env("GOTRUE_JWT_AUD", "authenticated"),
		Env("GOTRUE_JWT_DEFAULT_GROUP_NAME", "authenticated"),
		Env("GOTRUE_JWT_EXP", c.or("JWT_EXPIRY", "3600")),
		Env("GOTRUE_JWT_SECRET", c.e("JWT_SECRET")),
		Env("GOTRUE_JWT_ISSUER", c.e("API_EXTERNAL_URL")),
		Env("GOTRUE_EXTERNAL_EMAIL_ENABLED", c.or("ENABLE_EMAIL_SIGNUP", "true")),
		Env("GOTRUE_EXTERNAL_ANONYMOUS_USERS_ENABLED", c.or("ENABLE_ANONYMOUS_USERS", "false")),
		Env("GOTRUE_MAILER_AUTOCONFIRM", c.or("ENABLE_EMAIL_AUTOCONFIRM", "false")),
		Env("GOTRUE_SMTP_ADMIN_EMAIL", c.e("SMTP_ADMIN_EMAIL")),
		Env("GOTRUE_SMTP_HOST", c.e("SMTP_HOST")),
		Env("GOTRUE_SMTP_PORT", c.e("SMTP_PORT")),
		Env("GOTRUE_SMTP_USER", c.e("SMTP_USER")),
		Env("GOTRUE_SMTP_PASS", c.e("SMTP_PASS")),
		Env("GOTRUE_SMTP_SENDER_NAME", c.e("SMTP_SENDER_NAME")),
		Env("GOTRUE_MAILER_URLPATHS_INVITE", c.e("MAILER_URLPATHS_INVITE")),
		Env("GOTRUE_MAILER_URLPATHS_CONFIRMATION", c.e("MAILER_URLPATHS_CONFIRMATION")),
		Env("GOTRUE_MAILER_URLPATHS_RECOVERY", c.e("MAILER_URLPATHS_RECOVERY")),
		Env("GOTRUE_MAILER_URLPATHS_EMAIL_CHANGE", c.e("MAILER_URLPATHS_EMAIL_CHANGE")),
		Env("GOTRUE_EXTERNAL_PHONE_ENABLED", c.or("ENABLE_PHONE_SIGNUP", "true")),
		Env("GOTRUE_SMS_AUTOCONFIRM", c.or("ENABLE_PHONE_AUTOCONFIRM", "true")),
		c.envIf("JWT_KEYS", "GOTRUE_JWT_KEYS"),
	}
	return keep("auth", "supabase-auth", "supabase/gotrue:v2.197.0", opts...)
}

func (c stackConfig) rest() AddServiceOpt {
	secret := c.e("JWT_JWKS")
	if secret == "" {
		secret = c.e("JWT_SECRET")
	}
	return keep("rest", "supabase-rest", "postgrest/postgrest:v14.18",
		DependsOn(Name("db"), Condition(Healthy)),
		probe([]string{"CMD", "postgrest", "--ready"}, 5*time.Second, 5*time.Second, 3, 0),
		Env("PGRST_DB_URI", c.pg("authenticator", c.or("POSTGRES_DB", "postgres"))),
		Env("PGRST_DB_SCHEMAS", c.or("PGRST_DB_SCHEMAS", "public,graphql_public")),
		Env("PGRST_DB_MAX_ROWS", c.or("PGRST_DB_MAX_ROWS", "1000")),
		Env("PGRST_DB_EXTRA_SEARCH_PATH", c.or("PGRST_DB_EXTRA_SEARCH_PATH", "public")),
		Env("PGRST_DB_ANON_ROLE", "anon"),
		Env("PGRST_ADMIN_SERVER_PORT", "3001"),
		Env("PGRST_ADMIN_SERVER_HOST", "localhost"),
		Env("PGRST_JWT_SECRET", secret),
		Env("PGRST_DB_USE_LEGACY_GUCS", "false"),
		Env("PGRST_APP_SETTINGS_JWT_EXP", c.or("JWT_EXPIRY", "3600")),
		Command("postgrest"),
	)
}

func (c stackConfig) realtime() AddServiceOpt {
	opts := []ServiceOption{
		DependsOn(Name("db"), Condition(Healthy)),
		probe([]string{"CMD-SHELL", fmt.Sprintf(
			`curl -sSfL --head -o /dev/null -H "Authorization: Bearer %s" http://localhost:4000/api/tenants/realtime-dev/health`,
			c.e("ANON_KEY"),
		)}, 30*time.Second, 5*time.Second, 3, 10*time.Second),
		Env("PORT", "4000"),
		Env("DB_HOST", c.or("POSTGRES_HOST", "db")),
		Env("DB_PORT", c.or("POSTGRES_PORT", "5432")),
		Env("DB_USER", "supabase_admin"),
		Env("DB_PASSWORD", c.e("POSTGRES_PASSWORD")),
		Env("DB_NAME", c.or("POSTGRES_DB", "postgres")),
		Env("DB_AFTER_CONNECT_QUERY", "SET search_path TO _realtime"),
		Env("DB_ENC_KEY", c.or("REALTIME_DB_ENC_KEY", "supabaserealtime")),
		Env("API_JWT_SECRET", c.e("JWT_SECRET")),
		Env("SECRET_KEY_BASE", c.e("SECRET_KEY_BASE")),
		Env("METRICS_JWT_SECRET", c.e("JWT_SECRET")),
		Env("ERL_AFLAGS", "-proto_dist inet_tcp"),
		Env("DNS_NODES", "''"),
		Env("RLIMIT_NOFILE", "10000"),
		Env("APP_NAME", "realtime"),
		Env("SEED_SELF_HOST", "true"),
		Env("RUN_JANITOR", "true"),
		Env("DISABLE_HEALTHCHECK_LOGGING", "true"),
		c.envIf("JWT_JWKS", "API_JWT_JWKS"),
	}
	return keep("realtime", "realtime-dev.supabase-realtime", "supabase/realtime:v2.148.3", opts...)
}

func (c stackConfig) storageService() AddServiceOpt {
	opts := []ServiceOption{
		DependsOn(Name("db"), Condition(Healthy)),
		DependsOn(Name("rest"), Condition(Started)),
		DependsOn(Name("imgproxy"), Condition(Started)),
		probe([]string{"CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://storage:5000/status"}, 5*time.Second, 5*time.Second, 3, 10*time.Second),
		bind(c.file("volumes/storage"), "/var/lib/storage", false),
		Env("ANON_KEY", c.e("ANON_KEY")),
		Env("SERVICE_KEY", c.e("SERVICE_ROLE_KEY")),
		Env("POSTGREST_URL", "http://rest:3000"),
		Env("AUTH_JWT_SECRET", c.e("JWT_SECRET")),
		Env("DATABASE_URL", c.pg("supabase_storage_admin", c.or("POSTGRES_DB", "postgres"))),
		Env("STORAGE_PUBLIC_URL", c.e("SUPABASE_PUBLIC_URL")),
		Env("REQUEST_ALLOW_X_FORWARDED_PATH", "true"),
		Env("FILE_SIZE_LIMIT", "52428800"),
		Env("GLOBAL_S3_BUCKET", c.or("GLOBAL_S3_BUCKET", "stub")),
		Env("FILE_STORAGE_BACKEND_PATH", "/var/lib/storage"),
		Env("TENANT_ID", c.or("STORAGE_TENANT_ID", "stub")),
		Env("REGION", c.or("REGION", "stub")),
		Env("ENABLE_IMAGE_TRANSFORMATION", "true"),
		Env("IMGPROXY_URL", "http://imgproxy:5001"),
		Env("S3_PROTOCOL_ACCESS_KEY_ID", c.e("S3_PROTOCOL_ACCESS_KEY_ID")),
		Env("S3_PROTOCOL_ACCESS_KEY_SECRET", c.e("S3_PROTOCOL_ACCESS_KEY_SECRET")),
		c.envIf("JWT_JWKS", "JWT_JWKS"),
		logic.Switch(
			logic.Case(c.storage == "s3",
				DependsOn(Name("minio"), Condition(Healthy)),
				DependsOn(Name("minio-createbucket"), Condition(CompletedSuccessfully)),
				Env("STORAGE_BACKEND", "s3"),
				Env("GLOBAL_S3_ENDPOINT", "http://minio:9000"),
				Env("GLOBAL_S3_PROTOCOL", "http"),
				Env("GLOBAL_S3_FORCE_PATH_STYLE", "true"),
				Env("AWS_ACCESS_KEY_ID", c.e("MINIO_ROOT_USER")),
				Env("AWS_SECRET_ACCESS_KEY", c.e("MINIO_ROOT_PASSWORD")),
			),
			logic.Case(c.storage == "rustfs",
				DependsOn(Name("rustfs"), Condition(Healthy)),
				DependsOn(Name("rustfs-createbucket"), Condition(CompletedSuccessfully)),
				Env("STORAGE_BACKEND", "s3"),
				Env("GLOBAL_S3_ENDPOINT", "http://rustfs:9000"),
				Env("GLOBAL_S3_PROTOCOL", "http"),
				Env("GLOBAL_S3_FORCE_PATH_STYLE", "true"),
				Env("AWS_ACCESS_KEY_ID", c.e("MINIO_ROOT_USER")),
				Env("AWS_SECRET_ACCESS_KEY", c.e("MINIO_ROOT_PASSWORD")),
			),
			logic.Default(Env("STORAGE_BACKEND", "file")),
		),
	}
	return keep("storage", "supabase-storage", "supabase/storage-api:v1.89.12", opts...)
}

func (c stackConfig) imgproxy() AddServiceOpt {
	return keep("imgproxy", "supabase-imgproxy", "darthsim/imgproxy:v4.0.21",
		probe([]string{"CMD", "imgproxy", "health"}, 5*time.Second, 5*time.Second, 3, 0),
		bind(c.file("volumes/storage"), "/var/lib/storage", false),
		Env("IMGPROXY_BIND", ":5001"),
		Env("IMGPROXY_LOCAL_FILESYSTEM_ROOT", "/"),
		Env("IMGPROXY_USE_ETAG", "true"),
		Env("IMGPROXY_AUTO_WEBP", c.or("IMGPROXY_AUTO_WEBP", "true")),
		Env("IMGPROXY_MAX_SRC_RESOLUTION", "16.8"),
	)
}

func (c stackConfig) meta() AddServiceOpt {
	return keep("meta", "supabase-meta", "supabase/postgres-meta:v0.100.0",
		DependsOn(Name("db"), Condition(Healthy)),
		Env("PG_META_PORT", "8080"),
		Env("PG_META_DB_HOST", c.or("POSTGRES_HOST", "db")),
		Env("PG_META_DB_PORT", c.or("POSTGRES_PORT", "5432")),
		Env("PG_META_DB_NAME", c.or("POSTGRES_DB", "postgres")),
		Env("PG_META_DB_USER", "postgres"),
		Env("PG_META_DB_PASSWORD", c.e("POSTGRES_PASSWORD")),
		Env("CRYPTO_KEY", c.e("PG_META_CRYPTO_KEY")),
	)
}

func (c stackConfig) studio() AddServiceOpt {
	opts := []ServiceOption{
		probe([]string{"CMD-SHELL", `node -e "fetch('http://localhost:3000/api/platform/profile').then((r) => {if (r.status !== 200) throw new Error(r.status)})"`}, 5*time.Second, 10*time.Second, 3, 20*time.Second),
		bind(c.file("volumes/snippets"), "/app/snippets", false),
		bind(c.file("volumes/functions"), "/app/edge-functions", true),
		Env("HOSTNAME", "0.0.0.0"),
		Env("STUDIO_PG_META_URL", "http://meta:8080"),
		Env("POSTGRES_PORT", c.or("POSTGRES_PORT", "5432")),
		Env("POSTGRES_HOST", c.or("POSTGRES_HOST", "db")),
		Env("POSTGRES_DB", c.or("POSTGRES_DB", "postgres")),
		Env("POSTGRES_PASSWORD", c.e("POSTGRES_PASSWORD")),
		Env("POSTGRES_USER_READ_WRITE", "postgres"),
		Env("PG_META_CRYPTO_KEY", c.e("PG_META_CRYPTO_KEY")),
		Env("PGRST_DB_SCHEMAS", c.or("PGRST_DB_SCHEMAS", "public,graphql_public")),
		Env("PGRST_DB_MAX_ROWS", c.or("PGRST_DB_MAX_ROWS", "1000")),
		Env("PGRST_DB_EXTRA_SEARCH_PATH", c.or("PGRST_DB_EXTRA_SEARCH_PATH", "public")),
		Env("DEFAULT_ORGANIZATION_NAME", c.or("STUDIO_DEFAULT_ORGANIZATION", "Default Organization")),
		Env("DEFAULT_PROJECT_NAME", c.or("STUDIO_DEFAULT_PROJECT", "Default Project")),
		Env("OPENAI_API_KEY", c.e("OPENAI_API_KEY")),
		Env("SUPABASE_URL", "http://api-gw:8000"),
		Env("SUPABASE_PUBLIC_URL", c.e("SUPABASE_PUBLIC_URL")),
		Env("SUPABASE_ANON_KEY", c.e("ANON_KEY")),
		Env("SUPABASE_SERVICE_KEY", c.e("SERVICE_ROLE_KEY")),
		Env("AUTH_JWT_SECRET", c.e("JWT_SECRET")),
		Env("SUPABASE_PUBLISHABLE_KEY", c.e("SUPABASE_PUBLISHABLE_KEY")),
		Env("SUPABASE_SECRET_KEY", c.e("SUPABASE_SECRET_KEY")),
		Env("SNIPPETS_MANAGEMENT_FOLDER", "/app/snippets"),
		Env("EDGE_FUNCTIONS_MANAGEMENT_FOLDER", "/app/edge-functions"),
		logic.When(c.logs,
			DependsOn(Name("analytics"), Condition(Healthy)),
			Env("LOGFLARE_PRIVATE_ACCESS_TOKEN", c.e("LOGFLARE_PRIVATE_ACCESS_TOKEN")),
			Env("LOGFLARE_URL", "http://analytics:4000"),
			Env("ENABLED_FEATURES_LOGS_ALL", "true"),
		),
		logic.Unless(c.logs, Env("ENABLED_FEATURES_LOGS_ALL", "false")),
	}
	return keep("studio", "supabase-studio", "supabase/studio:2026.11.02-sha-a05e7ec", opts...)
}

func (c stackConfig) gatewayService() Grouped {
	return logic.Switch(
		logic.Case(c.gateway == "kong", c.kong()),
		logic.Default(c.envoy()),
	)
}

func (c stackConfig) envoy() AddServiceOpt {
	opts := []ServiceOption{
		gatewayNet(),
		DependsOn(Name("studio"), Condition(Healthy)),
		probe([]string{"CMD-SHELL", "timeout 1 bash -c '</dev/tcp/127.0.0.1/8000'"}, 10*time.Second, 5*time.Second, 3, 0),
		bind(c.file("volumes/api/envoy/envoy.yaml"), "/etc/envoy/envoy.yaml", true),
		bind(c.file("volumes/api/envoy/cds.yaml"), "/etc/envoy/cds.yaml", true),
		bind(c.file("volumes/api/envoy/lds.template.yaml"), "/etc/envoy/lds.template.yaml", true),
		bind(c.file("volumes/api/envoy/docker-entrypoint.sh"), "/docker-entrypoint.sh", true),
		Env("ANON_KEY", c.e("ANON_KEY")),
		Env("SERVICE_ROLE_KEY", c.e("SERVICE_ROLE_KEY")),
		Env("SUPABASE_PUBLISHABLE_KEY", c.e("SUPABASE_PUBLISHABLE_KEY")),
		Env("SUPABASE_SECRET_KEY", c.e("SUPABASE_SECRET_KEY")),
		Env("ANON_KEY_ASYMMETRIC", c.e("ANON_KEY_ASYMMETRIC")),
		Env("SERVICE_ROLE_KEY_ASYMMETRIC", c.e("SERVICE_ROLE_KEY_ASYMMETRIC")),
		Env("SUPABASE_PUBLIC_URL", c.e("SUPABASE_PUBLIC_URL")),
		Env("DASHBOARD_USERNAME", c.or("DASHBOARD_USERNAME", "supabase")),
		Env("DASHBOARD_PASSWORD", c.e("DASHBOARD_PASSWORD")),
		Entrypoint("/bin/sh", "/docker-entrypoint.sh"),
		logic.When(c.gateway == "envoy", tcp(c.httpPort(), 8000)),
	}
	// lan() would attach a second network without the envoy/kong aliases.
	head := []ServiceOption{Name("api-gw"), ContainerName("supabase-envoy"), Image("envoyproxy/envoy:v1.39.6"), Restart(Policy(UnlessStopped))}
	return AddService(append(head, opts...)...)
}

func (c stackConfig) kong() AddServiceOpt {
	head := []ServiceOption{Name("api-gw"), ContainerName("supabase-kong"), Image("kong/kong:3.9.3"), Restart(Policy(UnlessStopped))}
	return AddService(append(head,
		gatewayNet(),
		DependsOn(Name("studio"), Condition(Healthy)),
		probe([]string{"CMD", "kong", "health"}, 5*time.Second, 5*time.Second, 5, 0),
		tcp(c.httpPort(), 8000),
		tcp(c.port("KONG_HTTPS_PORT", 8443), 8443),
		bind(c.file("volumes/api/kong.yml"), "/home/kong/temp.yml", true),
		bind(c.file("volumes/api/kong-entrypoint.sh"), "/home/kong/kong-entrypoint.sh", true),
		Env("KONG_DATABASE", "off"),
		Env("KONG_DECLARATIVE_CONFIG", "/usr/local/kong/kong.yml"),
		Env("KONG_ROUTER_FLAVOR", "expressions"),
		Env("KONG_DNS_ORDER", "LAST,A,CNAME"),
		Env("KONG_DNS_NOT_FOUND_TTL", "1"),
		Env("KONG_DNS_VALID_TTL", "5"),
		Env("KONG_PLUGINS", "request-transformer,cors,key-auth,acl,basic-auth,request-termination,ip-restriction,post-function"),
		Env("KONG_NGINX_PROXY_PROXY_BUFFER_SIZE", "160k"),
		Env("KONG_NGINX_PROXY_PROXY_BUFFERS", "64 160k"),
		Env("KONG_PROXY_ACCESS_LOG", "/dev/stdout combined"),
		Env("SUPABASE_ANON_KEY", c.e("ANON_KEY")),
		Env("SUPABASE_SERVICE_KEY", c.e("SERVICE_ROLE_KEY")),
		Env("SUPABASE_PUBLISHABLE_KEY", c.e("SUPABASE_PUBLISHABLE_KEY")),
		Env("SUPABASE_SECRET_KEY", c.e("SUPABASE_SECRET_KEY")),
		Env("ANON_KEY_ASYMMETRIC", c.e("ANON_KEY_ASYMMETRIC")),
		Env("SERVICE_ROLE_KEY_ASYMMETRIC", c.e("SERVICE_ROLE_KEY_ASYMMETRIC")),
		Env("DASHBOARD_USERNAME", c.or("DASHBOARD_USERNAME", "supabase")),
		Env("DASHBOARD_PASSWORD", c.e("DASHBOARD_PASSWORD")),
		Entrypoint("/bin/sh", "/home/kong/kong-entrypoint.sh"),
	)...)
}

func (c stackConfig) caddy() AddServiceOpt {
	return keep("caddy", "supabase-caddy", "caddy:2",
		DependsOn(Name("api-gw"), Condition(Healthy)),
		DependsOn(Name("studio"), Condition(Healthy)),
		tcp(80, 80),
		tcp(443, 443),
		Publish(Protocol(UDP), HostPort(443), ContainerPort(443)),
		bind(c.file("volumes/proxy/caddy"), "/etc/caddy", false),
		Volume(Name("caddy_data"), Target("/data")),
		Volume(Name("caddy_config"), Target("/config")),
		Env("PROXY_DOMAIN", c.e("PROXY_DOMAIN")),
		Env("PROXY_AUTH_USERNAME", c.or("DASHBOARD_USERNAME", "supabase")),
		Env("PROXY_AUTH_PASSWORD", c.e("DASHBOARD_PASSWORD")),
		Command("/bin/sh", "-c", `PROXY_AUTH_PASSWORD=$(caddy hash-password --plaintext "$PROXY_AUTH_PASSWORD") && caddy run --config /etc/caddy/Caddyfile --adapter caddyfile`),
	)
}

func (c stackConfig) nginx() AddServiceOpt {
	return keep("nginx", "supabase-nginx", "jonasal/nginx-certbot:6.2.0-nginx1.31.8",
		DependsOn(Name("api-gw"), Condition(Healthy)),
		DependsOn(Name("studio"), Condition(Healthy)),
		tcp(80, 80),
		tcp(443, 443),
		bind(c.file("volumes/proxy/nginx/supabase-nginx.conf.tpl"), "/etc/nginx/supabase-nginx.conf.tpl", true),
		Volume(Name("nginx_letsencrypt"), Target("/etc/letsencrypt")),
		Env("PROXY_DOMAIN", c.e("PROXY_DOMAIN")),
		Env("CERTBOT_EMAIL", c.e("CERTBOT_EMAIL")),
		Env("PROXY_AUTH_USERNAME", c.or("DASHBOARD_USERNAME", "supabase")),
		Env("PROXY_AUTH_PASSWORD", c.e("DASHBOARD_PASSWORD")),
		Command("/bin/bash", "-c", `printf '%s:%s\n' "$PROXY_AUTH_USERNAME" "$(openssl passwd -apr1 "$PROXY_AUTH_PASSWORD")" > /etc/nginx/user_conf.d/dashboard-passwd && envsubst '$PROXY_DOMAIN' < /etc/nginx/supabase-nginx.conf.tpl > /etc/nginx/user_conf.d/nginx.conf && /scripts/start_nginx_certbot.sh`),
	)
}

func (c stackConfig) functions() AddServiceOpt {
	opts := []ServiceOption{
		DependsOn(Name("api-gw"), Condition(Healthy)),
		Memory(parseBytes(c.or("FUNCTIONS_MEMORY_LIMIT", "2g"))),
		probe([]string{"CMD-SHELL", "timeout 1 bash -c '</dev/tcp/127.0.0.1/9000'"}, 5*time.Second, 5*time.Second, 3, 0),
		bind(c.file("volumes/functions"), "/home/deno/functions", false),
		Volume(Name("deno-cache"), Target("/root/.cache/deno")),
		Env("JWT_SECRET", c.e("JWT_SECRET")),
		Env("SUPABASE_URL", "http://api-gw:8000"),
		Env("SUPABASE_PUBLIC_URL", c.e("SUPABASE_PUBLIC_URL")),
		Env("SUPABASE_ANON_KEY", c.e("ANON_KEY")),
		Env("SUPABASE_SERVICE_ROLE_KEY", c.e("SERVICE_ROLE_KEY")),
		Env("SUPABASE_PUBLISHABLE_KEYS", `{"default":"`+c.e("SUPABASE_PUBLISHABLE_KEY")+`"}`),
		Env("SUPABASE_SECRET_KEYS", `{"default":"`+c.e("SUPABASE_SECRET_KEY")+`"}`),
		Env("SUPABASE_DB_URL", "postgresql://postgres:"+c.e("POSTGRES_PASSWORD")+"@"+c.or("POSTGRES_HOST", "db")+":"+c.or("POSTGRES_PORT", "5432")+"/"+c.or("POSTGRES_DB", "postgres")),
		Env("VERIFY_JWT", c.or("FUNCTIONS_VERIFY_JWT", "false")),
		Command("start", "--user-worker-request-idle-timeout", "150000", "--main-service", "/home/deno/functions/main"),
		c.envIf("JWT_JWKS", "SUPABASE_JWKS"),
	}
	return keep("functions", "supabase-edge-functions", "supabase/edge-runtime:v1.77.9", opts...)
}

func (c stackConfig) supavisor() AddServiceOpt {
	return keep("supavisor", "supabase-pooler", "supabase/supavisor:2.9.13",
		DependsOn(Name("db"), Condition(Healthy)),
		tcp(c.port("POSTGRES_PORT", 5432), 5432),
		tcp(c.port("POOLER_PROXY_PORT_TRANSACTION", 6543), 6543),
		bind(c.file("volumes/pooler/pooler.exs"), "/etc/pooler/pooler.exs", true),
		probe([]string{"CMD", "curl", "-sSfL", "--head", "-o", "/dev/null", "http://127.0.0.1:4000/api/health"}, 10*time.Second, 5*time.Second, 10, 30*time.Second),
		Env("PORT", "4000"),
		Env("POSTGRES_PORT", c.or("POSTGRES_PORT", "5432")),
		Env("POSTGRES_HOST", c.or("POSTGRES_HOST", "db")),
		Env("POSTGRES_DB", c.or("POSTGRES_DB", "postgres")),
		Env("POSTGRES_PASSWORD", c.e("POSTGRES_PASSWORD")),
		Env("DATABASE_URL", "ecto://supabase_admin:"+c.e("POSTGRES_PASSWORD")+"@"+c.or("POSTGRES_HOST", "db")+":"+c.or("POSTGRES_PORT", "5432")+"/_supabase"),
		Env("CLUSTER_POSTGRES", "true"),
		Env("SECRET_KEY_BASE", c.e("SECRET_KEY_BASE")),
		Env("VAULT_ENC_KEY", c.e("VAULT_ENC_KEY")),
		Env("API_JWT_SECRET", c.e("JWT_SECRET")),
		Env("METRICS_JWT_SECRET", c.e("JWT_SECRET")),
		Env("REGION", "local"),
		Env("ERL_AFLAGS", "-proto_dist inet_tcp"),
		Env("POOLER_TENANT_ID", c.e("POOLER_TENANT_ID")),
		Env("POOLER_DEFAULT_POOL_SIZE", c.or("POOLER_DEFAULT_POOL_SIZE", "20")),
		Env("POOLER_MAX_CLIENT_CONN", c.or("POOLER_MAX_CLIENT_CONN", "100")),
		Env("POOLER_POOL_MODE", "transaction"),
		Env("DB_POOL_SIZE", c.or("POOLER_DB_POOL_SIZE", "5")),
		Command("/bin/sh", "-c", `/app/bin/migrate && /app/bin/supavisor eval "$(cat /etc/pooler/pooler.exs)" && /app/bin/server`),
	)
}

func (c stackConfig) pgbouncer() AddServiceOpt {
	return keep("pgbouncer", "supabase-pgbouncer", "edoburu/pgbouncer:v1.25.2-p0",
		DependsOn(Name("db"), Condition(Healthy)),
		tcp(c.port("POOLER_PROXY_PORT_TRANSACTION", 6543), 6432),
		probe([]string{"CMD", "pg_isready", "-h", "localhost", "-p", "6432"}, 10*time.Second, 5*time.Second, 3, 5*time.Second),
		Env("DB_USER", "pgbouncer"),
		Env("DB_PASSWORD", c.e("POSTGRES_PASSWORD")),
		Env("DB_HOST", c.or("POSTGRES_HOST", "db")),
		Env("DB_PORT", c.or("POSTGRES_PORT", "5432")),
		Env("DB_NAME", c.or("POSTGRES_DB", "postgres")),
		Env("LISTEN_PORT", "6432"),
		Env("AUTH_TYPE", "scram-sha-256"),
		Env("AUTH_QUERY", "SELECT * FROM pgbouncer.get_auth($1)"),
		Env("POOL_MODE", "transaction"),
		Env("DEFAULT_POOL_SIZE", c.or("POOLER_DEFAULT_POOL_SIZE", "20")),
		Env("MAX_CLIENT_CONN", c.or("POOLER_MAX_CLIENT_CONN", "100")),
		Env("ADMIN_USERS", "pgbouncer"),
		Env("STATS_USERS", "pgbouncer"),
	)
}

func (c stackConfig) minio() AddServiceOpt {
	return keep("minio", "supabase-minio", "cgr.dev/chainguard/minio",
		probe([]string{"CMD", "mc", "ready", "local"}, 2*time.Second, 10*time.Second, 5, 0),
		Volume(Name("minio-data"), Target("/data")),
		Env("MINIO_ROOT_USER", c.e("MINIO_ROOT_USER")),
		Env("MINIO_ROOT_PASSWORD", c.e("MINIO_ROOT_PASSWORD")),
		Command("server", "--console-address", ":9001", "/data"),
	)
}

func (c stackConfig) minioBucket() AddServiceOpt {
	return once("minio-createbucket", "supabase-minio-createbucket", "cgr.dev/chainguard/minio-client:latest-dev",
		DependsOn(Name("minio"), Condition(Healthy)),
		Env("MINIO_ROOT_USER", c.e("MINIO_ROOT_USER")),
		Env("MINIO_ROOT_PASSWORD", c.e("MINIO_ROOT_PASSWORD")),
		Env("GLOBAL_S3_BUCKET", c.or("GLOBAL_S3_BUCKET", "stub")),
		Entrypoint("/bin/sh", "-ec", `mc alias set supa-minio http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" && mc mb --ignore-existing "supa-minio/$GLOBAL_S3_BUCKET"`),
	)
}

func (c stackConfig) rustfs() AddServiceOpt {
	return keep("rustfs", "supabase-rustfs", "rustfs/rustfs:1.0.4",
		SecurityOpt("no-new-privileges:true"),
		probe([]string{"CMD", "curl", "-f", "http://rustfs:9000/health"}, 2*time.Second, 10*time.Second, 5, 0),
		Volume(Name("rustfs-data"), Target("/data")),
		Env("RUSTFS_ACCESS_KEY", c.e("MINIO_ROOT_USER")),
		Env("RUSTFS_SECRET_KEY", c.e("MINIO_ROOT_PASSWORD")),
		Env("RUSTFS_CONSOLE_ADDRESS", "0.0.0.0:9001"),
		Env("RUSTFS_CORS_ALLOWED_ORIGINS", "*"),
		Env("RUSTFS_CONSOLE_CORS_ALLOWED_ORIGINS", "*"),
	)
}

func (c stackConfig) rustfsBucket() AddServiceOpt {
	return once("rustfs-createbucket", "supabase-rustfs-createbucket", "rustfs/rc",
		DependsOn(Name("rustfs"), Condition(Healthy)),
		Env("MINIO_ROOT_USER", c.e("MINIO_ROOT_USER")),
		Env("MINIO_ROOT_PASSWORD", c.e("MINIO_ROOT_PASSWORD")),
		Env("GLOBAL_S3_BUCKET", c.or("GLOBAL_S3_BUCKET", "stub")),
		Entrypoint("/bin/sh", "-ec", `rc alias set supa-rustfs http://rustfs:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" && rc mb --ignore-existing "supa-rustfs/$GLOBAL_S3_BUCKET"`),
	)
}

func (c stackConfig) analytics() AddServiceOpt {
	return keep("analytics", "supabase-analytics", "supabase/logflare:1.50.15",
		DependsOn(Name("db"), Condition(Healthy)),
		probe([]string{"CMD-SHELL", "curl -sSfL -o /dev/null http://localhost:4000/health"}, 5*time.Second, 5*time.Second, 10, 30*time.Second),
		Env("LOGFLARE_NODE_HOST", "127.0.0.1"),
		Env("DB_USERNAME", "supabase_admin"),
		Env("DB_DATABASE", "_supabase"),
		Env("DB_HOSTNAME", c.or("POSTGRES_HOST", "db")),
		Env("DB_PORT", c.or("POSTGRES_PORT", "5432")),
		Env("DB_PASSWORD", c.e("POSTGRES_PASSWORD")),
		Env("DB_SCHEMA", "_analytics"),
		Env("LOGFLARE_SINGLE_TENANT", "true"),
		Env("LOGFLARE_SUPABASE_MODE", "true"),
		Env("LOGFLARE_PUBLIC_ACCESS_TOKEN", c.e("LOGFLARE_PUBLIC_ACCESS_TOKEN")),
		Env("LOGFLARE_PRIVATE_ACCESS_TOKEN", c.e("LOGFLARE_PRIVATE_ACCESS_TOKEN")),
		Env("LOGFLARE_FEATURE_FLAG_OVERRIDE", "multibackend=true"),
		Env("POSTGRES_BACKEND_URL", "postgresql://supabase_admin:"+c.e("POSTGRES_PASSWORD")+"@"+c.or("POSTGRES_HOST", "db")+":"+c.or("POSTGRES_PORT", "5432")+"/_supabase"),
		Env("POSTGRES_BACKEND_SCHEMA", "_analytics"),
	)
}

func (c stackConfig) vector() AddServiceOpt {
	sock := c.or("DOCKER_SOCKET_LOCATION", "/var/run/docker.sock")
	return keep("vector", "supabase-vector", "timberio/vector:0.53.0-alpine",
		DependsOn(Name("analytics"), Condition(Healthy)),
		SecurityOpt("label=disable"),
		probe([]string{"CMD-SHELL", "wget --no-verbose --tries=1 --spider http://vector:9001/health"}, 5*time.Second, 5*time.Second, 3, 0),
		bind(c.file("volumes/logs/vector.yml"), "/etc/vector/vector.yml", true),
		bind(sock, "/var/run/docker.sock", true),
		Env("LOGFLARE_PUBLIC_ACCESS_TOKEN", c.e("LOGFLARE_PUBLIC_ACCESS_TOKEN")),
		Command("--config", "/etc/vector/vector.yml"),
	)
}

func parseBytes(s string) int64 {
	s = strings.TrimSpace(strings.ToLower(s))
	mult := float64(1)
	switch {
	case strings.HasSuffix(s, "g"):
		mult = 1 << 30
		s = strings.TrimSuffix(s, "g")
	case strings.HasSuffix(s, "m"):
		mult = 1 << 20
		s = strings.TrimSuffix(s, "m")
	case strings.HasSuffix(s, "k"):
		mult = 1 << 10
		s = strings.TrimSuffix(s, "k")
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || n < 0 {
		return 2 << 30
	}
	return int64(n * mult)
}
