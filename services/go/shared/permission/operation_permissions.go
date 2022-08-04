package permission

// ViewAnyOperationPermissionName for ViewAnyOperation.
const ViewAnyOperationPermissionName Name = "operation.view.any"

// ViewAnyOperation allows listing and viewing all operations.
func ViewAnyOperation() Matcher {
	return Matcher{
		Name: "view-any-operation",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[ViewAnyOperationPermissionName]
			return ok, nil
		},
	}
}

// CreateOperationPermissionName for CreateOperation.
const CreateOperationPermissionName Name = "operation.create"

// CreateOperation allows creation of operations.
func CreateOperation() Matcher {
	return Matcher{
		Name: "create-operation",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[CreateOperationPermissionName]
			return ok, nil
		},
	}
}

// UpdateOperationPermissionName for UpdateOperation.
const UpdateOperationPermissionName Name = "operation.update"

// UpdateOperation allows updating of operations.
func UpdateOperation() Matcher {
	return Matcher{
		Name: "update-operation",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[UpdateOperationPermissionName]
			return ok, nil
		},
	}
}

// ViewOperationMembersPermissionName for ViewOperationMembers.
const ViewOperationMembersPermissionName Name = "operation.members.view"

// ViewOperationMembers allows retrieving members for an operation.
func ViewOperationMembers() Matcher {
	return Matcher{
		Name: "view-operation-members",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[ViewOperationMembersPermissionName]
			return ok, nil
		},
	}
}

// UpdateOperationMembersPermissionName for UpdateOperationMembers.
const UpdateOperationMembersPermissionName Name = "operation.members.update"

// UpdateOperationMembers allows setting members for an operation.
func UpdateOperationMembers() Matcher {
	return Matcher{
		Name: "update-operation-members",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[UpdateOperationMembersPermissionName]
			return ok, nil
		},
	}
}
