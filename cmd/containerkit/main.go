// Command containerkit runs a containerkit plan or compose YAML against a Docker daemon.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Set by GoReleaser / tools/crossbuild ldflags.
var (
	version = "0.8.3"
	commit  = "none"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
