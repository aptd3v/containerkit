package kit

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// AcceptAllPermissionsOpt is the value returned by [AcceptAllPermissions].
// It implements [EnginePluginInstallOption] and [EnginePluginUpgradeOption].
type AcceptAllPermissionsOpt struct{ v bool }

// AcceptAllPermissions skips the privilege prompt on [InstallPlugin] and
// [UpgradePlugin].
func AcceptAllPermissions(v bool) AcceptAllPermissionsOpt {
	return AcceptAllPermissionsOpt{v: v}
}

func (o AcceptAllPermissionsOpt) applyEnginePluginInstallOption(a *enginePluginInstallAct) error {
	a.opts.AcceptAllPermissions = o.v
	return nil
}

func (o AcceptAllPermissionsOpt) applyEnginePluginUpgradeOption(a *enginePluginUpgradeAct) error {
	a.opts.AcceptAllPermissions = o.v
	return nil
}

// PrivilegeOpt is the value returned by [Privilege].
// It implements [EnginePluginInstallOption] and [EnginePluginUpgradeOption].
type PrivilegeOpt struct {
	fn func(context.Context) (string, error)
}

// Privilege supplies a registry-auth retry callback for [InstallPlugin] and
// [UpgradePlugin].
func Privilege(fn func(context.Context) (string, error)) PrivilegeOpt {
	return PrivilegeOpt{fn: fn}
}

func (o PrivilegeOpt) applyEnginePluginInstallOption(a *enginePluginInstallAct) error {
	a.opts.PrivilegeFunc = o.fn
	return nil
}

func (o PrivilegeOpt) applyEnginePluginUpgradeOption(a *enginePluginUpgradeAct) error {
	a.opts.PrivilegeFunc = o.fn
	return nil
}

// AcceptPermissionsOpt is the value returned by [AcceptPermissions].
// It implements [EnginePluginInstallOption] and [EnginePluginUpgradeOption].
type AcceptPermissionsOpt struct {
	fn func(context.Context, PluginPrivileges) (bool, error)
}

// AcceptPermissions supplies a privilege-prompt callback for [InstallPlugin]
// and [UpgradePlugin].
func AcceptPermissions(fn func(context.Context, PluginPrivileges) (bool, error)) AcceptPermissionsOpt {
	return AcceptPermissionsOpt{fn: fn}
}

func (o AcceptPermissionsOpt) applyEnginePluginInstallOption(a *enginePluginInstallAct) error {
	a.opts.AcceptPermissionsFunc = o.fn
	return nil
}

func (o AcceptPermissionsOpt) applyEnginePluginUpgradeOption(a *enginePluginUpgradeAct) error {
	a.opts.AcceptPermissionsFunc = o.fn
	return nil
}

func (o TimeoutOpt) applyEnginePluginEnableOption(a *enginePluginEnableAct) error {
	sec := int(o.d / time.Second)
	if o.d > 0 && sec == 0 {
		sec = 1
	}
	a.opts.Timeout = sec
	return nil
}

func (o DisabledOpt) applyEnginePluginInstallOption(a *enginePluginInstallAct) error {
	a.opts.Disabled = o.v
	return nil
}

func (o DisabledOpt) applyEnginePluginUpgradeOption(a *enginePluginUpgradeAct) error {
	a.opts.Disabled = o.v
	return nil
}

func (o ArgsOpt) applyEnginePluginInstallOption(a *enginePluginInstallAct) error {
	a.opts.Args = append(a.opts.Args, o.args...)
	return nil
}

func (o ArgsOpt) applyEnginePluginUpgradeOption(a *enginePluginUpgradeAct) error {
	a.opts.Args = append(a.opts.Args, o.args...)
	return nil
}

// PackPlugin tars a plugin data directory for [CreatePlugin].
//
// dir must contain config.json and a rootfs/ directory, matching
// https://docs.docker.com/engine/extend/#developing-a-plugin.
func PackPlugin(dir string) (io.ReadCloser, error) {
	dir = filepath.Clean(dir)
	st, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("pack plugin: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("pack plugin: %s is not a directory", dir)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.json")); err != nil {
		return nil, fmt.Errorf("pack plugin: config.json: %w", err)
	}
	rootfs, err := os.Stat(filepath.Join(dir, "rootfs"))
	if err != nil {
		return nil, fmt.Errorf("pack plugin: rootfs: %w", err)
	}
	if !rootfs.IsDir() {
		return nil, fmt.Errorf("pack plugin: rootfs is not a directory")
	}
	pr, pw := io.Pipe()
	go func() {
		_ = pw.CloseWithError(writePluginTar(pw, dir))
	}()
	return pr, nil
}

func writePluginTar(w io.Writer, dir string) error {
	tw := tar.NewWriter(w)
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if !d.Type().IsRegular() && !d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if d.IsDir() {
			name += "/"
		}
		hdr.Name = name
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tw, f)
		closeErr := f.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	closeErr := tw.Close()
	if walkErr != nil {
		return walkErr
	}
	return closeErr
}

var (
	_ EnginePluginCreateOption  = Content(nil)
	_ EnginePluginCreateOption  = Name("plugin")
	_ EnginePluginEnableOption  = Timeout(0)
	_ EnginePluginInstallOption = Disabled(true)
	_ EnginePluginInstallOption = AcceptAllPermissions(true)
	_ EnginePluginInstallOption = Privilege(nil)
	_ EnginePluginInstallOption = AcceptPermissions(nil)
	_ EnginePluginInstallOption = Args()
	_ EnginePluginInstallOption = Name("remote")
	_ EnginePluginInstallOption = RegistryAuth("")
	_ EnginePluginUpgradeOption = Disabled(true)
	_ EnginePluginUpgradeOption = AcceptAllPermissions(true)
	_ EnginePluginUpgradeOption = Privilege(nil)
	_ EnginePluginUpgradeOption = AcceptPermissions(nil)
	_ EnginePluginUpgradeOption = Args()
	_ EnginePluginUpgradeOption = Name("remote")
	_ EnginePluginUpgradeOption = RegistryAuth("")
	_ EnginePluginSetOption     = Args()
	_ EnginePluginListOption    = Filter("enabled", "true")
	_ EnginePluginRemoveOption  = Force(true)
	_ EnginePluginDisableOption = Force(true)
	_ EnginePluginPushOption    = RegistryAuth("")
)
