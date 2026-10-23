package client

import (
	"context"

	"github.com/aptd3v/containerkit/pkg/client/auth"
	"github.com/moby/moby/api/types/plugin"
	sdk "github.com/moby/moby/client"
)

// PluginList is options for PluginList.
type PluginList struct {
	Filters []Filter
}

func (o *PluginList) apply() sdk.PluginListOptions {
	if o == nil {
		return sdk.PluginListOptions{Filters: argsFromFilters(nil)}
	}
	return sdk.PluginListOptions{Filters: argsFromFilters(o.Filters)}
}

// PluginInstall is options for PluginInstall.
type PluginInstall struct {
	Disabled              bool
	AcceptAllPermissions  bool
	Auth                  auth.Auth
	RemoteRef             string
	PrivilegeFunc         func(context.Context) (string, error)
	AcceptPermissionsFunc func(context.Context, plugin.Privileges) (bool, error)
	Args                  []string
}

func (o *PluginInstall) apply() (sdk.PluginInstallOptions, error) {
	if o == nil {
		return sdk.PluginInstallOptions{}, nil
	}
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.PluginInstallOptions{}, err
	}
	return sdk.PluginInstallOptions{
		Disabled:              o.Disabled,
		AcceptAllPermissions:  o.AcceptAllPermissions,
		RegistryAuth:          encoded,
		RemoteRef:             o.RemoteRef,
		PrivilegeFunc:         o.PrivilegeFunc,
		AcceptPermissionsFunc: o.AcceptPermissionsFunc,
		Args:                  o.Args,
	}, nil
}

// PluginUpgrade is options for PluginUpgrade.
type PluginUpgrade struct {
	Disabled              bool
	AcceptAllPermissions  bool
	Auth                  auth.Auth
	RemoteRef             string
	PrivilegeFunc         func(context.Context) (string, error)
	AcceptPermissionsFunc func(context.Context, plugin.Privileges) (bool, error)
	Args                  []string
}

func (o *PluginUpgrade) apply() (sdk.PluginUpgradeOptions, error) {
	if o == nil {
		return sdk.PluginUpgradeOptions{}, nil
	}
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.PluginUpgradeOptions{}, err
	}
	return sdk.PluginUpgradeOptions{
		Disabled:              o.Disabled,
		AcceptAllPermissions:  o.AcceptAllPermissions,
		RegistryAuth:          encoded,
		RemoteRef:             o.RemoteRef,
		PrivilegeFunc:         o.PrivilegeFunc,
		AcceptPermissionsFunc: o.AcceptPermissionsFunc,
		Args:                  o.Args,
	}, nil
}

// PluginPush is options for PluginPush.
type PluginPush struct {
	Auth auth.Auth
}

func (o *PluginPush) apply() (sdk.PluginPushOptions, error) {
	if o == nil {
		return sdk.PluginPushOptions{}, nil
	}
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.PluginPushOptions{}, err
	}
	return sdk.PluginPushOptions{RegistryAuth: encoded}, nil
}

// PluginCreate is options for PluginCreate.
type PluginCreate struct {
	RepoName string
}

func (o *PluginCreate) apply() sdk.PluginCreateOptions {
	if o == nil {
		return sdk.PluginCreateOptions{}
	}
	return sdk.PluginCreateOptions{RepoName: o.RepoName}
}

// PluginEnable is options for PluginEnable.
type PluginEnable struct {
	Timeout int
}

func (o *PluginEnable) apply() sdk.PluginEnableOptions {
	if o == nil {
		return sdk.PluginEnableOptions{}
	}
	return sdk.PluginEnableOptions{Timeout: o.Timeout}
}

// PluginDisable is options for PluginDisable.
type PluginDisable struct {
	Force bool
}

func (o *PluginDisable) apply() sdk.PluginDisableOptions {
	if o == nil {
		return sdk.PluginDisableOptions{}
	}
	return sdk.PluginDisableOptions{Force: o.Force}
}

// PluginRemove is options for PluginRemove.
type PluginRemove struct {
	Force bool
}

func (o *PluginRemove) apply() sdk.PluginRemoveOptions {
	if o == nil {
		return sdk.PluginRemoveOptions{}
	}
	return sdk.PluginRemoveOptions{Force: o.Force}
}

// PluginSet is options for PluginSet.
type PluginSet struct {
	Args []string
}

func (o *PluginSet) apply() sdk.PluginSetOptions {
	if o == nil {
		return sdk.PluginSetOptions{}
	}
	return sdk.PluginSetOptions{Args: o.Args}
}
