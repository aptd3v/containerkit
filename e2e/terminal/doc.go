// Package terminal holds the Engine TTY e2e test. It starts a real container
// with a TTY, attaches through pkg/kit AttachTerminal / ExecTerminal, sends
// input, and asserts interactive output.
//
// Default `go test ./...` does not run these tests. Against a live Engine:
//
//	go test -tags e2e -count=1 ./e2e/terminal
package terminal
