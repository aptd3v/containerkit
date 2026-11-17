package kit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	clienttypes "github.com/aptd3v/containerkit/internal/client/types"
	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
	containertypes "github.com/aptd3v/containerkit/internal/engine/container/types"
	"github.com/docker/cli/opts"
	buildapi "github.com/moby/moby/api/types/build"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type upAct struct {
	create   createAct
	start    startAct
	log      *Logger
	detached bool
}

func (a upAct) sdk() composeapi.UpOptions {
	if a.log != nil {
		a.create.quietPull.assign(true)
		if a.create.build != nil {
			a.create.build.log = a.log
		}
	}
	return composeapi.UpOptions{Create: a.create.sdk(), Start: a.start.sdk()}
}

type createAct struct {
	services     []string
	removeOrphan bool
	ignoreOrphan bool
	recreate     string
	recreateDeps string
	inherit      field[bool]
	timeout      *time.Duration
	quietPull    field[bool]
	assumeYes    bool
	build        *composeBuildAct
	log          *Logger
}

func (a createAct) sdk() composeapi.CreateOptions {
	quietPull := a.quietPull.val
	if a.log != nil {
		quietPull = true
		if a.build != nil {
			a.build.log = a.log
		}
	}
	out := composeapi.CreateOptions{
		Services:             append([]string(nil), a.services...),
		RemoveOrphans:        a.removeOrphan,
		IgnoreOrphans:        a.ignoreOrphan,
		Recreate:             a.recreate,
		RecreateDependencies: a.recreateDeps,
		Timeout:              a.timeout,
		QuietPull:            quietPull,
	}
	if a.inherit.ok() {
		out.Inherit = a.inherit.val
	}
	if a.build != nil {
		b := a.build.sdk()
		out.Build = &b
	}
	return out
}

func (a *createAct) ensureBuild() *composeBuildAct {
	if a.build == nil {
		a.build = &composeBuildAct{}
	}
	return a.build
}

type startAct struct {
	services    []string
	attachTo    []string
	wait        bool
	waitTimeout time.Duration
	watch       bool
	navMenu     bool
	exitFrom    string
	onExit      composeapi.Cascade
	consumer    composeapi.LogConsumer
	log         *Logger
}

func (a startAct) sdk() composeapi.StartOptions {
	return composeapi.StartOptions{
		Attach:         a.consumer,
		AttachTo:       append([]string(nil), a.attachTo...),
		OnExit:         a.onExit,
		ExitCodeFrom:   a.exitFrom,
		Wait:           a.wait,
		WaitTimeout:    a.waitTimeout,
		Services:       append([]string(nil), a.services...),
		Watch:          a.watch,
		NavigationMenu: a.navMenu,
	}
}

type downAct struct {
	removeOrphan bool
	timeout      *time.Duration
	images       string
	volumes      bool
	services     []string
	log          *Logger
}

func (a downAct) sdk() composeapi.DownOptions {
	return composeapi.DownOptions{
		RemoveOrphans: a.removeOrphan,
		Timeout:       a.timeout,
		Images:        composeapi.ImagePruneMode(a.images),
		Volumes:       a.volumes,
		Services:      append([]string(nil), a.services...),
	}
}

type composeBuildAct struct {
	pull         bool
	push         bool
	progress     string
	args         composetypes.MappingWithEquals
	noCache      bool
	quiet        bool
	services     []string
	deps         bool
	memory       int64
	builder      string
	print        bool
	check        bool
	attestations bool
	provenance   string
	sbom         string
	sshs         []composetypes.SSHKey
	out          io.Writer
	log          *Logger
}

func (a composeBuildAct) sdk() composeapi.BuildOptions {
	out := a.out
	quiet := a.quiet
	if a.log != nil && out == nil {
		quiet = true
		out = io.Discard
	}
	return composeapi.BuildOptions{
		Pull:         a.pull,
		Push:         a.push,
		Progress:     a.progress,
		Args:         a.args,
		NoCache:      a.noCache,
		Quiet:        quiet,
		Services:     append([]string(nil), a.services...),
		Deps:         a.deps,
		SSHs:         append([]composetypes.SSHKey(nil), a.sshs...),
		Memory:       a.memory,
		Builder:      a.builder,
		Print:        a.print,
		Check:        a.check,
		Attestations: a.attestations,
		Provenance:   a.provenance,
		SBOM:         a.sbom,
		Out:          out,
	}
}

type stopAct struct {
	timeout  *time.Duration
	services []string
	log      *Logger
}

func (a stopAct) sdk() composeapi.StopOptions {
	return composeapi.StopOptions{Timeout: a.timeout, Services: append([]string(nil), a.services...)}
}

type restartAct struct {
	timeout  *time.Duration
	services []string
	noDeps   bool
	log      *Logger
}

func (a restartAct) sdk() composeapi.RestartOptions {
	return composeapi.RestartOptions{Timeout: a.timeout, Services: append([]string(nil), a.services...), NoDeps: a.noDeps}
}

type pullAct struct {
	quiet           bool
	ignoreFailures  bool
	ignoreBuildable bool
	log             *Logger
}

func (a pullAct) sdk() composeapi.PullOptions {
	return composeapi.PullOptions{
		Quiet:           a.quiet || a.log != nil,
		IgnoreFailures:  a.ignoreFailures,
		IgnoreBuildable: a.ignoreBuildable,
	}
}

type pushAct struct {
	quiet          bool
	ignoreFailures bool
	imageMandatory bool
	log            *Logger
}

func (a pushAct) sdk() composeapi.PushOptions {
	return composeapi.PushOptions{
		Quiet:          a.quiet || a.log != nil,
		IgnoreFailures: a.ignoreFailures,
		ImageMandatory: a.imageMandatory,
	}
}

type logsAct struct {
	services   []string
	tail       string
	since      string
	until      string
	follow     bool
	timestamps bool
	index      int
	stdout     field[bool]
	stderr     field[bool]
	details    bool
	log        *Logger
}

func (a logsAct) composeSDK(proj *composetypes.Project) composeapi.LogOptions {
	return composeapi.LogOptions{
		Project:    proj,
		Index:      a.index,
		Services:   append([]string(nil), a.services...),
		Tail:       a.tail,
		Since:      a.since,
		Until:      a.until,
		Follow:     a.follow,
		Timestamps: a.timestamps,
	}
}

func (a logsAct) engineSDK() clienttypes.ContainerLogsOptions {
	out := clienttypes.ContainerLogsOptions{
		Since:      a.since,
		Until:      a.until,
		Timestamps: a.timestamps,
		Follow:     a.follow,
		Tail:       a.tail,
		Details:    a.details,
	}
	if a.stdout.ok() {
		out.ShowStdout = a.stdout.val
	} else {
		out.ShowStdout = true
	}
	if a.stderr.ok() {
		out.ShowStderr = a.stderr.val
	} else {
		out.ShowStderr = true
	}
	return out
}

type psAct struct {
	all      bool
	services []string
}

func (a psAct) sdk(proj *composetypes.Project) composeapi.PsOptions {
	return composeapi.PsOptions{Project: proj, All: a.all, Services: append([]string(nil), a.services...)}
}

type killAct struct {
	removeOrphan bool
	services     []string
	signal       string
	all          bool
}

func (a killAct) composeSDK(proj *composetypes.Project) composeapi.KillOptions {
	return composeapi.KillOptions{
		RemoveOrphans: a.removeOrphan,
		Project:       proj,
		Services:      append([]string(nil), a.services...),
		Signal:        a.signal,
		All:           a.all,
	}
}

func (a killAct) engineSDK() clienttypes.ContainerKillOptions {
	return clienttypes.ContainerKillOptions{Signal: a.signal}
}

type removeAct struct {
	stop     bool
	volumes  bool
	force    bool
	links    bool
	services []string
}

func (a removeAct) composeSDK(proj *composetypes.Project) composeapi.RemoveOptions {
	return composeapi.RemoveOptions{
		Project:  proj,
		Stop:     a.stop,
		Volumes:  a.volumes,
		Force:    a.force,
		Services: append([]string(nil), a.services...),
	}
}

func (a removeAct) engineSDK() clienttypes.ContainerRemoveOptions {
	return clienttypes.ContainerRemoveOptions{RemoveVolumes: a.volumes, RemoveLinks: a.links, Force: a.force}
}

type runAct struct {
	create   createAct
	name     string
	service  string
	command  []string
	entry    []string
	detach   bool
	autoRm   bool
	tty      field[bool]
	stdin    field[bool]
	workdir  string
	user     string
	env      []string
	capAdd   []string
	capDrop  []string
	labels   composetypes.Labels
	priv     bool
	noDeps   bool
	index    int
	useAlias bool
}

func (a runAct) sdk(proj *composetypes.Project, service string) composeapi.RunOptions {
	if service == "" {
		service = a.service
	}
	return composeapi.RunOptions{
		CreateOptions:     a.create.sdk(),
		Project:           proj,
		Name:              a.name,
		Service:           service,
		Command:           append([]string(nil), a.command...),
		Entrypoint:        append([]string(nil), a.entry...),
		Detach:            a.detach,
		AutoRemove:        a.autoRm,
		Tty:               a.tty.val,
		Interactive:       a.stdin.val,
		WorkingDir:        a.workdir,
		User:              a.user,
		Environment:       append([]string(nil), a.env...),
		CapAdd:            append([]string(nil), a.capAdd...),
		CapDrop:           append([]string(nil), a.capDrop...),
		Labels:            a.labels,
		Privileged:        a.priv,
		NoDeps:            a.noDeps,
		Index:             a.index,
		UseNetworkAliases: a.useAlias,
	}
}

type pauseAct struct {
	services []string
}

func (a pauseAct) sdk(proj *composetypes.Project) composeapi.PauseOptions {
	return composeapi.PauseOptions{Services: append([]string(nil), a.services...), Project: proj}
}

type eventsAct struct {
	services []string
	consumer func(composeapi.Event) error
	since    string
	until    string
}

func (a eventsAct) sdk() composeapi.EventsOptions {
	return composeapi.EventsOptions{Services: append([]string(nil), a.services...), Consumer: a.consumer, Since: a.since, Until: a.until}
}

type composePortAct struct {
	protocol string
	index    int
}

func (a composePortAct) sdk() composeapi.PortOptions {
	return composeapi.PortOptions{Protocol: a.protocol, Index: a.index}
}

type watchAct struct {
	build    *composeBuildAct
	logTo    composeapi.LogConsumer
	prune    bool
	services []string
	log      *Logger
}

func (a watchAct) sdk() composeapi.WatchOptions {
	logTo := a.logTo
	if logTo == nil && a.log != nil {
		logTo = a.log
	}
	out := composeapi.WatchOptions{LogTo: logTo, Prune: a.prune, Services: append([]string(nil), a.services...)}
	if a.build != nil {
		if a.log != nil {
			a.build.log = a.log
		}
		b := a.build.sdk()
		out.Build = &b
	}
	return out
}

func (a *watchAct) ensureBuild() *composeBuildAct {
	if a.build == nil {
		a.build = &composeBuildAct{}
	}
	return a.build
}

type waitAct struct {
	services []string
	downExit bool
}

func (a waitAct) sdk() composeapi.WaitOptions {
	return composeapi.WaitOptions{Services: append([]string(nil), a.services...), DownProjectOnContainerExit: a.downExit}
}

type scaleAct struct {
	services    []string
	replicas    int
	replicasSet bool
}

func (a scaleAct) sdk() composeapi.ScaleOptions {
	out := composeapi.ScaleOptions{Services: append([]string(nil), a.services...)}
	if a.replicasSet {
		out.Replicas = map[string]int{}
		for _, name := range a.services {
			out.Replicas[name] = a.replicas
		}
	}
	return out
}

type copyAct struct {
	source      string
	destination string
	all         bool
	index       int
	followLink  bool
	copyUIDGID  bool
}

func (a copyAct) sdk() composeapi.CopyOptions {
	return composeapi.CopyOptions{
		Source:      a.source,
		Destination: a.destination,
		All:         a.all,
		Index:       a.index,
		FollowLink:  a.followLink,
		CopyUIDGID:  a.copyUIDGID,
	}
}

type attachAct struct {
	service    string
	index      int
	detachKeys string
	noStdin    bool
	proxy      bool
}

func (a attachAct) sdk(proj *composetypes.Project) composeapi.AttachOptions {
	return composeapi.AttachOptions{Project: proj, Service: a.service, Index: a.index, DetachKeys: a.detachKeys, NoStdin: a.noStdin, Proxy: a.proxy}
}

type publishAct struct {
	resolveDigest bool
	application   bool
	withEnv       bool
	assumeYes     bool
	ociVersion    composeapi.OCIVersion
}

func (a publishAct) sdk() composeapi.PublishOptions {
	return composeapi.PublishOptions{
		ResolveImageDigests: a.resolveDigest,
		Application:         a.application,
		WithEnvironment:     a.withEnv,
		OCIVersion:          a.ociVersion,
	}
}

type imagesAct struct {
	services []string
}

func (a imagesAct) sdk() composeapi.ImagesOptions {
	return composeapi.ImagesOptions{Services: append([]string(nil), a.services...)}
}

type exportAct struct {
	service string
	index   int
	output  string
}

func (a exportAct) sdk() composeapi.ExportOptions {
	return composeapi.ExportOptions{Service: a.service, Index: a.index, Output: a.output}
}

type commitAct struct {
	service   string
	reference string
	pause     field[bool]
	comment   string
	author    string
	index     int
	changes   []string
}

func (a commitAct) sdk() composeapi.CommitOptions {
	out := composeapi.CommitOptions{Service: a.service, Reference: a.reference, Comment: a.comment, Author: a.author, Index: a.index}
	if a.pause.ok() {
		out.Pause = a.pause.val
	}
	if len(a.changes) > 0 {
		lo := opts.NewListOpts(nil)
		for _, c := range a.changes {
			_ = lo.Set(c)
		}
		out.Changes = lo
	}
	return out
}

type vizAct struct {
	networks bool
	ports    bool
	image    bool
	indent   string
}

func (a vizAct) sdk() composeapi.VizOptions {
	return composeapi.VizOptions{IncludeNetworks: a.networks, IncludePorts: a.ports, IncludeImageName: a.image, Indentation: a.indent}
}

type listAct struct {
	all bool
}

func (a listAct) sdk() composeapi.ListOptions {
	return composeapi.ListOptions{All: a.all}
}

type volumesListAct struct {
	services []string
}

func (a volumesListAct) sdk() composeapi.VolumesOptions {
	return composeapi.VolumesOptions{Services: append([]string(nil), a.services...)}
}

type engineCreateAct struct {
	name string
}

type engineStartAct struct {
	checkpoint    string
	checkpointDir string
}

func (a engineStartAct) sdk() clienttypes.ContainerStartOptions {
	return clienttypes.ContainerStartOptions{CheckpointID: a.checkpoint, CheckpointDir: a.checkpointDir}
}

type engineStopAct struct {
	signal  string
	timeout *int
}

func (a engineStopAct) stopSDK() clienttypes.ContainerStopOptions {
	return clienttypes.ContainerStopOptions{Signal: a.signal, Timeout: a.timeout}
}

func (a engineStopAct) restartSDK() clienttypes.ContainerRestartOptions {
	return clienttypes.ContainerRestartOptions{Signal: a.signal, Timeout: a.timeout}
}

type engineInspectAct struct {
	size bool
}

func (a engineInspectAct) sdk() clienttypes.ContainerInspectOptions {
	return clienttypes.ContainerInspectOptions{Size: a.size}
}

type engineWaitAct struct {
	condition WaitCondition
}

func (a engineWaitAct) sdk() clienttypes.ContainerWaitOptions {
	return clienttypes.ContainerWaitOptions{Condition: containertypes.WaitCondition(a.condition)}
}

type generateAct struct {
	name       string
	containers []string
}

func (a generateAct) sdk() composeapi.GenerateOptions {
	return composeapi.GenerateOptions{ProjectName: a.name, Containers: append([]string(nil), a.containers...)}
}

type enginePullAct struct {
	all       bool
	auth      string
	platforms []ocispec.Platform
	log       *Logger
}

func (a enginePullAct) sdk() clienttypes.ImagePullOptions {
	return clienttypes.ImagePullOptions{All: a.all, RegistryAuth: a.auth, Platforms: append([]ocispec.Platform(nil), a.platforms...)}
}

type engineBuildAct struct {
	tags       []string
	quiet      bool
	noCache    bool
	forceRm    bool
	remove     bool
	pull       bool
	memory     int64
	dockerfile string
	target     string
	labels     map[string]string
	args       map[string]*string
	platforms  []ocispec.Platform
	contextDir string
	df         *Dockerfile
	archive    io.Reader
	log        *Logger
}

func (a engineBuildAct) reader() (io.Reader, error) {
	switch {
	case a.archive != nil:
		return a.archive, nil
	case a.df != nil:
		return a.df.Archive(a.contextDir)
	case a.contextDir != "" || a.dockerfile != "":
		rel := a.dockerfile
		if rel == "" {
			rel = "Dockerfile"
		}
		path := rel
		if a.contextDir != "" && !filepath.IsAbs(rel) {
			path = filepath.Join(a.contextDir, rel)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read dockerfile %s: %w", path, err)
		}
		return packBuildContext(data, a.contextDir)
	default:
		return nil, cfgErr("engine", "build", "BuildImage", fmt.Errorf("%w: dockerfile or tar context", errRequired))
	}
}

func (a engineBuildAct) sdk() clienttypes.ImageBuildOptions {
	dockerfile := a.dockerfile
	if dockerfile == "" && a.df != nil {
		dockerfile = "Dockerfile"
	}
	return clienttypes.ImageBuildOptions{
		Tags:           append([]string(nil), a.tags...),
		SuppressOutput: a.quiet,
		NoCache:        a.noCache,
		Remove:         a.remove,
		ForceRemove:    a.forceRm,
		PullParent:     a.pull,
		Memory:         a.memory,
		Dockerfile:     dockerfile,
		Target:         a.target,
		Labels:         a.labels,
		BuildArgs:      a.args,
		Platforms:      append([]ocispec.Platform(nil), a.platforms...),
		// BuildKit accepts RUN --mount and COPY --exclude. The classic builder rejects them.
		Version: buildapi.BuilderBuildKit,
	}
}

type engineNetworkAct struct {
	driver     string
	internal   bool
	attachable bool
	labels     map[string]string
	options    map[string]string
}

func (a engineNetworkAct) sdk() clienttypes.NetworkCreateOptions {
	return clienttypes.NetworkCreateOptions{
		Driver:     a.driver,
		Internal:   a.internal,
		Attachable: a.attachable,
		Labels:     a.labels,
		Options:    a.options,
	}
}

type engineVolumeAct struct {
	name        string
	driver      string
	labels      map[string]string
	DriverValue map[string]string
}

func (a engineVolumeAct) sdk() clienttypes.VolumeCreateOptions {
	return clienttypes.VolumeCreateOptions{
		Name:       a.name,
		Driver:     a.driver,
		Labels:     a.labels,
		DriverOpts: a.DriverValue,
	}
}

func durPtr(d time.Duration) *time.Duration { return &d }

func secPtr(d time.Duration) *int {
	s := int(d / time.Second)
	return &s
}
