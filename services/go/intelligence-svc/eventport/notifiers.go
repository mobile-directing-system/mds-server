package eventport

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"golang.org/x/net/context"
)

// mapIntelTypeFromStore converts store.IntelType to event.IntelType.
func mapIntelTypeFromStore(s store.IntelType) (event.IntelType, error) {
	switch s {
	case store.IntelTypeAnalogRadioMessage:
		return event.IntelTypeAnalogRadioMessage, nil
	case store.IntelTypePlaintextMessage:
		return event.IntelTypePlaintextMessage, nil
	default:
		return "", meh.NewInternalErr("unsupported intel-type", meh.Details{"intel_type": s})
	}
}

// intelContentMapper unmarshals the given raw message, calls the mapper
// function and marshals back as JSON.
func intelContentMapper[From any, To any](mapFn func(From) (To, error)) func(s json.RawMessage) (json.RawMessage, error) {
	return func(s json.RawMessage) (json.RawMessage, error) {
		var f From
		err := json.Unmarshal(s, &f)
		if err != nil {
			return nil, meh.NewInternalErrFromErr(err, "unmarshal message content", meh.Details{"raw": string(s)})
		}
		mapped, err := mapFn(f)
		if err != nil {
			return nil, meh.Wrap(err, "map fn", meh.Details{"from": f})
		}
		toRaw, err := json.Marshal(mapped)
		if err != nil {
			return nil, meh.NewInternalErrFromErr(err, "marshal mapped", nil)
		}
		return toRaw, nil
	}
}

// mapIntelContentFromStore maps the content from store.Intel to its
// event-representation.
func mapIntelContentFromStore(sType store.IntelType, sContentRaw json.RawMessage) (json.RawMessage, error) {
	var mapper func(s json.RawMessage) (json.RawMessage, error)
	switch sType {
	case store.IntelTypeAnalogRadioMessage:
		mapper = intelContentMapper(mapIntelTypeAnalogRadioMessageContent)
	case store.IntelTypePlaintextMessage:
		mapper = intelContentMapper(mapIntelTypePlaintextMessageContent)
	}
	if mapper == nil {
		return nil, meh.NewInternalErr("no intel-content-mapper", meh.Details{"intel_type": sType})
	}
	mappedRaw, err := mapper(sContentRaw)
	if err != nil {
		return nil, meh.Wrap(err, "mapper fn", nil)
	}
	return mappedRaw, nil
}

// mapIntelTypeAnalogRadioMessageContent maps
// store.IntelTypeAnalogRadioMessageContent to
// event.mapIntelTypeAnalogRadioMessageContent.
func mapIntelTypeAnalogRadioMessageContent(s store.IntelTypeAnalogRadioMessageContent) (event.IntelTypeAnalogRadioMessageContent, error) {
	return event.IntelTypeAnalogRadioMessageContent{
		Channel:  s.Channel,
		Callsign: s.Callsign,
		Head:     s.Head,
		Content:  s.Content,
	}, nil
}

// mapIntelTypePlaintextMessageContent maps
// store.IntelTypePlaintextMessageContent to
// event.mapIntelTypePlaintextMessageContent.
func mapIntelTypePlaintextMessageContent(s store.IntelTypePlaintextMessageContent) (event.IntelTypePlaintextMessageContent, error) {
	return event.IntelTypePlaintextMessageContent{
		Text: s.Text,
	}, nil
}

// intelAssignmentsFromStore converts a store.IntelAssignment list to
// event.IntelAssignment list.
func intelAssignmentsFromStore(s []store.IntelAssignment) []event.IntelAssignment {
	assignments := make([]event.IntelAssignment, 0, len(s))
	for _, assignment := range s {
		assignments = append(assignments, event.IntelAssignment{
			ID: assignment.ID,
			To: assignment.To,
		})
	}
	return assignments
}

// NotifyIntelCreated notifies about created intel.
func (p *Port) NotifyIntelCreated(ctx context.Context, tx pgx.Tx, created store.Intel) error {
	mappedType, err := mapIntelTypeFromStore(created.Type)
	if err != nil {
		return meh.Wrap(err, "intel type from store", nil)
	}
	mappedContent, err := mapIntelContentFromStore(created.Type, created.Content)
	if err != nil {
		return meh.Wrap(err, "map intel-content from store", meh.Details{
			"intel_type":        created.Type,
			"intel_content_raw": string(created.Content),
		})
	}
	intelCreatedMessage := kafkautil.OutboundMessage{
		Topic:     event.IntelTopic,
		Key:       created.ID.String(),
		EventType: event.TypeIntelCreated,
		Value: event.IntelCreated{
			ID:          created.ID,
			CreatedAt:   created.CreatedAt,
			CreatedBy:   created.CreatedBy,
			Operation:   created.Operation,
			Type:        mappedType,
			Content:     mappedContent,
			SearchText:  created.SearchText,
			Importance:  created.Importance,
			IsValid:     created.IsValid,
			Assignments: intelAssignmentsFromStore(created.Assignments),
		},
	}
	err = p.writer.AddOutboxMessages(ctx, tx, intelCreatedMessage)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": intelCreatedMessage})
	}
	return nil
}

// NotifyIntelInvalidated notifies about invalidated intel.
func (p *Port) NotifyIntelInvalidated(ctx context.Context, tx pgx.Tx, intelID uuid.UUID, by uuid.UUID) error {
	intelInvalidatedMessage := kafkautil.OutboundMessage{
		Topic:     event.IntelTopic,
		Key:       intelID.String(),
		EventType: event.TypeIntelInvalidated,
		Value: event.IntelInvalidated{
			ID: intelID,
			By: by,
		},
	}
	err := p.writer.AddOutboxMessages(ctx, tx, intelInvalidatedMessage)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": intelInvalidatedMessage})
	}
	return nil
}
