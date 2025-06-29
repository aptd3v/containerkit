// package compose is a wrapper around the docker compose cli.
// it provides a programmatic interface for managing docker compose projects.
// it is not a complete 1:1 mapping of the docker compose cli, but provides a programmatic interface for managing
// docker compose projects in Go, making it useful for automation, tooling, or embedding docker compose behavior.
package compose

import (
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

type SetComposeUpOption func(*ComposeUpOptions) error

func (c *compose) Up(setters ...SetComposeUpOption) error {
	opt := &ComposeUpOptions{
		//default flags
		Flags: []string{"up"},
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

	cmd, err := c.command(flags)
	if err != nil {
		return NewComposeUpError(err)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return NewComposeUpError(err)
	}

	return nil
}

func (c *compose) command(args []string) (*exec.Cmd, error) {
	file, err := c.project.Marshal()
	if err != nil {
		return nil, NewComposeError(err)
	}
	cmd := exec.Command("docker", "compose", "-f", "-")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdin = strings.NewReader(string(file))
	return cmd, nil
}
