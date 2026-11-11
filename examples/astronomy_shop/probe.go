// Package main: kit-probe Dockerfile-as-Go image that curls the storefront.
package main

import (
	"context"
	"fmt"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
)

// kitProbeDockerfile builds a tiny alpine image with curl that hits the store.
func kitProbeDockerfile() *kit.Dockerfile {
	return kit.NewDockerFile().
		From("alpine:3.20").
		Run("apk add --no-cache curl").
		Label("example", "astronomy-shop-kit-probe")
}

// kitProbeBuild returns a kit.Build wrapping the probe Dockerfile.
func kitProbeBuild() kit.Build {
	return kit.NewBuild().Dockerfile(kitProbeDockerfile())
}

// kitProbeContainer curls the storefront. Prefer compose DNS (frontend-proxy);
// fall back to host.docker.internal for Docker Desktop when the Engine-API
// probe cannot join the compose network. extra_hosts host-gateway works on
// Desktop and Linux 20.10+ without a GOOS branch.
func kitProbeContainer(cfg *Config) *kit.Container {
	viaNet := fmt.Sprintf("http://frontend-proxy:%s/", cfg.EnvoyPort)
	viaHost := fmt.Sprintf("http://host.docker.internal:%s/", cfg.EnvoyPort)
	script := fmt.Sprintf(
		"curl -fsS --max-time 20 %s >/dev/null || curl -fsS --max-time 20 %s >/dev/null",
		viaNet, viaHost,
	)
	return kit.NewContainer("kit-probe").
		Image(probeImage).
		Endpoint(networkName).
		ExtraHost("host.docker.internal", "host-gateway").
		Command("sh", "-c", script+" && echo OK")
}

// imageExists reports whether ref is already present locally (skip rebuild).
func imageExists(ctx context.Context, ref string) (bool, error) {
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}
	defer cli.Close()
	img, err := cli.ImageInspect(ctx, ref, nil)
	if client.IsErrNotFound(err) {
		return false, nil
	}
	return img != nil, err
}
