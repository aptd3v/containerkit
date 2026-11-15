package kit

import "testing"

func TestSwarmInitJoinOptionsApply(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "init listen advertise datapath force",
			run: func(t *testing.T) {
				t.Helper()
				var act engineSwarmInitAct
				if err := ListenAddr("0.0.0.0:2377").applyEngineSwarmInitOption(&act); err != nil {
					t.Fatal(err)
				}
				if err := AdvertiseAddr("10.0.0.11:2377").applyEngineSwarmInitOption(&act); err != nil {
					t.Fatal(err)
				}
				if err := DataPathAddr("10.0.0.11").applyEngineSwarmInitOption(&act); err != nil {
					t.Fatal(err)
				}
				if err := ForceNewCluster(true).applyEngineSwarmInitOption(&act); err != nil {
					t.Fatal(err)
				}
				if act.opts.ListenAddr != "0.0.0.0:2377" || act.opts.AdvertiseAddr != "10.0.0.11:2377" {
					t.Fatalf("addrs = %+v", act.opts)
				}
				if act.opts.DataPathAddr != "10.0.0.11" || !act.opts.ForceNewCluster {
					t.Fatalf("datapath/force = %+v", act.opts)
				}
			},
		},
		{
			name: "join token remotes and addrs",
			run: func(t *testing.T) {
				t.Helper()
				var act engineSwarmJoinAct
				if err := ListenAddr("0.0.0.0:2377").applyEngineSwarmJoinOption(&act); err != nil {
					t.Fatal(err)
				}
				if err := AdvertiseAddr("10.0.0.12:2377").applyEngineSwarmJoinOption(&act); err != nil {
					t.Fatal(err)
				}
				if err := JoinToken("SWMTKN-1-test").applyEngineSwarmJoinOption(&act); err != nil {
					t.Fatal(err)
				}
				if err := RemoteAddrs("10.0.0.11:2377", "10.0.0.11:2378").applyEngineSwarmJoinOption(&act); err != nil {
					t.Fatal(err)
				}
				if act.opts.JoinToken != "SWMTKN-1-test" {
					t.Fatalf("token = %s", act.opts.JoinToken)
				}
				if len(act.opts.RemoteAddrs) != 2 || act.opts.RemoteAddrs[0] != "10.0.0.11:2377" {
					t.Fatalf("remotes = %#v", act.opts.RemoteAddrs)
				}
				if act.opts.ListenAddr != "0.0.0.0:2377" || act.opts.AdvertiseAddr != "10.0.0.12:2377" {
					t.Fatalf("addrs = %+v", act.opts)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
