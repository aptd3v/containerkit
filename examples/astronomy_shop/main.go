// Astronomy Shop showcase: OpenTelemetry demo defined entirely in Go with kit.
// Pulls published ghcr.io/open-telemetry/demo images, builds a small kit-probe
// via Dockerfile-as-Go, then runs a taskflow Build -> Up -> probe DAG and stays
// up until Ctrl+C. -follow also streams compose events and logs.
//
// Flags:
//
//	-profile: comma-separated shop,full,observability (default shop,observability)
//	-follow:  after Up, stream Events+Logs until Ctrl+C
//	-volumes-path: host dir for bind-mounted configs (defaults to <package>/volumes)
//	-export:  write resolved compose YAML (default ./astronomy-shop.compose.yaml)
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func main() {
	profileFlag := flag.String("profile", "", "comma-separated profiles: shop, full, observability (default shop,observability)")
	followFlag := flag.Bool("follow", false, "after Up, stream events and logs until Ctrl+C")
	volumesPathFlag := flag.String("volumes-path", "", "directory for bind-mounted configs (default <package>/volumes)")
	exportFlag := flag.String("export", "./astronomy-shop.compose.yaml", "write resolved compose YAML (empty to skip)")
	flag.Parse()

	cfg := DefaultConfig()
	baseVolumesPath := resolveVolumesPath(*volumesPathFlag)

	rawProfile := strings.TrimSpace(*profileFlag)
	if rawProfile == "" {
		rawProfile = strings.TrimSpace(os.Getenv("ASTRONOMY_SHOP_PROFILE"))
	}
	profiles := parseProfiles(rawProfile)

	bootstrapVolumes(baseVolumesPath)
	project := SetupProject(cfg, baseVolumesPath, profiles)
	if err := project.Validate(); err != nil {
		log.Fatalf("project validate: %v", err)
	}
	if path := strings.TrimSpace(*exportFlag); path != "" {
		if err := project.Export(path, 0644); err != nil {
			log.Printf("warning: export to %s: %v", path, err)
		} else {
			fmt.Printf("exported compose project to %s\n", path)
		}
	}

	ctx := context.Background()
	if *followFlag {
		if err := runFollow(ctx, cfg, project, profiles); err != nil {
			log.Fatal(err)
		}
		return
	}

	exists, err := imageExists(ctx, probeImage)
	if err != nil {
		log.Printf("warning: image inspect: %v (will rebuild)", err)
		exists = false
	}
	if err := runFlow(ctx, cfg, project, profiles, exists); err != nil {
		log.Fatal(err)
	}
}

func runFollow(ctx context.Context, cfg *Config, project *kit.Project, profiles []string) error {
	app := kit.NewCompose(project)
	waitTimeout := 900
	timeout := 5
	if err := app.Up(ctx, &kit.Up{
		Detach:        true,
		Wait:          true,
		WaitTimeout:   &waitTimeout,
		Timeout:       &timeout,
		RemoveOrphans: true,
		Profiles:      profiles,
		Writer:        os.Stdout,
	}); err != nil {
		return fmt.Errorf("up: %w", err)
	}
	printURLs(os.Stdout, cfg, profiles)

	runCtx, cancel := context.WithCancel(context.Background())
	go func() {
		waitInterrupt()
		cancel()
	}()

	eventsCh, eventsErrCh, err := app.Events(runCtx, &kit.EventsOpt{Profiles: profiles})
	if err != nil {
		log.Printf("events: %v", err)
	} else {
		go func() {
			const cyan, reset = "\033[36m", "\033[0m"
			for e := range eventsCh {
				b, _ := json.Marshal(e)
				os.Stdout.WriteString(cyan + string(b) + reset + "\n")
			}
		}()
		go func() {
			for err := range eventsErrCh {
				if err != nil {
					log.Printf("events err: %v", err)
				}
			}
		}()
	}

	if err := app.Logs(runCtx, &kit.Logs{Follow: true, NoLogPrefix: true, Profiles: profiles}); err != nil && err != context.Canceled {
		log.Printf("logs: %v", err)
	}

	killCtx := context.Background()
	sig := "SIGKILL"
	if err := app.Kill(killCtx, &kit.Kill{Signal: &sig, RemoveOrphans: true, Profiles: profiles}); err != nil {
		log.Printf("kill: %v", err)
	}
	if err := app.Down(killCtx, &kit.Down{RemoveOrphans: true, RemoveVolumes: true, Profiles: profiles}); err != nil {
		return fmt.Errorf("down: %w", err)
	}
	return nil
}

// resolveVolumesPath returns an absolute path for the volumes directory.
// Priority: explicit flag > env var > package-relative default (./volumes next to the example).
// Explicit values (flag/env) are resolved relative to the current working directory.
// The package-relative default is computed from runtime.Caller so the demo
// works regardless of the directory it is run from.
func resolveVolumesPath(explicit string) string {
	p := strings.TrimSpace(explicit)
	if p == "" {
		p = os.Getenv("ASTRONOMY_SHOP_VOLUMES_PATH")
	}
	if p != "" {
		abs, err := filepath.Abs(p)
		if err == nil {
			return abs
		}
		return p
	}

	// Package-relative default: <package-dir>/volumes
	// This makes "go run ./examples/astronomy_shop/" work from any cwd.
	if _, file, _, ok := runtime.Caller(0); ok {
		pkgDir := filepath.Dir(file)
		return filepath.Join(pkgDir, "volumes")
	}

	// Last resort (should never be reached in normal builds)
	return "./volumes"
}
