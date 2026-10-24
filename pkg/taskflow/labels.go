package taskflow

import (
	"context"
	"slices"
	"strings"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/client/response"
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
// stopped ones. Flow.List does the same for a flow in hand; this form works
// across processes.
func List(ctx context.Context, cli *client.Client, flowName string) ([]response.ContainerSummary, error) {
	if cli == nil {
		return nil, ErrNoClient
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return cli.ContainerList(ctx, &client.List{
		All:     true,
		Filters: make(client.Filters).Add("label", LabelFlow+"="+flowName),
	})
}

func (f *Flow) stamp() {
	if f == nil {
		return
	}
	for _, t := range f.taskList() {
		t.stamp()
	}
}

func (t *Task) stamp() {
	if t == nil {
		return
	}
	switch t.kind {
	case kindContainer:
		t.stampContainer()
	case kindCompose:
		t.stampCompose()
	}
}

func (t *Task) stampContainer() {
	if t.container == nil || t.alreadyCreated() {
		return
	}
	if t.container.Name == "" {
		t.container.Name = t.name
	}
	label(t.container, LabelFlow, t.flow.name)
	label(t.container, LabelTask, t.name)
	if deps := strings.Join(t.nearestNames(false), ","); deps != "" {
		label(t.container, LabelDepends, deps)
	}
	if succs := strings.Join(t.nearestNames(true), ","); succs != "" {
		label(t.container, LabelPrecedes, succs)
	}
}

func (t *Task) stampCompose() {
	proj := t.composeProject()
	if proj == nil {
		return
	}
	proj.ServiceLabel(LabelFlow, t.flow.name)
	proj.ServiceLabel(LabelTask, t.name)
}

func (t *Task) composeProject() *kit.Project {
	if t == nil || t.composeApp == nil {
		return nil
	}
	return t.composeApp.Project()
}

func label(c *kit.Container, key, value string) {
	if c == nil || key == "" || value == "" {
		return
	}
	c.Label(key, value)
}

func (t *Task) alreadyCreated() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.id != ""
}

func (t *Task) containerName() string {
	if t == nil || t.kind != kindContainer {
		return ""
	}
	if t.container != nil && t.container.Name != "" {
		return t.container.Name
	}
	return t.name
}

func (t *Task) composeContainerNames() []string {
	proj := t.composeProject()
	if proj == nil {
		return nil
	}
	raw := proj.Unwrap()
	if raw == nil {
		return nil
	}
	var names []string
	for _, svc := range raw.Services {
		if svc.ContainerName != "" {
			names = append(names, svc.ContainerName)
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
