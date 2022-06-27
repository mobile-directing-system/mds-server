package permission

import "github.com/lefinal/meh"

// Matcher for a Permission list that checks, whether permissions are granted.
type Matcher struct {
	// Name for better readability.
	Name string
	// MatchFn matches the given Permission list against criteria and returns
	// whether a permission was given or not.
	MatchFn func(granted map[Name]Permission) (bool, error)
}

// Has checks if the given Permission was wanted.
func Has(granted []Permission, toHave ...Matcher) (bool, error) {
	// Build map.
	permissions := make(map[Name]Permission, len(granted))
	for _, permission := range granted {
		permissions[permission.Name] = permission
	}
	// Match.
	for _, matcher := range toHave {
		ok, err := matcher.MatchFn(permissions)
		if err != nil {
			return false, meh.Wrap(err, "match permission", meh.Details{"matcher_name": matcher.Name})
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// Assure works similarly to Has but returns the result as error. If a Matcher
// returns not-ok, an meh.Forbidden error will be returned. If a Matcher fails,
// an meh.ErrInternal will be returned.
func Assure(granted []Permission, toHave ...Matcher) error {
	// Build map.
	permissions := make(map[Name]Permission, len(granted))
	for _, permission := range granted {
		permissions[permission.Name] = permission
	}
	// Match.
	for i, matcher := range toHave {
		ok, err := matcher.MatchFn(permissions)
		if err != nil {
			return meh.ApplyCode(meh.Wrap(err, "match permission", meh.Details{
				"matcher_name": matcher.Name,
				"matcher_pos":  i,
			}), meh.ErrInternal)
		}
		if !ok {
			return meh.NewForbiddenErr("permission not granted", meh.Details{
				"matcher_name": matcher.Name,
				"matcher_pos":  i,
			})
		}
	}
	return nil
}
