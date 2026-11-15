package kit

import (
	"fmt"

	engineclient "github.com/moby/moby/client"
)

type clientAct struct {
	opts    []engineclient.Opt
	headers map[string]string
	tls     tlsAct
	tlsSet  bool
}

type tlsAct struct {
	ca   string
	cert string
	key  string
}

func (a *clientAct) add(opt engineclient.Opt) {
	a.opts = append(a.opts, opt)
}

func (a clientAct) sdk() []engineclient.Opt {
	out := append([]engineclient.Opt(nil), a.opts...)
	if len(a.headers) > 0 {
		out = append(out, engineclient.WithHTTPHeaders(a.headers))
	}
	if a.tlsSet {
		out = append(out, engineclient.WithTLSClientConfig(a.tls.ca, a.tls.cert, a.tls.key))
	}
	return out
}

// NewEngine constructs an Engine client. Prefer [Connect] when the program
// also needs Compose. With no options it reads DOCKER_HOST, DOCKER_API_VERSION,
// and TLS settings from the environment ([FromEnv]).
func NewEngine(opts ...ClientOption) (Engine, error) {
	if len(opts) == 0 {
		opts = []ClientOption{FromEnv()}
	}
	var act clientAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyClient(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("engine", err)
	}
	cli, err := engineclient.New(act.sdk()...)
	if err != nil {
		return nil, cfgErr("engine", "client", "NewEngine", err)
	}
	return cli, nil
}

var _ Engine = (*engineclient.Client)(nil)
