package containerkit

import (
	"errors"
	"fmt"
	"testing"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/stretchr/testify/require"
)

func TestComposeErrorTypes(t *testing.T) {
	inner := errors.New("boom")

	cases := []struct {
		err   error
		is    func(error) bool
		check string
	}{
		{NewComposeFlagError("--x", "bad"), IsComposeFlagError, "compose flag error"},
		{NewComposeKillError(inner), IsComposeKillError, "compose kill error"},
		{NewComposeEventsError(inner), IsComposeEventsError, "compose events error"},
		{NewComposeError(inner), IsComposeError, "compose exec error"},
		{NewComposeUpError(inner), IsComposeUpError, "compose up error"},
		{NewComposeDownError(inner), IsComposeDownError, "compose down error"},
		{NewComposeLogsError(inner), IsComposeLogsError, "compose logs error"},
		{NewComposePsError(inner), IsComposePsError, "compose ps error"},
		{NewComposeStartError(inner), IsComposeStartError, "compose start error"},
		{NewComposeStopError(inner), IsComposeStopError, "compose stop error"},
		{NewComposeRestartError(inner), IsComposeRestartError, "compose restart error"},
		{NewComposeBuildError(inner), IsComposeBuildError, "compose build error"},
		{NewComposePullError(inner), IsComposePullError, "compose pull error"},
		{NewComposeExecError(inner), IsComposeExecError, "compose exec error"},
		{NewComposeAttachError(inner), IsComposeAttachError, "compose attach error"},
		{NewComposeCommitError(inner), IsComposeCommitError, "compose commit error"},
		{NewComposeConfigError(inner), IsComposeConfigError, "compose config error"},
		{NewComposeCpError(inner), IsComposeCpError, "compose cp error"},
		{NewComposeCreateError(inner), IsComposeCreateError, "compose create error"},
		{NewComposeExportError(inner), IsComposeExportError, "compose export error"},
		{NewComposeImagesError(inner), IsComposeImagesError, "compose images error"},
		{NewComposeLsError(inner), IsComposeLsError, "compose ls error"},
		{NewComposePauseError(inner), IsComposePauseError, "compose pause error"},
		{NewComposePortError(inner), IsComposePortError, "compose port error"},
		{NewComposePublishError(inner), IsComposePublishError, "compose publish error"},
		{NewComposePushError(inner), IsComposePushError, "compose push error"},
		{NewComposeRmError(inner), IsComposeRmError, "compose rm error"},
		{NewComposeRunError(inner), IsComposeRunError, "compose run error"},
		{NewComposeScaleError(inner), IsComposeScaleError, "compose scale error"},
		{NewComposeStatsError(inner), IsComposeStatsError, "compose stats error"},
		{NewComposeTopError(inner), IsComposeTopError, "compose top error"},
		{NewComposeUnpauseError(inner), IsComposeUnpauseError, "compose unpause error"},
		{NewComposeVersionError(inner), IsComposeVersionError, "compose version error"},
		{NewComposeVolumesError(inner), IsComposeVolumesError, "compose volumes error"},
		{NewComposeWaitError(inner), IsComposeWaitError, "compose wait error"},
		{NewComposeWatchError(inner), IsComposeWatchError, "compose watch error"},
	}
	for _, tc := range cases {
		require.True(t, tc.is(tc.err), tc.check)
		require.Contains(t, tc.err.Error(), tc.check)
		require.NotNil(t, errors.Unwrap(tc.err))
	}

	require.True(t, errors.Is(NewComposeKillError(inner), inner))
	require.True(t, errors.Is(NewComposeKillError(inner), ErrComposeKillError))
	require.True(t, IsComposeKillError(NewComposeKillError(nil)))
	require.True(t, errors.Is(NewComposeFlagError("--x", "bad"), ErrComposeFlagError))
}

func TestComposeErrorUnwrapsNotFound(t *testing.T) {
	cause := fmt.Errorf("no such container: missing: %w", cerrdefs.ErrNotFound)
	err := NewComposeUpError(cause)
	require.True(t, IsComposeUpError(err))
	require.True(t, cerrdefs.IsNotFound(err))
	require.True(t, errors.Is(err, cause))
}
