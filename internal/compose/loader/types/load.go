package types

import (
	"context"

	"github.com/compose-spec/compose-go/v2/loader"
	composetypes "github.com/compose-spec/compose-go/v2/types"
)

// LoadWithContext reads Compose YAML from ConfigDetails and returns a Project.
func LoadWithContext(ctx context.Context, configDetails composetypes.ConfigDetails, options ...func(*Options)) (*composetypes.Project, error) {
	return loader.LoadWithContext(ctx, configDetails, options...)
}
