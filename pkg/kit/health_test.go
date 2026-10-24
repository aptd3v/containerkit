package kit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHealthBuilder(t *testing.T) {
	h := NewHealth().CMD("true").Timeout(2).Interval(3).StartPeriod(1).Retries(4)
	cfg, err := h.healthConfig()
	require.NoError(t, err)
	require.Equal(t, []string{"CMD", "true"}, cfg.Test)
	require.Equal(t, 2*time.Second, cfg.Timeout)
	require.Equal(t, 3*time.Second, cfg.Interval)
	require.Equal(t, time.Second, cfg.StartPeriod)
	require.Equal(t, 4, cfg.Retries)

	shell, err := NewHealth().CMDShell("curl -f http://localhost || exit 1").TimeoutD("5s").IntervalD("10s").StartPeriodD("1s").healthConfig()
	require.NoError(t, err)
	require.Equal(t, []string{"CMD-SHELL", "curl -f http://localhost || exit 1"}, shell.Test)
	require.Equal(t, 5*time.Second, shell.Timeout)

	none, err := NewHealth().None().healthConfig()
	require.NoError(t, err)
	require.Equal(t, []string{"NONE"}, none.Test)

	empty, err := NewHealth().healthConfig()
	require.NoError(t, err)
	require.Equal(t, []string{"NONE"}, empty.Test)

	base := NewHealth().Timeout(1)
	a, err := base.CMD("a").healthConfig()
	require.NoError(t, err)
	b, err := base.CMD("b").healthConfig()
	require.NoError(t, err)
	require.Equal(t, []string{"CMD", "a"}, a.Test)
	require.Equal(t, []string{"CMD", "b"}, b.Test)
}
