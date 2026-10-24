package taskflow_test

import (
	"context"
	"fmt"

	"github.com/aptd3v/containerkit/pkg/taskflow"
)

func Example() {
	flow := taskflow.New("demo", nil)
	hello := flow.Task("hello", func(context.Context) error {
		fmt.Println("hello")
		return nil
	})
	done := flow.Task("done", func(context.Context) error {
		fmt.Println("done")
		return nil
	})
	hello.Precede(done)
	if err := flow.Run(context.Background(), nil); err != nil {
		fmt.Println(err)
	}
	// Output:
	// hello
	// done
}
