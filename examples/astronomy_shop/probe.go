// Package main: kit-probe Dockerfile-as-Go image that curls the storefront.
package main

import (
	"context"
	"fmt"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

// kitProbeDockerfile builds a tiny alpine image with curl that hits the store.
func kitProbeDockerfile(storeURL string) *kit.Dockerfile {
	return kit.NewDockerFile().
		From("alpine:3.20").
		Run("apk add --no-cache curl").
		Label("example", "astronomy-shop-kit-probe").
		Command("sh", "-c",
			fmt.Sprintf("curl -fsS --max-time 10 %s >/dev/null && echo OK", storeURL))
}

// kitProbeBuild returns a kit.Build wrapping the probe Dockerfile.
func kitProbeBuild(storeURL string) kit.Build {
	return kit.NewBuild().Dockerfile(kitProbeDockerfile(storeURL))
}

// kitProbeContainer runs the built probe once on the host network so it can
// reach the published frontend-proxy port (default http://127.0.0.1:8080/).
func kitProbeContainer() *kit.Container {
	return kit.NewContainer("kit-probe").
		Image(probeImage).
		NetworkMode(types.NetworkHost)
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
