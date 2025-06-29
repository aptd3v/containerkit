package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/aptd3v/go-contain/pkg/compose"
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
	NumWordPress = 3 //change me
)

func main() {

	project := SetupProject()
	err := project.Export("./examples/wordpress/docker-compose.yaml", 0644)
	if err != nil {
		log.Fatalf("failed to export to docker-compose.yaml: %v", err)
	}
	fmt.Println("docker-compose.yaml exported successfully")

	wordpress := compose.NewCompose(project)
	err = wordpress.Up(up.WithRemoveOrphans())
	if err != nil {
		log.Fatalf("failed to up: %v", err)
	}
	fmt.Println("up successfully")
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
	return create.NewContainer("wordpress-container").
		WithContainerConfig(
			cc.WithImage("wordpress:latest"),
			cc.WithEnv("WORDPRESS_DB_HOST", "database-example"),
			cc.WithEnv("WORDPRESS_DB_USER", "exampleuser"),
			cc.WithEnv("WORDPRESS_DB_PASSWORD", "examplepass"),
			cc.WithEnv("WORDPRESS_DB_NAME", "exampledb"),
			cc.WithExposedPort("tcp", "80"),
			cc.WithCurlHealthCheck("http://localhost/wp-login.php", 10),
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
	return create.NewContainer("portainer-container").
		WithContainerConfig(
			cc.WithImage("portainer/portainer-ce:latest"),
		).
		WithHostConfig(
			hc.WithPortBindings("tcp", "0.0.0.0", "9000", "9000"),
			hc.WithRWNamedVolumeMount("portainer-data", "/data"),
			hc.WithMountPoint(
				mount.WithSource("/var/run/docker.sock"),
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

	var backends []string
	for i, service := range services {
		backends = append(backends, fmt.Sprintf("    server wp%d %s:80 check", i+1, service))
	}

	cfg := fmt.Sprintf(`
global
    log stdout format raw local0

defaults
    log     global
    mode    http
    option  httplog
    option  dontlognull
    timeout connect 5000
    timeout client  50000
    timeout server  50000

frontend http_front
    bind *:80
    default_backend wordpress_back

backend wordpress_back
    balance roundrobin
%s
`, strings.Join(backends, "\n"))

	return os.WriteFile("./examples/wordpress/haproxy.cfg", []byte(cfg), 0644)
}
