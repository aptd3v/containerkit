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
	"sync"
	"syscall"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
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

	wordpress := kit.NewCompose(project)

	upTimeout := 3
	err = wordpress.Up(context.Background(), &kit.Up{
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

	err = wordpress.Logs(ctx, &kit.Logs{
		Writer:      NewLogger("logs"),
		NoLogPrefix: true,
		Follow:      true,
	})
	if err != nil {
		log.Fatalf("failed to execute logs: %v", err)
	}

	err = wordpress.Down(
		context.Background(),
		&kit.Down{
			Writer:        NewLogger("down"),
			RemoveOrphans: true,
			RemoveVolumes: true,
		},
	)
	if err != nil {
		log.Fatalf("failed to execute down: %v", err)
	}
}

func SetupProject() *kit.Project {
	project := kit.NewProject(fmt.Sprintf("gocontain-wp-scale-%d", NumWordPress))
	project.WithService(kit.NewService("database-example", DatabaseContainer()))

	if IsNotWindows {
		project.WithService(kit.NewService("portainer-container", PortainerContainer()))
	}
	proxy := kit.NewService("proxy-container", ProxyContainer())
	services := []string{}
	for i := 1; i <= NumWordPress; i++ {
		serviceName := fmt.Sprintf("wordpress-example-%d", i)
		services = append(services, serviceName)
		wordpress := kit.NewService(serviceName, WordPressContainer()).DependsOnHealthy("database-example")
		if i > 1 {
			wordpress.DependsOn(fmt.Sprintf("wordpress-example-%d", i-1))
		}
		project.WithService(wordpress)
		proxy.DependsOnHealthy(serviceName)
	}
	err := GenerateHAProxyConfig(services)
	if err != nil {
		log.Fatalf("failed to generate haproxy.cfg: %v", err)
	}

	project.WithService(proxy)

	project.
		WithVolume("wordpress-data").
		WithVolume("database-data").
		WithVolume("portainer-data").
		WithNetwork("wordpress-network")

	return project
}

func WordPressContainer() *kit.Container {
	c := kit.NewContainer().
		Image("wordpress:latest").
		Env("WORDPRESS_DB_HOST", "database-example").
		Env("WORDPRESS_DB_USER", "exampleuser").
		Env("WORDPRESS_DB_PASSWORD", "examplepass").
		Env("WORDPRESS_DB_NAME", "exampledb").
		ExposedPort("tcp", "80").
		HealthCheck(kit.NewHealth().
			CMD("curl", "-f", "http://localhost/wp-login.php").
			StartPeriod(5).
			Interval(10).
			Timeout(20).
			Retries(3)).
		RestartUnlessStopped().
		Endpoint("wordpress-network")
	if IsLinux {
		c.RWNamedVolumeMount("wordpress-data", "/var/www/html")
	} else {
		c.VolumeBinds("./examples/wordpress/src:/var/www/html:rw")
	}
	return c
}

func DatabaseContainer() *kit.Container {
	c := kit.NewContainer("database-container").
		Image("mysql:8.0").
		Env("MYSQL_DATABASE", "exampledb").
		Env("MYSQL_PASSWORD", "examplepass").
		Env("MYSQL_USER", "exampleuser").
		Env("MYSQL_RANDOM_ROOT_PASSWORD", "1").
		HealthCheck(kit.NewHealth().
			CMD("mysqladmin", "ping", "-h", "127.0.0.1", "-pexamplepass").
			StartPeriod(5).
			Interval(10).
			Timeout(20).
			Retries(3)).
		Endpoint("wordpress-network")
	if IsLinux {
		c.RWNamedVolumeMount("database-data", "/var/lib/mysql/")
	} else {
		c.VolumeBinds("./examples/wordpress/database/:/var/lib/mysql/:rw")
	}
	return c
}

func PortainerContainer() *kit.Container {
	rootless := fmt.Sprintf("/var/run/user/%d/docker.sock", syscall.Geteuid())
	_, err := os.Stat(rootless)
	isRootless := err == nil
	src := "/var/run/docker.sock"
	if isRootless {
		src = rootless
	}

	return kit.NewContainer().
		Image("portainer/portainer-ce:latest").
		PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "9000", ContainerPort: "9000"}).
		RWNamedVolumeMount("portainer-data", "/data").
		Mount(kit.Mount{
			Source: src, Target: "/var/run/docker.sock",
			Type: types.MountBind,
		}).
		Endpoint("wordpress-network")
}

func ProxyContainer() *kit.Container {
	return kit.NewContainer("proxy-container").
		Image("haproxy:latest").
		Command("-f", "/usr/local/etc/haproxy/haproxy.cfg").
		PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "80", ContainerPort: "80"}).
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

	mu     sync.Mutex
	buffer bytes.Buffer
}

func (l *Logger) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.buffer.Write(p); err != nil {
		return 0, err
	}
	for {
		line, err := l.buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// ReadString consumed the incomplete line; keep it for the next Write.
				_, _ = l.buffer.WriteString(line)
				return len(p), nil
			}
			return 0, err
		}
		if _, err := fmt.Fprintf(l.Target, "[\x1b[32m%s\x1b[0m] %s", l.action, line); err != nil {
			return 0, err
		}
	}
}

func NewLogger(action string) *Logger {

	return &Logger{
		Target: os.Stdout,
		action: action,
	}
}
