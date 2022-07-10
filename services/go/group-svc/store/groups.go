package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"time"
)

// Group represents a group with its members.
type Group struct {
	// ID identifies the group.
	ID uuid.UUID
	// Title of the group.
	Title string
	// Description of the group.
	Description string
	// Operation is the id an optional operation.
	Operation uuid.NullUUID
	// Members of the group represented by user ids.
	Members []uuid.UUID
}

// Validate that the group title is set.
func (g Group) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	if g.Title == "" {
		report.AddError("title must not be empty")
	}
	return report, nil
}

// CreateGroup creates the given group and returns the one with assigned id.
func (m *Mall) CreateGroup(ctx context.Context, tx pgx.Tx, create Group) (Group, error) {
	if hasDuplicateMembers(create.Members) {
		return Group{}, meh.NewBadInputErr("duplicate members", meh.Details{"members": create.Members})
	}
	// Create group.
	createGroupQuery, _, err := goqu.Insert(goqu.T("groups")).Rows(goqu.Record{
		"title":       create.Title,
		"description": create.Description,
		"operation":   create.Operation,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return Group{}, meh.NewInternalErrFromErr(err, "create-group-query to sql", nil)
	}
	rows, err := tx.Query(ctx, createGroupQuery)
	if err != nil {
		return Group{}, mehpg.NewQueryDBErr(err, "exec create-group-query", createGroupQuery)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return Group{}, mehpg.NewQueryDBErr(err, "next row", createGroupQuery)
		}
		return Group{}, meh.NewInternalErr("no rows returned", meh.Details{"query": createGroupQuery})
	}
	err = rows.Scan(&create.ID)
	if err != nil {
		return Group{}, mehpg.NewScanRowsErr(err, "scan row", createGroupQuery)
	}
	rows.Close()
	// Add members.
	members := make([]interface{}, 0, len(create.Members))
	for _, member := range create.Members {
		members = append(members, goqu.Record{
			"group":        create.ID,
			"user":         member,
			"member_since": time.Now().UTC(),
		})
	}
	if len(members) > 0 {
		addMembersQuery, _, err := goqu.Insert(goqu.T("members")).
			Rows(members...).ToSQL()
		if err != nil {
			return Group{}, meh.NewInternalErrFromErr(err, "add-members-query to sql", nil)
		}
		_, err = tx.Exec(ctx, addMembersQuery)
		if err != nil {
			return Group{}, mehpg.NewQueryDBErr(err, "exec add-members-query", addMembersQuery)
		}
	}
	return create, nil
}

// UpdateGroup updates the group identified by its id.
func (m *Mall) UpdateGroup(ctx context.Context, tx pgx.Tx, update Group) error {
	if hasDuplicateMembers(update.Members) {
		return meh.NewBadInputErr("duplicate members", meh.Details{"members": update.Members})
	}
	// Retrieve current group members.
	old, err := m.GroupByID(ctx, tx, update.ID)
	if err != nil {
		return meh.Wrap(err, "group by id", meh.Details{"group_id": update.ID})
	}
	// Save old an new members to map for fast checking which to delete and add.
	oldMembers := make(map[uuid.UUID]struct{}, len(old.Members))
	for _, member := range old.Members {
		oldMembers[member] = struct{}{}
	}
	updatedMembers := make(map[uuid.UUID]struct{}, len(update.Members))
	for _, member := range update.Members {
		updatedMembers[member] = struct{}{}
	}
	// Update group details.
	updateGroupQuery, _, err := goqu.Update(goqu.T("groups")).Set(goqu.Record{
		"title":       update.Title,
		"description": update.Description,
		"operation":   update.Operation,
	}).Where(goqu.C("id").Eq(update.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "update-group-query to sql", nil)
	}
	result, err := tx.Exec(ctx, updateGroupQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", updateGroupQuery)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", nil)
	}
	// Add new members.
	toAdd := make([]interface{}, 0)
	for _, member := range update.Members {
		if _, ok := oldMembers[member]; !ok {
			toAdd = append(toAdd, goqu.Record{
				"group":        update.ID,
				"user":         member,
				"member_since": time.Now().UTC(),
			})
		}
	}
	if len(toAdd) > 0 {
		addMembersQuery, _, err := goqu.Insert(goqu.T("members")).
			Rows(toAdd...).ToSQL()
		if err != nil {
			return meh.NewInternalErrFromErr(err, "add-members-query to sql", nil)
		}
		_, err = tx.Exec(ctx, addMembersQuery)
		if err != nil {
			return mehpg.NewQueryDBErr(err, "exec add-members-query", addMembersQuery)
		}
	}
	// Delete old members.
	toDelete := make([]uuid.UUID, 0)
	for oldMember := range oldMembers {
		if _, ok := updatedMembers[oldMember]; !ok {
			toDelete = append(toDelete, oldMember)
		}
	}
	if len(toDelete) > 0 {
		deleteMembersQuery, _, err := goqu.Delete(goqu.T("members")).
			Where(goqu.And(goqu.C("group").Eq(update.ID), goqu.C("user").In(toDelete))).ToSQL()
		if err != nil {
			return meh.NewInternalErrFromErr(err, "delete-members-query to sql", nil)
		}
		_, err = tx.Exec(ctx, deleteMembersQuery)
		if err != nil {
			return mehpg.NewQueryDBErr(err, "exec delete-members-query", deleteMembersQuery)
		}
	}
	return nil
}

// GroupByID retrieves a group by its id.
func (m *Mall) GroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) (Group, error) {
	group := Group{Members: make([]uuid.UUID, 0)}
	// Retrieve group details.
	groupDetailsQuery, _, err := goqu.From(goqu.T("groups")).
		Select(goqu.C("id"),
			goqu.C("title"),
			goqu.C("description"),
			goqu.C("operation")).
		Where(goqu.C("id").Eq(groupID)).ToSQL()
	if err != nil {
		return Group{}, meh.NewInternalErrFromErr(err, "group-details-query to sql", nil)
	}
	groupDetailsRows, err := tx.Query(ctx, groupDetailsQuery)
	if err != nil {
		return Group{}, mehpg.NewQueryDBErr(err, "exec group-details-query", groupDetailsQuery)
	}
	defer groupDetailsRows.Close()
	if !groupDetailsRows.Next() {
		return Group{}, meh.NewNotFoundErr("not found", nil)
	}
	err = groupDetailsRows.Scan(&group.ID,
		&group.Title,
		&group.Description,
		&group.Operation)
	if err != nil {
		return Group{}, mehpg.NewScanRowsErr(err, "scan group-details", groupDetailsQuery)
	}
	groupDetailsRows.Close()
	// Retrieve group members.
	membersQuery, _, err := goqu.From(goqu.T("members")).
		Select(goqu.C("user")).
		Where(goqu.C("group").Eq(groupID)).ToSQL()
	if err != nil {
		return Group{}, meh.NewInternalErrFromErr(err, "members-query to sql", nil)
	}
	membersRows, err := tx.Query(ctx, membersQuery)
	if err != nil {
		return Group{}, mehpg.NewQueryDBErr(err, "exec members-query", membersQuery)
	}
	defer membersRows.Close()
	for membersRows.Next() {
		var member uuid.UUID
		err = membersRows.Scan(&member)
		if err != nil {
			return Group{}, mehpg.NewScanRowsErr(err, "scan members-row", membersQuery)
		}
		group.Members = append(group.Members, member)
	}
	return group, nil
}

// GroupFilters are filters for Mall.Groups.
type GroupFilters struct {
	// ByUser includes only groups the user with the given id is member of.
	ByUser uuid.NullUUID
	// ForOperation includes only groups for the operation with this id. This
	// includes groups withotu set operation, unless ExcludeGlobal is set.
	ForOperation uuid.NullUUID
	// ExcludeGlobal excludes global groups without set operation, even if
	// ForOperation is set.
	ExcludeGlobal bool
}

// Groups retrieves a paginated Group list with optional GroupFilters.
func (m *Mall) Groups(ctx context.Context, tx pgx.Tx, filters GroupFilters, params pagination.Params) (pagination.Paginated[Group], error) {
	// Retrieve group details.
	groupDetailsQB := goqu.From(goqu.C("groups")).
		LeftJoin(goqu.T("members"), goqu.On(goqu.I("members.group").Eq(goqu.I("groups.id")))).
		Select(goqu.I("groups.id"),
			goqu.I("groups.title"),
			goqu.I("groups.description"),
			goqu.I("groups.operation"))
	groupDetailsFiltersRootAnd := make([]goqu.Expression, 0)
	groupDetailsFiltersOr := make([]goqu.Expression, 0)
	if filters.ByUser.Valid {
		err := m.AssureUserExists(ctx, tx, filters.ByUser.UUID)
		if err != nil {
			return pagination.Paginated[Group]{}, meh.Wrap(err, "assure user exists",
				meh.Details{"user_id": filters.ByUser.UUID})
		}
		groupDetailsFiltersRootAnd = append(groupDetailsFiltersRootAnd, goqu.I("members.user").Eq(filters.ByUser.UUID))
	}
	if filters.ForOperation.Valid {
		groupDetailsFiltersOr = append(groupDetailsFiltersOr, goqu.I("groups.operation").Eq(filters.ForOperation))
	}
	if !filters.ExcludeGlobal {
		groupDetailsFiltersOr = append(groupDetailsFiltersOr, goqu.I("groups.operation").IsNull())
	}
	if len(groupDetailsFiltersOr) > 0 {
		groupDetailsFiltersRootAnd = append(groupDetailsFiltersRootAnd, goqu.Or(groupDetailsFiltersOr...))
	}
	if len(groupDetailsFiltersRootAnd) > 0 {
		groupDetailsQB = groupDetailsQB.Where(goqu.And(groupDetailsFiltersRootAnd...))
	}
	groupDetailsQuery, _, err := pagination.QueryToSQLWithPagination(groupDetailsQB, params, map[string]exp.Orderable{
		"title":       goqu.I("groups.title"),
		"description": goqu.I("groups.description"),
	})
	if err != nil {
		return pagination.Paginated[Group]{}, meh.NewInternalErrFromErr(err, "group-details-query to sql", nil)
	}
	groupDetailsRows, err := tx.Query(ctx, groupDetailsQuery)
	if err != nil {
		return pagination.Paginated[Group]{}, mehpg.NewQueryDBErr(err, "exec group-details-query", groupDetailsQuery)
	}
	defer groupDetailsRows.Close()
	var groups []Group
	var total int
	for groupDetailsRows.Next() {
		group := Group{Members: make([]uuid.UUID, 0)}
		err = groupDetailsRows.Scan(&group.ID,
			&group.Title,
			&group.Description,
			&group.Operation,
			&total)
		if err != nil {
			return pagination.Paginated[Group]{}, mehpg.NewScanRowsErr(err, "scan group-details-row", groupDetailsQuery)
		}
		groups = append(groups, group)
	}
	groupDetailsRows.Close()
	if len(groups) > 0 {
		// Index groups by id with their index in list for faster adding of members
		// later.
		groupsInList := make(map[uuid.UUID]int, len(groups))
		groupIDs := make([]uuid.UUID, len(groups))
		for i, group := range groups {
			groupsInList[group.ID] = i
			groupIDs = append(groupIDs, group.ID)
		}
		// Retrieve members.
		membersQuery, _, err := goqu.From(goqu.T("members")).
			Select(goqu.C("user"),
				goqu.C("group")).
			Where(goqu.C("group").In(groupIDs)).ToSQL()
		if err != nil {
			return pagination.Paginated[Group]{}, meh.NewInternalErrFromErr(err, "members-query to sql", nil)
		}
		membersRows, err := tx.Query(ctx, membersQuery)
		if err != nil {
			return pagination.Paginated[Group]{}, mehpg.NewQueryDBErr(err, "exec members-query", membersQuery)
		}
		defer membersRows.Close()
		for membersRows.Next() {
			var member, group uuid.UUID
			err = membersRows.Scan(&member, &group)
			if err != nil {
				return pagination.Paginated[Group]{}, mehpg.NewScanRowsErr(err, "scan member-row", membersQuery)
			}
			memberGroupIndex, ok := groupsInList[group]
			if !ok {
				return pagination.Paginated[Group]{}, meh.NewInternalErr("group not found in groups list", meh.Details{
					"member": member,
					"groups": groups,
				})
			}
			memberGroup := groups[memberGroupIndex]
			memberGroup.Members = append(memberGroup.Members, member)
			groups[memberGroupIndex] = memberGroup
		}
	}
	return pagination.NewPaginated(params, groups, total), nil
}

// AssureUserExists assures that the user with the given id exists.
func (m *Mall) AssureUserExists(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	q, _, err := goqu.From(goqu.T("users")).
		Select(goqu.COUNT("*")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return mehpg.NewScanRowsErr(err, "scan row", q)
	}
	if count == 0 {
		return meh.NewNotFoundErr("not found", nil)
	}
	return nil
}

// DeleteGroupByID deletes the group with the given id.
func (m *Mall) DeleteGroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	q, _, err := goqu.Delete(goqu.T("groups")).
		Where(goqu.C("id").Eq(groupID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("group not found", nil)
	}
	return nil
}

// hasDuplicateMembers assures that there are not duplicate members in the given
// list.
func hasDuplicateMembers(members []uuid.UUID) bool {
	visited := make(map[uuid.UUID]struct{})
	for _, member := range members {
		if _, ok := visited[member]; ok {
			return true
		}
		visited[member] = struct{}{}
	}
	return false
}
