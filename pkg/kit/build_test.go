package kit

import (
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/require"
)

func TestNewBuild(t *testing.T) {
	cfg := &types.ServiceConfig{}
	err := NewBuild().
		Context(".").
		DockerfilePath("Dockerfile").
		Arg("A", "1").
		Args(map[string]string{"B": "2"}).
		NoCache().
		SSH("default", "/tmp/sock").
		Ulimit("nofile", 10, 20).
		Apply(cfg)
	require.NoError(t, err)
	require.Equal(t, ".", cfg.Build.Context)
	require.Equal(t, "Dockerfile", cfg.Build.Dockerfile)
	require.Equal(t, "1", *cfg.Build.Args["A"])
	require.Equal(t, "2", *cfg.Build.Args["B"])
	require.True(t, cfg.Build.NoCache)
	require.Equal(t, "default", cfg.Build.SSH[0].ID)
	require.Equal(t, 10, cfg.Build.Ulimits["nofile"].Soft)

	from := NewBuildFrom(BuildSpec{Context: "./app", Pull: true})
	cfg = &types.ServiceConfig{}
	require.NoError(t, from.Apply(cfg))
	require.Equal(t, "./app", cfg.Build.Context)
	require.True(t, cfg.Build.Pull)

	cfg = &types.ServiceConfig{}
	require.NoError(t, NewBuild().Dockerfile(NewDockerFile().From("scratch", "")).Apply(cfg))
	require.Equal(t, "FROM scratch\n", cfg.Build.DockerfileInline)
	require.NoError(t, NewBuild().Dockerfile(nil).Apply(cfg))
}
