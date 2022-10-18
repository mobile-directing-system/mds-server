package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"net/http"
	"time"
)

// publicAcceptedIntelDeliveryAttempt is the public representation of
// store.AcceptedIntelDeliveryAttempt.
type publicAcceptedIntelDeliveryAttempt struct {
	ID              uuid.UUID    `json:"id"`
	Intel           uuid.UUID    `json:"intel"`
	IntelOperation  uuid.UUID    `json:"intel_operation"`
	IntelImportance int          `json:"intel_importance"`
	AssignedTo      uuid.UUID    `json:"assigned_to"`
	AssignedToLabel string       `json:"assigned_to_label"`
	Delivery        uuid.UUID    `json:"delivery"`
	Channel         uuid.UUID    `json:"channel"`
	CreatedAt       time.Time    `json:"created_at"`
	StatusTS        time.Time    `json:"status_ts"`
	Note            nulls.String `json:"note"`
	AcceptedAt      time.Time    `json:"accepted_at"`
}

// publicAcceptedIntelDeliveryAttemptFromStore maps
// store.AcceptedIntelDeliveryAttempt to publicAcceptedIntelDeliveryAttempt.
func publicAcceptedIntelDeliveryAttemptFromStore(s store.AcceptedIntelDeliveryAttempt) publicAcceptedIntelDeliveryAttempt {
	return publicAcceptedIntelDeliveryAttempt{
		ID:              s.ID,
		Intel:           s.Intel,
		IntelOperation:  s.IntelOperation,
		IntelImportance: s.IntelImportance,
		AssignedTo:      s.AssignedTo,
		AssignedToLabel: s.AssignedToLabel,
		Delivery:        s.Delivery,
		Channel:         s.Channel,
		CreatedAt:       s.CreatedAt,
		StatusTS:        s.StatusTS,
		Note:            s.Note,
		AcceptedAt:      s.AcceptedAt,
	}
}

// handleGetNextRadioDeliveryStore are the dependencies needed for
// handleGetNextRadioDelivery.
type handleGetNextRadioDeliveryStore interface {
	PickUpNextRadioDelivery(ctx context.Context, operationID uuid.UUID, by uuid.UUID) (store.AcceptedIntelDeliveryAttempt, bool, error)
}

// handleGetNextRadioDelivery picks up the next radio delivery. If none is
// found, a http.StatusNoContent will be returned.
func handleGetNextRadioDelivery(s handleGetNextRadioDeliveryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract operation id.
		operationIDStr := c.Param("operationID")
		operationID, err := uuid.FromString(operationIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse operation id", meh.Details{"was": operationIDStr})
		}
		// Check permissions.
		err = auth.AssurePermission(token, permission.DeliverAnyRadioDelivery())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Pick up next.
		sAttempt, ok, err := s.PickUpNextRadioDelivery(c.Request.Context(), operationID, token.UserID)
		if err != nil {
			return meh.Wrap(err, "pick up next radio delivery", meh.Details{
				"operation_id": operationID,
				"by":           token.UserID,
			})
		}
		if !ok {
			c.Status(http.StatusNoContent)
			return nil
		}
		c.JSON(http.StatusOK, publicAcceptedIntelDeliveryAttemptFromStore(sAttempt))
		return nil
	}
}

// handleReleasePickedUpRadioDeliveryStore are the dependencies needed for
// handleReleasePickedUpRadioDelivery.
type handleReleasePickedUpRadioDeliveryStore interface {
	ReleasePickedUpRadioDelivery(ctx context.Context, attemptID uuid.UUID, limitToPickedUpBy uuid.NullUUID) error
}

// handleReleasePickedUpRadioDelivery releases the given picked up radio
// delivery.
func handleReleasePickedUpRadioDelivery(s handleReleasePickedUpRadioDeliveryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Check permissions.
		deliveryGranted, err := auth.HasPermission(token, permission.DeliverAnyRadioDelivery())
		if err != nil {
			return meh.Wrap(err, "check deliver-permission", nil)
		}
		manageDeliveriesGranted, err := auth.HasPermission(token, permission.ManageAnyRadioDelivery())
		if err != nil {
			return meh.Wrap(err, "check manage-permission", nil)
		}
		if !deliveryGranted && !manageDeliveriesGranted {
			return meh.NewForbiddenErr("missing permission", nil)
		}
		limitToPickedUpBy := nulls.NewUUID(token.UserID)
		if manageDeliveriesGranted {
			limitToPickedUpBy = uuid.NullUUID{}
		}
		// Extract attempt id.
		attemptIDStr := c.Param("attemptID")
		attemptID, err := uuid.FromString(attemptIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse attempt id", meh.Details{"was": attemptIDStr})
		}
		// Release.
		err = s.ReleasePickedUpRadioDelivery(c.Request.Context(), attemptID, limitToPickedUpBy)
		if err != nil {
			return meh.Wrap(err, "release picked up radio delivery", meh.Details{
				"attempt_id":            attemptID,
				"limit_to_picked_up_by": limitToPickedUpBy,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// publicFinishRadioDeliveryDetails is the expected payload in
// handleFinishRadioDelivery with details regarding finishing a radio delivery.
type publicFinishRadioDeliveryDetails struct {
	Success bool   `json:"success"`
	Note    string `json:"note"`
}

// handleFinishRadioDeliveryStore are the dependencies needed for
// handleFinishRadioDelivery.
type handleFinishRadioDeliveryStore interface {
	FinishRadioDelivery(ctx context.Context, attemptID uuid.UUID, success bool, note string, limitToPickedUpBy uuid.NullUUID) error
}

// handleFinishRadioDelivery finishes the radio delivery for the given attempt.
func handleFinishRadioDelivery(s handleFinishRadioDeliveryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Check permissions.
		deliveryGranted, err := auth.HasPermission(token, permission.DeliverAnyRadioDelivery())
		if err != nil {
			return meh.Wrap(err, "check deliver-permission", nil)
		}
		manageDeliveriesGranted, err := auth.HasPermission(token, permission.ManageAnyRadioDelivery())
		if err != nil {
			return meh.Wrap(err, "check manage-permission", nil)
		}
		if !deliveryGranted && !manageDeliveriesGranted {
			return meh.NewForbiddenErr("missing permission", nil)
		}
		limitToPickedUpBy := nulls.NewUUID(token.UserID)
		if manageDeliveriesGranted {
			limitToPickedUpBy = uuid.NullUUID{}
		}
		// Extract attempt id.
		attemptIDStr := c.Param("attemptID")
		attemptID, err := uuid.FromString(attemptIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse attempt id", meh.Details{"was": attemptIDStr})
		}
		// Parse body.
		var finishDetails publicFinishRadioDeliveryDetails
		err = c.BindJSON(&finishDetails)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse body", nil)
		}
		// Finish.
		err = s.FinishRadioDelivery(c.Request.Context(), attemptID, finishDetails.Success, finishDetails.Note, limitToPickedUpBy)
		if err != nil {
			return meh.Wrap(err, "finish radio delivery", meh.Details{
				"attempt_id":            attemptID,
				"success":               finishDetails.Success,
				"note":                  finishDetails.Note,
				"limit_to_picked_up_by": limitToPickedUpBy,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}
