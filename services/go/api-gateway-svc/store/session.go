package store

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/redisutil"
)

// UsernameBySessionToken returns the username for the given session token. If
// the token was not found, a meh.ErrNotFound will be returned.
func (m *Mall) UsernameBySessionToken(ctx context.Context, token string) (string, error) {
	username, err := m.Redis.Get(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token)).Result()
	if err != nil {
		if err != redis.Nil {
			return "", meh.NewInternalErrFromErr(err, "lookup session token in redis", nil)
		}
		// Not found.
		return "", meh.NewNotFoundErr("not found", nil)
	}
	// Found.
	return username, nil
}

// StoreUsernameBySessionToken stores the given username for the token.
func (m *Mall) StoreUsernameBySessionToken(ctx context.Context, token string, username string) error {
	err := m.Redis.Set(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token), username, 0).Err()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "set session token in redis", meh.Details{"username": username})
	}
	return nil
}
