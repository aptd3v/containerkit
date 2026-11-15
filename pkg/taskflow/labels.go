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
	if ctx == nil {
		ctx = context.Background()
	}
	return kit.ListContainers(ctx, cli.Engine(), kit.All(true), kit.Filter("label", LabelFlow+"="+flowName))
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
		opts = append(opts, kit.Name(t.name))
	}
	opts = append(opts, kit.Label(LabelFlow, t.flow.name), kit.Label(LabelTask, t.name))
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
		_ = svc.Apply(kit.Label(LabelFlow, t.flow.name), kit.Label(LabelTask, t.name))
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

func (t *Task) containerName() string {
	if t == nil || t.kind != kindContainer {
		return ""
	}
	if t.container != nil && t.container.Name() != "" {
		return t.container.Name()
	}
	if t.name != "" {
		return t.name
	}
	return ""
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

// List returns containers stamped with this flow's name, including stopped ones.
func (f *Flow) List(ctx context.Context) (kit.ContainerListResult, error) {
	if f == nil {
		return kit.ContainerListResult{}, errors.New("flow is nil")
	}
	return List(ctx, f.cli, f.name)
}

// Remove downs every Compose project the graph referenced, then force-removes
// Container-step IDs and any leftover containers stamped with this flow's name.
func (f *Flow) Remove(ctx context.Context) error {
	if f == nil {
		return errors.New("flow is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var errs []error
	apps := f.composeList()
	f.mu.Lock()
	ids := append([]string(nil), f.created...)
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
	leftover, err := f.List(ctx)
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
