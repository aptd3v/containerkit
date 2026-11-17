package kit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"

	engineclient "github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// EngineImageInspectOption configures [InspectImage].
type EngineImageInspectOption interface {
	applyEngineImageInspectOption(*engineImageInspectAct) error
}

type engineImageInspectAct struct {
	opts []engineclient.ImageInspectOption
}

// EngineImageHistoryOption configures [HistoryImage].
type EngineImageHistoryOption interface {
	applyEngineImageHistoryOption(*engineImageHistoryAct) error
}

type engineImageHistoryAct struct {
	opts []engineclient.ImageHistoryOption
}

// EngineImageLoadOption configures [LoadImage].
type EngineImageLoadOption interface {
	applyEngineImageLoadOption(*engineImageLoadAct) error
}

type engineImageLoadAct struct {
	quiet     bool
	platforms []ocispec.Platform
	log       *Logger
}

// EngineImageSaveOption configures [SaveImages].
type EngineImageSaveOption interface {
	applyEngineImageSaveOption(*engineImageSaveAct) error
}

type engineImageSaveAct struct {
	platforms []ocispec.Platform
}

// EngineImageAttestOption configures [AttestImage].
type EngineImageAttestOption interface {
	applyEngineImageAttestOption(*engineImageAttestAct) error
}

type engineImageAttestAct struct {
	opts []engineclient.ImageAttestationsOption
}

// ManifestsOpt is the value returned by [Manifests].
// It implements [EngineImageInspectOption].
type ManifestsOpt struct{ v bool }

// Manifests requests image manifests on [InspectImage].
func Manifests(v bool) ManifestsOpt { return ManifestsOpt{v: v} }

func (o ManifestsOpt) applyEngineImageInspectOption(a *engineImageInspectAct) error {
	a.opts = append(a.opts, engineclient.ImageInspectWithManifests(o.v))
	return nil
}

// ImagePlatformOpt is the value returned by [ImagePlatform].
// It implements [EngineImageAttestOption], [EngineImageHistoryOption],
// [EngineImageInspectOption], [EngineImageLoadOption], and [EngineImageSaveOption].
type ImagePlatformOpt struct{ p ocispec.Platform }

// ImagePlatform selects a platform for image inspect, history, load, save, or attestations.
func ImagePlatform(p ocispec.Platform) ImagePlatformOpt { return ImagePlatformOpt{p: p} }

func (o ImagePlatformOpt) applyEngineImageInspectOption(a *engineImageInspectAct) error {
	p := o.p
	a.opts = append(a.opts, engineclient.ImageInspectWithPlatform(&p))
	return nil
}
func (o ImagePlatformOpt) applyEngineImageHistoryOption(a *engineImageHistoryAct) error {
	a.opts = append(a.opts, engineclient.ImageHistoryWithPlatform(o.p))
	return nil
}
func (o ImagePlatformOpt) applyEngineImageLoadOption(a *engineImageLoadAct) error {
	a.platforms = append(a.platforms, o.p)
	return nil
}
func (o ImagePlatformOpt) applyEngineImageSaveOption(a *engineImageSaveAct) error {
	a.platforms = append(a.platforms, o.p)
	return nil
}
func (o ImagePlatformOpt) applyEngineImageAttestOption(a *engineImageAttestAct) error {
	a.opts = append(a.opts, engineclient.ImageAttestationsWithPlatform(o.p))
	return nil
}

// RawResponseOpt is the value returned by [RawResponse].
// It implements [EngineImageInspectOption].
type RawResponseOpt struct{ buf *bytes.Buffer }

// RawResponse stores the raw inspect JSON in buf on [InspectImage].
func RawResponse(buf *bytes.Buffer) RawResponseOpt { return RawResponseOpt{buf: buf} }

func (o RawResponseOpt) applyEngineImageInspectOption(a *engineImageInspectAct) error {
	if o.buf == nil {
		return cfgErr("", "raw", "RawResponse", fmt.Errorf("%w: buffer", errRequired))
	}
	a.opts = append(a.opts, engineclient.ImageInspectWithRawResponse(o.buf))
	return nil
}

func (o QuietOpt) applyEngineImageLoadOption(a *engineImageLoadAct) error {
	a.quiet = o.v
	return nil
}

// PredicateTypesOpt is the value returned by [PredicateTypes].
// It implements [EngineImageAttestOption].
type PredicateTypesOpt struct{ types []string }

// PredicateTypes filters image attestations by in-toto predicate type URIs.
func PredicateTypes(types ...string) PredicateTypesOpt {
	return PredicateTypesOpt{types: append([]string(nil), types...)}
}

func (o PredicateTypesOpt) applyEngineImageAttestOption(a *engineImageAttestAct) error {
	a.opts = append(a.opts, engineclient.ImageAttestationsWithPredicateTypes(o.types...))
	return nil
}

// StatementOpt is the value returned by [Statement].
// It implements [EngineImageAttestOption].
type StatementOpt struct{ v bool }

// Statement includes the verbatim in-toto statement body on [AttestImage].
func Statement(v bool) StatementOpt { return StatementOpt{v: v} }

func (o StatementOpt) applyEngineImageAttestOption(a *engineImageAttestAct) error {
	if o.v {
		a.opts = append(a.opts, engineclient.ImageAttestationsWithStatement())
	}
	return nil
}

// InspectImage wraps the corresponding Engine client method.
func InspectImage(ctx context.Context, api Engine, ref string, opts ...EngineImageInspectOption) (ImageInspectResult, error) {
	if err := engineAPI(api, "InspectImage"); err != nil {
		return ImageInspectResult{}, err
	}
	if ref == "" {
		return ImageInspectResult{}, cfgErr("engine", "ref", "InspectImage", fmt.Errorf("%w: image reference", errRequired))
	}
	var act engineImageInspectAct
	if err := applyEngineOptList("inspectimage", opts, func(o EngineImageInspectOption) error {
		return o.applyEngineImageInspectOption(&act)
	}); err != nil {
		return ImageInspectResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImageInspect(ctx, ref, act.opts...)
	if err != nil {
		return ImageInspectResult{}, cfgErr("engine", "inspectimage", "InspectImage", err)
	}
	return out, nil
}

// HistoryImage wraps the corresponding Engine client method.
func HistoryImage(ctx context.Context, api Engine, ref string, opts ...EngineImageHistoryOption) (ImageHistoryResult, error) {
	if err := engineAPI(api, "HistoryImage"); err != nil {
		return ImageHistoryResult{}, err
	}
	if ref == "" {
		return ImageHistoryResult{}, cfgErr("engine", "ref", "HistoryImage", fmt.Errorf("%w: image reference", errRequired))
	}
	var act engineImageHistoryAct
	if err := applyEngineOptList("historyimage", opts, func(o EngineImageHistoryOption) error {
		return o.applyEngineImageHistoryOption(&act)
	}); err != nil {
		return ImageHistoryResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImageHistory(ctx, ref, act.opts...)
	if err != nil {
		return ImageHistoryResult{}, cfgErr("engine", "historyimage", "HistoryImage", err)
	}
	return out, nil
}

// LoadImage loads an image from a tar stream.
// With a logger, from Log or the context, the progress body is read to the end
// and the returned reader is at EOF. Without a logger, the reader is that body.
func LoadImage(ctx context.Context, api Engine, input io.Reader, opts ...EngineImageLoadOption) (ImageLoadResult, error) {
	if err := engineAPI(api, "LoadImage"); err != nil {
		return nil, err
	}
	if input == nil {
		return nil, cfgErr("engine", "input", "LoadImage", fmt.Errorf("%w: reader", errRequired))
	}
	var act engineImageLoadAct
	if err := applyEngineOptList("loadimage", opts, func(o EngineImageLoadOption) error {
		return o.applyEngineImageLoadOption(&act)
	}); err != nil {
		return nil, wrapReceiver("engine", err)
	}
	var sdk []engineclient.ImageLoadOption
	if act.quiet {
		sdk = append(sdk, engineclient.ImageLoadWithQuiet(true))
	}
	if len(act.platforms) > 0 {
		sdk = append(sdk, engineclient.ImageLoadWithPlatforms(act.platforms...))
	}
	adoptLog(ctx, &act.log)
	const id = "load"
	if act.log != nil {
		emitPhase(act.log, KindEngine, "loading", id, StatusWorking, "")
	}
	out, err := api.ImageLoad(ctx, input, sdk...)
	if err != nil {
		failPhase(act.log, KindEngine, "loading", id, err)
		return nil, cfgErr("engine", id, "LoadImage", err)
	}
	if act.log == nil {
		return out, nil
	}
	if err := act.log.drainJSON(ctx, out, KindEngine, id); err != nil {
		failPhase(act.log, KindEngine, "loading", id, err)
		return nil, cfgErr("engine", id, "LoadImage", err)
	}
	finishPhase(act.log, KindEngine, "loaded", id)
	return io.NopCloser(bytes.NewReader(nil)), nil
}

// SaveImages wraps the corresponding Engine client method.
func SaveImages(ctx context.Context, api Engine, refs []string, opts ...EngineImageSaveOption) (ImageSaveResult, error) {
	if err := engineAPI(api, "SaveImages"); err != nil {
		return nil, err
	}
	var act engineImageSaveAct
	if err := applyEngineOptList("saveimages", opts, func(o EngineImageSaveOption) error {
		return o.applyEngineImageSaveOption(&act)
	}); err != nil {
		return nil, wrapReceiver("engine", err)
	}
	var sdk []engineclient.ImageSaveOption
	if len(act.platforms) > 0 {
		sdk = append(sdk, engineclient.ImageSaveWithPlatforms(act.platforms...))
	}
	out, err := api.ImageSave(ctx, refs, sdk...)
	if err != nil {
		return nil, cfgErr("engine", "saveimages", "SaveImages", err)
	}
	return out, nil
}

// AttestImage wraps Engine.ImageAttestations.
func AttestImage(ctx context.Context, api Engine, ref string, opts ...EngineImageAttestOption) (ImageAttestationsResult, error) {
	if err := engineAPI(api, "AttestImage"); err != nil {
		return ImageAttestationsResult{}, err
	}
	var act engineImageAttestAct
	if err := applyEngineOptList("attestimage", opts, func(o EngineImageAttestOption) error {
		return o.applyEngineImageAttestOption(&act)
	}); err != nil {
		return ImageAttestationsResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImageAttestations(ctx, ref, act.opts...)
	if err != nil {
		return ImageAttestationsResult{}, cfgErr("engine", "attestimage", "AttestImage", err)
	}
	return out, nil
}

// Dialer returns the Engine client's raw stream dialer.
func Dialer(api Engine) (func(context.Context) (net.Conn, error), error) {
	if err := engineAPI(api, "Dialer"); err != nil {
		return nil, err
	}
	return api.Dialer(), nil
}

// DialHijack opens a hijacked connection to the Engine daemon.
func DialHijack(ctx context.Context, api Engine, url, proto string, meta map[string][]string) (net.Conn, error) {
	if err := engineAPI(api, "DialHijack"); err != nil {
		return nil, err
	}
	conn, err := api.DialHijack(ctx, url, proto, meta)
	if err != nil {
		return nil, cfgErr("engine", "dialhijack", "DialHijack", err)
	}
	return conn, nil
}

var (
	_ EngineImageInspectOption = Manifests(true)
	_ EngineImageInspectOption = ImagePlatform(ocispec.Platform{})
	_ EngineImageLoadOption    = Quiet(true)
	_ EngineImageAttestOption  = Statement(true)
)
