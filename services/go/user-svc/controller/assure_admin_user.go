package controller

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
)

// adminUsername is the username of the admin. The user with that username is
// immutable.
const adminUsername = "admin"

// adminPassword is the default password for the admin account.
const adminPassword = "admin"

// AssureAdminUser assures that the user with the admin username exists.
func (c *Controller) AssureAdminUser(ctx context.Context) error {
	var adminUser store.User
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Check if admin user exists.
		var err error
		adminUser, err = c.Store.UserByUsername(ctx, tx, adminUsername)
		if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
			return meh.NewInternalErrFromErr(err, "user by username", meh.Details{"admin_username": adminUser})
		}
		if err == nil {
			// Admin user exists.
			return nil
		}
		adminPassHashed, err := auth.HashPassword(adminPassword)
		if err != nil {
			return meh.NewInternalErrFromErr(err, "hash password", nil)
		}
		// Admin user does not exist -> create.
		adminUserWithPass := store.UserWithPass{
			User: store.User{
				Username:  adminUsername,
				FirstName: "Admin",
				LastName:  "Admin",
				IsAdmin:   true,
			},
			Pass: adminPassHashed,
		}
		adminUser, err = c.Store.CreateUser(ctx, tx, adminUserWithPass)
		if err != nil {
			return meh.Wrap(err, "create admin user", nil)
		}
		// Write events.
		adminUserWithPass.User = adminUser
		err = c.Notifier.NotifyUserCreated(adminUserWithPass)
		if err != nil {
			return meh.Wrap(err, "notify user created", nil)
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
