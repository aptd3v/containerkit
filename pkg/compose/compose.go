// package compose is a wrapper around the docker compose cli.
// it provides a programmatic interface for managing docker compose projects.
// it is not a complete 1:1 mapping of the docker compose cli, but provides a programmatic interface for managing
// docker compose projects in Go, making it useful for automation, tooling, or embedding docker compose behavior.
package compose

import (
	"fmt"
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

func (c *compose) Up() error {
	cmd, err := c.command("up")
	if err != nil {
		return NewComposeExecError(err)
	}

	fmt.Println(cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return NewComposeExecError(err)
	}

	return nil
}

func (c *compose) command(args ...string) (*exec.Cmd, error) {
	file, err := c.project.Marshal()
	if err != nil {
		return nil, NewComposeExecError(err)
	}
	cmd := exec.Command("docker", "compose", "-f", "-")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdin = strings.NewReader(string(file))
	return cmd, nil
}
