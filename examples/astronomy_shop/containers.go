// Package main: Astronomy Shop service container builders (published demo images).
package main

import (
	"fmt"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

func base(name string) *kit.Container {
	return kit.NewContainer(name).
		RestartUnlessStopped().
		Endpoint(networkName)
}

// bashTCPHealth uses bash /dev/tcp (Java/Envoy images that ship bash but not nc).
func bashTCPHealth(port string, start, interval, timeout, retries int) kit.Health {
	return kit.NewHealth().
		CMDShell(fmt.Sprintf("bash -c 'echo > /dev/tcp/localhost/%s'", port)).
		StartPeriod(start).
		Interval(interval).
		Timeout(timeout).
		Retries(retries)
}

// ncHealth uses nc -z (Alpine images like recommendation that lack bash).
func ncHealth(port string, start, interval, timeout, retries int) kit.Health {
	return kit.NewHealth().
		CMDShell("nc -z localhost " + port).
		StartPeriod(start).
		Interval(interval).
		Timeout(timeout).
		Retries(retries)
}

func grpcHealth(port string) kit.Health {
	return kit.NewHealth().
		CMD("/bin/grpc_health_probe", "-addr=:"+port).
		StartPeriod(10).
		Interval(5).
		Timeout(5).
		Retries(20)
}

func publish(c *kit.Container, hostPort, containerPort string) *kit.Container {
	return c.PortBinding(types.PortBinding{
		Protocol:      types.TCP,
		HostPort:      hostPort,
		ContainerPort: containerPort,
	})
}

func astronomyDB(cfg *Config, vol func(string) string) *kit.Container {
	return base("astronomy-db").
		Image(cfg.PostgresImage).
		Env("POSTGRES_PASSWORD", cfg.PostgresPassword).
		Command("postgres", "-c", "shared_preload_libraries=pg_stat_statements").
		HealthCheck(kit.NewHealth().
			CMDShell("pg_isready").
			StartPeriod(10).
			Interval(5).
			Timeout(5).
			Retries(10)).
		VolumeBinds(vol("postgresql/init.sql") + ":/docker-entrypoint-initdb.d/init.sql:ro")
}

func valkeyCart(cfg *Config) *kit.Container {
	return base("valkey-cart").
		Image(cfg.ValkeyImage).
		User("valkey").
		HealthCheck(kit.NewHealth().
			CMDShell("valkey-cli ping").
			StartPeriod(5).
			Interval(5).
			Timeout(3).
			Retries(10))
}

func otelCollector(cfg *Config, vol func(string) string, withObs bool) *kit.Container {
	cmd := []string{
		"--config=/etc/otelcol-config.yml",
	}
	binds := []string{
		vol("otel-collector/otelcol-config.yml") + ":/etc/otelcol-config.yml:ro",
		vol("otel-collector/otelcol-config-extras.yml") + ":/etc/otelcol-config-extras.yml:ro",
	}
	if withObs {
		cmd = append(cmd, "--config=/etc/otelcol-config-observability.yml")
		binds = append(binds, vol("otel-collector/otelcol-config-observability.yml")+":/etc/otelcol-config-observability.yml:ro")
	}
	cmd = append(cmd, "--config=/etc/otelcol-config-extras.yml")

	return base("otel-collector").
		Image(cfg.CollectorImage).
		User("0:0").
		Command(cmd...).
		EnvMap(map[string]string{
			"OTEL_COLLECTOR_HOST":          "0.0.0.0",
			"OTEL_COLLECTOR_PORT_GRPC":     cfg.CollectorGRPCPort,
			"OTEL_COLLECTOR_PORT_HTTP":     cfg.CollectorHTTPPort,
			"FRONTEND_PROXY_ADDR":          "frontend-proxy:" + cfg.EnvoyPort,
			"IMAGE_PROVIDER_HOST":          "image-provider",
			"IMAGE_PROVIDER_PORT":          cfg.ImageProviderPort,
			"POSTGRES_HOST":                "astronomy-db",
			"POSTGRES_PORT":                cfg.PostgresPort,
			"POSTGRES_MONITORING_PASSWORD": cfg.PostgresMonitoringPassword,
			"AD_PROMETHEUS_PORT":           cfg.ADPrometheusPort,
			"GOMEMLIMIT":                   "160MiB",
		}).
		VolumeBinds(binds...).
		ExposedPort(types.TCP, cfg.CollectorGRPCPort).
		ExposedPort(types.TCP, cfg.CollectorHTTPPort)
}

func flagd(cfg *Config, vol func(string) string) *kit.Container {
	return base("flagd").
		Image(cfg.FlagdImage).
		Command("start", "--uri", "file:./etc/flagd/demo.flagd.json").
		EnvMap(mergeEnv(
			cfg.commonOTEL("flagd", "low", cfg.otlpHTTP()),
			map[string]string{
				"OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
				"GOMEMLIMIT":                  "60MiB",
			},
		)).
		VolumeBinds(vol("flagd")+":/etc/flagd:ro").
		ExposedPort(types.TCP, cfg.FlagdPort).
		ExposedPort(types.TCP, cfg.FlagdOFREPPort)
}

func flagdUI(cfg *Config, vol func(string) string) *kit.Container {
	return base("flagd-ui").
		Image(cfg.demoImage("flagd-ui")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("flagd-ui", "low", cfg.otlpHTTP()),
			map[string]string{
				"FLAGD_UI_PORT":   cfg.FlagdUIPort,
				"SECRET_KEY_BASE": "yYrECL4qbNwleYInGJYvVnSkwJuSQJ4ijPTx5tirGUXrbznFIBFVJdPl5t6O9ASw",
				"PHX_HOST":        "localhost",
			},
		)).
		HealthCheck(bashTCPHealth(cfg.FlagdUIPort, 10, 5, 5, 20)).
		VolumeBinds(vol("flagd")+":/app/data").
		ExposedPort(types.TCP, cfg.FlagdUIPort)
}

func ad(cfg *Config) *kit.Container {
	return base("ad").
		Image(cfg.demoImage("ad")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("ad", "medium", cfg.otlpHTTP()),
			map[string]string{
				"AD_PORT":            cfg.ADPort,
				"AD_PROMETHEUS_PORT": cfg.ADPrometheusPort,
				"FLAGD_HOST":         "flagd",
				"FLAGD_PORT":         cfg.FlagdPort,
				"OTEL_LOGS_EXPORTER": "otlp",
				"OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION": "latest",
			},
		)).
		HealthCheck(bashTCPHealth(cfg.ADPort, 20, 5, 5, 20)).
		ExposedPort(types.TCP, cfg.ADPort)
}

func cart(cfg *Config) *kit.Container {
	return base("cart").
		Image(cfg.demoImage("cart")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("cart", "high", cfg.otlpGRPC()),
			map[string]string{
				"CART_PORT":             cfg.CartPort,
				"FLAGD_HOST":            "flagd",
				"FLAGD_PORT":            cfg.FlagdPort,
				"VALKEY_ADDR":           "valkey-cart:" + cfg.ValkeyPort,
				"ASPNETCORE_HTTP_PORTS": cfg.CartPort,
			},
		)).
		ExposedPort(types.TCP, cfg.CartPort)
}

func checkout(cfg *Config) *kit.Container {
	return base("checkout").
		Image(cfg.demoImage("checkout")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("checkout", "critical", cfg.otlpHTTP()),
			map[string]string{
				"CHECKOUT_PORT":               cfg.CheckoutPort,
				"FLAGD_HOST":                  "flagd",
				"FLAGD_PORT":                  cfg.FlagdPort,
				"CART_ADDR":                   "cart:" + cfg.CartPort,
				"CURRENCY_ADDR":               "currency:" + cfg.CurrencyPort,
				"EMAIL_ADDR":                  "http://email:" + cfg.EmailPort,
				"PAYMENT_ADDR":                "payment:" + cfg.PaymentPort,
				"PRODUCT_CATALOG_ADDR":        "product-catalog:" + cfg.ProductCatalogPort,
				"SHIPPING_ADDR":               "http://shipping:" + cfg.ShippingPort,
				"KAFKA_ADDR":                  "kafka:" + cfg.KafkaPort,
				"KAFKA_TOPIC":                 "orders",
				"GOMEMLIMIT":                  "16MiB",
				"OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
				"OTEL_GO_X_OBSERVABILITY":     "true",
			},
		)).
		HealthCheck(grpcHealth(cfg.CheckoutPort)).
		ExposedPort(types.TCP, cfg.CheckoutPort)
}

func currency(cfg *Config) *kit.Container {
	return base("currency").
		Image(cfg.demoImage("currency")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("currency", "high", cfg.otlpGRPC()),
			map[string]string{
				"CURRENCY_PORT": cfg.CurrencyPort,
				"IPV6_ENABLED":  cfg.IPV6Enabled,
				"VERSION":       cfg.ImageVersion,
			},
		)).
		HealthCheck(kit.NewHealth().
			CMDShell("nc -z localhost "+cfg.CurrencyPort).
			StartPeriod(10).Interval(5).Timeout(5).Retries(20)).
		ExposedPort(types.TCP, cfg.CurrencyPort)
}

func email(cfg *Config) *kit.Container {
	return base("email").
		Image(cfg.demoImage("email")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("email", "medium", cfg.otlpHTTP()),
			map[string]string{
				"APP_ENV":    "production",
				"EMAIL_PORT": cfg.EmailPort,
				"FLAGD_HOST": "flagd",
				"FLAGD_PORT": cfg.FlagdPort,
			},
		)).
		HealthCheck(kit.NewHealth().
			CMD("ruby", "-e", fmt.Sprintf("require 'socket'; TCPSocket.new('localhost', %s).close", cfg.EmailPort)).
			StartPeriod(10).Interval(5).Timeout(5).Retries(20)).
		ExposedPort(types.TCP, cfg.EmailPort)
}

func payment(cfg *Config) *kit.Container {
	return base("payment").
		Image(cfg.demoImage("payment")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("payment", "critical", cfg.otlpGRPC()),
			map[string]string{
				"IPV6_ENABLED":                cfg.IPV6Enabled,
				"PAYMENT_PORT":                cfg.PaymentPort,
				"FLAGD_HOST":                  "flagd",
				"FLAGD_PORT":                  cfg.FlagdPort,
				"OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
			},
		)).
		HealthCheck(kit.NewHealth().
			CMD("/nodejs/bin/node", "-e",
				fmt.Sprintf("const net=require('net');const c=new net.Socket(); c.setTimeout(2000); c.connect(%s,'127.0.0.1',()=>{c.destroy();process.exit(0)}); c.on('error',()=>process.exit(1)); c.on('timeout',()=>process.exit(1))", cfg.PaymentPort)).
			StartPeriod(10).Interval(5).Timeout(5).Retries(20)).
		ExposedPort(types.TCP, cfg.PaymentPort)
}

func productCatalog(cfg *Config, vol func(string) string) *kit.Container {
	return base("product-catalog").
		Image(cfg.demoImage("product-catalog")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("product-catalog", "high", cfg.otlpGRPC()),
			map[string]string{
				"PRODUCT_CATALOG_PORT":          cfg.ProductCatalogPort,
				"FLAGD_HOST":                    "flagd",
				"FLAGD_PORT":                    cfg.FlagdPort,
				"GOMEMLIMIT":                    "16MiB",
				"OTEL_GO_X_OBSERVABILITY":       "true",
				"OTEL_CONFIG_FILE":              "/otel-config.yml",
				"OTEL_SEMCONV_STABILITY_OPT_IN": "database",
				"DB_CONNECTION_STRING": fmt.Sprintf(
					"postgres://astronomy_user:%s@astronomy-db/astronomy_db?sslmode=disable",
					cfg.PostgresAstronomyPassword,
				),
			},
		)).
		HealthCheck(grpcHealth(cfg.ProductCatalogPort)).
		VolumeBinds(vol("otel-config.yml")+":/otel-config.yml:ro").
		ExposedPort(types.TCP, cfg.ProductCatalogPort)
}

func quote(cfg *Config) *kit.Container {
	return base("quote").
		Image(cfg.demoImage("quote")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("quote", "low", cfg.otlpHTTP()),
			map[string]string{
				"IPV6_ENABLED":                      cfg.IPV6Enabled,
				"QUOTE_PORT":                        cfg.QuotePort,
				"OTEL_PHP_AUTOLOAD_ENABLED":         "true",
				"OTEL_PHP_INTERNAL_METRICS_ENABLED": "true",
			},
		)).
		HealthCheck(kit.NewHealth().
			CMD("php", "-r", fmt.Sprintf("fsockopen('localhost', %s) or die('fail');", cfg.QuotePort)).
			StartPeriod(10).Interval(5).Timeout(5).Retries(20)).
		ExposedPort(types.TCP, cfg.QuotePort)
}

func recommendation(cfg *Config) *kit.Container {
	return base("recommendation").
		Image(cfg.demoImage("recommendation")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("recommendation", "medium", cfg.otlpHTTP()),
			map[string]string{
				"RECOMMENDATION_PORT":  cfg.RecommendationPort,
				"PRODUCT_CATALOG_ADDR": "product-catalog:" + cfg.ProductCatalogPort,
				"FLAGD_HOST":           "flagd",
				"FLAGD_PORT":           cfg.FlagdPort,
				"OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION": "latest",
			},
		)).
		HealthCheck(ncHealth(cfg.RecommendationPort, 20, 5, 5, 20)).
		ExposedPort(types.TCP, cfg.RecommendationPort)
}

func shipping(cfg *Config) *kit.Container {
	return base("shipping").
		Image(cfg.demoImage("shipping")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("shipping", "high", cfg.otlpHTTP()),
			map[string]string{
				"IPV6_ENABLED":                cfg.IPV6Enabled,
				"SHIPPING_PORT":               cfg.ShippingPort,
				"QUOTE_ADDR":                  "http://quote:" + cfg.QuotePort,
				"FLAGD_HOST":                  "flagd",
				"FLAGD_PORT":                  cfg.FlagdPort,
				"OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
			},
		)).
		HealthCheck(kit.NewHealth().
			CMD("/app/healthcheck").
			StartPeriod(10).Interval(5).Timeout(5).Retries(20)).
		ExposedPort(types.TCP, cfg.ShippingPort)
}

func imageProvider(cfg *Config) *kit.Container {
	return base("image-provider").
		Image(cfg.demoImage("image-provider")).
		EnvMap(map[string]string{
			"IMAGE_PROVIDER_PORT":      cfg.ImageProviderPort,
			"OTEL_COLLECTOR_HOST":      "otel-collector",
			"OTEL_COLLECTOR_PORT_GRPC": cfg.CollectorGRPCPort,
			"OTEL_RESOURCE_ATTRIBUTES": cfg.resourceAttributes("low"),
			"OTEL_SERVICE_NAME":        "image-provider",
		}).
		HealthCheck(kit.NewHealth().
			CMD("wget", "--quiet", "--tries=1", "--spider", "http://localhost:"+cfg.ImageProviderPort+"/status").
			StartPeriod(10).Interval(5).Timeout(5).Retries(20)).
		ExposedPort(types.TCP, cfg.ImageProviderPort)
}

func frontend(cfg *Config) *kit.Container {
	return base("frontend").
		Image(cfg.demoImage("frontend")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("frontend", "critical", cfg.otlpGRPC()),
			map[string]string{
				"PORT":                 cfg.FrontendPort,
				"FRONTEND_ADDR":        "frontend:" + cfg.FrontendPort,
				"AD_ADDR":              "ad:" + cfg.ADPort,
				"CART_ADDR":            "cart:" + cfg.CartPort,
				"CHECKOUT_ADDR":        "checkout:" + cfg.CheckoutPort,
				"CURRENCY_ADDR":        "currency:" + cfg.CurrencyPort,
				"PRODUCT_CATALOG_ADDR": "product-catalog:" + cfg.ProductCatalogPort,
				"RECOMMENDATION_ADDR":  "recommendation:" + cfg.RecommendationPort,
				"SHIPPING_ADDR":        "http://shipping:" + cfg.ShippingPort,
				"ENV_PLATFORM":         cfg.EnvPlatform,
				"PUBLIC_OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://localhost:" + cfg.EnvoyPort + "/otlp-http/v1/traces",
				"WEB_OTEL_SERVICE_NAME":                     "frontend-web",
				"OTEL_COLLECTOR_HOST":                       "otel-collector",
				"FLAGD_HOST":                                "flagd",
				"FLAGD_PORT":                                cfg.FlagdPort,
			},
		)).
		HealthCheck(kit.NewHealth().
			CMD("/nodejs/bin/node", "-e",
				fmt.Sprintf("require('net').connect(%s,require('os').hostname(),function(){process.exit(0)}).on('error',function(){process.exit(1)})", cfg.FrontendPort)).
			StartPeriod(60).Interval(10).Timeout(10).Retries(5)).
		ExposedPort(types.TCP, cfg.FrontendPort)
}

func loadGenerator(cfg *Config) *kit.Container {
	c := base("load-generator").
		Image(cfg.demoImage("load-generator")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("load-generator", "low", cfg.otlpGRPC()),
			map[string]string{
				"LOCUST_WEB_PORT":                          cfg.LocustWebPort,
				"LOCUST_USERS":                             "5",
				"LOCUST_HOST":                              "http://frontend-proxy:" + cfg.EnvoyPort,
				"LOCUST_HEADLESS":                          "false",
				"LOCUST_AUTOSTART":                         "true",
				"LOCUST_BROWSER_TRAFFIC_ENABLED":           "true",
				"LOCUST_HTTP_USER_WEIGHT":                  "9",
				"LOCUST_BROWSER_USER_WEIGHT":               "1",
				"LOCUST_WEB_HOST":                          "0.0.0.0",
				"OTEL_PYTHON_SDK_INTERNAL_METRICS_ENABLED": "true",
				"PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION":   "python",
				"FLAGD_HOST":                               "flagd",
				"FLAGD_PORT":                               cfg.FlagdPort,
				"FLAGD_OFREP_PORT":                         cfg.FlagdOFREPPort,
			},
		)).
		HealthCheck(kit.NewHealth().
			CMD("python3", "-c",
				fmt.Sprintf("import socket; s=socket.create_connection(('localhost',%s),2); s.close()", cfg.LocustWebPort)).
			StartPeriod(30).Interval(5).Timeout(5).Retries(20))
	return publish(c, cfg.LocustWebPort, cfg.LocustWebPort)
}

func frontendProxy(cfg *Config) *kit.Container {
	c := base("frontend-proxy").
		Image(cfg.demoImage("frontend-proxy")).
		EnvMap(map[string]string{
			"FRONTEND_PORT":            cfg.FrontendPort,
			"FRONTEND_HOST":            "frontend",
			"LOCUST_WEB_HOST":          "load-generator",
			"LOCUST_WEB_PORT":          cfg.LocustWebPort,
			"GRAFANA_PORT":             cfg.GrafanaPort,
			"GRAFANA_HOST":             "grafana",
			"JAEGER_UI_PORT":           cfg.JaegerUIPort,
			"JAEGER_HOST":              "jaeger",
			"OPAMP_SERVER_HOST":        "opamp-server",
			"OPAMP_SERVER_UI_PORT":     "4321",
			"OTEL_COLLECTOR_HOST":      "otel-collector",
			"IMAGE_PROVIDER_HOST":      "image-provider",
			"IMAGE_PROVIDER_PORT":      cfg.ImageProviderPort,
			"OTEL_COLLECTOR_PORT_GRPC": cfg.CollectorGRPCPort,
			"OTEL_COLLECTOR_PORT_HTTP": cfg.CollectorHTTPPort,
			"OTEL_RESOURCE_ATTRIBUTES": cfg.resourceAttributes("critical"),
			"OTEL_SERVICE_NAME":        "frontend-proxy",
			"OTEL_SERVICE_NAMESPACE":   cfg.OTELServiceNamespace,
			"ENVOY_PORT":               cfg.EnvoyPort,
			"ENVOY_ADDR":               "0.0.0.0",
			"ENVOY_ADMIN_PORT":         cfg.EnvoyAdminPort,
			"FIREPIT_HOST":             "firepit",
			"FIREPIT_PORT":             "8080",
			"FLAGD_HOST":               "flagd",
			"FLAGD_PORT":               cfg.FlagdPort,
			"FLAGD_UI_HOST":            "flagd-ui",
			"FLAGD_UI_PORT":            cfg.FlagdUIPort,
			"TELEMETRY_DOCS_HOST":      "telemetry-docs",
			"TELEMETRY_DOCS_PORT":      "8000",
			"CHATBOT_HOST":             "chatbot",
			"CHATBOT_PORT":             "7860",
		}).
		HealthCheck(kit.NewHealth().
			CMD("bash", "-c", "echo > /dev/tcp/localhost/"+cfg.EnvoyPort).
			StartPeriod(10).Interval(5).Timeout(5).Retries(20))
	return publish(c, cfg.EnvoyPort, cfg.EnvoyPort)
}

func kafka(cfg *Config) *kit.Container {
	return base("kafka").
		Image(cfg.demoImage("kafka")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("kafka", "low", cfg.otlpHTTP()),
			map[string]string{
				"KAFKA_ADVERTISED_LISTENERS":              "PLAINTEXT://kafka:9092",
				"KAFKA_LISTENERS":                         "PLAINTEXT://kafka:9092,CONTROLLER://kafka:9093",
				"KAFKA_CONTROLLER_QUORUM_VOTERS":          "1@kafka:9093",
				"KAFKA_HEAP_OPTS":                         "-Xmx400m -Xms400m",
				"OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION": "latest",
			},
		)).
		HealthCheck(kit.NewHealth().
			CMDShell("nc -z kafka 9092").
			StartPeriod(10).Interval(5).Timeout(10).Retries(10)).
		ExposedPort(types.TCP, cfg.KafkaPort)
}

func accounting(cfg *Config) *kit.Container {
	return base("accounting").
		Image(cfg.demoImage("accounting")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("accounting", "low", cfg.otlpHTTP()),
			map[string]string{
				"KAFKA_ADDR":  "kafka:" + cfg.KafkaPort,
				"KAFKA_TOPIC": "orders",
				"DB_CONNECTION_STRING": fmt.Sprintf(
					"Host=astronomy-db;Username=astronomy_user;Password=%s;Database=astronomy_db",
					cfg.PostgresAstronomyPassword,
				),
				"OTEL_DOTNET_AUTO_TRACES_ENTITYFRAMEWORKCORE_INSTRUMENTATION_ENABLED": "false",
			},
		))
}

func fraudDetection(cfg *Config) *kit.Container {
	return base("fraud-detection").
		Image(cfg.demoImage("fraud-detection")).
		EnvMap(mergeEnv(
			cfg.commonOTEL("fraud-detection", "low", cfg.otlpHTTP()),
			map[string]string{
				"FLAGD_HOST":  "flagd",
				"FLAGD_PORT":  cfg.FlagdPort,
				"KAFKA_ADDR":  "kafka:" + cfg.KafkaPort,
				"KAFKA_TOPIC": "orders",
				"OTEL_INSTRUMENTATION_KAFKA_EXPERIMENTAL_SPAN_ATTRIBUTES":               "true",
				"OTEL_INSTRUMENTATION_MESSAGING_EXPERIMENTAL_RECEIVE_TELEMETRY_ENABLED": "true",
				"OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION":                               "latest",
			},
		))
}

func jaeger(cfg *Config, vol func(string) string) *kit.Container {
	c := base("jaeger").
		Image(cfg.JaegerImage).
		Command("--config=file:/etc/jaeger/config.yml").
		EnvMap(map[string]string{
			"JAEGER_HOST":              "0.0.0.0",
			"JAEGER_GRPC_PORT":         cfg.JaegerGRPCPort,
			"PROMETHEUS_ADDR":          "prometheus:" + cfg.PrometheusPort,
			"OTEL_COLLECTOR_HOST":      "otel-collector",
			"OTEL_COLLECTOR_PORT_HTTP": cfg.CollectorHTTPPort,
			"MEMORY_MAX_TRACES":        "25000",
		}).
		VolumeBinds(
			vol("jaeger/config.yml")+":/etc/jaeger/config.yml:ro",
			vol("jaeger/ui-config.json")+":/etc/jaeger/ui-config.json:ro",
		)
	return publish(c, cfg.JaegerUIPort, cfg.JaegerUIPort)
}

func prometheus(cfg *Config, vol func(string) string) *kit.Container {
	c := base("prometheus").
		Image(cfg.PrometheusImage).
		Command(
			"--web.console.templates=/etc/prometheus/consoles",
			"--web.console.libraries=/etc/prometheus/console_libraries",
			"--storage.tsdb.retention.time=7d",
			"--config.file=/etc/prometheus/prometheus-config.yaml",
			"--storage.tsdb.path=/prometheus",
			"--web.enable-lifecycle",
			"--web.route-prefix=/",
			"--web.enable-otlp-receiver",
			"--enable-feature=exemplar-storage",
		).
		VolumeBinds(vol("prometheus/prometheus-config.yaml") + ":/etc/prometheus/prometheus-config.yaml:ro")
	return publish(c, cfg.PrometheusPort, cfg.PrometheusPort)
}

func grafana(cfg *Config, vol func(string) string) *kit.Container {
	c := base("grafana").
		Image(cfg.GrafanaImage).
		VolumeBinds(
			vol("grafana/grafana.ini")+":/etc/grafana/grafana.ini:ro",
			vol("grafana/provisioning")+":/etc/grafana/provisioning:ro",
		)
	return publish(c, cfg.GrafanaPort, cfg.GrafanaPort)
}
