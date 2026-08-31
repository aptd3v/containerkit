package client

import (
	"context"
	"io"
	"runtime"

	"github.com/aptd3v/containerkit/pkg/client/auth"
	dBuild "github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ImagePull is options for ImagePull.
type ImagePull struct {
	Auth            auth.Auth
	All             bool
	CurrentPlatform bool
	Platform        string
	PrivilegeFunc   func(ctx context.Context) (string, error)
}

func (o *ImagePull) apply() (image.PullOptions, error) {
	if o == nil {
		return image.PullOptions{}, nil
	}
	op := image.PullOptions{
		All:           o.All,
		Platform:      o.Platform,
		PrivilegeFunc: o.PrivilegeFunc,
	}
	if o.CurrentPlatform {
		op.Platform = runtime.GOARCH
	}
	if o.Auth.Username != "" || o.Auth.Password != "" || o.Auth.ServerAddress != "" {
		encoded, err := auth.AuthToBase64(o.Auth)
		if err != nil {
			return image.PullOptions{}, err
		}
		op.RegistryAuth = encoded
	}
	return op, nil
}

// ImageCreate is options for ImageCreate.
type ImageCreate struct {
	Auth     auth.Auth
	Platform string
}

func (o *ImageCreate) apply() (image.CreateOptions, error) {
	if o == nil {
		return image.CreateOptions{}, nil
	}
	op := image.CreateOptions{Platform: o.Platform}
	if o.Auth.Username != "" || o.Auth.Password != "" || o.Auth.ServerAddress != "" {
		encoded, err := auth.AuthToBase64(o.Auth)
		if err != nil {
			return image.CreateOptions{}, err
		}
		op.RegistryAuth = encoded
	}
	return op, nil
}

// ImageList is options for ImageList.
type ImageList struct {
	All            bool
	SharedSize     bool
	ContainerCount bool
	Manifests      bool
	Filters        []Filter
}

func (o *ImageList) apply() image.ListOptions {
	op := image.ListOptions{Filters: argsFromFilters(nil)}
	if o == nil {
		return op
	}
	op.All = o.All
	op.SharedSize = o.SharedSize
	op.ContainerCount = o.ContainerCount
	op.Manifests = o.Manifests
	op.Filters = argsFromFilters(o.Filters)
	return op
}

// ImageBuild is options for ImageBuild. Fields match Docker's ImageBuildOptions.
type ImageBuild struct {
	dBuild.ImageBuildOptions
}

func (o *ImageBuild) apply() dBuild.ImageBuildOptions {
	if o == nil {
		return dBuild.ImageBuildOptions{}
	}
	return o.ImageBuildOptions
}

// ImageSave is options for ImageSave.
type ImageSave struct {
	ImageIDs  []string
	Platforms []ocispec.Platform
}

// ImageRemove is options for ImageRemove.
type ImageRemove struct {
	Force         bool
	PruneChildren bool
	Platforms     []ocispec.Platform
}

func (o *ImageRemove) apply() image.RemoveOptions {
	if o == nil {
		return image.RemoveOptions{}
	}
	return image.RemoveOptions{
		Force:         o.Force,
		PruneChildren: o.PruneChildren,
		Platforms:     o.Platforms,
	}
}

// ImageSearch is options for ImageSearch.
type ImageSearch struct {
	Auth          auth.Auth
	PrivilegeFunc func(ctx context.Context) (string, error)
	Filters       []Filter
	Limit         int
}

func (o *ImageSearch) apply() (registry.SearchOptions, error) {
	op := registry.SearchOptions{Filters: argsFromFilters(nil)}
	if o == nil {
		return op, nil
	}
	op.Filters = argsFromFilters(o.Filters)
	op.Limit = o.Limit
	op.PrivilegeFunc = o.PrivilegeFunc
	if o.Auth.Username != "" || o.Auth.Password != "" || o.Auth.ServerAddress != "" {
		encoded, err := auth.AuthToBase64(o.Auth)
		if err != nil {
			return registry.SearchOptions{}, err
		}
		op.RegistryAuth = encoded
	}
	return op, nil
}

// ImageImport is options for ImageImport.
type ImageImport struct {
	Source     io.Reader
	SourceName string
	Tag        string
	Message    string
	Changes    []string
	Platform   string
}

// ImageLoad is options for ImageLoad.
type ImageLoad struct {
	Input     io.Reader
	Quiet     bool
	Platforms []ocispec.Platform
}

// ImagePrune is options for ImagesPrune.
type ImagePrune struct {
	Filters []Filter
}

func (o *ImagePrune) apply() filters.Args {
	if o == nil {
		return argsFromFilters(nil)
	}
	return argsFromFilters(o.Filters)
}
