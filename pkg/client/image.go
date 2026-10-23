package client

import (
	"context"
	"io"

	"github.com/aptd3v/containerkit/pkg/client/response"
	sdk "github.com/moby/moby/client"
)

/*
ImagePull requests the docker host to pull an image from a remote registry.
It executes the privileged function if the operation is unauthorized and it
tries one more time. It's up to the caller to handle the io.ReadCloser and close it properly.
*/
func (c *Client) ImagePull(ctx context.Context, ref string, opt *ImagePull) (io.ReadCloser, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	return c.wrapped.ImagePull(ctx, ref, op)
}

// ImageCreate pulls ref and creates a local image. The engine no longer exposes a separate create call.
func (c *Client) ImageCreate(ctx context.Context, ref string, opt *ImageCreate) (io.ReadCloser, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	return c.wrapped.ImagePull(ctx, ref, op)
}

/*
ImageList returns a list of images in the docker host.

Setting Manifests populates image summaries with manifest information.
Identity requires API 1.54 and includes image identity with the manifests.
*/
func (c *Client) ImageList(ctx context.Context, opt *ImageList) ([]response.ImageSummary, error) {
	resp, err := c.wrapped.ImageList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	summaries := make([]response.ImageSummary, len(resp.Items))
	for i, item := range resp.Items {
		summaries[i] = response.ImageSummary{Summary: item}
	}
	return summaries, nil
}

// ImageInspect returns the image information. Nil options inspect without manifest or platform filters.
func (c *Client) ImageInspect(ctx context.Context, ref string, opt *ImageInspect) (*response.ImageInspect, error) {
	inspect, err := c.wrapped.ImageInspect(ctx, ref, opt.options()...)
	if err != nil {
		return nil, err
	}
	return &response.ImageInspect{InspectResponse: inspect.InspectResponse}, nil
}

// ImageHistory returns the changes in an image in history format.
func (c *Client) ImageHistory(ctx context.Context, ref string) ([]response.ImageHistoryItem, error) {
	resp, err := c.wrapped.ImageHistory(ctx, ref)
	if err != nil {
		return nil, err
	}
	items := make([]response.ImageHistoryItem, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = response.ImageHistoryItem{HistoryResponseItem: item}
	}
	return items, nil
}

// ImageBuild sends a request to the daemon to build images. The Body in the response implements
// an io.ReadCloser and it's up to the caller to close it.
func (c *Client) ImageBuild(ctx context.Context, context io.Reader, opt *ImageBuild) (*response.ImageBuild, error) {
	build, err := c.wrapped.ImageBuild(ctx, context, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.ImageBuild{Body: build.Body}, nil
}

// ImageSave retrieves one or more images from the docker host as an io.ReadCloser.
func (c *Client) ImageSave(ctx context.Context, opt *ImageSave) (io.ReadCloser, error) {
	if opt == nil {
		return c.wrapped.ImageSave(ctx, nil)
	}
	if len(opt.Platforms) > 0 {
		return c.wrapped.ImageSave(ctx, opt.ImageIDs, sdk.ImageSaveWithPlatforms(opt.Platforms...))
	}
	return c.wrapped.ImageSave(ctx, opt.ImageIDs)
}

// ImageTag tags an image in the docker host.
func (c *Client) ImageTag(ctx context.Context, source, target string) error {
	_, err := c.wrapped.ImageTag(ctx, sdk.ImageTagOptions{Source: source, Target: target})
	return err
}

// ImageRemove removes an image from the docker host.
func (c *Client) ImageRemove(ctx context.Context, ref string, opt *ImageRemove) ([]response.ImageDelete, error) {
	resp, err := c.wrapped.ImageRemove(ctx, ref, opt.apply())
	if err != nil {
		return nil, err
	}
	deletes := make([]response.ImageDelete, len(resp.Items))
	for i, del := range resp.Items {
		deletes[i] = response.ImageDelete{DeleteResponse: del}
	}
	return deletes, nil
}

// ImageSearch makes the docker host search by a term in a remote registry. The list of results is not sorted in any fashion.
func (c *Client) ImageSearch(ctx context.Context, term string, opt *ImageSearch) ([]response.ImageSearchResult, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	resp, err := c.wrapped.ImageSearch(ctx, term, op)
	if err != nil {
		return nil, err
	}
	results := make([]response.ImageSearchResult, len(resp.Items))
	for i, result := range resp.Items {
		results[i] = response.ImageSearchResult{SearchResult: result}
	}
	return results, nil
}

// ImagePush requests the docker host to push an image to a remote registry. It executes
// the privileged function if the operation is unauthorized and it tries one more time.
// Nil options push with an empty auth header. It's up to the caller to handle the io.ReadCloser and close it properly.
func (c *Client) ImagePush(ctx context.Context, ref string, opt *ImagePush) (io.ReadCloser, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	return c.wrapped.ImagePush(ctx, ref, op)
}

// ImageImport creates a new image based on the source options. It returns the JSON content in the response body.
func (c *Client) ImageImport(ctx context.Context, ref string, opt *ImageImport) (io.ReadCloser, error) {
	if opt == nil {
		opt = &ImageImport{}
	}
	options := sdk.ImageImportOptions{
		Tag:     opt.Tag,
		Message: opt.Message,
		Changes: opt.Changes,
	}
	if platform, ok := parsePlatform(opt.Platform); ok {
		options.Platform = platform
	}
	return c.wrapped.ImageImport(ctx, sdk.ImageImportSource{
		Source:     opt.Source,
		SourceName: opt.SourceName,
	}, ref, options)
}

/*
ImageLoad loads an image in the docker host from the client host.
It's up to the caller to close the io.ReadCloser in the ImageLoad response returned by this function.

Platforms selects which platforms to load from a multi-platform image.
*/
func (c *Client) ImageLoad(ctx context.Context, opt *ImageLoad) (*response.ImageLoad, error) {
	if opt == nil {
		opt = &ImageLoad{}
	}
	resp, err := c.wrapped.ImageLoad(ctx, opt.Input, sdk.ImageLoadWithPlatforms(opt.Platforms...), sdk.ImageLoadWithQuiet(opt.Quiet))
	if err != nil {
		return nil, err
	}
	return &response.ImageLoad{Body: resp}, nil
}

// ImagesPrune requests the daemon to delete unused data.
func (c *Client) ImagesPrune(ctx context.Context, opt *ImagePrune) (*response.PruneReport, error) {
	resp, err := c.wrapped.ImagePrune(ctx, sdk.ImagePruneOptions{Filters: opt.apply()})
	if err != nil {
		return nil, err
	}
	return &response.PruneReport{PruneReport: resp.Report}, nil
}

// BuildCancel asks the daemon to cancel the build identified by id.
func (c *Client) BuildCancel(ctx context.Context, id string) error {
	_, err := c.wrapped.BuildCancel(ctx, id, sdk.BuildCancelOptions{})
	return err
}

// BuildCachePrune deletes unused build-cache records.
func (c *Client) BuildCachePrune(ctx context.Context, opt *BuildCachePrune) (*response.BuildCachePrune, error) {
	pruned, err := c.wrapped.BuildCachePrune(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.BuildCachePrune{CachePruneReport: pruned.Report}, nil
}

// ImageAttestations returns in-toto statements attached to ref. This requires API 1.55.
func (c *Client) ImageAttestations(ctx context.Context, ref string, opt *ImageAttestations) ([]response.ImageAttestation, error) {
	resp, err := c.wrapped.ImageAttestations(ctx, ref, opt.options()...)
	if err != nil {
		return nil, err
	}
	items := make([]response.ImageAttestation, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = response.ImageAttestation{AttestationStatement: item}
	}
	return items, nil
}
