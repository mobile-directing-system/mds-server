package store

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// documentFromAddressBookEntrySuite tests documentFromAddressBookEntry.
type documentFromAddressBookEntrySuite struct {
	suite.Suite
	entry     AddressBookEntry
	user      User
	operation Operation
	visibleBy []uuid.UUID
}

func (suite *documentFromAddressBookEntrySuite) SetupTest() {
	suite.entry = AddressBookEntry{
		ID:          testutil.NewUUIDV4(),
		Label:       "plaster",
		Description: "airplane",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.user = User{
		ID:        testutil.NewUUIDV4(),
		Username:  "second",
		FirstName: "rapid",
		LastName:  "upset",
		IsActive:  true,
	}
	suite.operation = Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "consider",
		Description: "pressure",
		Start:       time.Date(2022, 9, 5, 16, 54, 13, 0, time.UTC),
		End:         nulls.Time{},
		IsArchived:  true,
	}
	suite.visibleBy = make([]uuid.UUID, 0, 32)
	for i := range suite.visibleBy {
		suite.visibleBy[i] = testutil.NewUUIDV4()
	}
}

func (suite *documentFromAddressBookEntrySuite) TestEntryOnly() {
	d := documentFromAddressBookEntry(suite.entry, nulls.JSONNullable[User]{}, nulls.JSONNullable[Operation]{}, nil)
	suite.Equal(search.Document{
		abEntrySearchAttrID:          suite.entry.ID,
		abEntrySearchAttrLabel:       suite.entry.Label,
		abEntrySearchAttrDescription: suite.entry.Description,
	}, d, "should return expected document")
}

func (suite *documentFromAddressBookEntrySuite) TestEntryWithUser() {
	d := documentFromAddressBookEntry(suite.entry, nulls.NewJSONNullable(suite.user), nulls.JSONNullable[Operation]{}, suite.visibleBy)
	suite.Equal(search.Document{
		abEntrySearchAttrID:            suite.entry.ID,
		abEntrySearchAttrLabel:         suite.entry.Label,
		abEntrySearchAttrDescription:   suite.entry.Description,
		abEntrySearchAttrUserID:        suite.user.ID,
		abEntrySearchAttrUserUsername:  suite.user.Username,
		abEntrySearchAttrUserFirstName: suite.user.FirstName,
		abEntrySearchAttrUserLastName:  suite.user.LastName,
		abEntrySearchAttrUserIsActive:  suite.user.IsActive,
		abEntrySearchAttrVisibleBy:     suite.visibleBy,
	}, d, "should return expected document")
}

func (suite *documentFromAddressBookEntrySuite) TestEntryWithOperation() {
	d := documentFromAddressBookEntry(suite.entry, nulls.JSONNullable[User]{}, nulls.NewJSONNullable(suite.operation), nil)
	suite.Equal(search.Document{
		abEntrySearchAttrID:                  suite.entry.ID,
		abEntrySearchAttrLabel:               suite.entry.Label,
		abEntrySearchAttrDescription:         suite.entry.Description,
		abEntrySearchAttrOperationID:         suite.operation.ID,
		abEntrySearchAttrOperationTitle:      suite.operation.Title,
		abEntrySearchAttrOperationIsArchived: suite.operation.IsArchived,
	}, d, "should return expected document")
}

func (suite *documentFromAddressBookEntrySuite) TestEntryWithUserOperation() {
	d := documentFromAddressBookEntry(suite.entry, nulls.NewJSONNullable(suite.user), nulls.NewJSONNullable(suite.operation), suite.visibleBy)
	suite.Equal(search.Document{
		abEntrySearchAttrID:                  suite.entry.ID,
		abEntrySearchAttrLabel:               suite.entry.Label,
		abEntrySearchAttrDescription:         suite.entry.Description,
		abEntrySearchAttrUserID:              suite.user.ID,
		abEntrySearchAttrUserUsername:        suite.user.Username,
		abEntrySearchAttrUserFirstName:       suite.user.FirstName,
		abEntrySearchAttrUserLastName:        suite.user.LastName,
		abEntrySearchAttrUserIsActive:        suite.user.IsActive,
		abEntrySearchAttrVisibleBy:           suite.visibleBy,
		abEntrySearchAttrOperationID:         suite.operation.ID,
		abEntrySearchAttrOperationTitle:      suite.operation.Title,
		abEntrySearchAttrOperationIsArchived: suite.operation.IsArchived,
	}, d, "should return expected document")
}

func Test_documentFromAddressBookEntry(t *testing.T) {
	suite.Run(t, new(documentFromAddressBookEntrySuite))
}
