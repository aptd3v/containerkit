package client

import (
	"context"
	"io"
	"runtime"
	"strings"

	"github.com/aptd3v/containerkit/pkg/client/auth"
	sdk "github.com/moby/moby/client"
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

func (o *ImagePull) apply() (sdk.ImagePullOptions, error) {
	if o == nil {
		return sdk.ImagePullOptions{}, nil
	}
	op := sdk.ImagePullOptions{
		All:           o.All,
		PrivilegeFunc: o.PrivilegeFunc,
	}
	if o.CurrentPlatform {
		osName := runtime.GOOS
		if osName == "darwin" {
			osName = "linux"
		}
		op.Platforms = []ocispec.Platform{{OS: osName, Architecture: runtime.GOARCH}}
	} else if platform, ok := parsePlatform(o.Platform); ok {
		op.Platforms = []ocispec.Platform{platform}
	}
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.ImagePullOptions{}, err
	}
	op.RegistryAuth = encoded
	return op, nil
}

// ImageCreate is options for ImageCreate. The engine creates the image by pulling ref.
type ImageCreate struct {
	Auth     auth.Auth
	Platform string
}

func (o *ImageCreate) apply() (sdk.ImagePullOptions, error) {
	if o == nil {
		return sdk.ImagePullOptions{}, nil
	}
	op := sdk.ImagePullOptions{}
	if platform, ok := parsePlatform(o.Platform); ok {
		op.Platforms = []ocispec.Platform{platform}
	}
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.ImagePullOptions{}, err
	}
	op.RegistryAuth = encoded
	return op, nil
}

// ImageList is options for ImageList.
type ImageList struct {
	All        bool
	SharedSize bool
	Manifests  bool
	Identity   bool
	Filters    []Filter
}

func (o *ImageList) apply() sdk.ImageListOptions {
	op := sdk.ImageListOptions{Filters: argsFromFilters(nil)}
	if o == nil {
		return op
	}
	op.All = o.All
	op.SharedSize = o.SharedSize
	op.Manifests = o.Manifests
	op.Identity = o.Identity
	op.Filters = argsFromFilters(o.Filters)
	return op
}

// ImageBuild is options for ImageBuild. Fields match the engine image build options.
type ImageBuild struct {
	sdk.ImageBuildOptions
}

func (o *ImageBuild) apply() sdk.ImageBuildOptions {
	if o == nil {
		return sdk.ImageBuildOptions{}
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

func (o *ImageRemove) apply() sdk.ImageRemoveOptions {
	if o == nil {
		return sdk.ImageRemoveOptions{}
	}
	return sdk.ImageRemoveOptions{
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

func (o *ImageSearch) apply() (sdk.ImageSearchOptions, error) {
	op := sdk.ImageSearchOptions{Filters: argsFromFilters(nil)}
	if o == nil {
		return op, nil
	}
	op.Filters = argsFromFilters(o.Filters)
	op.Limit = o.Limit
	op.PrivilegeFunc = o.PrivilegeFunc
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.ImageSearchOptions{}, err
	}
	op.RegistryAuth = encoded
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

// BuildCachePrune is options for BuildCachePrune. Space fields are byte counts.
type BuildCachePrune struct {
	All           bool
	ReservedSpace int64
	MaxUsedSpace  int64
	MinFreeSpace  int64
	Filters       []Filter
}

func (o *BuildCachePrune) apply() sdk.BuildCachePruneOptions {
	if o == nil {
		return sdk.BuildCachePruneOptions{Filters: argsFromFilters(nil)}
	}
	return sdk.BuildCachePruneOptions{
		All:           o.All,
		ReservedSpace: o.ReservedSpace,
		MaxUsedSpace:  o.MaxUsedSpace,
		MinFreeSpace:  o.MinFreeSpace,
		Filters:       argsFromFilters(o.Filters),
	}
}

// ImagePrune is options for ImagesPrune.
type ImagePrune struct {
	Filters []Filter
}

func (o *ImagePrune) apply() sdk.Filters {
	if o == nil {
		return argsFromFilters(nil)
	}
	return argsFromFilters(o.Filters)
}

// ImagePush is options for ImagePush. Nil options push without registry auth.
type ImagePush struct {
	Auth          auth.Auth
	All           bool
	Platform      *ocispec.Platform
	PrivilegeFunc func(ctx context.Context) (string, error)
}

func (o *ImagePush) apply() (sdk.ImagePushOptions, error) {
	if o == nil {
		return sdk.ImagePushOptions{}, nil
	}
	op := sdk.ImagePushOptions{
		All:           o.All,
		Platform:      o.Platform,
		PrivilegeFunc: o.PrivilegeFunc,
	}
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.ImagePushOptions{}, err
	}
	op.RegistryAuth = encoded
	return op, nil
}

// ImageInspect is options for ImageInspect. Nil options inspect the default platform manifest.
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

// ImageAttestations is options for ImageAttestations. The daemon must support API 1.55.
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

func encodeAuth(a auth.Auth) (string, error) {
	if a.Username == "" && a.Password == "" && a.ServerAddress == "" {
		return "", nil
	}
	return auth.AuthToBase64(a)
}

func parsePlatform(value string) (ocispec.Platform, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ocispec.Platform{}, false
	}
	parts := strings.Split(value, "/")
	platform := ocispec.Platform{OS: parts[0]}
	if len(parts) > 1 {
		platform.Architecture = parts[1]
	}
	if len(parts) > 2 {
		platform.Variant = strings.Join(parts[2:], "/")
	}
	return platform, true
}
