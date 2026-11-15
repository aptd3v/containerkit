package kit

import (
	clienttypes "github.com/aptd3v/containerkit/internal/client/types"
	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	clitypes "github.com/aptd3v/containerkit/internal/compose/cli/types"
	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

// Engine is the full Engine client surface. [NewEngine] returns one.
// Callers never import github.com/moby/moby/client: option and result types
// are aliased in this package. Fluent wrappers on [Container] and package
// functions such as [ListContainers] share the kit setter vocabulary.
type Engine = clienttypes.APIClient

// Composer is the Compose runtime surface. [NewComposer] returns one.
type Composer = composeapi.Compose

// Project is a loaded compose-go project. [LoadProject] and [Generate] return one.
type Project = composetypes.Project

// Compose SDK types used by [LoadProject] so callers never import compose/v5.
type (
	ProjectLoadOptions   = composeapi.ProjectLoadOptions
	LoadListener         = composeapi.LoadListener
	OCIOptions           = composeapi.OCIOptions
	UnsupportedAttribute = composeapi.UnsupportedAttribute
	ProjectOptionsFn     = clitypes.ProjectOptionsFn
)

// Created is the result of [Container.Create].
type Created struct {
	ID       string
	Warnings []string
}
