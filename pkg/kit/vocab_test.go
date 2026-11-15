package kit_test

import (
	"testing"

	"github.com/aptd3v/containerkit/pkg/avail"
	"github.com/aptd3v/containerkit/pkg/cap"
	"github.com/aptd3v/containerkit/pkg/cascade"
	"github.com/aptd3v/containerkit/pkg/cgroupns"
	"github.com/aptd3v/containerkit/pkg/depend"
	"github.com/aptd3v/containerkit/pkg/df"
	"github.com/aptd3v/containerkit/pkg/endpoint"
	"github.com/aptd3v/containerkit/pkg/failact"
	"github.com/aptd3v/containerkit/pkg/ipc"
	"github.com/aptd3v/containerkit/pkg/isolation"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/logdrv"
	"github.com/aptd3v/containerkit/pkg/netdrv"
	"github.com/aptd3v/containerkit/pkg/netmode"
	"github.com/aptd3v/containerkit/pkg/oci"
	"github.com/aptd3v/containerkit/pkg/order"
	"github.com/aptd3v/containerkit/pkg/pidns"
	"github.com/aptd3v/containerkit/pkg/portmode"
	"github.com/aptd3v/containerkit/pkg/progress"
	"github.com/aptd3v/containerkit/pkg/propagate"
	"github.com/aptd3v/containerkit/pkg/proto"
	"github.com/aptd3v/containerkit/pkg/pull"
	"github.com/aptd3v/containerkit/pkg/recursive"
	"github.com/aptd3v/containerkit/pkg/restart"
	"github.com/aptd3v/containerkit/pkg/role"
	"github.com/aptd3v/containerkit/pkg/scheme"
	"github.com/aptd3v/containerkit/pkg/selinux"
	"github.com/aptd3v/containerkit/pkg/signal"
	"github.com/aptd3v/containerkit/pkg/svcmode"
	"github.com/aptd3v/containerkit/pkg/userns"
	"github.com/aptd3v/containerkit/pkg/uts"
	"github.com/aptd3v/containerkit/pkg/vol"
	"github.com/aptd3v/containerkit/pkg/wait"
	"github.com/aptd3v/containerkit/pkg/watch"
)

func TestVocabularyPackagesWireIntoSetters(t *testing.T) {
	t.Parallel()
	c := kit.NewContainer(
		kit.Name("vocab"),
		kit.Image("alpine:latest"),
		kit.CapAdd(cap.NET_ADMIN),
		kit.CapDrop(cap.MKNOD),
		kit.IPC(ipc.Shareable),
		kit.PID(pidns.Host),
		kit.UTS(uts.Host),
		kit.UserNS(userns.Host),
		kit.CgroupNS(cgroupns.Private),
		kit.Isolation(isolation.Default),
		kit.NetworkMode(netmode.Bridge),
		kit.StopSignal(signal.SIGTERM),
		kit.Restart(kit.Policy(restart.Always)),
		kit.Publish(kit.ContainerPort(80), kit.Protocol(proto.TCP), kit.Mode(portmode.Host)),
		kit.Volume(
			kit.Source("/data"),
			kit.Target("/data"),
			kit.Kind(vol.Bind),
			kit.Propagation(propagate.RPrivate),
			kit.SELinux(selinux.Shared),
			kit.Recursive(recursive.Disabled),
		),
		kit.Logging(kit.Driver(logdrv.JSONFile)),
	)
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	s := kit.NewService(c,
		kit.PullPolicy(pull.Always),
		kit.Deploy(
			kit.Mode(svcmode.Replicated),
			kit.EndpointMode(endpoint.VIP),
			kit.Update(kit.FailureAction(failact.Pause), kit.Order(order.StartFirst)),
		),
		kit.DependsOn(kit.Name("db"), kit.Condition(depend.Healthy)),
	)
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	_ = kit.Scheme(scheme.HTTP)
	_ = kit.Progress(progress.Plain)
	_ = kit.WaitFor(wait.NotRunning)
	_ = kit.OnExit(cascade.Stop)
	_ = kit.OCI(oci.V1_1)
	_ = kit.Usage(df.Container)
	_ = kit.Driver(netdrv.Overlay)
	_ = kit.NetworkMode(netmode.Container("abc"))
	_ = kit.IPC(ipc.Container("abc"))
	_ = kit.PID(pidns.Container("abc"))
	if role.Manager != "manager" || avail.Active != "active" || watch.Sync != "sync" {
		t.Fatal("role/avail/watch constants")
	}
}

func TestVocabularyPackagesAllUnique(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		vals []string
	}{
		{"cap", strs(cap.All())},
		{"proto", strs(proto.All())},
		{"restart", strs(restart.All())},
		{"depend", strs(depend.All())},
		{"vol", strs(vol.All())},
		{"wait", strs(wait.All())},
		{"pull", strs(pull.All())},
		{"watch", strs(watch.All())},
		{"role", strs(role.All())},
		{"ipc", strs(ipc.All())},
		{"pidns", strs(pidns.All())},
		{"uts", strs(uts.All())},
		{"userns", strs(userns.All())},
		{"cgroupns", strs(cgroupns.All())},
		{"isolation", strs(isolation.All())},
		{"netmode", strs(netmode.All())},
		{"propagate", strs(propagate.All())},
		{"progress", strs(progress.All())},
		{"scheme", strs(scheme.All())},
		{"signal", strs(signal.All())},
		{"svcmode", strs(svcmode.All())},
		{"portmode", strs(portmode.All())},
		{"failact", strs(failact.All())},
		{"order", strs(order.All())},
		{"endpoint", strs(endpoint.All())},
		{"logdrv", strs(logdrv.All())},
		{"netdrv", strs(netdrv.All())},
		{"selinux", strs(selinux.All())},
		{"recursive", strs(recursive.All())},
		{"avail", strs(avail.All())},
		{"oci", strs(oci.All())},
		{"df", strs(df.All())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if len(tt.vals) == 0 {
				t.Fatal("empty vocabulary")
			}
			seen := map[string]struct{}{}
			for _, v := range tt.vals {
				if v == "" {
					t.Fatal("empty value")
				}
				if _, ok := seen[v]; ok {
					t.Fatalf("duplicate %q", v)
				}
				seen[v] = struct{}{}
			}
		})
	}
}

func TestCascadeVocabularyUnique(t *testing.T) {
	t.Parallel()
	seen := map[cascade.Action]struct{}{}
	for _, v := range cascade.All() {
		if _, ok := seen[v]; ok {
			t.Fatalf("duplicate %v", v)
		}
		seen[v] = struct{}{}
	}
	if len(cascade.All()) != 3 {
		t.Fatalf("cascade.All = %v", cascade.All())
	}
	if cascade.Ignore != kit.CascadeIgnore || cascade.Stop != kit.CascadeStop || cascade.Fail != kit.CascadeFail {
		t.Fatal("cascade aliases")
	}
}

func TestDFVocabularyAliases(t *testing.T) {
	t.Parallel()
	if df.Container != kit.UsageContainer || df.Image != kit.UsageImage || df.Volume != kit.UsageVolume || df.BuildCache != kit.UsageBuildCache {
		t.Fatal("df aliases")
	}
	if len(df.All()) != 4 {
		t.Fatalf("df.All = %v", df.All())
	}
}

func strs[T ~string](in []T) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	return out
}
