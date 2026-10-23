package containerkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
)

type dockerFile struct {
	builder      strings.Builder
	errs         []error
	lastCmdIsRun bool
	cmdSet       bool
	bodyStarted  bool
	syntaxSet    bool
	escapeSet    bool
	checkSet     bool
}

// CopyOpts are optional flags on a COPY instruction.
type CopyOpts struct {
	From    string
	Chown   string
	Chmod   string
	Link    bool
	Exclude []string
	Parents bool
}

// AddOpts are optional flags on an ADD instruction.
type AddOpts struct {
	Chown      string
	Chmod      string
	Checksum   string
	KeepGitDir bool
	Link       bool
	Exclude    []string
}

// RunOpts are BuildKit flags on a RUN instruction.
// Setting any flag starts a new RUN. Later Run calls with empty options still
// fold into that instruction.
type RunOpts struct {
	Mount    []string
	Network  string
	Security string
}

// NewDockerFile creates a dockerfile builder.
//
// Not safe for concurrent use.
func NewDockerFile() *dockerFile {
	return &dockerFile{}
}

// Syntax sets the # syntax= parser directive.
func (d *dockerFile) Syntax(frontend string) *dockerFile {
	return d.directive("syntax", frontend, &d.syntaxSet)
}

// Escape sets the # escape= parser directive. char must be \ or `.
func (d *dockerFile) Escape(char string) *dockerFile {
	if char != `\` && char != "`" {
		return d.fail(fmt.Errorf("escape directive must be \\ or `"))
	}
	return d.directive("escape", char, &d.escapeSet)
}

// Check sets the # check= parser directive.
func (d *dockerFile) Check(value string) *dockerFile {
	return d.directive("check", value, &d.checkSet)
}

func (d *dockerFile) directive(name, value string, seen *bool) *dockerFile {
	if d.bodyStarted {
		return d.fail(fmt.Errorf("%s directive must precede instructions", name))
	}
	if *seen {
		return d.fail(fmt.Errorf("%s directive is already set", name))
	}
	if strings.TrimSpace(value) == "" {
		return d.fail(fmt.Errorf("%s directive is empty", name))
	}
	*seen = true
	d.builder.WriteString("# " + name + "=" + value + "\n")
	return d
}

// From sets the FROM instruction. An empty tag omits ":tag".
func (d *dockerFile) From(image string, tag string) *dockerFile {
	return d.from("", image, tag, "")
}

// FromPlatform sets FROM --platform=platform.
func (d *dockerFile) FromPlatform(platform, image, tag string) *dockerFile {
	if strings.TrimSpace(platform) == "" {
		return d.fail(fmt.Errorf("from platform is empty"))
	}
	return d.from(platform, image, tag, "")
}

// FromAs sets FROM image AS alias. image may already include a tag.
func (d *dockerFile) FromAs(image, alias string) *dockerFile {
	if strings.TrimSpace(alias) == "" {
		return d.fail(fmt.Errorf("from alias is empty"))
	}
	return d.from("", image, "", alias)
}

// FromAsPlatform sets FROM --platform=platform image AS alias.
func (d *dockerFile) FromAsPlatform(platform, image, alias string) *dockerFile {
	if strings.TrimSpace(platform) == "" {
		return d.fail(fmt.Errorf("from platform is empty"))
	}
	if strings.TrimSpace(alias) == "" {
		return d.fail(fmt.Errorf("from alias is empty"))
	}
	return d.from(platform, image, "", alias)
}

func (d *dockerFile) from(platform, image, tag, alias string) *dockerFile {
	image = strings.TrimSpace(image)
	tag = strings.TrimSpace(tag)
	alias = strings.TrimSpace(alias)
	platform = strings.TrimSpace(platform)
	if image == "" {
		return d.fail(fmt.Errorf("from image is empty"))
	}
	if tag != "" && strings.Contains(image, ":") {
		return d.fail(fmt.Errorf("from tag must be passed separately from the image"))
	}
	for _, part := range []string{platform, image, tag, alias} {
		if strings.ContainsAny(part, " \t\r\n") {
			return d.fail(fmt.Errorf("from contains whitespace"))
		}
	}
	var b strings.Builder
	b.WriteString("FROM ")
	if platform != "" {
		b.WriteString("--platform=")
		b.WriteString(platform)
		b.WriteByte(' ')
	}
	b.WriteString(image)
	if tag != "" {
		b.WriteByte(':')
		b.WriteString(tag)
	}
	if alias != "" {
		b.WriteString(" AS ")
		b.WriteString(alias)
	}
	b.WriteByte('\n')
	d.commit(b.String(), false)
	return d
}

// Arg sets an ARG instruction without a default.
func (d *dockerFile) Arg(arg string) *dockerFile {
	if strings.TrimSpace(arg) == "" {
		return d.fail(fmt.Errorf("arg name is empty"))
	}
	d.commit("ARG "+arg+"\n", false)
	return d
}

// ArgKV sets an ARG instruction with a default value.
func (d *dockerFile) ArgKV(key string, value string) *dockerFile {
	if strings.TrimSpace(key) == "" {
		return d.fail(fmt.Errorf("arg name is empty"))
	}
	d.commit(fmt.Sprintf("ARG %s=%s\n", key, value), false)
	return d
}

// Env sets an ENV instruction.
func (d *dockerFile) Env(key string, value string) *dockerFile {
	if strings.TrimSpace(key) == "" {
		return d.fail(fmt.Errorf("env name is empty"))
	}
	d.commit(fmt.Sprintf("ENV %s=%s\n", key, value), false)
	return d
}

// Copy sets a COPY instruction. At most one CopyOpts value is accepted.
func (d *dockerFile) Copy(src string, dest string, opts ...CopyOpts) *dockerFile {
	o, err := oneOpt("copy", opts)
	if err != nil {
		return d.fail(err)
	}
	flags, err := copyFlags(o)
	if err != nil {
		return d.fail(err)
	}
	line, err := fileInstruction("COPY", flags, src, dest)
	if err != nil {
		return d.fail(err)
	}
	d.commit(line, false)
	return d
}

// Add sets an ADD instruction. At most one AddOpts value is accepted.
func (d *dockerFile) Add(src string, dest string, opts ...AddOpts) *dockerFile {
	o, err := oneOpt("add", opts)
	if err != nil {
		return d.fail(err)
	}
	flags, err := addFlags(o)
	if err != nil {
		return d.fail(err)
	}
	line, err := fileInstruction("ADD", flags, src, dest)
	if err != nil {
		return d.fail(err)
	}
	d.commit(line, false)
	return d
}

func copyFlags(o CopyOpts) (string, error) {
	var b strings.Builder
	if err := writeFlag(&b, "--from", o.From); err != nil {
		return "", err
	}
	if err := writeFlag(&b, "--chown", o.Chown); err != nil {
		return "", err
	}
	if err := writeFlag(&b, "--chmod", o.Chmod); err != nil {
		return "", err
	}
	for _, pattern := range o.Exclude {
		if pattern == "" {
			return "", fmt.Errorf("--exclude is empty")
		}
		if err := writeFlag(&b, "--exclude", pattern); err != nil {
			return "", err
		}
	}
	if o.Parents {
		b.WriteString("--parents ")
	}
	if o.Link {
		b.WriteString("--link ")
	}
	return b.String(), nil
}

func addFlags(o AddOpts) (string, error) {
	var b strings.Builder
	if err := writeFlag(&b, "--chown", o.Chown); err != nil {
		return "", err
	}
	if err := writeFlag(&b, "--chmod", o.Chmod); err != nil {
		return "", err
	}
	if err := writeFlag(&b, "--checksum", o.Checksum); err != nil {
		return "", err
	}
	if o.KeepGitDir {
		b.WriteString("--keep-git-dir=true ")
	}
	for _, pattern := range o.Exclude {
		if pattern == "" {
			return "", fmt.Errorf("--exclude is empty")
		}
		if err := writeFlag(&b, "--exclude", pattern); err != nil {
			return "", err
		}
	}
	if o.Link {
		b.WriteString("--link ")
	}
	return b.String(), nil
}

// Entrypoint sets the JSON exec form of ENTRYPOINT.
func (d *dockerFile) Entrypoint(executable string, args ...string) *dockerFile {
	line, err := execInstruction("ENTRYPOINT", executable, args)
	if err != nil {
		return d.fail(err)
	}
	d.commit(line, false)
	return d
}

// EntrypointShell sets the shell form of ENTRYPOINT.
func (d *dockerFile) EntrypointShell(executable string, args ...string) *dockerFile {
	line, err := shellInstruction("ENTRYPOINT", executable, args)
	if err != nil {
		return d.fail(err)
	}
	d.commit(line, false)
	return d
}

// Expose sets an EXPOSE instruction.
func (d *dockerFile) Expose(port string) *dockerFile {
	if strings.TrimSpace(port) == "" {
		return d.fail(fmt.Errorf("expose port is empty"))
	}
	d.commit("EXPOSE "+port+"\n", false)
	return d
}

// Label sets a LABEL instruction.
func (d *dockerFile) Label(key string, value string) *dockerFile {
	if strings.TrimSpace(key) == "" {
		return d.fail(fmt.Errorf("label name is empty"))
	}
	d.commit(fmt.Sprintf("LABEL %s=%s\n", key, value), false)
	return d
}

// Onbuild sets an ONBUILD instruction.
func (d *dockerFile) Onbuild(cmd string) *dockerFile {
	if strings.TrimSpace(cmd) == "" {
		return d.fail(fmt.Errorf("onbuild command is empty"))
	}
	d.commit("ONBUILD "+cmd+"\n", false)
	return d
}

// Workdir sets a WORKDIR instruction.
func (d *dockerFile) Workdir(path string) *dockerFile {
	if strings.TrimSpace(path) == "" {
		return d.fail(fmt.Errorf("workdir path is empty"))
	}
	d.commit("WORKDIR "+path+"\n", false)
	return d
}

// StopSignal sets a STOPSIGNAL instruction.
func (d *dockerFile) StopSignal(signal string) *dockerFile {
	if strings.TrimSpace(signal) == "" {
		return d.fail(fmt.Errorf("stopsignal is empty"))
	}
	d.commit("STOPSIGNAL "+signal+"\n", false)
	return d
}

// User sets a USER instruction.
func (d *dockerFile) User(user string) *dockerFile {
	if strings.TrimSpace(user) == "" {
		return d.fail(fmt.Errorf("user is empty"))
	}
	d.commit("USER "+user+"\n", false)
	return d
}

// Comment writes a dockerfile comment.
func (d *dockerFile) Comment(comment string) *dockerFile {
	d.commit("# "+comment+"\n", false)
	return d
}

// Volumes sets a VOLUME instruction.
func (d *dockerFile) Volumes(volumes ...string) *dockerFile {
	if len(volumes) == 0 {
		return d.fail(fmt.Errorf("volume requires at least one path"))
	}
	for _, volume := range volumes {
		if strings.TrimSpace(volume) == "" {
			return d.fail(fmt.Errorf("volume path is empty"))
		}
	}
	arr, err := jsonArray(volumes)
	if err != nil {
		return d.fail(err)
	}
	d.commit("VOLUME "+arr+"\n", false)
	return d
}

// Healthcheck sets a HEALTHCHECK instruction from a container Health value.
// Zero interval, timeout, start period, and retries are omitted.
func (d *dockerFile) Healthcheck(h Health) *dockerFile {
	if len(h.Test) > 0 && h.Test[0] == "NONE" {
		return d.HealthcheckNone()
	}
	c := NewContainer().HealthCheck(h)
	d.errs = append(d.errs, c.Errors...)
	if len(c.Errors) > 0 {
		return d
	}
	hc := container.HealthConfig{}
	if c.Config != nil && c.Config.Container != nil && c.Config.Container.Healthcheck != nil {
		hc = *c.Config.Container.Healthcheck
	}
	cmd, err := formatHealthCmd(hc.Test)
	if err != nil {
		return d.fail(err)
	}
	if cmd == "NONE" {
		return d.HealthcheckNone()
	}
	var flags []string
	if hc.Interval > 0 {
		flags = append(flags, "--interval="+hc.Interval.String())
	}
	if hc.Timeout > 0 {
		flags = append(flags, "--timeout="+hc.Timeout.String())
	}
	if hc.StartPeriod > 0 {
		flags = append(flags, "--start-period="+hc.StartPeriod.String())
	}
	if hc.Retries > 0 {
		flags = append(flags, fmt.Sprintf("--retries=%d", hc.Retries))
	}
	line := "HEALTHCHECK "
	if len(flags) > 0 {
		line += strings.Join(flags, " ") + " \\\n\t"
	}
	d.commit(line+cmd+"\n", false)
	return d
}

// HealthcheckNone sets HEALTHCHECK NONE.
func (d *dockerFile) HealthcheckNone() *dockerFile {
	d.commit("HEALTHCHECK NONE\n", false)
	return d
}

// Shell sets the SHELL instruction (JSON exec form).
func (d *dockerFile) Shell(executable string, args ...string) *dockerFile {
	line, err := execInstruction("SHELL", executable, args)
	if err != nil {
		return d.fail(err)
	}
	d.commit(line, false)
	return d
}

// CommandExec sets CMD in JSON exec form. A second CMD records an error and is not written.
func (d *dockerFile) CommandExec(executable string, args ...string) *dockerFile {
	if d.cmdSet {
		return d.fail(fmt.Errorf("command has already been set"))
	}
	line, err := execInstruction("CMD", executable, args)
	if err != nil {
		return d.fail(err)
	}
	d.cmdSet = true
	d.commit(line, false)
	return d
}

// CommandShell sets CMD in shell form. A second CMD records an error and is not written.
func (d *dockerFile) CommandShell(executable string, args ...string) *dockerFile {
	if d.cmdSet {
		return d.fail(fmt.Errorf("command has already been set"))
	}
	line, err := shellInstruction("CMD", executable, args)
	if err != nil {
		return d.fail(err)
	}
	d.cmdSet = true
	d.commit(line, false)
	return d
}

// Run sets a RUN instruction.
// Consecutive Run calls with no flags are joined with "&& \" into one layer.
// A call that sets RunOpts starts a new RUN.
func (d *dockerFile) Run(cmd string, opts ...RunOpts) *dockerFile {
	return d.run(cmd, false, opts...)
}

// RunLayer starts a new RUN instruction. Unlike Run, it does not fold into the previous RUN.
func (d *dockerFile) RunLayer(cmd string, opts ...RunOpts) *dockerFile {
	return d.run(cmd, true, opts...)
}

func (d *dockerFile) run(cmd string, fresh bool, opts ...RunOpts) *dockerFile {
	o, err := oneOpt("run", opts)
	if err != nil {
		return d.fail(err)
	}
	if strings.TrimSpace(cmd) == "" {
		return d.fail(fmt.Errorf("run command is empty"))
	}
	flags, flagged, err := runFlags(o)
	if err != nil {
		return d.fail(err)
	}
	if d.lastCmdIsRun && !fresh && !flagged {
		nStr := strings.TrimSuffix(d.builder.String(), "\n")
		d.builder.Reset()
		d.builder.WriteString(fmt.Sprintf("%s && \\\n\t%s\n", nStr, cmd))
		d.lastCmdIsRun = true
		d.bodyStarted = true
		return d
	}
	d.commit("RUN "+flags+cmd+"\n", true)
	return d
}

// RunArgs continues the previous RUN with newline continuation and no "&&".
func (d *dockerFile) RunArgs(args ...string) *dockerFile {
	if !d.lastCmdIsRun {
		return d.fail(fmt.Errorf("runargs was called but no run command was called before it. Args: %s", strings.Join(args, ",")))
	}
	if len(args) == 0 {
		return d.fail(fmt.Errorf("runargs requires at least one argument"))
	}
	for _, arg := range args {
		if strings.TrimSpace(arg) == "" {
			return d.fail(fmt.Errorf("runargs argument is empty"))
		}
	}
	nStr := strings.TrimSuffix(d.builder.String(), "\n")
	d.builder.Reset()
	d.builder.WriteString(fmt.Sprintf("%s \\\n\t%s\n", nStr, strings.Join(args, " \\\n\t")))
	d.lastCmdIsRun = true
	d.bodyStarted = true
	return d
}

// Format formats the dockerfile with the given arguments and replaces its contents.
func (d *dockerFile) Format(args ...any) *dockerFile {
	formatted := fmt.Sprintf(d.builder.String(), args...)
	d.builder.Reset()
	d.builder.WriteString(formatted)
	d.lastCmdIsRun = false
	if formatted != "" {
		d.bodyStarted = true
	}
	return d
}

// String returns the dockerfile text.
func (d *dockerFile) String() string {
	return d.builder.String()
}

// Export writes the dockerfile to path after validation.
func (d *dockerFile) Export(path string, mode os.FileMode) error {
	if err := d.Validate(); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(d.String()), mode)
}

// Validate returns any errors recorded while building the dockerfile.
func (d *dockerFile) Validate() error {
	if len(d.errs) > 0 {
		return errors.Join(d.errs...)
	}
	return nil
}

// WithInline returns a BuildSpec that sets the service dockerfile inline.
// When the dockerfile is invalid, Project.WithService records the error and
// does not apply the spec.
func (d *dockerFile) WithInline() BuildSpec {
	spec := BuildSpec{DockerfileInline: d.String()}
	if err := d.Validate(); err != nil {
		spec.err = err
	}
	return spec
}

func (d *dockerFile) fail(err error) *dockerFile {
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

func (d *dockerFile) commit(line string, run bool) {
	d.bodyStarted = true
	d.lastCmdIsRun = run
	d.builder.WriteString(line)
}

func oneOpt[T any](kind string, opts []T) (T, error) {
	var zero T
	switch len(opts) {
	case 0:
		return zero, nil
	case 1:
		return opts[0], nil
	default:
		return zero, fmt.Errorf("%s accepts at most one options value", kind)
	}
}

func writeFlag(b *strings.Builder, name, value string) error {
	if value == "" {
		return nil
	}
	if strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("%s contains whitespace", name)
	}
	b.WriteString(name)
	b.WriteByte('=')
	b.WriteString(value)
	b.WriteByte(' ')
	return nil
}

func runFlags(o RunOpts) (string, bool, error) {
	var b strings.Builder
	for _, mount := range o.Mount {
		if mount == "" {
			return "", false, fmt.Errorf("--mount is empty")
		}
		if err := writeFlag(&b, "--mount", mount); err != nil {
			return "", false, err
		}
	}
	if err := writeFlag(&b, "--network", o.Network); err != nil {
		return "", false, err
	}
	if err := writeFlag(&b, "--security", o.Security); err != nil {
		return "", false, err
	}
	flags := b.String()
	return flags, flags != "", nil
}

func fileInstruction(instr, flags, src, dest string) (string, error) {
	src = strings.TrimSpace(src)
	dest = strings.TrimSpace(dest)
	if src == "" || dest == "" {
		return "", fmt.Errorf("%s source and destination are required", strings.ToLower(instr))
	}
	if strings.Contains(src, " ") || strings.Contains(dest, " ") {
		arr, err := jsonArray([]string{src, dest})
		if err != nil {
			return "", err
		}
		return instr + " " + flags + arr + "\n", nil
	}
	return instr + " " + flags + src + " " + dest + "\n", nil
}

func execInstruction(instr, executable string, args []string) (string, error) {
	if executable == "" {
		return "", fmt.Errorf("%s executable is empty", strings.ToLower(instr))
	}
	arr, err := jsonArray(append([]string{executable}, args...))
	if err != nil {
		return "", err
	}
	return instr + " " + arr + "\n", nil
}

func shellInstruction(instr, executable string, args []string) (string, error) {
	if strings.TrimSpace(executable) == "" {
		return "", fmt.Errorf("%s executable is empty", strings.ToLower(instr))
	}
	line := instr + " " + executable
	if len(args) > 0 {
		line += " " + strings.Join(args, " ")
	}
	return line + "\n", nil
}

func jsonArray(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("json array is empty")
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		encoded, err := json.Marshal(arg)
		if err != nil {
			return "", err
		}
		parts[i] = string(encoded)
	}
	return "[" + strings.Join(parts, ", ") + "]", nil
}

func formatHealthCmd(test []string) (string, error) {
	if len(test) == 0 {
		return "", fmt.Errorf("healthcheck test is empty")
	}
	switch test[0] {
	case "NONE":
		return "NONE", nil
	case "CMD-SHELL":
		if len(test) != 2 || strings.TrimSpace(test[1]) == "" {
			return "", fmt.Errorf("healthcheck shell command is empty")
		}
		return "CMD " + test[1], nil
	case "CMD":
		if len(test) < 2 || strings.TrimSpace(test[1]) == "" {
			return "", fmt.Errorf("healthcheck command is empty")
		}
		arr, err := jsonArray(test[1:])
		if err != nil {
			return "", err
		}
		return "CMD " + arr, nil
	default:
		return "", fmt.Errorf("healthcheck test must start with CMD, CMD-SHELL, or NONE")
	}
}
