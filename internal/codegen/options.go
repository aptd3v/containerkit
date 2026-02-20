package codegen

// Options configures code generation.
type Options struct {
	// PackageName is the Go package name for the generated file (e.g. "main").
	PackageName string
	// EmitMain, if true, generates a func main() that creates the project and runs compose.Up.
	EmitMain bool
	// ProjectName overrides the compose project name in generated code. If empty, the loaded project name is used.
	ProjectName string
	// Profiles, if non-empty, are passed to Up/Down in the generated main so only services with these profiles run.
	Profiles []string
}
