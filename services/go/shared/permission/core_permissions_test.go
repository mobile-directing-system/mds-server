package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestRebuildSearchIndex(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "rebuild-search-index",
		Matcher:     RebuildSearchIndex(),
		Granted:     RebuildSearchIndexPermissionName,
		Others: []Name{
			CreateGroupPermissionName,
			ViewUserPermissionName,
			ViewPermissionsPermissionName,
		},
	})
}
