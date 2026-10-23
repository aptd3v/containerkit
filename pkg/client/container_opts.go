package client

import (
	"io"

	"github.com/aptd3v/containerkit/pkg/containerkit"
	"github.com/docker/go-units"
	"github.com/moby/moby/api/types/blkiodev"
	"github.com/moby/moby/api/types/container"
	sdk "github.com/moby/moby/client"
)

// List is options for ContainerList.
// Latest sets Limit to 1 when Limit is 0. Since and Before are sent as filters.
type List struct {
	Size    bool
	All     bool
	Latest  bool
	Since   string
	Before  string
	Limit   int
	Filters []Filter
}

func (o *List) apply() sdk.ContainerListOptions {
	op := sdk.ContainerListOptions{Filters: argsFromFilters(nil)}
	if o == nil {
		return op
	}
	op.Size = o.Size
	op.All = o.All
	op.Limit = o.Limit
	if o.Latest && op.Limit == 0 {
		op.Limit = 1
	}
	filters := append([]Filter(nil), o.Filters...)
	if o.Since != "" {
		filters = append(filters, Filter{Key: "since", Value: o.Since})
	}
	if o.Before != "" {
		filters = append(filters, Filter{Key: "before", Value: o.Before})
	}
	op.Filters = argsFromFilters(filters)
	return op
}

// Start is options for ContainerStart.
type Start struct {
	CheckpointID  string
	CheckpointDir string
}

func (o *Start) apply() sdk.ContainerStartOptions {
	if o == nil {
		return sdk.ContainerStartOptions{}
	}
	return sdk.ContainerStartOptions{
		CheckpointID:  o.CheckpointID,
		CheckpointDir: o.CheckpointDir,
	}
}

// Stop is options for ContainerStop and ContainerRestart.
type Stop struct {
	Timeout *int
	Signal  string
}

func (o *Stop) apply() sdk.ContainerStopOptions {
	if o == nil {
		return sdk.ContainerStopOptions{}
	}
	return sdk.ContainerStopOptions{
		Timeout: o.Timeout,
		Signal:  o.Signal,
	}
}

// Remove is options for ContainerRemove.
type Remove struct {
	Force   bool
	Volumes bool
	Links   bool
}

func (o *Remove) apply() sdk.ContainerRemoveOptions {
	if o == nil {
		return sdk.ContainerRemoveOptions{}
	}
	return sdk.ContainerRemoveOptions{
		Force:         o.Force,
		RemoveVolumes: o.Volumes,
		RemoveLinks:   o.Links,
	}
}

// Logs is options for ContainerLogs.
type Logs struct {
	ShowStdout bool
	ShowStderr bool
	Since      string
	Until      string
	Timestamps bool
	Follow     bool
	Tail       string
	Details    bool
}

func (o *Logs) apply() sdk.ContainerLogsOptions {
	if o == nil {
		return sdk.ContainerLogsOptions{}
	}
	return sdk.ContainerLogsOptions{
		ShowStdout: o.ShowStdout,
		ShowStderr: o.ShowStderr,
		Since:      o.Since,
		Until:      o.Until,
		Timestamps: o.Timestamps,
		Follow:     o.Follow,
		Tail:       o.Tail,
		Details:    o.Details,
	}
}

// Exec is options for ContainerExecCreate.
type Exec struct {
	User          string
	Privileged    bool
	Tty           bool
	ConsoleWidth  uint
	ConsoleHeight uint
	AttachStdin   bool
	AttachStderr  bool
	AttachStdout  bool
	DetachKeys    string
	Env           []string
	WorkingDir    string
	Command       []string
}

func (o *Exec) apply() sdk.ExecCreateOptions {
	if o == nil {
		return sdk.ExecCreateOptions{}
	}
	return sdk.ExecCreateOptions{
		User:         o.User,
		Privileged:   o.Privileged,
		TTY:          o.Tty,
		ConsoleSize:  sdk.ConsoleSize{Width: o.ConsoleWidth, Height: o.ConsoleHeight},
		AttachStdin:  o.AttachStdin,
		AttachStderr: o.AttachStderr,
		AttachStdout: o.AttachStdout,
		DetachKeys:   o.DetachKeys,
		Env:          o.Env,
		WorkingDir:   o.WorkingDir,
		Cmd:          o.Command,
	}
}

// ExecStart is options for ContainerExecStart.
type ExecStart struct {
	Detach        bool
	Tty           bool
	ConsoleWidth  uint
	ConsoleHeight uint
}

func (o *ExecStart) apply() sdk.ExecStartOptions {
	if o == nil {
		return sdk.ExecStartOptions{}
	}
	return sdk.ExecStartOptions{
		Detach:      o.Detach,
		TTY:         o.Tty,
		ConsoleSize: sdk.ConsoleSize{Width: o.ConsoleWidth, Height: o.ConsoleHeight},
	}
}

// Resize is options for ContainerExecResize and ContainerResize.
type Resize struct {
	Width  uint
	Height uint
}

func (o *Resize) apply() sdk.ContainerResizeOptions {
	if o == nil {
		return sdk.ContainerResizeOptions{}
	}
	return sdk.ContainerResizeOptions{
		Width:  o.Width,
		Height: o.Height,
	}
}

// ExecAttach is options for ContainerExecAttach and ContainerExecAttachTerminal.
type ExecAttach struct {
	Detach        bool
	Tty           bool
	ConsoleWidth  uint
	ConsoleHeight uint
}

func (o *ExecAttach) apply() sdk.ExecAttachOptions {
	if o == nil {
		return sdk.ExecAttachOptions{}
	}
	return sdk.ExecAttachOptions{
		TTY:         o.Tty,
		ConsoleSize: sdk.ConsoleSize{Width: o.ConsoleWidth, Height: o.ConsoleHeight},
	}
}

// Attach is options for ContainerAttach.
type Attach struct {
	Stream     bool
	Stdin      bool
	Stdout     bool
	Stderr     bool
	DetachKeys string
	Logs       bool
}

func (o *Attach) apply() sdk.ContainerAttachOptions {
	if o == nil {
		return sdk.ContainerAttachOptions{}
	}
	return sdk.ContainerAttachOptions{
		Stream:     o.Stream,
		Stdin:      o.Stdin,
		Stdout:     o.Stdout,
		Stderr:     o.Stderr,
		DetachKeys: o.DetachKeys,
		Logs:       o.Logs,
	}
}

// Prune is options for ContainerPrune.
type Prune struct {
	Filters []Filter
}

func (o *Prune) apply() sdk.Filters {
	if o == nil {
		return argsFromFilters(nil)
	}
	return argsFromFilters(o.Filters)
}

// Commit is options for ContainerCommit.
type Commit struct {
	Reference string
	Comment   string
	Author    string
	Changes   []string
	Pause     bool
	Config    *containerkit.Container
}

func (o *Commit) apply() (sdk.ContainerCommitOptions, error) {
	if o == nil {
		return sdk.ContainerCommitOptions{}, nil
	}
	op := sdk.ContainerCommitOptions{
		Reference: o.Reference,
		Comment:   o.Comment,
		Author:    o.Author,
		Changes:   o.Changes,
		NoPause:   !o.Pause,
	}
	if o.Config != nil && o.Config.Config != nil {
		cfg, err := convert[container.Config](o.Config.Config.Container)
		if err != nil {
			return sdk.ContainerCommitOptions{}, err
		}
		op.Config = cfg
	}
	return op, nil
}

// Update is options for ContainerUpdate. Resource fields match Docker's update config.
type Update struct {
	RestartPolicy        containerkit.RestartPolicy
	RestartMaxRetry      int
	CPUShares            int64
	Memory               int64
	NanoCPUs             int64
	CgroupParent         string
	BlkioWeight          uint16
	BlkioWeightDevice    []*blkiodev.WeightDevice
	BlkioDeviceReadBps   []*blkiodev.ThrottleDevice
	BlkioDeviceWriteBps  []*blkiodev.ThrottleDevice
	BlkioDeviceReadIOps  []*blkiodev.ThrottleDevice
	BlkioDeviceWriteIOps []*blkiodev.ThrottleDevice
	CPUPeriod            int64
	CPUQuota             int64
	CPURealtimePeriod    int64
	CPURealtimeRuntime   int64
	CpusetCpus           string
	CpusetMems           string
	Devices              []container.DeviceMapping
	DeviceCgroupRules    []string
	DeviceRequests       []container.DeviceRequest
	MemoryReservation    int64
	MemorySwap           int64
	MemorySwappiness     *int64
	OomKillDisable       *bool
	PidsLimit            *int64
	Ulimits              []*units.Ulimit
	CPUCount             int64
	CPUPercent           int64
	IOMaximumIOps        uint64
	IOMaximumBandwidth   uint64
}

func (o *Update) apply() sdk.ContainerUpdateOptions {
	if o == nil {
		return sdk.ContainerUpdateOptions{}
	}
	op := sdk.ContainerUpdateOptions{
		Resources: &container.Resources{
			CPUShares:            o.CPUShares,
			Memory:               o.Memory,
			NanoCPUs:             o.NanoCPUs,
			CgroupParent:         o.CgroupParent,
			BlkioWeight:          o.BlkioWeight,
			BlkioWeightDevice:    o.BlkioWeightDevice,
			BlkioDeviceReadBps:   o.BlkioDeviceReadBps,
			BlkioDeviceWriteBps:  o.BlkioDeviceWriteBps,
			BlkioDeviceReadIOps:  o.BlkioDeviceReadIOps,
			BlkioDeviceWriteIOps: o.BlkioDeviceWriteIOps,
			CPUPeriod:            o.CPUPeriod,
			CPUQuota:             o.CPUQuota,
			CPURealtimePeriod:    o.CPURealtimePeriod,
			CPURealtimeRuntime:   o.CPURealtimeRuntime,
			CpusetCpus:           o.CpusetCpus,
			CpusetMems:           o.CpusetMems,
			Devices:              o.Devices,
			DeviceCgroupRules:    o.DeviceCgroupRules,
			DeviceRequests:       o.DeviceRequests,
			MemoryReservation:    o.MemoryReservation,
			MemorySwap:           o.MemorySwap,
			MemorySwappiness:     o.MemorySwappiness,
			OomKillDisable:       o.OomKillDisable,
			PidsLimit:            o.PidsLimit,
			Ulimits:              o.Ulimits,
			CPUCount:             o.CPUCount,
			CPUPercent:           o.CPUPercent,
			IOMaximumIOps:        o.IOMaximumIOps,
			IOMaximumBandwidth:   o.IOMaximumBandwidth,
		},
	}
	if o.RestartPolicy.String() != "" {
		op.RestartPolicy = &container.RestartPolicy{
			Name:              container.RestartPolicyMode(o.RestartPolicy.String()),
			MaximumRetryCount: o.RestartMaxRetry,
		}
	}
	return op
}

// CopyTo is options for ContainerCopyToContainer.
type CopyTo struct {
	AllowOverwriteDirWithFile bool
	CopyUIDGID                bool
	Content                   io.Reader
}

func (o *CopyTo) apply() sdk.CopyToContainerOptions {
	if o == nil {
		return sdk.CopyToContainerOptions{}
	}
	return sdk.CopyToContainerOptions{
		AllowOverwriteDirWithFile: o.AllowOverwriteDirWithFile,
		CopyUIDGID:                o.CopyUIDGID,
		Content:                   o.Content,
	}
}

// CheckpointCreate is options for ContainerCheckpointCreate.
type CheckpointCreate struct {
	CheckpointID  string
	CheckpointDir string
	Exit          bool
}

func (o *CheckpointCreate) apply() sdk.CheckpointCreateOptions {
	if o == nil {
		return sdk.CheckpointCreateOptions{}
	}
	return sdk.CheckpointCreateOptions{
		CheckpointID:  o.CheckpointID,
		CheckpointDir: o.CheckpointDir,
		Exit:          o.Exit,
	}
}

// CheckpointList is options for ContainerCheckpointList.
type CheckpointList struct {
	CheckpointDir string
}

func (o *CheckpointList) apply() sdk.CheckpointListOptions {
	if o == nil {
		return sdk.CheckpointListOptions{}
	}
	return sdk.CheckpointListOptions{CheckpointDir: o.CheckpointDir}
}

// CheckpointDelete is options for ContainerCheckpointDelete.
type CheckpointDelete struct {
	CheckpointID  string
	CheckpointDir string
}

func (o *CheckpointDelete) apply() sdk.CheckpointRemoveOptions {
	if o == nil {
		return sdk.CheckpointRemoveOptions{}
	}
	return sdk.CheckpointRemoveOptions{
		CheckpointID:  o.CheckpointID,
		CheckpointDir: o.CheckpointDir,
	}
}
