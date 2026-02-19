// Package main: Supabase service container definitions (vector, db, analytics, auth, rest, realtime, storage, imgproxy, meta, functions, kong, studio, supavisor).
package main

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
	"github.com/aptd3v/go-contain/pkg/create/config/nc"
)

func vectorContainer(cfg *SupabaseConfig, vol func(string) string) *create.Container {
	dockerSocket := "/var/run/docker.sock"
	return create.NewContainer("supabase-vector").
		WithContainerConfig(
			cc.WithImage("timberio/vector:0.53.0-alpine"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envVector())),
			cc.WithCommand("--config", "/etc/vector/vector.yml"),
			cc.WithHealthCheck(
				health.WithTest("CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:9001/health"),
				health.WithTimeout(5),
				health.WithInterval(5),
				health.WithRetries(3),
				health.WithStartPeriod(10),
			),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			hc.WithVolumeBinds(vol("logs/vector.yml")+":/etc/vector/vector.yml:ro,z"),
			hc.WithMountPoint(
				mount.WithSource(dockerSocket),
				mount.WithTarget("/var/run/docker.sock"),
				mount.WithType("bind"),
				mount.WithReadOnly(),
			),
			hc.WithSecurityOpts("label=disable"),
		).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func dbContainer(cfg *SupabaseConfig, vol func(string) string) *create.Container {
	env := cfg.envDB()
	binds := []string{
		vol("db/realtime.sql") + ":/docker-entrypoint-initdb.d/migrations/99-realtime.sql:ro,Z",
		vol("db/webhooks.sql") + ":/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql:ro,Z",
		vol("db/roles.sql") + ":/docker-entrypoint-initdb.d/init-scripts/99-roles.sql:ro,Z",
		vol("db/jwt.sql") + ":/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql:ro,Z",
		vol("db/data") + ":/var/lib/postgresql/data:Z",
		vol("db/_supabase.sql") + ":/docker-entrypoint-initdb.d/migrations/97-_supabase.sql:ro,Z",
		vol("db/logs.sql") + ":/docker-entrypoint-initdb.d/migrations/99-logs.sql:ro,Z",
		vol("db/pooler.sql") + ":/docker-entrypoint-initdb.d/migrations/99-pooler.sql:ro,Z",
	}
	return create.NewContainer("supabase-db").
		WithContainerConfig(
			cc.WithImage("supabase/postgres:15.8.1.085"),
			cc.WithEnvMap(envMapNonEmpty(env)),
			cc.WithCommand("postgres", "-c", "config_file=/etc/postgresql/postgresql.conf", "-c", "log_min_messages=fatal"),
			cc.WithHealthCheck(
				health.WithTest("CMD", "pg_isready", "-U", "postgres", "-h", "localhost"),
				health.WithInterval(5),
				health.WithTimeout(5),
				health.WithRetries(10),
				health.WithStartPeriod(120), // initdb runs many SQL scripts; allow time before health checks count
			),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			hc.WithVolumeBinds(binds...),
			hc.WithRWNamedVolumeMount("db-config", "/etc/postgresql-custom"),
		).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func analyticsContainer(cfg *SupabaseConfig) *create.Container {
	return create.NewContainer("supabase-analytics").
		WithContainerConfig(
			cc.WithImage("supabase/logflare:1.31.2"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envAnalytics())),
			cc.WithHealthCheck(
				health.WithTest("CMD", "curl", "http://localhost:4000/health"),
				health.WithTimeout(5),
				health.WithInterval(5),
				health.WithRetries(10),
			),
		).
		WithHostConfig(hc.WithRestartPolicyUnlessStopped()).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func authContainer(cfg *SupabaseConfig) *create.Container {
	return create.NewContainer("supabase-auth").
		WithContainerConfig(
			cc.WithImage("supabase/gotrue:v2.186.0"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envAuth())),
			cc.WithHealthCheck(
				health.WithTest("CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9999/health"),
				health.WithTimeout(5),
				health.WithInterval(5),
				health.WithRetries(3),
			),
		).
		WithHostConfig(hc.WithRestartPolicyUnlessStopped()).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func restContainer(cfg *SupabaseConfig) *create.Container {
	return create.NewContainer("supabase-rest").
		WithContainerConfig(
			cc.WithImage("postgrest/postgrest:v14.5"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envRest())),
			cc.WithCommand("postgrest"),
		).
		WithHostConfig(hc.WithRestartPolicyUnlessStopped()).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func realtimeContainer(cfg *SupabaseConfig) *create.Container {
	healthTest := fmt.Sprintf("curl -sSfL --head -o /dev/null -H \"Authorization: Bearer %s\" http://localhost:4000/api/tenants/realtime-dev/health", cfg.AnonKey)
	return create.NewContainer("realtime-dev.supabase-realtime").
		WithContainerConfig(
			cc.WithImage("supabase/realtime:v2.76.5"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envRealtime())),
			cc.WithHealthCheck(
				health.WithTest("CMD-SHELL", healthTest),
				health.WithTimeout(5),
				health.WithInterval(30),
				health.WithRetries(3),
				health.WithStartPeriod(10),
			),
		).
		WithHostConfig(hc.WithRestartPolicyUnlessStopped()).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func storageContainer(cfg *SupabaseConfig, vol func(string) string) *create.Container {
	return create.NewContainer("supabase-storage").
		WithContainerConfig(
			cc.WithImage("supabase/storage-api:v1.37.8"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envStorage())),
			cc.WithHealthCheck(
				health.WithTest("CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://storage:5000/status"),
				health.WithTimeout(5),
				health.WithInterval(5),
				health.WithRetries(3),
			),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			hc.WithVolumeBinds(vol("storage")+":/var/lib/storage:z"),
		).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func imgproxyContainer(cfg *SupabaseConfig, vol func(string) string) *create.Container {
	return create.NewContainer("supabase-imgproxy").
		WithContainerConfig(
			cc.WithImage("darthsim/imgproxy:v3.30.1"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envImgproxy())),
			cc.WithHealthCheck(
				health.WithTest("CMD", "imgproxy", "health"),
				health.WithTimeout(5),
				health.WithInterval(5),
				health.WithRetries(3),
			),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			hc.WithVolumeBinds(vol("storage")+":/var/lib/storage:z"),
		).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func metaContainer(cfg *SupabaseConfig) *create.Container {
	return create.NewContainer("supabase-meta").
		WithContainerConfig(
			cc.WithImage("supabase/postgres-meta:v0.95.2"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envMeta())),
		).
		WithHostConfig(hc.WithRestartPolicyUnlessStopped()).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func functionsContainer(cfg *SupabaseConfig, vol func(string) string) *create.Container {
	return create.NewContainer("supabase-edge-functions").
		WithContainerConfig(
			cc.WithImage("supabase/edge-runtime:v1.70.3"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envFunctions())),
			cc.WithCommand("start", "--main-service", "/home/deno/functions/main"),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			hc.WithVolumeBinds(vol("functions")+":/home/deno/functions:Z"),
			hc.WithRWNamedVolumeMount("deno-cache", "/root/.cache/deno"),
		).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func kongContainer(cfg *SupabaseConfig, vol func(string) string) *create.Container {
	entrypointScript := `eval "echo \"$(cat ~/temp.yml)\"" > ~/kong.yml && /docker-entrypoint.sh kong docker-start`
	return create.NewContainer("supabase-kong").
		WithContainerConfig(
			cc.WithImage("kong:2.8.1"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envKong())),
			cc.WithEntrypoint("bash", "-c", entrypointScript),
			cc.WithExposedPort("tcp", "8000"),
			cc.WithExposedPort("tcp", "8443"),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			hc.WithPortBindings("tcp", "0.0.0.0", cfg.KongHTTPPort, "8000"),
			hc.WithPortBindings("tcp", "0.0.0.0", cfg.KongHTTPSPort, "8443"),
			hc.WithVolumeBinds(vol("api/kong.yml")+":/home/kong/temp.yml:ro,z"),
		).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func studioContainer(cfg *SupabaseConfig, vol func(string) string) *create.Container {
	return create.NewContainer("supabase-studio").
		WithContainerConfig(
			cc.WithImage("supabase/studio:2026.02.16-sha-26c615c"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envStudio())),
			cc.WithExposedPort("tcp", "3000"),
			cc.WithHealthCheck(
				health.WithTest("CMD", "node", "-e", "fetch('http://studio:3000/api/platform/profile').then((r) => {if (r.status !== 200) throw new Error(r.status)})"),
				health.WithTimeout(10),
				health.WithInterval(5),
				health.WithRetries(3),
			),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			hc.WithPortBindings("tcp", "0.0.0.0", cfg.StudioPort, "3000"),
			hc.WithVolumeBinds(vol("snippets")+":/app/snippets:Z", vol("functions")+":/app/edge-functions:Z"),
		).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}

func supavisorContainer(cfg *SupabaseConfig, vol func(string) string) *create.Container {
	script := `/app/bin/migrate && /app/bin/supavisor eval "$(cat /etc/pooler/pooler.exs)" && /app/bin/server`
	return create.NewContainer("supabase-pooler").
		WithContainerConfig(
			cc.WithImage("supabase/supavisor:2.7.4"),
			cc.WithEnvMap(envMapNonEmpty(cfg.envSupavisor())),
			cc.WithEntrypoint("/bin/sh", "-c", script),
			cc.WithHealthCheck(
				health.WithTest("CMD", "curl", "-sSfL", "--head", "-o", "/dev/null", "http://127.0.0.1:4000/api/health"),
				health.WithInterval(10),
				health.WithTimeout(5),
				health.WithRetries(5),
			),
			cc.WithExposedPort("tcp", "5432"),
			cc.WithExposedPort("tcp", "6543"),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			hc.WithPortBindings("tcp", "0.0.0.0", cfg.PostgresPort, "5432"),
			hc.WithPortBindings("tcp", "0.0.0.0", cfg.PoolerProxyPortTx, "6543"),
			hc.WithVolumeBinds(vol("pooler/pooler.exs")+":/etc/pooler/pooler.exs:ro,z"),
		).
		WithNetworkConfig(nc.WithEndpoint("supabase-network"))
}
