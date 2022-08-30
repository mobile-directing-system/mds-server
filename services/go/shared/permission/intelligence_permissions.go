package permission

// CreateIntelPermissionName for CreateIntel.
const CreateIntelPermissionName Name = "intelligence.intel.create"

// CreateIntel allows creation of intel.
func CreateIntel() Matcher {
	return Matcher{
		Name: "create-intel",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[CreateIntelPermissionName]
			return ok, nil
		},
	}
}

// InvalidateIntelPermissionName for InvalidateIntel.
const InvalidateIntelPermissionName Name = "intelligence.intel.invalidate"

// InvalidateIntel allows invalidating intel.
func InvalidateIntel() Matcher {
	return Matcher{
		Name: "invalidate-intel",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[InvalidateIntelPermissionName]
			return ok, nil
		},
	}
}

// ViewAnyIntelPermissionName for ViewAnyIntel.
const ViewAnyIntelPermissionName Name = "intelligence.intel.view.any"

// ViewAnyIntel allows viewing any intel, even if not assigned to.
func ViewAnyIntel() Matcher {
	return Matcher{
		Name: "view-any-intel",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[ViewAnyIntelPermissionName]
			return ok, nil
		},
	}
}
