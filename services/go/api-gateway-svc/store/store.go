package store

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
)

// Mall holds the goqu.DialectWrapper for correct SQL-queries and the
// redis.Client for caching.
type Mall struct {
	Dialect goqu.DialectWrapper
	Redis   *redis.Client
}

const (
	redisSessionTokenPrefix = "session_token"
)
