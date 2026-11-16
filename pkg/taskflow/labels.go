package taskflow

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
)

const (
	// LabelFlow is the Docker label for the flow name.
	LabelFlow = "containerkit.flow"
	// LabelRun is the Docker label for one Flow.Run.
	LabelRun = "containerkit.run"
	// LabelTask is the Docker label for the step name.
	LabelTask = "containerkit.task"
	// LabelDepends lists predecessor container names, sorted and comma-separated.
	LabelDepends = "containerkit.depends"
	// LabelPrecedes lists successor container names, sorted and comma-separated.
	LabelPrecedes = "containerkit.precedes"
)

// List returns containers stamped with LabelFlow equal to flowName, including
// stopped ones.
func List(ctx context.Context, cli *kit.Client, flowName string) (kit.ContainerListResult, error) {
	if cli == nil || cli.Engine() == nil {
		return kit.ContainerListResult{}, ErrNoClient
	}
	return listLabeled(ctx, cli, flowName, "")
}

func listLabeled(ctx context.Context, cli *kit.Client, flowName, runID string) (kit.ContainerListResult, error) {
	if cli == nil || cli.Engine() == nil {
		return kit.ContainerListResult{}, ErrNoClient
	}
	opts := []kit.EngineContainerListOption{kit.All(true), kit.Filter("label", LabelFlow+"="+flowName)}
	if runID != "" {
		opts = append(opts, kit.Filter("label", LabelRun+"="+runID))
	}
	return kit.ListContainers(ctx, cli.Engine(), opts...)
}

func (f *Flow) stamp() {
	if f == nil {
		return
	}
	for _, t := range f.taskList() {
		_ = t.stampContainer()
		t.stampCompose()
	}
}

func (t *Task) stampContainer() error {
	if t == nil || t.kind != kindContainer || t.container == nil || t.alreadyCreated() {
		return nil
	}
	var opts []kit.ContainerOption
	if t.container.Name() == "" && t.name != "" {
		opts = append(opts, kit.Name(t.implicitName()))
	}
	opts = append(opts, kit.Label(LabelFlow, t.flow.name), kit.Label(LabelTask, t.name))
	if t.flow.runID != "" {
		opts = append(opts, kit.Label(LabelRun, t.flow.runID))
	}
	if deps := strings.Join(t.nearestNames(false), ","); deps != "" {
		opts = append(opts, kit.Label(LabelDepends, deps))
	}
	if succs := strings.Join(t.nearestNames(true), ","); succs != "" {
		opts = append(opts, kit.Label(LabelPrecedes, succs))
	}
	return t.container.Apply(opts...)
}

func (t *Task) stampCompose() {
	if t == nil || t.kind != kindCompose || t.composeApp == nil {
		return
	}
	for _, svc := range t.composeApp.Services() {
		opts := []kit.ServiceOption{kit.Label(LabelFlow, t.flow.name), kit.Label(LabelTask, t.name)}
		if t.flow.runID != "" {
			opts = append(opts, kit.Label(LabelRun, t.flow.runID))
		}
		_ = svc.Apply(opts...)
	}
}

func (t *Task) alreadyCreated() bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.id != ""
}

func (t *Task) implicitName() string {
	if t == nil || t.flow == nil || t.flow.runID == "" || t.name == "" {
		return ""
	}
	return t.flow.name + "-" + t.name + "-" + t.flow.runID
}

func (t *Task) containerName() string {
	if t == nil || t.kind != kindContainer {
		return ""
	}
	if t.container != nil && t.container.Name() != "" {
		return t.container.Name()
	}
	if name := t.implicitName(); name != "" {
		return name
	}
	return t.name
}

func (t *Task) composeContainerNames() []string {
	if t == nil || t.composeApp == nil {
		return nil
	}
	var names []string
	for _, svc := range t.composeApp.Services() {
		if name := svc.ContainerName(); name != "" {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

func (t *Task) nearestNames(successors bool) []string {
	seen := map[*Task]struct{}{t: {}}
	var names []string
	var walk func(*Task)
	walk = func(n *Task) {
		neighbors := n.preds
		if successors {
			neighbors = n.succs
		}
		for _, nb := range neighbors {
			if nb == nil {
				continue
			}
			if _, ok := seen[nb]; ok {
				continue
			}
			seen[nb] = struct{}{}
			if found := nb.containerName(); found != "" {
				names = append(names, found)
				continue
			}
			if found := nb.composeContainerNames(); len(found) > 0 {
				names = append(names, found...)
				continue
			}
			walk(nb)
		}
	}
	walk(t)
	slices.Sort(names)
	return slices.Compact(names)
}

// List returns containers stamped with this run, including stopped ones.
// Before Run it is empty and does not query the engine.
func (f *Flow) List(ctx context.Context) (kit.ContainerListResult, error) {
	if f == nil {
		return kit.ContainerListResult{}, errors.New("flow is nil")
	}
	f.mu.Lock()
	runID := f.runID
	f.mu.Unlock()
	if runID == "" {
		return kit.ContainerListResult{}, nil
	}
	return listLabeled(ctx, f.cli, f.name, runID)
}

// Remove downs every Compose project the graph referenced, then force-removes
// Container-step IDs and leftover containers stamped with this run.
// Before Run the leftover sweep is empty and does not query the engine.
func (f *Flow) Remove(ctx context.Context) error {
	if f == nil {
		return errors.New("flow is nil")
	}
	var errs []error
	apps := f.composeList()
	f.mu.Lock()
	ids := append([]string(nil), f.created...)
	runID := f.runID
	f.mu.Unlock()
	api := f.engine()
	if api == nil {
		if len(ids) > 0 || len(apps) > 0 {
			errs = append(errs, ErrNoClient)
		}
		return errors.Join(errs...)
	}
	for _, app := range apps {
		if app == nil {
			continue
		}
		if err := app.Down(ctx, f.composer(), kit.Volumes(true)); err != nil {
			errs = append(errs, err)
		}
	}
	for _, id := range ids {
		if err := removeContainer(ctx, api, id); err != nil {
			errs = append(errs, err)
		}
	}
	if runID == "" {
		return errors.Join(errs...)
	}
	leftover, err := listLabeled(ctx, f.cli, f.name, runID)
	if err != nil {
		errs = append(errs, err)
		return errors.Join(errs...)
	}
	for _, ctr := range leftover.Items {
		if err := removeContainer(ctx, api, ctr.ID); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func removeContainer(ctx context.Context, api kit.Engine, id string) error {
	if id == "" {
		return nil
	}
	err := kit.NewContainer().Remove(ctx, api, id, kit.Force(true))
	if err != nil && !kit.IsNotFound(err) {
		return err
	}
	return nil
}
