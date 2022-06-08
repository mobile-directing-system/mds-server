package controller

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// adminUsername is the username of the admin. The user with that username is
// immutable.
const adminUsername = "admin"

// adminPassword is the default password for the admin account.
const adminPassword = "admin"

// Controller manages all operations regarding users.
type Controller struct {
	Logger      *zap.Logger
	DB          *pgxpool.Pool
	Mall        *store.Mall
	KafkaWriter *kafka.Writer
}

// Run the controller.
func (c *Controller) Run(lifetime context.Context) error {
	err := c.AssureAdminUser(lifetime)
	if err != nil {
		return meh.Wrap(err, "assure admin user", nil)
	}
	<-lifetime.Done()
	return nil
}

// AssureAdminUser assures that the user with the admin username exists.
func (c *Controller) AssureAdminUser(ctx context.Context) error {
	var adminUser store.User
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Check if admin user exists.
		var err error
		adminUser, err = c.Mall.UserByUsername(ctx, tx, adminUsername)
		if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
			return meh.NewInternalErrFromErr(err, "user by username", meh.Details{"admin_username": adminUser})
		}
		if err == nil {
			// Admin user exists.
			return nil
		}
		adminPassHashed, err := bcrypt.GenerateFromPassword([]byte(adminPassword), auth.BCryptHashCost)
		if err != nil {
			return meh.NewInternalErrFromErr(err, "hash password", nil)
		}
		// Admin user does not exist -> create.
		adminUser = store.User{
			Username:  adminUsername,
			FirstName: "Admin",
			LastName:  "Admin",
			IsAdmin:   true,
			Pass:      adminPassHashed,
		}
		_, err = c.Mall.CreateUser(ctx, tx, adminUser)
		if err != nil {
			return meh.Wrap(err, "create admin user", nil)
		}
		// Write events.
		err = kafkautil.WriteMessages(c.KafkaWriter, kafkautil.Message{
			Topic:     event.KafkaUsersTopic,
			Key:       adminUser.Username,
			EventType: event.TypeUserCreated,
			Value: event.UserCreated{
				Username:  adminUser.Username,
				FirstName: adminUser.FirstName,
				LastName:  adminUser.LastName,
				IsAdmin:   adminUser.IsAdmin,
				Pass:      adminPassHashed,
			},
		})
		if err != nil {
			return meh.Wrap(err, "write kafka messages", nil)
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
