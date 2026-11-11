// Package main: compose project assembly with shop / full / observability profiles.
package main

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
)

// SetupProject builds the Astronomy Shop Compose project for the given profiles.
// Collector observability exporters are enabled when "observability" is active.
func SetupProject(cfg *Config, baseVolumesPath string, profiles []string) *kit.Project {
	project := kit.NewProject(projectName)
	vol := func(p string) string { return filepath.Join(baseVolumesPath, p) }
	withObs := slices.Contains(profiles, profileObservability)
	withFull := slices.Contains(profiles, profileFull)

	add := func(name string, c *kit.Container, profile string) *kit.Service {
		s := kit.NewService(name, c).Profiles(profile)
		project.WithService(s)
		return s
	}

	// Infrastructure (shop)
	add("astronomy-db", astronomyDB(cfg, vol), profileShop)
	add("valkey-cart", valkeyCart(cfg), profileShop)
	add("otel-collector", otelCollector(cfg, vol, withObs), profileShop)
	add("flagd", flagd(cfg, vol), profileShop)

	add("flagd-ui", flagdUI(cfg, vol), profileShop).
		DependsOn("otel-collector").
		DependsOn("flagd")

	add("ad", ad(cfg), profileShop).
		DependsOn("otel-collector").
		DependsOn("flagd")

	add("cart", cart(cfg), profileShop).
		DependsOnHealthy("valkey-cart").
		DependsOn("otel-collector").
		DependsOn("flagd")

	add("currency", currency(cfg), profileShop).
		DependsOn("otel-collector")

	add("email", email(cfg), profileShop).
		DependsOn("otel-collector")

	add("payment", payment(cfg), profileShop).
		DependsOn("otel-collector").
		DependsOn("flagd")

	add("product-catalog", productCatalog(cfg, vol), profileShop).
		DependsOn("otel-collector").
		DependsOn("flagd").
		DependsOnHealthy("astronomy-db")

	add("quote", quote(cfg), profileShop).
		DependsOn("otel-collector")

	add("recommendation", recommendation(cfg), profileShop).
		DependsOn("product-catalog").
		DependsOn("otel-collector").
		DependsOn("flagd")

	add("shipping", shipping(cfg), profileShop).
		DependsOn("otel-collector").
		DependsOn("flagd")

	add("image-provider", imageProvider(cfg), profileShop).
		DependsOn("otel-collector")

	checkoutSvc := add("checkout", checkout(cfg, withFull), profileShop).
		DependsOn("cart").
		DependsOnHealthy("currency").
		DependsOnHealthy("email").
		DependsOnHealthy("payment").
		DependsOn("product-catalog").
		DependsOn("shipping").
		DependsOn("otel-collector").
		DependsOn("flagd")

	add("frontend", frontend(cfg), profileShop).
		DependsOnHealthy("ad").
		DependsOn("cart").
		DependsOn("checkout").
		DependsOnHealthy("currency").
		DependsOn("product-catalog").
		DependsOnHealthy("quote").
		DependsOnHealthy("recommendation").
		DependsOn("shipping").
		DependsOn("otel-collector").
		DependsOnHealthy("image-provider").
		DependsOn("flagd")

	add("load-generator", loadGenerator(cfg), profileShop).
		DependsOnHealthy("frontend").
		DependsOn("flagd")

	add("frontend-proxy", frontendProxy(cfg), profileShop).
		DependsOnHealthy("frontend").
		DependsOn("load-generator").
		DependsOnHealthy("flagd-ui")

	// Full: Kafka event path
	add("kafka", kafka(cfg), profileFull)
	add("accounting", accounting(cfg), profileFull).
		DependsOn("otel-collector").
		DependsOnHealthy("kafka")
	add("fraud-detection", fraudDetection(cfg), profileFull).
		DependsOn("otel-collector").
		DependsOnHealthy("kafka")
	if slices.Contains(profiles, profileFull) {
		checkoutSvc.DependsOnHealthy("kafka")
	}

	// Observability: traces + metrics UIs
	add("jaeger", jaeger(cfg, vol), profileObservability)
	add("prometheus", prometheus(cfg, vol), profileObservability)
	add("grafana", grafana(cfg, vol), profileObservability).
		DependsOn("prometheus").
		DependsOn("jaeger")

	// IPv4-only and attachable: Docker Desktop DNS otherwise prefers AAAA, and
	// the kit-probe (Engine API, not Compose) must be able to join this network.
	project.WithNetwork(networkName, kit.ProjectNetwork{
		Driver:      "bridge",
		Attachable:  true,
		DisableIPv6: true,
	})
	return project
}

// parseProfiles splits a comma-separated -profile flag into a unique ordered list.
// Default when empty: shop,observability.
func parseProfiles(raw string) []string {
	if raw == "" {
		return []string{profileShop, profileObservability}
	}
	seen := map[string]bool{}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		switch p {
		case profileShop, profileFull, profileObservability:
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	if !seen[profileShop] {
		// Always include shop as the base store profile.
		out = append([]string{profileShop}, out...)
	}
	if len(out) == 0 {
		return []string{profileShop, profileObservability}
	}
	return out
}
