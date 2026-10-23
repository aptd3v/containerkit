package client

import (
	"encoding/json"
	"reflect"
)

// convert copies src into a new T by a JSON round-trip.
//
// containerkit stores Docker API types from github.com/docker/docker/api/types.
// The wrapped Moby client expects the same shapes from github.com/moby/moby/api/types.
// Those packages define distinct Go types, so the fields are copied through JSON.
// Only exported fields that survive encoding/json are transferred, and the JSON
// tags on both sides must agree. Marshal and unmarshal errors are returned as-is.
//
// A nil interface or a nil pointer returns (nil, nil) so an unset config stays unset.
func convert[T any](src any) (*T, error) {
	if src == nil {
		return nil, nil
	}
	// A typed nil pointer is an interface value that still holds a type.
	value := reflect.ValueOf(src)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return nil, nil
	}
	raw, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var dst T
	if err := json.Unmarshal(raw, &dst); err != nil {
		return nil, err
	}
	return &dst, nil
}
