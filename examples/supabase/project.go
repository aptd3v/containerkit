// Package main: compose project assembly and service dependencies.
// Core services use containerkit.Profiles("minimal"); optional services use containerkit.Profiles("full").
// Up/Down/Logs must pass at least one profile (e.g. Up.Profiles = []string{"minimal"} or "minimal","full" for full stack).
// Resource limits are applied when enableResourceLimits is true.
package main

import (
	"path/filepath"

	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func SetupProject(cfg *SupabaseConfig, baseVolumesPath string, enableResourceLimits bool) *containerkit.Project {
	project := containerkit.NewProject("supabase")
	vol := func(p string) string { return filepath.Join(baseVolumesPath, p) }

	const profileMinimal = "minimal"
	const profileFull = "full"

	limits := func(mem uint64, cpus int64) containerkit.Deploy {
		return containerkit.Deploy{Limits: &containerkit.Resources{MemoryBytes: mem, NanoCPUs: cpus}}
	}

	project.WithService("vector", vectorContainer(cfg, vol),
		containerkit.Profiles(profileFull))

	dbExtras := []any{containerkit.Profiles(profileMinimal)}
	if enableResourceLimits {
		dbExtras = append(dbExtras, limits(2*1024*1024*1024, 2))
	}
	project.WithService("db", dbContainer(cfg, vol), dbExtras...)

	project.WithService("analytics", analyticsContainer(cfg),
		containerkit.Profiles(profileMinimal),
		containerkit.DependsOnHealthy("db"))

	project.WithService("auth", authContainer(cfg),
		containerkit.Profiles(profileMinimal),
		containerkit.DependsOnHealthy("db"),
		containerkit.DependsOnHealthy("analytics"))
	project.WithService("rest", restContainer(cfg),
		containerkit.Profiles(profileMinimal),
		containerkit.DependsOnHealthy("db"),
		containerkit.DependsOnHealthy("analytics"))

	project.WithService("realtime", realtimeContainer(cfg),
		containerkit.Profiles(profileFull),
		containerkit.DependsOnHealthy("db"),
		containerkit.DependsOnHealthy("analytics"))
	project.WithService("meta", metaContainer(cfg),
		containerkit.Profiles(profileFull),
		containerkit.DependsOnHealthy("db"),
		containerkit.DependsOnHealthy("analytics"))
	project.WithService("supavisor", supavisorContainer(cfg, vol),
		containerkit.Profiles(profileFull),
		containerkit.DependsOnHealthy("db"),
		containerkit.DependsOnHealthy("analytics"))

	project.WithService("imgproxy", imgproxyContainer(cfg, vol),
		containerkit.Profiles(profileFull))

	project.WithService("storage", storageContainer(cfg, vol),
		containerkit.Profiles(profileFull),
		containerkit.DependsOnHealthy("db"),
		containerkit.DependsOn("rest"),
		containerkit.DependsOn("imgproxy"))

	project.WithService("functions", functionsContainer(cfg, vol),
		containerkit.Profiles(profileFull),
		containerkit.DependsOnHealthy("analytics"))

	kongExtras := []any{containerkit.Profiles(profileMinimal), containerkit.DependsOnHealthy("analytics")}
	if enableResourceLimits {
		kongExtras = append(kongExtras, limits(512*1024*1024, 1))
	}
	project.WithService("kong", kongContainer(cfg, vol), kongExtras...)

	studioExtras := []any{containerkit.Profiles(profileMinimal), containerkit.DependsOnHealthy("analytics")}
	if enableResourceLimits {
		studioExtras = append(studioExtras, limits(512*1024*1024, 1))
	}
	project.WithService("studio", studioContainer(cfg, vol), studioExtras...)

	project.WithVolume("db-config").WithVolume("deno-cache").WithNetwork("supabase-network")
	return project
}
