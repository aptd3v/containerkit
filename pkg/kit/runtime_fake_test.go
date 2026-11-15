package kit

import (
	"context"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

type fakeComposer struct {
	up        *composetypes.Project
	upOpts    composeapi.UpOptions
	down      string
	downOpts  composeapi.DownOptions
	created   composeapi.CreateOptions
	built     composeapi.BuildOptions
	pulled    composeapi.PullOptions
	started   composeapi.StartOptions
	watched   composeapi.WatchOptions
	published composeapi.PublishOptions
	committed composeapi.CommitOptions
	ran       composeapi.RunOptions
	gen       composeapi.GenerateOptions
	loaded    composeapi.ProjectLoadOptions
}

func (f *fakeComposer) Build(_ context.Context, _ *composetypes.Project, options composeapi.BuildOptions) error {
	f.built = options
	return nil
}
func (f *fakeComposer) Push(context.Context, *composetypes.Project, composeapi.PushOptions) error {
	return nil
}
func (f *fakeComposer) Pull(_ context.Context, _ *composetypes.Project, options composeapi.PullOptions) error {
	f.pulled = options
	return nil
}
func (f *fakeComposer) Create(_ context.Context, _ *composetypes.Project, options composeapi.CreateOptions) error {
	f.created = options
	return nil
}
func (f *fakeComposer) Start(_ context.Context, _ string, options composeapi.StartOptions) error {
	f.started = options
	return nil
}
func (f *fakeComposer) Restart(context.Context, string, composeapi.RestartOptions) error {
	return nil
}
func (f *fakeComposer) Stop(context.Context, string, composeapi.StopOptions) error { return nil }
func (f *fakeComposer) Up(_ context.Context, project *composetypes.Project, options composeapi.UpOptions) error {
	f.up = project
	f.upOpts = options
	return nil
}
func (f *fakeComposer) Down(_ context.Context, name string, options composeapi.DownOptions) error {
	f.down = name
	f.downOpts = options
	return nil
}
func (f *fakeComposer) Logs(context.Context, string, composeapi.LogConsumer, composeapi.LogOptions) error {
	return nil
}
func (f *fakeComposer) Ps(context.Context, string, composeapi.PsOptions) ([]composeapi.ContainerSummary, error) {
	return nil, nil
}
func (f *fakeComposer) List(context.Context, composeapi.ListOptions) ([]composeapi.Stack, error) {
	return nil, nil
}
func (f *fakeComposer) Kill(context.Context, string, composeapi.KillOptions) error { return nil }
func (f *fakeComposer) RunOneOffContainer(_ context.Context, _ *composetypes.Project, options composeapi.RunOptions) (int, error) {
	f.ran = options
	return 0, nil
}
func (f *fakeComposer) Remove(context.Context, string, composeapi.RemoveOptions) error {
	return nil
}
func (f *fakeComposer) Exec(context.Context, string, composeapi.RunOptions) (int, error) {
	return 0, nil
}
func (f *fakeComposer) Attach(context.Context, string, composeapi.AttachOptions) error { return nil }
func (f *fakeComposer) Copy(context.Context, string, composeapi.CopyOptions) error     { return nil }
func (f *fakeComposer) Pause(context.Context, string, composeapi.PauseOptions) error   { return nil }
func (f *fakeComposer) UnPause(context.Context, string, composeapi.PauseOptions) error { return nil }
func (f *fakeComposer) Top(context.Context, string, []string) ([]composeapi.ContainerProcSummary, error) {
	return nil, nil
}
func (f *fakeComposer) Events(context.Context, string, composeapi.EventsOptions) error { return nil }
func (f *fakeComposer) Ports(context.Context, string, string, uint16, composeapi.PortOptions) (composeapi.PortPublishers, error) {
	return nil, nil
}
func (f *fakeComposer) Publish(_ context.Context, _ *composetypes.Project, _ string, options composeapi.PublishOptions) error {
	f.published = options
	return nil
}
func (f *fakeComposer) Images(context.Context, string, composeapi.ImagesOptions) (map[string]composeapi.ImageSummary, error) {
	return nil, nil
}
func (f *fakeComposer) MaxConcurrency(int) {}
func (f *fakeComposer) DryRunMode(ctx context.Context, _ bool) (context.Context, error) {
	return ctx, nil
}
func (f *fakeComposer) Watch(_ context.Context, _ *composetypes.Project, options composeapi.WatchOptions) error {
	f.watched = options
	return nil
}
func (f *fakeComposer) Viz(context.Context, *composetypes.Project, composeapi.VizOptions) (string, error) {
	return "", nil
}
func (f *fakeComposer) Wait(context.Context, string, composeapi.WaitOptions) (int64, error) {
	return 0, nil
}
func (f *fakeComposer) Scale(context.Context, *composetypes.Project, composeapi.ScaleOptions) error {
	return nil
}
func (f *fakeComposer) Export(context.Context, string, composeapi.ExportOptions) error { return nil }
func (f *fakeComposer) Commit(_ context.Context, _ string, options composeapi.CommitOptions) error {
	f.committed = options
	return nil
}
func (f *fakeComposer) Generate(_ context.Context, options composeapi.GenerateOptions) (*composetypes.Project, error) {
	f.gen = options
	return &composetypes.Project{Name: options.ProjectName}, nil
}
func (f *fakeComposer) Volumes(context.Context, string, composeapi.VolumesOptions) ([]composeapi.VolumesSummary, error) {
	return nil, nil
}
func (f *fakeComposer) LoadProject(_ context.Context, options composeapi.ProjectLoadOptions) (*composetypes.Project, error) {
	f.loaded = options
	return &composetypes.Project{Name: options.ProjectName}, nil
}

var _ Composer = (*fakeComposer)(nil)

type fakeLogConsumer struct{}

func (fakeLogConsumer) Log(string, string)    {}
func (fakeLogConsumer) Err(string, string)    {}
func (fakeLogConsumer) Status(string, string) {}
