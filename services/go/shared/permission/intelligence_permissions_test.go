package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestCreateIntelPermission(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "create-intel",
		Matcher:     CreateIntel(),
		Granted:     CreateIntelPermissionName,
		Others: []Name{
			InvalidateIntelPermissionName,
			ViewAnyIntelPermissionName,
			CreateGroupPermissionName,
			ViewAnyAddressBookEntryPermissionName,
		},
	})
}

func TestInvalidateIntel(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "invalidate-intel",
		Matcher:     InvalidateIntel(),
		Granted:     InvalidateIntelPermissionName,
		Others: []Name{
			CreateIntelPermissionName,
			ViewAnyIntelPermissionName,
			CreateGroupPermissionName,
			ViewAnyAddressBookEntryPermissionName,
		},
	})
}

func TestViewAnyIntel(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "view-any-intel",
		Matcher:     ViewAnyIntel(),
		Granted:     ViewAnyIntelPermissionName,
		Others: []Name{
			CreateIntelPermissionName,
			InvalidateIntelPermissionName,
			CreateGroupPermissionName,
			ViewAnyAddressBookEntryPermissionName,
		},
	})
}
