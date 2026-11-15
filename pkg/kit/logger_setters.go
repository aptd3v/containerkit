package kit

import (
	"fmt"
	"strings"
)

// ColorOpt is the value returned by [Color].
// It implements [LoggerOption].
type ColorOpt struct{ v bool }

// Color enables or disables ANSI color on [NewLogger].
func Color(v bool) ColorOpt { return ColorOpt{v: v} }

func (o ColorOpt) applyLogger(l *Logger) error {
	l.color.assign(o.v)
	return nil
}

// LogOpt is the value returned by [Log].
// It implements [ComposeBuildOption], [ComposeCreateOption], [ComposeRestartOption],
// [ComposeStartOption], [ComposeStopOption], [ComposerClientOption], [DownOption],
// [EngineBuildOption], [EngineImagePushOption], [EnginePullOption], [PullOption], [PushOption],
// [UpOption], and [WatchOption].
type LogOpt struct{ l *Logger }

// Log attaches a [Logger] to Engine and Compose runtime actions.
// Container log drivers are [Logging], not this setter.
func Log(l *Logger) LogOpt { return LogOpt{l: l} }

func (o LogOpt) applyUp(a *upAct) error {
	if o.l == nil {
		return cfgErr("", "log", "Log", fmt.Errorf("%w: logger", errRequired))
	}
	a.log = o.l
	if a.start.consumer == nil {
		a.start.consumer = o.l
	}
	return nil
}
func (o LogOpt) applyDown(a *downAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyComposeBuild(a *composeBuildAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyPull(a *pullAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyPush(a *pushAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyComposeCreate(a *createAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyComposeStart(a *startAct) error {
	if err := attachLog(&a.log, o.l); err != nil {
		return err
	}
	if a.consumer == nil {
		a.consumer = o.l
	}
	return nil
}
func (o LogOpt) applyComposeStop(a *stopAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyComposeRestart(a *restartAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyWatch(a *watchAct) error {
	if err := attachLog(&a.log, o.l); err != nil {
		return err
	}
	if a.logTo == nil {
		a.logTo = o.l
	}
	return nil
}
func (o LogOpt) applyEnginePull(a *enginePullAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyEngineBuild(a *engineBuildAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyEngineImagePushOption(a *engineImagePushAct) error {
	return attachLog(&a.log, o.l)
}
func (o LogOpt) applyComposerClient(a *composerAct) error {
	if o.l == nil {
		return cfgErr("", "log", "Log", fmt.Errorf("%w: logger", errRequired))
	}
	a.log = o.l
	return nil
}

func attachLog(dst **Logger, l *Logger) error {
	if l == nil {
		return cfgErr("", "log", "Log", fmt.Errorf("%w: logger", errRequired))
	}
	*dst = l
	return nil
}

func (o OutputStreamOpt) applyLogger(l *Logger) error {
	if o.w == nil {
		return cfgErr("", "output", "OutputStream", fmt.Errorf("%w: writer", errRequired))
	}
	l.out = o.w
	return nil
}

func (o QuietOpt) applyLogger(l *Logger) error {
	l.quiet = o.v
	return nil
}

func (o TimestampsOpt) applyLogger(l *Logger) error {
	l.timestamps = o.v
	return nil
}

var (
	_ LoggerOption          = ColorOpt{}
	_ LoggerOption          = OutputStreamOpt{}
	_ LoggerOption          = QuietOpt{}
	_ LoggerOption          = TimestampsOpt{}
	_ LoggerOption          = ProgressOpt{}
	_ UpOption              = LogOpt{}
	_ DownOption            = LogOpt{}
	_ ComposeBuildOption    = LogOpt{}
	_ PullOption            = LogOpt{}
	_ PushOption            = LogOpt{}
	_ ComposeCreateOption   = LogOpt{}
	_ ComposeStartOption    = LogOpt{}
	_ ComposeStopOption     = LogOpt{}
	_ ComposeRestartOption  = LogOpt{}
	_ WatchOption           = LogOpt{}
	_ EnginePullOption      = LogOpt{}
	_ EngineBuildOption     = LogOpt{}
	_ EngineImagePushOption = LogOpt{}
	_ ComposerClientOption  = LogOpt{}
	_ ComposerClientOption  = OutputStream(nil)
)

func (o ProgressOpt) applyLogger(l *Logger) error {
	mode := strings.ToLower(strings.TrimSpace(o.s))
	switch mode {
	case "auto", "plain", "tty", "quiet", "json":
		l.mode = mode
		return nil
	case "":
		return cfgErr("", "progress", "Progress", fmt.Errorf("%w: progress mode", errEmpty))
	default:
		return cfgErr("", "progress", "Progress", fmt.Errorf("%w: progress %q", errInvalid, o.s))
	}
}
