package redisutil

import "strings"

// BuildKey returns the given segments joined by a colon.
func BuildKey(segments ...string) string {
	return strings.Join(segments, ":")
}
