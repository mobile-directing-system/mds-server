package pagination

import (
	"github.com/gin-gonic/gin"
	"github.com/lefinal/nulls"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

// ParamsFromRequestSuite tests ParamsFromRequest.
type ParamsFromRequestSuite struct {
	suite.Suite
}

func (suite *ParamsFromRequestSuite) genContext(limit, offset, orderBy, orderDir string) *gin.Context {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	q.Add(LimitQueryParam, limit)
	q.Add(OffsetQueryParam, offset)
	q.Add(OrderByQueryParam, orderBy)
	q.Add(OrderDirQueryParam, orderDir)
	req.URL.RawQuery = q.Encode()
	return &gin.Context{Request: req}
}

func (suite *ParamsFromRequestSuite) TestBadLimit() {
	_, err := ParamsFromRequest(suite.genContext("hello", "", "", ""))
	suite.Error(err, "should fail")
}

func (suite *ParamsFromRequestSuite) TestBadOffset() {
	_, err := ParamsFromRequest(suite.genContext("", "hello", "", ""))
	suite.Error(err, "should fail")
}

func (suite *ParamsFromRequestSuite) TestOK1() {
	params, err := ParamsFromRequest(suite.genContext("1", "", "hello", ""))
	suite.Require().NoError(err, "should not fail")
	suite.Equal(Params{
		Limit:          nulls.NewInt(1),
		Offset:         0,
		OrderBy:        nulls.NewString("hello"),
		OrderDirection: "",
	}, params, "should return correct params")
}

func (suite *ParamsFromRequestSuite) TestOK2() {
	params, err := ParamsFromRequest(suite.genContext("", "32", "", "desc"))
	suite.Require().NoError(err, "should not fail")
	suite.Equal(Params{
		Limit:          nulls.Int{},
		Offset:         32,
		OrderBy:        nulls.String{},
		OrderDirection: "desc",
	}, params, "should return correct params")
}

func (suite *ParamsFromRequestSuite) TestOK3() {
	params, err := ParamsFromRequest(suite.genContext("1", "1", "", ""))
	suite.Require().NoError(err, "should not fail")
	suite.Equal(Params{
		Limit:          nulls.NewInt(1),
		Offset:         1,
		OrderBy:        nulls.String{},
		OrderDirection: "",
	}, params, "should return correct params")
}

func (suite *ParamsFromRequestSuite) TestOK4() {
	params, err := ParamsFromRequest(suite.genContext("1", "2", "3", "4"))
	suite.Require().NoError(err, "should not fail")
	suite.Equal(Params{
		Limit:          nulls.NewInt(1),
		Offset:         2,
		OrderBy:        nulls.NewString("3"),
		OrderDirection: "4",
	}, params, "should return correct params")
}

func TestParamsFromRequest(t *testing.T) {
	suite.Run(t, new(ParamsFromRequestSuite))
}
