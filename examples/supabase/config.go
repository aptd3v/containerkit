// Package main: Supabase stack configuration and per-service environment maps.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// SupabaseConfig holds env and ports for the Supabase stack. Defaults are for local dev.
type SupabaseConfig struct {
	PostgresPort              string
	PostgresHost              string
	PostgresDB                string
	PostgresPassword          string
	JWTSecret                 string
	JWTExpiry                 string
	AnonKey                   string
	ServiceRoleKey            string
	PGMetaCryptoKey           string
	KongHTTPPort              string
	KongHTTPSPort             string
	DashboardUser             string
	DashboardPass             string
	APIExternalURL            string
	SiteURL                   string
	SupabasePublicURL         string
	StudioPort                string
	StudioDefaultOrg          string
	StudioDefaultProj         string
	OpenAIAPIKey              string
	LogflarePublic            string
	LogflarePrivate           string
	SecretKeyBase             string
	VaultEncKey               string
	DisableSignup             string
	EnableEmailSignup         string
	EnableAnonymous           string
	EnableEmailAutoconfirm    string
	EnablePhoneSignup         string
	EnablePhoneAutoconfirm    string
	SMTPAdminEmail            string
	SMTPHost                  string
	SMTPPort                  string
	SMTPUser                  string
	SMTPPass                  string
	SMTPSenderName            string
	MailerPathsInvite         string
	MailerPathsConfirmation   string
	MailerPathsRecovery       string
	MailerPathsEmailChange    string
	AdditionalRedirectURLs    string
	PGRSTDBSchemas            string
	StorageTenantID           string
	Region                    string
	GlobalS3Bucket            string
	S3ProtocolAccessKeyID     string
	S3ProtocolAccessKeySecret string
	IMGProxyEnableWebP        string
	FunctionsVerifyJWT        string
	PoolerProxyPortTx         string
	PoolerTenantID            string
	PoolerDefaultPoolSize     string
	PoolerMaxClientConn       string
	PoolerDBPoolSize          string
}

// DefaultSupabaseConfig returns a config with local dev defaults. Generates secrets if empty.
func DefaultSupabaseConfig() *SupabaseConfig {
	c := &SupabaseConfig{
		PostgresPort:           "5432",
		PostgresHost:           "db",
		PostgresDB:             "postgres",
		PostgresPassword:       "postgres",
		JWTExpiry:              "3600",
		KongHTTPPort:           "8000",
		KongHTTPSPort:          "8443",
		DashboardUser:          "kong",
		DashboardPass:          "kong",
		APIExternalURL:         "http://localhost:8000",
		SiteURL:                "http://localhost:3000",
		SupabasePublicURL:      "http://localhost:8000",
		StudioPort:             "3000",
		StudioDefaultOrg:       "Default Organization",
		StudioDefaultProj:      "Default Project",
		DisableSignup:          "false",
		EnableEmailSignup:      "true",
		EnableAnonymous:        "true",
		EnableEmailAutoconfirm: "true",
		EnablePhoneSignup:      "false",
		EnablePhoneAutoconfirm: "false",
		PGRSTDBSchemas:         "public,storage,graphql_public",
		StorageTenantID:        "default",
		Region:                 "local",
		GlobalS3Bucket:         "supabase-storage",
		IMGProxyEnableWebP:     "true",
		FunctionsVerifyJWT:     "true",
		PoolerProxyPortTx:      "6543",
		PoolerTenantID:         "default",
		PoolerDefaultPoolSize:  "20",
		PoolerMaxClientConn:    "100",
		PoolerDBPoolSize:       "10",
	}
	if c.JWTSecret == "" {
		c.JWTSecret = mustRandBase64(32)
	}
	if c.AnonKey == "" {
		c.AnonKey = mustRandBase64(32)
	}
	if c.ServiceRoleKey == "" {
		c.ServiceRoleKey = mustRandBase64(32)
	}
	if c.PGMetaCryptoKey == "" {
		c.PGMetaCryptoKey = mustRandBase64(32)
	}
	if c.LogflarePublic == "" {
		c.LogflarePublic = mustRandBase64(16)
	}
	if c.LogflarePrivate == "" {
		c.LogflarePrivate = mustRandBase64(16)
	}
	if c.SecretKeyBase == "" {
		c.SecretKeyBase = mustRandBase64(64)
	}
	if c.VaultEncKey == "" {
		c.VaultEncKey = mustRandBase64(32)
	}
	return c
}

func mustRandBase64(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:n]
}

func (c *SupabaseConfig) envStudio() map[string]string {
	return map[string]string{
		"HOSTNAME":                         "::",
		"STUDIO_PG_META_URL":               "http://meta:8080",
		"POSTGRES_PORT":                    c.PostgresPort,
		"POSTGRES_HOST":                    c.PostgresHost,
		"POSTGRES_DB":                      c.PostgresDB,
		"POSTGRES_PASSWORD":                c.PostgresPassword,
		"PG_META_CRYPTO_KEY":               c.PGMetaCryptoKey,
		"DEFAULT_ORGANIZATION_NAME":        c.StudioDefaultOrg,
		"DEFAULT_PROJECT_NAME":             c.StudioDefaultProj,
		"OPENAI_API_KEY":                   c.OpenAIAPIKey,
		"SUPABASE_URL":                     "http://kong:8000",
		"SUPABASE_PUBLIC_URL":              c.SupabasePublicURL,
		"SUPABASE_ANON_KEY":                c.AnonKey,
		"SUPABASE_SERVICE_KEY":             c.ServiceRoleKey,
		"AUTH_JWT_SECRET":                  c.JWTSecret,
		"LOGFLARE_API_KEY":                 c.LogflarePublic,
		"LOGFLARE_PUBLIC_ACCESS_TOKEN":     c.LogflarePublic,
		"LOGFLARE_PRIVATE_ACCESS_TOKEN":    c.LogflarePrivate,
		"LOGFLARE_URL":                     "http://analytics:4000",
		"NEXT_PUBLIC_ENABLE_LOGS":          "true",
		"NEXT_ANALYTICS_BACKEND_PROVIDER":  "postgres",
		"SNIPPETS_MANAGEMENT_FOLDER":       "/app/snippets",
		"EDGE_FUNCTIONS_MANAGEMENT_FOLDER": "/app/edge-functions",
	}
}

func (c *SupabaseConfig) envKong() map[string]string {
	return map[string]string{
		"KONG_DATABASE":                      "off",
		"KONG_DECLARATIVE_CONFIG":            "/home/kong/kong.yml",
		"KONG_DNS_ORDER":                     "LAST,A,CNAME",
		"KONG_PLUGINS":                       "request-transformer,cors,key-auth,acl,basic-auth,request-termination,ip-restriction",
		"KONG_NGINX_PROXY_PROXY_BUFFER_SIZE": "160k",
		"KONG_NGINX_PROXY_PROXY_BUFFERS":     "64 160k",
		"SUPABASE_ANON_KEY":                  c.AnonKey,
		"SUPABASE_SERVICE_KEY":               c.ServiceRoleKey,
		"DASHBOARD_USERNAME":                 c.DashboardUser,
		"DASHBOARD_PASSWORD":                 c.DashboardPass,
	}
}

func (c *SupabaseConfig) envAuth() map[string]string {
	return map[string]string{
		"GOTRUE_API_HOST":                         "0.0.0.0",
		"GOTRUE_API_PORT":                         "9999",
		"API_EXTERNAL_URL":                        c.APIExternalURL,
		"GOTRUE_DB_DRIVER":                        "postgres",
		"GOTRUE_DB_DATABASE_URL":                  fmt.Sprintf("postgres://supabase_auth_admin:%s@%s:%s/%s", c.PostgresPassword, c.PostgresHost, c.PostgresPort, c.PostgresDB),
		"GOTRUE_SITE_URL":                         c.SiteURL,
		"GOTRUE_URI_ALLOW_LIST":                   c.AdditionalRedirectURLs,
		"GOTRUE_DISABLE_SIGNUP":                   c.DisableSignup,
		"GOTRUE_JWT_ADMIN_ROLES":                  "service_role",
		"GOTRUE_JWT_AUD":                          "authenticated",
		"GOTRUE_JWT_DEFAULT_GROUP_NAME":           "authenticated",
		"GOTRUE_JWT_EXP":                          c.JWTExpiry,
		"GOTRUE_JWT_SECRET":                       c.JWTSecret,
		"GOTRUE_EXTERNAL_EMAIL_ENABLED":           c.EnableEmailSignup,
		"GOTRUE_EXTERNAL_ANONYMOUS_USERS_ENABLED": c.EnableAnonymous,
		"GOTRUE_MAILER_AUTOCONFIRM":               c.EnableEmailAutoconfirm,
		"GOTRUE_SMTP_ADMIN_EMAIL":                 c.SMTPAdminEmail,
		"GOTRUE_SMTP_HOST":                        c.SMTPHost,
		"GOTRUE_SMTP_PORT":                        c.SMTPPort,
		"GOTRUE_SMTP_USER":                        c.SMTPUser,
		"GOTRUE_SMTP_PASS":                        c.SMTPPass,
		"GOTRUE_SMTP_SENDER_NAME":                 c.SMTPSenderName,
		"GOTRUE_MAILER_URLPATHS_INVITE":           c.MailerPathsInvite,
		"GOTRUE_MAILER_URLPATHS_CONFIRMATION":     c.MailerPathsConfirmation,
		"GOTRUE_MAILER_URLPATHS_RECOVERY":         c.MailerPathsRecovery,
		"GOTRUE_MAILER_URLPATHS_EMAIL_CHANGE":     c.MailerPathsEmailChange,
		"GOTRUE_EXTERNAL_PHONE_ENABLED":           c.EnablePhoneSignup,
		"GOTRUE_SMS_AUTOCONFIRM":                  c.EnablePhoneAutoconfirm,
	}
}

func (c *SupabaseConfig) envRest() map[string]string {
	return map[string]string{
		"PGRST_DB_URI":                  fmt.Sprintf("postgres://authenticator:%s@%s:%s/%s", c.PostgresPassword, c.PostgresHost, c.PostgresPort, c.PostgresDB),
		"PGRST_DB_SCHEMAS":              c.PGRSTDBSchemas,
		"PGRST_DB_ANON_ROLE":            "anon",
		"PGRST_JWT_SECRET":              c.JWTSecret,
		"PGRST_DB_USE_LEGACY_GUCS":      "false",
		"PGRST_APP_SETTINGS_JWT_SECRET": c.JWTSecret,
		"PGRST_APP_SETTINGS_JWT_EXP":    c.JWTExpiry,
	}
}

func (c *SupabaseConfig) envRealtime() map[string]string {
	return map[string]string{
		"PORT":                        "4000",
		"DB_HOST":                     c.PostgresHost,
		"DB_PORT":                     c.PostgresPort,
		"DB_USER":                     "supabase_admin",
		"DB_PASSWORD":                 c.PostgresPassword,
		"DB_NAME":                     c.PostgresDB,
		"DB_AFTER_CONNECT_QUERY":      "SET search_path TO _realtime",
		"DB_ENC_KEY":                  "supabaserealtime",
		"API_JWT_SECRET":              c.JWTSecret,
		"SECRET_KEY_BASE":             c.SecretKeyBase,
		"ERL_AFLAGS":                  "-proto_dist inet_tcp",
		"DNS_NODES":                   "''",
		"RLIMIT_NOFILE":               "10000",
		"APP_NAME":                    "realtime",
		"SEED_SELF_HOST":              "true",
		"RUN_JANITOR":                 "true",
		"DISABLE_HEALTHCHECK_LOGGING": "true",
	}
}

func (c *SupabaseConfig) envStorage() map[string]string {
	return map[string]string{
		"ANON_KEY":                       c.AnonKey,
		"SERVICE_KEY":                    c.ServiceRoleKey,
		"POSTGREST_URL":                  "http://rest:3000",
		"PGRST_JWT_SECRET":               c.JWTSecret,
		"DATABASE_URL":                   fmt.Sprintf("postgres://supabase_storage_admin:%s@%s:%s/%s", c.PostgresPassword, c.PostgresHost, c.PostgresPort, c.PostgresDB),
		"REQUEST_ALLOW_X_FORWARDED_PATH": "true",
		"FILE_SIZE_LIMIT":                "52428800",
		"STORAGE_BACKEND":                "file",
		"GLOBAL_S3_BUCKET":               c.GlobalS3Bucket,
		"FILE_STORAGE_BACKEND_PATH":      "/var/lib/storage",
		"TENANT_ID":                      c.StorageTenantID,
		"REGION":                         c.Region,
		"ENABLE_IMAGE_TRANSFORMATION":    "true",
		"IMGPROXY_URL":                   "http://imgproxy:5001",
		"S3_PROTOCOL_ACCESS_KEY_ID":      c.S3ProtocolAccessKeyID,
		"S3_PROTOCOL_ACCESS_KEY_SECRET":  c.S3ProtocolAccessKeySecret,
	}
}

func (c *SupabaseConfig) envImgproxy() map[string]string {
	return map[string]string{
		"IMGPROXY_BIND":                  ":5001",
		"IMGPROXY_LOCAL_FILESYSTEM_ROOT": "/",
		"IMGPROXY_USE_ETAG":              "true",
		"IMGPROXY_ENABLE_WEBP_DETECTION": c.IMGProxyEnableWebP,
		"IMGPROXY_MAX_SRC_RESOLUTION":    "16.8",
	}
}

func (c *SupabaseConfig) envMeta() map[string]string {
	return map[string]string{
		"PG_META_PORT":        "8080",
		"PG_META_DB_HOST":     c.PostgresHost,
		"PG_META_DB_PORT":     c.PostgresPort,
		"PG_META_DB_NAME":     c.PostgresDB,
		"PG_META_DB_USER":     "supabase_admin",
		"PG_META_DB_PASSWORD": c.PostgresPassword,
		"CRYPTO_KEY":          c.PGMetaCryptoKey,
	}
}

func (c *SupabaseConfig) envFunctions() map[string]string {
	return map[string]string{
		"JWT_SECRET":                c.JWTSecret,
		"SUPABASE_URL":              "http://kong:8000",
		"SUPABASE_ANON_KEY":         c.AnonKey,
		"SUPABASE_SERVICE_ROLE_KEY": c.ServiceRoleKey,
		"SUPABASE_DB_URL":           fmt.Sprintf("postgresql://postgres:%s@%s:%s/%s", c.PostgresPassword, c.PostgresHost, c.PostgresPort, c.PostgresDB),
		"VERIFY_JWT":                c.FunctionsVerifyJWT,
	}
}

func (c *SupabaseConfig) envAnalytics() map[string]string {
	return map[string]string{
		"LOGFLARE_NODE_HOST":             "127.0.0.1",
		"DB_USERNAME":                    "supabase_admin",
		"DB_DATABASE":                    "_supabase",
		"DB_HOSTNAME":                    c.PostgresHost,
		"DB_PORT":                        c.PostgresPort,
		"DB_PASSWORD":                    c.PostgresPassword,
		"DB_SCHEMA":                      "_analytics",
		"LOGFLARE_PUBLIC_ACCESS_TOKEN":   c.LogflarePublic,
		"LOGFLARE_PRIVATE_ACCESS_TOKEN":  c.LogflarePrivate,
		"LOGFLARE_SINGLE_TENANT":         "true",
		"LOGFLARE_SUPABASE_MODE":         "true",
		"POSTGRES_BACKEND_URL":           fmt.Sprintf("postgresql://supabase_admin:%s@%s:%s/_supabase", c.PostgresPassword, c.PostgresHost, c.PostgresPort),
		"POSTGRES_BACKEND_SCHEMA":        "_analytics",
		"LOGFLARE_FEATURE_FLAG_OVERRIDE": "multibackend=true",
	}
}

func (c *SupabaseConfig) envDB() map[string]string {
	return map[string]string{
		"POSTGRES_HOST":     "/var/run/postgresql",
		"PGPORT":            c.PostgresPort,
		"POSTGRES_PORT":     c.PostgresPort,
		"PGPASSWORD":        c.PostgresPassword,
		"POSTGRES_PASSWORD": c.PostgresPassword,
		"PGDATABASE":        c.PostgresDB,
		"POSTGRES_DB":       c.PostgresDB,
		"JWT_SECRET":        c.JWTSecret,
		"JWT_EXP":           c.JWTExpiry,
	}
}

func (c *SupabaseConfig) envVector() map[string]string {
	return map[string]string{
		"LOGFLARE_PUBLIC_ACCESS_TOKEN": c.LogflarePublic,
	}
}

func (c *SupabaseConfig) envSupavisor() map[string]string {
	return map[string]string{
		"PORT":                     "4000",
		"POSTGRES_PORT":            c.PostgresPort,
		"POSTGRES_DB":              c.PostgresDB,
		"POSTGRES_PASSWORD":        c.PostgresPassword,
		"DATABASE_URL":             fmt.Sprintf("ecto://supabase_admin:%s@%s:%s/_supabase", c.PostgresPassword, c.PostgresHost, c.PostgresPort),
		"CLUSTER_POSTGRES":         "true",
		"SECRET_KEY_BASE":          c.SecretKeyBase,
		"VAULT_ENC_KEY":            c.VaultEncKey,
		"API_JWT_SECRET":           c.JWTSecret,
		"METRICS_JWT_SECRET":       c.JWTSecret,
		"REGION":                   "local",
		"ERL_AFLAGS":               "-proto_dist inet_tcp",
		"POOLER_TENANT_ID":         c.PoolerTenantID,
		"POOLER_DEFAULT_POOL_SIZE": c.PoolerDefaultPoolSize,
		"POOLER_MAX_CLIENT_CONN":   c.PoolerMaxClientConn,
		"POOLER_POOL_MODE":         "transaction",
		"DB_POOL_SIZE":             c.PoolerDBPoolSize,
	}
}

// envMapNonEmpty returns a copy of m with empty values omitted so cc.WithEnvMap accepts it.
func envMapNonEmpty(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if v != "" {
			out[k] = v
		}
	}
	return out
}
