package client

import (
	"context"
	"io"

	"github.com/aptd3v/containerkit/pkg/client/response"
	sdk "github.com/moby/moby/client"
)

// PluginList returns plugins installed on the engine.
func (c *Client) PluginList(ctx context.Context, opt *PluginList) ([]response.Plugin, error) {
	listed, err := c.wrapped.PluginList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	plugins := make([]response.Plugin, len(listed.Items))
	for i, item := range listed.Items {
		plugins[i] = response.Plugin{Plugin: item}
	}
	return plugins, nil
}

// PluginInspect returns one plugin. Raw is the daemon JSON.
func (c *Client) PluginInspect(ctx context.Context, name string) (*response.Plugin, error) {
	inspected, err := c.wrapped.PluginInspect(ctx, name, sdk.PluginInspectOptions{})
	if err != nil {
		return nil, err
	}
	return &response.Plugin{Plugin: inspected.Plugin, Raw: cloneRaw(inspected.Raw)}, nil
}

// PluginInstall installs a plugin. The caller closes the progress stream.
func (c *Client) PluginInstall(ctx context.Context, name string, opt *PluginInstall) (io.ReadCloser, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	return c.wrapped.PluginInstall(ctx, name, op)
}

// PluginUpgrade upgrades an installed plugin. The caller closes the progress stream.
func (c *Client) PluginUpgrade(ctx context.Context, name string, opt *PluginUpgrade) (io.ReadCloser, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	return c.wrapped.PluginUpgrade(ctx, name, op)
}

// PluginCreate creates a plugin from a local rootfs tar. createContext is that tar.
func (c *Client) PluginCreate(ctx context.Context, createContext io.Reader, opt *PluginCreate) error {
	_, err := c.wrapped.PluginCreate(ctx, createContext, opt.apply())
	return err
}

// PluginEnable enables a plugin.
func (c *Client) PluginEnable(ctx context.Context, name string, opt *PluginEnable) error {
	_, err := c.wrapped.PluginEnable(ctx, name, opt.apply())
	return err
}

// PluginDisable disables a plugin.
func (c *Client) PluginDisable(ctx context.Context, name string, opt *PluginDisable) error {
	_, err := c.wrapped.PluginDisable(ctx, name, opt.apply())
	return err
}

// PluginRemove removes a plugin.
func (c *Client) PluginRemove(ctx context.Context, name string, opt *PluginRemove) error {
	_, err := c.wrapped.PluginRemove(ctx, name, opt.apply())
	return err
}

// PluginSet sets plugin arguments.
func (c *Client) PluginSet(ctx context.Context, name string, opt *PluginSet) error {
	_, err := c.wrapped.PluginSet(ctx, name, opt.apply())
	return err
}

// PluginPush pushes a plugin to a registry. The caller closes the progress stream.
func (c *Client) PluginPush(ctx context.Context, name string, opt *PluginPush) (io.ReadCloser, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	return c.wrapped.PluginPush(ctx, name, op)
}
