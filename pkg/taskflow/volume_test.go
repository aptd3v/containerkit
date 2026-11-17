package taskflow

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	cerrdefs "github.com/containerd/errdefs"
)

func TestVolumeNamesDifferAndStayPrivate(t *testing.T) {
	eng := &scriptedEngine{}
	eng.rememberVolume("data", map[string]string{"keep": "yes"})
	f := New("priv", newClient(eng, &fakeComposer{}))
	one := f.Volume("data")
	two := f.Volume("data")
	if one == nil || two == nil || one.Name() == two.Name() || one.Name() == "data" || two.Name() == "data" {
		t.Fatalf("%v %v", one, two)
	}
	f.Container("job", kit.NewContainer(kit.Image("app"), kit.Name("job")), nil)
	if err := f.Run(context.Background(), discardRun()); err != nil {
		t.Fatal(err)
	}
	if len(eng.volCreates) != 2 {
		t.Fatalf("creates %+v", eng.volCreates)
	}
	for _, created := range eng.volCreates {
		if created.Labels[LabelFlow] != "priv" || created.Labels[LabelRun] == "" {
			t.Fatalf("labels %+v", created.Labels)
		}
	}
	eng.volumes[one.Name()][LabelRun] = "other"
	if err := f.Remove(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := eng.volumes["data"]; !ok {
		t.Fatal("sentinel removed")
	}
	if _, ok := eng.volumes[one.Name()]; !ok {
		t.Fatal("mismatched volume removed")
	}
	if _, ok := eng.volumes[two.Name()]; ok {
		t.Fatal("run volume remains")
	}
	var containerAt, volumeAt int
	for i, ev := range eng.events {
		if strings.HasPrefix(ev, "container:") {
			containerAt = i
		}
		if ev == "volume:"+two.Name() {
			volumeAt = i
		}
	}
	if containerAt > volumeAt {
		t.Fatalf("order %v", eng.events)
	}
}

func TestVolumeRemoveReportsInspectAndUse(t *testing.T) {
	eng := &scriptedEngine{volumeRemoveErr: errors.New("in use")}
	f := New("use", newClient(eng, nil))
	vol := f.Volume("data")
	if err := f.Run(context.Background(), discardRun()); err != nil {
		t.Fatal(err)
	}
	if err := f.Remove(context.Background()); err == nil || !strings.Contains(err.Error(), "in use") {
		t.Fatal(err)
	}
	delete(eng.volumes, vol.Name())
	eng.volumeRemoveErr = nil
	if err := f.Remove(context.Background()); err != nil {
		t.Fatal(err)
	}
	eng.volumeInspectErr = errors.New("inspect")
	f = New("inspect", newClient(eng, nil))
	f.Volume("data")
	if err := f.Run(context.Background(), discardRun()); err != nil {
		t.Fatal(err)
	}
	if err := f.Remove(context.Background()); err == nil || !strings.Contains(err.Error(), "inspect") {
		t.Fatal(err)
	}
}

func TestVolumeCreateFailures(t *testing.T) {
	f := New("empty", newClient(&scriptedEngine{}, nil))
	if f.Volume(" ") != nil {
		t.Fatal("blank")
	}
	if err := f.Run(context.Background(), discardRun()); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatal(err)
	}

	prev := runIDEntropy
	runIDEntropy = bytes.NewReader(nil)
	t.Cleanup(func() { runIDEntropy = prev })
	f = New("entropy", newClient(&scriptedEngine{}, nil))
	if f.Volume("data") != nil {
		t.Fatal("entropy")
	}
	if err := f.Run(context.Background(), discardRun()); err == nil {
		t.Fatal("missing entropy error")
	}
	runIDEntropy = prev

	f = New("nocli", nil)
	f.Volume("data")
	if err := f.Run(context.Background(), discardRun()); !errors.Is(err, ErrNoClient) {
		t.Fatal(err)
	}
	eng := &scriptedEngine{volumeCreateErr: errors.New("create")}
	f = New("create", newClient(eng, nil))
	f.Volume("data")
	f.Task("later", func(context.Context) error { return errors.New("ran") })
	if err := f.Run(context.Background(), discardRun()); err == nil || strings.Contains(err.Error(), "ran") {
		t.Fatal(err)
	}
}

func TestImportVolumes(t *testing.T) {
	_, err := Import([]byte(`{"name":"app","volumes":["data","data"],"steps":[]}`), nil)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatal(err)
	}
	_, err = Import([]byte(`{"name":"app","volumes":[""],"steps":[]}`), nil)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatal(err)
	}
	_, err = Import([]byte(`{"name":"app","steps":[{"name":"job","op":"containerRun","options":{"image":"app","volume":{"source":"volume:data","target":"/data"}}}]}`), nil)
	if err == nil || !strings.Contains(err.Error(), "unknown volume") {
		t.Fatal(err)
	}
	eng := &scriptedEngine{}
	doc := `{"name":"app","volumes":["data"],"steps":[{"name":"job","op":"containerRun","options":{"image":"app","volume":{"source":"volume:data","target":"/data"}}}]}`
	f, err := Import([]byte(doc), newClient(eng, nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(context.Background(), discardRun()); err != nil {
		t.Fatal(err)
	}
	if len(eng.volCreates) != 1 || eng.volCreates[0].Name == "data" {
		t.Fatalf("creates %+v", eng.volCreates)
	}
	if eng.created.HostConfig == nil || len(eng.created.HostConfig.Mounts) != 1 {
		t.Fatalf("mounts %+v", eng.created.HostConfig)
	}
	if eng.created.HostConfig.Mounts[0].Source != eng.volCreates[0].Name {
		t.Fatalf("mount %s volume %s", eng.created.HostConfig.Mounts[0].Source, eng.volCreates[0].Name)
	}

	if err := rewriteVolumeOptions(nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := rewriteVolumeOptions(map[string]any{}, nil); err != nil {
		t.Fatal(err)
	}
	if err := rewriteVolumeOptions(map[string]any{"volume": true}, nil); err != nil {
		t.Fatal(err)
	}
	if err := rewriteVolumeOptions(map[string]any{"volume": []any{"skip"}}, nil); err != nil {
		t.Fatal(err)
	}
	item := map[string]any{"source": "volume:data", "target": "/data"}
	if err := rewriteVolumeOptions(map[string]any{"volume": []any{item}}, map[string]string{"data": "data-abc"}); err != nil || item["source"] != "data-abc" {
		t.Fatal(err)
	}
	gone := map[string]any{"source": "volume:gone"}
	if err := rewriteVolumeOptions(map[string]any{"volume": []any{gone}}, map[string]string{}); err == nil {
		t.Fatal("array unknown")
	}
	kept := map[string]any{"source": "data"}
	if err := rewriteVolumeOptions(map[string]any{"volume": kept}, nil); err != nil || kept["source"] != "data" {
		t.Fatal(err)
	}
	items := []any{map[string]any{"source": "volume:data"}, "skip"}
	if err := rewriteVolumeOptions(map[string]any{"volume": items}, map[string]string{"data": "data-phys"}); err != nil {
		t.Fatal(err)
	}
	if items[0].(map[string]any)["source"] != "data-phys" {
		t.Fatal(items[0])
	}
	if (*Volume)(nil).Name() != "" {
		t.Fatal("nil volume")
	}
	if (*Flow)(nil).volumeList() != nil || (*Flow)(nil).volumeNames() != nil {
		t.Fatal("nil flow volumes")
	}
	if err := New("idle", nil).removeVolumes(context.Background(), nil, ""); err != nil {
		t.Fatal(err)
	}
	(*Task)(nil).noteConfig(errors.New("x"))
	(*Task)(nil).noteConfig(nil)
	if (*Task)(nil).configErr() != nil {
		t.Fatal("nil config")
	}
	missing := map[string]any{"source": "volume:gone"}
	if err := rewriteVolumeOptions(map[string]any{"volume": missing}, map[string]string{}); err == nil {
		t.Fatal("missing")
	}
	_ = cerrdefs.ErrNotFound
}
