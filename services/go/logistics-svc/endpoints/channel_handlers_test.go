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
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
)

// storeChannelFromPublicSuite tests storeChannelFromPublic.
type storeChannelFromPublicSuite struct {
	suite.Suite
	samplePublicChannel publicChannel
	sampleStoreChannel  store.Channel
}

func (suite *storeChannelFromPublicSuite) SetupTest() {
	details := publicPushChannelDetails{}
	suite.samplePublicChannel = publicChannel{
		ID:            testutil.NewUUIDV4(),
		Entry:         testutil.NewUUIDV4(),
		Label:         "flavor",
		Type:          string(store.ChannelTypePush),
		Priority:      900,
		MinImportance: 210,
		Details:       testutil.MarshalJSONMust(details),
		Timeout:       851,
	}
	suite.sampleStoreChannel = store.Channel{
		ID:            suite.samplePublicChannel.ID,
		Entry:         suite.samplePublicChannel.Entry,
		Label:         suite.samplePublicChannel.Label,
		Type:          store.ChannelType(suite.samplePublicChannel.Type),
		Priority:      suite.samplePublicChannel.Priority,
		MinImportance: suite.samplePublicChannel.MinImportance,
		Details:       storePushChannelDetailsFromPublic(details),
		Timeout:       suite.samplePublicChannel.Timeout,
	}
}

func (suite *storeChannelFromPublicSuite) TestUnsupportedChannelType() {
	pChan := suite.samplePublicChannel
	pChan.Type = testutil.NewUUIDV4().String()
	_, err := storeChannelFromPublic(pChan)
	suite.Error(err, "should fail")
}

func (suite *storeChannelFromPublicSuite) TestDetailsConversionFail() {
	pChan := suite.samplePublicChannel
	pChan.Details = []byte(`{invalid`)
	_, err := storeChannelFromPublic(pChan)
	suite.Error(err, "should fail")
}

func (suite *storeChannelFromPublicSuite) TestOK() {
	s, err := storeChannelFromPublic(suite.samplePublicChannel)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(suite.sampleStoreChannel, s, "should return correct value")
}

func Test_storeChannelFromPublic(t *testing.T) {
	suite.Run(t, new(storeChannelFromPublicSuite))
}

func testStoreChannelDetailsFromPublic[From any, To store.ChannelDetails](t *testing.T, chanType store.ChannelType, from From, to To) {
	fromRaw, err := json.Marshal(from)
	require.NoError(t, err, "marshal public channel details should not fail")
	pChan := publicChannel{
		ID:            testutil.NewUUIDV4(),
		Entry:         testutil.NewUUIDV4(),
		Label:         "flavor",
		Type:          string(chanType),
		Priority:      900,
		MinImportance: 210,
		Details:       fromRaw,
		Timeout:       851,
	}
	sChan, err := storeChannelFromPublic(pChan)
	require.NoError(t, err, "store channel conversion from public should not fail")
	assert.Equal(t, store.Channel{
		ID:            pChan.ID,
		Entry:         pChan.Entry,
		Label:         pChan.Label,
		Type:          chanType,
		Priority:      pChan.Priority,
		MinImportance: pChan.MinImportance,
		Details:       to,
		Timeout:       pChan.Timeout,
	}, sChan, "conversion should return correct value")
}

func Test_storeDirectChannelDetailsFromPublic(t *testing.T) {
	p := publicDirectChannelDetails{
		Info: "board",
	}
	s := store.DirectChannelDetails{
		Info: p.Info,
	}
	testStoreChannelDetailsFromPublic(t, store.ChannelTypeDirect, p, s)
}

func Test_storeEmailChannelDetailsFromPublic(t *testing.T) {
	p := publicEmailChannelDetails{
		Email: "meow@meow.com",
	}
	s := store.EmailChannelDetails{
		Email: p.Email,
	}
	testStoreChannelDetailsFromPublic(t, store.ChannelTypeEmail, p, s)
}

func Test_storeForwardToGroupChannelDetailsFromPublic(t *testing.T) {
	p := publicForwardToGroupChannelDetails{
		ForwardToGroup: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	s := store.ForwardToGroupChannelDetails{
		ForwardToGroup: p.ForwardToGroup,
	}
	testStoreChannelDetailsFromPublic(t, store.ChannelTypeForwardToGroup, p, s)
}

func Test_storeForwardToUserChannelDetailsFromPublic(t *testing.T) {
	p := publicForwardToUserChannelDetails{
		ForwardToUser: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	s := store.ForwardToUserChannelDetails{
		ForwardToUser: p.ForwardToUser,
	}
	testStoreChannelDetailsFromPublic(t, store.ChannelTypeForwardToUser, p, s)
}

func Test_storePhoneCallChannelDetailsFromPublic(t *testing.T) {
	p := publicPhoneCallChannelDetails{
		Phone: "00123456789",
	}
	s := store.PhoneCallChannelDetails{
		Phone: p.Phone,
	}
	testStoreChannelDetailsFromPublic(t, store.ChannelTypePhoneCall, p, s)
}

func Test_storePushChannelDetailsFromPublic(t *testing.T) {
	p := publicPushChannelDetails{}
	s := store.PushChannelDetails{}
	testStoreChannelDetailsFromPublic(t, store.ChannelTypePush, p, s)
}

func Test_storeRadioChannelDetailsFromPublic(t *testing.T) {
	p := publicRadioChannelDetails{
		Info: "board",
	}
	s := store.RadioChannelDetails{
		Info: p.Info,
	}
	testStoreChannelDetailsFromPublic(t, store.ChannelTypeRadio, p, s)
}

// publicChannelFromStoreSuite test publicChannelFromStore.
type publicChannelFromStoreSuite struct {
	suite.Suite
	samplePublicChannel publicChannel
	sampleStoreChannel  store.Channel
}

func (suite *publicChannelFromStoreSuite) SetupTest() {
	details := store.PushChannelDetails{}
	suite.sampleStoreChannel = store.Channel{
		ID:            testutil.NewUUIDV4(),
		Entry:         testutil.NewUUIDV4(),
		Label:         "flavor",
		Type:          store.ChannelTypePush,
		Priority:      900,
		MinImportance: 210,
		Details:       details,
		Timeout:       851,
	}
	suite.samplePublicChannel = publicChannel{
		ID:            suite.sampleStoreChannel.ID,
		Entry:         suite.sampleStoreChannel.Entry,
		Label:         suite.sampleStoreChannel.Label,
		Type:          string(suite.sampleStoreChannel.Type),
		Priority:      suite.sampleStoreChannel.Priority,
		MinImportance: suite.sampleStoreChannel.MinImportance,
		Details:       testutil.MarshalJSONMust(suite.sampleStoreChannel.Details),
		Timeout:       suite.sampleStoreChannel.Timeout,
	}
}

// ValidatableMock mocks entityvalidation.Validatable.
type ValidatableMock struct {
	mock.Mock
}

func (m *ValidatableMock) Validate() (entityvalidation.Report, error) {
	args := m.Called()
	return args.Get(0).(entityvalidation.Report), args.Error(1)
}

func (suite *publicChannelFromStoreSuite) TestUnsupportedChannelType() {
	sChan := suite.sampleStoreChannel
	sChan.Details = &ValidatableMock{}
	_, err := publicChannelFromStore(sChan)
	suite.Error(err, "should fail")
}

func (suite *publicChannelFromStoreSuite) TestOK() {
	p, err := publicChannelFromStore(suite.sampleStoreChannel)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(suite.samplePublicChannel, p, "should return correct value")
}

func Test_publicChannelFromStore(t *testing.T) {
	suite.Run(t, new(publicChannelFromStoreSuite))
}

func testPublicChannelDetailsFromStore[From store.ChannelDetails, To any](t *testing.T, chanType store.ChannelType, from From, to To) {
	toRaw, err := json.Marshal(to)
	require.NoError(t, err, "marshal public channel details should not fail")
	sChan := store.Channel{
		ID:            testutil.NewUUIDV4(),
		Entry:         testutil.NewUUIDV4(),
		Label:         "flavor",
		Type:          chanType,
		Priority:      900,
		MinImportance: 210,
		Details:       from,
		Timeout:       851,
	}
	pChan, err := publicChannelFromStore(sChan)
	require.NoError(t, err, "store channel conversion from public should not fail")
	assert.Equal(t, publicChannel{
		ID:            sChan.ID,
		Entry:         sChan.Entry,
		Label:         sChan.Label,
		Type:          string(chanType),
		Priority:      sChan.Priority,
		MinImportance: sChan.MinImportance,
		Details:       toRaw,
		Timeout:       sChan.Timeout,
	}, pChan, "conversion should return correct value")
}

func Test_publicDirectChannelDetailsFromStore(t *testing.T) {
	p := store.DirectChannelDetails{
		Info: "board",
	}
	s := publicDirectChannelDetails{
		Info: p.Info,
	}
	testPublicChannelDetailsFromStore(t, store.ChannelTypeDirect, p, s)
}

func Test_publicEmailChannelDetailsFromStore(t *testing.T) {
	p := store.EmailChannelDetails{
		Email: "meow@meow.com",
	}
	s := publicEmailChannelDetails{
		Email: p.Email,
	}
	testPublicChannelDetailsFromStore(t, store.ChannelTypeEmail, p, s)
}

func Test_publicForwardToGroupChannelDetailsFromStore(t *testing.T) {
	p := store.ForwardToGroupChannelDetails{
		ForwardToGroup: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	s := publicForwardToGroupChannelDetails{
		ForwardToGroup: p.ForwardToGroup,
	}
	testPublicChannelDetailsFromStore(t, store.ChannelTypeForwardToGroup, p, s)
}

func Test_publicForwardToUserChannelDetailsFromStore(t *testing.T) {
	p := store.ForwardToUserChannelDetails{
		ForwardToUser: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	s := publicForwardToUserChannelDetails{
		ForwardToUser: p.ForwardToUser,
	}
	testPublicChannelDetailsFromStore(t, store.ChannelTypeForwardToUser, p, s)
}

func Test_publicPhoneCallChannelDetailsFromStore(t *testing.T) {
	p := store.PhoneCallChannelDetails{
		Phone: "00123456789",
	}
	s := publicPhoneCallChannelDetails{
		Phone: p.Phone,
	}
	testPublicChannelDetailsFromStore(t, store.ChannelTypePhoneCall, p, s)
}

func Test_publicPushChannelDetailsFromStore(t *testing.T) {
	p := store.PushChannelDetails{}
	s := publicPushChannelDetails{}
	testPublicChannelDetailsFromStore(t, store.ChannelTypePush, p, s)
}

func Test_publicRadioChannelDetailsFromStore(t *testing.T) {
	p := store.RadioChannelDetails{
		Info: "board",
	}
	s := publicRadioChannelDetails{
		Info: p.Info,
	}
	testPublicChannelDetailsFromStore(t, store.ChannelTypeRadio, p, s)
}

// handleUpdateChannelsByAddressBookEntrySuite tests
// handleUpdateChannelsByAddressBookEntry.
type handleUpdateChannelsByAddressBookEntrySuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	tokenOK            auth.Token
	sampleEntryID      uuid.UUID
	samplePublicUpdate []publicChannel
	sampleStoreUpdate  []store.Channel
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "motor",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.UpdateAnyAddressBookEntryPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleEntryID = testutil.NewUUIDV4()
	suite.samplePublicUpdate = []publicChannel{
		{
			ID:            testutil.NewUUIDV4(),
			Entry:         suite.sampleEntryID,
			Label:         "quarter",
			Type:          string(store.ChannelTypePush),
			Priority:      966,
			MinImportance: 292,
			Details:       testutil.MarshalJSONMust(publicPushChannelDetails{}),
			Timeout:       386,
		},
		{
			Entry:         suite.sampleEntryID,
			Label:         "towel",
			Type:          string(store.ChannelTypePhoneCall),
			Priority:      40,
			MinImportance: 832,
			Details: testutil.MarshalJSONMust(publicPhoneCallChannelDetails{
				Phone: "0123456789",
			}),
			Timeout: 542,
		},
	}
	suite.sampleStoreUpdate = make([]store.Channel, 0, len(suite.samplePublicUpdate))
	for _, pChan := range suite.samplePublicUpdate {
		sChan, err := storeChannelFromPublic(pChan)
		if err != nil {
			panic(err)
		}
		suite.sampleStoreUpdate = append(suite.sampleStoreUpdate, sChan)
	}
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s/channels", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "woof",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s/channels", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/address-book/entries/abc/channels",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s/channels", suite.sampleEntryID.String()),
		Body:   strings.NewReader(`{invalid`),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) TestConversionToStoreFromPublicFail() {
	suite.samplePublicUpdate[0].Type = testutil.NewUUIDV4().String()

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s/channels", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) TestUpdateFail() {
	suite.s.On("UpdateChannelsByAddressBookEntry", mock.Anything, suite.sampleEntryID, suite.sampleStoreUpdate, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s/channels", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) TestOKWithUpdateAnyPermission() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{{Name: permission.UpdateAnyAddressBookEntryPermissionName}}
	suite.s.On("UpdateChannelsByAddressBookEntry", mock.Anything, suite.sampleEntryID, suite.sampleStoreUpdate, uuid.NullUUID{}).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s/channels", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleUpdateChannelsByAddressBookEntrySuite) TestOKWithoutUpdateAnyPermission() {
	token := suite.tokenOK
	token.Permissions = nil
	suite.s.On("UpdateChannelsByAddressBookEntry", mock.Anything, suite.sampleEntryID, suite.sampleStoreUpdate, nulls.NewUUID(token.UserID)).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/address-book/entries/%s/channels", suite.sampleEntryID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  token,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleUpdateChannelsByAddressBookEntry(t *testing.T) {
	suite.Run(t, new(handleUpdateChannelsByAddressBookEntrySuite))
}
