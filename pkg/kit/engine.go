package kit

// EngineConfig translates the container into Engine create structs.
func (c *Container) EngineConfig() (*ContainerConfig, *HostConfig, *NetworkingConfig, error) {
	if c == nil {
		return nil, nil, nil, cfgErr("container", ".", "", errRequired)
	}
	if err := c.Validate(); err != nil {
		return nil, nil, nil, wrapReceiver("container", err)
	}
	cfg := &ContainerConfig{}
	host := &HostConfig{}
	net := &NetworkingConfig{}
	if err := c.spec.applyEngine(cfg, host, net); err != nil {
		return nil, nil, nil, wrapReceiver("container", err)
	}
	if c.autoRemove.ok() {
		host.AutoRemove = c.autoRemove.val
	}
	if c.stdinOnce.ok() {
		cfg.StdinOnce = c.stdinOnce.val
	}
	if c.argsEscaped.ok() {
		cfg.ArgsEscaped = c.argsEscaped.val
	}
	if len(c.shell) > 0 {
		cfg.Shell = append([]string(nil), c.shell...)
	}
	if len(c.onBuild) > 0 {
		cfg.OnBuild = append([]string(nil), c.onBuild...)
	}
	if c.attachStdin.ok() {
		cfg.AttachStdin = c.attachStdin.val
	}
	if c.attachStdout.ok() {
		cfg.AttachStdout = c.attachStdout.val
	}
	if c.attachStderr.ok() {
		cfg.AttachStderr = c.attachStderr.val
	}
	if c.publishAll.ok() {
		host.PublishAllPorts = c.publishAll.val
	}
	if c.idFile.ok() {
		host.ContainerIDFile = c.idFile.val
	}
	if len(c.maskedPaths) > 0 {
		host.MaskedPaths = append([]string(nil), c.maskedPaths...)
	}
	if len(c.readonlyPaths) > 0 {
		host.ReadonlyPaths = append([]string(nil), c.readonlyPaths...)
	}
	return cfg, host, net, nil
}
