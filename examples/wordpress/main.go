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

	"github.com/aptd3v/containerkit/pkg/containerkit"
)

var (
	IsLinux      = runtime.GOOS == "linux"
	IsNotWindows = runtime.GOOS != "windows"
	NumWordPress = 3
)

func main() {

	project := SetupProject()
	err := project.Export("./examples/wordpress/docker-compose.yaml", 0644)
	if err != nil {
		log.Fatalf("failed to export to docker-compose.yaml: %v", err)
	}

	wordpress := containerkit.NewCompose(project)

	upTimeout := 3
	err = wordpress.Up(context.Background(), &containerkit.Up{
		Writer:        NewLogger("up"),
		RemoveOrphans: true,
		NoLogPrefix:   true,
		Detach:        true,
		Timeout:       &upTimeout,
	})
	if err != nil {
		log.Fatalf("failed to execute up: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	go func() {
		<-ctrlc
		cancel()
	}()

	err = wordpress.Logs(ctx, &containerkit.Logs{
		Writer:      NewLogger("logs"),
		NoLogPrefix: true,
		Follow:      true,
	})
	if err != nil {
		log.Fatalf("failed to execute logs: %v", err)
	}

	err = wordpress.Down(
		context.Background(),
		&containerkit.Down{
			Writer:        NewLogger("down"),
			RemoveOrphans: true,
			RemoveVolumes: true,
		},
	)
	if err != nil {
		log.Fatalf("failed to execute down: %v", err)
	}
}

func SetupProject() *containerkit.Project {
	project := containerkit.NewProject(fmt.Sprintf("gocontain-wp-scale-%d", NumWordPress))
	project.WithService("database-example", DatabaseContainer())

	if IsNotWindows {
		project.WithService("portainer-container", PortainerContainer())
	}
	deps := []any{}
	services := []string{}
	for i := 1; i <= NumWordPress; i++ {
		serviceName := fmt.Sprintf("wordpress-example-%d", i)
		services = append(services, serviceName)
		extras := []any{containerkit.DependsOnHealthy("database-example")}
		if i > 1 {
			extras = append(extras, containerkit.DependsOn(fmt.Sprintf("wordpress-example-%d", i-1)))
		}
		project.WithService(serviceName, WordPressContainer(), extras...)
		deps = append(deps, containerkit.DependsOnHealthy(fmt.Sprintf("wordpress-example-%d", i)))
	}
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

func WordPressContainer() *containerkit.Container {
	c := containerkit.NewContainer().
		Image("wordpress:latest").
		Env("WORDPRESS_DB_HOST", "database-example").
		Env("WORDPRESS_DB_USER", "exampleuser").
		Env("WORDPRESS_DB_PASSWORD", "examplepass").
		Env("WORDPRESS_DB_NAME", "exampledb").
		ExposedPort("tcp", "80").
		HealthCheck(containerkit.Health{
			Test:        []string{"CMD", "curl", "-f", "http://localhost/wp-login.php"},
			StartPeriod: 5,
			Interval:    10,
			Timeout:     20,
			Retries:     3,
		}).
		RestartUnlessStopped().
		Endpoint("wordpress-network")
	if IsLinux {
		c.RWNamedVolumeMount("wordpress-data", "/var/www/html")
	} else {
		c.VolumeBinds("./examples/wordpress/src:/var/www/html:rw")
	}
	return c
}

func DatabaseContainer() *containerkit.Container {
	c := containerkit.NewContainer("database-container").
		Image("mysql:8.0").
		Env("MYSQL_DATABASE", "exampledb").
		Env("MYSQL_PASSWORD", "examplepass").
		Env("MYSQL_USER", "exampleuser").
		Env("MYSQL_RANDOM_ROOT_PASSWORD", "1").
		HealthCheck(containerkit.Health{
			Test:        []string{"CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-pexamplepass"},
			StartPeriod: 5,
			Interval:    10,
			Timeout:     20,
			Retries:     3,
		}).
		Endpoint("wordpress-network")
	if IsLinux {
		c.RWNamedVolumeMount("database-data", "/var/lib/mysql/")
	} else {
		c.VolumeBinds("./examples/wordpress/database/:/var/lib/mysql/:rw")
	}
	return c
}

func PortainerContainer() *containerkit.Container {
	rootless := fmt.Sprintf("/var/run/user/%d/docker.sock", syscall.Geteuid())
	_, err := os.Stat(rootless)
	isRootless := err == nil
	src := "/var/run/docker.sock"
	if isRootless {
		src = rootless
	}

	return containerkit.NewContainer().
		Image("portainer/portainer-ce:latest").
		PortBindings("tcp", "0.0.0.0", "9000", "9000").
		RWNamedVolumeMount("portainer-data", "/data").
		Mount(containerkit.Mount{
			Source: src, Target: "/var/run/docker.sock",
			Type: containerkit.MountBind,
		}).
		Endpoint("wordpress-network")
}

func ProxyContainer() *containerkit.Container {
	return containerkit.NewContainer("proxy-container").
		Image("haproxy:latest").
		Command("-f", "/usr/local/etc/haproxy/haproxy.cfg").
		PortBindings("tcp", "0.0.0.0", "80", "80").
		VolumeBinds("./examples/wordpress/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro").
		Endpoint("wordpress-network")
}

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
