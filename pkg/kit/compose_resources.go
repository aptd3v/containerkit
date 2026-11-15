package kit

import "fmt"

func (n *projectNet) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("network", err)
	}
	n.name.assign(name)
	return nil
}

func (n *projectNet) setDriver(d string) error {
	n.driver.assign(d)
	return nil
}

func (n *projectNet) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if n.labels == nil {
		n.labels = map[string]string{}
	}
	n.labels[k] = val
	return nil
}

func (v *projectVol) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("volume", err)
	}
	v.name.assign(name)
	return nil
}

func (v *projectVol) setDriver(d string) error {
	v.driver.assign(d)
	return nil
}

func (v *projectVol) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if v.labels == nil {
		v.labels = map[string]string{}
	}
	v.labels[k] = val
	return nil
}

func (s *projectSecret) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("secret", err)
	}
	s.name.assign(name)
	return nil
}

func (s *projectSecret) setFile(p string) error {
	s.file.assign(p)
	return nil
}

func (s *projectSecret) setDriver(d string) error {
	s.driver.assign(d)
	return nil
}

func (s *projectSecret) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if s.labels == nil {
		s.labels = map[string]string{}
	}
	s.labels[k] = val
	return nil
}

func (f *projectFile) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("config", err)
	}
	f.name.assign(name)
	return nil
}

func (f *projectFile) setFile(p string) error {
	f.file.assign(p)
	return nil
}

func (f *projectFile) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if f.labels == nil {
		f.labels = map[string]string{}
	}
	f.labels[k] = val
	return nil
}

// InternalOpt is the value returned by [Internal].
// It implements [ComposeNetOption] and [EngineNetworkOption].
type InternalOpt struct{ v bool }

// Internal marks a project network as internal.
func Internal(v bool) InternalOpt { return InternalOpt{v: v} }

func (o InternalOpt) applyComposeNet(n *projectNet) error {
	n.internal.assign(o.v)
	return nil
}

// AttachableOpt is the value returned by [Attachable].
// It implements [ComposeNetOption] and [EngineNetworkOption].
type AttachableOpt struct{ v bool }

// Attachable marks a project network as attachable.
func Attachable(v bool) AttachableOpt { return AttachableOpt{v: v} }

func (o AttachableOpt) applyComposeNet(n *projectNet) error {
	n.attachable.assign(o.v)
	return nil
}

// SubnetOpt is the value returned by [Subnet].
// It implements [ComposeNetOption].
type SubnetOpt struct{ s string }

// Subnet sets an IPAM subnet inside [AddNetwork].
func Subnet(s string) SubnetOpt { return SubnetOpt{s: s} }

func (o SubnetOpt) applyComposeNet(n *projectNet) error {
	n.subnet.assign(o.s)
	return nil
}

// GatewayOpt is the value returned by [Gateway].
// It implements [ComposeNetOption].
type GatewayOpt struct{ g string }

// Gateway sets an IPAM gateway inside [AddNetwork].
func Gateway(g string) GatewayOpt { return GatewayOpt{g: g} }

func (o GatewayOpt) applyComposeNet(n *projectNet) error {
	n.gateway.assign(o.g)
	return nil
}

// AddNetworkOpt is the value returned by [AddNetwork].
// It implements [ComposeOption].
type AddNetworkOpt struct{ opts []ComposeNetOption }

// AddNetwork defines a project-level network.
func AddNetwork(opts ...ComposeNetOption) AddNetworkOpt { return AddNetworkOpt{opts: opts} }

func (o AddNetworkOpt) applyCompose(p *Compose) error {
	var n projectNet
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("network.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyComposeNet(&n); err != nil {
			errs = append(errs, wrapPath("network", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !n.name.ok() {
		return cfgErr("", "network.name", "Name", fmt.Errorf("%w: network name", errRequired))
	}
	p.networks = append(p.networks, n)
	return nil
}

// AddVolumeOpt is the value returned by [AddVolume].
// It implements [ComposeOption].
type AddVolumeOpt struct{ opts []ComposeVolOption }

// AddVolume defines a project-level named volume.
func AddVolume(opts ...ComposeVolOption) AddVolumeOpt { return AddVolumeOpt{opts: opts} }

func (o AddVolumeOpt) applyCompose(p *Compose) error {
	var v projectVol
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("volume.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyComposeVol(&v); err != nil {
			errs = append(errs, wrapPath("volume", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !v.name.ok() {
		return cfgErr("", "volume.name", "Name", fmt.Errorf("%w: volume name", errRequired))
	}
	p.volumes = append(p.volumes, v)
	return nil
}

// AddSecretOpt is the value returned by [AddSecret].
// It implements [ComposeOption].
type AddSecretOpt struct{ opts []ComposeSecretOption }

// AddSecret defines a project-level secret.
func AddSecret(opts ...ComposeSecretOption) AddSecretOpt { return AddSecretOpt{opts: opts} }

func (o AddSecretOpt) applyCompose(p *Compose) error {
	var s projectSecret
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("secret.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyComposeSecret(&s); err != nil {
			errs = append(errs, wrapPath("secret", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !s.name.ok() {
		return cfgErr("", "secret.name", "Name", fmt.Errorf("%w: secret name", errRequired))
	}
	p.secrets = append(p.secrets, s)
	return nil
}

// AddFileOpt is the value returned by [AddFile].
// It implements [ComposeOption].
type AddFileOpt struct{ opts []ComposeFileOption }

// AddFile defines a project-level Compose config object.
func AddFile(opts ...ComposeFileOption) AddFileOpt { return AddFileOpt{opts: opts} }

func (o AddFileOpt) applyCompose(p *Compose) error {
	var f projectFile
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("config.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyComposeFile(&f); err != nil {
			errs = append(errs, wrapPath("config", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !f.name.ok() {
		return cfgErr("", "config.name", "Name", fmt.Errorf("%w: config name", errRequired))
	}
	p.configs = append(p.configs, f)
	return nil
}
