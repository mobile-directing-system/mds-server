package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
)

// handleGetAddressBookEntryByIDSuite tests handleGetAddressBookEntryByID.
type handleGetAddressBookEntryByIDSuite struct {
	suite.Suite
	s                 *StoreMock
	r                 *gin.Engine
	tokenOK           auth.Token
	sampleEntryID     uuid.UUID
	sampleStoreEntry  store.AddressBookEntryDetailed
	samplePublicEntry publicAddressBookEntryDetailed
}

func (suite *handleGetAddressBookEntryByIDSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "pound",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ViewAnyAddressBookEntryPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleEntryID = testutil.NewUUIDV4()
	suite.sampleStoreEntry = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          suite.sampleEntryID,
			Label:       "birth",
			Description: "correct",
			Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
			User:        nulls.NewUUID(testutil.NewUUIDV4()),
		},
		UserDetails: nulls.NewJSONNullable(store.User{
			ID:        uuid.UUID{},
			Username:  "rot",
			FirstName: "result",
			LastName:  "disgust",
			IsActive:  true,
		}),
	}
	suite.samplePublicEntry = publicAddressBookEntryDetailedFromStore(suite.sampleStoreEntry)
}

func (suite *handleGetAddressBookEntryByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetAddressBookEntryByIDSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetAddressBookEntryByIDSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries/abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetAddressBookEntryByIDSuite) TestRetrieveFail() {
	suite.s.On("AddressBookEntryByID", mock.Anything, suite.sampleEntryID, mock.Anything).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetAddressBookEntryByIDSuite) TestOK() {
	suite.s.On("AddressBookEntryByID", mock.Anything, suite.sampleEntryID, mock.Anything).
		Return(suite.sampleStoreEntry, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got publicAddressBookEntryDetailed
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicEntry, got, "should return correct body")
}

func (suite *handleGetAddressBookEntryByIDSuite) TestOKWithViewAnyPermission() {
	suite.tokenOK.Permissions = []permission.Permission{{Name: permission.ViewAnyAddressBookEntryPermissionName}}
	suite.s.On("AddressBookEntryByID", mock.Anything, suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.sampleStoreEntry, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleGetAddressBookEntryByIDSuite) TestOKWithoutViewAnyPermission() {
	suite.tokenOK.Permissions = nil
	suite.s.On("AddressBookEntryByID", mock.Anything, suite.sampleEntryID, nulls.NewUUID(suite.tokenOK.UserID)).
		Return(suite.sampleStoreEntry, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleGetAddressBookEntryByID(t *testing.T) {
	suite.Run(t, new(handleGetAddressBookEntryByIDSuite))
}

// handleGetAllAddressBookEntriesSuite tests handleGetAllAddressBookEntries.
type handleGetAllAddressBookEntriesSuite struct {
	suite.Suite
	s                   *StoreMock
	r                   *gin.Engine
	tokenOK             auth.Token
	sampleStoreEntries  pagination.Paginated[store.AddressBookEntryDetailed]
	samplePublicEntries pagination.Paginated[publicAddressBookEntryDetailed]
}

func (suite *handleGetAllAddressBookEntriesSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "jewel",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ViewAnyAddressBookEntryPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleStoreEntries = pagination.NewPaginated[store.AddressBookEntryDetailed](pagination.Params{}, []store.AddressBookEntryDetailed{
		{
			AddressBookEntry: store.AddressBookEntry{
				ID:          testutil.NewUUIDV4(),
				Label:       "shilling",
				Description: "sail",
				Operation:   uuid.NullUUID{},
				User:        uuid.NullUUID{},
			},
			UserDetails: nulls.JSONNullable[store.User]{},
		},
		{
			AddressBookEntry: store.AddressBookEntry{
				ID:          testutil.NewUUIDV4(),
				Label:       "discuss",
				Description: "strange",
				Operation:   nulls.NewUUID((testutil.NewUUIDV4())),
				User:        nulls.NewUUID(testutil.NewUUIDV4()),
			},
		},
	}, 14)
	suite.samplePublicEntries = pagination.MapPaginated(suite.sampleStoreEntries, publicAddressBookEntryDetailedFromStore)
}

func (suite *handleGetAllAddressBookEntriesSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries",
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries",
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestInvalidByUserFilter() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries?by_user=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestInvalidForOperationFilter() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries?for_operation=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestInvalidExcludeGlobalFilter() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries?exclude_global=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestInvalidVisibleByFilter() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries?visible_by=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestInvalidIncludeForInactiveUsersFilter() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries?include_for_inactive_users=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestInvalidPagination() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries?%s=abc", pagination.LimitQueryParam),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestRetrieveFail() {
	suite.s.On("AddressBookEntries", mock.Anything, mock.Anything, mock.Anything).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestOKResponse() {
	suite.s.On("AddressBookEntries", mock.Anything, mock.Anything, mock.Anything).
		Return(suite.sampleStoreEntries, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries",
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got pagination.Paginated[publicAddressBookEntryDetailed]
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicEntries, got, "should return correct body")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestOKParams1() {
	filters := store.AddressBookEntryFilters{
		ByUser:        nulls.NewUUID(testutil.NewUUIDV4()),
		ForOperation:  nulls.NewUUID(testutil.NewUUIDV4()),
		ExcludeGlobal: true,
		VisibleBy:     nulls.NewUUID(testutil.NewUUIDV4()),
	}
	paginationParams := pagination.Params{
		Limit:          14,
		Offset:         82,
		OrderBy:        nulls.NewString("label"),
		OrderDirection: pagination.OrderDirDesc,
	}
	suite.s.On("AddressBookEntries", mock.Anything, filters, paginationParams).
		Return(suite.sampleStoreEntries, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL: fmt.Sprintf("/address-book/entries?by_user=%s&for_operation=%s&exclude_global=true&visible_by=%s&%s",
			filters.ByUser.UUID.String(), filters.ForOperation.UUID.String(), filters.VisibleBy.UUID.String(),
			pagination.ParamsToQueryString(paginationParams)),
		Token: suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestOKParams2() {
	filters := store.AddressBookEntryFilters{
		ByUser:        uuid.NullUUID{},
		ForOperation:  uuid.NullUUID{},
		ExcludeGlobal: false,
		VisibleBy:     uuid.NullUUID{},
	}
	paginationParams := pagination.Params{
		Limit:          14,
		Offset:         82,
		OrderBy:        nulls.String{},
		OrderDirection: pagination.OrderDirAsc,
	}
	suite.s.On("AddressBookEntries", mock.Anything, filters, paginationParams).
		Return(suite.sampleStoreEntries, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries?%s", pagination.ParamsToQueryString(paginationParams)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestOKWithViewAnyPermission1() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{{Name: permission.ViewAnyAddressBookEntryPermissionName}}
	filters := store.AddressBookEntryFilters{
		VisibleBy: uuid.NullUUID{},
	}
	suite.s.On("AddressBookEntries", mock.Anything, filters, mock.Anything).
		Return(suite.sampleStoreEntries, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries",
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestOKWithViewAnyPermission2() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{{Name: permission.ViewAnyAddressBookEntryPermissionName}}
	filters := store.AddressBookEntryFilters{
		VisibleBy: nulls.NewUUID(token.UserID),
	}
	suite.s.On("AddressBookEntries", mock.Anything, filters, mock.Anything).
		Return(suite.sampleStoreEntries, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries?visible_by=%s", filters.VisibleBy.UUID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestOKWithoutViewAnyPermission1() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{}
	filters := store.AddressBookEntryFilters{
		VisibleBy: nulls.NewUUID(token.UserID),
	}
	suite.s.On("AddressBookEntries", mock.Anything, filters, mock.Anything).
		Return(suite.sampleStoreEntries, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries",
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleGetAllAddressBookEntriesSuite) TestOKWithoutViewAnyPermission2() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{}
	filters := store.AddressBookEntryFilters{
		VisibleBy: nulls.NewUUID(token.UserID),
	}
	suite.s.On("AddressBookEntries", mock.Anything, filters, mock.Anything).
		Return(suite.sampleStoreEntries, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries?visible_by=%s", testutil.NewUUIDV4().String()),
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleGetAllAddressBookEntries(t *testing.T) {
	suite.Run(t, new(handleGetAllAddressBookEntriesSuite))
}

// handleCreateAddressBookEntrySuite tests handleCreateAddressBookEntry.
type handleCreateAddressBookEntrySuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	tokenOK            auth.Token
	samplePublicCreate publicAddressBookEntry
	sampleStoreCreate  store.AddressBookEntry
}

func (suite *handleCreateAddressBookEntrySuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "within",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.CreateAnyAddressBookEntryPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleStoreCreate = store.AddressBookEntry{
		Label:       "insure",
		Description: "radio",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.samplePublicCreate = publicAddressBookEntryFromStore(suite.sampleStoreCreate)
}

func (suite *handleCreateAddressBookEntrySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateAddressBookEntrySuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleCreateAddressBookEntrySuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries",
		Body:   strings.NewReader(`{invalid`),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateAddressBookEntrySuite) TestMissingPermissionForGlobal() {
	token := suite.tokenOK
	token.Permissions = nil
	create := suite.samplePublicCreate
	create.User = uuid.NullUUID{}

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(create)),
		Token:  token,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleCreateAddressBookEntrySuite) TestMissingPermissionForOtherUser() {
	token := suite.tokenOK
	token.Permissions = nil
	create := suite.samplePublicCreate
	create.User = nulls.NewUUID(testutil.NewUUIDV4())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(create)),
		Token:  token,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleCreateAddressBookEntrySuite) TestCreateFail() {
	suite.s.On("CreateAddressBookEntry", mock.Anything, mock.Anything).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateAddressBookEntrySuite) TestOK() {
	created := store.AddressBookEntryDetailed{
		AddressBookEntry: suite.sampleStoreCreate,
	}
	created.ID = testutil.NewUUIDV4()
	suite.s.On("CreateAddressBookEntry", mock.Anything, suite.sampleStoreCreate).
		Return(created, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusCreated, rr.Code, "should return correct code")
	var got publicAddressBookEntryDetailed
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(publicAddressBookEntryDetailedFromStore(created), got, "should return correct body")
}

func Test_handleCreateAddressBookEntry(t *testing.T) {
	suite.Run(t, new(handleCreateAddressBookEntrySuite))
}

// handleUpdateAddressBookEntrySuite tests handleUpdateAddressBookEntry.
type handleUpdateAddressBookEntrySuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	tokenOK            auth.Token
	sampleEntryID      uuid.UUID
	samplePublicUpdate publicAddressBookEntry
	sampleStoreUpdate  store.AddressBookEntry
}

func (suite *handleUpdateAddressBookEntrySuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "within",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.UpdateAnyAddressBookEntryPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleEntryID = testutil.NewUUIDV4()
	suite.sampleStoreUpdate = store.AddressBookEntry{
		ID:          suite.sampleEntryID,
		Label:       "insure",
		Description: "radio",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.samplePublicUpdate = publicAddressBookEntryFromStore(suite.sampleStoreUpdate)
}

func (suite *handleUpdateAddressBookEntrySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateAddressBookEntrySuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleUpdateAddressBookEntrySuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/address-book/entries/abc",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateAddressBookEntrySuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Body:   strings.NewReader(`{invalid`),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateAddressBookEntrySuite) TestUpdateFail() {
	suite.s.On("UpdateAddressBookEntry", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateAddressBookEntrySuite) TestOK() {
	suite.s.On("UpdateAddressBookEntry", mock.Anything, suite.sampleStoreUpdate, mock.Anything).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleUpdateAddressBookEntrySuite) TestOKWithUpdateAnyPermission() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{{Name: permission.UpdateAnyAddressBookEntryPermissionName}}
	suite.s.On("UpdateAddressBookEntry", mock.Anything, suite.sampleStoreUpdate, uuid.NullUUID{}).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleUpdateAddressBookEntrySuite) TestOKWithoutUpdateAnyPermission() {
	token := suite.tokenOK
	token.Permissions = nil
	suite.s.On("UpdateAddressBookEntry", mock.Anything, suite.sampleStoreUpdate, nulls.NewUUID(token.UserID)).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleUpdateAddressBookEntry(t *testing.T) {
	suite.Run(t, new(handleUpdateAddressBookEntrySuite))
}

// handleDeleteAddressBookEntryByIDSuite tests handleDeleteAddressBookEntryByID.
type handleDeleteAddressBookEntryByIDSuite struct {
	suite.Suite
	s             *StoreMock
	r             *gin.Engine
	tokenOK       auth.Token
	sampleEntryID uuid.UUID
}

func (suite *handleDeleteAddressBookEntryByIDSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "within",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.DeleteAnyAddressBookEntryPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleEntryID = testutil.NewUUIDV4()
}

func (suite *handleDeleteAddressBookEntryByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDeleteAddressBookEntryByIDSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleDeleteAddressBookEntryByIDSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    "/address-book/entries/abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleDeleteAddressBookEntryByIDSuite) TestDeleteFail() {
	suite.s.On("DeleteAddressBookEntryByID", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDeleteAddressBookEntryByIDSuite) TestOK() {
	suite.s.On("DeleteAddressBookEntryByID", mock.Anything, suite.sampleEntryID, mock.Anything).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleDeleteAddressBookEntryByIDSuite) TestOKWithDeleteAnyPermission() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{{Name: permission.DeleteAnyAddressBookEntryPermissionName}}
	suite.s.On("DeleteAddressBookEntryByID", mock.Anything, suite.sampleEntryID, uuid.NullUUID{}).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleDeleteAddressBookEntryByIDSuite) TestOKWithoutDeleteAnyPermission() {
	token := suite.tokenOK
	token.Permissions = nil
	suite.s.On("DeleteAddressBookEntryByID", mock.Anything, suite.sampleEntryID, nulls.NewUUID(token.UserID)).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/address-book/entries/%s", suite.sampleEntryID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleDeleteAddressBookEntryByID(t *testing.T) {
	suite.Run(t, new(handleDeleteAddressBookEntryByIDSuite))
}
