// Package swarm holds the 4-node Docker Swarm e2e cluster used to exercise
// pkg/kit against real Engine daemons.
//
// Default `go test ./...` does not run these tests. Start the cluster with
// e2e/swarm/up.sh, then:
//
//	go test -tags e2e -count=1 ./e2e/swarm
package swarm
