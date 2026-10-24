package client

import (
	"io"

	sdk "github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ImagePull is [sdk.ImagePullOptions].
type ImagePull sdk.ImagePullOptions

func (o *ImagePull) apply() sdk.ImagePullOptions { return sdk.ImagePullOptions(apply(o)) }

// ImageCreate is [sdk.ImagePullOptions]. The engine creates the image by pulling ref.
type ImageCreate sdk.ImagePullOptions

func (o *ImageCreate) apply() sdk.ImagePullOptions { return sdk.ImagePullOptions(apply(o)) }

// ImageList is [sdk.ImageListOptions].
type ImageList sdk.ImageListOptions

func (o *ImageList) apply() sdk.ImageListOptions { return sdk.ImageListOptions(apply(o)) }

// ImageBuild is [sdk.ImageBuildOptions].
type ImageBuild sdk.ImageBuildOptions

func (o *ImageBuild) apply() sdk.ImageBuildOptions { return sdk.ImageBuildOptions(apply(o)) }

// ImageSave is options for ImageSave. The SDK takes image IDs as a method argument
// plus functional options; this struct holds both.
type ImageSave struct {
	ImageIDs  []string
	Platforms []ocispec.Platform
}

// ImageRemove is [sdk.ImageRemoveOptions].
type ImageRemove sdk.ImageRemoveOptions

func (o *ImageRemove) apply() sdk.ImageRemoveOptions { return sdk.ImageRemoveOptions(apply(o)) }

// ImageSearch is [sdk.ImageSearchOptions].
type ImageSearch sdk.ImageSearchOptions

func (o *ImageSearch) apply() sdk.ImageSearchOptions { return sdk.ImageSearchOptions(apply(o)) }

// ImageImport combines [sdk.ImageImportSource] and [sdk.ImageImportOptions].
type ImageImport struct {
	Source     io.Reader
	SourceName string
	Tag        string
	Message    string
	Changes    []string
	Platform   ocispec.Platform
}

// ImageLoad is options for ImageLoad. The SDK takes the input reader as a method
// argument plus functional options.
type ImageLoad struct {
	Input     io.Reader
	Quiet     bool
	Platforms []ocispec.Platform
}

// BuildCachePrune is [sdk.BuildCachePruneOptions].
type BuildCachePrune sdk.BuildCachePruneOptions

func (o *BuildCachePrune) apply() sdk.BuildCachePruneOptions {
	return sdk.BuildCachePruneOptions(apply(o))
}

// ImagePrune is [sdk.ImagePruneOptions].
type ImagePrune sdk.ImagePruneOptions

func (o *ImagePrune) apply() sdk.ImagePruneOptions { return sdk.ImagePruneOptions(apply(o)) }

// ImagePush is [sdk.ImagePushOptions].
type ImagePush sdk.ImagePushOptions

func (o *ImagePush) apply() sdk.ImagePushOptions { return sdk.ImagePushOptions(apply(o)) }

// ImageInspect is options for ImageInspect. The SDK uses functional options.
type ImageInspect struct {
	Manifests bool
	Platform  *ocispec.Platform
}

func (o *ImageInspect) options() []sdk.ImageInspectOption {
	if o == nil {
		return nil
	}
	var opts []sdk.ImageInspectOption
	if o.Manifests {
		opts = append(opts, sdk.ImageInspectWithManifests(true))
	}
	if o.Platform != nil {
		opts = append(opts, sdk.ImageInspectWithPlatform(o.Platform))
	}
	return opts
}

// ImageAttestations is options for ImageAttestations. The SDK uses functional options.
type ImageAttestations struct {
	Platform       *ocispec.Platform
	PredicateTypes []string
}

func (o *ImageAttestations) options() []sdk.ImageAttestationsOption {
	if o == nil {
		return nil
	}
	var opts []sdk.ImageAttestationsOption
	if o.Platform != nil {
		opts = append(opts, sdk.ImageAttestationsWithPlatform(*o.Platform))
	}
	if len(o.PredicateTypes) > 0 {
		opts = append(opts, sdk.ImageAttestationsWithPredicateTypes(o.PredicateTypes...))
	}
	return opts
}
