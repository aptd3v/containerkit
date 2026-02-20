// package compose is a wrapper around the docker compose cli.
// it provides a programmatic interface for managing docker compose projects.
// it is not a complete 1:1 mapping of the docker compose cli, but provides a programmatic interface for managing
// docker compose projects in Go, making it useful for automation, tooling, or embedding docker compose behavior.
package compose

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/compose-spec/compose-go/v2/types"
)

type compose struct {
	project *create.Project
}

func NewCompose(project *create.Project) *compose {
	return &compose{
		project: project,
	}
}

// SetComposeUpOption is a function that sets a ComposeUpOptions
type SetComposeUpOption func(*ComposeUpOptions) error

// SetComposeDownOption is a function that sets a ComposeDownOptions
type SetComposeDownOption func(*ComposeDownOptions) error

// SetComposeLogsOption is a function that sets a ComposeLogsOptions
type SetComposeLogsOption func(*ComposeLogsOptions) error

// SetComposeKillOption is a function that sets a ComposeKillOptions
type SetComposeKillOption func(*ComposeKillOptions) error

// SetComposePsOption is a function that sets a ComposePsOptions
type SetComposePsOption func(*ComposePsOptions) error

// SetComposeStartOption is a function that sets a ComposeStartOptions
type SetComposeStartOption func(*ComposeStartOptions) error

// SetComposeStopOption is a function that sets a ComposeStopOptions
type SetComposeStopOption func(*ComposeStopOptions) error

// SetComposeRestartOption is a function that sets a ComposeRestartOptions
type SetComposeRestartOption func(*ComposeRestartOptions) error

// SetComposeBuildOption is a function that sets a ComposeBuildOptions
type SetComposeBuildOption func(*ComposeBuildOptions) error

// SetComposePullOption is a function that sets a ComposePullOptions
type SetComposePullOption func(*ComposePullOptions) error

// SetComposeExecOption is a function that sets a ComposeExecOptions
type SetComposeExecOption func(*ComposeExecOptions) error

// Events runs the docker compose events command.
// Pass an empty service string to receive events for all services.
// When the project uses profiles, pass the same profiles used for up/down (e.g. Events(ctx, "", "minimal", "full")).
func (c *compose) Events(ctx context.Context, service string, profiles ...string) (<-chan Events, <-chan error, error) {
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

// Kill is a function that runs the docker compose kill command
// it returns an error if the command fails, or nil if the command succeeds
// it also returns nil if the context is canceled
func (c *compose) Kill(ctx context.Context, setters ...SetComposeKillOption) error {
	opt := &ComposeKillOptions{
		Flags:  []string{"kill"},
		Writer: os.Stdout,
	}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeKillError(err)
		}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeKillError(err)
	}
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposeKillError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

func (c *compose) Up(ctx context.Context, setters ...SetComposeUpOption) error {
	opt := &ComposeUpOptions{
		//default flags
		Flags:  []string{"up"},
		Writer: os.Stdout,
	}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeUpError(err)
		}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeUpError(err)
	}
	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposeUpError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

func (c *compose) Down(ctx context.Context, setters ...SetComposeDownOption) error {
	opt := &ComposeDownOptions{
		Flags:  []string{"down"},
		Writer: os.Stdout,
	}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeDownError(err)
		}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeDownError(err)
	}

	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposeDownError(err)
	}
	return handleContextCancellation(ctx, cmd.Run())
}

// Logs is a function that runs the docker compose logs command
// it returns an error if the command fails, or nil if the command succeeds
// it also returns nil if the context is canceled
func (c *compose) Logs(ctx context.Context, setters ...SetComposeLogsOption) error {
	opt := &ComposeLogsOptions{
		Flags:  []string{"logs"},
		Writer: os.Stdout,
	}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeLogsError(err)
		}
	}
	flags, err := opt.GenerateFlags()
	if err != nil {
		return NewComposeLogsError(err)
	}

	cmd, err := c.command(ctx, opt.Writer, flags, opt.Profiles, nil)
	if err != nil {
		return NewComposeLogsError(err)
	}

	return handleContextCancellation(ctx, cmd.Run())
}

// Ps runs the docker compose ps command.
func (c *compose) Ps(ctx context.Context, setters ...SetComposePsOption) error {
	opt := &ComposePsOptions{Writer: os.Stdout}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposePsError(err)
		}
	}
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
func (c *compose) Start(ctx context.Context, setters ...SetComposeStartOption) error {
	opt := &ComposeStartOptions{Writer: os.Stdout}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeStartError(err)
		}
	}
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
func (c *compose) Stop(ctx context.Context, setters ...SetComposeStopOption) error {
	opt := &ComposeStopOptions{Writer: os.Stdout}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeStopError(err)
		}
	}
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
func (c *compose) Restart(ctx context.Context, setters ...SetComposeRestartOption) error {
	opt := &ComposeRestartOptions{Writer: os.Stdout}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeRestartError(err)
		}
	}
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
func (c *compose) Build(ctx context.Context, setters ...SetComposeBuildOption) error {
	opt := &ComposeBuildOptions{Writer: os.Stdout}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeBuildError(err)
		}
	}
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
func (c *compose) Pull(ctx context.Context, setters ...SetComposePullOption) error {
	opt := &ComposePullOptions{Writer: os.Stdout}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposePullError(err)
		}
	}
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
// Service and Command must be set (e.g. WithService("web"), WithCommand("sh", "-c", "echo hi")).
// Use WithStdin(reader) to forward stdin to the container (e.g. os.Stdin for interactive).
func (c *compose) Exec(ctx context.Context, setters ...SetComposeExecOption) error {
	opt := &ComposeExecOptions{Writer: os.Stdout}
	for _, setter := range setters {
		if err := setter(opt); err != nil {
			return NewComposeExecError(err)
		}
	}
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

func (c *compose) command(ctx context.Context, writer io.Writer, args []string, profiles []string, stdin io.Reader) (*exec.Cmd, error) {
	file, err := c.project.Marshal()
	if err != nil {
		return nil, NewComposeError(err)
	}
	base := []string{"compose"}

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
			return nil, NewComposeError(fmt.Errorf("when using %s.WithProfiles, you must create a service with a profile via sc.WithProfiles", args[0]))
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
			return nil, NewComposeError(fmt.Errorf("when you have profiles, you must specify them via %s.WithProfiles\nprofiles found: %s", args[0], strings.Join(names, ", ")))
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
