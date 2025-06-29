// package compose is a wrapper around the docker compose cli.
// it provides a programmatic interface for managing docker compose projects.
// it is not a complete 1:1 mapping of the docker compose cli, but provides a programmatic interface for managing
// docker compose projects in Go, making it useful for automation, tooling, or embedding docker compose behavior.
package compose

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/aptd3v/go-contain/pkg/create"
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

	cmd, err := c.command(ctx, opt.Writer, flags)
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

	cmd, err := c.command(ctx, opt.Writer, flags)
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

	cmd, err := c.command(ctx, opt.Writer, flags)
	if err != nil {
		return NewComposeLogsError(err)
	}

	return handleContextCancellation(ctx, cmd.Run())
}
func (c *compose) command(ctx context.Context, writer io.Writer, args []string) (*exec.Cmd, error) {
	file, err := c.project.Marshal()
	if err != nil {
		return nil, NewComposeError(err)
	}
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "-")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdin = strings.NewReader(string(file))
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd, nil
}

func handleContextCancellation(ctx context.Context, err error) error {
	if errors.Is(err, context.Canceled) || ctx.Err() == context.Canceled {
		return nil
	}
	return err
}
