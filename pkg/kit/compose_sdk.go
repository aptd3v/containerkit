package kit

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/distribution/reference"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/compose/v5/cmd/display"
	"github.com/docker/compose/v5/cmd/formatter"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
	"github.com/docker/compose/v5/pkg/dryrun"
	"github.com/docker/go-units"
	"github.com/moby/moby/api/types/image"
	moby "github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/stringid"
	"golang.org/x/term"
)

// selectMode decides how a service list narrows the project model.
// The Compose SDK applies an operation to every service still in the project.
type selectMode int

const (
	selectKeep selectMode = iota
	selectIgnoreDeps
	selectIncludeDeps
)

func (c *Compose) prepare(cmd string, profiles, services []string, mode selectMode) (*types.Project, error) {
	if c == nil || c.project == nil {
		return nil, fmt.Errorf("project is required")
	}
	if err := c.project.Validate(); err != nil {
		return nil, err
	}
	if err := c.checkProfiles(cmd, profiles); err != nil {
		return nil, err
	}
	project, err := c.project.Unwrap().WithSelectedServices(nil)
	if err != nil {
		return nil, err
	}
	if len(profiles) > 0 {
		project, err = project.WithProfiles(profiles)
		if err != nil {
			return nil, err
		}
	}
	if !c.global.AllResources {
		project = project.WithoutUnnecessaryResources()
	}
	if c.global.ProjectDirectory != "" {
		project.WorkingDir = c.global.ProjectDirectory
	}
	if err := applyEnvFiles(project, c.global.EnvFile); err != nil {
		return nil, err
	}
	stampComposeLabels(project)
	if len(services) == 0 || mode == selectKeep {
		return project, nil
	}
	opt := types.IncludeDependencies
	if mode == selectIgnoreDeps {
		opt = types.IgnoreDependencies
	}
	return project.WithSelectedServices(services, opt)
}

func (c *Compose) checkProfiles(cmd string, profiles []string) error {
	if profiles != nil {
		if len(c.project.declaredProfiles()) == 0 {
			return errdefs.NewComposeError(fmt.Errorf("when using %s.Profiles, you must create a service with a profile via Service.Profiles", cmd))
		}
	}
	if len(profiles) == 0 {
		names := c.project.declaredProfiles()
		if len(names) > 0 {
			return errdefs.NewComposeError(fmt.Errorf("when you have profiles, you must specify them via %s.Profiles\nprofiles found: %s", cmd, strings.Join(names, ", ")))
		}
	}
	return nil
}

// stampComposeLabels matches compose.Service.postProcessProject. The SDK
// lists containers by these labels, including oneoff=False.
func stampComposeLabels(project *types.Project) {
	files := strings.Join(project.ComposeFiles, ",")
	for name, service := range project.Services {
		service.CustomLabels = service.CustomLabels.Add(api.ProjectLabel, project.Name).Add(api.ServiceLabel, name).Add(api.VersionLabel, api.ComposeVersion).Add(api.WorkingDirLabel, project.WorkingDir).Add(api.ConfigFilesLabel, files).Add(api.OneoffLabel, "False")
		project.Services[name] = service
	}
}

func applyEnvFiles(project *types.Project, files []string) error {
	if len(files) == 0 {
		return nil
	}
	if project.Environment == nil {
		project.Environment = types.Mapping{}
	}
	for _, file := range files {
		if file == "" {
			continue
		}
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		err = mergeEnvFile(project.Environment, f)
		_ = f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeEnvFile(dst types.Mapping, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, ok := strings.Cut(line, "=")
		if !ok || key == "" {
			continue
		}
		if _, exists := dst[key]; exists {
			continue
		}
		dst[key] = val
	}
	return scanner.Err()
}

func (c *Compose) open(out io.Writer, in io.Reader, progress string) (api.Compose, func(), error) {
	if in == nil {
		in = bytes.NewReader(nil)
	}
	out = defaultWriter(out)
	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		return nil, nil, err
	}
	if err := dockerCLI.Initialize(flags.NewClientOptions()); err != nil {
		return nil, nil, err
	}
	if progress == "" {
		progress = c.global.Progress
	}
	options := []compose.Option{
		compose.WithStreams(out, out, in),
		compose.WithPrompt(compose.AlwaysOkPrompt()),
		compose.WithEventProcessor(progressProcessor(progress, c.global.ANSI, out)),
	}
	if c.global.Parallel != nil {
		options = append(options, compose.WithMaxConcurrency(*c.global.Parallel))
	}
	if c.global.DryRun {
		options = append(options, compose.WithDryRun)
	}
	svc, err := compose.NewComposeService(dockerCLI, options...)
	if err != nil {
		return nil, nil, err
	}
	return svc, func() { closeService(svc) }, nil
}

func (c *Compose) engineCLI(out io.Writer, _ io.Reader) (command.Cli, func(), error) {
	out = defaultWriter(out)
	base, err := command.NewDockerCli(command.WithOutputStream(out), command.WithErrorStream(out))
	if err != nil {
		return nil, nil, err
	}
	if err := base.Initialize(flags.NewClientOptions()); err != nil {
		return nil, nil, err
	}
	closeFn := func() { _ = base.Client().Close() }
	if !c.global.DryRun {
		return base, closeFn, nil
	}
	dry, err := command.NewDockerCli(command.WithOutputStream(out), command.WithErrorStream(out))
	if err != nil {
		closeFn()
		return nil, nil, err
	}
	err = dry.Initialize(flags.NewClientOptions(), command.WithInitializeClient(func(*command.DockerCli) (moby.APIClient, error) {
		return dryrun.NewDryRunClient(base.Client(), base)
	}))
	if err != nil {
		closeFn()
		return nil, nil, err
	}
	return dry, closeFn, nil
}

func vizIndentation(size *int, spaces bool) (string, error) {
	n := 1
	if size != nil {
		n = *size
	}
	if n < 0 {
		return "", fmt.Errorf("invalid indentation size: %d", n)
	}
	unit := "\t"
	if spaces {
		unit = " "
	}
	return strings.Repeat(unit, n), nil
}

func writeTransformers(w io.Writer, images []image.Summary, format string, quiet bool) error {
	if quiet {
		for _, img := range images {
			name := img.ID
			if len(img.RepoTags) > 0 {
				name = img.RepoTags[0]
			}
			if _, err := fmt.Fprintln(w, name); err != nil {
				return err
			}
		}
		return nil
	}
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		return enc.Encode(images)
	}
	if _, err := fmt.Fprintln(w, "IMAGE ID\tREPO\tTAG\tSIZE"); err != nil {
		return err
	}
	for _, img := range images {
		repo, tag := "<none>", "<none>"
		if len(img.RepoTags) > 0 {
			ref, err := reference.ParseDockerRef(img.RepoTags[0])
			if err == nil {
				repo = reference.FamiliarName(ref)
				if tagged, ok := ref.(reference.Tagged); ok {
					tag = tagged.Tag()
				}
			}
		}
		_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			stringid.TruncateID(img.ID), repo, tag, units.HumanSizeWithPrecision(float64(img.Size), 3))
		if err != nil {
			return err
		}
	}
	return nil
}

func closeService(svc api.Compose) {
	type closer interface{ Close() error }
	if c, ok := svc.(closer); ok {
		_ = c.Close()
	}
}

func progressProcessor(progress, ansi string, out io.Writer) api.EventProcessor {
	switch progress {
	case "quiet":
		return display.Quiet()
	case "json":
		return display.JSON(out)
	case "plain":
		return display.Plain(out)
	case "tty":
		if ansi == "never" {
			return display.Plain(out)
		}
		return display.Full(out, out, false)
	default:
		if ansi == "never" || !isTerminal(out) {
			return display.Plain(out)
		}
		return display.Full(out, out, false)
	}
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func seconds(n *int) *time.Duration {
	if n == nil {
		return nil
	}
	d := time.Duration(*n) * time.Second
	return &d
}

func exitStatus(code int) error {
	if code == 0 {
		return nil
	}
	return fmt.Errorf("exit status %d", code)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func indexOrZero(n *int) int {
	if n == nil {
		return 0
	}
	return *n
}

func applyPullPolicy(project *types.Project, policy string) {
	if policy == "" {
		return
	}
	for name, service := range project.Services {
		service.PullPolicy = policy
		project.Services[name] = service
	}
}

func applyScale(project *types.Project, scales []UpScale) {
	for _, scale := range scales {
		service, ok := project.Services[scale.Service]
		if !ok {
			continue
		}
		n := scale.Num
		service.Scale = &n
		project.Services[scale.Service] = service
	}
}

func buildArgs(values []string) types.MappingWithEquals {
	if len(values) == 0 {
		return nil
	}
	args := types.MappingWithEquals{}
	for _, value := range values {
		key, raw, ok := strings.Cut(value, "=")
		if !ok {
			args[value] = nil
			continue
		}
		val := raw
		args[key] = &val
	}
	return args
}

func sshKeys(values []string) []types.SSHKey {
	if len(values) == 0 {
		return nil
	}
	keys := make([]types.SSHKey, 0, len(values))
	for _, value := range values {
		id, path, ok := strings.Cut(value, "=")
		if !ok {
			keys = append(keys, types.SSHKey{ID: value})
			continue
		}
		keys = append(keys, types.SSHKey{ID: id, Path: path})
	}
	return keys
}

func (opt *ComposeBuild) apiBuild(services []string) (api.BuildOptions, error) {
	var memory int64
	if opt.Memory != "" {
		n, err := units.RAMInBytes(opt.Memory)
		if err != nil {
			return api.BuildOptions{}, err
		}
		memory = n
	}
	provenance, sbom := "", ""
	if opt.Provenance {
		provenance = "true"
	}
	if opt.SBOM {
		sbom = "true"
	}
	return api.BuildOptions{
		Pull:       opt.Pull,
		Push:       opt.Push,
		Args:       buildArgs(opt.BuildArg),
		NoCache:    opt.NoCache,
		Quiet:      opt.Quiet,
		Services:   services,
		Deps:       opt.WithDependencies,
		SSHs:       sshKeys(opt.SSH),
		Memory:     memory,
		Builder:    opt.Builder,
		Print:      opt.Print,
		Check:      opt.Check,
		Provenance: provenance,
		SBOM:       sbom,
	}, nil
}

func upBuildOptions(opt *Up, services []string) *api.BuildOptions {
	if opt.NoBuild {
		return nil
	}
	return &api.BuildOptions{
		Quiet:    opt.QuietPull,
		Services: services,
		Deps:     !opt.NoDeps,
	}
}

func recreate(force, never, renew bool) (string, string) {
	switch {
	case never:
		return api.RecreateNever, api.RecreateNever
	case force || renew:
		return api.RecreateForce, api.RecreateDiverged
	default:
		return api.RecreateDiverged, api.RecreateDiverged
	}
}

const composeIgnoreOrphans = "COMPOSE_IGNORE_ORPHANS"

func orphanMode(explicit, remove bool, env types.Mapping) (bool, error) {
	ignore := explicit || composeBool(lookupComposeEnv(env, composeIgnoreOrphans))
	if ignore && remove {
		return false, fmt.Errorf("cannot combine COMPOSE_IGNORE_ORPHANS and --remove-orphans")
	}
	return ignore, nil
}

func lookupComposeEnv(env types.Mapping, key string) string {
	if env != nil {
		if value, ok := env[key]; ok {
			return value
		}
	}
	return os.Getenv(key)
}

func composeBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "y" {
		return true
	}
	parsed, _ := strconv.ParseBool(value)
	return parsed
}

func (c *Compose) projectLoadOptions(opt *ProjectLoad) api.ProjectLoadOptions {
	workingDir := opt.WorkingDir
	if workingDir == "" {
		workingDir = c.global.ProjectDirectory
	}
	envFiles := append(append([]string{}, c.global.EnvFile...), opt.EnvFiles...)
	load := api.ProjectLoadOptions{
		ProjectName:   opt.Name,
		ConfigPaths:   append([]string{}, opt.Files...),
		WorkingDir:    workingDir,
		EnvFiles:      envFiles,
		Profiles:      append([]string{}, opt.Profiles...),
		Services:      append([]string{}, opt.Services...),
		Offline:       opt.Offline,
		All:           opt.AllResources || c.global.AllResources,
		Compatibility: opt.Compatibility || c.global.Compatibility,
		OCI:           api.OCIOptions{InsecureRegistries: append([]string{}, opt.InsecureRegistries...)},
	}
	if opt.OnUnsupported != nil {
		load.OnUnsupportedAttribute = func(found []api.UnsupportedAttribute) {
			out := make([]UnsupportedAttribute, len(found))
			for i, attr := range found {
				out[i] = UnsupportedAttribute{Service: attr.Service, Path: attr.Path, Reason: attr.Reason}
			}
			opt.OnUnsupported(out)
		}
	}
	return load
}

func (opt *Up) createOptions(services []string, ignoreOrphans bool) api.CreateOptions {
	recreate, deps := recreate(opt.ForceRecreate, opt.NoRecreate, opt.RenewAnonVolumes)
	if opt.AlwaysRecreateDeps && !opt.NoRecreate {
		deps = api.RecreateForce
	}
	return api.CreateOptions{
		Build:                upBuildOptions(opt, services),
		Services:             services,
		RemoveOrphans:        opt.RemoveOrphans,
		IgnoreOrphans:        ignoreOrphans,
		Recreate:             recreate,
		RecreateDependencies: deps,
		Inherit:              !opt.RenewAnonVolumes,
		Timeout:              seconds(opt.Timeout),
		QuietPull:            opt.QuietPull,
	}
}

func (opt *Up) onExit() api.Cascade {
	switch {
	case opt.AbortOnContainerExit:
		return api.CascadeStop
	case opt.AbortOnContainerFailure:
		return api.CascadeFail
	default:
		return api.CascadeIgnore
	}
}

func (opt *Up) attachTo(names []string) []string {
	if opt.Detach || opt.Wait {
		return nil
	}
	if opt.Attach != nil {
		return []string{*opt.Attach}
	}
	attach := append([]string{}, names...)
	if opt.NoAttach != nil {
		filtered := attach[:0]
		for _, name := range attach {
			if name != *opt.NoAttach {
				filtered = append(filtered, name)
			}
		}
		attach = filtered
	}
	return attach
}

func logConsumer(ctx context.Context, w io.Writer, color, prefix, timestamps bool) api.LogConsumer {
	return formatter.NewLogConsumer(ctx, w, w, color, prefix, timestamps)
}

func writeLines(w io.Writer, lines []string) error {
	sort.Strings(lines)
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compose) renderConfig(w io.Writer, project *types.Project, opt *ComposeConfig) error {
	if opt.Quiet {
		return nil
	}
	var body []byte
	var err error
	switch {
	case opt.PrintServices:
		return writeLines(w, project.ServiceNames())
	case opt.PrintVolumes:
		return writeLines(w, mapKeys(project.Volumes))
	case opt.PrintNetworks:
		return writeLines(w, mapKeys(project.Networks))
	case opt.PrintModels:
		return writeLines(w, mapKeys(project.Models))
	case opt.PrintProfiles:
		return writeLines(w, c.project.declaredProfiles())
	case opt.PrintImages:
		var images []string
		for _, service := range project.Services {
			if service.Image != "" {
				images = append(images, service.Image)
			}
		}
		return writeLines(w, images)
	case opt.PrintEnvironment || opt.PrintVariables:
		var lines []string
		for key, val := range project.Environment {
			lines = append(lines, key+"="+val)
		}
		return writeLines(w, lines)
	case strings.EqualFold(opt.Format, "json"):
		body, err = json.Marshal(project)
		if err != nil {
			return err
		}
		body = append(body, '\n')
	default:
		body, err = project.MarshalYAML()
		if err != nil {
			return err
		}
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	if opt.Output != "" {
		return os.WriteFile(opt.Output, body, 0o644)
	}
	return nil
}

func mapKeys[M ~map[string]V, V any](m M) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func writePs(w io.Writer, rows []api.ContainerSummary, opt *Ps) error {
	rows = filterPs(rows, opt)
	switch {
	case opt.Services:
		names := make([]string, 0, len(rows))
		seen := map[string]struct{}{}
		for _, row := range rows {
			if _, ok := seen[row.Service]; ok {
				continue
			}
			seen[row.Service] = struct{}{}
			names = append(names, row.Service)
		}
		return writeLines(w, names)
	case opt.Quiet:
		for _, row := range rows {
			id := row.ID
			if !opt.NoTrunc && len(id) > 12 {
				id = id[:12]
			}
			if _, err := fmt.Fprintln(w, id); err != nil {
				return err
			}
		}
		return nil
	case strings.EqualFold(opt.Format, "json"):
		body, err := json.Marshal(rows)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(body))
		return err
	default:
		for _, row := range rows {
			name := row.Name
			if !opt.NoTrunc && len(name) > 40 {
				name = name[:40]
			}
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", row.ID, name, row.Service, row.State); err != nil {
				return err
			}
		}
		return nil
	}
}

func filterPs(rows []api.ContainerSummary, opt *Ps) []api.ContainerSummary {
	if opt.Status == "" && len(opt.Filter) == 0 {
		return rows
	}
	want := map[string]struct{}{}
	if opt.Status != "" {
		want[strings.ToLower(opt.Status)] = struct{}{}
	}
	for _, filter := range opt.Filter {
		key, val, ok := strings.Cut(filter, "=")
		if ok && key == "status" {
			want[strings.ToLower(val)] = struct{}{}
		}
	}
	if len(want) == 0 {
		return rows
	}
	out := make([]api.ContainerSummary, 0, len(rows))
	for _, row := range rows {
		state := strings.ToLower(string(row.State))
		if _, ok := want[state]; ok {
			out = append(out, row)
		}
	}
	return out
}

func writeImages(w io.Writer, images map[string]api.ImageSummary, format string, quiet bool) error {
	names := mapKeys(images)
	sort.Strings(names)
	if strings.EqualFold(format, "json") {
		body, err := json.Marshal(images)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(body))
		return err
	}
	for _, name := range names {
		image := images[name]
		if quiet {
			if _, err := fmt.Fprintln(w, image.ID); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", name, image.Repository, image.Tag); err != nil {
			return err
		}
	}
	return nil
}

func writeTop(w io.Writer, rows []api.ContainerProcSummary) error {
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", row.Service, row.Name); err != nil {
			return err
		}
		if len(row.Titles) > 0 {
			if _, err := fmt.Fprintln(w, strings.Join(row.Titles, "\t")); err != nil {
				return err
			}
		}
		for _, proc := range row.Processes {
			if _, err := fmt.Fprintln(w, strings.Join(proc, "\t")); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeStacks(w io.Writer, stacks []api.Stack, opt *Ls) error {
	filtered := make([]api.Stack, 0, len(stacks))
	for _, stack := range stacks {
		if stackMatches(stack, opt.Filter) {
			filtered = append(filtered, stack)
		}
	}
	if strings.EqualFold(opt.Format, "json") {
		body, err := json.Marshal(filtered)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(body))
		return err
	}
	for _, stack := range filtered {
		if opt.Quiet {
			if _, err := fmt.Fprintln(w, stack.Name); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\n", stack.Name, stack.Status); err != nil {
			return err
		}
	}
	return nil
}

func stackMatches(stack api.Stack, filters []string) bool {
	for _, filter := range filters {
		key, val, ok := strings.Cut(filter, "=")
		if !ok {
			continue
		}
		switch key {
		case "name":
			if stack.Name != val {
				return false
			}
		case "status":
			if !strings.EqualFold(stack.Status, val) {
				return false
			}
		}
	}
	return true
}

func writePublishers(w io.Writer, publishers api.PortPublishers, protocol string) error {
	for _, publisher := range publishers {
		if protocol != "" && !strings.EqualFold(publisher.Protocol, protocol) {
			continue
		}
		if _, err := fmt.Fprintln(w, publisher.HostPort()); err != nil {
			return err
		}
	}
	return nil
}

func commitChanges(changes []string) (opts.ListOpts, error) {
	list := opts.NewListOpts(nil)
	for _, change := range changes {
		if err := list.Set(change); err != nil {
			return opts.ListOpts{}, err
		}
	}
	return list, nil
}

func labelsFrom(values []string) (types.Labels, error) {
	if len(values) == 0 {
		return nil, nil
	}
	labels := types.Labels{}
	for _, value := range values {
		key, val, ok := strings.Cut(value, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid label %q", value)
		}
		labels[key] = val
	}
	return labels, nil
}

func envFileLines(path string) ([]string, error) {
	mapping := types.Mapping{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := mergeEnvFile(mapping, f); err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(mapping))
	for key, val := range mapping {
		lines = append(lines, key+"="+val)
	}
	return lines, nil
}

func composeModuleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, mod := range info.Deps {
		if mod.Path == "github.com/docker/compose/v5" {
			if mod.Replace != nil && mod.Replace.Version != "" {
				return mod.Replace.Version
			}
			return mod.Version
		}
	}
	return "unknown"
}

func (c *Compose) containerStats(ctx context.Context, project *types.Project, opt *Stats, w io.Writer) error {
	eng, err := moby.New(moby.FromEnv)
	if err != nil {
		return err
	}
	defer eng.Close()
	filters := make(moby.Filters).Add("label", api.ProjectLabel+"="+project.Name)
	listed, err := eng.ContainerList(ctx, moby.ContainerListOptions{All: opt.All, Filters: filters})
	if err != nil {
		return err
	}
	for _, ctr := range listed.Items {
		if len(opt.ServiceNames) > 0 && !serviceSelected(ctr.Labels, opt.ServiceNames) {
			continue
		}
		stats, err := eng.ContainerStats(ctx, ctr.ID, moby.ContainerStatsOptions{Stream: !opt.NoStream})
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, stats.Body)
		closeErr := stats.Body.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if !opt.NoStream {
			return nil
		}
	}
	return nil
}

func serviceSelected(labels map[string]string, names []string) bool {
	service := labels[api.ServiceLabel]
	return slices.Contains(names, service)
}
