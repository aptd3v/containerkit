// Three WordPress containers behind HAProxy, with MySQL and, off Windows, Portainer.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	composeapi "github.com/docker/compose/v5/pkg/api"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

var (
	isLinux      = runtime.GOOS == "linux"
	isNotWindows = runtime.GOOS != "windows"
	numWordPress = 3
)

func main() {
	app := setup()
	yml, err := app.ExportYAML()
	if err != nil {
		log.Fatalf("export: %v", err)
	}
	if err := os.WriteFile("./examples/wordpress/docker-compose.yaml", yml, 0644); err != nil {
		log.Fatalf("write compose: %v", err)
	}

	cli, err := Connect(FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	api := cli.Composer()

	if err := app.Up(context.Background(), api,
		ConsumeLogs(prefixed("up")),
		RemoveOrphans(true),
		Detach(true),
		Timeout(3*time.Second),
	); err != nil {
		log.Fatalf("up: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	go func() {
		<-ctrlc
		cancel()
	}()

	if err := app.Logs(ctx, api, prefixed("logs"), Follow(true)); err != nil && ctx.Err() == nil {
		log.Fatalf("logs: %v", err)
	}

	if err := app.Down(context.Background(), api,
		Log(NewLogger()),
		RemoveOrphans(true),
		Volumes(true),
	); err != nil {
		log.Fatalf("down: %v", err)
	}
}

func setup() *Compose {
	app := NewCompose(fmt.Sprintf("gocontain-wp-scale-%d", numWordPress),
		AddVolume(Name("wordpress-data")),
		AddVolume(Name("database-data")),
		AddVolume(Name("portainer-data")),
		AddNetwork(Name("wordpress-network"), Driver("bridge")),
	).WithService(NewService(append([]ServiceOption{
		Name("database-example"),
	}, database()...)...))

	if isNotWindows {
		app.WithService(NewService(append([]ServiceOption{Name("portainer-container")}, portainer()...)...))
	}

	services := make([]string, 0, numWordPress)
	for i := 1; i <= numWordPress; i++ {
		serviceName := fmt.Sprintf("wordpress-example-%d", i)
		services = append(services, serviceName)
		opts := append(wordpress(), Name(serviceName), DependsOn(Name("database-example"), Condition(Healthy)))
		if i > 1 {
			opts = append(opts, DependsOn(Name(fmt.Sprintf("wordpress-example-%d", i-1))))
		}
		app.WithService(NewService(opts...))
	}
	if err := writeHAProxy(services); err != nil {
		log.Fatalf("haproxy.cfg: %v", err)
	}

	proxyOpts := append(proxy(), Name("proxy-container"))
	for _, serviceName := range services {
		proxyOpts = append(proxyOpts, DependsOn(Name(serviceName), Condition(Healthy)))
	}
	return app.WithService(NewService(proxyOpts...))
}

func wordpress() []ServiceOption {
	opts := []ServiceOption{
		Image("wordpress:latest"),
		Env("WORDPRESS_DB_HOST", "database-example"),
		Env("WORDPRESS_DB_USER", "exampleuser"),
		Env("WORDPRESS_DB_PASSWORD", "examplepass"),
		Env("WORDPRESS_DB_NAME", "exampledb"),
		Expose(80, TCP),
		Health(
			Test("CMD", "curl", "-f", "http://localhost/wp-login.php"),
			StartPeriod(5*time.Second),
			Interval(10*time.Second),
			Timeout(20*time.Second),
			Retries(3),
		),
		Restart(Policy(UnlessStopped)),
		Network(Name("wordpress-network")),
	}
	if isLinux {
		opts = append(opts, Volume(Source("wordpress-data"), Target("/var/www/html")))
	} else {
		opts = append(opts, Volume(Kind(Bind), Source("./examples/wordpress/src"), Target("/var/www/html")))
	}
	return opts
}

func database() []ServiceOption {
	opts := []ServiceOption{
		Image("mysql:8.0"),
		Env("MYSQL_DATABASE", "exampledb"),
		Env("MYSQL_PASSWORD", "examplepass"),
		Env("MYSQL_USER", "exampleuser"),
		Env("MYSQL_RANDOM_ROOT_PASSWORD", "1"),
		Health(
			Test("CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-pexamplepass"),
			StartPeriod(5*time.Second),
			Interval(10*time.Second),
			Timeout(20*time.Second),
			Retries(3),
		),
		Network(Name("wordpress-network")),
	}
	if isLinux {
		opts = append(opts, Volume(Source("database-data"), Target("/var/lib/mysql/")))
	} else {
		opts = append(opts, Volume(Kind(Bind), Source("./examples/wordpress/database/"), Target("/var/lib/mysql/")))
	}
	return opts
}

func portainer() []ServiceOption {
	rootless := fmt.Sprintf("/var/run/user/%d/docker.sock", syscall.Geteuid())
	src := "/var/run/docker.sock"
	if _, err := os.Stat(rootless); err == nil {
		src = rootless
	}
	return []ServiceOption{
		Image("portainer/portainer-ce:latest"),
		Publish(Protocol(TCP), HostIP("0.0.0.0"), HostPort(9000), ContainerPort(9000)),
		Volume(Source("portainer-data"), Target("/data")),
		Volume(Kind(Bind), Source(src), Target("/var/run/docker.sock")),
		Network(Name("wordpress-network")),
	}
}

func proxy() []ServiceOption {
	return []ServiceOption{
		Image("haproxy:latest"),
		Command("-f", "/usr/local/etc/haproxy/haproxy.cfg"),
		Publish(Protocol(TCP), HostIP("0.0.0.0"), HostPort(80), ContainerPort(80)),
		Volume(
			Kind(Bind),
			Source("./examples/wordpress/haproxy.cfg"),
			Target("/usr/local/etc/haproxy/haproxy.cfg"),
			ReadOnly(true),
		),
		Network(Name("wordpress-network")),
	}
}

func writeHAProxy(services []string) error {
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
		fmt.Fprintf(&sb, "	server wp%d %s:80 check\n", i+1, backend)
	}
	sb.WriteString("\n")
	return os.WriteFile("./examples/wordpress/haproxy.cfg", []byte(sb.String()), 0644)
}

type prefixedLog struct{ action string }

func prefixed(action string) *prefixedLog { return &prefixedLog{action: action} }

func (l *prefixedLog) Log(_, message string)    { l.write(message) }
func (l *prefixedLog) Err(_, message string)    { l.write(message) }
func (l *prefixedLog) Status(_, message string) { l.write(message) }

func (l *prefixedLog) write(message string) {
	fmt.Printf("[\x1b[32m%s\x1b[0m] %s\n", l.action, message)
}

var _ composeapi.LogConsumer = (*prefixedLog)(nil)
