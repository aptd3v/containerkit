// Package main: Astronomy Shop image tags, ports, and shared environment maps.
// Values mirror open-telemetry/opentelemetry-demo .env (IMAGE_VERSION 3.1.0).
package main

import "fmt"

const (
	profileShop          = "shop"
	profileFull          = "full"
	profileObservability = "observability"

	networkName = "astronomy-shop"
	projectName = "astronomy-shop"
	probeImage  = "astronomy-shop-kit-probe:local"
)

// Config holds pinned image tags and ports for the demo stack.
type Config struct {
	ImageName    string
	ImageVersion string

	CollectorImage string
	FlagdImage     string
	PostgresImage  string
	ValkeyImage    string
	GrafanaImage   string
	JaegerImage    string
	PrometheusImage string

	EnvPlatform string

	ADPort             string
	ADPrometheusPort   string
	CartPort           string
	CheckoutPort       string
	CurrencyPort       string
	EmailPort          string
	PaymentPort        string
	ProductCatalogPort string
	QuotePort          string
	RecommendationPort string
	ShippingPort       string
	FrontendPort       string
	EnvoyPort          string
	EnvoyAdminPort     string
	ImageProviderPort  string
	LocustWebPort      string
	FlagdPort          string
	FlagdOFREPPort     string
	FlagdUIPort        string
	ValkeyPort         string
	PostgresPort       string
	CollectorGRPCPort  string
	CollectorHTTPPort  string
	GrafanaPort        string
	JaegerUIPort       string
	JaegerGRPCPort     string
	PrometheusPort     string
	KafkaPort          string

	PostgresPassword           string
	PostgresAstronomyPassword  string
	PostgresMonitoringPassword string

	OTELServiceNamespace string
	IPV6Enabled          string
}

// DefaultConfig returns local-dev defaults matching upstream demo .env.
func DefaultConfig() *Config {
	return &Config{
		ImageName:       "ghcr.io/open-telemetry/demo",
		ImageVersion:    "3.1.0",
		CollectorImage:  "ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib:0.160.0",
		FlagdImage:      "ghcr.io/open-feature/flagd:v0.16.0",
		PostgresImage:   "postgres:18.6",
		ValkeyImage:     "ghcr.io/valkey-io/valkey:9.1.2-alpine3.24",
		GrafanaImage:    "grafana/grafana:13.2.1",
		JaegerImage:     "quay.io/jaegertracing/jaeger:2.21.0",
		PrometheusImage: "quay.io/prometheus/prometheus:v3.13.3",

		EnvPlatform: "local",

		ADPort:             "9555",
		ADPrometheusPort:   "9465",
		CartPort:           "7070",
		CheckoutPort:       "5050",
		CurrencyPort:       "7001",
		EmailPort:          "6060",
		PaymentPort:        "50051",
		ProductCatalogPort: "3550",
		QuotePort:          "8090",
		RecommendationPort: "9001",
		ShippingPort:       "50050",
		FrontendPort:       "8080",
		EnvoyPort:          "8080",
		EnvoyAdminPort:     "10000",
		ImageProviderPort:  "8081",
		LocustWebPort:      "8089",
		FlagdPort:          "8013",
		FlagdOFREPPort:     "8016",
		FlagdUIPort:        "4000",
		ValkeyPort:         "6379",
		PostgresPort:       "5432",
		CollectorGRPCPort:  "4317",
		CollectorHTTPPort:  "4318",
		GrafanaPort:        "3000",
		JaegerUIPort:       "16686",
		JaegerGRPCPort:     "4317",
		PrometheusPort:     "9090",
		KafkaPort:          "9092",

		PostgresPassword:           "changeit",
		PostgresAstronomyPassword:  "astronomy_password",
		PostgresMonitoringPassword: "monitoring_password",

		OTELServiceNamespace: "opentelemetry-demo",
		IPV6Enabled:          "false",
	}
}

func (c *Config) demoImage(svc string) string {
	return fmt.Sprintf("%s:%s-%s", c.ImageName, c.ImageVersion, svc)
}

func (c *Config) resourceAttributes(criticality string) string {
	base := fmt.Sprintf("service.namespace=%s,service.version=%s", c.OTELServiceNamespace, c.ImageVersion)
	if criticality == "" {
		return base
	}
	return base + ",service.criticality=" + criticality
}

func (c *Config) otlpGRPC() string {
	return fmt.Sprintf("http://otel-collector:%s", c.CollectorGRPCPort)
}

func (c *Config) otlpHTTP() string {
	return fmt.Sprintf("http://otel-collector:%s", c.CollectorHTTPPort)
}

func (c *Config) otlpMetricsPrefs() map[string]string {
	return map[string]string{
		"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE":         "cumulative",
		"OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION": "explicit_bucket_histogram",
	}
}

func (c *Config) commonOTEL(service, criticality, endpoint string) map[string]string {
	m := map[string]string{
		"OTEL_SERVICE_NAME":        service,
		"OTEL_RESOURCE_ATTRIBUTES": c.resourceAttributes(criticality),
		"OTEL_EXPORTER_OTLP_ENDPOINT": endpoint,
	}
	for k, v := range c.otlpMetricsPrefs() {
		m[k] = v
	}
	return m
}

func mergeEnv(parts ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, p := range parts {
		for k, v := range p {
			out[k] = v
		}
	}
	return out
}
