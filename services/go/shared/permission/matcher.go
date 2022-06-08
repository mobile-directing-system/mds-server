package permission

// Matcher for a Permission list that checks, whether permissions are granted.
type Matcher func([]Permission) (bool, error)

// Has checks if the given Permission was wanted.
func Has(permissionToHave Permission) Matcher {
	return func(permissions []Permission) (bool, error) {
		for _, perm := range permissions {
			if perm == permissionToHave {
				return true, nil
			}
		}
		return false, nil
	}
}
