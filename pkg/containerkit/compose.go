// Package containerkit is a wrapper around the docker compose CLI.
// It provides a programmatic interface for managing docker compose projects
// by mapping docker compose commands onto typed option structs.
package containerkit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

// Compose runs docker compose commands against a Project (YAML on stdin).
type Compose struct {
	project *Project
}

// NewCompose returns a Compose runner for the given project.
func NewCompose(project *Project) *Compose {
	return &Compose{
		project: project,
	}
}

func defaultWriter(w io.Writer) io.Writer {
	if w == nil {
		return os.Stdout
	}
	return w
}

func extraArgs(parts ...[]string) []string {
	var out []string
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func asComposeErr[T error](fn func(error) T) func(error) error {
	return func(err error) error { return fn(err) }
}

func (c *Compose) runCompose(ctx context.Context, writer io.Writer, args []string, profiles []string, stdin io.Reader, wrap func(error) error) error {
	writer = defaultWriter(writer)
	cmd, err := c.command(ctx, writer, args, profiles, stdin)
	if err != nil {
		return wrap(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Events runs the docker compose events command.
// Pass an empty service string to receive events for all services.
// When the project uses profiles, pass the same profiles used for up/down (e.g. Events(ctx, "", "minimal", "full")).
func (c *Compose) Events(ctx context.Context, service string, profiles ...string) (<-chan Events, <-chan error, error) {
	eventsCh := make(chan Events, 1)
	errCh := make(chan error, 1)

	writer := newEventsWriter(ctx, eventsCh, errCh)
	args := []string{"events", "--json"}
	if service != "" {
		args = append(args, service)
	}
	cmd, err := c.command(ctx, writer, args, profiles, nil)
	if err != nil {
		return nil, nil, NewComposeEventsError(err)
	}

	go func() {
		defer close(eventsCh)
		defer close(errCh)

		if err := handleContextCancellation(ctx, cmd.Run()); err != nil {
			errCh <- NewComposeEventsError(err)
		}
	}()

	return eventsCh, errCh, nil
}

// Kill runs the docker compose kill command.
// A nil opt uses defaults (writer is os.Stdout).
func (c *Compose) Kill(ctx context.Context, opt *Kill) error {
	if opt == nil {
		opt = &Kill{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeKillError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeKillError))
}

// Up runs the docker compose up command.
// A nil opt uses defaults (writer is os.Stdout).
func (c *Compose) Up(ctx context.Context, opt *Up) error {
	if opt == nil {
		opt = &Up{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeUpError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeUpError))
}

// Down runs the docker compose down command.
// A nil opt uses defaults (writer is os.Stdout).
func (c *Compose) Down(ctx context.Context, opt *Down) error {
	if opt == nil {
		opt = &Down{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeDownError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeDownError))
}

// Logs runs the docker compose logs command.
// A nil opt uses defaults (writer is os.Stdout).
func (c *Compose) Logs(ctx context.Context, opt *Logs) error {
	if opt == nil {
		opt = &Logs{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeLogsError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeLogsError))
}

// Ps runs the docker compose ps command.
func (c *Compose) Ps(ctx context.Context, opt *Ps) error {
	if opt == nil {
		opt = &Ps{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposePsError(err)
	}
	flags = append(flags, opt.ServiceNames...)
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposePsError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Start runs the docker compose start command.
func (c *Compose) Start(ctx context.Context, opt *Start) error {
	if opt == nil {
		opt = &Start{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeStartError(err)
	}
	flags = append(flags, opt.ServiceNames...)
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposeStartError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Stop runs the docker compose stop command.
func (c *Compose) Stop(ctx context.Context, opt *Stop) error {
	if opt == nil {
		opt = &Stop{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeStopError(err)
	}
	flags = append(flags, opt.ServiceNames...)
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposeStopError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Restart runs the docker compose restart command.
func (c *Compose) Restart(ctx context.Context, opt *Restart) error {
	if opt == nil {
		opt = &Restart{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeRestartError(err)
	}
	flags = append(flags, opt.ServiceNames...)
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposeRestartError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Build runs the docker compose build command.
func (c *Compose) Build(ctx context.Context, opt *ComposeBuild) error {
	if opt == nil {
		opt = &ComposeBuild{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeBuildError(err)
	}
	flags = append(flags, opt.Flags...)
	flags = append(flags, opt.ServiceNames...)
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposeBuildError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Pull runs the docker compose pull command.
func (c *Compose) Pull(ctx context.Context, opt *Pull) error {
	if opt == nil {
		opt = &Pull{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposePullError(err)
	}
	flags = append(flags, opt.Flags...)
	flags = append(flags, opt.ServiceNames...)
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposePullError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Exec runs the docker compose exec command.
// Service and Command must be set on opt.
func (c *Compose) Exec(ctx context.Context, opt *Exec) error {
	if opt == nil {
		opt = &Exec{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Service == "" {
		return NewComposeExecError(fmt.Errorf("service is required"))
	}
	if len(opt.Command) == 0 {
		return NewComposeExecError(fmt.Errorf("command is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeExecError(err)
	}
	flags = append(flags, opt.Service)
	flags = append(flags, opt.Command...)
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, opt.Stdin)
	if err != nil {
		return NewComposeExecError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Attach runs the docker compose attach command. Service is required.
func (c *Compose) Attach(ctx context.Context, opt *Attach) error {
	if opt == nil {
		opt = &Attach{}
	}
	if opt.Service == "" {
		return NewComposeAttachError(fmt.Errorf("service is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeAttachError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, []string{opt.Service})...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, opt.Stdin, asComposeErr(NewComposeAttachError))
}

// Commit runs the docker compose commit command. Service is required.
func (c *Compose) Commit(ctx context.Context, opt *Commit) error {
	if opt == nil {
		opt = &Commit{}
	}
	if opt.Service == "" {
		return NewComposeCommitError(fmt.Errorf("service is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeCommitError(err)
	}
	pos := []string{opt.Service}
	if opt.Repository != "" {
		pos = append(pos, opt.Repository)
	}
	flags = append(flags, extraArgs(opt.Flags, pos)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeCommitError))
}

// Config runs the docker compose config command.
func (c *Compose) Config(ctx context.Context, opt *ComposeConfig) error {
	if opt == nil {
		opt = &ComposeConfig{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeConfigError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeConfigError))
}

// Cp runs the docker compose cp command. Source and Dest are required.
func (c *Compose) Cp(ctx context.Context, opt *Cp) error {
	if opt == nil {
		opt = &Cp{}
	}
	if opt.Source == "" || opt.Dest == "" {
		return NewComposeCpError(fmt.Errorf("source and dest are required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeCpError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, []string{opt.Source, opt.Dest})...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, opt.Stdin, asComposeErr(NewComposeCpError))
}

// Create runs the docker compose create command.
func (c *Compose) Create(ctx context.Context, opt *Create) error {
	if opt == nil {
		opt = &Create{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeCreateError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeCreateError))
}

// Export runs the docker compose export command. Service is required.
func (c *Compose) Export(ctx context.Context, opt *ComposeExport) error {
	if opt == nil {
		opt = &ComposeExport{}
	}
	if opt.Service == "" {
		return NewComposeExportError(fmt.Errorf("service is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeExportError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, []string{opt.Service})...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeExportError))
}

// Images runs the docker compose images command.
func (c *Compose) Images(ctx context.Context, opt *Images) error {
	if opt == nil {
		opt = &Images{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeImagesError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeImagesError))
}

// Ls runs the docker compose ls command.
func (c *Compose) Ls(ctx context.Context, opt *Ls) error {
	if opt == nil {
		opt = &Ls{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeLsError(err)
	}
	flags = append(flags, opt.Flags...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeLsError))
}

// Pause runs the docker compose pause command.
func (c *Compose) Pause(ctx context.Context, opt *Pause) error {
	if opt == nil {
		opt = &Pause{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposePauseError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposePauseError))
}

// Port runs the docker compose port command. Service and PrivatePort are required.
func (c *Compose) Port(ctx context.Context, opt *Port) error {
	if opt == nil {
		opt = &Port{}
	}
	if opt.Service == "" {
		return NewComposePortError(fmt.Errorf("service is required"))
	}
	if opt.PrivatePort <= 0 {
		return NewComposePortError(fmt.Errorf("private port is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposePortError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, []string{opt.Service, strconv.Itoa(opt.PrivatePort)})...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposePortError))
}

// Publish runs the docker compose publish command. Repository is required.
func (c *Compose) Publish(ctx context.Context, opt *Publish) error {
	if opt == nil {
		opt = &Publish{}
	}
	if opt.Repository == "" {
		return NewComposePublishError(fmt.Errorf("repository is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposePublishError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, []string{opt.Repository})...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposePublishError))
}

// Push runs the docker compose push command.
func (c *Compose) Push(ctx context.Context, opt *Push) error {
	if opt == nil {
		opt = &Push{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposePushError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposePushError))
}

// Rm runs the docker compose rm command.
func (c *Compose) Rm(ctx context.Context, opt *Rm) error {
	if opt == nil {
		opt = &Rm{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeRmError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeRmError))
}

// Run runs the docker compose run command. Service is required.
func (c *Compose) Run(ctx context.Context, opt *Run) error {
	if opt == nil {
		opt = &Run{}
	}
	if opt.Service == "" {
		return NewComposeRunError(fmt.Errorf("service is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeRunError(err)
	}
	pos := append([]string{opt.Service}, opt.Command...)
	flags = append(flags, extraArgs(opt.Flags, pos)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, opt.Stdin, asComposeErr(NewComposeRunError))
}

// Scale runs the docker compose scale command. Replicas is required.
func (c *Compose) Scale(ctx context.Context, opt *Scale) error {
	if opt == nil {
		opt = &Scale{}
	}
	if len(opt.Replicas) == 0 {
		return NewComposeScaleError(fmt.Errorf("at least one service replica is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeScaleError(err)
	}
	var reps []string
	for _, s := range opt.Replicas {
		if s.Service == "" || s.Num < 0 {
			return NewComposeScaleError(fmt.Errorf("service name required and scale must be non-negative"))
		}
		reps = append(reps, fmt.Sprintf("%s=%d", s.Service, s.Num))
	}
	flags = append(flags, extraArgs(opt.Flags, reps)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeScaleError))
}

// Stats runs the docker compose stats command.
func (c *Compose) Stats(ctx context.Context, opt *Stats) error {
	if opt == nil {
		opt = &Stats{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeStatsError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeStatsError))
}

// Top runs the docker compose top command.
func (c *Compose) Top(ctx context.Context, opt *Top) error {
	if opt == nil {
		opt = &Top{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeTopError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeTopError))
}

// Unpause runs the docker compose unpause command.
func (c *Compose) Unpause(ctx context.Context, opt *Unpause) error {
	if opt == nil {
		opt = &Unpause{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeUnpauseError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeUnpauseError))
}

// Version runs the docker compose version command.
func (c *Compose) Version(ctx context.Context, opt *Version) error {
	if opt == nil {
		opt = &Version{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeVersionError(err)
	}
	flags = append(flags, opt.Flags...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeVersionError))
}

// Volumes runs the docker compose volumes command.
func (c *Compose) Volumes(ctx context.Context, opt *Volumes) error {
	if opt == nil {
		opt = &Volumes{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeVolumesError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeVolumesError))
}

// Wait runs the docker compose wait command. ServiceNames is required.
func (c *Compose) Wait(ctx context.Context, opt *Wait) error {
	if opt == nil {
		opt = &Wait{}
	}
	if len(opt.ServiceNames) == 0 {
		return NewComposeWaitError(fmt.Errorf("at least one service is required"))
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeWaitError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeWaitError))
}

// Watch runs the docker compose watch command.
func (c *Compose) Watch(ctx context.Context, opt *Watch) error {
	if opt == nil {
		opt = &Watch{}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeWatchError(err)
	}
	flags = append(flags, extraArgs(opt.Flags, opt.ServiceNames)...)
	return c.runCompose(ctx, opt.Writer, flags, opt.Profiles, nil, asComposeErr(NewComposeWatchError))
}

func (c *Compose) command(ctx context.Context, writer io.Writer, args []string, profiles []string, stdin io.Reader) (*exec.Cmd, error) {
	file, err := c.project.Marshal()
	if err != nil {
		return nil, NewComposeError(err)
	}
	base := []string{"compose"}

	cmdName := "compose"
	if len(args) > 0 {
		cmdName = args[0]
	}

	// if profiles args are passed, we need to check if any services have profiles
	// if they do not, we need to return an error
	if profiles != nil {
		pmap := map[string]struct{}{}
		c.project.ForEachService(func(name string, service *types.ServiceConfig) error {
			if len(service.Profiles) > 0 {
				for _, profile := range service.Profiles {
					pmap[profile] = struct{}{}
				}
			}
			return nil
		})
		if len(pmap) == 0 {
			return nil, NewComposeError(fmt.Errorf("when using %s.Profiles, you must create a service with a profile via containerkit.Profiles", cmdName))
		}
	}
	// if no profiles args are passed, we need to check if any services have profiles
	// if they do, we need to return an error
	if len(profiles) == 0 {
		pmap := map[string]struct{}{}

		c.project.ForEachService(func(name string, service *types.ServiceConfig) error {
			if len(service.Profiles) > 0 {
				for _, profile := range service.Profiles {
					pmap[profile] = struct{}{}
				}
			}
			return nil
		})
		if len(pmap) > 0 {
			names := []string{}
			for profile := range pmap {
				names = append(names, profile)
			}
			return nil, NewComposeError(fmt.Errorf("when you have profiles, you must specify them via %s.Profiles\nprofiles found: %s", cmdName, strings.Join(names, ", ")))
		}
	}
	// if profiles are set, add them to the base command
	if len(profiles) > 0 {
		for _, profile := range profiles {
			base = append(base, "--profile", profile)
		}
	}

	// for file passed via stdin, we need to add the -f flag
	base = append(base, "-f", "-")
	cmd := exec.CommandContext(ctx, "docker", base...)
	cmd.Args = append(cmd.Args, args...)
	fileReader := strings.NewReader(string(file))
	if stdin != nil {
		cmd.Stdin = io.MultiReader(fileReader, stdin)
	} else {
		cmd.Stdin = fileReader
	}
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd, nil
}

func handleContextCancellation(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return nil
	}
	if ctx.Err() == context.Canceled {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil
	}
	return err
}
