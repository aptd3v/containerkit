package containerkit

import (
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckSecondsAndDurations(t *testing.T) {
	sec := NewContainer("h1").HealthCheck(Health{
		Test:        []string{"CMD", "true"},
		Timeout:     2,
		Interval:    3,
		StartPeriod: 1,
		Retries:     4,
	})
	require.NoError(t, sec.Validate())
	require.NotNil(t, sec.Config.Container.Healthcheck)
	require.Equal(t, []string{"CMD", "true"}, sec.Config.Container.Healthcheck.Test)

	dur := NewContainer("h2").HealthCheck(Health{
		Test:         []string{"CMD-SHELL", "true"},
		TimeoutD:     "2s",
		IntervalD:    "5s",
		StartPeriodD: "1s",
		Retries:      2,
	})
	require.NoError(t, dur.Validate())
	require.NotNil(t, dur.Config.Container.Healthcheck)

	off := NewContainer("h3").DisabledHealthCheck()
	require.NoError(t, off.Validate())
	require.Equal(t, []string{"NONE"}, off.Config.Container.Healthcheck.Test)

	bad := NewContainer("h4").HealthCheck(Health{TimeoutD: "nope", Test: []string{"CMD", "true"}})
	require.Error(t, bad.Validate())
}

func TestMountAndEndpointAndRestartAlias(t *testing.T) {
	c := NewContainer("net").RestartUnlessStopped().Mount(Mount{Source: "/tmp", Target: "/mnt", ReadOnly: true}).Mount(Mount{Source: "/tmp", Target: "/mnt2", Type: MountVolume}).Endpoint("front").Endpoint("back", EndpointOpts{
		Aliases:       []string{"api"},
		Links:         []string{"db"},
		MacAddress:    "02:42:ac:11:00:02",
		DriverOptions: map[string]string{"com.docker.network.endpoint.sysctls": "net.ipv4.conf.IFNAME.rp_filter=0"},
		IPv4:          "10.0.0.10",
		IPv6:          "fd00::10",
		LinkLocalIPs:  []string{"169.254.1.1"},
	})
	require.NoError(t, c.Validate())
	require.Len(t, c.Config.Host.Mounts, 2)
	require.Equal(t, mount.TypeBind, c.Config.Host.Mounts[0].Type)
	require.True(t, c.Config.Host.Mounts[0].ReadOnly)

	create := false
	full := NewContainer("full").Image("alpine").Mount(Mount{
		Source: "data", Target: "/var/lib/data",
		Consistency: ConsistencyDelegated,
		Volume: &VolumeOptions{
			NoCopy: true, Subpath: "nested", Labels: map[string]string{"foo": "bar"},
			Driver: "local", DriverOptions: map[string]string{"type": "tmpfs"},
		},
	}).Mount(Mount{
		Source: "/src", Target: "/dst", ReadOnly: true,
		Bind: &BindOptions{
			Propagation: PropagationRShared, SELinux: SELinuxShared,
			NonRecursive: true, CreateHostPath: &create,
		},
	}).Mount(Mount{
		Target: "/run",
		Tmpfs:  &TmpfsOptions{SizeBytes: 4096, Mode: 0o755},
	}).Mount(Mount{
		Source: "alpine", Target: "/img",
		Image: &ImageOptions{Subpath: "bin"},
	})
	require.NoError(t, full.Validate())
	require.Equal(t, mount.TypeVolume, full.Config.Host.Mounts[0].Type)
	require.Equal(t, mount.ConsistencyDelegated, full.Config.Host.Mounts[0].Consistency)
	require.True(t, full.Config.Host.Mounts[0].VolumeOptions.NoCopy)
	require.Equal(t, "local", full.Config.Host.Mounts[0].VolumeOptions.DriverConfig.Name)
	require.Equal(t, "tmpfs", full.Config.Host.Mounts[0].VolumeOptions.DriverConfig.Options["type"])
	require.Equal(t, mount.PropagationRShared, full.Config.Host.Mounts[1].BindOptions.Propagation)
	require.False(t, full.Config.Host.Mounts[1].BindOptions.CreateMountpoint)
	require.Equal(t, mount.TypeTmpfs, full.Config.Host.Mounts[2].Type)
	require.Equal(t, int64(4096), full.Config.Host.Mounts[2].TmpfsOptions.SizeBytes)
	require.Equal(t, mount.TypeImage, full.Config.Host.Mounts[3].Type)
	require.Equal(t, "bin", full.Config.Host.Mounts[3].ImageOptions.Subpath)

	p := NewProject("mounts")
	p.WithService("app", full)
	require.NoError(t, p.Validate())
	svc, err := p.GetService("app")
	require.NoError(t, err)
	for i := range svc.Volumes {
		switch svc.Volumes[i].Target {
		case "/dst":
			require.Equal(t, SELinuxShared, svc.Volumes[i].Bind.SELinux)
			require.Equal(t, string(PropagationRShared), svc.Volumes[i].Bind.Propagation)
			require.False(t, bool(svc.Volumes[i].Bind.CreateHostPath))
			require.Equal(t, "disabled", svc.Volumes[i].Bind.Recursive)
		case "/var/lib/data":
			require.True(t, svc.Volumes[i].Volume.NoCopy)
			require.Equal(t, "nested", svc.Volumes[i].Volume.Subpath)
			require.Equal(t, "bar", svc.Volumes[i].Volume.Labels["foo"])
			require.Equal(t, "delegated", svc.Volumes[i].Consistency)
		case "/img":
			require.Equal(t, "bin", svc.Volumes[i].Image.SubPath)
		case "/run":
			require.Equal(t, "tmpfs", svc.Volumes[i].Type)
			require.Equal(t, uint32(0o755), svc.Volumes[i].Tmpfs.Mode)
		}
	}
	require.Equal(t, "local", p.Unwrap().Volumes["data"].Driver)
	require.Equal(t, "tmpfs", p.Unwrap().Volumes["data"].DriverOpts["type"])

	bad := NewContainer("bad").Mount(Mount{Source: "/src", Target: "/dst", Bind: &BindOptions{SELinux: "nope"}})
	require.Error(t, bad.Validate())
	require.Error(t, NewContainer("tmp").Mount(Mount{Source: "/src", Target: "/run", Type: MountTmpfs}).Validate())
	require.Contains(t, c.Config.Network.EndpointsConfig, "front")
	require.Contains(t, c.Config.Network.EndpointsConfig, "back")
	ep := c.Config.Network.EndpointsConfig["back"]
	require.Equal(t, "02:42:ac:11:00:02", ep.MacAddress)
	require.NotNil(t, ep.IPAMConfig)
	require.Equal(t, "10.0.0.10", ep.IPAMConfig.IPv4Address)
}
