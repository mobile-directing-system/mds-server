package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"time"
)

// IntelType of intel. Also describes the content.
type IntelType string

// CreateIntel for creating intel.
type CreateIntel struct {
	// CreatedBy is the id of the user, who created the intent.
	CreatedBy uuid.UUID
	// Operation is the id of the associated operation.
	Operation uuid.UUID
	// Type of the intel.
	Type IntelType
	// Content is the actual information.
	Content json.RawMessage
	// SearchText for better searching. Used with higher priority than Content.
	SearchText nulls.String
	// Importance of the intel. Used for example for prioritizing delivery methods.
	Importance int
	// Assignments contains the recipients to assign the intel to.
	Assignments []IntelAssignment
}

// Validate the CreateIntel for Type, Content and Assignments.
func (i CreateIntel) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	// Validate intel-type and content.
	subReport, err := validateCreateIntelTypeAndContent(i.Type, i.Content)
	if err != nil {
		return entityvalidation.Report{}, meh.Wrap(err, "validate create-intel-type and content", nil)
	}
	report.Include(subReport)
	// Assure no duplicate assignments.
	assignedTo := make(map[uuid.UUID]struct{}, len(i.Assignments))
	for _, assignment := range i.Assignments {
		if _, ok := assignedTo[assignment.To]; ok {
			report.AddError(fmt.Sprintf("duplicate assignment to %s", assignment.To.String()))
			continue
		}
		assignedTo[assignment.To] = struct{}{}
	}
	return report, nil
}

// IntelAssignment for delivering Intel.
type IntelAssignment struct {
	// ID identifies the assignment.
	ID uuid.UUID
	// Intel is the id of the intel to deliver.
	Intel uuid.UUID
	// To is the id of the target address book entry.
	To uuid.UUID
}

// Intel for intel information with deliveries.
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
	// SearchText for better searching. Used with higher priority than Content.
	SearchText nulls.String
	// Importance of the intel. Used for example for prioritizing delivery methods.
	Importance int
	// IsValid describes whether the intel is still valid or marked as invalid
	// (equals deletion).
	IsValid bool
	// Assignments of the intel.
	Assignments []IntelAssignment
}

// CreateIntel creates the given intel and returns it with its assigned id.
func (m *Mall) CreateIntel(ctx context.Context, tx pgx.Tx, create CreateIntel) (Intel, error) {
	// Create intel.
	q, _, err := m.dialect.Insert(goqu.T("intel")).Rows(goqu.Record{
		"created_at":  time.Now().UTC(),
		"created_by":  create.CreatedBy,
		"operation":   create.Operation,
		"type":        create.Type,
		"content":     []byte(create.Content),
		"search_text": create.SearchText,
		"importance":  create.Importance,
		"is_valid":    true,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return Intel{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Intel{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return Intel{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return Intel{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	var intelID uuid.UUID
	err = rows.Scan(&intelID)
	if err != nil {
		return Intel{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	// Cerate deliveries.
	for _, createAssignment := range create.Assignments {
		_, err = m.createIntelAssignment(ctx, tx, intelID, createAssignment)
		if err != nil {
			return Intel{}, meh.Wrap(err, "create intel assignment", meh.Details{
				"intel_id": intelID,
				"create":   createAssignment,
			})
		}
	}
	// Return created intel.
	created, err := m.IntelByID(ctx, tx, intelID)
	if err != nil {
		return Intel{}, meh.Wrap(err, "created intel by id", meh.Details{"intel_id": intelID})
	}
	return created, nil
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

// createIntelAssignment creates the given IntelAssignment.
func (m *Mall) createIntelAssignment(ctx context.Context, tx pgx.Tx, intelID uuid.UUID, create IntelAssignment) (IntelAssignment, error) {
	q, _, err := m.dialect.Insert(goqu.T("intel_assignments")).Rows(goqu.Record{
		"intel": intelID,
		"to":    create.To,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return IntelAssignment{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return IntelAssignment{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return IntelAssignment{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return IntelAssignment{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	var assignmentID uuid.UUID
	err = rows.Scan(&assignmentID)
	if err != nil {
		return IntelAssignment{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	created, err := m.IntelAssignmentByID(ctx, tx, assignmentID)
	if err != nil {
		return IntelAssignment{}, meh.Wrap(err, "created intel assignment by id", meh.Details{"id": assignmentID})
	}
	return created, nil
}

// IntelAssignmentByID retrieves the IntelAssignment with the given id.
func (m *Mall) IntelAssignmentByID(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) (IntelAssignment, error) {
	q, _, err := goqu.From(goqu.T("intel_assignments")).
		Select(goqu.C("id"),
			goqu.C("intel"),
			goqu.C("to")).
		Where(goqu.C("id").Eq(deliveryID)).ToSQL()
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
	return assignment, nil
}

func (m *Mall) intelAssignmentsByIntel(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) ([]IntelAssignment, error) {
	q, _, err := m.dialect.From(goqu.T("intel_assignments")).
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
		return Intel{}, meh.Wrap(err, "intel deliveries by id", meh.Details{"intel_id": intelID})
	}
	return intel, nil
}
