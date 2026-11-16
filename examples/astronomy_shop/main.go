// Astronomy Shop showcase: OpenTelemetry demo defined with the kit.
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
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
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
		yml, err := project.ExportYAML()
		if err != nil {
			log.Printf("warning: export to %s: %v", path, err)
		} else if err := os.WriteFile(path, yml, 0644); err != nil {
			log.Printf("warning: export to %s: %v", path, err)
		} else {
			fmt.Printf("exported compose project to %s\n", path)
		}
	}

	logger := NewLogger(OutputStream(os.Stdout))
	ctx := logger.Context(context.Background())
	fmt.Printf("profiles: %s\n", strings.Join(profiles, ","))
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
	if exists {
		fmt.Printf("probe image %s already exists\n", probeImage)
	} else {
		fmt.Printf("building probe image %s\n", probeImage)
	}
	if err := runFlow(ctx, cfg, project, profiles, exists); err != nil {
		log.Fatal(err)
	}
}

func runFollow(ctx context.Context, cfg *Config, project *Compose, profiles []string) error {
	cli, err := connectShop(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	active, err := withProfiles(project, profiles...)
	if err != nil {
		return err
	}
	api := cli.Composer()
	if err := active.Up(ctx, api,
		Detach(true),
		Wait(true),
		Timeout(900*time.Second),
		RemoveOrphans(true),
		Log(LoggerFrom(ctx)),
	); err != nil {
		return fmt.Errorf("up: %w", err)
	}
	printURLs(os.Stdout, cfg, profiles)

	runCtx, cancel := context.WithCancel(context.Background())
	go func() {
		waitInterrupt()
		cancel()
	}()

	go func() {
		err := active.Events(runCtx, api, EventConsumer(func(event Event) error {
			b, _ := json.Marshal(event)
			os.Stdout.WriteString("\033[36m" + string(b) + "\033[0m\n")
			return nil
		}))
		if err != nil && runCtx.Err() == nil {
			log.Printf("events: %v", err)
		}
	}()

	if err := active.Logs(runCtx, api, NewLogger(OutputStream(os.Stdout)), Follow(true)); err != nil && err != context.Canceled {
		log.Printf("logs: %v", err)
	}

	if err := active.Kill(context.Background(), api, StopSignal(SIGKILL), RemoveOrphans(true)); err != nil {
		log.Printf("kill: %v", err)
	}
	if err := active.Down(context.Background(), api, RemoveOrphans(true), Volumes(true)); err != nil {
		return fmt.Errorf("down: %w", err)
	}
	return nil
}

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
	if _, file, _, ok := runtime.Caller(0); ok {
		return filepath.Join(filepath.Dir(file), "volumes")
	}
	return "./volumes"
}

func connectShop(ctx context.Context) (*Client, error) {
	logger := LoggerFrom(ctx)
	if logger == nil {
		return Connect(FromEnv())
	}
	return Connect(FromEnv(), Log(logger))
}
