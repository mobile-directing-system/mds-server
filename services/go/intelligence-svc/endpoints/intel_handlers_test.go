package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"testing"
	"time"
)

// Test_publicIntelTypeFromStore reads all constants of store.IntelType and
// assures that the mapper does not fail.
func Test_publicIntelTypeFromStore(t *testing.T) {
	testutil.TestMapperWithConstExtraction(t, publicIntelTypeFromStore, "../store/intel.go", nulls.String{})
}

// Test_storeIntelTypeFromPublic reads all constants of public intel-type and
// assures that the mapper does not fail.
func Test_storeIntelTypeFromPublic(t *testing.T) {
	testutil.TestMapperWithConstExtraction(t, storeIntelTypeFromPublic, "intel_handlers.go", nulls.NewString("intelType"))
}

// handleCreateIntelSuite tests handleCreateIntel.
type handleCreateIntelSuite struct {
	suite.Suite
	s                   *StoreMock
	r                   *gin.Engine
	tokenOK             auth.Token
	sampleStoreCreate   store.CreateIntel
	samplePublicCreate  publicCreateIntel
	sampleStoreCreated  store.Intel
	samplePublicCreated publicIntel
}

func (suite *handleCreateIntelSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "cloud",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.CreateIntelPermissionName}},
		RandomSalt:      nil,
	}
	suite.samplePublicCreate = publicCreateIntel{
		Operation:  testutil.NewUUIDV4(),
		Type:       intelTypePlainTextMessage,
		Content:    json.RawMessage(`{"hello":"world"}`),
		SearchText: nulls.NewString("hello world"),
		Assignments: []publicIntelAssignment{
			{
				ID: uuid.UUID{},
				To: testutil.NewUUIDV4(),
			},
			{
				ID: uuid.UUID{},
				To: testutil.NewUUIDV4(),
			},
		},
	}
	suite.sampleStoreCreate = store.CreateIntel{
		CreatedBy:  suite.tokenOK.UserID,
		Operation:  suite.samplePublicCreate.Operation,
		Type:       store.IntelTypePlainTextMessage,
		Content:    json.RawMessage(`{"hello":"world"}`),
		SearchText: nulls.NewString("hello world"),
		Assignments: []store.IntelAssignment{
			{
				ID:    uuid.UUID{},
				Intel: uuid.UUID{},
				To:    suite.samplePublicCreate.Assignments[0].To,
			},
			{
				ID:    uuid.UUID{},
				Intel: uuid.UUID{},
				To:    suite.samplePublicCreate.Assignments[1].To,
			},
		},
	}
	suite.sampleStoreCreated = store.Intel{
		ID:         testutil.NewUUIDV4(),
		CreatedAt:  time.Now().UTC(),
		CreatedBy:  suite.sampleStoreCreate.CreatedBy,
		Operation:  suite.sampleStoreCreate.Operation,
		Type:       suite.sampleStoreCreate.Type,
		Content:    suite.sampleStoreCreate.Content,
		SearchText: suite.sampleStoreCreate.SearchText,
		IsValid:    true,
	}
	suite.sampleStoreCreated.Assignments = []store.IntelAssignment{
		{
			ID:    testutil.NewUUIDV4(),
			Intel: suite.sampleStoreCreated.ID,
			To:    suite.sampleStoreCreate.Assignments[0].To,
		},
		{
			ID:    testutil.NewUUIDV4(),
			Intel: suite.sampleStoreCreated.ID,
			To:    suite.sampleStoreCreate.Assignments[1].To,
		},
	}
	suite.samplePublicCreated = publicIntel{
		ID:         suite.sampleStoreCreated.ID,
		CreatedAt:  suite.sampleStoreCreated.CreatedAt,
		CreatedBy:  suite.sampleStoreCreated.CreatedBy,
		Operation:  suite.sampleStoreCreated.Operation,
		Type:       intelTypePlainTextMessage,
		Content:    suite.sampleStoreCreated.Content,
		SearchText: suite.sampleStoreCreated.SearchText,
		IsValid:    true,
		Assignments: []publicIntelAssignment{
			{
				ID: suite.sampleStoreCreated.Assignments[0].ID,
				To: suite.sampleStoreCreated.Assignments[0].To,
			},
			{
				ID: suite.sampleStoreCreated.Assignments[1].ID,
				To: suite.sampleStoreCreated.Assignments[1].To,
			},
		},
	}
}

func (suite *handleCreateIntelSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel",
		Body:   bytes.NewReader([]byte(`{invalid`)),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelSuite) TestUnsupportedType() {
	suite.samplePublicCreate.Type = publicIntelType(testutil.NewUUIDV4().String())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelSuite) TestInvalidIntel() {
	// Duplicate assignment to same recipient.
	suite.samplePublicCreate.Assignments[1] = suite.samplePublicCreate.Assignments[0]

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelSuite) TestCreateFail() {
	suite.s.On("CreateIntel", mock.Anything, suite.sampleStoreCreate).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelSuite) TestStoreToPublicConversionFail() {
	suite.sampleStoreCreated.Type = store.IntelType(testutil.NewUUIDV4().String())
	suite.s.On("CreateIntel", mock.Anything, suite.sampleStoreCreate).
		Return(suite.sampleStoreCreated, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelSuite) TestOK() {
	suite.s.On("CreateIntel", mock.Anything, suite.sampleStoreCreate).
		Return(suite.sampleStoreCreated, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got publicIntel
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicCreated, got, "should return expected body")
}

func Test_handleCreateIntel(t *testing.T) {
	suite.Run(t, new(handleCreateIntelSuite))
}

// handleInvalidateIntelByIDSuite tests handleInvalidateIntelByID.
type handleInvalidateIntelByIDSuite struct {
	suite.Suite
	s        *StoreMock
	r        *gin.Engine
	tokenOK  auth.Token
	sampleID uuid.UUID
}

func (suite *handleInvalidateIntelByIDSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "arrow",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.InvalidateIntelPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleID = testutil.NewUUIDV4()
}

func (suite *handleInvalidateIntelByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel/%s/invalidate", suite.sampleID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleInvalidateIntelByIDSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel/%s/invalidate", suite.sampleID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleInvalidateIntelByIDSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel/abc/invalidate",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleInvalidateIntelByIDSuite) TestInvalidateFail() {
	suite.s.On("InvalidateIntelByID", mock.Anything, suite.sampleID, suite.tokenOK.UserID).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel/%s/invalidate", suite.sampleID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleInvalidateIntelByIDSuite) TestOK() {
	suite.s.On("InvalidateIntelByID", mock.Anything, suite.sampleID, suite.tokenOK.UserID).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel/%s/invalidate", suite.sampleID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleInvalidateIntelByID(t *testing.T) {
	suite.Run(t, new(handleInvalidateIntelByIDSuite))
}

// handleGetIntelByIDSuite tests handleGetIntelByID.
type handleGetIntelByIDSuite struct {
	suite.Suite
	s                 *StoreMock
	r                 *gin.Engine
	tokenOK           auth.Token
	sampleID          uuid.UUID
	sampleStoreIntel  store.Intel
	samplePublicIntel publicIntel
}

func (suite *handleGetIntelByIDSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "fair",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.InvalidateIntelPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleID = testutil.NewUUIDV4()
	suite.sampleStoreIntel = store.Intel{
		ID:         suite.sampleID,
		CreatedAt:  time.Now().UTC(),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       store.IntelTypePlainTextMessage,
		Content:    json.RawMessage(`null`),
		SearchText: nulls.NewString("Hello World!"),
		IsValid:    true,
		Assignments: []store.IntelAssignment{
			{
				ID:    testutil.NewUUIDV4(),
				Intel: suite.sampleID,
				To:    testutil.NewUUIDV4(),
			},
			{
				ID:    testutil.NewUUIDV4(),
				Intel: suite.sampleID,
				To:    testutil.NewUUIDV4(),
			},
		},
	}
	suite.samplePublicIntel = publicIntel{
		ID:         suite.sampleID,
		CreatedAt:  suite.sampleStoreIntel.CreatedAt,
		CreatedBy:  suite.sampleStoreIntel.CreatedBy,
		Operation:  suite.sampleStoreIntel.Operation,
		Type:       intelTypePlainTextMessage,
		Content:    json.RawMessage(`null`),
		SearchText: nulls.NewString("Hello World!"),
		IsValid:    true,
		Assignments: []publicIntelAssignment{
			{
				ID: suite.sampleStoreIntel.Assignments[0].ID,
				To: suite.sampleStoreIntel.Assignments[0].To,
			},
			{
				ID: suite.sampleStoreIntel.Assignments[1].ID,
				To: suite.sampleStoreIntel.Assignments[1].To,
			},
		},
	}
}

func (suite *handleGetIntelByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel/%s", suite.sampleID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetIntelByIDSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel/%s", suite.sampleID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetIntelByIDSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/intel/abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetIntelByIDSuite) TestRetrieveFail() {
	suite.s.On("IntelByID", mock.Anything, suite.sampleID, mock.Anything).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel/%s", suite.sampleID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetIntelByIDSuite) TestUnsupportedStoreIntelType() {
	suite.sampleStoreIntel.Type = store.IntelType(testutil.NewUUIDV4().String())
	suite.s.On("IntelByID", mock.Anything, suite.sampleID, mock.Anything).
		Return(suite.sampleStoreIntel, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel/%s", suite.sampleID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetIntelByIDSuite) TestOK() {
	suite.s.On("IntelByID", mock.Anything, suite.sampleID, mock.Anything).
		Return(suite.sampleStoreIntel, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel/%s", suite.sampleID.String()),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got publicIntel
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicIntel, got, "should return correct body")
}

func Test_handleGetIntelByID(t *testing.T) {
	suite.Run(t, new(handleGetIntelByIDSuite))
}
