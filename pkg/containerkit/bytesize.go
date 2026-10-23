package containerkit

import "github.com/docker/go-units"

// ByteSize is a memory or rate size in bytes.
// Multiply an integer by B, KiB, MiB, GiB, or TiB.
type ByteSize int64

const (
	B   ByteSize = 1
	KiB          = 1024 * B
	MiB          = 1024 * KiB
	GiB          = 1024 * MiB
	TiB          = 1024 * GiB
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
