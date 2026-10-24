package taskflow

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/client/response"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

// Container adds a one-shot container step: remove by name, create, start,
// wait for exit. A non-zero exit fails the step. Restart policies
// (always, unless-stopped, on-failure) and AutoRemove are rejected: this
// step records the first exit, then copies logs; use Compose with *kit.Up
// for a long-running service and Retry for a flaky job. The step writes to ctr:
// it sets Name when empty and adds the containerkit.* labels.
func (f *Flow) Container(name string, ctr *kit.Container, opt *ContainerRun) *Task {
	if f == nil {
		return nil
	}
	return f.scope().container(name, ctr, opt)
}

// Container adds a one-shot container step to the subflow. See Flow.Container.
func (s *Subflow) Container(name string, ctr *kit.Container, opt *ContainerRun) *Task {
	sc, ok := s.scope()
	if !ok {
		return nil
	}
	return sc.container(name, ctr, opt)
}

// Build adds an image build step from a kit.Build.
func (f *Flow) Build(name string, b kit.Build, opt *client.ImageBuild) *Task {
	if f == nil {
		return nil
	}
	return f.scope().build(name, b, opt)
}

// Build adds an image build step to the subflow.
func (s *Subflow) Build(name string, b kit.Build, opt *client.ImageBuild) *Task {
	sc, ok := s.scope()
	if !ok {
		return nil
	}
	return sc.build(name, b, opt)
}

// Compose adds a step that runs a compose command. cmd is a *kit option
// (*kit.Up, *kit.Down, *kit.Logs, …) or a func(context.Context, *kit.Compose) error
// for Events, LoadProject, or options that depend on earlier steps. The
// option is copied so the caller's value is not mutated. A nil Writer on
// the option means Run.Writer. Prefer defer flow.Remove for end-of-run
// teardown so List can still see the stack; use a *kit.Down step when the
// graph itself must bring the project down (and Always if that must happen
// after a failure).
func (f *Flow) Compose(name string, app *kit.Compose, cmd any) *Task {
	if f == nil {
		return nil
	}
	return f.scope().compose(name, app, cmd)
}

// Compose adds a Compose step to the subflow. See Flow.Compose.
func (s *Subflow) Compose(name string, app *kit.Compose, cmd any) *Task {
	sc, ok := s.scope()
	if !ok {
		return nil
	}
	return sc.compose(name, app, cmd)
}

func (sc scope) container(name string, ctr *kit.Container, opt *ContainerRun) *Task {
	if ctr != nil && ctr.Name == "" {
		ctr.Name = name
	}
	t := sc.newTask(name, kindContainer)
	t.container = ctr
	t.containerRun = opt
	t.gtf = sc.adder.NewTask(name, func() {
		t.observe(t.runContainer)
	})
	return t
}

func (sc scope) build(name string, b kit.Build, opt *client.ImageBuild) *Task {
	t := sc.newTask(name, kindBuild)
	t.build = b
	t.buildOpt = opt
	t.gtf = sc.adder.NewTask(name, func() {
		t.observe(t.runBuild)
	})
	return t
}

func (sc scope) compose(name string, app *kit.Compose, cmd any) *Task {
	sc.flow.trackCompose(app)
	t := sc.newTask(name, kindCompose)
	t.composeApp = app
	fn, err := composeCommand(sc.flow, name, cmd)
	if err != nil {
		fn = func(context.Context, *kit.Compose) error { return err }
	}
	t.composeFn = fn
	t.gtf = sc.adder.NewTask(name, func() {
		t.observe(func() error {
			if t.composeFn == nil {
				return errors.New("compose func is nil")
			}
			return t.composeFn(t.flow.context(), t.composeApp)
		})
	})
	return t
}

func checkContainerJob(ctr *kit.Container) error {
	if ctr == nil || ctr.Config == nil || ctr.Config.Host == nil {
		return nil
	}
	host := ctr.Config.Host
	if host.AutoRemove {
		return errors.New("autoremove: Container waits for exit and then copies logs; use flow.Remove")
	}
	switch host.RestartPolicy.Name {
	case "", types.RestartPolicyNo:
		return nil
	case types.RestartPolicyAlways, types.RestartPolicyUnlessStopped:
		return fmt.Errorf("restart policy %s: Container waits for exit; use Compose with *kit.Up", host.RestartPolicy.Name)
	case types.RestartPolicyOnFailure:
		return errors.New("restart policy on-failure: Container already records the first exit; use Retry")
	default:
		return fmt.Errorf("restart policy %s: Container waits for exit", host.RestartPolicy.Name)
	}
}

func (t *Task) runContainer() error {
	if t.container == nil {
		return errors.New("container is nil")
	}
	if err := checkContainerJob(t.container); err != nil {
		return err
	}
	f := t.flow
	if f.cli == nil {
		return ErrNoClient
	}
	t.stamp()
	ctx := f.context()
	name := t.container.Name
	if err := f.cli.ContainerRemove(ctx, name, &client.Remove{Force: true}); err != nil && !client.IsErrNotFound(err) {
		return err
	}
	created, err := f.cli.ContainerCreate(ctx, t.container)
	if err != nil {
		return err
	}
	t.setID(created.ID)
	f.trackID(created.ID)
	if err := f.cli.ContainerStart(ctx, created.ID, nil); err != nil {
		return err
	}
	waitCh, errCh := f.cli.ContainerWait(ctx, created.ID, client.WaitConditionNotRunning)
	code, err := waitStatus(ctx, waitCh, errCh)
	t.setExitCode(int(code))
	if logErr := copyLogs(ctx, f.cli, created.ID, t.logDest()); logErr != nil && err == nil {
		err = logErr
	}
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("exit code %d", code)
	}
	return nil
}

func (t *Task) runBuild() error {
	f := t.flow
	if f.cli == nil {
		return ErrNoClient
	}
	ctx := f.context()
	r, err := t.build.Reader()
	if err != nil {
		return err
	}
	resp, err := f.cli.ImageBuild(ctx, r, t.buildOpt)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(stepWriter(f, t.name, nil), resp.Body)
	return err
}

// stepWriter is where a step's output goes: w, else Run.Writer, with a tee
// into Run.Log when set. The tee hides ReadFrom/WriterTo so io.Copy cannot
// bypass it.
func stepWriter(f *Flow, step string, w io.Writer) io.Writer {
	if w == nil {
		w = f.out()
	}
	if !f.hasLog() {
		return w
	}
	return writerOnly{io.MultiWriter(w, logWriter{f: f, step: step})}
}

type writerOnly struct{ io.Writer }

func (t *Task) logDest() io.Writer {
	var w io.Writer
	if t.containerRun != nil {
		w = t.containerRun.Logs
	}
	return stepWriter(t.flow, t.name, w)
}

func waitStatus(ctx context.Context, waitCh <-chan response.ContainerWait, errCh <-chan error) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case err := <-errCh:
		if err != nil {
			return 0, err
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case w, ok := <-waitCh:
			if !ok {
				return 0, ctx.Err()
			}
			return waitCode(w)
		}
	case w, ok := <-waitCh:
		if !ok {
			return 0, ctx.Err()
		}
		return waitCode(w)
	}
}

func waitCode(w response.ContainerWait) (int64, error) {
	if w.Error != nil && w.Error.Message != "" {
		return w.StatusCode, fmt.Errorf("%s", w.Error.Message)
	}
	return w.StatusCode, nil
}

func copyLogs(ctx context.Context, cli *client.Client, id string, w io.Writer) error {
	r, err := cli.ContainerLogs(ctx, id, &client.Logs{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = client.NewLogCopier(w, w).Copy(r)
	return err
}
