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
			DeleteUserPermissionName,
			UpdateUserPermissionName,
			SetAdminUserPermissionName,
			ViewUserPermissionName,
			UpdateUserPassPermissionName,
		},
	})
}

func TestDeleteUser(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "delete-user",
		Matcher:     DeleteUser(),
		Granted:     DeleteUserPermissionName,
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
			DeleteUserPermissionName,
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
			DeleteUserPermissionName,
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
			DeleteUserPermissionName,
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
			DeleteUserPermissionName,
			CreateUserPermissionName,
			UpdateUserPermissionName,
			UpdateUserPassPermissionName,
			SetAdminUserPermissionName,
		},
	})
}
