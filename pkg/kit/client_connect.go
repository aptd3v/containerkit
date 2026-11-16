package kit

import (
	"context"
	"fmt"
)

// Client is one connection to a Docker daemon: Engine API and Compose runtime.
// Construct with [Connect] (dials the environment) or [NewClient] (tests and
// already-open clients).
type Client struct {
	engine   Engine
	composer Composer
}

// NewClient binds an existing Engine and Composer. [Connect] is the usual
// constructor. Like other constructors except [Connect], [NewEngine], and
// [NewComposer], NewClient does not return an error; methods fail if a client
// is missing.
func NewClient(engine Engine, composer Composer) *Client {
	return &Client{engine: engine, composer: composer}
}

// Connect dials one Engine and one in-process Composer from the same options.
// With none, it reads the daemon from the environment ([FromEnv]).
func Connect(opts ...ComposerClientOption) (*Client, error) {
	act, err := collectComposerAct(opts)
	if err != nil {
		return nil, wrapReceiver("client", err)
	}
	api, err := engineFromAct(act)
	if err != nil {
		return nil, cfgErr("client", "connect", "Connect", err)
	}
	composer, err := composerFromEngine(api, act)
	if err != nil {
		return nil, err
	}
	return &Client{engine: api, composer: composer}, nil
}

// Engine returns the Engine API client.
func (c *Client) Engine() Engine {
	if c == nil {
		return nil
	}
	return c.engine
}

// Composer returns the Compose runtime client.
func (c *Client) Composer() Composer {
	if c == nil {
		return nil
	}
	return c.composer
}

// Close closes the Engine connection. The Composer uses the same connection.
func (c *Client) Close() error {
	if c == nil || c.engine == nil {
		return nil
	}
	type closer interface{ Close() error }
	if cl, ok := any(c.engine).(closer); ok {
		return cl.Close()
	}
	return nil
}

// Loaded is a compose-go project bound to the [Client] that loaded it.
type Loaded struct {
	Project *Project
	client  *Client
}

// Load loads compose YAML through this client's Composer. Local YAML with no
// daemon is [ImportFile].
func (c *Client) Load(ctx context.Context, path string, opts ...LoadProjectOption) (*Loaded, error) {
	if c == nil || c.composer == nil {
		return nil, cfgErr("client", "load", "Load", fmt.Errorf("%w: compose client", errRequired))
	}
	if path == "" {
		return nil, cfgErr("client", "load", "Load", fmt.Errorf("%w: compose file", errEmpty))
	}
	all := make([]LoadProjectOption, 0, 1+len(opts))
	all = append(all, File(path))
	all = append(all, opts...)
	proj, err := LoadProject(ctx, c.composer, all...)
	if err != nil {
		return nil, err
	}
	return &Loaded{Project: proj, client: c}, nil
}

// Up brings a loaded project up.
func (l *Loaded) Up(ctx context.Context, opts ...UpOption) error {
	if l == nil || l.client == nil {
		return cfgErr("client", "up", "Up", fmt.Errorf("%w: compose client", errRequired))
	}
	name := "load"
	if l.Project != nil && l.Project.Name != "" {
		name = l.Project.Name
	}
	return upProject(ctx, l.client.composer, l.Project, name, opts...)
}

// Up brings a defined Compose file up through this client.
func (c *Client) Up(ctx context.Context, p *Compose, opts ...UpOption) error {
	if c == nil || c.composer == nil {
		return cfgErr("client", "up", "Up", fmt.Errorf("%w: compose client", errRequired))
	}
	return p.Up(ctx, c.composer, opts...)
}

// Run creates and starts a container on this client's Engine.
// The returned [Created] value is the container id from create, including when start fails.
func (c *Client) Run(ctx context.Context, opts ...ContainerOption) (Created, error) {
	if c == nil || c.engine == nil {
		return Created{}, cfgErr("client", "run", "Run", fmt.Errorf("%w: engine client", errRequired))
	}
	ctr := NewContainer(opts...)
	created, err := ctr.Create(ctx, c.engine)
	if err != nil {
		return Created{}, err
	}
	if err := ctr.Start(ctx, c.engine, created.ID); err != nil {
		return created, err
	}
	return created, nil
}

// Run dials [Connect], creates and starts a container, then closes the client.
func Run(ctx context.Context, opts ...ContainerOption) (Created, error) {
	c, err := Connect()
	if err != nil {
		return Created{}, err
	}
	defer c.Close()
	return c.Run(ctx, opts...)
}
