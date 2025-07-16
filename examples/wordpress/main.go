package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/down"
	"github.com/aptd3v/go-contain/pkg/compose/options/logs"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
	"github.com/aptd3v/go-contain/pkg/create/config/nc"
	"github.com/aptd3v/go-contain/pkg/create/config/sc"
	"github.com/aptd3v/go-contain/pkg/tools"
)

var (
	IsLinux      = runtime.GOOS == "linux"
	IsNotWindows = runtime.GOOS != "windows"
	NumWordPress = 3
)

func main() {

	project := SetupProject()
	//export the project to a docker-compose.yaml file
	err := project.Export("./examples/wordpress/docker-compose.yaml", 0644)
	if err != nil {
		log.Fatalf("failed to export to docker-compose.yaml: %v", err)
	}

	//create a new compose instance
	wordpress := compose.NewCompose(project)

	//execute the up command
	err = wordpress.Up(context.Background(),
		up.WithWriter(NewLogger("up")),
		up.WithRemoveOrphans(),
		up.WithNoLogPrefix(),
		up.WithDetach(),
		up.WithTimeout(3),
	)
	if err != nil {
		log.Fatalf("failed to execute up: %v", err)
	}

	//create a new context with a cancel function
	ctx, cancel := context.WithCancel(context.Background())
	// wait for ctrl+c to cancel the context
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	go func() {
		<-ctrlc
		cancel()
	}()

	//execute the logs command
	err = wordpress.Logs(ctx,
		logs.WithWriter(NewLogger("logs")),
		logs.WithNoLogPrefix(),
		logs.WithFollow(),
	)
	if err != nil {
		log.Fatalf("failed to execute logs: %v", err)
	}

	//cleanup
	err = wordpress.Down(
		//we use background context because we want the down command to run even if the context is canceled,
		context.Background(),
		down.WithWriter(NewLogger("down")),
		down.WithRemoveOrphans(),
		down.WithRemoveVolumes(),
	)
	if err != nil {
		log.Fatalf("failed to execute down: %v", err)
	}
}

func SetupProject() *create.Project {
	project := create.NewProject(fmt.Sprintf("gocontain-wp-scale-%d", NumWordPress))
	project.WithService("database-example", DatabaseContainer())

	// portainer involes extra steps on windows so we skip it
	if IsNotWindows {
		project.WithService("portainer-container", PortainerContainer())
	}
	deps := []create.SetServiceConfig{}
	// Creates x amount of separate service entries with individual names.
	// NOTE: This differs from scaling a single service.
	//
	// If you want to scale a single service, you can use the following code:
	// project.WithService("wordpress-example",
	// 	WordPressContainer(),
	// 	sc.WithDependsOnHealthy("database-example"),
	// 	sc.WithDeploy(
	// 		deploy.WithReplicas(5),
	// 	),
	// )
	services := []string{}
	for i := 1; i <= NumWordPress; i++ {
		serviceName := fmt.Sprintf("wordpress-example-%d", i)
		services = append(services, serviceName)
		project.WithService(serviceName,
			WordPressContainer(),
			sc.WithDependsOnHealthy("database-example"),
			//dependancy chain so each service depends on the previous one 1<-2<-3
			tools.WhenTrueElse(i > 1,
				sc.WithDependsOn(fmt.Sprintf("wordpress-example-%d", i-1)),
				nil,
			),
		)
		//add to deps so we can use it in the proxy container
		//this means that the proxy container will wait for all the wordpress containers to be created to start
		deps = append(deps, sc.WithDependsOnHealthy(fmt.Sprintf("wordpress-example-%d", i)))
	}
	//generate dynamic haproxy config
	err := GenerateHAProxyConfig(services)
	if err != nil {
		log.Fatalf("failed to generate haproxy.cfg: %v", err)
	}

	project.WithService("proxy-container", ProxyContainer(), deps...)

	project.
		WithVolume("wordpress-data").
		WithVolume("database-data").
		WithVolume("portainer-data").
		WithNetwork("wordpress-network")

	return project
}

func WordPressContainer() *create.Container {
	return create.NewContainer().
		WithContainerConfig(
			cc.WithImage("wordpress:latest"),
			cc.WithEnv("WORDPRESS_DB_HOST", "database-example"),
			cc.WithEnv("WORDPRESS_DB_USER", "exampleuser"),
			cc.WithEnv("WORDPRESS_DB_PASSWORD", "examplepass"),
			cc.WithEnv("WORDPRESS_DB_NAME", "exampledb"),
			cc.WithExposedPort("tcp", "80"),
			cc.WithHealthCheck(
				health.WithTest("CMD", "curl", "-f", "http://localhost/wp-login.php"),
				health.WithStartPeriod(5),
				health.WithInterval(10),
				health.WithTimeout(20),
				health.WithRetries(3),
			),
		).
		WithHostConfig(
			hc.WithRestartPolicyUnlessStopped(),
			tools.WhenTrueElse(IsLinux,
				hc.WithRWNamedVolumeMount("wordpress-data", "/var/www/html"),    // Linux to avoid permission issues
				hc.WithVolumeBinds("./examples/wordpress/src:/var/www/html:rw"), //not created by root so its not a problem
			),
		).
		WithNetworkConfig(nc.WithEndpoint("wordpress-network"))
}

func DatabaseContainer() *create.Container {
	return create.NewContainer("database-container").
		WithContainerConfig(
			cc.WithImage("mysql:8.0"),
			cc.WithEnv("MYSQL_DATABASE", "exampledb"),
			cc.WithEnv("MYSQL_PASSWORD", "examplepass"),
			cc.WithEnv("MYSQL_USER", "exampleuser"),
			cc.WithEnv("MYSQL_RANDOM_ROOT_PASSWORD", "1"),
			cc.WithHealthCheck(
				health.WithTest("CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-pexamplepass"),
				health.WithStartPeriod(5), // Wait for 5 seconds before starting the health check
				health.WithInterval(10),   // Check every 10 seconds
				health.WithTimeout(20),    // Timeout after 20 seconds
				health.WithRetries(3),     // Retry 3 times
			),
		).
		WithHostConfig(
			tools.WhenTrueElse(IsLinux,
				hc.WithRWNamedVolumeMount("database-data", "/var/lib/mysql/"),           // Linux to avoid permission issues
				hc.WithVolumeBinds("./examples/wordpress/database/:/var/lib/mysql/:rw"), //not created by root so its not a problem
			),
		).
		WithNetworkConfig(nc.WithEndpoint("wordpress-network"))
}

func PortainerContainer() *create.Container {
	// Portainer requires access to the Docker socket, which is typically at /var/run/docker.sock
	// For rootless Docker, the socket is at /var/run/user/<UID>/docker.sock
	rootless := fmt.Sprintf("/var/run/user/%d/docker.sock", syscall.Geteuid())
	_, err := os.Stat(rootless)
	isRootless := err == nil

	return create.NewContainer().
		WithContainerConfig(
			cc.WithImage("portainer/portainer-ce:latest"),
		).
		WithHostConfig(
			hc.WithPortBindings("tcp", "0.0.0.0", "9000", "9000"),
			hc.WithRWNamedVolumeMount("portainer-data", "/data"),
			hc.WithMountPoint(
				tools.WhenTrueElse(isRootless,
					mount.WithSource(rootless),
					mount.WithSource("/var/run/docker.sock"),
				),
				mount.WithTarget("/var/run/docker.sock"),
				mount.WithType("bind"),
				mount.WithReadWrite(),
			),
		).
		WithNetworkConfig(nc.WithEndpoint("wordpress-network"))
}

func ProxyContainer() *create.Container {
	return create.NewContainer("proxy-container").
		WithContainerConfig(
			cc.WithImage("haproxy:latest"),
			cc.WithCommand("-f", "/usr/local/etc/haproxy/haproxy.cfg"),
		).
		WithHostConfig(
			hc.WithPortBindings("tcp", "0.0.0.0", "80", "80"),
			hc.WithVolumeBinds("./examples/wordpress/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro"),
		).
		WithNetworkConfig(nc.WithEndpoint("wordpress-network"))
}

// Generate HAProxy configuration for round-robin load balancing
func GenerateHAProxyConfig(services []string) error {

	var sb strings.Builder
	sb.WriteString("global\n")
	sb.WriteString("	log stdout format raw local0\n")
	sb.WriteString("defaults\n")
	sb.WriteString("	log		global\n")
	sb.WriteString("	mode	http\n")
	sb.WriteString("	option	httplog\n")
	sb.WriteString("	option	dontlognull\n")
	sb.WriteString("	timeout connect 5000\n")
	sb.WriteString("	timeout client  50000\n")
	sb.WriteString("	timeout server  50000\n")
	sb.WriteString("frontend http_front\n")
	sb.WriteString("	bind *:80\n")
	sb.WriteString("	default_backend wordpress_back\n")
	sb.WriteString("backend wordpress_back\n")
	sb.WriteString("	balance roundrobin\n")
	for i, backend := range services {
		sb.WriteString(fmt.Sprintf("	server wp%d %s:80 check\n", i+1, backend))
	}
	sb.WriteString("\n")
	cfg := sb.String()

	return os.WriteFile("./examples/wordpress/haproxy.cfg", []byte(cfg), 0644)
}

type Logger struct {
	Target io.Writer
	action string
	buffer bytes.Buffer
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.buffer.Write(p)

	for {
		line, err := l.buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}

		_, werr := fmt.Fprintf(l.Target, "[\x1b[32m%s\x1b[0m] %s", l.action, line)
		if werr != nil {
			return 0, werr
		}

	}

	return len(p), nil
}

func NewLogger(action string) *Logger {

	return &Logger{
		Target: os.Stdout,
		action: action,
	}
}
