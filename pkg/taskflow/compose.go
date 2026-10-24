package taskflow

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aptd3v/containerkit/pkg/kit"
)

var errNilCompose = errors.New("compose is nil")

func bindCompose[T any](
	f *Flow,
	name string,
	src *T,
	writer func(*T) *io.Writer,
	call func(*kit.Compose, context.Context, *T) error,
) func(context.Context, *kit.Compose) error {
	return func(ctx context.Context, app *kit.Compose) error {
		if app == nil {
			return errNilCompose
		}
		var opt T
		if src != nil {
			opt = *src
		}
		if w := writer(&opt); w != nil {
			*w = stepWriter(f, name, *w)
		}
		return call(app, ctx, &opt)
	}
}

func composeCommand(f *Flow, name string, cmd any) (func(context.Context, *kit.Compose) error, error) {
	if cmd == nil {
		return nil, errors.New("compose command is nil")
	}
	switch o := cmd.(type) {
	case func(context.Context, *kit.Compose) error:
		if o == nil {
			return nil, errors.New("compose func is nil")
		}
		return o, nil
	case *kit.Up:
		return bindCompose(f, name, o, func(o *kit.Up) *io.Writer { return &o.Writer }, (*kit.Compose).Up), nil
	case *kit.Down:
		return bindCompose(f, name, o, func(o *kit.Down) *io.Writer { return &o.Writer }, (*kit.Compose).Down), nil
	case *kit.Logs:
		return bindCompose(f, name, o, func(o *kit.Logs) *io.Writer { return &o.Writer }, (*kit.Compose).Logs), nil
	case *kit.Kill:
		return bindCompose(f, name, o, func(o *kit.Kill) *io.Writer { return &o.Writer }, (*kit.Compose).Kill), nil
	case *kit.Ps:
		return bindCompose(f, name, o, func(o *kit.Ps) *io.Writer { return &o.Writer }, (*kit.Compose).Ps), nil
	case *kit.Start:
		return bindCompose(f, name, o, func(o *kit.Start) *io.Writer { return &o.Writer }, (*kit.Compose).Start), nil
	case *kit.Stop:
		return bindCompose(f, name, o, func(o *kit.Stop) *io.Writer { return &o.Writer }, (*kit.Compose).Stop), nil
	case *kit.Restart:
		return bindCompose(f, name, o, func(o *kit.Restart) *io.Writer { return &o.Writer }, (*kit.Compose).Restart), nil
	case *kit.ComposeBuild:
		return bindCompose(f, name, o, func(o *kit.ComposeBuild) *io.Writer { return &o.Writer }, (*kit.Compose).Build), nil
	case *kit.Pull:
		return bindCompose(f, name, o, func(o *kit.Pull) *io.Writer { return &o.Writer }, (*kit.Compose).Pull), nil
	case *kit.Exec:
		return bindCompose(f, name, o, func(o *kit.Exec) *io.Writer { return &o.Writer }, (*kit.Compose).Exec), nil
	case *kit.Attach:
		return bindCompose(f, name, o, func(o *kit.Attach) *io.Writer { return &o.Writer }, (*kit.Compose).Attach), nil
	case *kit.Commit:
		return bindCompose(f, name, o, func(o *kit.Commit) *io.Writer { return &o.Writer }, (*kit.Compose).Commit), nil
	case *kit.ComposeConfig:
		return bindCompose(f, name, o, func(o *kit.ComposeConfig) *io.Writer { return &o.Writer }, (*kit.Compose).Config), nil
	case *kit.Cp:
		return bindCompose(f, name, o, func(o *kit.Cp) *io.Writer { return &o.Writer }, (*kit.Compose).Cp), nil
	case *kit.Create:
		return bindCompose(f, name, o, func(o *kit.Create) *io.Writer { return &o.Writer }, (*kit.Compose).Create), nil
	case *kit.ComposeExport:
		return bindCompose(f, name, o, func(o *kit.ComposeExport) *io.Writer { return &o.Writer }, (*kit.Compose).Export), nil
	case *kit.Images:
		return bindCompose(f, name, o, func(o *kit.Images) *io.Writer { return &o.Writer }, (*kit.Compose).Images), nil
	case *kit.Ls:
		return bindCompose(f, name, o, func(o *kit.Ls) *io.Writer { return &o.Writer }, (*kit.Compose).Ls), nil
	case *kit.Pause:
		return bindCompose(f, name, o, func(o *kit.Pause) *io.Writer { return &o.Writer }, (*kit.Compose).Pause), nil
	case *kit.Port:
		return bindCompose(f, name, o, func(o *kit.Port) *io.Writer { return &o.Writer }, (*kit.Compose).Port), nil
	case *kit.Publish:
		return bindCompose(f, name, o, func(o *kit.Publish) *io.Writer { return &o.Writer }, (*kit.Compose).Publish), nil
	case *kit.Push:
		return bindCompose(f, name, o, func(o *kit.Push) *io.Writer { return &o.Writer }, (*kit.Compose).Push), nil
	case *kit.Rm:
		return bindCompose(f, name, o, func(o *kit.Rm) *io.Writer { return &o.Writer }, (*kit.Compose).Rm), nil
	case *kit.Run:
		return bindCompose(f, name, o, func(o *kit.Run) *io.Writer { return &o.Writer }, (*kit.Compose).Run), nil
	case *kit.Scale:
		return bindCompose(f, name, o, func(o *kit.Scale) *io.Writer { return &o.Writer }, (*kit.Compose).Scale), nil
	case *kit.Stats:
		return bindCompose(f, name, o, func(o *kit.Stats) *io.Writer { return &o.Writer }, (*kit.Compose).Stats), nil
	case *kit.Top:
		return bindCompose(f, name, o, func(o *kit.Top) *io.Writer { return &o.Writer }, (*kit.Compose).Top), nil
	case *kit.Unpause:
		return bindCompose(f, name, o, func(o *kit.Unpause) *io.Writer { return &o.Writer }, (*kit.Compose).Unpause), nil
	case *kit.Version:
		return bindCompose(f, name, o, func(o *kit.Version) *io.Writer { return &o.Writer }, (*kit.Compose).Version), nil
	case *kit.Volumes:
		return bindCompose(f, name, o, func(o *kit.Volumes) *io.Writer { return &o.Writer }, (*kit.Compose).Volumes), nil
	case *kit.Wait:
		return bindCompose(f, name, o, func(o *kit.Wait) *io.Writer { return &o.Writer }, (*kit.Compose).Wait), nil
	case *kit.Watch:
		return bindCompose(f, name, o, func(o *kit.Watch) *io.Writer { return &o.Writer }, (*kit.Compose).Watch), nil
	case *kit.Viz:
		return bindCompose(f, name, o, func(o *kit.Viz) *io.Writer { return &o.Writer }, (*kit.Compose).Viz), nil
	case *kit.Generate:
		return bindCompose(f, name, o, func(o *kit.Generate) *io.Writer { return &o.Writer }, (*kit.Compose).Generate), nil
	case *kit.BridgeConvert:
		return bindCompose(f, name, o, func(o *kit.BridgeConvert) *io.Writer { return &o.Writer }, (*kit.Compose).BridgeConvert), nil
	case *kit.BridgeTransformationsList:
		return bindCompose(f, name, o, func(o *kit.BridgeTransformationsList) *io.Writer { return &o.Writer }, (*kit.Compose).BridgeTransformationsList), nil
	case *kit.BridgeTransformationsCreate:
		return bindCompose(f, name, o, func(o *kit.BridgeTransformationsCreate) *io.Writer { return &o.Writer }, (*kit.Compose).BridgeTransformationsCreate), nil
	default:
		return nil, fmt.Errorf("unsupported compose command %T; use a *kit option or func(context.Context, *kit.Compose) error", cmd)
	}
}
