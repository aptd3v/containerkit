package kit

import "fmt"

// Job is a Compose job: shared spec plus job-only build/secrets/depends.
type Job struct {
	spec
	depends       []depend
	build         field[build]
	secrets       []secretRef
	configs       []fileConfigRef
	envFiles      []envFileRef
	pullPolicy    field[string]
	labelFiles    []string
	useAPISocket  field[bool]
	errs          []error
	nameGenerated bool
}

// NewJob applies typed options and always returns a job.
func NewJob(opts ...JobOption) *Job {
	j := &Job{}
	for i, opt := range opts {
		if opt == nil {
			j.errs = addErr(j.errs, cfgErr("job", fmt.Sprintf("option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyJob(j); err != nil {
			j.errs = addErr(j.errs, wrapReceiver("job", err))
		}
	}
	j.ensureName()
	return j
}

func (j *Job) ensureName() {
	if j == nil || j.name.ok() {
		return
	}
	s := &Service{spec: j.spec, build: j.build}
	if n := generateServiceName(s); n != "" {
		j.name.assign(n)
		j.nameGenerated = true
	}
}

func (j *Job) setName(name string) error {
	j.nameGenerated = false
	return j.spec.setName(name)
}

func (j *Job) validate() error {
	if j == nil {
		return cfgErr("job", ".", "", errRequired)
	}
	if !j.name.ok() {
		return cfgErr("job", "name", "Name", fmt.Errorf("%w: name", errRequired))
	}
	if !j.image.ok() && !j.build.ok() {
		return cfgErr("job", "image", "Image", fmt.Errorf("%w: image or build", errRequired))
	}
	return nil
}

// Validate reports accumulated option errors and missing required fields.
func (j *Job) Validate() error {
	if j == nil {
		return cfgErr("job", ".", "", errRequired)
	}
	return join(append(append([]error(nil), j.errs...), j.validate())...)
}

// Name returns the job name.
func (j *Job) Name() string {
	if j == nil || !j.name.ok() {
		return ""
	}
	return j.name.val
}
