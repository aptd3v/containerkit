package containerkit

import (
	"errors"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/require"
)

func TestApplyAndConditionals(t *testing.T) {
	c := NewContainer("t").Image("alpine")
	c.Apply(nil, func(c *Container) { c.Env("ON", "1") })
	require.Equal(t, []string{"ON=1"}, c.Config.Container.Env)

	c.Apply(WhenTrue(true, func(c *Container) { c.Label("a", "1") }))
	c.Apply(WhenTrue(false, func(c *Container) { c.Label("skip", "1") }))
	c.Apply(WhenTrueFn(nil, func(c *Container) { c.Label("nilpred", "1") }))
	c.Apply(WhenTrueFn(func() bool { return true }, nil, func(c *Container) { c.Label("b", "1") }))
	require.Equal(t, "1", c.Config.Container.Labels["a"])
	require.Equal(t, "1", c.Config.Container.Labels["b"])
	_, skip := c.Config.Container.Labels["skip"]
	require.False(t, skip)

	c.Apply(WhenTrueElse(true, func(c *Container) { c.Env("T", "yes") }, func(c *Container) { c.Env("T", "no") }))
	c.Apply(WhenTrueElse(false, func(c *Container) { c.Env("F", "yes") }, func(c *Container) { c.Env("F", "no") }))
	c.Apply(WhenTrueElseFn(nil, func(c *Container) { c.Env("N", "yes") }, func(c *Container) { c.Env("N", "else") }))

	require.True(t, And(true, true)())
	require.False(t, And(true, false)())
	require.True(t, AndFn(nil, func() bool { return true })())
	require.False(t, AndFn(func() bool { return false }, func() bool { return true })())
	require.True(t, Or(false, true)())
	require.False(t, Or(false, false)())
	require.True(t, OrFn(nil, func() bool { return true })())
	require.False(t, OrFn(func() bool { return false })())

	c.Apply(Group(nil, func(c *Container) { c.User("nobody") }))
	require.Equal(t, "nobody", c.Config.Container.User)

	c.Apply(OnlyIf(nil, func(c *Container) { c.WorkingDir("/x") }))
	require.Error(t, c.Validate())

	ok := NewContainer("ok")
	ok.Apply(OnlyIf(func() (bool, error) { return false, errors.New("boom") }, func(c *Container) { c.User("x") }))
	require.Error(t, ok.Validate())

	ok2 := NewContainer("ok2")
	ok2.Apply(OnlyIf(func() (bool, error) { return true, nil }, func(c *Container) { c.User("y") }))
	ok2.Apply(OnlyIf(func() (bool, error) { return false, nil }, func(c *Container) { c.User("z") }))
	require.NoError(t, ok2.Validate())
	require.Equal(t, "y", ok2.Config.Container.User)

	ok2.Apply(Each[string](nil, func(i int, s string) Mutator { return func(c *Container) { c.Env(s, "1") } }))
	ok2.Apply(Each([]string{"K1", "K2"}, nil))
	ok2.Apply(Each([]string{"K1", "K2"}, func(i int, s string) Mutator {
		if i == 0 {
			return nil
		}
		return func(c *Container) { c.Env(s, "1") }
	}))
	require.Contains(t, ok2.Config.Container.Env, "K2=1")
}

func TestWhenTrueOpt(t *testing.T) {
	called := 0
	fn := WhenTrueOpt(true, SetServiceConfig(nil), SetServiceConfig(func(s *types.ServiceConfig) error {
		called++
		s.Init = ptr(true)
		return nil
	}))
	svc := &types.ServiceConfig{}
	require.NoError(t, fn(svc))
	require.Equal(t, 1, called)

	skip := WhenTrueOpt(false, SetServiceConfig(func(s *types.ServiceConfig) error {
		t.Fatal("should not run")
		return nil
	}))
	require.NoError(t, skip(&types.ServiceConfig{}))

	fail := WhenTrueOpt(true, SetServiceConfig(func(s *types.ServiceConfig) error {
		return errors.New("nope")
	}))
	require.Error(t, fail(&types.ServiceConfig{}))
}

func ptr[T any](v T) *T { return &v }
