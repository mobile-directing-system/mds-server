package store

import (
	"context"
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"time"
)

// IntelType is the type of intel. Also describes the content.
type IntelType string

// Intel holds intel information.
type Intel struct {
	// ID identifies the intel.
	ID uuid.UUID
	// CreatedAt is the timestamp, the intel was created.
	CreatedAt time.Time
	// CreatedBy is the id of the user, who created the intent.
	CreatedBy uuid.UUID
	// Operation is the id of the associated operation.
	Operation uuid.UUID
	// Type of the intel.
	Type IntelType
	// Content is the actual information.
	Content json.RawMessage
	// SearchText for searching with higher priority than Content.
	SearchText nulls.String
	// Importance of the intel.
	Importance int
	// IsValid describes whether the intel is still valid or marked as invalid
	// (equals deletion).
	IsValid bool
	// Deliveries holds the associated deliveries for the Intel.
	Assignments []IntelAssignment
}

// IntelAssignment of Intel.
type IntelAssignment struct {
	// ID identifies the assignment.
	ID uuid.UUID
	// Intel is the id of the assigned intel.
	Intel uuid.UUID
	// To is the id of the assigned address book entry.
	To uuid.UUID
}

// CreateIntel creates the given intel with its assignments.
func (m *Mall) CreateIntel(ctx context.Context, tx pgx.Tx, create Intel) error {
	// Create intel.
	q, _, err := m.dialect.Insert(goqu.T("intel")).Rows(goqu.Record{
		"id":          create.ID,
		"created_at":  create.CreatedAt.UTC(),
		"created_by":  create.CreatedBy,
		"operation":   create.Operation,
		"type":        create.Type,
		"content":     []byte(create.Content),
		"search_text": create.SearchText,
		"importance":  create.Importance,
		"is_valid":    create.IsValid,
	}).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	// Create assignments.
	for _, assignment := range create.Assignments {
		err = m.createIntelAssignment(ctx, tx, assignment)
		if err != nil {
			return meh.Wrap(err, "create intel assignment", meh.Details{
				"create": assignment,
			})
		}
	}
	return nil
}

// createIntelAssignment creates the given IntelAssignment in the database.
func (m *Mall) createIntelAssignment(ctx context.Context, tx pgx.Tx, create IntelAssignment) error {
	q, _, err := m.dialect.Insert(goqu.T("intel_assignments")).Rows(goqu.Record{
		"id":    create.ID,
		"intel": create.Intel,
		"to":    create.To,
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

// InvalidateIntelByID sets the valid-field of the intel with the given id to
// false.
func (m *Mall) InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	q, _, err := m.dialect.Update(goqu.T("intel")).Set(goqu.Record{
		"is_valid": false,
	}).Where(goqu.C("id").Eq(intelID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", nil)
	}
	return nil
}

// IntelByID retrieves an Intel by its id.
func (m *Mall) IntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) (Intel, error) {
	// Retrieve intel information.
	q, _, err := m.dialect.From(goqu.T("intel")).
		Select(goqu.C("id"),
			goqu.C("created_at"),
			goqu.C("created_by"),
			goqu.C("operation"),
			goqu.C("type"),
			goqu.C("content"),
			goqu.C("search_text"),
			goqu.C("importance"),
			goqu.C("is_valid")).
		Where(goqu.C("id").Eq(intelID)).ToSQL()
	if err != nil {
		return Intel{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Intel{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return Intel{}, meh.NewNotFoundErr("not found", nil)
	}
	intel := Intel{
		Assignments: make([]IntelAssignment, 0),
	}
	err = rows.Scan(&intel.ID,
		&intel.CreatedAt,
		&intel.CreatedBy,
		&intel.Operation,
		&intel.Type,
		&intel.Content,
		&intel.SearchText,
		&intel.Importance,
		&intel.IsValid)
	if err != nil {
		return Intel{}, mehpg.NewScanRowsErr(err, "scan rows", q)
	}
	rows.Close()
	// Retrieve assignments.
	intel.Assignments, err = m.intelAssignmentsByIntel(ctx, tx, intelID)
	if err != nil {
		return Intel{}, meh.Wrap(err, "intel assignments by intel", meh.Details{"intel_id": intelID})
	}
	return intel, nil
}

// intelAssignmentsByIntel retrieves the IntelAssignment list for the intel with
// the given id.
func (m *Mall) intelAssignmentsByIntel(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) ([]IntelAssignment, error) {
	q, _, err := goqu.From(goqu.T("intel_assignments")).
		Select(goqu.C("id"),
			goqu.C("intel"),
			goqu.C("to")).
		Where(goqu.C("intel").Eq(intelID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	assignments := make([]IntelAssignment, 0)
	for rows.Next() {
		var assignment IntelAssignment
		err = rows.Scan(&assignment.ID,
			&assignment.Intel,
			&assignment.To)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		assignments = append(assignments, assignment)
	}
	rows.Close()
	return assignments, nil
}

// IntelAssignmentByID retrieves the IntelAssignment with the given id.
func (m *Mall) IntelAssignmentByID(ctx context.Context, tx pgx.Tx, assignmentID uuid.UUID) (IntelAssignment, error) {
	q, _, err := goqu.From(goqu.T("intel_assignments")).
		Select(goqu.C("id"),
			goqu.C("intel"),
			goqu.C("to")).
		Where(goqu.C("id").Eq(assignmentID)).ToSQL()
	if err != nil {
		return IntelAssignment{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return IntelAssignment{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return IntelAssignment{}, meh.NewNotFoundErr("not found", nil)
	}
	var assignment IntelAssignment
	err = rows.Scan(&assignment.Intel,
		&assignment.Intel,
		&assignment.To)
	if err != nil {
		return IntelAssignment{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	return assignment, nil
}

// TODO: FOR LATER PROPER SYNC AND LOCKS FOR SCHEDULER on attempts, deliveries, channel choice, etc.

// TODO: SEARCH, BATCH RETRIEVAL, BLABLABLA
