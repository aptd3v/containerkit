// Package main: compose project assembly and service dependencies.
// Core services use Service.Profiles("minimal"); optional services use Service.Profiles("full").
// Up/Down/Logs must pass at least one profile (e.g. Up.Profiles = []string{"minimal"} or "minimal","full" for full stack).
// Resource limits are applied when enableResourceLimits is true.
package main

import (
	"path/filepath"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

func SetupProject(cfg *SupabaseConfig, baseVolumesPath string, enableResourceLimits bool) *kit.Project {
	project := kit.NewProject("supabase")
	vol := func(p string) string { return filepath.Join(baseVolumesPath, p) }

	const profileMinimal = "minimal"
	const profileFull = "full"

	limits := func(mem uint64, cpus int64) kit.Deploy {
		return kit.Deploy{Limits: &types.Resources{
			MemoryBytes: types.ByteSize(mem),
			NanoCPUs:    types.NanoCPUs(cpus),
		}}
	}

	project.WithService(kit.NewService("vector", vectorContainer(cfg, vol)).Profiles(profileFull))

	db := kit.NewService("db", dbContainer(cfg, vol)).Profiles(profileMinimal)
	if enableResourceLimits {
		db.Deploy(limits(2*1024*1024*1024, 2))
	}
	project.WithService(db)

	project.WithService(kit.NewService("analytics", analyticsContainer(cfg)).
		Profiles(profileMinimal).
		DependsOnHealthy("db"))

	project.WithService(kit.NewService("auth", authContainer(cfg)).
		Profiles(profileMinimal).
		DependsOnHealthy("db").
		DependsOnHealthy("analytics"))
	project.WithService(kit.NewService("rest", restContainer(cfg)).
		Profiles(profileMinimal).
		DependsOnHealthy("db").
		DependsOnHealthy("analytics"))

	project.WithService(kit.NewService("realtime", realtimeContainer(cfg)).
		Profiles(profileFull).
		DependsOnHealthy("db").
		DependsOnHealthy("analytics"))
	project.WithService(kit.NewService("meta", metaContainer(cfg)).
		Profiles(profileMinimal).
		DependsOnHealthy("db").
		DependsOnHealthy("analytics"))
	project.WithService(kit.NewService("supavisor", supavisorContainer(cfg, vol)).
		Profiles(profileFull).
		DependsOnHealthy("db").
		DependsOnHealthy("analytics"))

	project.WithService(kit.NewService("imgproxy", imgproxyContainer(cfg, vol)).
		Profiles(profileFull))

	project.WithService(kit.NewService("storage", storageContainer(cfg, vol)).
		Profiles(profileFull).
		DependsOnHealthy("db").
		DependsOn("rest").
		DependsOn("imgproxy"))

	project.WithService(kit.NewService("functions", functionsContainer(cfg, vol)).
		Profiles(profileFull).
		DependsOnHealthy("analytics"))

	kong := kit.NewService("kong", kongContainer(cfg, vol)).
		Profiles(profileMinimal).
		DependsOnHealthy("analytics")
	if enableResourceLimits {
		kong.Deploy(limits(512*1024*1024, 1))
	}
	project.WithService(kong)

	studio := kit.NewService("studio", studioContainer(cfg, vol)).
		Profiles(profileMinimal).
		DependsOnHealthy("analytics").
		DependsOn("meta")
	if enableResourceLimits {
		studio.Deploy(limits(512*1024*1024, 1))
	}
	project.WithService(studio)

	project.WithVolume("db-config").WithVolume("deno-cache").WithNetwork("supabase-network")
	return project
}
