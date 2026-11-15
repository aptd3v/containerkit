// Package registry holds the authenticated Docker registry e2e test.
// A QEMU VM (preferred) or DinD stack runs Engine plus registry:2 with
// htpasswd auth. The test drives pkg/kit Login, PushImage, and PullImage.
//
// Default `go test ./...` does not run these tests. Start the environment
// with e2e/registry/up.sh, then:
//
//	go test -tags e2e -count=1 ./e2e/registry
package registry
