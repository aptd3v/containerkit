package taskflow

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/moby/moby/api/pkg/stdcopy"
	containertypes "github.com/moby/moby/api/types/container"
)

// Container adds a one-shot container step: remove by name, create, start,
// wait for exit. A non-zero exit fails the step. Restart policies
// (always, unless-stopped, on-failure) and AutoRemove are rejected.
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

// Build adds an image build step. Options are the same ones BuildImage takes.
func (f *Flow) Build(name string, opts ...kit.EngineBuildOption) *Task {
	if f == nil {
		return nil
	}
	return f.scope().build(name, opts)
}

// Build adds an image build step to the subflow.
func (s *Subflow) Build(name string, opts ...kit.EngineBuildOption) *Task {
	sc, ok := s.scope()
	if !ok {
		return nil
	}
	return sc.build(name, opts)
}

// Compose adds a step that runs a compose command. cmd is a [ComposeCmd]
// (Up, Down, Logs, …) or a func(context.Context, *kit.Compose) error.
// A func(context.Context, *kit.Compose, kit.Composer) error is accepted too.
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
	t := sc.newTask(name, kindContainer)
	t.container = ctr
	t.containerRun = opt
	t.gtf = sc.adder.NewTask(name, func() {
		t.observe(t.runContainer)
	})
	return t
}

func (sc scope) build(name string, opts []kit.EngineBuildOption) *Task {
	t := sc.newTask(name, kindBuild)
	t.build = append([]kit.EngineBuildOption(nil), opts...)
	t.gtf = sc.adder.NewTask(name, func() {
		t.observe(t.runBuild)
	})
	return t
}

func (sc scope) compose(name string, app *kit.Compose, cmd any) *Task {
	sc.flow.trackCompose(app)
	t := sc.newTask(name, kindCompose)
	t.composeApp = app
	fn, err := composeCommand(cmd)
	if err != nil {
		fn = func(context.Context, *kit.Compose, kit.Composer) error { return err }
	}
	t.composeFn = fn
	t.gtf = sc.adder.NewTask(name, func() {
		t.observe(func() error {
			if t.composeApp == nil {
				return errors.New("compose is nil")
			}
			if t.composeFn == nil {
				return errors.New("compose func is nil")
			}
			return t.composeFn(t.flow.context(), t.composeApp, t.flow.composer())
		})
	})
	return t
}

func rejectJobPolicy(auto bool, policy string) error {
	if auto {
		return errors.New("autoremove: Container waits for exit and then copies logs; use flow.Remove")
	}
	switch policy {
	case "", "no":
		return nil
	case "on-failure":
		return errors.New("restart policy on-failure: Container already records the first exit; use Retry")
	case "always", "unless-stopped":
		return fmt.Errorf("restart policy %s: Container waits for exit; use Compose with Up", policy)
	default:
		return fmt.Errorf("restart policy %s: Container waits for exit", policy)
	}
}

func (t *Task) runContainer() error {
	if t.container == nil {
		return errors.New("container is nil")
	}
	f := t.flow
	api := f.engine()
	if api == nil {
		return ErrNoClient
	}
	if err := t.stampContainer(); err != nil {
		return err
	}
	cfg, host, _, err := t.container.EngineConfig()
	if err != nil {
		return err
	}
	_ = cfg
	if err := rejectJobPolicy(host.AutoRemove, string(host.RestartPolicy.Name)); err != nil {
		return err
	}
	ctx := f.context()
	name := t.container.Name()
	if err := t.container.Remove(ctx, api, name, kit.Force(true)); err != nil && !kit.IsNotFound(err) {
		return err
	}
	created, err := t.container.Create(ctx, api)
	if err != nil {
		return err
	}
	t.setID(created.ID)
	f.trackID(created.ID)
	if err := t.container.Start(ctx, api, created.ID); err != nil {
		return err
	}
	code, err := waitStatus(ctx, t.container.Wait(ctx, api, created.ID, kit.WaitFor(kit.NotRunning)))
	t.setExitCode(int(code))
	if logErr := copyLogs(ctx, t.container, api, created.ID, t.logDest()); logErr != nil && err == nil {
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
	api := f.engine()
	if api == nil {
		return ErrNoClient
	}
	ctx := f.context()
	w := stepWriter(f, t.name, nil)
	logger := kit.NewLogger(kit.OutputStream(w), kit.Color(false), kit.Progress("plain"))
	opts := append([]kit.EngineBuildOption{}, t.build...)
	opts = append(opts, kit.Log(logger))
	_, err := kit.BuildImage(ctx, api, opts...)
	return err
}

// stepWriter is where a step's output goes: w, else Run.Writer, with a tee
// into Run.Log and the context logger. The tee hides ReadFrom so io.Copy
// cannot bypass it.
func stepWriter(f *Flow, step string, w io.Writer) io.Writer {
	if w == nil {
		w = f.out()
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

func waitStatus(ctx context.Context, res kit.ContainerWaitResult) (int64, error) {
	if res.Result == nil && res.Error == nil {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return 0, errors.New("wait: no result")
		}
	}
	if res.Error == nil {
		return readWait(ctx, res.Result)
	}
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case err, ok := <-res.Error:
		if !ok {
			return 0, errors.New("wait: error channel closed")
		}
		if err != nil {
			return 0, err
		}
		return readWait(ctx, res.Result)
	}
}

func readWait(ctx context.Context, ch <-chan containertypes.WaitResponse) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case w, ok := <-ch:
		if !ok {
			return 0, errors.New("wait: result channel closed")
		}
		return waitCode(w.StatusCode, exitMessage(w.Error))
	}
}

func exitMessage(err *containertypes.WaitExitError) string {
	if err == nil {
		return ""
	}
	return err.Message
}

func waitCode(code int64, msg string) (int64, error) {
	if msg != "" {
		return code, errors.New(msg)
	}
	return code, nil
}

func copyLogs(ctx context.Context, c *kit.Container, api kit.Engine, id string, w io.Writer) error {
	rc, err := c.Logs(ctx, api, id, kit.AttachStdout(true), kit.AttachStderr(true))
	if err != nil {
		return err
	}
	if rc == nil {
		return nil
	}
	defer rc.Close()
	_, err = stdcopy.StdCopy(w, w, rc)
	return err
}
