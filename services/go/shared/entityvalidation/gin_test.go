package entityvalidation

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

// validatableMock mocks Validatable.
type validatableMock struct {
	mock.Mock
}

func (m *validatableMock) Validate() (Report, error) {
	args := m.Called()
	return args.Get(0).(Report), args.Error(1)
}

// ValidateInRequestSuite tests ValidateInRequest.
type ValidateInRequestSuite struct {
	suite.Suite
	v *validatableMock
	r *gin.Engine
}

func (suite *ValidateInRequestSuite) SetupTest() {
	suite.v = &validatableMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.GET("/", func(c *gin.Context) {
		ok, err := ValidateInRequest(c, suite.v)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		if !ok {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusOK)
	})
}

func (suite *ValidateInRequestSuite) TestValidationFail() {
	suite.v.On("Validate").Return(Report{}, errors.New("sad life"))
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *ValidateInRequestSuite) TestNotOK() {
	report := NewReport()
	report.AddError("meow")
	report.AddError("sad life")
	suite.v.On("Validate").Return(report, nil)
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
	})
	suite.Require().Equal(http.StatusBadRequest, rr.Code, "should return correct code")
	var got publicReport
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(got, report.toPublic(), "should return expected response")
}

func (suite *ValidateInRequestSuite) TestOK() {
	suite.v.On("Validate").Return(NewReport(), nil)
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func TestValidateInRequest(t *testing.T) {
	suite.Run(t, new(ValidateInRequestSuite))
}
