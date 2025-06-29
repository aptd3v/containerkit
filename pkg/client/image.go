package client

import (
	"context"
	"io"

	"github.com/aptd3v/go-contain/pkg/client/options/image/build"
	"github.com/aptd3v/go-contain/pkg/client/options/image/create"
	"github.com/aptd3v/go-contain/pkg/client/options/image/imports"
	"github.com/aptd3v/go-contain/pkg/client/options/image/list"
	"github.com/aptd3v/go-contain/pkg/client/options/image/load"
	"github.com/aptd3v/go-contain/pkg/client/options/image/prune"
	"github.com/aptd3v/go-contain/pkg/client/options/image/pull"
	"github.com/aptd3v/go-contain/pkg/client/options/image/remove"
	"github.com/aptd3v/go-contain/pkg/client/options/image/save"
	"github.com/aptd3v/go-contain/pkg/client/options/image/search"
	"github.com/aptd3v/go-contain/pkg/client/response"
	dBuild "github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

/*
ImagePull requests the docker host to pull an image from a remote registry.
It executes the privileged function if the operation is unauthorized and it
tries one more time. It's up to the caller to handle the io.ReadCloser and close it properly.
*/
func (c *Client) ImagePull(ctx context.Context, ref string, setters ...pull.SetImagePullOption) (io.ReadCloser, error) {
	op := image.PullOptions{}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	return c.wrapped.ImagePull(ctx, ref, op)
}

// ImageCreate creates a new image based on the parent options. It returns the JSON content in the response body.
func (c *Client) ImageCreate(ctx context.Context, ref string, setters ...create.SetImageCreateOption) (io.ReadCloser, error) {
	op := image.CreateOptions{}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	return c.wrapped.ImageCreate(ctx, ref, op)
}

/*
ImageList returns a list of images in the docker host.

Experimental: Setting the [options.Manifest] will
populate image.Summary.Manifests with information about image manifests.
This is experimental and might change in the future without any backward compatibility.
*/
func (c *Client) ImageList(ctx context.Context, setters ...list.SetImageListOption) ([]response.ImageSummary, error) {
	op := image.ListOptions{
		Filters: filters.NewArgs(),
	}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	resp, err := c.wrapped.ImageList(ctx, op)
	if err != nil {
		return nil, err
	}
	summaries := make([]response.ImageSummary, len(resp))
	for i, image := range resp {
		summaries[i] = response.ImageSummary{
			Summary: image,
		}
	}
	return summaries, nil
}

// ImageInspect returns the image information.
func (c *Client) ImageInspect(ctx context.Context, ref string) (*response.ImageInspect, error) {
	inspect, err := c.wrapped.ImageInspect(ctx, ref)
	if err != nil {
		return nil, err
	}
	return &response.ImageInspect{
		InspectResponse: inspect,
	}, nil
}

// ImageHistory returns the changes in an image in history format.
func (c *Client) ImageHistory(ctx context.Context, ref string) ([]response.ImageHistoryItem, error) {
	resp, err := c.wrapped.ImageHistory(ctx, ref)
	if err != nil {
		return nil, err
	}
	items := make([]response.ImageHistoryItem, len(resp))
	for i, item := range resp {
		items[i] = response.ImageHistoryItem{
			HistoryResponseItem: item,
		}
	}
	return items, nil
}

// ImageBuild sends a request to the daemon to build images. The Body in the response implements
// an io.ReadCloser and it's up to the caller to close it.
func (c *Client) ImageBuild(ctx context.Context, setters ...build.SetImageBuildOption) (*response.ImageBuild, error) {
	op := dBuild.ImageBuildOptions{}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	build, err := c.wrapped.ImageBuild(ctx, nil, op)
	if err != nil {
		return nil, err
	}
	return &response.ImageBuild{
		ImageBuildResponse: build,
	}, nil
}

// ImageSave retrieves one or more images from the docker host as an io.ReadCloser.
func (c *Client) ImageSave(ctx context.Context, setters ...save.SetImageSaveOption) (io.ReadCloser, error) {
	op := save.ImageSaveOptions{}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	return c.wrapped.ImageSave(ctx, op.ImageIDs, client.ImageSaveWithPlatforms(op.Platforms...))
}

// ImageTag tags an image in the docker host.
func (c *Client) ImageTag(ctx context.Context, source, target string) error {
	return c.wrapped.ImageTag(ctx, source, target)
}

// ImageRemove removes an image from the docker host.
func (c *Client) ImageRemove(ctx context.Context, ref string, setters ...remove.SetImageRemoveOption) ([]response.ImageDelete, error) {
	op := image.RemoveOptions{}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	resp, err := c.wrapped.ImageRemove(ctx, ref, op)
	if err != nil {
		return nil, err
	}
	deletes := make([]response.ImageDelete, len(resp))
	for i, del := range resp {
		deletes[i] = response.ImageDelete{
			DeleteResponse: del,
		}
	}
	return deletes, nil
}

// ImageSearch makes the docker host search by a term in a remote registry. The list of results is not sorted in any fashion.
func (c *Client) ImageSearch(ctx context.Context, term string, setters ...search.SetImageSearchOption) ([]response.ImageSearchResult, error) {
	op := registry.SearchOptions{
		Filters: filters.NewArgs(),
	}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	resp, err := c.wrapped.ImageSearch(ctx, term, op)
	if err != nil {
		return nil, err
	}
	results := make([]response.ImageSearchResult, len(resp))
	for i, result := range resp {
		results[i] = response.ImageSearchResult{
			SearchResult: result,
		}
	}
	return results, nil
}

// ImagePush requests the docker host to push an image to a remote registry. It executes
// the privileged function if the operation is unauthorized and it tries one more time.
// It's up to the caller to handle the io.ReadCloser and close it properly.
func (c *Client) ImagePush(ctx context.Context, ref string) (io.ReadCloser, error) {
	return c.wrapped.ImagePush(ctx, ref, image.PushOptions{})
}

// ImageImport creates a new image based on the source options. It returns the JSON content in the response body.
func (c *Client) ImageImport(ctx context.Context, ref string, setters ...imports.SetImageImportOption) (io.ReadCloser, error) {
	op := imports.ImageImportOptions{}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	return c.wrapped.ImageImport(ctx, image.ImportSource{
		Source:     op.Source,
		SourceName: op.SourceName,
	}, ref, image.ImportOptions{
		Tag:      op.Tag,
		Message:  op.Message,
		Changes:  op.Changes,
		Platform: op.Platform,
	})
}

/*
ImageLoad loads an image in the docker host from the client host.
It's up to the caller to close the io.ReadCloser in the ImageLoadResponse returned by this function.

WithPlatform is an optional parameter that specifies the platform to
load from the provided multi-platform image. This is only has effect if the input image is a multi-platform image.
*/
func (c *Client) ImageLoad(ctx context.Context, setters ...load.SetImageLoadOption) (*response.ImageLoad, error) {
	op := load.ImageLoadOptions{}
	for _, setter := range setters {
		if setter != nil {
			if err := setter(&op); err != nil {
				return nil, err
			}
		}
	}
	resp, err := c.wrapped.ImageLoad(ctx, op.Input, client.ImageLoadWithPlatforms(op.Platforms...), client.ImageLoadWithQuiet(op.Quiet))
	if err != nil {
		return nil, err
	}
	return &response.ImageLoad{
		LoadResponse: resp,
	}, nil
}

// ImagesPrune requests the daemon to delete unused data
func (c *Client) ImagesPrune(ctx context.Context, setters ...prune.SetImagePruneOption) (*response.PruneReport, error) {
	filters := filters.NewArgs()
	for _, setter := range setters {
		if setter != nil {
			if err := setter(filters); err != nil {
				return nil, err
			}
		}
	}
	resp, err := c.wrapped.ImagesPrune(ctx, filters)
	if err != nil {
		return nil, err
	}
	return &response.PruneReport{
		PruneReport: resp,
	}, nil
}
