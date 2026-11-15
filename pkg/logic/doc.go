// Package logic groups kit setters and applies them under control-flow
// helpers. It is the containerkit equivalent of go-contain's tools package.
//
//	c := kit.NewContainer(
//		logic.Group(
//			kit.Name("api"),
//			kit.Image("alpine:latest"),
//			logic.When(isProd, kit.Env("CACHE", "1")),
//		),
//	)
//
// Helpers return [kit.Grouped], which implements every kit option
// interface, so a group can be passed to NewContainer, NewService, nested
// Health/Volume scopes, Engine actions, and Compose runtime methods.
package logic
