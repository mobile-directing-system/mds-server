package permission

// CreateUser allows user creation.
const CreateUser Permission = "user.create"

// DeleteUser allows deletion of users.
const DeleteUser Permission = "user.delete"

// UpdateUser allows updating of user details without changing the password and
// is-admin-state. Of course, a user can always change its own password.
const UpdateUser Permission = "user.update"

// SetAdminUser allows setting the is-admin-state for users.
const SetAdminUser Permission = "user.set-admin"

// UpdateUserPass allows updating the password for other users.
const UpdateUserPass Permission = "user.update-pass"

// ViewUser allows viewing details regarding a user.
const ViewUser Permission = "user.view"
