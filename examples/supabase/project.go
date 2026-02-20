// Package main: compose project assembly and service dependencies.
// Core services use sc.WithProfiles("minimal"); optional services use sc.WithProfiles("full").
// Up/Down/Logs must pass at least one profile (e.g. up.WithProfiles("minimal") or "minimal","full" for full stack).
// Resource limits are applied when enableResourceLimits is true.
package main

import (
	"path/filepath"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/sc"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource"
	"github.com/aptd3v/go-contain/pkg/tools"
)

// SetupProject builds the Supabase compose project with all services and dependencies.
// When enableResourceLimits is true, memory/CPU limits are applied to db, kong, and studio via deploy setters.
func SetupProject(cfg *SupabaseConfig, baseVolumesPath string, enableResourceLimits bool) *create.Project {
	project := create.NewProject("supabase")
	vol := func(p string) string { return filepath.Join(baseVolumesPath, p) }

	// Profiles: "minimal" = core services only; "full" = adds optional services. Compose requires at least one profile when any service has a profile.
	const profileMinimal = "minimal"
	const profileFull = "full"

	// vector: full profile only (no deps on other optional services)
	project.WithService("vector", vectorContainer(cfg, vol),
		sc.WithProfiles(profileFull))

	// db: core (minimal profile). Conditional resource limits.
	project.WithService("db", dbContainer(cfg, vol),
		sc.WithProfiles(profileMinimal),
		tools.WhenTrue(enableResourceLimits, sc.WithDeploy(deploy.WithResourceLimits(
			resource.WithMemoryBytes(2*1024*1024*1024), // 2GiB
			resource.WithNanoCPUs(2),                    // 2 CPUs (compose-go cpus is decimal, 0.01–10)
		))))

	// analytics: core (minimal)
	project.WithService("analytics", analyticsContainer(cfg),
		sc.WithProfiles(profileMinimal),
		sc.WithDependsOnHealthy("db"))

	// auth, rest: core (minimal)
	project.WithService("auth", authContainer(cfg),
		sc.WithProfiles(profileMinimal),
		sc.WithDependsOnHealthy("db"),
		sc.WithDependsOnHealthy("analytics"))
	project.WithService("rest", restContainer(cfg),
		sc.WithProfiles(profileMinimal),
		sc.WithDependsOnHealthy("db"),
		sc.WithDependsOnHealthy("analytics"))

	// realtime, meta, supavisor: full profile only
	project.WithService("realtime", realtimeContainer(cfg),
		sc.WithProfiles(profileFull),
		sc.WithDependsOnHealthy("db"),
		sc.WithDependsOnHealthy("analytics"))
	project.WithService("meta", metaContainer(cfg),
		sc.WithProfiles(profileFull),
		sc.WithDependsOnHealthy("db"),
		sc.WithDependsOnHealthy("analytics"))
	project.WithService("supavisor", supavisorContainer(cfg, vol),
		sc.WithProfiles(profileFull),
		sc.WithDependsOnHealthy("db"),
		sc.WithDependsOnHealthy("analytics"))

	// imgproxy: full profile only
	project.WithService("imgproxy", imgproxyContainer(cfg, vol),
		sc.WithProfiles(profileFull))

	// storage: full profile only
	project.WithService("storage", storageContainer(cfg, vol),
		sc.WithProfiles(profileFull),
		sc.WithDependsOnHealthy("db"),
		sc.WithDependsOn("rest"),
		sc.WithDependsOn("imgproxy"))

	// functions: full profile only
	project.WithService("functions", functionsContainer(cfg, vol),
		sc.WithProfiles(profileFull),
		sc.WithDependsOnHealthy("analytics"))

	// kong: core (minimal). Conditional resource limits.
	project.WithService("kong", kongContainer(cfg, vol),
		sc.WithProfiles(profileMinimal),
		sc.WithDependsOnHealthy("analytics"),
		tools.WhenTrue(enableResourceLimits, sc.WithDeploy(deploy.WithResourceLimits(
			resource.WithMemoryBytes(512*1024*1024), // 512MiB
			resource.WithNanoCPUs(1),                // 1 CPU (compose-go cpus is decimal, 0.01–10)
		))))

	// studio: core (minimal). Conditional resource limits.
	project.WithService("studio", studioContainer(cfg, vol),
		sc.WithProfiles(profileMinimal),
		sc.WithDependsOnHealthy("analytics"),
		tools.WhenTrue(enableResourceLimits, sc.WithDeploy(deploy.WithResourceLimits(
			resource.WithMemoryBytes(512*1024*1024), // 512MiB
			resource.WithNanoCPUs(1),                // 1 CPU (compose-go cpus is decimal, 0.01–10)
		))))

	project.WithVolume("db-config").WithVolume("deno-cache").WithNetwork("supabase-network")
	return project
}
