package kit_test

import "github.com/aptd3v/containerkit/pkg/kit"

// Compile-time proof that setter option types are visible outside the package
// and assignable to the interfaces pkg.go.dev lists under Implemented by.
var (
	_ kit.ContainerOption  = kit.NameOpt{}
	_ kit.ContainerOption  = kit.Name("worker")
	_ kit.ContainerOption  = kit.ImageOpt{}
	_ kit.ContainerOption  = kit.SpecOpt{}
	_ kit.PortOption       = kit.HostPortOpt{}
	_ kit.PortOption       = kit.HostPort(8080)
	_ kit.LogOption        = kit.DriverValue{}
	_ kit.LogOption        = kit.DriverOptKV{}
	_ kit.ClientOption     = kit.FromEnvOpt{}
	_ kit.ComposeNetOption = kit.IPAMOptKV{}
)
