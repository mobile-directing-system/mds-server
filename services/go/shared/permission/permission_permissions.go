package permission

// UpdatePermissionsPermissionName for UpdatePermissions.
const UpdatePermissionsPermissionName Name = "permissions.update"

// UpdatePermissions allows setting the permissions of a user.
func UpdatePermissions() Matcher {
	return Matcher{
		Name: "update-permissions",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[UpdatePermissionsPermissionName]
			return ok, nil
		},
	}
}

// ViewPermissionsPermissionName for ViewPermissions.
const ViewPermissionsPermissionName Name = "permissions.view"

// ViewPermissions allows setting the permissions of a user.
func ViewPermissions() Matcher {
	return Matcher{
		Name: "view-permissions",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[ViewPermissionsPermissionName]
			return ok, nil
		},
	}
}
