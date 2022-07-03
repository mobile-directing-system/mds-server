package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestUpdatePermissions(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "update-permissions",
		Matcher:     UpdatePermissions(),
		Granted:     UpdatePermissionsPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
			ViewPermissionsPermissionName,
		},
	})
}

func TestViewPermissions(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "view-permissions",
		Matcher:     ViewPermissions(),
		Granted:     ViewPermissionsPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
			UpdatePermissionsPermissionName,
		},
	})
}
