// Package response provides thin wrappers around the docker client response types.
// Inspect results that include daemon JSON expose it on Raw.
package response

import (
	"bufio"
	"io"
	"net"

	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/api/types/checkpoint"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/plugin"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/api/types/volume"
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

// ContainerStatsReader is a single stats stream from the daemon.
type ContainerStatsReader struct {
	Body   io.ReadCloser
	OSType string
}

// ContainerExecCreate is a wrapper around the container.ExecCreateResponse type.
type ContainerExecCreate struct {
	container.ExecCreateResponse
}

// ContainerHijackedResponse is a hijacked connection to a container or exec stream.
type ContainerHijackedResponse struct {
	Conn   net.Conn
	Reader *bufio.Reader
}

// Close closes the hijacked connection.
func (h *ContainerHijackedResponse) Close() error {
	if h == nil || h.Conn == nil {
		return nil
	}
	return h.Conn.Close()
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

// ContainerExecInspect is a wrapper around the container.ExecInspectResponse type.
type ContainerExecInspect struct {
	container.ExecInspectResponse
}

// ContainerStatsOneShot is one stats sample from a container.
type ContainerStatsOneShot struct {
	Body   io.ReadCloser
	OSType string
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

// BuildCachePrune is the result of pruning the build cache.
type BuildCachePrune struct {
	build.CachePruneReport
}

// ImageBuild is the body of an image build response.
type ImageBuild struct {
	Body io.ReadCloser
}

// ImageDelete is a wrapper around the image.DeleteResponse type.
type ImageDelete struct {
	image.DeleteResponse
}

// ImageSearchResult is a wrapper around the image.SearchResponse type.
type ImageSearchResult struct {
	registry.SearchResult
}

// ImageLoad is the body of an image load response.
type ImageLoad struct {
	Body io.ReadCloser
}

// PruneReport is a wrapper around the image.PruneReport type.
type PruneReport struct {
	image.PruneReport
}

// Volume is a wrapper around the volume.Volume type.
type Volume struct {
	volume.Volume
}

// VolumeList is a wrapper around the volume.ListResponse type.
type VolumeList struct {
	Volumes  []*Volume
	Warnings []string
}

// VolumePruneReport is a wrapper around the volume.PruneReport type.
type VolumePruneReport struct {
	volume.PruneReport
}

// NetworkCreate is a wrapper around the network.CreateResponse type.
type NetworkCreate struct {
	network.CreateResponse
}

// NetworkInspect is a wrapper around the network.InspectResponse type.
type NetworkInspect struct {
	network.Inspect
}

// NetworkSummary is a wrapper around the network.Summary type.
type NetworkSummary struct {
	network.Summary
}

// NetworkPruneReport is a wrapper around the network.PruneReport type.
type NetworkPruneReport struct {
	network.PruneReport
}

// Swarm is a wrapper around the swarm type.
type Swarm struct {
	swarm.Swarm
}

// Node is a swarm node.
type Node struct {
	swarm.Node
	Raw []byte `json:"-"`
}

// Service is a swarm service.
type Service struct {
	swarm.Service
	Raw []byte `json:"-"`
}

// ServiceCreate is the result of creating a swarm service.
type ServiceCreate struct {
	ID       string
	Warnings []string
}

// ServiceUpdate is the result of updating a swarm service.
type ServiceUpdate struct {
	Warnings []string
}

// Task is a swarm task.
type Task struct {
	swarm.Task
	Raw []byte `json:"-"`
}

// Secret is a swarm secret.
type Secret struct {
	swarm.Secret
	Raw []byte `json:"-"`
}

// SwarmConfig is a swarm config object.
type SwarmConfig struct {
	swarm.Config
	Raw []byte `json:"-"`
}

// ImageAttestation is one in-toto statement attached to an image.
type ImageAttestation struct {
	image.AttestationStatement
}

// Info is daemon information.
type Info struct {
	system.Info
}

// Ping is the result of a daemon ping.
type Ping struct {
	APIVersion     string
	OSType         string
	Experimental   bool
	BuilderVersion string
}

// ServerVersion is the daemon version.
type ServerVersion struct {
	Platform      string
	Version       string
	APIVersion    string
	MinAPIVersion string
	Os            string
	Arch          string
	Experimental  bool
}

// DiskUsage is engine disk usage.
type DiskUsage struct {
	Containers ContainersDiskUsage
	Images     ImagesDiskUsage
	Volumes    VolumesDiskUsage
	BuildCache BuildCacheDiskUsage
}

// ContainersDiskUsage is container disk usage.
type ContainersDiskUsage struct {
	ActiveCount int64
	TotalCount  int64
	Reclaimable int64
	TotalSize   int64
	Items       []container.Summary
}

// ImagesDiskUsage is image disk usage.
type ImagesDiskUsage struct {
	ActiveCount int64
	TotalCount  int64
	Reclaimable int64
	TotalSize   int64
	Items       []image.Summary
}

// VolumesDiskUsage is volume disk usage.
type VolumesDiskUsage struct {
	ActiveCount int64
	TotalCount  int64
	Reclaimable int64
	TotalSize   int64
	Items       []volume.Volume
}

// BuildCacheDiskUsage is build-cache disk usage.
type BuildCacheDiskUsage struct {
	ActiveCount int64
	TotalCount  int64
	Reclaimable int64
	TotalSize   int64
	Items       []build.CacheRecord
}

// EngineEvent is one engine event.
type EngineEvent struct {
	events.Message
}

// DistributionInspect is a registry distribution inspect response.
type DistributionInspect struct {
	registry.DistributionInspect
}

// RegistryLogin is a registry login response.
type RegistryLogin struct {
	registry.AuthResponse
}

// Plugin is an engine plugin. Raw is the daemon JSON from inspect.
type Plugin struct {
	plugin.Plugin
	Raw []byte `json:"-"`
}
