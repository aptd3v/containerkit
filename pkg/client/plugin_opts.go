package client

import sdk "github.com/moby/moby/client"

// PluginList is [sdk.PluginListOptions].
type PluginList sdk.PluginListOptions

func (o *PluginList) apply() sdk.PluginListOptions { return sdk.PluginListOptions(apply(o)) }

// PluginInstall is [sdk.PluginInstallOptions].
type PluginInstall sdk.PluginInstallOptions

func (o *PluginInstall) apply() sdk.PluginInstallOptions { return sdk.PluginInstallOptions(apply(o)) }

// PluginUpgrade is [sdk.PluginUpgradeOptions].
type PluginUpgrade sdk.PluginUpgradeOptions

func (o *PluginUpgrade) apply() sdk.PluginUpgradeOptions { return sdk.PluginUpgradeOptions(apply(o)) }

// PluginPush is [sdk.PluginPushOptions].
type PluginPush sdk.PluginPushOptions

func (o *PluginPush) apply() sdk.PluginPushOptions { return sdk.PluginPushOptions(apply(o)) }

// PluginCreate is [sdk.PluginCreateOptions].
type PluginCreate sdk.PluginCreateOptions

func (o *PluginCreate) apply() sdk.PluginCreateOptions { return sdk.PluginCreateOptions(apply(o)) }

// PluginEnable is [sdk.PluginEnableOptions].
type PluginEnable sdk.PluginEnableOptions

func (o *PluginEnable) apply() sdk.PluginEnableOptions { return sdk.PluginEnableOptions(apply(o)) }

// PluginDisable is [sdk.PluginDisableOptions].
type PluginDisable sdk.PluginDisableOptions

func (o *PluginDisable) apply() sdk.PluginDisableOptions { return sdk.PluginDisableOptions(apply(o)) }

// PluginRemove is [sdk.PluginRemoveOptions].
type PluginRemove sdk.PluginRemoveOptions

func (o *PluginRemove) apply() sdk.PluginRemoveOptions { return sdk.PluginRemoveOptions(apply(o)) }

// PluginSet is [sdk.PluginSetOptions].
type PluginSet sdk.PluginSetOptions

func (o *PluginSet) apply() sdk.PluginSetOptions { return sdk.PluginSetOptions(apply(o)) }
