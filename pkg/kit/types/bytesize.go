package types

import (
	composetypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/go-units"
)

// ByteSize is Compose [composetypes.UnitBytes].
type ByteSize = composetypes.UnitBytes

const (
	B   ByteSize = 1
	KiB          = ByteSize(units.KiB)
	MiB          = ByteSize(units.MiB)
	GiB          = ByteSize(units.GiB)
	TiB          = ByteSize(units.TiB)
)

// ParseByteSize parses a Docker size string.
// "100M" and "100MiB" are both 100 mebibytes.
func ParseByteSize(size string) (ByteSize, error) {
	n, err := units.RAMInBytes(size)
	if err != nil {
		return 0, err
	}
	return ByteSize(n), nil
}
