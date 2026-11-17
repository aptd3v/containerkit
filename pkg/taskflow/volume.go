package taskflow

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
)

const volumePrefix = "volume:"

// Volume is a run-private Engine volume. Name is the physical name allocated
// when Volume is called, before Run creates a run id.
type Volume struct {
	logical string
	name    string
}

// Name is the Engine volume name. Mount it with kit.Volume(kit.Name(name), ...).
// The logical name passed to Flow.Volume is not this string.
func (v *Volume) Name() string {
	if v == nil {
		return ""
	}
	return v.name
}

// Volume allocates a physical name for logical. Two calls, including two
// processes, get two names. Run creates the volume. Remove deletes it only
// when its containerkit.run label is this run.
func (f *Flow) Volume(logical string) *Volume {
	if f == nil {
		return nil
	}
	if strings.TrimSpace(logical) == "" {
		f.noteEdge(&terminalError{msg: "taskflow: volume name is empty"})
		return nil
	}
	name, err := physicalVolume(logical)
	if err != nil {
		f.noteEdge(&terminalError{msg: err.Error()})
		return nil
	}
	v := &Volume{logical: logical, name: name}
	f.mu.Lock()
	f.volumes = append(f.volumes, v)
	if f.volNames == nil {
		f.volNames = map[string]string{}
	}
	if _, exists := f.volNames[logical]; !exists {
		f.volNames[logical] = name
	}
	f.mu.Unlock()
	return v
}

func physicalVolume(logical string) (string, error) {
	var b [8]byte
	if _, err := io.ReadFull(runIDEntropy, b[:]); err != nil {
		return "", err
	}
	return logical + "-" + hex.EncodeToString(b[:]), nil
}

func (f *Flow) addDocumentVolumes(names []string) error {
	seen := map[string]struct{}{}
	for _, name := range names {
		if _, ok := seen[name]; ok {
			return fmt.Errorf("taskflow: duplicate volume %q", name)
		}
		seen[name] = struct{}{}
		if f.Volume(name) == nil {
			return f.foreign()
		}
	}
	return nil
}

func (f *Flow) volumeList() []*Volume {
	if f == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*Volume(nil), f.volumes...)
}

func (f *Flow) volumeNames() map[string]string {
	if f == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[string]string, len(f.volNames))
	for k, v := range f.volNames {
		out[k] = v
	}
	return out
}

func (f *Flow) createVolumes(ctx context.Context) error {
	vols := f.volumeList()
	if len(vols) == 0 {
		return nil
	}
	api := f.engine()
	if api == nil {
		return ErrNoClient
	}
	f.mu.Lock()
	runID := f.runID
	name := f.name
	f.mu.Unlock()
	for _, v := range vols {
		_, err := kit.CreateVolume(ctx, api,
			kit.Name(v.name),
			kit.Label(LabelFlow, name),
			kit.Label(LabelRun, runID),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Flow) removeVolumes(ctx context.Context, api kit.Engine, runID string) error {
	var errs []error
	for _, v := range f.volumeList() {
		inspected, err := kit.InspectVolume(ctx, api, v.name)
		if err != nil {
			if kit.IsNotFound(err) {
				continue
			}
			errs = append(errs, err)
			continue
		}
		if inspected.Volume.Labels[LabelRun] != runID {
			continue
		}
		if _, err := kit.RemoveVolume(ctx, api, v.name); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func rewriteVolumeOptions(options map[string]any, names map[string]string) error {
	if options == nil {
		return nil
	}
	raw, ok := options["volume"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case map[string]any:
		return rewriteOneVolume(v, names)
	case []any:
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if err := rewriteOneVolume(m, names); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

func rewriteOneVolume(m map[string]any, names map[string]string) error {
	src, _ := m["source"].(string)
	logical, ok := strings.CutPrefix(src, volumePrefix)
	if !ok {
		return nil
	}
	phys, found := names[logical]
	if !found {
		return fmt.Errorf("taskflow: unknown volume %q", logical)
	}
	m["source"] = phys
	return nil
}
