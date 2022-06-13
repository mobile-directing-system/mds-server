package permission

import "github.com/lefinal/meh"

// Matcher for a Permission list that checks, whether permissions are granted.
type Matcher func(grantedPermissions []Permission) (bool, error)

// Has checks if the given Permission was wanted.
func Has(permissionsToHave ...Permission) Matcher {
	return func(grantedPermissions []Permission) (bool, error) {
		if len(permissionsToHave) == 0 {
			return true, nil
		}
		// If only one permission needed, we can simply search the list.
		if len(permissionsToHave) == 1 {
			permissionToHave := permissionsToHave[0]
			for _, perm := range grantedPermissions {
				if perm == permissionToHave {
					return true, nil
				}
			}
			return false, nil
		}
		// For multiple expected ones:
		ok, err := hasAll(permissionsToHave)(grantedPermissions)
		if err != nil {
			return false, meh.Wrap(err, "has all", meh.Details{"permissions_to_have": permissionsToHave})
		}
		return ok, nil
	}
}

// hasAll returns a Matcher, that assures all the given permissions to have been
// granted.
func hasAll(permissionsToHave []Permission) Matcher {
	return func(permissions []Permission) (bool, error) {
		remainingPermissionsToHave := make(map[Permission]struct{})
		for _, permissionToHave := range permissionsToHave {
			remainingPermissionsToHave[permissionToHave] = struct{}{}
		}
		for _, grantedPermission := range permissions {
			delete(remainingPermissionsToHave, grantedPermission)
		}
		return len(remainingPermissionsToHave) == 0, nil
	}
}
