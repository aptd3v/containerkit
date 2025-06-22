// Package response provides thin wrappers around the docker client response types.
package response

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/checkpoint"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/volume"
)

// ContainerCreate is a wrapper around the container.CreateResponse type.
type ContainerCreate struct {
	container.CreateResponse
}

// ContainerSummary is a wrapper around the container.Summary type.
type ContainerSummary struct {
	container.Summary
}

// ContainerInspect is a wrapper around the container.InspectResponse type.
type ContainerInspect struct {
	container.InspectResponse
}

// ContainerWait is a wrapper around the container.WaitResponse type.
type ContainerWait struct {
	container.WaitResponse
}

// ContainerStatsReader is a wrapper around the container.StatsResponseReader type.
type ContainerStatsReader struct {
	container.StatsResponseReader
}

// ContainerExecCreate is a wrapper around the container.ExecCreateResponse type.
type ContainerExecCreate struct {
	container.ExecCreateResponse
}

// ContainerHijackedResponse is a wrapper around the types.HijackedResponse type.
type ContainerHijackedResponse struct {
	types.HijackedResponse
}

// ContainerPruneReport is a wrapper around the container.PruneReport type.
type ContainerPruneReport struct {
	container.PruneReport
}

// ContainerCommitResponse is a wrapper around the container.CommitResponse type.
type ContainerCommitResponse struct {
	container.CommitResponse
}

// ContainerFilesystemChange is a wrapper around the container.FilesystemChange type.
type ContainerFilesystemChange struct {
	container.FilesystemChange
}

// ContainerUpdateResponse is a wrapper around the container.UpdateResponse type.
type ContainerUpdateResponse struct {
	container.UpdateResponse
}

// ContainerTopResponse is a wrapper around the container.TopResponse type.
type ContainerTopResponse struct {
	container.TopResponse
}

// ContainerExecInspect is a wrapper around the container.ExecInspect type.
type ContainerExecInspect struct {
	container.ExecInspect
}

// ContainerStatsOneShot is a wrapper around the container.StatsResponseReader type.
type ContainerStatsOneShot struct {
	container.StatsResponseReader
}

// ContainerCheckpointSummary is a wrapper around the checkpoint.Summary type.
type ContainerCheckpointSummary struct {
	checkpoint.Summary
}

// ContainerPathStat is a wrapper around the container.PathStat type.
type ContainerPathStat struct {
	container.PathStat
}

// ImageSummary is a wrapper around the image.Summary type.
type ImageSummary struct {
	image.Summary
}

// ImageInspect is a wrapper around the image.InspectResponse type.
type ImageInspect struct {
	image.InspectResponse
}

// ImageHistoryItem is a wrapper around the image.HistoryResponseItem type.
type ImageHistoryItem struct {
	image.HistoryResponseItem
}

// ImageBuild is a wrapper around the image.BuildResponse type.
type ImageBuild struct {
	build.ImageBuildResponse
}

// ImageDelete is a wrapper around the image.DeleteResponse type.
type ImageDelete struct {
	image.DeleteResponse
}

// ImageSearchResult is a wrapper around the image.SearchResponse type.
type ImageSearchResult struct {
	registry.SearchResult
}

// ImageLoad is a wrapper around the image.LoadResponse type.
type ImageLoad struct {
	image.LoadResponse
}

// PruneReport is a wrapper around the image.PruneReport type.
type PruneReport struct {
	image.PruneReport
}

// Volume is a wrapper around the volume.Volume type.
type Volume struct {
	volume.Volume
}

type VolumeList struct {
	Volumes  []*Volume
	Warnings []string
}
type VolumePruneReport struct {
	volume.PruneReport
}
