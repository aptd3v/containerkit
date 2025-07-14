package create

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/build"
	"github.com/docker/docker/api/types/container"
)

type dockerFile struct {
	strings.Builder
	errs         []error
	lastCmdIsRun bool
	cmdSet       bool
}

// NewDockerFile creates a new dockerfile which
// allows you to create a dockerfile string with step by step instructions via builder pattern
//
// note: Not safe for concurrent use.
func NewDockerFile() *dockerFile {
	return &dockerFile{
		Builder: strings.Builder{},
	}
}

// From sets the FROM instruction in the Dockerfile
func (d *dockerFile) From(image string, tag string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("FROM %s:%s\n", image, tag))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// FromAs sets the FROM instruction in the Dockerfile with an alias
func (d *dockerFile) FromAs(image, alias string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("FROM %s AS %s\n", image, alias))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Arg sets the ARG instruction in the Dockerfile
func (d *dockerFile) Arg(arg string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("ARG %s\n", arg))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// ArgKey sets the ARG instruction in the Dockerfile
func (d *dockerFile) ArgKV(key string, value string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("ARG %s=%s\n", key, value))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Env sets the ENV instruction in the Dockerfile
func (d *dockerFile) Env(key string, value string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("ENV %s=%s\n", key, value))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Copy sets the COPY instruction in the Dockerfile
func (d *dockerFile) Copy(src string, dest string) *dockerFile {
	defer d.setRunState(false)

	src = strings.TrimSpace(src)
	dest = strings.TrimSpace(dest)
	if strings.Contains(src, " ") || strings.Contains(dest, " ") {
		// Required for paths containing whitespace
		_, err := d.Builder.WriteString(fmt.Sprintf("COPY [\"%s\", \"%s\"]\n", src, dest))
		if err != nil {
			d.errs = append(d.errs, err)
		}
		return d
	}
	_, err := d.Builder.WriteString(fmt.Sprintf("COPY %s %s\n", src, dest))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Entrypoint sets the ENTRYPOINT instruction in the Dockerfile
func (d *dockerFile) Entrypoint(executable string, args ...string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("ENTRYPOINT [\"%s\", \"%s\"]\n", executable, strings.Join(args, "\", \"")))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Expose sets the EXPOSE instruction in the Dockerfile
func (d *dockerFile) Expose(port string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("EXPOSE %s\n", port))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Label sets the LABEL instruction in the Dockerfile
func (d *dockerFile) Label(key string, value string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("LABEL %s=%s\n", key, value))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Onbuild sets the ONBUILD instruction in the Dockerfile
func (d *dockerFile) Onbuild(cmd string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("ONBUILD %s\n", cmd))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Workdir sets the WORKDIR instruction in the Dockerfile
func (d *dockerFile) Workdir(path string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("WORKDIR %s\n", path))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Stopsignal sets the STOPSIGNAL instruction in the Dockerfile
func (d *dockerFile) StopSignal(signal string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("STOPSIGNAL %s\n", signal))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// User sets the USER instruction in the Dockerfile
func (d *dockerFile) User(user string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("USER %s\n", user))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// comment sets the comment instruction in the Dockerfile
func (d *dockerFile) Comment(comment string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("# %s\n", comment))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Volumes sets the VOLUME instruction in the Dockerfile
func (d *dockerFile) Volumes(volumes ...string) *dockerFile {
	defer d.setRunState(false)
	_, err := d.Builder.WriteString(fmt.Sprintf("VOLUME [\"%s\"]\n", strings.Join(volumes, "\", \"")))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Healthcheck sets the HEALTHCHECK instruction in the Dockerfile
func (d *dockerFile) Healthcheck(setters ...health.SetHealthcheckConfig) *dockerFile {
	defer d.setRunState(false)

	hc := container.HealthConfig{}
	for _, setter := range setters {
		if err := setter(&hc); err != nil {
			d.errs = append(d.errs, err)
		}
	}

	str := fmt.Sprintf("HEALTHCHECK --interval=%s --timeout=%s --start-period=%s --retries=%d \\\n\t%s\n",
		hc.Interval,
		hc.Timeout,
		hc.StartPeriod,
		hc.Retries,
		strings.Join(hc.Test, " "),
	)
	_, err := d.Builder.WriteString(str)
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Add sets the ADD instruction in the Dockerfile
func (d *dockerFile) Add(src string, dest string) *dockerFile {
	defer d.setRunState(false)

	src = strings.TrimSpace(src)
	dest = strings.TrimSpace(dest)
	if strings.Contains(src, " ") || strings.Contains(dest, " ") {
		// Required for paths containing whitespace
		_, err := d.Builder.WriteString(fmt.Sprintf("ADD [\"%s\", \"%s\"]\n", src, dest))
		if err != nil {
			d.errs = append(d.errs, err)
		}
		return d
	}
	_, err := d.Builder.WriteString(fmt.Sprintf("ADD %s %s\n", src, dest))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// CommandExec sets the command to be executed in the Dockerfile
// it is CMD ["executable","param1","param2"] (exec form)
func (d *dockerFile) CommandExec(executable string, args ...string) *dockerFile {
	defer d.setRunState(false)
	defer d.setCmdState(true)
	if d.cmdSet {
		d.errs = append(d.errs, fmt.Errorf("command has already been set"))
		return d
	}
	_, err := d.Builder.WriteString(fmt.Sprintf("CMD [\"%s\", \"%s\"]\n", executable, strings.Join(args, "\", \"")))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// CommandShell sets the command to be executed in the Dockerfile
// it is CMD param1 param2 (shell form)
func (d *dockerFile) CommandShell(executable string, args ...string) *dockerFile {
	defer d.setRunState(false)
	defer d.setCmdState(true)
	if d.cmdSet {
		d.errs = append(d.errs, fmt.Errorf("command has already been set"))
		return d
	}
	_, err := d.Builder.WriteString(fmt.Sprintf("CMD %s %s\n", executable, strings.Join(args, " ")))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Run sets the RUN instruction in the Dockerfile
func (d *dockerFile) Run(cmd string) *dockerFile {
	defer d.setRunState(true)
	if d.lastCmdIsRun {
		// trim last run newline to add continuation
		nStr := strings.TrimSuffix(d.Builder.String(), "\n")
		d.Builder.Reset()
		_, err := d.Builder.WriteString(fmt.Sprintf("%s && \\\n\t%s\n", nStr, cmd))
		if err != nil {
			d.errs = append(d.errs, err)
		}
		return d
	}
	_, err := d.Builder.WriteString(fmt.Sprintf("RUN %s\n", cmd))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// RunArgs sets the RUN instruction arguments in the dockerfile with newline continuation without '&&'
func (d *dockerFile) RunArgs(args ...string) *dockerFile {
	defer d.setRunState(true)

	if !d.lastCmdIsRun {
		d.errs = append(d.errs, fmt.Errorf("runargs was called but no run command was called before it. Args: %s", strings.Join(args, ",")))
		return d
	}
	// trim last run newline to add continuation
	nStr := strings.TrimSuffix(d.Builder.String(), "\n")
	d.Builder.Reset()
	_, err := d.Builder.WriteString(fmt.Sprintf("%s \\\n\t%s\n", nStr, strings.Join(args, " \\\n\t")))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// Format formats the Dockerfile with the given arguments and resets the dockerfile with the formatted string
func (d *dockerFile) Format(args ...any) *dockerFile {
	defer d.setRunState(false)
	nStr := d.Builder.String()
	d.Builder.Reset()
	_, err := d.Builder.WriteString(fmt.Sprintf(nStr, args...))
	if err != nil {
		d.errs = append(d.errs, err)
	}
	return d
}

// String returns the Dockerfile as a string
func (d *dockerFile) String() string {
	return d.Builder.String()
}

// Export exports the Dockerfile to a file
func (d *dockerFile) Export(path string, mode os.FileMode) error {
	if err := d.Validate(); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(d.String()), mode)
}

func (d *dockerFile) setRunState(state bool) {
	d.lastCmdIsRun = state
}
func (d *dockerFile) setCmdState(state bool) {
	d.cmdSet = state
}

// Validate validates the Dockerfile by checking for errors and returns a joined error if there are any
func (d *dockerFile) Validate() error {
	if len(d.errs) > 0 {
		return errors.Join(d.errs...)
	}
	return nil
}

// WithInline returns a build.SetBuildConfig that can be used to set the dockerfile inline
// within a service config.
//
// if there are errors in the dockerfile, it will return a fail setter that will return an error
func (d *dockerFile) WithInline() build.SetBuildConfig {
	if len(d.errs) > 0 {
		return build.Failf("dockerfile is invalid: %s", errors.Join(d.errs...))
	}
	return build.WithDockerfileInline(d.String())
}
