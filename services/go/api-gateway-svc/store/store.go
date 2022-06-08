package store

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
)

// Mall holds the goqu.DialectWrapper for correct SQL-queries and the
// redis.Client for caching.
type Mall struct {
	dialect goqu.DialectWrapper
	redis   *redis.Client
}

const (
	redisSessionTokenPrefix = "session_token"
)

// NewMall creates a new Mall.
func NewMall(redisClient *redis.Client) *Mall {
	return &Mall{
		dialect: goqu.Dialect("postgres"),
		redis:   redisClient,
	}
}
