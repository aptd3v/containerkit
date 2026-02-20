package codegen

import (
	"bytes"
	"sort"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/imports"
)

const (
	pkgCreate   = "github.com/aptd3v/go-contain/pkg/create"
	pkgCC       = "github.com/aptd3v/go-contain/pkg/create/config/cc"
	pkgHealth   = "github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	pkgHC       = "github.com/aptd3v/go-contain/pkg/create/config/hc"
	pkgNC       = "github.com/aptd3v/go-contain/pkg/create/config/nc"
	pkgEndpoint   = "github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint"
	pkgEndpointIPAM = "github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint/ipam"
	pkgPC         = "github.com/aptd3v/go-contain/pkg/create/config/pc"
	pkgSC         = "github.com/aptd3v/go-contain/pkg/create/config/sc"
	pkgBuild      = "github.com/aptd3v/go-contain/pkg/create/config/sc/build"
	pkgBuildUlimit = "github.com/aptd3v/go-contain/pkg/create/config/sc/build/ulimit"
	pkgDeploy     = "github.com/aptd3v/go-contain/pkg/create/config/sc/deploy"
	pkgUpdate     = "github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/update"
	pkgResource   = "github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource"
	pkgDevice     = "github.com/aptd3v/go-contain/pkg/create/config/sc/deploy/resource/device"
	pkgSecretSvc  = "github.com/aptd3v/go-contain/pkg/create/config/sc/secrets/secretservice"
	pkgNetwork  = "github.com/aptd3v/go-contain/pkg/create/config/sc/network"
	pkgPool     = "github.com/aptd3v/go-contain/pkg/create/config/sc/network/pool"
	pkgVolume   = "github.com/aptd3v/go-contain/pkg/create/config/sc/volume"
	pkgCompose  = "github.com/aptd3v/go-contain/pkg/compose"
	pkgUp       = "github.com/aptd3v/go-contain/pkg/compose/options/up"
	pkgDown     = "github.com/aptd3v/go-contain/pkg/compose/options/down"
	pkgTools    = "github.com/aptd3v/go-contain/pkg/tools"
)

// Generate produces a Go source file that builds the given compose project using go-contain.
func Generate(project *types.Project, opts Options) ([]byte, error) {
	pkg := opts.PackageName
	if pkg == "" {
		pkg = "main"
	}
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = project.Name
	}
	if projectName == "" {
		projectName = "project"
	}

	f := jen.NewFile(pkg)

	// IIFE body: project := NewProject; [Networks]; [Volumes]; Services; return project
	body := []jen.Code{
		jen.Id("project").Op(":=").Qual(pkgCreate, "NewProject").Call(jen.Lit(projectName)),
	}
	networkStmts := genNetworks(project)
	if len(networkStmts) > 0 {
		body = append(body, jen.Line(), jen.Comment("// Networks"))
		body = append(body, networkStmts...)
	}
	volumeStmts := genVolumes(project)
	if len(volumeStmts) > 0 {
		body = append(body, jen.Line(), jen.Comment("// Volumes"))
		body = append(body, volumeStmts...)
	}
	serviceNames := make([]string, 0, len(project.Services))
	for name := range project.Services {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)
	if len(serviceNames) > 0 {
		body = append(body, jen.Line(), jen.Comment("// Services"))
		for _, name := range serviceNames {
			svc := project.Services[name]
			_, _, callStmt := genServiceFunc(name, &svc)
			body = append(body, callStmt)
		}
	}
	body = append(body, jen.Line(), jen.Return(jen.Id("project")))

	// Emit container and optional service-config function per service before the project var/func.
	for _, name := range serviceNames {
		svc := project.Services[name]
		containerFunc, serviceConfigFunc, _ := genServiceFunc(name, &svc)
		f.Add(containerFunc)
		if serviceConfigFunc != nil {
			f.Add(serviceConfigFunc)
		}
	}

	if pkg == "main" {
		f.Var().Id("project").Op("=").Parens(
			jen.Func().Params().Op("*").Qual(pkgCreate, "Project").Block(body...),
		).Call()
	} else {
		// Library package: expose a function so other packages can use it.
		f.Comment("// WithCompose returns the compose project for use by other packages.")
		f.Func().Id("WithCompose").Params().Op("*").Qual(pkgCreate, "Project").Block(body...)
	}

	if opts.EmitMain {
		projectExpr := jen.Id("project")
		if pkg != "main" {
			projectExpr = jen.Id("WithCompose").Call()
		}
		upArgs := []jen.Code{
			jen.Id("ctx"),
			jen.Qual(pkgUp, "WithWriter").Call(jen.Qual("os", "Stdout")),
			jen.Qual(pkgUp, "WithRemoveOrphans").Call(),
		}
		if len(opts.Profiles) > 0 {
			profileArgs := make([]jen.Code, len(opts.Profiles))
			for i, p := range opts.Profiles {
				profileArgs[i] = jen.Lit(p)
			}
			upArgs = append(upArgs, jen.Qual(pkgUp, "WithProfiles").Call(profileArgs...))
		}
		downArgs := []jen.Code{jen.Qual("context", "Background").Call(), jen.Qual(pkgDown, "WithRemoveOrphans").Call()}
		if len(opts.Profiles) > 0 {
			profileArgs := make([]jen.Code, len(opts.Profiles))
			for i, p := range opts.Profiles {
				profileArgs[i] = jen.Lit(p)
			}
			downArgs = append(downArgs, jen.Qual(pkgDown, "WithProfiles").Call(profileArgs...))
		}
		f.Line()
		f.Func().Id("main").Params().Block(
			jen.Id("ctx").Op(",").Id("cancel").Op(":=").Qual("context", "WithCancel").Call(jen.Qual("context", "Background").Call()),
			jen.Defer().Id("cancel").Call(),
			jen.Id("sigCh").Op(":=").Make(jen.Chan().Qual("os", "Signal"), jen.Lit(1)),
			jen.Qual("os/signal", "Notify").Call(jen.Id("sigCh"), jen.Qual("os", "Interrupt"), jen.Qual("syscall", "SIGTERM")),
			jen.Go().Func().Params().Block(
				jen.Op("<-").Id("sigCh"),
				jen.Id("cancel").Call(),
			).Call(),
			jen.Id("app").Op(":=").Qual(pkgCompose, "NewCompose").Call(projectExpr),
			jen.Defer().Func().Params().Block(
				jen.If(jen.Err().Op(":=").Id("app").Dot("Down").Call(downArgs...), jen.Err().Op("!=").Nil()).Block(
					jen.Qual("log", "Print").Call(jen.Err()),
				),
			).Call(),
			jen.Err().Op(":=").Id("app").Dot("Up").Call(upArgs...),
			jen.If(jen.Err().Op("!=").Nil().Op("&&").Op("!").Qual("errors", "Is").Call(jen.Err(), jen.Qual("context", "Canceled"))).Block(
				jen.Qual("log", "Fatal").Call(jen.Err()),
			),
		)
	}

	var buf bytes.Buffer
	if err := f.Render(&buf); err != nil {
		return nil, err
	}
	out := buf.Bytes()
	// Break long container/service chains so each .With* starts on its own line.
	out = bytes.ReplaceAll(out, []byte(").With"), []byte(").\n\t\tWith"))
	// Put each With* argument on its own line (including nested calls like deploy.WithRollbackConfig(update.With...)).
	pkgs := []string{"cc.", "hc.", "nc.", "sc.", "health.", "network.", "build.", "deploy.", "endpoint.", "resource.", "ipam.", "update.", "device.", "secretservice.", "ulimit."}
	for _, pkg := range pkgs {
		for n := 1; n <= 6; n++ {
			old := append(bytes.Repeat([]byte(")"), n), []byte(", "+pkg)...)
			new := append(bytes.Repeat([]byte(")"), n), []byte(",\n\t\t\t"+pkg)...)
			out = bytes.ReplaceAll(out, old, new)
		}
	}
	// Opening paren on new line for multi-line With* config calls.
	for _, name := range []string{
		"WithContainerConfig(", "WithHostConfig(", "WithNetworkConfig(", "WithPlatformConfig(", "WithHealthCheck(",
		"WithUpdateConfig(", "WithRollbackConfig(",
		"WithResourceLimits(", "WithResourceReservations(",
	} {
		out = bytes.ReplaceAll(out, []byte(name), []byte(name[:len(name)-1]+"(\n\t\t\t"))
	}
	// Run goimports-style formatting: group imports (stdlib, blank, third-party), gofmt.
	formatted, err := imports.Process("", out, nil)
	if err != nil {
		return nil, err
	}
	return formatted, nil
}
