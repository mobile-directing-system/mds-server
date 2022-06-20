package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/redisutil"
	"time"
)

// UserIDBySessionToken returns the user id for the given session token. If the
// token was not found, a meh.ErrNotFound will be returned.
func (m *Mall) UserIDBySessionToken(ctx context.Context, txSupplier pgutil.DBTxSupplier, token string) (uuid.UUID, error) {
	userIDRaw, err := m.redis.Get(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token)).Result()
	if err != nil {
		if err != redis.Nil {
			return uuid.Nil, meh.NewInternalErrFromErr(err, "lookup session token in redis", nil)
		}
	} else {
		// Parse.
		userID, err := uuid.Parse(userIDRaw)
		if err != nil {
			return uuid.Nil, meh.NewInternalErrFromErr(err, "parse raw user id", meh.Details{"raw": userIDRaw})
		}
		return userID, nil
	}
	// Not found -> lookup in database.
	q, _, err := goqu.From(goqu.T("session_tokens")).
		Select(goqu.C("user")).
		Where(goqu.C("token").Eq(token)).ToSQL()
	if err != nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	tx, err := txSupplier.Begin(ctx)
	if err != nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "begin tx", nil)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return uuid.Nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		_ = tx.Commit(ctx)
		return uuid.Nil, meh.NewNotFoundErr("not found", nil)
	}
	var userID uuid.UUID
	err = rows.Scan(&userID)
	if err != nil {
		return uuid.Nil, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	err = tx.Commit(ctx)
	if err != nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "commit tx", nil)
	}
	// Set in cache.
	err = m.redis.Set(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token), userID.String(), 0).Err()
	if err != nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "set token in cache", nil)
	}
	return userID, nil
}

// StoreSessionTokenForUser stores the given token for the user with the given
// id.
func (m *Mall) StoreSessionTokenForUser(ctx context.Context, tx pgx.Tx, token string, userID uuid.UUID) error {
	// Save to database.
	q, _, err := goqu.Insert(goqu.T("session_tokens")).Rows(goqu.Record{
		"user":       userID,
		"token":      token,
		"created_ts": time.Now().UTC(),
	}).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// GetAndDeleteUserIDBySessionToken gets and then deletes the mapping of the
// given session token to a user id.
func (m *Mall) GetAndDeleteUserIDBySessionToken(ctx context.Context, tx pgx.Tx, token string) (uuid.UUID, error) {
	// Remove from cache.
	userIDRaw, err := m.redis.GetDel(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token)).Result()
	if err != nil && err != redis.Nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "get and delete user id by session token in redis", nil)
	}
	// Remove from database.
	q, _, err := goqu.Delete(goqu.T("session_tokens")).
		Where(goqu.C("token").Eq(token)).ToSQL()
	if err != nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return uuid.Nil, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return uuid.Nil, meh.NewInternalErr("token not found in database but in cache", nil)
	}
	// Parse.
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return uuid.Nil, meh.NewInternalErrFromErr(err, "parse raw user id", meh.Details{"raw": userIDRaw})
	}
	return userID, nil
}

// DeleteSessionTokensByUser deletes all session tokens for the given user from
// the database and from cache.
func (m *Mall) DeleteSessionTokensByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Retrieve.
	retrieveQuery, _, err := goqu.From(goqu.T("session_tokens")).
		Select(goqu.C("token")).
		Where(goqu.C("user").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "retrieve-query to sql", nil)
	}
	rows, err := tx.Query(ctx, retrieveQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "perform retrieve-query", retrieveQuery)
	}
	defer rows.Close()
	for rows.Next() {
		var token string
		err = rows.Scan(&token)
		if err != nil {
			return mehpg.NewScanRowsErr(err, "scan row", retrieveQuery)
		}
		// Delete in cache.
		err = m.redis.Del(ctx, redisutil.BuildKey(redisSessionTokenPrefix, token)).Err()
		if err != nil && err != redis.Nil {
			return meh.NewInternalErrFromErr(err, "delete token in redis", nil)
		}
	}
	// Delete in database.
	deleteQuery, _, err := goqu.Delete(goqu.T("session_tokens")).
		Where(goqu.C("user").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "delete-query to sql", nil)
	}
	_, err = tx.Exec(ctx, deleteQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "perform delete-query", deleteQuery)
	}
	return nil
}
