// Astronomy Shop services. Images, ports, and env match the OpenTelemetry demo.
package main

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/selinux"
)

func base() []ServiceOption {
	return []ServiceOption{
		Restart(Policy(UnlessStopped)),
		Network(Name(networkName)),
	}
}

func seconds(n int) time.Duration { return time.Duration(n) * time.Second }

func mustPort(s string) uint16 {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("port %q: %w", s, err))
	}
	return uint16(n)
}

func expose(ports ...string) []ServiceOption {
	out := make([]ServiceOption, 0, len(ports))
	for _, p := range ports {
		out = append(out, Expose(mustPort(p), TCP))
	}
	return out
}

func publish(host, container string) ServiceOption {
	return Publish(Protocol(TCP), HostPort(mustPort(host)), ContainerPort(mustPort(container)))
}

func healthShell(cmd string, start, interval, timeout, retries int) ServiceOption {
	return Health(
		Test("CMD-SHELL", cmd),
		StartPeriod(seconds(start)),
		Interval(seconds(interval)),
		Timeout(seconds(timeout)),
		Retries(retries),
	)
}

func healthCmd(start, interval, timeout, retries int, argv ...string) ServiceOption {
	return Health(
		Test(append([]string{"CMD"}, argv...)...),
		StartPeriod(seconds(start)),
		Interval(seconds(interval)),
		Timeout(seconds(timeout)),
		Retries(retries),
	)
}

func bashTCPHealth(port string, start, interval, timeout, retries int) ServiceOption {
	return healthShell(fmt.Sprintf("bash -c 'echo > /dev/tcp/127.0.0.1/%s'", port), start, interval, timeout, retries)
}

func ncHealth(port string, start, interval, timeout, retries int) ServiceOption {
	return healthShell("nc -z 127.0.0.1 "+port, start, interval, timeout, retries)
}

func grpcHealth(port string) ServiceOption {
	return healthCmd(90, 5, 10, 30, "/bin/grpc_health_probe", "-addr=127.0.0.1:"+port)
}

func envs(m map[string]string) []ServiceOption {
	out := make([]ServiceOption, 0, len(m))
	for _, key := range slices.Sorted(maps.Keys(m)) {
		out = append(out, Env(key, m[key]))
	}
	return out
}

func mount(spec string) ServiceOption {
	flags := ""
	body := spec
	if i := strings.LastIndex(spec, ":"); i >= 0 && !strings.HasPrefix(spec[i+1:], "/") {
		flags = spec[i+1:]
		body = spec[:i]
	}
	host, container, ok := strings.Cut(body, ":")
	if !ok || host == "" || container == "" {
		panic(fmt.Errorf("volume bind %q", spec))
	}
	opts := []VolumeOption{Kind(Bind), Source(host), Target(container), SELinux(selinux.Shared)}
	if strings.Contains(flags, "ro") {
		opts = append(opts, ReadOnly(true))
	}
	return Volume(opts...)
}

func binds(specs ...string) []ServiceOption {
	out := make([]ServiceOption, 0, len(specs))
	for _, spec := range specs {
		out = append(out, mount(spec))
	}
	return out
}

func cat(groups ...[]ServiceOption) []ServiceOption {
	var out []ServiceOption
	for _, group := range groups {
		out = append(out, group...)
	}
	return out
}

func astronomyDB(cfg *Config, vol func(string) string) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.PostgresImage),
		Env("POSTGRES_PASSWORD", cfg.PostgresPassword),
		Command("postgres", "-c", "shared_preload_libraries=pg_stat_statements"),
		healthShell("pg_isready -U astronomy_user -d astronomy_db", 30, 5, 10, 20),
	}, binds(vol("postgresql/init.sql")+":/docker-entrypoint-initdb.d/init.sql:ro,z"))
}

func valkeyCart(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.ValkeyImage),
		User("valkey"),
		healthShell("valkey-cli ping", 5, 5, 3, 10),
	})
}

func otelCollector(cfg *Config, vol func(string) string, withObs bool) []ServiceOption {
	cmd := []string{"--config=/etc/otelcol-config.yml"}
	specs := []string{
		vol("otel-collector/otelcol-config.yml") + ":/etc/otelcol-config.yml:ro,z",
		vol("otel-collector/otelcol-config-extras.yml") + ":/etc/otelcol-config-extras.yml:ro,z",
	}
	if withObs {
		cmd = append(cmd, "--config=/etc/otelcol-config-observability.yml")
		specs = append(specs, vol("otel-collector/otelcol-config-observability.yml")+":/etc/otelcol-config-observability.yml:ro,z")
	}
	cmd = append(cmd, "--config=/etc/otelcol-config-extras.yml")
	return cat(base(), []ServiceOption{
		Image(cfg.CollectorImage),
		User("0:0"),
		Command(cmd...),
	}, envs(map[string]string{
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
	}), binds(specs...), expose(cfg.CollectorGRPCPort, cfg.CollectorHTTPPort))
}

func flagd(cfg *Config, vol func(string) string) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.FlagdImage),
		Command("start", "--uri", "file:./etc/flagd/demo.flagd.json"),
	}, envs(mergeEnv(
		cfg.commonOTEL("flagd", "low", cfg.otlpHTTP()),
		map[string]string{
			"OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
			"GOMEMLIMIT":                  "60MiB",
		},
	)), binds(vol("flagd")+":/etc/flagd:ro,z"), expose(cfg.FlagdPort, cfg.FlagdOFREPPort))
}

func flagdUI(cfg *Config, vol func(string) string) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("flagd-ui")),
		bashTCPHealth(cfg.FlagdUIPort, 10, 5, 5, 20),
	}, envs(mergeEnv(
		cfg.commonOTEL("flagd-ui", "low", cfg.otlpHTTP()),
		map[string]string{
			"FLAGD_UI_PORT":   cfg.FlagdUIPort,
			"SECRET_KEY_BASE": "yYrECL4qbNwleYInGJYvVnSkwJuSQJ4ijPTx5tirGUXrbznFIBFVJdPl5t6O9ASw",
			"PHX_HOST":        "localhost",
		},
	)), binds(vol("flagd")+":/app/data:z"), expose(cfg.FlagdUIPort))
}

func ad(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("ad")),
		bashTCPHealth(cfg.ADPort, 20, 5, 5, 20),
	}, envs(mergeEnv(
		cfg.commonOTEL("ad", "medium", cfg.otlpHTTP()),
		map[string]string{
			"AD_PORT":            cfg.ADPort,
			"AD_PROMETHEUS_PORT": cfg.ADPrometheusPort,
			"FLAGD_HOST":         "flagd",
			"FLAGD_PORT":         cfg.FlagdPort,
			"OTEL_LOGS_EXPORTER": "otlp",
			"OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION": "latest",
		},
	)), expose(cfg.ADPort))
}

func cart(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{Image(cfg.demoImage("cart"))}, envs(mergeEnv(
		cfg.commonOTEL("cart", "high", cfg.otlpGRPC()),
		map[string]string{
			"CART_PORT":             cfg.CartPort,
			"FLAGD_HOST":            "flagd",
			"FLAGD_PORT":            cfg.FlagdPort,
			"VALKEY_ADDR":           "valkey-cart:" + cfg.ValkeyPort,
			"ASPNETCORE_HTTP_PORTS": cfg.CartPort,
		},
	)), expose(cfg.CartPort))
}

func checkout(cfg *Config, withKafka bool) []ServiceOption {
	env := map[string]string{
		"CHECKOUT_PORT":               cfg.CheckoutPort,
		"FLAGD_HOST":                  "flagd",
		"FLAGD_PORT":                  cfg.FlagdPort,
		"CART_ADDR":                   "cart:" + cfg.CartPort,
		"CURRENCY_ADDR":               "currency:" + cfg.CurrencyPort,
		"EMAIL_ADDR":                  "http://email:" + cfg.EmailPort,
		"PAYMENT_ADDR":                "payment:" + cfg.PaymentPort,
		"PRODUCT_CATALOG_ADDR":        "product-catalog:" + cfg.ProductCatalogPort,
		"SHIPPING_ADDR":               "http://shipping:" + cfg.ShippingPort,
		"GOMEMLIMIT":                  "64MiB",
		"OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
		"OTEL_GO_X_OBSERVABILITY":     "true",
	}
	if withKafka {
		env["KAFKA_ADDR"] = "kafka:" + cfg.KafkaPort
		env["KAFKA_TOPIC"] = "orders"
	}
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("checkout")),
		grpcHealth(cfg.CheckoutPort),
	}, envs(mergeEnv(cfg.commonOTEL("checkout", "critical", cfg.otlpHTTP()), env)), expose(cfg.CheckoutPort))
}

func currency(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("currency")),
		healthShell("nc -z 127.0.0.1 "+cfg.CurrencyPort, 10, 5, 5, 20),
	}, envs(mergeEnv(
		cfg.commonOTEL("currency", "high", cfg.otlpGRPC()),
		map[string]string{
			"CURRENCY_PORT": cfg.CurrencyPort,
			"IPV6_ENABLED":  cfg.IPV6Enabled,
			"VERSION":       cfg.ImageVersion,
		},
	)), expose(cfg.CurrencyPort))
}

func email(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("email")),
		healthCmd(10, 5, 5, 20, "ruby", "-e", fmt.Sprintf("require 'socket'; TCPSocket.new('127.0.0.1', %s).close", cfg.EmailPort)),
	}, envs(mergeEnv(
		cfg.commonOTEL("email", "medium", cfg.otlpHTTP()),
		map[string]string{
			"APP_ENV":    "production",
			"EMAIL_PORT": cfg.EmailPort,
			"FLAGD_HOST": "flagd",
			"FLAGD_PORT": cfg.FlagdPort,
		},
	)), expose(cfg.EmailPort))
}

func payment(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("payment")),
		healthCmd(10, 5, 5, 20, "/nodejs/bin/node", "-e",
			fmt.Sprintf("const net=require('net');const c=new net.Socket(); c.setTimeout(2000); c.connect(%s,'127.0.0.1',()=>{c.destroy();process.exit(0)}); c.on('error',()=>process.exit(1)); c.on('timeout',()=>process.exit(1))", cfg.PaymentPort)),
	}, envs(mergeEnv(
		cfg.commonOTEL("payment", "critical", cfg.otlpGRPC()),
		map[string]string{
			"IPV6_ENABLED":                cfg.IPV6Enabled,
			"PAYMENT_PORT":                cfg.PaymentPort,
			"FLAGD_HOST":                  "flagd",
			"FLAGD_PORT":                  cfg.FlagdPort,
			"OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
		},
	)), expose(cfg.PaymentPort))
}

func productCatalog(cfg *Config, vol func(string) string) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("product-catalog")),
		healthCmd(90, 5, 10, 36, "/bin/grpc_health_probe", "-addr=127.0.0.1:"+cfg.ProductCatalogPort),
	}, envs(mergeEnv(
		cfg.commonOTEL("product-catalog", "high", cfg.otlpGRPC()),
		map[string]string{
			"PRODUCT_CATALOG_PORT":          cfg.ProductCatalogPort,
			"FLAGD_HOST":                    "flagd",
			"FLAGD_PORT":                    cfg.FlagdPort,
			"GOMEMLIMIT":                    "64MiB",
			"OTEL_GO_X_OBSERVABILITY":       "true",
			"OTEL_CONFIG_FILE":              "/otel-config.yml",
			"OTEL_SEMCONV_STABILITY_OPT_IN": "database",
			"DB_CONNECTION_STRING": fmt.Sprintf(
				"postgres://astronomy_user:%s@astronomy-db/astronomy_db?sslmode=disable",
				cfg.PostgresAstronomyPassword,
			),
		},
	)), binds(vol("otel-config.yml")+":/otel-config.yml:ro,z"), expose(cfg.ProductCatalogPort))
}

func quote(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("quote")),
		healthCmd(10, 5, 5, 20, "php", "-r", fmt.Sprintf("fsockopen('127.0.0.1', %s) or die('fail');", cfg.QuotePort)),
	}, envs(mergeEnv(
		cfg.commonOTEL("quote", "low", cfg.otlpHTTP()),
		map[string]string{
			"IPV6_ENABLED":                      cfg.IPV6Enabled,
			"QUOTE_PORT":                        cfg.QuotePort,
			"OTEL_PHP_AUTOLOAD_ENABLED":         "true",
			"OTEL_PHP_INTERNAL_METRICS_ENABLED": "true",
		},
	)), expose(cfg.QuotePort))
}

func recommendation(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("recommendation")),
		ncHealth(cfg.RecommendationPort, 20, 5, 5, 20),
	}, envs(mergeEnv(
		cfg.commonOTEL("recommendation", "medium", cfg.otlpHTTP()),
		map[string]string{
			"RECOMMENDATION_PORT":  cfg.RecommendationPort,
			"PRODUCT_CATALOG_ADDR": "product-catalog:" + cfg.ProductCatalogPort,
			"FLAGD_HOST":           "flagd",
			"FLAGD_PORT":           cfg.FlagdPort,
			"OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION": "latest",
		},
	)), expose(cfg.RecommendationPort))
}

func shipping(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("shipping")),
		healthCmd(10, 5, 5, 20, "/app/healthcheck"),
	}, envs(mergeEnv(
		cfg.commonOTEL("shipping", "high", cfg.otlpHTTP()),
		map[string]string{
			"IPV6_ENABLED":                cfg.IPV6Enabled,
			"SHIPPING_PORT":               cfg.ShippingPort,
			"QUOTE_ADDR":                  "http://quote:" + cfg.QuotePort,
			"FLAGD_HOST":                  "flagd",
			"FLAGD_PORT":                  cfg.FlagdPort,
			"OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
		},
	)), expose(cfg.ShippingPort))
}

func imageProvider(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("image-provider")),
		healthCmd(10, 5, 5, 20, "wget", "--quiet", "--tries=1", "--spider", "http://127.0.0.1:"+cfg.ImageProviderPort+"/status"),
	}, envs(map[string]string{
		"IMAGE_PROVIDER_PORT":      cfg.ImageProviderPort,
		"OTEL_COLLECTOR_HOST":      "otel-collector",
		"OTEL_COLLECTOR_PORT_GRPC": cfg.CollectorGRPCPort,
		"OTEL_RESOURCE_ATTRIBUTES": cfg.resourceAttributes("low"),
		"OTEL_SERVICE_NAME":        "image-provider",
	}), expose(cfg.ImageProviderPort))
}

func frontend(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("frontend")),
		healthCmd(60, 10, 10, 5, "/nodejs/bin/node", "-e",
			fmt.Sprintf("require('net').connect(%s,require('os').hostname(),function(){process.exit(0)}).on('error',function(){process.exit(1)})", cfg.FrontendPort)),
	}, envs(mergeEnv(
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
	)), expose(cfg.FrontendPort))
}

func loadGenerator(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("load-generator")),
		healthCmd(30, 5, 5, 20, "python3", "-c",
			fmt.Sprintf("import socket; s=socket.create_connection(('127.0.0.1',%s),2); s.close()", cfg.LocustWebPort)),
		publish(cfg.LocustWebPort, cfg.LocustWebPort),
	}, envs(mergeEnv(
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
	)))
}

func frontendProxy(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("frontend-proxy")),
		healthCmd(10, 5, 5, 20, "bash", "-c", "echo > /dev/tcp/127.0.0.1/"+cfg.EnvoyPort),
		publish(cfg.EnvoyPort, cfg.EnvoyPort),
	}, envs(map[string]string{
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
	}))
}

func kafka(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.demoImage("kafka")),
		healthShell("nc -z kafka 9092", 10, 5, 10, 10),
	}, envs(mergeEnv(
		cfg.commonOTEL("kafka", "low", cfg.otlpHTTP()),
		map[string]string{
			"KAFKA_ADVERTISED_LISTENERS":              "PLAINTEXT://kafka:9092",
			"KAFKA_LISTENERS":                         "PLAINTEXT://kafka:9092,CONTROLLER://kafka:9093",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":          "1@kafka:9093",
			"KAFKA_HEAP_OPTS":                         "-Xmx400m -Xms400m",
			"OTEL_EXPERIMENTAL_SDK_TELEMETRY_VERSION": "latest",
		},
	)), expose(cfg.KafkaPort))
}

func accounting(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{Image(cfg.demoImage("accounting"))}, envs(mergeEnv(
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
	)))
}

func fraudDetection(cfg *Config) []ServiceOption {
	return cat(base(), []ServiceOption{Image(cfg.demoImage("fraud-detection"))}, envs(mergeEnv(
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
	)))
}

func jaeger(cfg *Config, vol func(string) string) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.JaegerImage),
		Command("--config=file:/etc/jaeger/config.yml"),
		publish(cfg.JaegerUIPort, cfg.JaegerUIPort),
	}, envs(map[string]string{
		"JAEGER_HOST":              "0.0.0.0",
		"JAEGER_GRPC_PORT":         cfg.JaegerGRPCPort,
		"PROMETHEUS_ADDR":          "prometheus:" + cfg.PrometheusPort,
		"OTEL_COLLECTOR_HOST":      "otel-collector",
		"OTEL_COLLECTOR_PORT_HTTP": cfg.CollectorHTTPPort,
		"MEMORY_MAX_TRACES":        "25000",
	}), binds(
		vol("jaeger/config.yml")+":/etc/jaeger/config.yml:ro,z",
		vol("jaeger/ui-config.json")+":/etc/jaeger/ui-config.json:ro,z",
	))
}

func prometheus(cfg *Config, vol func(string) string) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.PrometheusImage),
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
		),
		publish(cfg.PrometheusPort, cfg.PrometheusPort),
	}, binds(vol("prometheus/prometheus-config.yaml")+":/etc/prometheus/prometheus-config.yaml:ro,z"))
}

func grafana(cfg *Config, vol func(string) string) []ServiceOption {
	return cat(base(), []ServiceOption{
		Image(cfg.GrafanaImage),
		publish(cfg.GrafanaPort, cfg.GrafanaPort),
	}, binds(
		vol("grafana/grafana.ini")+":/etc/grafana/grafana.ini:ro,z",
		vol("grafana/provisioning")+":/etc/grafana/provisioning:ro,z",
	))
}
