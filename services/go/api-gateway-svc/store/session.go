package store

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/redisutil"
)

// UserIDBySessionToken returns the user id for the given session token. If the
// token was not found, a meh.ErrNotFound will be returned.
func (m *Mall) UserIDBySessionToken(ctx context.Context, token string) (uuid.UUID, error) {
	userIDRaw, err := m.redis.Get(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token)).Result()
	if err != nil {
		if err != redis.Nil {
			return uuid.Nil, meh.NewInternalErrFromErr(err, "lookup session token in redis", nil)
		}
		// Not found.
		return uuid.Nil, meh.NewNotFoundErr("not found", nil)
	}
	// Parse.
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "parse raw user id", meh.Details{"raw": userIDRaw})
	}
	return userID, nil
}

// StoreUserIDBySessionToken stores the given user id for the token.
func (m *Mall) StoreUserIDBySessionToken(ctx context.Context, token string, userID uuid.UUID) error {
	err := m.redis.Set(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token), userID.String(), 0).Err()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "set session token in redis", meh.Details{"user_id": userID})
	}
	return nil
}

// GetAndDeleteUserIDBySessionToken gets and then deletes the mapping of the
// given session token to a user id.
func (m *Mall) GetAndDeleteUserIDBySessionToken(ctx context.Context, token string) (uuid.UUID, error) {
	userIDRaw, err := m.redis.GetDel(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token)).Result()
	if err != nil {
		if err != redis.Nil {
			return uuid.Nil, meh.NewInternalErrFromErr(err, "get and delete user id by session token in redis", nil)
		}
		// Not found.
		return uuid.Nil, meh.NewNotFoundErr("not found", nil)
	}
	// Parse.
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "parse raw user id", meh.Details{"raw": userIDRaw})
	}
	return userID, nil
}
