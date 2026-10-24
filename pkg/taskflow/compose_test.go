package taskflow

import (
	"context"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/stretchr/testify/require"
)

func TestComposeCommandOptions(t *testing.T) {
	flow := New("cmp", nil)
	cmds := []any{
		&kit.Up{},
		&kit.Down{},
		&kit.Logs{Follow: true},
		&kit.Kill{},
		&kit.Ps{},
		&kit.Start{},
		&kit.Stop{},
		&kit.Restart{},
		&kit.ComposeBuild{},
		&kit.Pull{},
		&kit.Exec{},
		&kit.Attach{},
		&kit.Commit{},
		&kit.ComposeConfig{},
		&kit.Cp{},
		&kit.Create{},
		&kit.ComposeExport{},
		&kit.Images{},
		&kit.Ls{},
		&kit.Pause{},
		&kit.Port{},
		&kit.Publish{},
		&kit.Push{},
		&kit.Rm{},
		&kit.Run{},
		&kit.Scale{},
		&kit.Stats{},
		&kit.Top{},
		&kit.Unpause{},
		&kit.Version{},
		&kit.Volumes{},
		&kit.Wait{},
		&kit.Watch{},
		&kit.Viz{},
		&kit.Generate{},
		&kit.BridgeConvert{},
		&kit.BridgeTransformationsList{},
		&kit.BridgeTransformationsCreate{},
		(*kit.Up)(nil),
	}
	for _, cmd := range cmds {
		fn, err := composeCommand(flow, "step", cmd)
		require.NoError(t, err, "%T", cmd)
		require.ErrorIs(t, fn(context.Background(), nil), errNilCompose, "%T", cmd)
	}
}

func TestComposeCommandFuncAndRejects(t *testing.T) {
	flow := New("cmp", nil)
	called := false
	fn, err := composeCommand(flow, "ps", func(context.Context, *kit.Compose) error {
		called = true
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, fn(context.Background(), nil))
	require.True(t, called)

	_, err = composeCommand(flow, "x", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil")

	var nilFn func(context.Context, *kit.Compose) error
	_, err = composeCommand(flow, "x", nilFn)
	require.Error(t, err)

	_, err = composeCommand(flow, "x", 42)
	require.Error(t, err)
	require.Contains(t, err.Error(), "int")

	_, err = composeCommand(flow, "x", &kit.EventsOpt{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "EventsOpt")
}

func TestComposeCopiesOption(t *testing.T) {
	flow := New("cmp", nil)
	opt := &kit.Logs{Follow: true}
	fn, err := composeCommand(flow, "logs", opt)
	require.NoError(t, err)
	require.ErrorIs(t, fn(context.Background(), nil), errNilCompose)
	require.Nil(t, opt.Writer)
	require.True(t, opt.Follow)
}

func TestComposeLogsWithoutClient(t *testing.T) {
	flow := New("cmp", nil)
	logs := flow.Compose("logs", nil, &kit.Logs{Follow: true})
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	require.ErrorIs(t, logs.Err(), errNilCompose)
}
