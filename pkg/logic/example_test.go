package logic_test

import (
	"fmt"
	"os"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/logic"
)

func Example() {
	prod := true
	c := kit.NewContainer(
		logic.Group(
			kit.Name("api"),
			kit.Image("alpine:latest"),
			logic.When(prod, kit.Env("CACHE", "1")),
			logic.WhenFn(logic.Or(os.Getenv("DEBUG") != "", prod), kit.Env("LOG", "debug")),
			logic.Each([]string{"ROLE", "TIER"}, func(_ int, k string) any {
				return kit.Env(k, "api")
			}),
		),
	)
	fmt.Println(c.Name(), c.Image())
	// Output: api alpine:latest
}
