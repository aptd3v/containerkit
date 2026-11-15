// Package plugin holds the Engine plugin e2e test. It builds a local volume
// plugin (config.json + rootfs) following
// https://docs.docker.com/engine/extend/#developing-a-plugin and drives
// pkg/kit Create/Set/Enable/Inspect/List/Install/Upgrade/Disable/Remove.
//
// Default `go test ./...` does not run these tests. Against a live Linux Engine:
//
//	go test -tags e2e -count=1 ./e2e/plugin
package plugin
