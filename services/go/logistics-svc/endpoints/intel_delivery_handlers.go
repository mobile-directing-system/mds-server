package endpoints

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"net/http"
	"time"
)

// publicIntelDeliveryAttempt is the public representation of
// store.IntelDeliveryAttempt.
type publicIntelDeliveryAttempt struct {
	ID        uuid.UUID                 `json:"id"`
	Delivery  uuid.UUID                 `json:"delivery"`
	Channel   uuid.UUID                 `json:"channel"`
	CreatedAt time.Time                 `json:"created_at"`
	IsActive  bool                      `json:"is_active"`
	Status    publicIntelDeliveryStatus `json:"status"`
	StatusTS  time.Time                 `json:"status_ts"`
	Note      nulls.String              `json:"note"`
}

// publicIntelDeliveryAttemptFromStore maps store.IntelDeliveryAttempt to
// publicIntelDeliveryAttempt.
func publicIntelDeliveryAttemptFromStore(s store.IntelDeliveryAttempt) (publicIntelDeliveryAttempt, error) {
	status, err := publicIntelDeliveryStatusFromStore(s.Status)
	if err != nil {
		return publicIntelDeliveryAttempt{}, meh.Wrap(err, "map status", meh.Details{"status": s.Status})
	}
	return publicIntelDeliveryAttempt{
		ID:        s.ID,
		Delivery:  s.Delivery,
		Channel:   s.Channel,
		CreatedAt: s.CreatedAt,
		IsActive:  s.IsActive,
		Status:    status,
		StatusTS:  s.StatusTS,
		Note:      s.Note,
	}, nil
}

// publicIntelDeliveryStatus is the public representation of
// store.IntelDeliveryStatus.
type publicIntelDeliveryStatus string

const (
	publicIntelDeliveryStatusOpen             publicIntelDeliveryStatus = "open"
	publicIntelDeliveryStatusAwaitingDelivery publicIntelDeliveryStatus = "awaiting-delivery"
	publicIntelDeliveryStatusDelivering       publicIntelDeliveryStatus = "delivering"
	publicIntelDeliveryStatusAwaitingAck      publicIntelDeliveryStatus = "awaiting-ack"
	publicIntelDeliveryStatusDelivered        publicIntelDeliveryStatus = "delivered"
	publicIntelDeliveryStatusTimeout          publicIntelDeliveryStatus = "timeout"
	publicIntelDeliveryStatusCanceled         publicIntelDeliveryStatus = "canceled"
	publicIntelDeliveryStatusFailed           publicIntelDeliveryStatus = "failed"
)

// publicIntelDeliveryStatusFromStore maps store.IntelDeliveryStatus to
// publicIntelDeliveryStatus.
func publicIntelDeliveryStatusFromStore(s store.IntelDeliveryStatus) (publicIntelDeliveryStatus, error) {
	switch s {
	case store.IntelDeliveryStatusOpen:
		return publicIntelDeliveryStatusOpen, nil
	case store.IntelDeliveryStatusAwaitingDelivery:
		return publicIntelDeliveryStatusAwaitingDelivery, nil
	case store.IntelDeliveryStatusDelivering:
		return publicIntelDeliveryStatusDelivering, nil
	case store.IntelDeliveryStatusAwaitingAck:
		return publicIntelDeliveryStatusAwaitingAck, nil
	case store.IntelDeliveryStatusDelivered:
		return publicIntelDeliveryStatusDelivered, nil
	case store.IntelDeliveryStatusTimeout:
		return publicIntelDeliveryStatusTimeout, nil
	case store.IntelDeliveryStatusCanceled:
		return publicIntelDeliveryStatusCanceled, nil
	case store.IntelDeliveryStatusFailed:
		return publicIntelDeliveryStatusFailed, nil
	default:
		return "", meh.NewInternalErr(fmt.Sprintf("unknown status: %v", s), nil)
	}
}

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

// handleCreateIntelDeliveryAttemptForDeliveryStore are the dependencies needed
// for handleCreateIntelDeliveryAttemptForDelivery.
type handleCreateIntelDeliveryAttemptForDeliveryStore interface {
	CreateIntelDeliveryAttempt(ctx context.Context, deliveryID uuid.UUID, channelID uuid.UUID) (store.IntelDeliveryAttempt, error)
}

// handleCreateIntelDeliveryAttemptForDelivery creates an delivery attempt for
// the intel delivery with the given id.
func handleCreateIntelDeliveryAttemptForDelivery(s handleCreateIntelDeliveryAttemptForDeliveryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Check permissions.
		err := auth.AssurePermission(token, permission.ManageIntelDelivery())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Extract ids.
		deliveryIDStr := c.Param("deliveryID")
		deliveryID, err := uuid.FromString(deliveryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse delivery id", meh.Details{"was": deliveryIDStr})
		}
		channelIDStr := c.Param("channelID")
		channelID, err := uuid.FromString(channelIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse channel id", meh.Details{"was": channelIDStr})
		}
		// Create.
		sCreatedAttempt, err := s.CreateIntelDeliveryAttempt(c.Request.Context(), deliveryID, channelID)
		if err != nil {
			return meh.Wrap(err, "create intel delivery attempt", meh.Details{
				"delivery_id": deliveryID,
				"channel_id":  channelID,
			})
		}
		pCreatedAttempt, err := publicIntelDeliveryAttemptFromStore(sCreatedAttempt)
		if err != nil {
			return meh.Wrap(err, "convert to public", meh.Details{"store_created_attempt": sCreatedAttempt})
		}
		c.JSON(http.StatusCreated, pCreatedAttempt)
		return nil
	}
}

// handleGetIntelDeliveryAttemptsByDeliveryStore are the dependencies needed for
// handleGetIntelDeliveryAttemptsByDelivery.
type handleGetIntelDeliveryAttemptsByDeliveryStore interface {
	IntelDeliveryAttemptsByDelivery(ctx context.Context, deliveryID uuid.UUID) ([]store.IntelDeliveryAttempt, error)
}

// handleGetIntelDeliveryAttemptsByDelivery retrieves the intel delivery attempts
// for the intel delivery with the given id.
func handleGetIntelDeliveryAttemptsByDelivery(s handleGetIntelDeliveryAttemptsByDeliveryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Check permissions.
		err := auth.AssurePermission(token, permission.ManageIntelDelivery())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Extract intel delivery id.
		deliveryIDStr := c.Param("deliveryID")
		deliveryID, err := uuid.FromString(deliveryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse delivery id", meh.Details{"was": deliveryIDStr})
		}
		// Retrieve.
		sAttempts, err := s.IntelDeliveryAttemptsByDelivery(c.Request.Context(), deliveryID)
		if err != nil {
			return meh.Wrap(err, "intel delivery attempts by delivery", meh.Details{"delivery_id": deliveryID})
		}
		pAttempts := make([]publicIntelDeliveryAttempt, 0, len(sAttempts))
		for _, sAttempt := range sAttempts {
			pAttempt, err := publicIntelDeliveryAttemptFromStore(sAttempt)
			if err != nil {
				return meh.Wrap(err, "convert to public", meh.Details{"store_attempt": sAttempt})
			}
			pAttempts = append(pAttempts, pAttempt)
		}
		c.JSON(http.StatusOK, pAttempts)
		return nil
	}
}
