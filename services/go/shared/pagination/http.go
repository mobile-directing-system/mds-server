package pagination

import (
	"fmt"
	"strings"
)

// ParamsToQueryString returns the given Params as HTTP request query string.
//
// Example: 'limit=12&offset=2'
func ParamsToQueryString(params Params) string {
	str := strings.Builder{}
	if params.Limit != 0 {
		str.WriteString(fmt.Sprintf("&%s=%d", LimitQueryParam, params.Limit))
	}
	if params.Offset != 0 {
		str.WriteString(fmt.Sprintf("&%s=%d", OffsetQueryParam, params.Offset))
	}
	if params.OrderBy.Valid {
		str.WriteString(fmt.Sprintf("&%s=%v", OrderByQueryParam, params.OrderBy.String))
	}
	if params.OrderDirection != "" {
		str.WriteString(fmt.Sprintf("&%s=%s", OrderDirQueryParam, params.OrderDirection))
	}
	return strings.TrimPrefix(str.String(), "&")
}
