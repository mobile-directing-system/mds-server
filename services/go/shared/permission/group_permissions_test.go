package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestCreateGroup(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "create-group",
		Matcher:     CreateGroup(),
		Granted:     CreateGroupPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
		},
	})
}

func TestUpdateGroup(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "update-group",
		Matcher:     UpdateGroup(),
		Granted:     UpdateGroupPermissionName,
		Others: []Name{
			CreateGroupPermissionName,
			DeleteGroupPermissionName,
			CreateOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
		},
	})
}

func TestDeleteGroup(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "delete-group",
		Matcher:     DeleteGroup(),
		Granted:     DeleteGroupPermissionName,
		Others: []Name{
			CreateGroupPermissionName,
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
		},
	})
}

func TestViewGroup(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "view-group",
		Matcher:     ViewGroup(),
		Granted:     ViewGroupPermissionName,
		Others: []Name{
			CreateGroupPermissionName,
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
		},
	})
}
