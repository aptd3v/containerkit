// Compose project assembly with shop / full / observability profiles.
package main

import (
	"path/filepath"
	"slices"
	"strings"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

// SetupProject builds the Astronomy Shop Compose project for the given profiles.
// Collector observability exporters are enabled when "observability" is active.
func SetupProject(cfg *Config, baseVolumesPath string, profiles []string) *Compose {
	vol := func(p string) string { return filepath.Join(baseVolumesPath, p) }
	withObs := slices.Contains(profiles, profileObservability)
	withFull := slices.Contains(profiles, profileFull)

	app := NewCompose(projectName, AddNetwork(
		Name(networkName),
		Driver("bridge"),
		Attachable(true),
		EnableIPv6(false),
	))

	add := func(name, profile string, opts []ServiceOption, extra ...ServiceOption) {
		all := append([]ServiceOption{Name(name), Profile(profile)}, opts...)
		all = append(all, extra...)
		app.WithService(NewService(all...))
	}

	add("astronomy-db", profileShop, astronomyDB(cfg, vol))
	add("valkey-cart", profileShop, valkeyCart(cfg))
	add("otel-collector", profileShop, otelCollector(cfg, vol, withObs))
	add("flagd", profileShop, flagd(cfg, vol))

	add("flagd-ui", profileShop, flagdUI(cfg, vol),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("flagd")),
	)
	add("ad", profileShop, ad(cfg),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("flagd")),
	)
	add("cart", profileShop, cart(cfg),
		DependsOn(Name("valkey-cart"), Condition(Healthy)),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("flagd")),
	)
	add("currency", profileShop, currency(cfg),
		DependsOn(Name("otel-collector")),
	)
	add("email", profileShop, email(cfg),
		DependsOn(Name("otel-collector")),
	)
	add("payment", profileShop, payment(cfg),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("flagd")),
	)
	add("product-catalog", profileShop, productCatalog(cfg, vol),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("flagd")),
		DependsOn(Name("astronomy-db"), Condition(Healthy)),
	)
	add("quote", profileShop, quote(cfg),
		DependsOn(Name("otel-collector")),
	)
	add("recommendation", profileShop, recommendation(cfg),
		DependsOn(Name("product-catalog")),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("flagd")),
	)
	add("shipping", profileShop, shipping(cfg),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("flagd")),
	)
	add("image-provider", profileShop, imageProvider(cfg),
		DependsOn(Name("otel-collector")),
	)

	checkoutOpts := checkout(cfg, withFull)
	checkoutDeps := []ServiceOption{
		DependsOn(Name("cart")),
		DependsOn(Name("currency"), Condition(Healthy)),
		DependsOn(Name("email"), Condition(Healthy)),
		DependsOn(Name("payment"), Condition(Healthy)),
		DependsOn(Name("product-catalog")),
		DependsOn(Name("shipping")),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("flagd")),
	}
	if withFull {
		checkoutDeps = append(checkoutDeps, DependsOn(Name("kafka"), Condition(Healthy)))
	}
	add("checkout", profileShop, checkoutOpts, checkoutDeps...)

	add("frontend", profileShop, frontend(cfg),
		DependsOn(Name("ad"), Condition(Healthy)),
		DependsOn(Name("cart")),
		DependsOn(Name("checkout")),
		DependsOn(Name("currency"), Condition(Healthy)),
		DependsOn(Name("product-catalog")),
		DependsOn(Name("quote"), Condition(Healthy)),
		DependsOn(Name("recommendation"), Condition(Healthy)),
		DependsOn(Name("shipping")),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("image-provider"), Condition(Healthy)),
		DependsOn(Name("flagd")),
	)
	add("load-generator", profileShop, loadGenerator(cfg),
		DependsOn(Name("frontend"), Condition(Healthy)),
		DependsOn(Name("flagd")),
	)
	add("frontend-proxy", profileShop, frontendProxy(cfg),
		DependsOn(Name("frontend"), Condition(Healthy)),
		DependsOn(Name("load-generator")),
		DependsOn(Name("flagd-ui"), Condition(Healthy)),
	)

	add("kafka", profileFull, kafka(cfg))
	add("accounting", profileFull, accounting(cfg),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("kafka"), Condition(Healthy)),
	)
	add("fraud-detection", profileFull, fraudDetection(cfg),
		DependsOn(Name("otel-collector")),
		DependsOn(Name("kafka"), Condition(Healthy)),
	)

	add("jaeger", profileObservability, jaeger(cfg, vol))
	add("prometheus", profileObservability, prometheus(cfg, vol))
	add("grafana", profileObservability, grafana(cfg, vol),
		DependsOn(Name("prometheus")),
		DependsOn(Name("jaeger")),
	)
	return app
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
		out = append([]string{profileShop}, out...)
	}
	if len(out) == 0 {
		return []string{profileShop, profileObservability}
	}
	return out
}

// withProfiles keeps services whose profile is active. Unprofiled services stay.
func withProfiles(base *Compose, profiles ...string) (*Compose, error) {
	yml, err := base.ExportYAML()
	if err != nil {
		return nil, err
	}
	opts := make([]ImportOption, 0, len(profiles))
	for _, profile := range profiles {
		opts = append(opts, Profile(profile))
	}
	loaded := ImportYAML(yml, opts...)
	if err := loaded.Validate(); err != nil {
		return nil, err
	}
	return loaded, nil
}
