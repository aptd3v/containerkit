package taskflow

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
	composetypes "github.com/compose-spec/compose-go/v2/types"
	cerrdefs "github.com/containerd/errdefs"
	composeapi "github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/api/types/container"
	volumetypes "github.com/moby/moby/api/types/volume"
	clienttypes "github.com/moby/moby/client"
)

type scriptedEngine struct {
	engineStub
	createErr        error
	startErr         error
	removeErr        error
	logsErr          error
	logs             clienttypes.ContainerLogsResult
	wait             clienttypes.ContainerWaitResult
	list             clienttypes.ContainerListResult
	listErr          error
	buildBody        io.ReadCloser
	buildErr         error
	builds           int
	built            clienttypes.ImageBuildOptions
	inspects         []string
	images           map[string]error
	tags             int
	tagErr           error
	created          clienttypes.ContainerCreateOptions
	id               string
	seq              int
	removed          []string
	lists            int
	live             []scriptedContainer
	events           []string
	volumes          map[string]map[string]string
	volCreates       []clienttypes.VolumeCreateOptions
	volInspects      []string
	volRemoves       []string
	volumeCreateErr  error
	volumeInspectErr error
	volumeRemoveErr  error
}

type scriptedContainer struct {
	id     string
	labels map[string]string
}

func (s *scriptedEngine) ContainerCreate(_ context.Context, options clienttypes.ContainerCreateOptions) (clienttypes.ContainerCreateResult, error) {
	s.created = options
	if s.createErr != nil {
		return clienttypes.ContainerCreateResult{}, s.createErr
	}
	id := s.id
	if id == "" {
		s.seq++
		id = fmt.Sprintf("ctr-%d", s.seq)
	}
	labels := map[string]string{}
	if options.Config != nil {
		for k, v := range options.Config.Labels {
			labels[k] = v
		}
	}
	s.live = append(s.live, scriptedContainer{id: id, labels: labels})
	return clienttypes.ContainerCreateResult{ID: id}, nil
}

func (s *scriptedEngine) ContainerStart(context.Context, string, clienttypes.ContainerStartOptions) (clienttypes.ContainerStartResult, error) {
	return clienttypes.ContainerStartResult{}, s.startErr
}

func (s *scriptedEngine) ContainerRemove(_ context.Context, id string, _ clienttypes.ContainerRemoveOptions) (clienttypes.ContainerRemoveResult, error) {
	s.removed = append(s.removed, id)
	s.events = append(s.events, "container:"+id)
	if s.removeErr != nil {
		return clienttypes.ContainerRemoveResult{}, s.removeErr
	}
	kept := s.live[:0]
	for _, c := range s.live {
		if c.id != id {
			kept = append(kept, c)
		}
	}
	s.live = kept
	return clienttypes.ContainerRemoveResult{}, nil
}

func (s *scriptedEngine) rememberVolume(name string, labels map[string]string) {
	if s.volumes == nil {
		s.volumes = map[string]map[string]string{}
	}
	s.volumes[name] = labels
}

func (s *scriptedEngine) VolumeCreate(_ context.Context, options clienttypes.VolumeCreateOptions) (clienttypes.VolumeCreateResult, error) {
	s.volCreates = append(s.volCreates, options)
	if s.volumeCreateErr != nil {
		return clienttypes.VolumeCreateResult{}, s.volumeCreateErr
	}
	copied := map[string]string{}
	for k, v := range options.Labels {
		copied[k] = v
	}
	s.rememberVolume(options.Name, copied)
	return clienttypes.VolumeCreateResult{}, nil
}

func (s *scriptedEngine) VolumeInspect(_ context.Context, id string, _ clienttypes.VolumeInspectOptions) (clienttypes.VolumeInspectResult, error) {
	s.volInspects = append(s.volInspects, id)
	if s.volumeInspectErr != nil {
		return clienttypes.VolumeInspectResult{}, s.volumeInspectErr
	}
	labels, ok := s.volumes[id]
	if !ok {
		return clienttypes.VolumeInspectResult{}, cerrdefs.ErrNotFound
	}
	return clienttypes.VolumeInspectResult{Volume: volumetypes.Volume{Name: id, Labels: labels}}, nil
}

func (s *scriptedEngine) VolumeRemove(_ context.Context, id string, _ clienttypes.VolumeRemoveOptions) (clienttypes.VolumeRemoveResult, error) {
	s.volRemoves = append(s.volRemoves, id)
	s.events = append(s.events, "volume:"+id)
	if s.volumeRemoveErr != nil {
		return clienttypes.VolumeRemoveResult{}, s.volumeRemoveErr
	}
	delete(s.volumes, id)
	return clienttypes.VolumeRemoveResult{}, nil
}

func (s *scriptedEngine) ContainerLogs(context.Context, string, clienttypes.ContainerLogsOptions) (clienttypes.ContainerLogsResult, error) {
	return s.logs, s.logsErr
}

func (s *scriptedEngine) ContainerWait(context.Context, string, clienttypes.ContainerWaitOptions) clienttypes.ContainerWaitResult {
	if s.wait.Result == nil && s.wait.Error == nil {
		ch := make(chan container.WaitResponse, 1)
		ch <- container.WaitResponse{}
		return clienttypes.ContainerWaitResult{Result: ch}
	}
	return s.wait
}

func (s *scriptedEngine) ContainerList(_ context.Context, opts clienttypes.ContainerListOptions) (clienttypes.ContainerListResult, error) {
	s.lists++
	if s.listErr != nil {
		return clienttypes.ContainerListResult{}, s.listErr
	}
	if s.list.Items != nil {
		return s.list, nil
	}
	var items []container.Summary
	for _, c := range s.live {
		if labelMatch(c.labels, opts.Filters) {
			items = append(items, container.Summary{ID: c.id, Labels: c.labels})
		}
	}
	return clienttypes.ContainerListResult{Items: items}, nil
}

func labelMatch(labels map[string]string, filters clienttypes.Filters) bool {
	for expr := range filters["label"] {
		key, val, ok := strings.Cut(expr, "=")
		if !ok || labels[key] != val {
			return false
		}
	}
	return true
}

func (s *scriptedEngine) ImageBuild(_ context.Context, _ io.Reader, options clienttypes.ImageBuildOptions) (clienttypes.ImageBuildResult, error) {
	s.builds++
	s.built = options
	if s.buildErr != nil {
		return clienttypes.ImageBuildResult{}, s.buildErr
	}
	return clienttypes.ImageBuildResult{Body: s.buildBody}, nil
}

func (s *scriptedEngine) ImageInspect(_ context.Context, image string, _ ...clienttypes.ImageInspectOption) (clienttypes.ImageInspectResult, error) {
	s.inspects = append(s.inspects, image)
	if s.images == nil {
		return clienttypes.ImageInspectResult{}, nil
	}
	err, ok := s.images[image]
	if !ok {
		return clienttypes.ImageInspectResult{}, cerrdefs.ErrNotFound
	}
	return clienttypes.ImageInspectResult{}, err
}

func (s *scriptedEngine) ImageTag(context.Context, clienttypes.ImageTagOptions) (clienttypes.ImageTagResult, error) {
	s.tags++
	return clienttypes.ImageTagResult{}, s.tagErr
}

func waitResult(code int64, msg string) clienttypes.ContainerWaitResult {
	ch := make(chan container.WaitResponse, 1)
	w := container.WaitResponse{StatusCode: code}
	if msg != "" {
		w.Error = &container.WaitExitError{Message: msg}
	}
	ch <- w
	return clienttypes.ContainerWaitResult{Result: ch}
}

func mux(payload string) io.ReadCloser {
	p := []byte(payload)
	hdr := make([]byte, 8)
	hdr[0] = 1
	binary.BigEndian.PutUint32(hdr[4:], uint32(len(p)))
	return io.NopCloser(bytes.NewReader(append(hdr, p...)))
}

func newClient(eng *scriptedEngine, comp kit.Composer) *kit.Client {
	return kit.NewClient(eng, comp)
}

type fakeComposer struct {
	op           string
	code         int
	err          error
	downErr      error
	downs        int
	ups          int
	upOpts       composeapi.UpOptions
	downOpts     []composeapi.DownOptions
	downCtx      context.Context
	downLive     bool
	downDeadline bool
}

func (f *fakeComposer) note(op string) error {
	f.op = op
	return f.err
}

func (f *fakeComposer) Build(context.Context, *composetypes.Project, composeapi.BuildOptions) error {
	return f.note("build")
}
func (f *fakeComposer) Push(context.Context, *composetypes.Project, composeapi.PushOptions) error {
	return f.note("push")
}
func (f *fakeComposer) Pull(context.Context, *composetypes.Project, composeapi.PullOptions) error {
	return f.note("pull")
}
func (f *fakeComposer) Create(context.Context, *composetypes.Project, composeapi.CreateOptions) error {
	return f.note("create")
}
func (f *fakeComposer) Start(context.Context, string, composeapi.StartOptions) error {
	return f.note("start")
}
func (f *fakeComposer) Restart(context.Context, string, composeapi.RestartOptions) error {
	return f.note("restart")
}
func (f *fakeComposer) Stop(context.Context, string, composeapi.StopOptions) error {
	return f.note("stop")
}
func (f *fakeComposer) Up(_ context.Context, _ *composetypes.Project, opts composeapi.UpOptions) error {
	f.ups++
	f.upOpts = opts
	return f.note("up")
}
func (f *fakeComposer) Down(ctx context.Context, _ string, opts composeapi.DownOptions) error {
	f.downs++
	f.downOpts = append(f.downOpts, opts)
	f.downCtx = ctx
	f.downLive = ctx.Err() == nil
	_, f.downDeadline = ctx.Deadline()
	if f.downErr != nil {
		return f.downErr
	}
	return f.note("down")
}
func (f *fakeComposer) Logs(context.Context, string, composeapi.LogConsumer, composeapi.LogOptions) error {
	return f.note("logs")
}
func (f *fakeComposer) Ps(context.Context, string, composeapi.PsOptions) ([]composeapi.ContainerSummary, error) {
	return nil, f.note("ps")
}
func (f *fakeComposer) List(context.Context, composeapi.ListOptions) ([]composeapi.Stack, error) {
	return nil, f.note("list")
}
func (f *fakeComposer) Kill(context.Context, string, composeapi.KillOptions) error {
	return f.note("kill")
}
func (f *fakeComposer) RunOneOffContainer(context.Context, *composetypes.Project, composeapi.RunOptions) (int, error) {
	f.op = "run"
	return f.code, f.err
}
func (f *fakeComposer) Remove(context.Context, string, composeapi.RemoveOptions) error {
	return f.note("remove")
}
func (f *fakeComposer) Exec(context.Context, string, composeapi.RunOptions) (int, error) {
	f.op = "exec"
	return f.code, f.err
}
func (f *fakeComposer) Attach(context.Context, string, composeapi.AttachOptions) error {
	return f.note("attach")
}
func (f *fakeComposer) Copy(context.Context, string, composeapi.CopyOptions) error {
	return f.note("copy")
}
func (f *fakeComposer) Pause(context.Context, string, composeapi.PauseOptions) error {
	return f.note("pause")
}
func (f *fakeComposer) UnPause(context.Context, string, composeapi.PauseOptions) error {
	return f.note("unpause")
}
func (f *fakeComposer) Top(context.Context, string, []string) ([]composeapi.ContainerProcSummary, error) {
	return nil, f.note("top")
}
func (f *fakeComposer) Events(context.Context, string, composeapi.EventsOptions) error {
	return f.note("events")
}
func (f *fakeComposer) Ports(context.Context, string, string, uint16, composeapi.PortOptions) (composeapi.PortPublishers, error) {
	return nil, f.note("port")
}
func (f *fakeComposer) Publish(context.Context, *composetypes.Project, string, composeapi.PublishOptions) error {
	return f.note("publish")
}
func (f *fakeComposer) Images(context.Context, string, composeapi.ImagesOptions) (map[string]composeapi.ImageSummary, error) {
	return nil, f.note("images")
}
func (f *fakeComposer) MaxConcurrency(_ int) { f.op = "max" }
func (f *fakeComposer) DryRunMode(ctx context.Context, _ bool) (context.Context, error) {
	f.op = "dry"
	return ctx, f.err
}
func (f *fakeComposer) Watch(context.Context, *composetypes.Project, composeapi.WatchOptions) error {
	return f.note("watch")
}
func (f *fakeComposer) Viz(context.Context, *composetypes.Project, composeapi.VizOptions) (string, error) {
	return "", f.note("viz")
}
func (f *fakeComposer) Wait(context.Context, string, composeapi.WaitOptions) (int64, error) {
	f.op = "wait"
	return 0, f.err
}
func (f *fakeComposer) Scale(context.Context, *composetypes.Project, composeapi.ScaleOptions) error {
	return f.note("scale")
}
func (f *fakeComposer) Export(context.Context, string, composeapi.ExportOptions) error {
	return f.note("export")
}
func (f *fakeComposer) Commit(context.Context, string, composeapi.CommitOptions) error {
	return f.note("commit")
}
func (f *fakeComposer) Generate(context.Context, composeapi.GenerateOptions) (*composetypes.Project, error) {
	f.op = "generate"
	return &composetypes.Project{}, f.err
}
func (f *fakeComposer) Volumes(context.Context, string, composeapi.VolumesOptions) ([]composeapi.VolumesSummary, error) {
	return nil, f.note("volumes")
}
func (f *fakeComposer) LoadProject(context.Context, composeapi.ProjectLoadOptions) (*composetypes.Project, error) {
	f.op = "load"
	return &composetypes.Project{Name: "app"}, f.err
}

func sampleApp() *kit.Compose {
	return kit.NewCompose("app").WithService(kit.NewService(kit.Name("db"), kit.Image("postgres:16")))
}

func discardRun() *Run {
	return &Run{Writer: io.Discard, Concurrency: 1}
}

var _ = cerrdefs.ErrNotFound
