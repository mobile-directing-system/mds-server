package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// Group for grouping users for different purposes.
type Group struct {
	// ID identifies the group.
	ID uuid.UUID
	// Title of the group.
	Title string
	// Description is optional description content.
	Description string
	// Operation is the id of an optionally assigned operation.
	Operation uuid.NullUUID
	// Members of the group.
	Members []uuid.UUID
}

// CreateGroup creates the given Group.
func (m *Mall) CreateGroup(ctx context.Context, tx pgx.Tx, create Group) error {
	// Insert metadata.
	metadataQuery, _, err := m.dialect.Insert(goqu.T("groups")).Rows(goqu.Record{
		"id":          create.ID,
		"title":       create.Title,
		"description": create.Description,
		"operation":   create.Operation,
	}).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "metadata-query to sql", nil)
	}
	_, err = tx.Exec(ctx, metadataQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec metadata-query", metadataQuery)
	}
	// Insert members.
	err = m.updateGroupMembers(ctx, tx, create.ID, create.Members)
	if err != nil {
		return meh.Wrap(err, "update group members", meh.Details{
			"group_id":      create.ID,
			"group_members": create.Members,
		})
	}
	return nil
}

// updateGroupMembers clears all members for the group with the given id and
// inserts the new given ones.
func (m *Mall) updateGroupMembers(ctx context.Context, tx pgx.Tx, groupID uuid.UUID, members []uuid.UUID) error {
	// Clear.
	clearMembersQuery, _, err := m.dialect.Delete(goqu.T("group_members")).
		Where(goqu.C("group").Eq(groupID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "clear-members-query to sql", nil)
	}
	_, err = tx.Exec(ctx, clearMembersQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec clear-members-query", clearMembersQuery)
	}
	// Insert members.
	if len(members) == 0 {
		return nil
	}
	records := make([]any, 0, len(members))
	for _, member := range members {
		records = append(records, goqu.Record{
			"group": groupID,
			"user":  member,
		})
	}
	insertMembersQuery, _, err := m.dialect.Insert(goqu.T("group_members")).
		Rows(records...).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "insert-members-query to sql", nil)
	}
	_, err = tx.Exec(ctx, insertMembersQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec insert-members-query", insertMembersQuery)
	}
	return nil
}

// UpdateGroup updates the given Group, identified by its id.
func (m *Mall) UpdateGroup(ctx context.Context, tx pgx.Tx, update Group) error {
	// Update metadata.
	metadataQuery, _, err := m.dialect.Update(goqu.T("groups")).Set(goqu.Record{
		"title":       update.Title,
		"description": update.Description,
		"operation":   update.Operation,
	}).Where(goqu.C("id").Eq(update.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "metadata-query to sql", nil)
	}
	_, err = tx.Exec(ctx, metadataQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec metadata-query", metadataQuery)
	}
	// Update members.
	err = m.updateGroupMembers(ctx, tx, update.ID, update.Members)
	if err != nil {
		return meh.Wrap(err, "update group members", meh.Details{
			"group_id":      update.ID,
			"group_members": update.Members,
		})
	}
	return nil
}

// DeleteGroupByID deletes the group with the given id.
func (m *Mall) DeleteGroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	q, _, err := m.dialect.Delete(goqu.T("groups")).
		Where(goqu.C("id").Eq(groupID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}
