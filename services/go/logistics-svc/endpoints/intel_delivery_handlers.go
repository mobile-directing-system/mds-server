package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"net/http"
)

// handleMarkIntelDeliveryAttemptAsDeliveredStore are the dependencies needed
// for handleMarkIntelDeliveryAttemptAsDelivered.
type handleMarkIntelDeliveryAttemptAsDeliveredStore interface {
	MarkIntelDeliveryAttemptAsDelivered(ctx context.Context, attemptID uuid.UUID, by uuid.NullUUID) error
}

// handleMarkIntelDeliveryAttemptAsDelivered marks the intel-delivery with the given id
// as delivered.
func handleMarkIntelDeliveryAttemptAsDelivered(s handleMarkIntelDeliveryAttemptAsDeliveredStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract attempt id.
		attemptIDStr := c.Param("attemptID")
		attemptID, err := uuid.FromString(attemptIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse delivery id", meh.Details{"was": attemptIDStr})
		}
		// Mark.
		var by uuid.NullUUID
		by = nulls.NewUUID(token.UserID)
		err = s.MarkIntelDeliveryAttemptAsDelivered(c.Request.Context(), attemptID, by)
		if err != nil {
			return meh.Wrap(err, "mark intel-delivery-attempt as delivered", meh.Details{
				"attempt_id": attemptID,
				"by":         by,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleMarkIntelDeliveryAsDeliveredStore are the dependencies needed for
// handleMarkIntelDeliveryAsDelivered.
type handleMarkIntelDeliveryAsDeliveredStore interface {
	MarkIntelDeliveryAsDelivered(ctx context.Context, deliveryID uuid.UUID, by uuid.NullUUID) error
}

// handleMarkIntelDeliveryAsDelivered marks the intel-delivery with the given id
// as delivered.
func handleMarkIntelDeliveryAsDelivered(s handleMarkIntelDeliveryAsDeliveredStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract delivery id.
		deliveryIDStr := c.Param("deliveryID")
		deliveryID, err := uuid.FromString(deliveryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse delivery id", meh.Details{"was": deliveryIDStr})
		}
		// Mark.
		var by uuid.NullUUID
		by = nulls.NewUUID(token.UserID)
		err = s.MarkIntelDeliveryAsDelivered(c.Request.Context(), deliveryID, by)
		if err != nil {
			return meh.Wrap(err, "mark intel-delivery as delivered", meh.Details{
				"delivery_id": deliveryID,
				"by":          by,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}
