package permission

// CreateUserPermissionName for CreateUser.
const CreateUserPermissionName Name = "user.create"

// CreateUser allows creating users. For creating admin-users, the SetAdminUserPermissionName
// permission is needed as well.
func CreateUser() Matcher {
	return Matcher{
		Name: "create-user",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[CreateUserPermissionName]
			return ok, nil
		},
	}
}

// DeleteUserPermissionName for DeleteUser.
const DeleteUserPermissionName Name = "user.delete"

// DeleteUser allows deletion of users.
func DeleteUser() Matcher {
	return Matcher{
		Name: "delete-user",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[DeleteUserPermissionName]
			return ok, nil
		},
	}
}

// UpdateUserPermissionName for UpdateUser.
const UpdateUserPermissionName Name = "user.update"

// UpdateUser allows updating of user details without changing the password and
// is-admin-state. Of course, a user can always change its own password.
func UpdateUser() Matcher {
	return Matcher{
		Name: "update-user",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[UpdateUserPermissionName]
			return ok, nil
		},
	}
}

// SetAdminUserPermissionName for SetAdminUser.
const SetAdminUserPermissionName Name = "user.set-admin"

// SetAdminUser allows setting the is-admin-state for users.
func SetAdminUser() Matcher {
	return Matcher{
		Name: "set-admin-user",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[SetAdminUserPermissionName]
			return ok, nil
		},
	}
}

// UpdateUserPassPermissionName for UpdateUserPass.
const UpdateUserPassPermissionName Name = "user.update-pass"

// UpdateUserPass allows updating the password for other users.
func UpdateUserPass() Matcher {
	return Matcher{
		Name: "update-user-pass",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[UpdateUserPassPermissionName]
			return ok, nil
		},
	}
}

// ViewUserPermissionName for ViewUser.
const ViewUserPermissionName Name = "user.view"

// ViewUser allows viewing details regarding foreign users.
func ViewUser() Matcher {
	return Matcher{
		Name: "view-user",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[ViewUserPermissionName]
			return ok, nil
		},
	}
}
