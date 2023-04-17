package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
)

// ActiveIntelDelivery represents an intel delivery that is currently considered
// active.
type ActiveIntelDelivery struct {
	// ID identifies the intel delivery.
	ID uuid.UUID
	// Intel is the id of the Intel to deliver.
	Intel uuid.UUID
	// To is the id of the recipient address book entry.
	To uuid.UUID
	// Note contains additional (debug) information.
	Note nulls.String
}

// ActiveIntelDeliveryAttempt represents an intel delivery attempt that is
// currently active.
type ActiveIntelDeliveryAttempt struct {
	// ID identifies the delivery attempt.
	ID uuid.UUID
	// Delivery is the id of the intel delivery, the attempt is for.
	Delivery uuid.UUID
}

// OpenIntelDeliverySummary is a summary for an intel that is open for intel
// delivery.
//
// This is used for communicating intel deliveries that have no attempts and
// therefore an attempt is expected to be created by manual intel delivery.
type OpenIntelDeliverySummary struct {
	Delivery ActiveIntelDelivery
	Intel    Intel
}

// IntelOperationByDeliveryAttempt retrieves the operation id for the intel the
// delivery associated with the given attempt is for. This is mainly used for
// reducing database calls when trying to get the operation id for change
// notifications.
func (m *Mall) IntelOperationByDeliveryAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (uuid.UUID, error) {
	q, _, err := m.dialect.From(goqu.T("active_intel_delivery_attempts")).
		InnerJoin(goqu.T("active_intel_deliveries"),
			goqu.On(goqu.I("active_intel_deliveries.id").Eq(goqu.I("active_intel_delivery_attempts.delivery")))).
		InnerJoin(goqu.T("intel"),
			goqu.On(goqu.I("intel.id").Eq(goqu.I("active_intel_deliveries.intel")))).
		Select(goqu.I("intel.operation")).
		Where(goqu.I("active_intel_delivery_attempts.id").Eq(attemptID)).ToSQL()
	if err != nil {
		return uuid.UUID{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return uuid.UUID{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return uuid.UUID{}, meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	var operationID uuid.UUID
	err = rows.Scan(&operationID)
	if err != nil {
		return uuid.UUID{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return operationID, nil
}

// IntelOperationByDelivery retrieves the operation id for the intel the
// given delivery is associated with. This is mainly used for
// reducing database calls when trying to get the operation id for change
// notifications.
func (m *Mall) IntelOperationByDelivery(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (uuid.UUID, error) {
	q, _, err := m.dialect.From(goqu.T("active_intel_deliveries")).
		InnerJoin(goqu.T("intel"),
			goqu.On(goqu.I("intel.id").Eq(goqu.I("active_intel_deliveries.intel")))).
		Select(goqu.I("intel.operation")).
		Where(goqu.I("active_intel_deliveries.id").Eq(attemptID)).ToSQL()
	if err != nil {
		return uuid.UUID{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return uuid.UUID{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return uuid.UUID{}, meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	var operationID uuid.UUID
	err = rows.Scan(&operationID)
	if err != nil {
		return uuid.UUID{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return operationID, nil
}

// CreateActiveIntelDelivery creates the given ActiveIntelDelivery in the store.
func (m *Mall) CreateActiveIntelDelivery(ctx context.Context, tx pgx.Tx, create ActiveIntelDelivery) error {
	q, _, err := m.dialect.Insert(goqu.T("active_intel_deliveries")).Rows(goqu.Record{
		"id":    create.ID,
		"intel": create.Intel,
		"to":    create.To,
		"note":  create.Note,
	}).ToSQL()
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// DeleteActiveIntelDeliveryByID deletes the ActiveIntelDelivery with the given
// id.
func (m *Mall) DeleteActiveIntelDeliveryByID(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) error {
	q, _, err := m.dialect.Delete(goqu.T("active_intel_deliveries")).
		Where(goqu.C("id").Eq(deliveryID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	return nil
}

// CreateActiveIntelDeliveryAttempt creates the given ActiveIntelDeliveryAttempt.
func (m *Mall) CreateActiveIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create ActiveIntelDeliveryAttempt) error {
	q, _, err := m.dialect.Insert(goqu.T("active_intel_delivery_attempts")).Rows(goqu.Record{
		"id":       create.ID,
		"delivery": create.Delivery,
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

// DeleteActiveIntelDeliveryAttemptByID deletes the ActiveIntelDeliveryAttempt
// with the given id.
func (m *Mall) DeleteActiveIntelDeliveryAttemptByID(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) error {
	q, _, err := m.dialect.Delete(goqu.T("active_intel_delivery_attempts")).
		Where(goqu.C("id").Eq(attemptID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	return nil
}

// IsAutoIntelDeliveryEnabledForEntry checks whether auto-intel-delivery is
// enabled for the address book entry with the given id.
func (m *Mall) IsAutoIntelDeliveryEnabledForEntry(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID) (bool, error) {
	q, _, err := m.dialect.From(goqu.T("auto_intel_delivery_address_book_entries")).
		Select(goqu.C("entry")).
		Where(goqu.C("entry").Eq(addressBookEntryID)).ToSQL()
	if err != nil {
		return false, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return false, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	return rows.Next(), nil
}

// SetAutoIntelDeliveryEnabledForEntry sets the auto-intel-delivery flag for the
// address book entry with the given id.
func (m *Mall) SetAutoIntelDeliveryEnabledForEntry(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID, enabled bool) error {
	isAlreadyEnabled, err := m.IsAutoIntelDeliveryEnabledForEntry(ctx, tx, addressBookEntryID)
	if err != nil {
		return meh.Wrap(err, "check if auto-intel-delivery is already enabled for entry", meh.Details{"entry_id": addressBookEntryID})
	}
	var q string
	if !isAlreadyEnabled {
		q, _, err = m.dialect.Insert(goqu.C("auto_intel_delivery_address_book_entries")).Rows(goqu.Record{
			"entry": addressBookEntryID,
		}).ToSQL()
		if err != nil {
			return meh.NewInternalErrFromErr(err, "insert-query to sql", nil)
		}
	} else {
		q, _, err = m.dialect.Delete(goqu.T("auto_intel_delivery_address_book_entries")).
			Where(goqu.C("entry").Eq(addressBookEntryID)).ToSQL()
		if err != nil {
			return meh.NewInternalErrFromErr(err, "delete-query to sql", nil)
		}
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	// Ignore no-rows.
	return nil
}

// OpenIntelDeliveriesByOperation retrieves the intel deliveries that are active
// but have no active delivery attempt and are not marked for
// auto-intel-delivery.
func (m *Mall) OpenIntelDeliveriesByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) ([]OpenIntelDeliverySummary, error) {
	q, _, err := m.dialect.From(goqu.T("active_intel_deliveries")).
		InnerJoin(goqu.T("intel"),
			goqu.On(goqu.I("active_intel_deliveries.intel").Eq(goqu.I("intel.id")))).
		LeftJoin(goqu.T("active_intel_delivery_attempts"),
			goqu.On(goqu.I("active_intel_deliveries.id").Eq(goqu.I("active_intel_delivery_attempts.delivery")))).
		Select(goqu.I("active_intel_deliveries.id"),
			goqu.I("active_intel_deliveries.intel"),
			goqu.I("active_intel_deliveries.to"),
			goqu.I("active_intel_deliveries.note"),
			goqu.I("intel.id"),
			goqu.I("intel.created_at"),
			goqu.I("intel.created_by"),
			goqu.I("intel.operation"),
			goqu.I("intel.importance"),
			goqu.I("intel.is_valid")).
		Where( // Only include the ones, not having active delivery attempts.
			goqu.I("active_intel_delivery_attempts.id").IsNull(),
			// Only include for the given operation.
			goqu.I("intel.operation").Eq(operationID),
			// Only include deliveries not for entries that have auto-intel-delivery enabled.
			goqu.I("active_intel_deliveries.to").NotIn(m.dialect.From(goqu.T("auto_intel_delivery_address_book_entries")).
				Select(goqu.C("entry")))).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	deliveries := make([]OpenIntelDeliverySummary, 0)
	for rows.Next() {
		var delivery OpenIntelDeliverySummary
		err = rows.Scan(&delivery.Delivery.ID,
			&delivery.Delivery.Intel,
			&delivery.Delivery.To,
			&delivery.Delivery.Note,
			&delivery.Intel.ID,
			&delivery.Intel.CreatedAt,
			&delivery.Intel.CreatedBy,
			&delivery.Intel.Operation,
			&delivery.Intel.Importance,
			&delivery.Intel.IsValid)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		deliveries = append(deliveries, delivery)
	}
	return deliveries, nil
}

// IntelOperationsByActiveIntelDeliveryRecipient retrieves a list of distinct
// operation ids. These operations have intel that currently has active intel
// deliveries with the given address book entry being the recipient.
//
// This is used for finding affected (and therefore potentially changed)
// operations when auto-intel-delivery is changed for the given address book
// entry.
func (m *Mall) IntelOperationsByActiveIntelDeliveryRecipient(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID) ([]uuid.UUID, error) {
	q, _, err := m.dialect.From(goqu.T("active_intel_deliveries")).
		InnerJoin(goqu.T("intel"),
			goqu.On(goqu.I("intel.id").Eq(goqu.I("active_intel_deliveries.intel")))).
		Select(goqu.DISTINCT(goqu.I("intel.operation"))).
		Where(goqu.I("active_intel_deliveries.to").Eq(addressBookEntryID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	operations := make([]uuid.UUID, 0)
	for rows.Next() {
		var operation uuid.UUID
		err = rows.Scan(&operation)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		operations = append(operations, operation)
	}
	return operations, nil
}
