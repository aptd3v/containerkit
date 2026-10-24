package client

import sdk "github.com/moby/moby/client"

// List is [sdk.ContainerListOptions].
// Latest sets Limit to 1 when Limit is 0. Since and Before are also sent as filters.
type List sdk.ContainerListOptions

func (o *List) apply() sdk.ContainerListOptions {
	if o == nil {
		return sdk.ContainerListOptions{}
	}
	op := sdk.ContainerListOptions(*o)
	if op.Latest && op.Limit == 0 {
		op.Limit = 1
	}
	if op.Since != "" {
		if op.Filters == nil {
			op.Filters = make(sdk.Filters)
		}
		op.Filters = op.Filters.Add("since", op.Since)
	}
	if op.Before != "" {
		if op.Filters == nil {
			op.Filters = make(sdk.Filters)
		}
		op.Filters = op.Filters.Add("before", op.Before)
	}
	return op
}

// Start is [sdk.ContainerStartOptions].
type Start sdk.ContainerStartOptions

func (o *Start) apply() sdk.ContainerStartOptions { return sdk.ContainerStartOptions(apply(o)) }

// Stop is [sdk.ContainerStopOptions].
type Stop sdk.ContainerStopOptions

func (o *Stop) apply() sdk.ContainerStopOptions { return sdk.ContainerStopOptions(apply(o)) }

// Remove is [sdk.ContainerRemoveOptions].
type Remove sdk.ContainerRemoveOptions

func (o *Remove) apply() sdk.ContainerRemoveOptions { return sdk.ContainerRemoveOptions(apply(o)) }

// Logs is [sdk.ContainerLogsOptions].
type Logs sdk.ContainerLogsOptions

func (o *Logs) apply() sdk.ContainerLogsOptions { return sdk.ContainerLogsOptions(apply(o)) }

// Exec is [sdk.ExecCreateOptions].
type Exec sdk.ExecCreateOptions

func (o *Exec) apply() sdk.ExecCreateOptions { return sdk.ExecCreateOptions(apply(o)) }

// ExecStart is [sdk.ExecStartOptions].
type ExecStart sdk.ExecStartOptions

func (o *ExecStart) apply() sdk.ExecStartOptions { return sdk.ExecStartOptions(apply(o)) }

// Resize is [sdk.ContainerResizeOptions].
type Resize sdk.ContainerResizeOptions

func (o *Resize) apply() sdk.ContainerResizeOptions { return sdk.ContainerResizeOptions(apply(o)) }

// ExecAttach is [sdk.ExecAttachOptions].
type ExecAttach sdk.ExecAttachOptions

func (o *ExecAttach) apply() sdk.ExecAttachOptions { return sdk.ExecAttachOptions(apply(o)) }

// Attach is [sdk.ContainerAttachOptions].
type Attach sdk.ContainerAttachOptions

func (o *Attach) apply() sdk.ContainerAttachOptions { return sdk.ContainerAttachOptions(apply(o)) }

// Prune is [sdk.ContainerPruneOptions].
type Prune sdk.ContainerPruneOptions

func (o *Prune) apply() sdk.ContainerPruneOptions { return sdk.ContainerPruneOptions(apply(o)) }

// Commit is [sdk.ContainerCommitOptions].
type Commit sdk.ContainerCommitOptions

func (o *Commit) apply() sdk.ContainerCommitOptions { return sdk.ContainerCommitOptions(apply(o)) }

// Update is [sdk.ContainerUpdateOptions].
type Update sdk.ContainerUpdateOptions

func (o *Update) apply() sdk.ContainerUpdateOptions { return sdk.ContainerUpdateOptions(apply(o)) }

// CopyTo is [sdk.CopyToContainerOptions].
// DestinationPath is filled from the method argument.
type CopyTo sdk.CopyToContainerOptions

func (o *CopyTo) apply() sdk.CopyToContainerOptions { return sdk.CopyToContainerOptions(apply(o)) }

// CheckpointCreate is [sdk.CheckpointCreateOptions].
type CheckpointCreate sdk.CheckpointCreateOptions

func (o *CheckpointCreate) apply() sdk.CheckpointCreateOptions {
	return sdk.CheckpointCreateOptions(apply(o))
}

// CheckpointList is [sdk.CheckpointListOptions].
type CheckpointList sdk.CheckpointListOptions

func (o *CheckpointList) apply() sdk.CheckpointListOptions {
	return sdk.CheckpointListOptions(apply(o))
}

// CheckpointDelete is [sdk.CheckpointRemoveOptions].
type CheckpointDelete sdk.CheckpointRemoveOptions

func (o *CheckpointDelete) apply() sdk.CheckpointRemoveOptions {
	return sdk.CheckpointRemoveOptions(apply(o))
}
