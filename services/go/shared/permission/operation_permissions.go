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
