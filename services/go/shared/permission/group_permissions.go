package permission

// CreateGroupPermissionName for CreateGroup.
const CreateGroupPermissionName Name = "group.create"

// CreateGroup allows creation of groups.
func CreateGroup() Matcher {
	return Matcher{
		Name: "create-group",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[CreateGroupPermissionName]
			return ok, nil
		},
	}
}

// UpdateGroupPermissionName for UpdateGroup.
const UpdateGroupPermissionName Name = "group.update"

// UpdateGroup allows updating of groups.
func UpdateGroup() Matcher {
	return Matcher{
		Name: "update-group",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[UpdateGroupPermissionName]
			return ok, nil
		},
	}
}

// DeleteGroupPermissionName for DeleteGroup.
const DeleteGroupPermissionName = "group.delete"

// DeleteGroup allows deleting groups.
func DeleteGroup() Matcher {
	return Matcher{
		Name: "delete-group",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[DeleteGroupPermissionName]
			return ok, nil
		},
	}
}

// ViewGroupPermissionName for ViewGroup.
const ViewGroupPermissionName Name = "group.view"

// ViewGroup allows retrieval of all groups.
func ViewGroup() Matcher {
	return Matcher{
		Name: "view-group",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[ViewGroupPermissionName]
			return ok, nil
		},
	}
}
