package kit

import (
	"fmt"
	"io"

	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	compose "github.com/docker/compose/v5/pkg/compose"
	engineclient "github.com/moby/moby/client"
)

type composerAct struct {
	client         clientAct
	log            *Logger
	out            io.Writer
	maxConcurrency int
	maxSet         bool
	dryRun         bool
}

func applyAsComposerClient(a *composerAct, opt ClientOption) error {
	return opt.applyClient(&a.client)
}

func collectComposerAct(opts []ComposerClientOption) (composerAct, error) {
	if len(opts) == 0 {
		opts = []ComposerClientOption{FromEnv()}
	}
	var act composerAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("composer", fmt.Sprintf("option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyComposerClient(&act))
	}
	if err := join(errs...); err != nil {
		return composerAct{}, err
	}
	return act, nil
}

func engineFromAct(act composerAct) (Engine, error) {
	engineOpts := act.client.sdk()
	if len(engineOpts) == 0 {
		engineOpts = []engineclient.Opt{engineclient.FromEnv}
	}
	api, err := engineclient.New(engineOpts...)
	if err != nil {
		return nil, err
	}
	return api, nil
}

func composerFromEngine(api Engine, act composerAct) (Composer, error) {
	cli, err := command.NewDockerCli(command.WithAPIClient(api))
	if err != nil {
		closeEngine(api)
		return nil, cfgErr("composer", "client", "NewComposer", err)
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		closeEngine(api)
		return nil, cfgErr("composer", "client", "NewComposer", err)
	}
	var composeOpts []compose.Option
	if act.log != nil {
		composeOpts = append(composeOpts, compose.WithEventProcessor(act.log))
	}
	if act.out != nil {
		composeOpts = append(composeOpts, compose.WithOutputStream(act.out))
	}
	if act.maxSet {
		composeOpts = append(composeOpts, compose.WithMaxConcurrency(act.maxConcurrency))
	}
	if act.dryRun {
		composeOpts = append(composeOpts, compose.WithDryRun)
	}
	svc, err := compose.NewComposeService(cli, composeOpts...)
	if err != nil {
		closeEngine(api)
		return nil, cfgErr("composer", "client", "NewComposer", err)
	}
	return svc, nil
}

func closeEngine(api Engine) {
	type closer interface{ Close() error }
	if cl, ok := any(api).(closer); ok {
		_ = cl.Close()
	}
}

// NewComposer constructs an in-process Compose SDK service (no docker CLI
// shell-out). Prefer [Connect] when the program also needs an Engine client.
// Options are the same [ClientOption] values as [NewEngine], plus [Log],
// [MaxConcurrency], [DryRun], and [OutputStream]. With none, it reads the
// daemon from the environment ([FromEnv]).
func NewComposer(opts ...ComposerClientOption) (Composer, error) {
	act, err := collectComposerAct(opts)
	if err != nil {
		return nil, wrapReceiver("composer", err)
	}
	api, err := engineFromAct(act)
	if err != nil {
		return nil, cfgErr("composer", "client", "NewComposer", err)
	}
	return composerFromEngine(api, act)
}

// MaxConcurrencyOpt is the value returned by [MaxConcurrency].
// It implements [ComposerClientOption].
type MaxConcurrencyOpt struct{ n int }

// MaxConcurrency limits concurrent Engine operations on [NewComposer].
func MaxConcurrency(n int) MaxConcurrencyOpt { return MaxConcurrencyOpt{n: n} }

func (o MaxConcurrencyOpt) applyComposerClient(a *composerAct) error {
	a.maxConcurrency = o.n
	a.maxSet = true
	return nil
}

// DryRunOpt is the value returned by [DryRun].
// It implements [ComposerClientOption].
type DryRunOpt struct{ v bool }

// DryRun configures the Compose SDK to plan without applying changes
// on [NewComposer].
func DryRun(v bool) DryRunOpt { return DryRunOpt{v: v} }

func (o DryRunOpt) applyComposerClient(a *composerAct) error {
	a.dryRun = o.v
	return nil
}

func applyEngineOptList[T any](op string, opts []T, apply func(T) error) error {
	var errs []error
	for i, opt := range opts {
		if any(opt) == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("%s.option[%d]", op, i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, apply(opt))
	}
	return join(errs...)
}
