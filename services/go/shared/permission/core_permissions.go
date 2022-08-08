package permission

// RebuildSearchIndexPermissionName for RebuildSearchIndex.
const RebuildSearchIndexPermissionName = "core.search.rebuild-index"

// RebuildSearchIndex allows performing a full rebuild on search indices.
func RebuildSearchIndex() Matcher {
	return Matcher{
		Name: "rebuild-search-index",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[RebuildSearchIndexPermissionName]
			return ok, nil
		},
	}
}
