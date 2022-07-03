package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestViewAnyOperation(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "view-any-operation",
		Matcher:     ViewAnyOperation(),
		Granted:     ViewAnyOperationPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			UpdateOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
		},
	})
}

func TestCreateOperation(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "create-operation",
		Matcher:     CreateOperation(),
		Granted:     CreateOperationPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			ViewAnyOperationPermissionName,
			UpdateOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
		},
	})
}

func TestUpdateOperation(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "update-operation",
		Matcher:     UpdateOperation(),
		Granted:     UpdateOperationPermissionName,
		Others: []Name{
			UpdateGroupPermissionName,
			CreateOperationPermissionName,
			ViewAnyOperationPermissionName,
			CreateUserPermissionName,
			ViewUserPermissionName,
		},
	})
}
