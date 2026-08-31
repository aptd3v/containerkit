package client

import (
	"io"

	"github.com/aptd3v/containerkit/pkg/containerkit"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/checkpoint"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-units"
)

// List is options for ContainerList.
type List struct {
	Size    bool
	All     bool
	Latest  bool
	Since   string
	Before  string
	Limit   int
	Filters []Filter
}

func (o *List) apply() container.ListOptions {
	op := container.ListOptions{Filters: argsFromFilters(nil)}
	if o == nil {
		return op
	}
	op.Size = o.Size
	op.All = o.All
	op.Latest = o.Latest
	op.Since = o.Since
	op.Before = o.Before
	op.Limit = o.Limit
	op.Filters = argsFromFilters(o.Filters)
	return op
}

// Start is options for ContainerStart.
type Start struct {
	CheckpointID  string
	CheckpointDir string
}

func (o *Start) apply() container.StartOptions {
	if o == nil {
		return container.StartOptions{}
	}
	return container.StartOptions{
		CheckpointID:  o.CheckpointID,
		CheckpointDir: o.CheckpointDir,
	}
}

// Stop is options for ContainerStop and ContainerRestart.
type Stop struct {
	Timeout *int
	Signal  string
}

func (o *Stop) apply() container.StopOptions {
	if o == nil {
		return container.StopOptions{}
	}
	return container.StopOptions{
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

func (o *Remove) apply() container.RemoveOptions {
	if o == nil {
		return container.RemoveOptions{}
	}
	return container.RemoveOptions{
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

func (o *Logs) apply() container.LogsOptions {
	if o == nil {
		return container.LogsOptions{}
	}
	return container.LogsOptions{
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
	Detach        bool
	DetachKeys    string
	Env           []string
	WorkingDir    string
	Command       []string
}

func (o *Exec) apply() container.ExecOptions {
	if o == nil {
		return container.ExecOptions{}
	}
	op := container.ExecOptions{
		User:         o.User,
		Privileged:   o.Privileged,
		Tty:          o.Tty,
		AttachStdin:  o.AttachStdin,
		AttachStderr: o.AttachStderr,
		AttachStdout: o.AttachStdout,
		Detach:       o.Detach,
		DetachKeys:   o.DetachKeys,
		Env:          o.Env,
		WorkingDir:   o.WorkingDir,
		Cmd:          o.Command,
	}
	if o.ConsoleWidth != 0 || o.ConsoleHeight != 0 {
		op.ConsoleSize = &[2]uint{o.ConsoleWidth, o.ConsoleHeight}
	}
	return op
}

// ExecStart is options for ContainerExecStart.
type ExecStart struct {
	Detach        bool
	Tty           bool
	ConsoleWidth  uint
	ConsoleHeight uint
}

func (o *ExecStart) apply() container.ExecStartOptions {
	if o == nil {
		return container.ExecStartOptions{}
	}
	op := container.ExecStartOptions{
		Detach: o.Detach,
		Tty:    o.Tty,
	}
	if o.ConsoleWidth != 0 || o.ConsoleHeight != 0 {
		op.ConsoleSize = &[2]uint{o.ConsoleWidth, o.ConsoleHeight}
	}
	return op
}

// Resize is options for ContainerExecResize and ContainerResize.
type Resize struct {
	Width  uint
	Height uint
}

func (o *Resize) apply() container.ResizeOptions {
	if o == nil {
		return container.ResizeOptions{}
	}
	return container.ResizeOptions{
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

func (o *ExecAttach) apply() container.ExecAttachOptions {
	if o == nil {
		return container.ExecAttachOptions{}
	}
	op := container.ExecAttachOptions{
		Detach: o.Detach,
		Tty:    o.Tty,
	}
	if o.ConsoleWidth != 0 || o.ConsoleHeight != 0 {
		op.ConsoleSize = &[2]uint{o.ConsoleWidth, o.ConsoleHeight}
	}
	return op
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

func (o *Attach) apply() container.AttachOptions {
	if o == nil {
		return container.AttachOptions{}
	}
	return container.AttachOptions{
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

func (o *Prune) apply() filters.Args {
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

func (o *Commit) apply() container.CommitOptions {
	if o == nil {
		return container.CommitOptions{}
	}
	op := container.CommitOptions{
		Reference: o.Reference,
		Comment:   o.Comment,
		Author:    o.Author,
		Changes:   o.Changes,
		Pause:     o.Pause,
	}
	if o.Config != nil && o.Config.Config != nil {
		op.Config = o.Config.Config.Container
	}
	return op
}

// Update is options for ContainerUpdate. Resource fields match Docker's UpdateConfig.
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

func (o *Update) apply() container.UpdateConfig {
	if o == nil {
		return container.UpdateConfig{}
	}
	op := container.UpdateConfig{
		Resources: container.Resources{
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
	if o.RestartPolicy != "" {
		op.RestartPolicy = container.RestartPolicy{
			Name:              container.RestartPolicyMode(o.RestartPolicy),
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

func (o *CopyTo) apply() container.CopyToContainerOptions {
	if o == nil {
		return container.CopyToContainerOptions{}
	}
	return container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: o.AllowOverwriteDirWithFile,
		CopyUIDGID:                o.CopyUIDGID,
	}
}

// CheckpointCreate is options for ContainerCheckpointCreate.
type CheckpointCreate struct {
	CheckpointID  string
	CheckpointDir string
	Exit          bool
}

func (o *CheckpointCreate) apply() checkpoint.CreateOptions {
	if o == nil {
		return checkpoint.CreateOptions{}
	}
	return checkpoint.CreateOptions{
		CheckpointID:  o.CheckpointID,
		CheckpointDir: o.CheckpointDir,
		Exit:          o.Exit,
	}
}

// CheckpointList is options for ContainerCheckpointList.
type CheckpointList struct {
	CheckpointDir string
}

func (o *CheckpointList) apply() checkpoint.ListOptions {
	if o == nil {
		return checkpoint.ListOptions{}
	}
	return checkpoint.ListOptions{CheckpointDir: o.CheckpointDir}
}

// CheckpointDelete is options for ContainerCheckpointDelete.
type CheckpointDelete struct {
	CheckpointID  string
	CheckpointDir string
}

func (o *CheckpointDelete) apply() checkpoint.DeleteOptions {
	if o == nil {
		return checkpoint.DeleteOptions{}
	}
	return checkpoint.DeleteOptions{
		CheckpointID:  o.CheckpointID,
		CheckpointDir: o.CheckpointDir,
	}
}
