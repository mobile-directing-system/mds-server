package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestCreateUser(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "create-user",
		Matcher:     CreateUser(),
		Granted:     CreateUserPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			SetUserActiveStatePermission,
			UpdateUserPermissionName,
			SetAdminUserPermissionName,
			ViewUserPermissionName,
			UpdateUserPassPermissionName,
		},
	})
}

func TestSetUserActiveState(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "set-user-active-state",
		Matcher:     SetUserActiveState(),
		Granted:     SetUserActiveStatePermission,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			CreateUserPermissionName,
			UpdateUserPermissionName,
			SetAdminUserPermissionName,
			ViewUserPermissionName,
			UpdateUserPassPermissionName,
		},
	})
}

func TestUpdateUser(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "update-user",
		Matcher:     UpdateUser(),
		Granted:     UpdateUserPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			SetUserActiveStatePermission,
			CreateUserPermissionName,
			SetAdminUserPermissionName,
			ViewUserPermissionName,
			UpdateUserPassPermissionName,
		},
	})
}

func TestSetAdminUser(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "set-admin-user",
		Matcher:     SetAdminUser(),
		Granted:     SetAdminUserPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			SetUserActiveStatePermission,
			CreateUserPermissionName,
			UpdateUserPermissionName,
			ViewUserPermissionName,
			UpdateUserPassPermissionName,
		},
	})
}

func TestUpdateUserPass(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "update-user-pass",
		Matcher:     UpdateUserPass(),
		Granted:     UpdateUserPassPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			SetUserActiveStatePermission,
			CreateUserPermissionName,
			UpdateUserPermissionName,
			ViewUserPermissionName,
			SetAdminUserPermissionName,
		},
	})
}

func TestViewUser(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "view-user",
		Matcher:     ViewUser(),
		Granted:     ViewUserPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			SetUserActiveStatePermission,
			CreateUserPermissionName,
			UpdateUserPermissionName,
			UpdateUserPassPermissionName,
			SetAdminUserPermissionName,
		},
	})
}
