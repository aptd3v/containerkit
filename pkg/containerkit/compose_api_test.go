package containerkit

import (
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptrInt(v int) *int       { return &v }
func ptrStr(v string) *string { return &v }
func ptrBool(v bool) *bool    { return &v }

func TestUpGenerateFlags(t *testing.T) {
	opt := &Up{
		NoAttach:             ptrStr("db"),
		AbortOnContainerExit: true,
		AlwaysRecreateDeps:   true,
		Build:                true,
		NoBuild:              true,
		Menu:                 true,
		ExitCodeFrom:         ptrStr("web"),
		ForceRecreate:        true,
		NoColor:              true,
		NoDeps:               true,
		NoLogPrefix:          true,
		NoRecreate:           true,
		Pull:                 ptrStr("never"),
		NoStart:              true,
		QuietPull:            true,
		RemoveOrphans:        true,
		RenewAnonVolumes:     true,
		Scale:                []UpScale{{Service: "web", Num: 2}},
		Timeout:              ptrInt(5),
		Timestamps:           true,
		WaitTimeout:          ptrInt(10),
		Yes:                  true,
	}
	flags, err := opt.GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, "up", flags[0])
	joined := join(flags)
	require.Contains(t, joined, "--no-attach")
	require.Contains(t, joined, "--scale")
	require.Contains(t, joined, "web=2")

	_, err = (&Up{Detach: true, Attach: ptrStr("web")}).GenerateFlags()
	require.Error(t, err)
	require.True(t, IsComposeFlagError(err))

	_, err = (&Up{Wait: true, Detach: true}).GenerateFlags()
	require.Error(t, err)

	_, err = (&Up{Attach: ptrStr("web"), AttachDependencies: true}).GenerateFlags()
	require.Error(t, err)

	_, err = (&Up{Scale: []UpScale{{Service: "", Num: -1}}}).GenerateFlags()
	require.Error(t, err)

	flags, err = (&Up{Detach: true}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, flags, "--detach")

	flags, err = (&Up{AttachDependencies: true}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, flags, "--attach-dependencies")

	flags, err = (&Up{Wait: true}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, flags, "--wait")

	flags, err = (&Up{Watch: true}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, flags, "--watch")
}

func TestOtherComposeGenerateFlags(t *testing.T) {
	down, err := (&Down{RemoveOrphans: true, Timeout: ptrInt(3), RemoveImage: ptrStr("local"), RemoveVolumes: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"down", "--remove-orphans", "--timeout", "3", "--rmi", "local", "--volumes"}, down)

	logs, err := (&Logs{Tail: ptrInt(20), Follow: true, NoLogPrefix: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"logs", "--tail", "20", "--follow", "--no-log-prefix"}, logs)

	kill, err := (&Kill{Signal: ptrStr("SIGTERM"), RemoveOrphans: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"kill", "--signal", "SIGTERM", "--remove-orphans"}, kill)

	ps, err := (&Ps{
		All: true, Filter: []string{"status=running"}, Format: "json", NoTrunc: true,
		Orphans: ptrBool(false), Quiet: true, Services: true, Status: "running",
	}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, ps, "--all")
	require.Contains(t, ps, "--no-orphans")
	require.Contains(t, ps, "--services")

	start, err := (&Start{Wait: true, WaitTimeout: ptrInt(9)}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"start", "--wait", "--wait-timeout", "9"}, start)

	stop, err := (&Stop{Timeout: ptrInt(4)}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"stop", "--timeout", "4"}, stop)

	restart, err := (&Restart{NoDeps: true, Timeout: ptrInt(2)}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"restart", "--no-deps", "--timeout", "2"}, restart)

	b, err := (&ComposeBuild{
		BuildArg: []string{"A=1"}, Builder: "default", Check: true, Memory: "2G",
		NoCache: true, Print: true, Provenance: true, Pull: true, Push: true,
		Quiet: true, SBOM: true, SSH: []string{"default"}, WithDependencies: true,
	}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, b, "--sbom")
	require.Contains(t, b, "--with-dependencies")

	pull, err := (&Pull{
		IgnoreBuildable: true, IgnorePullFailures: true, IncludeDeps: true,
		Policy: "missing", Quiet: true,
	}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"pull", "--ignore-buildable", "--ignore-pull-failures", "--include-deps", "--policy", "missing", "--quiet"}, pull)

	ex, err := (&Exec{
		Detach: true, Env: []string{"A=1"}, Index: ptrInt(1), NoTTY: true,
		Privileged: true, User: "root", Workdir: "/tmp",
	}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, ex, "--no-tty")
	require.Contains(t, ex, "--index")
	require.Equal(t, "1", strconv.Itoa(*ptrInt(1)))
	_ = io.Discard
}

func TestNewComposeGenerateFlags(t *testing.T) {
	attach, err := (&Attach{DetachKeys: "ctrl-p,ctrl-q", Index: ptrInt(2), NoStdin: true, NoSigProxy: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"attach", "--detach-keys", "ctrl-p,ctrl-q", "--index", "2", "--no-stdin", "--sig-proxy=false"}, attach)

	commit, err := (&Commit{Author: "a", Change: []string{"CMD /bin/sh"}, Index: ptrInt(1), Message: "m", NoPause: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"commit", "--author", "a", "--change", "CMD /bin/sh", "--index", "1", "--message", "m", "--pause=false"}, commit)

	cfg, err := (&ComposeConfig{
		PrintEnvironment: true, Format: "json", Hash: "*", PrintImages: true,
		LockImageDigests: true, PrintModels: true, PrintNetworks: true,
		NoConsistency: true, NoEnvResolution: true, NoInterpolate: true,
		NoNormalize: true, NoPathResolution: true, Output: "out.yml",
		PrintProfiles: true, Quiet: true, ResolveImageDigests: true,
		PrintServices: true, PrintVariables: true, PrintVolumes: true,
	}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, "config", cfg[0])
	require.Contains(t, cfg, "--environment")
	require.Contains(t, cfg, "--lock-image-digests")
	require.Contains(t, cfg, "--services")

	cp, err := (&Cp{All: true, Archive: true, FollowLink: true, Index: ptrInt(1)}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"cp", "--all", "--archive", "--follow-link", "--index", "1"}, cp)

	create, err := (&Create{
		Build: true, ForceRecreate: true, NoBuild: true, NoRecreate: true,
		Pull: "never", QuietPull: true, RemoveOrphans: true,
		Scale: []UpScale{{Service: "web", Num: 2}}, Yes: true,
	}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, create, "--force-recreate")
	require.Contains(t, create, "web=2")

	exp, err := (&ComposeExport{Index: ptrInt(1), Output: "fs.tar"}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"export", "--index", "1", "--output", "fs.tar"}, exp)

	images, err := (&Images{Format: "json", Quiet: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"images", "--format", "json", "--quiet"}, images)

	ls, err := (&Ls{All: true, Filter: []string{"status=running"}, Format: "json", Quiet: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"ls", "--all", "--filter", "status=running", "--format", "json", "--quiet"}, ls)

	pause, err := (&Pause{}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"pause"}, pause)

	port, err := (&Port{Index: ptrInt(1), Protocol: "udp"}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"port", "--index", "1", "--protocol", "udp"}, port)

	pub, err := (&Publish{App: true, OCIVersion: "1.1", ResolveImageDigests: true, WithEnv: true, Yes: true}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, pub, "--app")
	require.Contains(t, pub, "--with-env")

	push, err := (&Push{IgnorePushFailures: true, IncludeDeps: true, Quiet: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"push", "--ignore-push-failures", "--include-deps", "--quiet"}, push)

	rm, err := (&Rm{Force: true, Stop: true, Volumes: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"rm", "--force", "--stop", "--volumes"}, rm)

	run, err := (&Run{
		Build: true, CapAdd: []string{"NET_ADMIN"}, CapDrop: []string{"MKNOD"},
		Detach: true, Entrypoint: "/bin/sh", Env: []string{"A=1"}, EnvFromFile: []string{".env"},
		Interactive: ptrBool(false), Label: []string{"k=v"}, Name: "once", NoDeps: true,
		NoTTY: true, Publish: []string{"8080:80"}, Pull: "never", Quiet: true,
		QuietBuild: true, QuietPull: true, RemoveOrphans: true, Rm: true,
		ServicePorts: true, UseAliases: true, User: "root", Volume: []string{"/tmp:/tmp"},
		Workdir: "/app",
	}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, "run", run[0])
	require.Contains(t, run, "--interactive=false")
	require.Contains(t, run, "--no-TTY")
	require.Contains(t, run, "--service-ports")
	require.Contains(t, run, "--rm")

	scale, err := (&Scale{NoDeps: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"scale", "--no-deps"}, scale)

	stats, err := (&Stats{All: true, Format: "json", NoStream: true, NoTrunc: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"stats", "--all", "--format", "json", "--no-stream", "--no-trunc"}, stats)

	top, err := (&Top{}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"top"}, top)

	unpause, err := (&Unpause{}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"unpause"}, unpause)

	ver, err := (&Version{Format: "json", Short: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"version", "--format", "json", "--short"}, ver)

	vols, err := (&Volumes{Format: "json", Quiet: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"volumes", "--format", "json", "--quiet"}, vols)

	wait, err := (&Wait{DownProject: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"wait", "--down-project"}, wait)

	watch, err := (&Watch{NoUp: true, NoPrune: true, Quiet: true}).GenerateFlags()
	require.NoError(t, err)
	require.Equal(t, []string{"watch", "--no-up", "--prune=false", "--quiet"}, watch)

	runTrue, err := (&Run{Interactive: ptrBool(true)}).GenerateFlags()
	require.NoError(t, err)
	require.Contains(t, runTrue, "--interactive")
}

func join(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}
