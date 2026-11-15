// Bring up one Alpine container that prints hello and exits.
package main

import (
	"context"
	"log"
	"os"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

func main() {
	cli, err := Connect(FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	app := NewCompose("simple").WithService(NewService(
		Name("hello"),
		Image("alpine:latest"),
		Command("echo", "hello"),
	))
	if err := cli.Up(context.Background(), app, Log(NewLogger(OutputStream(os.Stdout)))); err != nil {
		log.Fatal(err)
	}
}
