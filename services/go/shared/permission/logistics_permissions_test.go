package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestCreateAddressBookEntry(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "create-any-address-book-entry",
		Matcher:     CreateAnyAddressBookEntry(),
		Granted:     CreateAnyAddressBookEntryPermissionName,
		Others: []Name{
			UpdateAnyAddressBookEntryPermissionName,
			DeleteAnyAddressBookEntryPermissionName,
			CreateGroupPermissionName,
			ViewAnyAddressBookEntryPermissionName,
		},
	})
}

func TestUpdateAddressBookEntry(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "update-any-address-book-entry",
		Matcher:     UpdateAnyAddressBookEntry(),
		Granted:     UpdateAnyAddressBookEntryPermissionName,
		Others: []Name{
			CreateAnyAddressBookEntryPermissionName,
			DeleteAnyAddressBookEntryPermissionName,
			UpdateGroupPermissionName,
			ViewAnyAddressBookEntryPermissionName,
		},
	})
}

func TestDeleteAddressBookEntry(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "delete-any-address-book-entry",
		Matcher:     DeleteAnyAddressBookEntry(),
		Granted:     DeleteAnyAddressBookEntryPermissionName,
		Others: []Name{
			CreateAnyAddressBookEntryPermissionName,
			UpdateAnyAddressBookEntryPermissionName,
			DeleteGroupPermissionName,
			ViewAnyAddressBookEntryPermissionName,
		},
	})
}

func TestViewAnyAddressBookEntry(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "view-any-address-book-entry",
		Matcher:     ViewAnyAddressBookEntry(),
		Granted:     ViewAnyAddressBookEntryPermissionName,
		Others: []Name{
			CreateAnyAddressBookEntryPermissionName,
			UpdateAnyAddressBookEntryPermissionName,
			DeleteAnyAddressBookEntryPermissionName,
			ViewAnyOperationPermissionName,
		},
	})
}
