package containerkit

import (
	"testing"

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
}

func TestMountAndEndpointAndRestartAlias(t *testing.T) {
	c := NewContainer("net").
		RestartUnlessStopped().
		Mount(Mount{Source: "/tmp", Target: "/mnt", ReadOnly: true}).
		Mount(Mount{Source: "/tmp", Target: "/mnt2", Type: MountVolume}).
		Endpoint("front").
		Endpoint("back", EndpointOpts{
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
	require.Contains(t, c.Config.Network.EndpointsConfig, "front")
	require.Contains(t, c.Config.Network.EndpointsConfig, "back")
	ep := c.Config.Network.EndpointsConfig["back"]
	require.Equal(t, "02:42:ac:11:00:02", ep.MacAddress)
	require.NotNil(t, ep.IPAMConfig)
	require.Equal(t, "10.0.0.10", ep.IPAMConfig.IPv4Address)
}
