package search

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"strconv"
	"strings"
)

// Params to use for performing a search.
type Params struct {
	Query  string
	Offset int
	Limit  int
}

// Query param names for ParamsFromRequest and ParamsToQueryString.
const (
	QueryParamQuery  = "q"
	QueryParamOffset = "offset"
	QueryParamLimit  = "limit"
)

// ParamsFromRequest extracts Params from the given gin.Context.
func ParamsFromRequest(c *gin.Context) (Params, error) {
	var err error
	params := Params{}
	params.Query = c.Query(QueryParamQuery)
	offsetStr := c.Query(QueryParamOffset)
	if offsetStr != "" {
		params.Offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return Params{}, meh.NewBadInputErrFromErr(err, "parse offset", meh.Details{"was": offsetStr})
		}
	}
	limitStr := c.Query(QueryParamLimit)
	if limitStr != "" {
		params.Limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return Params{}, meh.NewBadInputErrFromErr(err, "parse limit", meh.Details{"was": limitStr})
		}
	}
	return params, nil
}

// ParamsToQueryString returns the list of query params for the given Params,
// concatenated with '&'.
func ParamsToQueryString(p Params) string {
	qp := make([]string, 0)
	if p.Query != "" {
		qp = append(qp, fmt.Sprintf("%s=%s", QueryParamQuery, p.Query))
	}
	if p.Offset != 0 {
		qp = append(qp, fmt.Sprintf("%s=%d", QueryParamOffset, p.Offset))
	}
	if p.Limit != 0 {
		qp = append(qp, fmt.Sprintf("%s=%d", QueryParamLimit, p.Limit))
	}
	return strings.Join(qp, "&")
}
