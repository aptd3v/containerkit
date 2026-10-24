package taskflow

import (
	"context"
	"errors"
	"testing"

	"github.com/aptd3v/containerkit/pkg/client/response"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/require"
)

func TestWaitCode(t *testing.T) {
	code, err := waitCode(response.ContainerWait{WaitResponse: container.WaitResponse{StatusCode: 0}})
	require.NoError(t, err)
	require.Equal(t, int64(0), code)

	code, err = waitCode(response.ContainerWait{WaitResponse: container.WaitResponse{StatusCode: 2}})
	require.NoError(t, err)
	require.Equal(t, int64(2), code)

	code, err = waitCode(response.ContainerWait{WaitResponse: container.WaitResponse{
		StatusCode: 1,
		Error:      &container.WaitExitError{Message: "oom"},
	}})
	require.EqualError(t, err, "oom")
	require.Equal(t, int64(1), code)
}

func TestWaitStatus(t *testing.T) {
	t.Run("status", func(t *testing.T) {
		waitCh := make(chan response.ContainerWait, 1)
		errCh := make(chan error, 1)
		waitCh <- response.ContainerWait{WaitResponse: container.WaitResponse{StatusCode: 3}}
		code, err := waitStatus(context.Background(), waitCh, errCh)
		require.NoError(t, err)
		require.Equal(t, int64(3), code)
	})
	t.Run("transport error", func(t *testing.T) {
		waitCh := make(chan response.ContainerWait)
		errCh := make(chan error, 1)
		errCh <- errors.New("daemon")
		_, err := waitStatus(context.Background(), waitCh, errCh)
		require.EqualError(t, err, "daemon")
	})
	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := waitStatus(ctx, make(chan response.ContainerWait), make(chan error))
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestCheckContainerJob(t *testing.T) {
	require.NoError(t, checkContainerJob(nil))
	require.NoError(t, checkContainerJob(kit.NewContainer().Image("alpine")))
	require.NoError(t, checkContainerJob(kit.NewContainer().Image("alpine").RestartPolicyNever()))

	err := checkContainerJob(kit.NewContainer().Image("alpine").RestartAlways())
	require.Error(t, err)
	require.Contains(t, err.Error(), "always")
	require.Contains(t, err.Error(), "*kit.Up")

	err = checkContainerJob(kit.NewContainer().Image("alpine").RestartPolicyUnlessStopped())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unless-stopped")

	err = checkContainerJob(kit.NewContainer().Image("alpine").RestartPolicyOnFailure(3))
	require.Error(t, err)
	require.Contains(t, err.Error(), "on-failure")
	require.Contains(t, err.Error(), "Retry")

	err = checkContainerJob(kit.NewContainer().Image("alpine").AutoRemove())
	require.Error(t, err)
	require.Contains(t, err.Error(), "autoremove")
}

func TestContainerRejectsRestartPolicyWithoutClient(t *testing.T) {
	flow := New("job", nil)
	flow.Container("c", kit.NewContainer("c").Image("alpine").RestartAlways(), nil)
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "always")
	require.NotErrorIs(t, err, ErrNoClient)
}

func TestTaskErrorUnwrap(t *testing.T) {
	inner := errors.New("exit code 1")
	err := taskErr("unit", inner)
	var te *TaskError
	require.ErrorAs(t, err, &te)
	require.Equal(t, "unit", te.Task)
	require.ErrorIs(t, err, inner)
	require.Equal(t, "task unit: exit code 1", err.Error())
	require.ErrorIs(t, taskErr("unit", ErrNoClient), ErrNoClient)
}
