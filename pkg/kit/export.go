package kit

import (
	"fmt"
	"io"

	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
	yaml "go.yaml.in/yaml/v3"
)

// Export is an explicit action struct for writing Compose YAML.
type Export struct {
	Writer io.Writer
}

// ComposeConfig translates the service into a Compose SDK service config.
func (s *Service) ComposeConfig() (composetypes.ServiceConfig, error) {
	if s == nil {
		return composetypes.ServiceConfig{}, cfgErr("service", ".", "", errRequired)
	}
	if err := s.Validate(); err != nil {
		return composetypes.ServiceConfig{}, wrapReceiver("service", err)
	}
	svc := composetypes.ServiceConfig{Name: s.Name()}
	if err := s.spec.applyCompose(&svc); err != nil {
		return composetypes.ServiceConfig{}, wrapReceiver("service", err)
	}
	applyServiceExtras(s, &svc)
	return svc, nil
}

// ComposeConfig translates the job into a Compose SDK job config.
func (j *Job) ComposeConfig() (composetypes.JobConfig, error) {
	if j == nil {
		return composetypes.JobConfig{}, cfgErr("job", ".", "", errRequired)
	}
	if err := j.Validate(); err != nil {
		return composetypes.JobConfig{}, wrapReceiver("job", err)
	}
	job := composetypes.JobConfig{Name: j.Name()}
	svc := composetypes.ServiceConfig{}
	if err := j.spec.applyCompose(&svc); err != nil {
		return composetypes.JobConfig{}, wrapReceiver("job", err)
	}
	job.ContainerSpec = svc.ContainerSpec
	job.WorkloadSpec = svc.WorkloadSpec
	applyJobExtras(j, &job)
	return job, nil
}

// Project returns the compose-go object model for this Compose file.
func (p *Compose) Project() (*composetypes.Project, error) {
	if p == nil {
		return nil, cfgErr("compose", ".", "", errRequired)
	}
	if err := p.Validate(); err != nil {
		return nil, wrapReceiver("compose", err)
	}
	out := &composetypes.Project{
		Name:     p.Name(),
		Services: composetypes.Services{},
	}
	for _, s := range p.services {
		svc, err := s.ComposeConfig()
		if err != nil {
			return nil, wrapPath("service."+s.Name(), err)
		}
		out.Services[svc.Name] = svc
	}
	if len(p.jobs) > 0 {
		out.Jobs = composetypes.Jobs{}
		for _, j := range p.jobs {
			job, err := j.ComposeConfig()
			if err != nil {
				return nil, wrapPath("job."+j.Name(), err)
			}
			out.Jobs[job.Name] = job
		}
	}
	if len(p.networks) > 0 {
		out.Networks = composetypes.Networks{}
		for _, n := range p.networks {
			cfg := projectNetwork(n)
			cfg.Name = engineNetworkName(p.Name(), n.name.val)
			out.Networks[n.name.val] = cfg
		}
	}
	if len(p.volumes) > 0 {
		out.Volumes = composetypes.Volumes{}
		for _, v := range p.volumes {
			out.Volumes[v.name.val] = projectVolume(v)
		}
	}
	if len(p.secrets) > 0 {
		out.Secrets = composetypes.Secrets{}
		for _, s := range p.secrets {
			out.Secrets[s.name.val] = projectSecretCfg(s)
		}
	}
	if len(p.configs) > 0 {
		out.Configs = composetypes.Configs{}
		for _, f := range p.configs {
			out.Configs[f.name.val] = projectFileCfg(f)
		}
	}
	if len(p.models) > 0 {
		out.Models = composetypes.Models{}
		for _, m := range p.models {
			cfg := composetypes.ModelConfig{Name: m.name.val}
			if m.model.ok() {
				cfg.Model = m.model.val
			}
			if m.contextSize.ok() {
				cfg.ContextSize = m.contextSize.val
			}
			cfg.RuntimeFlags = append([]string(nil), m.runtimeFlags...)
			out.Models[m.name.val] = cfg
		}
	}
	return out, nil
}

// ExportYAML returns the YAML encoding of this Compose file.
func (p *Compose) ExportYAML() ([]byte, error) {
	proj, err := p.Project()
	if err != nil {
		return nil, err
	}
	b, err := proj.MarshalYAML()
	if err != nil {
		return nil, cfgErr("compose", "export", "Export", err)
	}
	if len(p.includes) == 0 {
		return b, nil
	}
	incYAML, err := yaml.Marshal(map[string]any{"include": composeIncludes(p.includes)})
	if err != nil {
		return nil, cfgErr("compose", "export", "Export", err)
	}
	out := make([]byte, 0, len(incYAML)+len(b))
	out = append(out, incYAML...)
	out = append(out, b...)
	return out, nil
}

func composeIncludes(incs []includeRef) []any {
	out := make([]any, 0, len(incs))
	for _, inc := range incs {
		if len(inc.paths) == 1 && !inc.projectDir.ok() && len(inc.envFiles) == 0 {
			out = append(out, inc.paths[0])
			continue
		}
		m := map[string]any{}
		if len(inc.paths) == 1 {
			m["path"] = inc.paths[0]
		} else if len(inc.paths) > 0 {
			m["path"] = append([]string(nil), inc.paths...)
		}
		if inc.projectDir.ok() {
			m["project_directory"] = inc.projectDir.val
		}
		if len(inc.envFiles) > 0 {
			m["env_file"] = append([]string(nil), inc.envFiles...)
		}
		out = append(out, m)
	}
	return out
}

// Write exports Compose YAML using an explicit [Export] action.
func (p *Compose) Write(act Export) error {
	if act.Writer == nil {
		return cfgErr("compose", "export.writer", "Writer", fmt.Errorf("%w: writer", errRequired))
	}
	b, err := p.ExportYAML()
	if err != nil {
		return err
	}
	_, err = act.Writer.Write(b)
	if err != nil {
		return cfgErr("compose", "export", "Export", err)
	}
	return nil
}
