// kit-probe image that curls the storefront.
package main

import (
	"context"
	"fmt"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

func kitProbeDockerfile() *Dockerfile {
	return NewDockerfile().
		From("alpine", "3.20").
		Run("apk add --no-cache curl").
		Label("example", "astronomy-shop-kit-probe")
}

// kitProbeContainer curls the storefront. Prefer compose DNS (frontend-proxy);
// fall back to host.docker.internal for Docker Desktop when the Engine-API
// probe cannot join the compose network. extra_hosts host-gateway works on
// Desktop and Linux 20.10+ without a GOOS branch.
func kitProbeContainer(cfg *Config) *Container {
	viaNet := fmt.Sprintf("http://frontend-proxy:%s/", cfg.EnvoyPort)
	viaHost := fmt.Sprintf("http://host.docker.internal:%s/", cfg.EnvoyPort)
	script := fmt.Sprintf(
		"curl -fsS --max-time 20 %s >/dev/null || curl -fsS --max-time 20 %s >/dev/null",
		viaNet, viaHost,
	)
	return NewContainer(
		Name("kit-probe"),
		Image(probeImage),
		Network(Name(networkName)),
		ExtraHost("host.docker.internal", "host-gateway"),
		Command("sh", "-c", script+" && echo OK"),
	)
}

func imageExists(ctx context.Context, ref string) (bool, error) {
	cli, err := Connect(FromEnv())
	if err != nil {
		return false, err
	}
	defer cli.Close()
	_, err = InspectImage(ctx, cli.Engine(), ref)
	if IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}
