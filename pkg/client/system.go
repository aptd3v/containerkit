package client

import (
	"context"

	"github.com/aptd3v/containerkit/pkg/client/auth"
	"github.com/aptd3v/containerkit/pkg/client/response"
	sdk "github.com/moby/moby/client"
)

// DiskUsage is [sdk.DiskUsageOptions].
// Nil options report containers, images, volumes, and build cache.
type DiskUsage sdk.DiskUsageOptions

func (o *DiskUsage) apply() sdk.DiskUsageOptions {
	if o == nil {
		return sdk.DiskUsageOptions{Containers: true, Images: true, BuildCache: true, Volumes: true}
	}
	return sdk.DiskUsageOptions(*o)
}

// EngineEvents is [sdk.EventsListOptions]. Nil watches every event.
type EngineEvents sdk.EventsListOptions

func (o *EngineEvents) apply() sdk.EventsListOptions { return sdk.EventsListOptions(apply(o)) }

// DistributionInspect is [sdk.DistributionInspectOptions].
type DistributionInspect sdk.DistributionInspectOptions

func (o *DistributionInspect) apply() sdk.DistributionInspectOptions {
	return sdk.DistributionInspectOptions(apply(o))
}

// Info returns daemon information.
func (c *Client) Info(ctx context.Context) (*response.Info, error) {
	info, err := c.wrapped.Info(ctx, sdk.InfoOptions{})
	if err != nil {
		return nil, err
	}
	return &response.Info{Info: info.Info}, nil
}

// Ping is [sdk.PingOptions]. Nil leaves API version negotiation to the client constructor.
type Ping sdk.PingOptions

func (o *Ping) apply() sdk.PingOptions { return sdk.PingOptions(apply(o)) }

// Ping checks that the daemon is reachable.
func (c *Client) Ping(ctx context.Context, opt *Ping) (*response.Ping, error) {
	pong, err := c.wrapped.Ping(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.Ping{
		APIVersion:     pong.APIVersion,
		OSType:         pong.OSType,
		Experimental:   pong.Experimental,
		BuilderVersion: string(pong.BuilderVersion),
	}, nil
}

// ServerVersion returns the daemon version.
func (c *Client) ServerVersion(ctx context.Context) (*response.ServerVersion, error) {
	version, err := c.wrapped.ServerVersion(ctx, sdk.ServerVersionOptions{})
	if err != nil {
		return nil, err
	}
	return &response.ServerVersion{
		Platform:      version.Platform.Name,
		Version:       version.Version,
		APIVersion:    version.APIVersion,
		MinAPIVersion: version.MinAPIVersion,
		Os:            version.Os,
		Arch:          version.Arch,
		Experimental:  version.Experimental,
	}, nil
}

// DiskUsage reports disk usage for the selected object types.
func (c *Client) DiskUsage(ctx context.Context, opt *DiskUsage) (*response.DiskUsage, error) {
	usage, err := c.wrapped.DiskUsage(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.DiskUsage{
		Containers: response.ContainersDiskUsage{
			ActiveCount: usage.Containers.ActiveCount,
			TotalCount:  usage.Containers.TotalCount,
			Reclaimable: usage.Containers.Reclaimable,
			TotalSize:   usage.Containers.TotalSize,
			Items:       usage.Containers.Items,
		},
		Images: response.ImagesDiskUsage{
			ActiveCount: usage.Images.ActiveCount,
			TotalCount:  usage.Images.TotalCount,
			Reclaimable: usage.Images.Reclaimable,
			TotalSize:   usage.Images.TotalSize,
			Items:       usage.Images.Items,
		},
		Volumes: response.VolumesDiskUsage{
			ActiveCount: usage.Volumes.ActiveCount,
			TotalCount:  usage.Volumes.TotalCount,
			Reclaimable: usage.Volumes.Reclaimable,
			TotalSize:   usage.Volumes.TotalSize,
			Items:       usage.Volumes.Items,
		},
		BuildCache: response.BuildCacheDiskUsage{
			ActiveCount: usage.BuildCache.ActiveCount,
			TotalCount:  usage.BuildCache.TotalCount,
			Reclaimable: usage.BuildCache.Reclaimable,
			TotalSize:   usage.BuildCache.TotalSize,
			Items:       usage.BuildCache.Items,
		},
	}, nil
}

// Events returns a stream of engine events. Cancel ctx to close the stream.
// The name is Engine-scoped so it does not collide with compose Events.
func (c *Client) Events(ctx context.Context, opt *EngineEvents) (<-chan response.EngineEvent, <-chan error) {
	if opt == nil {
		opt = &EngineEvents{}
	}
	stream := c.wrapped.Events(ctx, opt.apply())
	out := make(chan response.EngineEvent)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-stream.Err:
				if ok && err != nil {
					errs <- err
				}
				return
			case msg, ok := <-stream.Messages:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case out <- response.EngineEvent{Message: msg}:
				}
			}
		}
	}()
	return out, errs
}

// DistributionInspect returns the image digest and manifest descriptor from the registry.
func (c *Client) DistributionInspect(ctx context.Context, ref string, opt *DistributionInspect) (*response.DistributionInspect, error) {
	inspected, err := c.wrapped.DistributionInspect(ctx, ref, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.DistributionInspect{DistributionInspect: inspected.DistributionInspect}, nil
}

// RegistryLogin authenticates against a registry.
func (c *Client) RegistryLogin(ctx context.Context, creds auth.Auth) (*response.RegistryLogin, error) {
	logged, err := c.wrapped.RegistryLogin(ctx, sdk.RegistryLoginOptions{
		Username:      creds.Username,
		Password:      creds.Password,
		ServerAddress: creds.ServerAddress,
	})
	if err != nil {
		return nil, err
	}
	return &response.RegistryLogin{AuthResponse: logged.Auth}, nil
}
