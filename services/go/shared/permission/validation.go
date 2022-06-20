package permission

import "github.com/lefinal/meh"

// validator accepts the granted permissions and returns a list of validation
// errors. An empty list describes, that no issues where found.
type validator func(granted map[Name]Permission) ([]string, error)

// Validate validates the given Permission list and returns a list of validation
// errors. An empty list describes, that no issues where found.
func Validate(granted []Permission) ([]string, error) {
	// Build map.
	grantedMap := make(map[Name]Permission)
	for _, permission := range granted {
		grantedMap[permission.Name] = permission
	}
	// Validate.
	validationErrors := make([]string, 0)
	validators := validators()
	for vName, v := range validators {
		vErrs, err := v(grantedMap)
		if err != nil {
			return nil, meh.Wrap(err, "run validator", meh.Details{"validator_name": vName})
		}
		validationErrors = append(validationErrors, vErrs...)
	}
	return validationErrors, nil
}

// validators provides a central location to for registering validators.
func validators() map[string]validator {
	return map[string]validator{}
}
