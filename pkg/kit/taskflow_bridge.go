package kit

import "context"

// ContainerFromOptions decodes a plan container option bag.
func ContainerFromOptions(v any) (*Container, error) {
	opts, err := decodeContainerOptions(v)
	if err != nil {
		return nil, err
	}
	c := NewContainer(opts...)
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// EngineBuildOptions decodes a plan buildImage option bag.
func EngineBuildOptions(v any) ([]EngineBuildOption, error) {
	return decodeBuildImageOptions(v)
}

// UpOptions decodes a plan compose up option bag.
func UpOptions(v any) ([]UpOption, error) {
	return decodeUpOptions(v)
}

// DecodeAction checks a plan option bag for op. Unknown ops are not an error
// here; execution reports those.
func DecodeAction(op string, v any) error {
	return decodeActionOptions(op, v)
}

// ExecAction runs one plan op on cli. compose is required for compose ops.
// fields supplies positional arguments such as service, id, path, ref, and port.
func ExecAction(ctx context.Context, cli *Client, compose *Compose, op string, options map[string]any, fields map[string]any) ([]byte, error) {
	act := planAction{Op: op, Options: options, fields: fields}
	if fields != nil {
		act.Service, _ = asString(fields["service"])
		act.ID, _ = asString(fields["id"])
		act.Path, _ = asString(fields["path"])
		act.Ref, _ = asString(fields["ref"])
	}
	return execKnownAction(ctx, cli, compose, act)
}
