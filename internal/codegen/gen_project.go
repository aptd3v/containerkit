package codegen

import (
	"sort"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/dave/jennifer/jen"
)

// genNetworks returns project.WithNetwork(...) statements for each network.
func genNetworks(project *types.Project) []jen.Code {
	var stmts []jen.Code
	names := make([]string, 0, len(project.Networks))
	for name := range project.Networks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		cfg := project.Networks[name]
		if cfg.External {
			continue
		}
		enableIPv6 := cfg.EnableIPv6 != nil && *cfg.EnableIPv6
		if cfg.Name == "" && len(cfg.Driver) == 0 && len(cfg.DriverOpts) == 0 && !cfg.Internal && !cfg.Attachable && !enableIPv6 && len(cfg.Labels) == 0 {
			stmts = append(stmts, jen.Id("project").Dot("WithNetwork").Call(jen.Lit(name)))
			continue
		}
		args := []jen.Code{jen.Lit(name)}
		if cfg.Driver != "" {
			args = append(args, jen.Qual(pkgNetwork, "WithDriver").Call(jen.Lit(cfg.Driver)))
		}
		for k, v := range cfg.DriverOpts {
			args = append(args, jen.Qual(pkgNetwork, "WithDriverOptions").Call(jen.Lit(k), jen.Lit(v)))
		}
		if cfg.Internal {
			args = append(args, jen.Qual(pkgNetwork, "WithInternal").Call())
		}
		if cfg.Attachable {
			args = append(args, jen.Qual(pkgNetwork, "WithAttachable").Call())
		}
		if enableIPv6 {
			args = append(args, jen.Qual(pkgNetwork, "WithEnableIPv6").Call())
		}
		for k, v := range cfg.Labels {
			args = append(args, jen.Qual(pkgNetwork, "WithLabel").Call(jen.Lit(k), jen.Lit(v)))
		}
		if len(cfg.Ipam.Config) > 0 {
			for _, ipam := range cfg.Ipam.Config {
				var poolArgs []jen.Code
				if ipam.Subnet != "" {
					poolArgs = append(poolArgs, jen.Qual(pkgPool, "WithSubnet").Call(jen.Lit(ipam.Subnet)))
				}
				if ipam.Gateway != "" {
					poolArgs = append(poolArgs, jen.Qual(pkgPool, "WithGateway").Call(jen.Lit(ipam.Gateway)))
				}
				if ipam.IPRange != "" {
					poolArgs = append(poolArgs, jen.Qual(pkgPool, "WithIpRange").Call(jen.Lit(ipam.IPRange)))
				}
				for k, v := range ipam.AuxiliaryAddresses {
					poolArgs = append(poolArgs, jen.Qual(pkgPool, "WithAuxiliaryAddresses").Call(jen.Lit(k), jen.Lit(v)))
				}
				if len(poolArgs) > 0 {
					args = append(args, jen.Qual(pkgNetwork, "WithIpamPool").Call(poolArgs...))
				}
			}
		}
		stmts = append(stmts, jen.Id("project").Dot("WithNetwork").Call(args...))
	}
	return stmts
}

// genVolumes returns project.WithVolume(...) statements for each volume.
func genVolumes(project *types.Project) []jen.Code {
	var stmts []jen.Code
	names := make([]string, 0, len(project.Volumes))
	for name := range project.Volumes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		cfg := project.Volumes[name]
		if cfg.External {
			continue
		}
		if cfg.Driver == "" && len(cfg.DriverOpts) == 0 && len(cfg.Labels) == 0 {
			stmts = append(stmts, jen.Id("project").Dot("WithVolume").Call(jen.Lit(name)))
			continue
		}
		args := []jen.Code{jen.Lit(name)}
		if cfg.Driver != "" {
			args = append(args, jen.Qual(pkgVolume, "WithDriver").Call(jen.Lit(cfg.Driver)))
		}
		for k, v := range cfg.DriverOpts {
			args = append(args, jen.Qual(pkgVolume, "WithDriverOptions").Call(jen.Lit(k), jen.Lit(v)))
		}
		for k, v := range cfg.Labels {
			args = append(args, jen.Qual(pkgVolume, "WithLabel").Call(jen.Lit(k), jen.Lit(v)))
		}
		stmts = append(stmts, jen.Id("project").Dot("WithVolume").Call(args...))
	}
	return stmts
}
