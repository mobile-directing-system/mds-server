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
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

type addressBookEntryFiltersFromQuerySuite struct {
	suite.Suite
	sampleFilters store.AddressBookEntryFilters
	qOK           url.Values
}

func (suite *addressBookEntryFiltersFromQuerySuite) SetupTest() {
	suite.sampleFilters = store.AddressBookEntryFilters{
		ByUser:                  nulls.NewUUID(testutil.NewUUIDV4()),
		ForOperation:            nulls.NewUUID(testutil.NewUUIDV4()),
		ExcludeGlobal:           true,
		VisibleBy:               nulls.NewUUID(testutil.NewUUIDV4()),
		IncludeForInactiveUsers: true,
	}
	suite.qOK = map[string][]string{
		"by_user":                    {suite.sampleFilters.ByUser.UUID.String()},
		"for_operation":              {suite.sampleFilters.ForOperation.UUID.String()},
		"exclude_global":             {fmt.Sprintf("%t", suite.sampleFilters.ExcludeGlobal)},
		"visible_by":                 {suite.sampleFilters.VisibleBy.UUID.String()},
		"include_for_inactive_users": {fmt.Sprintf("%t", suite.sampleFilters.IncludeForInactiveUsers)},
	}
}

func (suite *addressBookEntryFiltersFromQuerySuite) TestInvalidByUserFilter() {
	q := suite.qOK
	q["by_user"] = []string{"abc"}
	_, err := addressBookEntryFiltersFromQuery(q)
	suite.Error(err, "should fail")
}

func (suite *addressBookEntryFiltersFromQuerySuite) TestInvalidForOperationFilter() {
	q := suite.qOK
	q["for_operation"] = []string{"abc"}
	_, err := addressBookEntryFiltersFromQuery(q)
	suite.Error(err, "should fail")
}

func (suite *addressBookEntryFiltersFromQuerySuite) TestInvalidExcludeGlobalFilter() {
	q := suite.qOK
	q["exclude_global"] = []string{"abc"}
	_, err := addressBookEntryFiltersFromQuery(q)
	suite.Error(err, "should fail")
}

func (suite *addressBookEntryFiltersFromQuerySuite) TestInvalidVisibleByFilter() {
	q := suite.qOK
	q["visible_by"] = []string{"abc"}
	_, err := addressBookEntryFiltersFromQuery(q)
	suite.Error(err, "should fail")
}

func (suite *addressBookEntryFiltersFromQuerySuite) TestInvalidIncludeForInactiveUsersFilter() {
	q := suite.qOK
	q["include_for_inactive_users"] = []string{"abc"}
	_, err := addressBookEntryFiltersFromQuery(q)
	suite.Error(err, "should fail")
}

func (suite *addressBookEntryFiltersFromQuerySuite) TestOK() {
	got, err := addressBookEntryFiltersFromQuery(suite.qOK)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(suite.sampleFilters, got, "should return correct value")
}

func Test_addressBookEntryFiltersFromQuery(t *testing.T) {
	suite.Run(t, new(addressBookEntryFiltersFromQuerySuite))
}

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

func (suite *handleGetAllAddressBookEntriesSuite) TestInvalidFilter() {
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
	suite.s.On("DeleteAddressBookEntryWithChannelsByID", mock.Anything, mock.Anything, mock.Anything).
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
	suite.s.On("DeleteAddressBookEntryWithChannelsByID", mock.Anything, suite.sampleEntryID, mock.Anything).
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
	suite.s.On("DeleteAddressBookEntryWithChannelsByID", mock.Anything, suite.sampleEntryID, uuid.NullUUID{}).
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
	suite.s.On("DeleteAddressBookEntryWithChannelsByID", mock.Anything, suite.sampleEntryID, nulls.NewUUID(token.UserID)).
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

// handleSearchAddressBookEntriesSuite tests handleSearchAddressBookEntries.
type handleSearchAddressBookEntriesSuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	tokenOK            auth.Token
	sampleResult       search.Result[store.AddressBookEntryDetailed]
	samplePublicResult search.Result[publicAddressBookEntryDetailed]
	sampleFilters      store.AddressBookEntryFilters
	sampleParams       search.Params
	sampleURL          string
}

func (suite *handleSearchAddressBookEntriesSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "future",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     nil,
		RandomSalt:      nil,
	}
	suite.sampleResult = search.Result[store.AddressBookEntryDetailed]{
		Hits: []store.AddressBookEntryDetailed{
			{
				AddressBookEntry: store.AddressBookEntry{
					ID:          testutil.NewUUIDV4(),
					Label:       "bribe",
					Description: "ground",
					Operation:   uuid.NullUUID{},
					User:        uuid.NullUUID{},
				},
				UserDetails: nulls.JSONNullable[store.User]{},
			},
			{
				AddressBookEntry: store.AddressBookEntry{
					ID:          testutil.NewUUIDV4(),
					Label:       "unite",
					Description: "lessen",
					Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
					User:        nulls.NewUUID(testutil.NewUUIDV4()),
				},
				UserDetails: nulls.NewJSONNullable(store.User{
					ID:        testutil.NewUUIDV4(),
					Username:  "anything",
					FirstName: "solemn",
					LastName:  "tough",
					IsActive:  true,
				}),
			},
		},
	}
	suite.samplePublicResult = search.Result[publicAddressBookEntryDetailed]{
		Hits: []publicAddressBookEntryDetailed{
			{
				publicAddressBookEntry: publicAddressBookEntry{
					ID:          suite.sampleResult.Hits[0].ID,
					Label:       suite.sampleResult.Hits[0].Label,
					Description: suite.sampleResult.Hits[0].Description,
					Operation:   suite.sampleResult.Hits[0].Operation,
					User:        suite.sampleResult.Hits[0].User,
				},
				UserDetails: nulls.JSONNullable[publicUser]{},
			},
			{
				publicAddressBookEntry: publicAddressBookEntry{
					ID:          suite.sampleResult.Hits[1].ID,
					Label:       suite.sampleResult.Hits[1].Label,
					Description: suite.sampleResult.Hits[1].Description,
					Operation:   suite.sampleResult.Hits[1].Operation,
					User:        suite.sampleResult.Hits[1].User,
				},
				UserDetails: nulls.NewJSONNullable(publicUser{
					ID:        suite.sampleResult.Hits[1].UserDetails.V.ID,
					Username:  suite.sampleResult.Hits[1].UserDetails.V.Username,
					FirstName: suite.sampleResult.Hits[1].UserDetails.V.FirstName,
					LastName:  suite.sampleResult.Hits[1].UserDetails.V.LastName,
					IsActive:  suite.sampleResult.Hits[1].UserDetails.V.IsActive,
				}),
			},
		},
	}
	suite.sampleFilters = store.AddressBookEntryFilters{
		ByUser:                  nulls.NewUUID(testutil.NewUUIDV4()),
		ForOperation:            nulls.NewUUID(testutil.NewUUIDV4()),
		ExcludeGlobal:           true,
		VisibleBy:               nulls.NewUUID(testutil.NewUUIDV4()),
		IncludeForInactiveUsers: true,
	}
	suite.sampleParams = search.Params{
		Query:  "brick",
		Offset: 480,
		Limit:  114,
	}
	q := fmt.Sprintf("q=%s", suite.sampleParams.Query)
	q += fmt.Sprintf("&offset=%d", suite.sampleParams.Offset)
	q += fmt.Sprintf("&limit=%d", suite.sampleParams.Limit)
	q += fmt.Sprintf("&by_user=%s", suite.sampleFilters.ByUser.UUID.String())
	q += fmt.Sprintf("&for_operation=%s", suite.sampleFilters.ForOperation.UUID.String())
	q += fmt.Sprintf("&exclude_global=%t", suite.sampleFilters.ExcludeGlobal)
	q += fmt.Sprintf("&visible_by=%s", suite.sampleFilters.VisibleBy.UUID.String())
	q += fmt.Sprintf("&include_for_inactive_users=%t", suite.sampleFilters.IncludeForInactiveUsers)
	suite.sampleURL = fmt.Sprintf("/address-book/entries/search?%s", q)
}

func (suite *handleSearchAddressBookEntriesSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    suite.sampleURL,
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleSearchAddressBookEntriesSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    suite.sampleURL,
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleSearchAddressBookEntriesSuite) TestInvalidFilterParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries/search?by_user=abc",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleSearchAddressBookEntriesSuite) TestInvalidSearchParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries/search?limit=abc",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleSearchAddressBookEntriesSuite) TestSearchFail() {
	suite.s.On("SearchAddressBookEntries", mock.Anything, mock.Anything, mock.Anything).
		Return(search.Result[store.AddressBookEntryDetailed]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    suite.sampleURL,
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleSearchAddressBookEntriesSuite) TestOK() {
	suite.sampleFilters.VisibleBy = nulls.NewUUID(suite.tokenOK.UserID)
	suite.s.On("SearchAddressBookEntries", mock.Anything, suite.sampleFilters, suite.sampleParams).
		Return(suite.sampleResult, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    suite.sampleURL,
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got search.Result[publicAddressBookEntryDetailed]
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicResult, got, "should return correct body")
}

func (suite *handleSearchAddressBookEntriesSuite) TestOKWithViewAny() {
	suite.tokenOK.Permissions = []permission.Permission{{Name: permission.ViewAnyAddressBookEntryPermissionName}}
	suite.s.On("SearchAddressBookEntries", mock.Anything, suite.sampleFilters, suite.sampleParams).
		Return(suite.sampleResult, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    suite.sampleURL,
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleSearchAddressBookEntries(t *testing.T) {
	suite.Run(t, new(handleSearchAddressBookEntriesSuite))
}

// handleRebuildAddressBookEntrySearchSuite tests
// handleRebuildAddressBookEntrySearch.
type handleRebuildAddressBookEntrySearchSuite struct {
	suite.Suite
	s       *StoreMock
	r       *gin.Engine
	tokenOK auth.Token
}

func (suite *handleRebuildAddressBookEntrySearchSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.POST("/search/rebuild", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleRebuildAddressBookEntrySearch(suite.s)))
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "organ",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.RebuildSearchIndexPermissionName}},
		RandomSalt:      nil,
	}
}

func (suite *handleRebuildAddressBookEntrySearchSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleRebuildAddressBookEntrySearchSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleRebuildAddressBookEntrySearchSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleRebuildAddressBookEntrySearchSuite) TestOK() {
	_, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.s.On("RebuildAddressBookEntrySearch", mock.Anything).Run(func(_ mock.Arguments) {
		cancel()
	}).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
	wait()
}

func Test_handleRebuildAddressBookEntrySearch(t *testing.T) {
	suite.Run(t, new(handleRebuildAddressBookEntrySearchSuite))
}

// handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite tests handleSetAddressBookEntriesWithAutoDeliveryEnabled.
type handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite struct {
	suite.Suite
	s             *StoreMock
	r             *gin.Engine
	tokenOK       auth.Token
	sampleEntries []uuid.UUID
}

func (suite *handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "within",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ManageIntelDeliveryPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleEntries = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
}

func (suite *handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/address-book/entries-with-auto-intel-delivery",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleEntries)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/address-book/entries-with-auto-intel-delivery",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleEntries)),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestInvalidIDs() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/address-book/entries-with-auto-intel-delivery",
		Body:   bytes.NewReader(testutil.MarshalJSONMust([]string{"hello", "world"})),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestSetFail() {
	suite.s.On("SetAddressBookEntriesWithAutoDeliveryEnabled", mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/address-book/entries-with-auto-intel-delivery",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleEntries)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestOK() {
	suite.s.On("SetAddressBookEntriesWithAutoDeliveryEnabled", mock.Anything, suite.sampleEntries).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/address-book/entries-with-auto-intel-delivery",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleEntries)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleSetAddressBookEntriesWithAutoDeliveryEnabled(t *testing.T) {
	suite.Run(t, new(handleSetAddressBookEntriesWithAutoDeliveryEnabledSuite))
}

// intelDeliveryAttemptFiltersFromRequestSuite tests
// intelDeliveryAttemptFiltersFromRequest.
type intelDeliveryAttemptFiltersFromRequestSuite struct {
	suite.Suite
	sampleFilters store.IntelDeliveryAttemptFilters
	qOK           url.Values
}

func (suite *intelDeliveryAttemptFiltersFromRequestSuite) SetupTest() {
	suite.sampleFilters = store.IntelDeliveryAttemptFilters{
		ByOperation: nulls.NewUUID(testutil.NewUUIDV4()),
		ByDelivery:  nulls.NewUUID(testutil.NewUUIDV4()),
		ByChannel:   nulls.NewUUID(testutil.NewUUIDV4()),
		ByActive:    nulls.NewBool(true),
	}
	suite.qOK = map[string][]string{
		"by_operation": {suite.sampleFilters.ByOperation.UUID.String()},
		"by_delivery":  {suite.sampleFilters.ByDelivery.UUID.String()},
		"by_channel":   {suite.sampleFilters.ByChannel.UUID.String()},
		"by_active":    {fmt.Sprintf("%t", suite.sampleFilters.ByActive.Bool)},
	}
}

func (suite *intelDeliveryAttemptFiltersFromRequestSuite) TestInvalidByOperationFilter() {
	suite.qOK["by_operation"] = []string{"abc"}
	_, err := intelDeliveryAttemptFiltersFromRequest(suite.qOK)
	suite.Error(err, "should fail")
}

func (suite *intelDeliveryAttemptFiltersFromRequestSuite) TestInvalidByDeliveryFilter() {
	suite.qOK["by_delivery"] = []string{"abc"}
	_, err := intelDeliveryAttemptFiltersFromRequest(suite.qOK)
	suite.Error(err, "should fail")
}

func (suite *intelDeliveryAttemptFiltersFromRequestSuite) TestInvalidByChannelFilter() {
	suite.qOK["by_channel"] = []string{"abc"}
	_, err := intelDeliveryAttemptFiltersFromRequest(suite.qOK)
	suite.Error(err, "should fail")
}

func (suite *intelDeliveryAttemptFiltersFromRequestSuite) TestInvalidByActiveFilter() {
	suite.qOK["by_active"] = []string{"abc"}
	_, err := intelDeliveryAttemptFiltersFromRequest(suite.qOK)
	suite.Error(err, "should fail")
}

func (suite *intelDeliveryAttemptFiltersFromRequestSuite) TestOK() {
	got, err := intelDeliveryAttemptFiltersFromRequest(suite.qOK)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(suite.sampleFilters, got, "should return correct value")
}

func Test_intelDeliveryAttemptFiltersFromRequest(t *testing.T) {
	suite.Run(t, new(intelDeliveryAttemptFiltersFromRequestSuite))
}

// handleGetIntelDeliveryAttemptsSuite tests handleGetIntelDeliveryAttempts.
type handleGetIntelDeliveryAttemptsSuite struct {
	suite.Suite
	s                           *StoreMock
	r                           *gin.Engine
	tokenOK                     auth.Token
	sampleStoreAttempts         pagination.Paginated[store.IntelDeliveryAttempt]
	samplePublicAttempts        pagination.Paginated[publicIntelDeliveryAttempt]
	sampleStoreFilters          store.IntelDeliveryAttemptFilters
	samplePublicFilters         url.Values
	sampleStorePaginationParams pagination.Params
}

func (suite *handleGetIntelDeliveryAttemptsSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "steel",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.DeliverIntelPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleStoreAttempts = pagination.NewPaginated[store.IntelDeliveryAttempt](pagination.Params{}, []store.IntelDeliveryAttempt{
		{
			ID:        testutil.NewUUIDV4(),
			Delivery:  testutil.NewUUIDV4(),
			Channel:   testutil.NewUUIDV4(),
			CreatedAt: testutil.NewRandomTime(),
			IsActive:  true,
			Status:    store.IntelDeliveryStatusAwaitingAck,
			StatusTS:  testutil.NewRandomTime(),
			Note:      nulls.String{},
		},
		{
			ID:        testutil.NewUUIDV4(),
			Delivery:  testutil.NewUUIDV4(),
			Channel:   testutil.NewUUIDV4(),
			CreatedAt: testutil.NewRandomTime(),
			IsActive:  true,
			Status:    store.IntelDeliveryStatusOpen,
			StatusTS:  testutil.NewRandomTime(),
			Note:      nulls.NewString("ride"),
		},
		{
			ID:        testutil.NewUUIDV4(),
			Delivery:  testutil.NewUUIDV4(),
			Channel:   testutil.NewUUIDV4(),
			CreatedAt: testutil.NewRandomTime(),
			IsActive:  false,
			Status:    store.IntelDeliveryStatusFailed,
			StatusTS:  testutil.NewRandomTime(),
			Note:      nulls.NewString("this failed"),
		},
	}, 52)
	suite.samplePublicAttempts = pagination.MapPaginated(suite.sampleStoreAttempts, func(from store.IntelDeliveryAttempt) publicIntelDeliveryAttempt {
		p, err := publicIntelDeliveryAttemptFromStore(from)
		if err != nil {
			suite.Fail("fail", "map store to public failed")
		}
		return p
	})
	suite.sampleStoreFilters = store.IntelDeliveryAttemptFilters{
		ByOperation: nulls.NewUUID(testutil.NewUUIDV4()),
		ByDelivery:  nulls.NewUUID(testutil.NewUUIDV4()),
		ByChannel:   nulls.NewUUID(testutil.NewUUIDV4()),
		ByActive:    nulls.NewBool(true),
	}
	suite.samplePublicFilters = map[string][]string{
		"by_operation": {suite.sampleStoreFilters.ByOperation.UUID.String()},
		"by_delivery":  {suite.sampleStoreFilters.ByDelivery.UUID.String()},
		"by_channel":   {suite.sampleStoreFilters.ByChannel.UUID.String()},
		"by_active":    {fmt.Sprintf("%t", suite.sampleStoreFilters.ByActive.Bool)},
	}
	suite.sampleStorePaginationParams = pagination.Params{
		Limit:          14,
		Offset:         82,
		OrderBy:        nulls.NewString("label"),
		OrderDirection: pagination.OrderDirDesc,
	}
}

func (suite *handleGetIntelDeliveryAttemptsSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/intel-delivery-attempts",
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/intel-delivery-attempts",
		Token:  token,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsSuite) TestInvalidFilter() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/intel-delivery-attempts?by_operation=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsSuite) TestInvalidPagination() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/intel-delivery-attempts?limit=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsSuite) TestRetrieveFail() {
	suite.s.On("IntelDeliveryAttempts", mock.Anything, mock.Anything, mock.Anything).
		Return(pagination.Paginated[store.IntelDeliveryAttempt]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/intel-delivery-attempts",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsSuite) TestOK() {
	suite.s.On("IntelDeliveryAttempts", mock.Anything, suite.sampleStoreFilters, suite.sampleStorePaginationParams).
		Return(suite.sampleStoreAttempts, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/intel-delivery-attempts?" + suite.samplePublicFilters.Encode() + "&" + pagination.ParamsToQueryString(suite.sampleStorePaginationParams),
		Token:  suite.tokenOK,
	})
	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got pagination.Paginated[publicIntelDeliveryAttempt]
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicAttempts, got, "should return correct body")
}

func Test_handleGetIntelDeliveryAttempts(t *testing.T) {
	suite.Run(t, new(handleGetIntelDeliveryAttemptsSuite))
}

// handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite tests
// handleGetAutoIntelDeliveryEnabledForAddressBookEntry.
type handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite struct {
	suite.Suite
	r         *gin.Engine
	s         *StoreMock
	entryID   uuid.UUID
	isEnabled bool
	tokenOK   auth.Token
}

func (suite *handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite) SetupTest() {
	suite.r = testutil.NewGinEngine()
	suite.s = &StoreMock{}
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.entryID = testutil.NewUUIDV4()
	suite.isEnabled = true
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "thorough",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ManageIntelDeliveryPermissionName}},
		RandomSalt:      nil,
	}
}

func (suite *handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery", suite.entryID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery", suite.entryID.String()),
		Token:  token,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestMissingPermission() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{}

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery", suite.entryID.String()),
		Token:  token,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/address-book/entries/abc/auto-intel-delivery",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestRetrieveFail() {
	suite.s.On("IsAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, mock.Anything).
		Return(false, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery", suite.entryID.String()),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestOKEnabled() {
	suite.s.On("IsAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, suite.entryID).
		Return(suite.isEnabled, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery", suite.entryID.String()),
		Token:  suite.tokenOK,
	})
	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	gotStr, err := io.ReadAll(rr.Body)
	suite.Require().NoError(err, "should return valid body")
	got, err := strconv.ParseBool(string(gotStr))
	suite.Require().NoError(err, "should return bool in body")
	suite.Equal(suite.isEnabled, got, "shoudl return correct bool in body")
}

func (suite *handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestOKDisabled() {
	suite.isEnabled = false
	suite.s.On("IsAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, suite.entryID).
		Return(suite.isEnabled, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery", suite.entryID.String()),
		Token:  suite.tokenOK,
	})
	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	gotStr, err := io.ReadAll(rr.Body)
	suite.Require().NoError(err, "should return valid body")
	got, err := strconv.ParseBool(string(gotStr))
	suite.Require().NoError(err, "should return bool in body")
	suite.Equal(suite.isEnabled, got, "shoudl return correct bool in body")
}

func Test_handleGetAutoIntelDeliveryEnabledForAddressBookEntry(t *testing.T) {
	suite.Run(t, new(handleGetAutoIntelDeliveryEnabledForAddressBookEntrySuite))
}

// handleEnableAutoIntelDeliveryForAddressBookEntrySuite tests
// handleEnableAutoIntelDeliveryForAddressBookEntry.
type handleEnableAutoIntelDeliveryForAddressBookEntrySuite struct {
	suite.Suite
	r       *gin.Engine
	s       *StoreMock
	entryID uuid.UUID
	tokenOK auth.Token
}

func (suite *handleEnableAutoIntelDeliveryForAddressBookEntrySuite) SetupTest() {
	suite.r = testutil.NewGinEngine()
	suite.s = &StoreMock{}
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.entryID = testutil.NewUUIDV4()
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "thorough",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ManageIntelDeliveryPermissionName}},
		RandomSalt:      nil,
	}
}

func (suite *handleEnableAutoIntelDeliveryForAddressBookEntrySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/enable", suite.entryID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleEnableAutoIntelDeliveryForAddressBookEntrySuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/enable", suite.entryID.String()),
		Token:  token,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleEnableAutoIntelDeliveryForAddressBookEntrySuite) TestMissingPermission() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{}

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/enable", suite.entryID.String()),
		Token:  token,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleEnableAutoIntelDeliveryForAddressBookEntrySuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries/abc/auto-intel-delivery/enable",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleEnableAutoIntelDeliveryForAddressBookEntrySuite) TestSetFail() {
	suite.s.On("SetAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/enable", suite.entryID.String()),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleEnableAutoIntelDeliveryForAddressBookEntrySuite) TestOK() {
	suite.s.On("SetAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, suite.entryID, true).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/enable", suite.entryID.String()),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleEnableAutoIntelDeliveryForAddressBookEntry(t *testing.T) {
	suite.Run(t, new(handleEnableAutoIntelDeliveryForAddressBookEntrySuite))
}

// handleDisableAutoIntelDeliveryForAddressBookEntrySuite tests
// handleDisableAutoIntelDeliveryForAddressBookEntry.
type handleDisableAutoIntelDeliveryForAddressBookEntrySuite struct {
	suite.Suite
	r       *gin.Engine
	s       *StoreMock
	entryID uuid.UUID
	tokenOK auth.Token
}

func (suite *handleDisableAutoIntelDeliveryForAddressBookEntrySuite) SetupTest() {
	suite.r = testutil.NewGinEngine()
	suite.s = &StoreMock{}
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.entryID = testutil.NewUUIDV4()
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "thorough",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ManageIntelDeliveryPermissionName}},
		RandomSalt:      nil,
	}
}

func (suite *handleDisableAutoIntelDeliveryForAddressBookEntrySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/disable", suite.entryID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDisableAutoIntelDeliveryForAddressBookEntrySuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/disable", suite.entryID.String()),
		Token:  token,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleDisableAutoIntelDeliveryForAddressBookEntrySuite) TestMissingPermission() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{}

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/disable", suite.entryID.String()),
		Token:  token,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleDisableAutoIntelDeliveryForAddressBookEntrySuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/address-book/entries/abc/auto-intel-delivery/disable",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleDisableAutoIntelDeliveryForAddressBookEntrySuite) TestSetFail() {
	suite.s.On("SetAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/disable", suite.entryID.String()),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDisableAutoIntelDeliveryForAddressBookEntrySuite) TestOK() {
	suite.s.On("SetAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, suite.entryID, false).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/address-book/entries/%s/auto-intel-delivery/disable", suite.entryID.String()),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleDisableAutoIntelDeliveryForAddressBookEntry(t *testing.T) {
	suite.Run(t, new(handleDisableAutoIntelDeliveryForAddressBookEntrySuite))
}
