// Package main: Supabase service container definitions (vector, db, analytics, auth, rest, realtime, storage, imgproxy, meta, functions, kong, studio, supavisor).
package main

import (
	"fmt"

	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func vectorContainer(cfg *SupabaseConfig, vol func(string) string) *containerkit.Container {
	return containerkit.NewContainer("supabase-vector").
		Image("timberio/vector:0.53.0-alpine").
		EnvMap(envMapNonEmpty(cfg.envVector())).
		Command("--config", "/etc/vector/vector.yml").
		HealthCheck(containerkit.Health{
			Test:        []string{"CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:9001/health"},
			Timeout:     5,
			Interval:    5,
			Retries:     3,
			StartPeriod: 10,
		}).
		RestartUnlessStopped().
		VolumeBinds(vol("logs/vector.yml") + ":/etc/vector/vector.yml:ro,z").
		Mount(containerkit.Mount{
			Source: "/var/run/docker.sock", Target: "/var/run/docker.sock",
			Type: containerkit.MountBind, ReadOnly: true,
		}).
		SecurityOpts("label=disable").
		Endpoint("supabase-network")
}

func dbContainer(cfg *SupabaseConfig, vol func(string) string) *containerkit.Container {
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
	return containerkit.NewContainer("supabase-db").
		Image("supabase/postgres:15.8.1.085").
		EnvMap(envMapNonEmpty(env)).
		Command("postgres", "-c", "config_file=/etc/postgresql/postgresql.conf", "-c", "log_min_messages=fatal").
		HealthCheck(containerkit.Health{
			Test:        []string{"CMD", "pg_isready", "-U", "postgres", "-h", "localhost"},
			Interval:    5,
			Timeout:     5,
			Retries:     10,
			StartPeriod: 120,
		}).
		RestartUnlessStopped().
		VolumeBinds(binds...).
		RWNamedVolumeMount("db-config", "/etc/postgresql-custom").
		Endpoint("supabase-network")
}

func analyticsContainer(cfg *SupabaseConfig) *containerkit.Container {
	return containerkit.NewContainer("supabase-analytics").
		Image("supabase/logflare:1.31.2").
		EnvMap(envMapNonEmpty(cfg.envAnalytics())).
		HealthCheck(containerkit.Health{
			Test:     []string{"CMD", "curl", "http://localhost:4000/health"},
			Timeout:  5,
			Interval: 5,
			Retries:  10,
		}).
		RestartUnlessStopped().
		Endpoint("supabase-network")
}

func authContainer(cfg *SupabaseConfig) *containerkit.Container {
	return containerkit.NewContainer("supabase-auth").
		Image("supabase/gotrue:v2.186.0").
		EnvMap(envMapNonEmpty(cfg.envAuth())).
		HealthCheck(containerkit.Health{
			Test:     []string{"CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9999/health"},
			Timeout:  5,
			Interval: 5,
			Retries:  3,
		}).
		RestartUnlessStopped().
		Endpoint("supabase-network")
}

func restContainer(cfg *SupabaseConfig) *containerkit.Container {
	return containerkit.NewContainer("supabase-rest").
		Image("postgrest/postgrest:v14.5").
		EnvMap(envMapNonEmpty(cfg.envRest())).
		Command("postgrest").
		RestartUnlessStopped().
		Endpoint("supabase-network")
}

func realtimeContainer(cfg *SupabaseConfig) *containerkit.Container {
	healthTest := fmt.Sprintf("curl -sSfL --head -o /dev/null -H \"Authorization: Bearer %s\" http://localhost:4000/api/tenants/realtime-dev/health", cfg.AnonKey)
	return containerkit.NewContainer("realtime-dev.supabase-realtime").
		Image("supabase/realtime:v2.76.5").
		EnvMap(envMapNonEmpty(cfg.envRealtime())).
		HealthCheck(containerkit.Health{
			Test:        []string{"CMD-SHELL", healthTest},
			Timeout:     5,
			Interval:    30,
			Retries:     3,
			StartPeriod: 10,
		}).
		RestartUnlessStopped().
		Endpoint("supabase-network")
}

func storageContainer(cfg *SupabaseConfig, vol func(string) string) *containerkit.Container {
	return containerkit.NewContainer("supabase-storage").
		Image("supabase/storage-api:v1.37.8").
		EnvMap(envMapNonEmpty(cfg.envStorage())).
		HealthCheck(containerkit.Health{
			Test:     []string{"CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://storage:5000/status"},
			Timeout:  5,
			Interval: 5,
			Retries:  3,
		}).
		RestartUnlessStopped().
		VolumeBinds(vol("storage") + ":/var/lib/storage:z").
		Endpoint("supabase-network")
}

func imgproxyContainer(cfg *SupabaseConfig, vol func(string) string) *containerkit.Container {
	return containerkit.NewContainer("supabase-imgproxy").
		Image("darthsim/imgproxy:v3.30.1").
		EnvMap(envMapNonEmpty(cfg.envImgproxy())).
		HealthCheck(containerkit.Health{
			Test:     []string{"CMD", "imgproxy", "health"},
			Timeout:  5,
			Interval: 5,
			Retries:  3,
		}).
		RestartUnlessStopped().
		VolumeBinds(vol("storage") + ":/var/lib/storage:z").
		Endpoint("supabase-network")
}

func metaContainer(cfg *SupabaseConfig) *containerkit.Container {
	return containerkit.NewContainer("supabase-meta").
		Image("supabase/postgres-meta:v0.95.2").
		EnvMap(envMapNonEmpty(cfg.envMeta())).
		RestartUnlessStopped().
		Endpoint("supabase-network")
}

func functionsContainer(cfg *SupabaseConfig, vol func(string) string) *containerkit.Container {
	return containerkit.NewContainer("supabase-edge-functions").
		Image("supabase/edge-runtime:v1.70.3").
		EnvMap(envMapNonEmpty(cfg.envFunctions())).
		Command("start", "--main-service", "/home/deno/functions/main").
		RestartUnlessStopped().
		VolumeBinds(vol("functions")+":/home/deno/functions:Z").
		RWNamedVolumeMount("deno-cache", "/root/.cache/deno").
		Endpoint("supabase-network")
}

func kongContainer(cfg *SupabaseConfig, vol func(string) string) *containerkit.Container {
	entrypointScript := `eval "echo \"$(cat ~/temp.yml)\"" > ~/kong.yml && /docker-entrypoint.sh kong docker-start`
	return containerkit.NewContainer("supabase-kong").
		Image("kong:2.8.1").
		EnvMap(envMapNonEmpty(cfg.envKong())).
		Entrypoint("bash", "-c", entrypointScript).
		ExposedPort("tcp", "8000").
		ExposedPort("tcp", "8443").
		RestartUnlessStopped().
		PortBindings("tcp", "0.0.0.0", cfg.KongHTTPPort, "8000").
		PortBindings("tcp", "0.0.0.0", cfg.KongHTTPSPort, "8443").
		VolumeBinds(vol("api/kong.yml") + ":/home/kong/temp.yml:ro,z").
		Endpoint("supabase-network")
}

func studioContainer(cfg *SupabaseConfig, vol func(string) string) *containerkit.Container {
	return containerkit.NewContainer("supabase-studio").
		Image("supabase/studio:2026.02.16-sha-26c615c").
		EnvMap(envMapNonEmpty(cfg.envStudio())).
		ExposedPort("tcp", "3000").
		HealthCheck(containerkit.Health{
			Test:     []string{"CMD", "node", "-e", "fetch('http://studio:3000/api/platform/profile').then((r) => {if (r.status !== 200) throw new Error(r.status)})"},
			Timeout:  10,
			Interval: 5,
			Retries:  3,
		}).
		RestartUnlessStopped().
		PortBindings("tcp", "0.0.0.0", cfg.StudioPort, "3000").
		VolumeBinds(vol("snippets")+":/app/snippets:Z", vol("functions")+":/app/edge-functions:Z").
		Endpoint("supabase-network")
}

func supavisorContainer(cfg *SupabaseConfig, vol func(string) string) *containerkit.Container {
	script := `/app/bin/migrate && /app/bin/supavisor eval "$(cat /etc/pooler/pooler.exs)" && /app/bin/server`
	return containerkit.NewContainer("supabase-pooler").
		Image("supabase/supavisor:2.7.4").
		EnvMap(envMapNonEmpty(cfg.envSupavisor())).
		Entrypoint("/bin/sh", "-c", script).
		HealthCheck(containerkit.Health{
			Test:     []string{"CMD", "curl", "-sSfL", "--head", "-o", "/dev/null", "http://127.0.0.1:4000/api/health"},
			Interval: 10,
			Timeout:  5,
			Retries:  5,
		}).
		ExposedPort("tcp", "5432").
		ExposedPort("tcp", "6543").
		RestartUnlessStopped().
		PortBindings("tcp", "0.0.0.0", cfg.PostgresPort, "5432").
		PortBindings("tcp", "0.0.0.0", cfg.PoolerProxyPortTx, "6543").
		VolumeBinds(vol("pooler/pooler.exs") + ":/etc/pooler/pooler.exs:ro,z").
		Endpoint("supabase-network")
}
