package taskflow

import (
	"context"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/stretchr/testify/require"
)

func labelsOf(t *testing.T, ctr *kit.Container) map[string]string {
	t.Helper()
	require.NotNil(t, ctr.Config)
	require.NotNil(t, ctr.Config.Container)
	return ctr.Config.Container.Labels
}

func TestStampRelationLabels(t *testing.T) {
	flow := New("report", nil)
	export := flow.Container("export", kit.NewContainer("export").Image("alpine"), nil)
	gate := flow.Condition("gate", func(context.Context) uint { return 0 })
	schema := flow.Container("schema", kit.NewContainer("schema").Image("alpine"), nil)
	bundle := flow.Container("bundle", kit.NewContainer("bundle").Image("alpine"), nil)
	publish := flow.Task("publish", func(context.Context) error { return nil })
	export.Precede(gate)
	gate.Precede(schema, bundle)
	schema.Precede(publish)
	bundle.Precede(publish)
	flow.stamp()

	el := labelsOf(t, export.container)
	require.Equal(t, "report", el[LabelFlow])
	require.Equal(t, "export", el[LabelTask])
	require.Equal(t, "bundle,schema", el[LabelPrecedes])
	require.Empty(t, el[LabelDepends])

	sl := labelsOf(t, schema.container)
	require.Equal(t, "export", sl[LabelDepends])
	require.Empty(t, sl[LabelPrecedes])

	bl := labelsOf(t, bundle.container)
	require.Equal(t, "export", bl[LabelDepends])

	require.Nil(t, publish.container)
}

func TestStampSkipsSelfEdgeOnRetry(t *testing.T) {
	flow := New("loop", nil)
	export := flow.Container("export", kit.NewContainer("export").Image("alpine"), nil)
	inspect := flow.Task("inspect", func(context.Context) error { return nil })
	gate := flow.Condition("gate", func(context.Context) uint { return 1 })
	retry := flow.Condition("retry", func(context.Context) uint { return 0 })
	fail := flow.Task("fail", func(context.Context) error { return nil })
	export.Precede(inspect)
	inspect.Precede(gate)
	gate.Precede(fail, retry)
	retry.Precede(export)
	flow.stamp()

	el := labelsOf(t, export.container)
	require.Empty(t, el[LabelDepends])
	require.Empty(t, el[LabelPrecedes])
}

func TestEmptyContainerNameUsesTaskName(t *testing.T) {
	flow := New("names", nil)
	c := flow.Container("unit", kit.NewContainer().Image("alpine"), nil)
	require.Equal(t, "unit", c.container.Name)
	flow.stamp()
	require.Equal(t, "unit", labelsOf(t, c.container)[LabelTask])
}

func TestListRequiresClient(t *testing.T) {
	_, err := List(context.Background(), nil, "ci")
	require.ErrorIs(t, err, ErrNoClient)
}

func TestStampComposeServices(t *testing.T) {
	proj := kit.NewProject("db")
	ctr := kit.NewContainer("ck-db").Image("postgres:16-alpine")
	proj.WithService(kit.NewService("db", ctr))
	db := kit.NewCompose(proj)

	flow := New("ci", nil)
	up := flow.Compose("up-db", db, &kit.Up{})
	unit := flow.Container("unit", kit.NewContainer("ck-unit").Image("alpine"), nil)
	down := flow.Compose("down-db", db, &kit.Down{}).Always()
	up.Precede(unit)
	unit.Precede(down)
	flow.stamp()

	svc, err := proj.GetService("db")
	require.NoError(t, err)
	require.Equal(t, "ci", svc.Labels[LabelFlow])
	require.Equal(t, "up-db", svc.Labels[LabelTask])
	require.Empty(t, svc.Labels[LabelDepends])
	require.Empty(t, svc.Labels[LabelPrecedes])

	ul := labelsOf(t, unit.container)
	require.Equal(t, "ck-db", ul[LabelDepends])
	require.Equal(t, "ck-db", ul[LabelPrecedes])
}
