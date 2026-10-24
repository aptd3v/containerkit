package kit

// Architecture sets the platform CPU architecture, for example amd64 or ppc64le.
// Parameter:
//   - architecture: the architecture to be used for the platform
//
// Architecture field specifies the CPU architecture, for example `amd64` or `ppc64le`.
func (c *Container) Architecture(architecture string) *Container {
	c.Config.Platform.Architecture = architecture
	return c
}

// OS sets the OS for the platform
// Parameter:
//   - OS: the OS to be used for the platform
//
// OS specifies the operating system, for example `linux` or `windows`.
func (c *Container) OS(OS string) *Container {
	c.Config.Platform.OS = OS
	return c
}

// OSVersion sets the OS version for the platform
// Parameter:
//   - OSVersion: the OS version to be used for the platform
//
// OSVersion is an optional field specifying the operating system version, for example on Windows `10.0.14393.1066`.
func (c *Container) OSVersion(OSVersion string) *Container {
	c.Config.Platform.OSVersion = OSVersion
	return c
}

// OSFeatures sets the OS features for the platform
// Parameter:
//   - OSFeatures: the OS features to be used for the platform
//
// OSFeatures is an optional field specifying an array of strings, each listing a required OS feature (for example on Windows `win32k`).
func (c *Container) OSFeatures(OSFeatures ...string) *Container {
	if c.Config.Platform.OSFeatures == nil {
		c.Config.Platform.OSFeatures = make([]string, 0)
	}
	c.Config.Platform.OSFeatures = append(c.Config.Platform.OSFeatures, OSFeatures...)
	return c
}

// Variant sets the variant for the platform
// Parameter:
//   - variant: the variant to be used for the platform
//
// Variant is an optional field specifying a variant of the CPU, for example `v7` to specify ARMv7 when architecture is `arm`.
func (c *Container) Variant(variant string) *Container {
	c.Config.Platform.Variant = variant
	return c
}
