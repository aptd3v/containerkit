package kit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpValidate(t *testing.T) {
	require.NoError(t, (&Up{Detach: true}).validate())
	require.NoError(t, (&Up{AttachDependencies: true}).validate())
	require.NoError(t, (&Up{Wait: true}).validate())
	require.NoError(t, (&Up{Watch: true}).validate())
	require.NoError(t, (&Up{Scale: []UpScale{{Service: "web", Num: 2}}}).validate())

	web := "web"
	require.NoError(t, (&Up{Wait: true, Detach: true}).validate())
	require.Error(t, (&Up{Detach: true, Attach: &web}).validate())
	require.Error(t, (&Up{Wait: true, Attach: &web}).validate())
	require.Error(t, (&Up{Attach: &web, AttachDependencies: true}).validate())
	require.Error(t, (&Up{Scale: []UpScale{{Service: "", Num: -1}}}).validate())

	err := (&Up{Detach: true, Wait: true, Attach: &web}).validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "cannot be combined with detach")
	require.ErrorContains(t, err, "cannot be combined with wait")
}

func TestUpAttachTo(t *testing.T) {
	names := []string{"web", "db"}
	require.Equal(t, names, (&Up{}).attachTo(names))
	require.Nil(t, (&Up{Detach: true}).attachTo(names))
	require.Nil(t, (&Up{Wait: true}).attachTo(names))
	require.Equal(t, []string{"web"}, (&Up{Attach: &names[0]}).attachTo(names))
}
