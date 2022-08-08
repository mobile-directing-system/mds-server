package search

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strconv"
	"testing"
)

// ParamsFromRequestSuite tests ParamsFromRequest.
type ParamsFromRequestSuite struct {
	suite.Suite
}

func (suite *ParamsFromRequestSuite) genContext(query, offset, limit string) *gin.Context {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	q.Add(QueryParamQuery, query)
	q.Add(QueryParamOffset, offset)
	q.Add(QueryParamLimit, limit)
	req.URL.RawQuery = q.Encode()
	return &gin.Context{Request: req}
}

func (suite *ParamsFromRequestSuite) TestBadOffset() {
	_, err := ParamsFromRequest(suite.genContext("", "abc", ""))
	suite.Error(err, "should fail")
}

func (suite *ParamsFromRequestSuite) TestBadLimit() {
	_, err := ParamsFromRequest(suite.genContext("", "", "abc"))
	suite.Error(err, "should fail")
}

func (suite *ParamsFromRequestSuite) TestOK1() {
	params, err := ParamsFromRequest(suite.genContext("1", "", ""))
	suite.Require().NoError(err, "should not fail")
	suite.Equal(Params{
		Query:  "1",
		Offset: 0,
		Limit:  0,
	}, params, "should return correct params")
}

func (suite *ParamsFromRequestSuite) TestOK2() {
	params, err := ParamsFromRequest(suite.genContext("", "53", ""))
	suite.Require().NoError(err, "should not fail")
	suite.Equal(Params{
		Query:  "",
		Offset: 53,
		Limit:  0,
	}, params, "should return correct params")
}

func (suite *ParamsFromRequestSuite) TestOK3() {
	params, err := ParamsFromRequest(suite.genContext("", "", "32"))
	suite.Require().NoError(err, "should not fail")
	suite.Equal(Params{
		Query:  "",
		Offset: 0,
		Limit:  32,
	}, params, "should return correct params")
}

func (suite *ParamsFromRequestSuite) TestOK4() {
	params, err := ParamsFromRequest(suite.genContext("1", "3", "2"))
	suite.Require().NoError(err, "should not fail")
	suite.Equal(Params{
		Query:  "1",
		Offset: 3,
		Limit:  2,
	}, params, "should return correct params")
}

func TestParamsFromRequest(t *testing.T) {
	suite.Run(t, new(ParamsFromRequestSuite))
}

// ParamsToQueryStringSuite tests ParamsToQueryString.
type ParamsToQueryStringSuite struct {
	suite.Suite
	sampleParams Params
}

func (suite *ParamsToQueryStringSuite) SetupTest() {
	suite.sampleParams = Params{
		Query:  "west",
		Offset: 55,
		Limit:  940,
	}
}

func (suite *ParamsToQueryStringSuite) TestEmpty() {
	suite.Equal("", ParamsToQueryString(Params{}))
}

func (suite *ParamsToQueryStringSuite) TestQuery() {
	p := Params{Query: suite.sampleParams.Query}
	expect := fmt.Sprintf("%s=%s", QueryParamQuery, p.Query)
	suite.Equal(expect, ParamsToQueryString(p))
}

func (suite *ParamsToQueryStringSuite) TestOffset() {
	p := Params{Offset: suite.sampleParams.Offset}
	expect := fmt.Sprintf("%s=%d", QueryParamOffset, p.Offset)
	suite.Equal(expect, ParamsToQueryString(p))
}

func (suite *ParamsToQueryStringSuite) TestLimit() {
	p := Params{Limit: suite.sampleParams.Limit}
	expect := fmt.Sprintf("%s=%d", QueryParamLimit, p.Limit)
	suite.Equal(expect, ParamsToQueryString(p))
}

func (suite *ParamsToQueryStringSuite) TestAll() {
	p := suite.sampleParams
	expect := QueryParamQuery + "=" + p.Query
	expect += "&" + QueryParamOffset + "=" + strconv.Itoa(p.Offset)
	expect += "&" + QueryParamLimit + "=" + strconv.Itoa(p.Limit)
	suite.Equal(expect, ParamsToQueryString(p))
}

func TestParamsToQueryString(t *testing.T) {
	suite.Run(t, new(ParamsToQueryStringSuite))
}
