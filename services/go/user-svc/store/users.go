package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
)

const userSearchIndex search.Index = "users"

const userSearchAttrID search.Attribute = "id"
const userSearchAttrUsername search.Attribute = "username"
const userSearchAttrFirstName search.Attribute = "first_name"
const userSearchAttrLastName search.Attribute = "last_name"

var userSearchIndexConfig = search.IndexConfig{
	PrimaryKey: userSearchAttrID,
	Searchable: []search.Attribute{
		userSearchAttrID,
		userSearchAttrUsername,
		userSearchAttrFirstName,
		userSearchAttrLastName,
	},
	Filterable: nil,
	Sortable:   nil,
}

// User contains all stored user information.
type User struct {
	// ID identifies the user.
	ID uuid.UUID
	// Username for the user.
	Username string
	// FirstName of the user.
	FirstName string
	// LastName of the user.
	LastName string
	// IsAdmin describes whether the User is an administrator.
	IsAdmin bool
}

// Validate assures that Username, FirstName and LastName are not empty.
func (u User) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	if u.Username == "" {
		report.AddError("username must not be empty")
	}
	if u.FirstName == "" {
		report.AddError("first name must not be empty")
	}
	if u.LastName == "" {
		report.AddError("last name must not be empty")
	}
	return report, nil
}

// documentFromUser generates a search.Document from the given User.
func documentFromUser(u User) search.Document {
	return search.Document{
		userSearchAttrID:        u.ID,
		userSearchAttrUsername:  u.Username,
		userSearchAttrFirstName: u.FirstName,
		userSearchAttrLastName:  u.LastName,
	}
}

// UserWithPass is a User with a Pass field.
type UserWithPass struct {
	User
	// Pass is the hashed password for the user.
	Pass []byte
}

// Validate the User and assure the Pass not being empty.
func (u UserWithPass) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	if string(u.Pass) == "" {
		report.AddError("password must not be empty")
	}
	subReport, err := u.User.Validate()
	if err != nil {
		return entityvalidation.Report{}, meh.Wrap(err, "validate user", meh.Details{"user": u.User})
	}
	report.Include(subReport)
	return report, nil
}

// UserByID retrieves a User by its User.ID.
func (m *Mall) UserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (User, error) {
	// Build query.
	q, _, err := m.dialect.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_admin")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return User{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return User{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	if !rows.Next() {
		return User{}, meh.NewNotFoundErr("user not found", nil)
	}
	var user User
	err = rows.Scan(&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsAdmin)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// UserByUsername retrieves a User by its User.Username.
func (m *Mall) UserByUsername(ctx context.Context, tx pgx.Tx, username string) (User, error) {
	// Build query.
	q, _, err := m.dialect.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_admin")).
		Where(goqu.C("username").Eq(username)).ToSQL()
	if err != nil {
		return User{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return User{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	if !rows.Next() {
		return User{}, meh.NewNotFoundErr("user not found", nil)
	}
	var user User
	err = rows.Scan(&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsAdmin)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// Users retrieves all known users.
func (m *Mall) Users(ctx context.Context, tx pgx.Tx, params pagination.Params) (pagination.Paginated[User], error) {
	// Build query.
	q, _, err := pagination.QueryToSQLWithPagination(m.dialect.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_admin")).
		Order(goqu.C("username").Asc()), params, pagination.FieldMap{
		"username":   goqu.C("username"),
		"first_name": goqu.C("first_name"),
		"last_name":  goqu.C("last_name"),
		"is_admin":   goqu.C("is_admin"),
	})
	if err != nil {
		return pagination.Paginated[User]{}, meh.Wrap(err, "query to sql with pagination", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return pagination.Paginated[User]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	users := make([]User, 0)
	var total int
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.IsAdmin,
			&total)
		if err != nil {
			return pagination.Paginated[User]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		users = append(users, user)
	}
	return pagination.NewPaginated(params, users, total), nil
}

// CreateUser creates the given user.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, user UserWithPass) (User, error) {
	// Build query.
	q, _, err := m.dialect.Insert(goqu.T("users")).Rows(goqu.Record{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_admin":   user.IsAdmin,
		"pass":       user.Pass,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return User{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return User{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return User{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return User{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	err = rows.Scan(&user.ID)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	// Add to search.
	err = search.AddOrUpdateDocuments(m.searchClient, userSearchIndex, documentFromUser(user.User))
	if err != nil {
		return User{}, meh.Wrap(err, "add or update in search", nil)
	}
	return user.User, nil
}

// UpdateUser updates the given User, identified by its User.ID. This will not
// change the password!
func (m *Mall) UpdateUser(ctx context.Context, tx pgx.Tx, user User) error {
	// Build query.
	q, _, err := m.dialect.Update(goqu.T("users")).Set(goqu.Record{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_admin":   user.IsAdmin,
	}).Where(goqu.C("id").Eq(user.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("user not found", nil)
	}
	// Update in search.
	err = search.AddOrUpdateDocuments(m.searchClient, userSearchIndex, documentFromUser(user))
	if err != nil {
		return meh.Wrap(err, "add or update in search", nil)
	}
	return nil
}

// DeleteUserByID deletes the user with the given id.
func (m *Mall) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Build query.
	q, _, err := m.dialect.Delete(goqu.T("users")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("user not found", nil)
	}
	// Delete from search.
	err = search.DeleteDocumentsByUUID(m.searchClient, userSearchIndex, userID)
	if err != nil {
		return meh.Wrap(err, "delete in search", nil)
	}
	return nil
}

// UpdateUserPassByUserID updates the hashed password of the user with the given
// id.
func (m *Mall) UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, pass []byte) error {
	// Build query.
	q, _, err := m.dialect.Update(goqu.T("users")).Set(goqu.Record{
		"pass": pass,
	}).Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("user not found", nil)
	}
	return nil
}

// SearchUsers searches for users with the given search.Params.
func (m *Mall) SearchUsers(ctx context.Context, tx pgx.Tx, searchParams search.Params) (search.Result[User], error) {
	// Search.
	resultUUIDs, err := search.UUIDSearch(m.searchClient, userSearchIndex, searchParams)
	if err != nil {
		return search.Result[User]{}, meh.Wrap(err, "search uuids", meh.Details{
			"index":  userSearchIndex,
			"params": searchParams,
		})
	}
	// Query.
	q, _, err := pgutil.QueryWithOrdinalityUUID(m.dialect.From(goqu.T("users")).
		Select(goqu.I("users.id"),
			goqu.I("users.username"),
			goqu.I("users.first_name"),
			goqu.I("users.last_name"),
			goqu.I("users.is_admin")), goqu.I("users.id"), resultUUIDs.Hits).ToSQL()
	if err != nil {
		return search.Result[User]{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return search.Result[User]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	users := make([]User, 0, len(resultUUIDs.Hits))
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.IsAdmin)
		if err != nil {
			return search.Result[User]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		users = append(users, user)
	}
	return search.ResultFromResult(resultUUIDs, users), nil
}

// RebuildUserSearch rebuilds the user search.
func (m *Mall) RebuildUserSearch(ctx context.Context, tx pgx.Tx) error {
	err := search.Rebuild(ctx, m.searchClient, userSearchIndex, search.DefaultBatchSize,
		func(ctx context.Context, next chan<- search.Document) error {
			defer close(next)
			// Build query.
			q, _, err := m.dialect.From(goqu.T("users")).
				Select(goqu.C("id"),
					goqu.C("username"),
					goqu.C("first_name"),
					goqu.C("last_name"),
					goqu.C("is_admin")).ToSQL()
			if err != nil {
				return meh.Wrap(err, "query to sql", nil)
			}
			// Query.
			rows, err := tx.Query(ctx, q)
			if err != nil {
				return mehpg.NewQueryDBErr(err, "query db", q)
			}
			defer rows.Close()
			// Scan.
			for rows.Next() {
				var user User
				err = rows.Scan(&user.ID,
					&user.Username,
					&user.FirstName,
					&user.LastName,
					&user.IsAdmin)
				if err != nil {
					return mehpg.NewScanRowsErr(err, "scan row", q)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case next <- documentFromUser(user):
				}
			}
			return nil
		})
	if err != nil {
		return meh.Wrap(err, "rebuild search", nil)
	}
	return nil
}
