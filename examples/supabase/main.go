// Supabase bootstrap example: one-execute install of the Supabase stack using kit.
// Required volume files are auto-downloaded from the Supabase repo when missing.
// Streams compose events (start/stop/health) alongside logs; Ctrl+C kills containers and brings the stack down.
//
// Flags:
//
//	-profile: minimal (default) or full. Full adds vector, realtime, storage, imgproxy, functions, supavisor.
//	-resource-limits: apply memory/CPU limits to db, kong, and studio.
//	-volumes-path: directory for volume files (default: ./volumes or SUPABASE_VOLUMES_PATH).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
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
	exportPath := "./docker-compose.yaml"
	if err := project.Export(exportPath, 0644); err != nil {
		log.Printf("warning: export to %s: %v", exportPath, err)
	}

	supabase := kit.NewCompose(project)
	ctx := context.Background()
	profiles := []string{"minimal"}
	if profile == "full" {
		profiles = append(profiles, "full")
	}
	timeout := 5
	waitTimeout := 300
	if err := supabase.Up(ctx, &kit.Up{
		RemoveOrphans: true,
		Detach:        true,
		Timeout:       &timeout,
		WaitTimeout:   &waitTimeout,
		Profiles:      profiles,
	}); err != nil {
		log.Fatalf("up: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
	}()

	eventsCh, eventsErrCh, err := supabase.Events(ctx, &kit.EventsOpt{Profiles: profiles})
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

	if err := supabase.Logs(ctx, &kit.Logs{Follow: true, NoLogPrefix: true, Profiles: profiles}); err != nil && err != context.Canceled {
		log.Printf("logs: %v", err)
	}

	killCtx := context.Background()
	sig := "SIGKILL"
	if err := supabase.Kill(killCtx, &kit.Kill{Signal: &sig, RemoveOrphans: true, Profiles: profiles}); err != nil {
		log.Printf("kill: %v", err)
	}
	if err := supabase.Down(killCtx, &kit.Down{RemoveOrphans: true, RemoveVolumes: true, Profiles: profiles}); err != nil {
		log.Fatalf("down: %v", err)
	}
}
