// Supabase bootstrap example: one-execute install of the Supabase stack using go-contain.
// Required volume files are auto-downloaded from the Supabase repo when missing.
// Streams compose events (start/stop/health) alongside logs; Ctrl+C kills containers and brings the stack down.
//
// Flags:
//   -profile: minimal (default) or full. Full adds vector, realtime, storage, imgproxy, meta, functions, supavisor.
//   -resource-limits: apply memory/CPU limits to db, kong, and studio.
//   -volumes-path: directory for volume files (default: ./volumes or SUPABASE_VOLUMES_PATH).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/down"
	"github.com/aptd3v/go-contain/pkg/compose/options/kill"
	"github.com/aptd3v/go-contain/pkg/compose/options/logs"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
)

func main() {
	profileFlag := flag.String("profile", "", "run mode: minimal (default) or full")
	resourceLimitsFlag := flag.Bool("resource-limits", false, "apply memory/CPU limits to db, kong, studio")
	volumesPathFlag := flag.String("volumes-path", "", "directory for volume files (default ./volumes or SUPABASE_VOLUMES_PATH)")
	flag.Parse()

	cfg := DefaultSupabaseConfig()
	baseVolumesPath := *volumesPathFlag
	if baseVolumesPath == "" {
		baseVolumesPath = os.Getenv("SUPABASE_VOLUMES_PATH")
	}
	if baseVolumesPath == "" {
		baseVolumesPath = "./volumes"
	}
	profile := strings.TrimSpace(*profileFlag)
	if profile == "" {
		profile = strings.TrimSpace(os.Getenv("SUPABASE_PROFILE"))
	}
	enableResourceLimits := *resourceLimitsFlag || os.Getenv("SUPABASE_RESOURCE_LIMITS") == "1"

	bootstrapSupabaseVolumes(baseVolumesPath)

	project := SetupProject(cfg, baseVolumesPath, enableResourceLimits)
	if err := project.Validate(); err != nil {
		log.Fatalf("project validate: %v", err)
	}
	exportPath := "./docker-compose.yaml" // write to repo root (run from repo root)
	if err := project.Export(exportPath, 0644); err != nil {
		log.Printf("warning: export to %s: %v", exportPath, err)
	}

	supabase := compose.NewCompose(project)
	ctx := context.Background()
	// Compose requires at least one profile when any service has a profile. Use same profiles for Up, Logs, and Down.
	profiles := []string{"minimal"}
	if profile == "full" {
		profiles = append(profiles, "full")
	}
	upOpts := []compose.SetComposeUpOption{
		up.WithRemoveOrphans(),
		up.WithDetach(),
		up.WithTimeout(5),
		up.WithWaitTimeout(300), // db init can take 1–2 min on first run; wait up to 5 min for healthy deps
		up.WithProfiles(profiles...),
	}
	if err := supabase.Up(ctx, upOpts...); err != nil {
		log.Fatalf("up: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
	}()

	// Stream real-time compose events (start, stop, health, etc.) for all services.
	eventsCh, eventsErrCh, err := supabase.Events(ctx, "", profiles...)
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

	logOpts := []compose.SetComposeLogsOption{logs.WithFollow(), logs.WithNoLogPrefix(), logs.WithProfiles(profiles...)}
	if err := supabase.Logs(ctx, logOpts...); err != nil && err != context.Canceled {
		log.Printf("logs: %v", err)
	}

	// On Ctrl+C, kill all containers (SIGKILL) then down so nothing is left running.
	killCtx := context.Background()
	if err := supabase.Kill(killCtx, kill.WithSignal("SIGKILL"), kill.WithRemoveOrphans(), kill.WithProfiles(profiles...)); err != nil {
		log.Printf("kill: %v", err)
	}
	downOpts := []compose.SetComposeDownOption{down.WithRemoveOrphans(), down.WithRemoveVolumes(), down.WithProfiles(profiles...)}
	if err := supabase.Down(killCtx, downOpts...); err != nil {
		log.Fatalf("down: %v", err)
	}
}
