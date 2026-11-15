// Package df is the Engine DiskUsage object-type vocabulary.
package df

// Object is a [DiskUsage] type filter (container, image, volume, build-cache).
type Object string

const (
	Container  Object = "container"
	Image      Object = "image"
	Volume     Object = "volume"
	BuildCache Object = "build-cache"
)

// All returns every known disk-usage object type.
func All() []Object { return []Object{Container, Image, Volume, BuildCache} }
