package taskflow

import (
	"context"
	"fmt"

	"github.com/aptd3v/containerkit/pkg/kit"
	composeapi "github.com/docker/compose/v5/pkg/api"
)

// ComposeCmd runs one method of *kit.Compose. Construct it with Up, Down, and
// the other constructors in this file. The option slice is copied.
type ComposeCmd struct {
	call func(context.Context, *kit.Compose, kit.Composer) error
}

func command(fn func(context.Context, *kit.Compose, kit.Composer) error) ComposeCmd {
	return ComposeCmd{call: fn}
}

func copyOpts[T any](opts []T) []T {
	return append([]T(nil), opts...)
}

type discardLogs struct{}

func (discardLogs) Log(string, string)    { discardLogsSeen++ }
func (discardLogs) Err(string, string)    { discardLogsSeen++ }
func (discardLogs) Status(string, string) { discardLogsSeen++ }

var discardLogsSeen int

var _ composeapi.LogConsumer = discardLogs{}

func composeCommand(cmd any) (func(context.Context, *kit.Compose, kit.Composer) error, error) {
	switch c := cmd.(type) {
	case nil:
		return nil, fmt.Errorf("compose command is nil")
	case ComposeCmd:
		if c.call == nil {
			return nil, fmt.Errorf("compose command is nil")
		}
		return c.call, nil
	case func(context.Context, *kit.Compose) error:
		return func(ctx context.Context, app *kit.Compose, _ kit.Composer) error {
			return c(ctx, app)
		}, nil
	case func(context.Context, *kit.Compose, kit.Composer) error:
		return c, nil
	default:
		return nil, fmt.Errorf("unsupported compose command %T", cmd)
	}
}

func exitCode(code int, err error) error {
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("exit code %d", code)
	}
	return nil
}

// Up brings the project up.
func Up(opts ...kit.UpOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Up(ctx, api, opts...)
	})
}

// Down tears the project down.
func Down(opts ...kit.DownOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Down(ctx, api, opts...)
	})
}

// ComposeBuild builds images for the project.
func ComposeBuild(opts ...kit.ComposeBuildOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Build(ctx, api, opts...)
	})
}

// Create creates containers for the project.
func Create(opts ...kit.ComposeCreateOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Create(ctx, api, opts...)
	})
}

// Start starts the project's services.
func Start(opts ...kit.ComposeStartOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Start(ctx, api, opts...)
	})
}

// Stop stops the project's services.
func Stop(opts ...kit.ComposeStopOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Stop(ctx, api, opts...)
	})
}

// Restart restarts the project's services.
func Restart(opts ...kit.ComposeRestartOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Restart(ctx, api, opts...)
	})
}

// Pull pulls images for the project.
func Pull(opts ...kit.PullOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Pull(ctx, api, opts...)
	})
}

// Push pushes images for the project.
func Push(opts ...kit.PushOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Push(ctx, api, opts...)
	})
}

// Logs streams project logs. The consumer is discarded; progress follows the
// logger on the run context.
func Logs(opts ...kit.LogsOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Logs(ctx, api, discardLogs{}, opts...)
	})
}

// Ps lists project containers.
func Ps(opts ...kit.PsOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		_, err := app.Ps(ctx, api, opts...)
		return err
	})
}

// Kill signals project containers.
func Kill(opts ...kit.KillOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Kill(ctx, api, opts...)
	})
}

// Rm removes project containers.
func Rm(opts ...kit.RemoveOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Remove(ctx, api, opts...)
	})
}

// RunService starts a one-off service command. A non-zero exit fails the step.
func RunService(service string, opts ...kit.RunOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		code, err := app.Run(ctx, api, service, opts...)
		return exitCode(code, err)
	})
}

// Exec runs a command in a service container. A non-zero exit fails the step.
func Exec(service string, opts ...kit.RunOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		code, err := app.Exec(ctx, api, service, opts...)
		return exitCode(code, err)
	})
}

// Attach attaches to a service container.
func Attach(service string, opts ...kit.AttachOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Attach(ctx, api, service, opts...)
	})
}

// Copy copies files between the host and a service.
func Copy(opts ...kit.CopyOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Copy(ctx, api, opts...)
	})
}

// Pause pauses the project's services.
func Pause(opts ...kit.PauseOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Pause(ctx, api, opts...)
	})
}

// Unpause resumes the project's services.
func Unpause(opts ...kit.PauseOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Unpause(ctx, api, opts...)
	})
}

// Top lists processes in the named services.
func Top(services ...string) ComposeCmd {
	services = copyOpts(services)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		_, err := app.Top(ctx, api, services...)
		return err
	})
}

// Events streams project events.
func Events(opts ...kit.EventsOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Events(ctx, api, opts...)
	})
}

// Port prints the public address of a service port.
func Port(service string, port uint16, opts ...kit.ComposePortOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		_, _, err := app.Port(ctx, api, service, port, opts...)
		return err
	})
}

// PublishOCI publishes the project as an OCI artifact.
func PublishOCI(repository string, opts ...kit.PublishOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.PublishOCI(ctx, api, repository, opts...)
	})
}

// Images lists images used by the project.
func Images(opts ...kit.ImagesOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		_, err := app.Images(ctx, api, opts...)
		return err
	})
}

// Watch rebuilds and syncs the project.
func Watch(opts ...kit.WatchOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Watch(ctx, api, opts...)
	})
}

// Viz renders the project graph.
func Viz(opts ...kit.VizOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		_, err := app.Viz(ctx, api, opts...)
		return err
	})
}

// Wait waits for the project to exit.
func Wait(opts ...kit.WaitOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		_, err := app.Wait(ctx, api, opts...)
		return err
	})
}

// Scale changes service replica counts.
func Scale(opts ...kit.ScaleOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Scale(ctx, api, opts...)
	})
}

// Export copies a service container's filesystem out.
func Export(service string, opts ...kit.ExportOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Export(ctx, api, service, opts...)
	})
}

// Commit creates an image from a service container.
func Commit(service string, opts ...kit.CommitOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		return app.Commit(ctx, api, service, opts...)
	})
}

// VolumeList lists volumes for the project.
func VolumeList(opts ...kit.VolumesListOption) ComposeCmd {
	opts = copyOpts(opts)
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		_, err := app.VolumeList(ctx, api, opts...)
		return err
	})
}

// MaxConcurrency sets the Compose runtime's parallel operation limit.
func MaxConcurrency(n int) ComposeCmd {
	return command(func(_ context.Context, app *kit.Compose, api kit.Composer) error {
		return app.MaxConcurrency(api, n)
	})
}

// DryRun toggles Compose dry-run mode.
func DryRun(v bool) ComposeCmd {
	return command(func(ctx context.Context, app *kit.Compose, api kit.Composer) error {
		_, err := app.DryRunMode(ctx, api, v)
		return err
	})
}
