package eventport

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// PortNotifyAddressBookEntryCreatedSuite tests Port.NotifyAddressBookEntryCreated.
type PortNotifyAddressBookEntryCreatedSuite struct {
	suite.Suite
	port                   *PortMock
	sampleAddressBookEntry store.AddressBookEntry
	expectedMessage        kafkautil.OutboundMessage
}

func (suite *PortNotifyAddressBookEntryCreatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleAddressBookEntry = store.AddressBookEntry{
		ID:          testutil.NewUUIDV4(),
		Label:       "bag",
		Description: "proof",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.AddressBookTopic,
		Key:       suite.sampleAddressBookEntry.ID.String(),
		EventType: event.TypeAddressBookEntryCreated,
		Value: event.AddressBookEntryCreated{
			ID:          suite.sampleAddressBookEntry.ID,
			Label:       suite.sampleAddressBookEntry.Label,
			Description: suite.sampleAddressBookEntry.Description,
			Operation:   suite.sampleAddressBookEntry.Operation,
			User:        suite.sampleAddressBookEntry.User,
		},
	}
}

func (suite *PortNotifyAddressBookEntryCreatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryCreated(timeout, &testutil.DBTx{}, suite.sampleAddressBookEntry)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyAddressBookEntryCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryCreated(timeout, &testutil.DBTx{}, suite.sampleAddressBookEntry)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyAddressBookEntryCreated(t *testing.T) {
	suite.Run(t, new(PortNotifyAddressBookEntryCreatedSuite))
}

// PortNotifyAddressBookEntryUpdatedSuite tests Port.NotifyAddressBookEntryUpdated.
type PortNotifyAddressBookEntryUpdatedSuite struct {
	suite.Suite
	port                   *PortMock
	sampleAddressBookEntry store.AddressBookEntry
	expectedMessage        kafkautil.OutboundMessage
}

func (suite *PortNotifyAddressBookEntryUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleAddressBookEntry = store.AddressBookEntry{
		ID:          testutil.NewUUIDV4(),
		Label:       "bag",
		Description: "proof",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.AddressBookTopic,
		Key:       suite.sampleAddressBookEntry.ID.String(),
		EventType: event.TypeAddressBookEntryUpdated,
		Value: event.AddressBookEntryUpdated{
			ID:          suite.sampleAddressBookEntry.ID,
			Label:       suite.sampleAddressBookEntry.Label,
			Description: suite.sampleAddressBookEntry.Description,
			Operation:   suite.sampleAddressBookEntry.Operation,
			User:        suite.sampleAddressBookEntry.User,
		},
	}
}

func (suite *PortNotifyAddressBookEntryUpdatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryUpdated(timeout, &testutil.DBTx{}, suite.sampleAddressBookEntry)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyAddressBookEntryUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryUpdated(timeout, &testutil.DBTx{}, suite.sampleAddressBookEntry)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyAddressBookEntryUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyAddressBookEntryUpdatedSuite))
}

// PortNotifyAddressBookEntryDeletedSuite tests
// Port.NotifyAddressBookEntryDeleted.
type PortNotifyAddressBookEntryDeletedSuite struct {
	suite.Suite
	port                     *PortMock
	sampleAddressBookEntryID uuid.UUID
	expectedMessage          kafkautil.OutboundMessage
}

func (suite *PortNotifyAddressBookEntryDeletedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleAddressBookEntryID = testutil.NewUUIDV4()
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.AddressBookTopic,
		Key:       suite.sampleAddressBookEntryID.String(),
		EventType: event.TypeAddressBookEntryDeleted,
		Value: event.AddressBookEntryDeleted{
			ID: suite.sampleAddressBookEntryID,
		},
	}
}

func (suite *PortNotifyAddressBookEntryDeletedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryDeleted(timeout, &testutil.DBTx{}, suite.sampleAddressBookEntryID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyAddressBookEntryDeletedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryDeleted(timeout, &testutil.DBTx{}, suite.sampleAddressBookEntryID)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyAddressBookEntryDeleted(t *testing.T) {
	suite.Run(t, new(PortNotifyAddressBookEntryDeletedSuite))
}

// mapDirectChannelDetailsSuite tests mapDirectChannelDetails.
type mapDirectChannelDetailsSuite struct {
	suite.Suite
	sampleDetails store.DirectChannelDetails
	mapper        channelDetailsMapper
}

func (suite *mapDirectChannelDetailsSuite) SetupTest() {
	suite.sampleDetails = store.DirectChannelDetails{
		Info: "article",
	}
	var err error
	suite.mapper, err = mapChannelDetails(store.ChannelTypeDirect)
	if err != nil {
		suite.FailNow("get channel-details-mapper failed")
	}
}

func (suite *mapDirectChannelDetailsSuite) TestInvalidDetails() {
	_, err := suite.mapper(store.EmailChannelDetails{})
	suite.Error(err, "should fail")
}

func (suite *mapDirectChannelDetailsSuite) TestOK() {
	raw, err := suite.mapper(suite.sampleDetails)
	suite.Require().NoError(err, "should not fail")
	var got event.AddressBookEntryDirectChannelDetails
	suite.Require().NoError(json.Unmarshal(raw, &got), "should return valid details")
	suite.Equal(event.AddressBookEntryDirectChannelDetails{
		Info: suite.sampleDetails.Info,
	}, got, "should return correct details")
}

func Test_mapDirectChannelDetails(t *testing.T) {
	suite.Run(t, new(mapDirectChannelDetailsSuite))
}

// mapEmailChannelDetailsSuite tests mapEmailChannelDetails.
type mapEmailChannelDetailsSuite struct {
	suite.Suite
	sampleDetails store.EmailChannelDetails
	mapper        channelDetailsMapper
}

func (suite *mapEmailChannelDetailsSuite) SetupTest() {
	suite.sampleDetails = store.EmailChannelDetails{
		Email: "meow@meow.com",
	}
	var err error
	suite.mapper, err = mapChannelDetails(store.ChannelTypeEmail)
	if err != nil {
		suite.FailNow("get channel-details-mapper failed")
	}
}

func (suite *mapEmailChannelDetailsSuite) TestInvalidDetails() {
	_, err := suite.mapper(store.RadioChannelDetails{})
	suite.Error(err, "should fail")
}

func (suite *mapEmailChannelDetailsSuite) TestOK() {
	raw, err := suite.mapper(suite.sampleDetails)
	suite.Require().NoError(err, "should not fail")
	var got event.AddressBookEntryEmailChannelDetails
	suite.Require().NoError(json.Unmarshal(raw, &got), "should return valid details")
	suite.Equal(event.AddressBookEntryEmailChannelDetails{
		Email: suite.sampleDetails.Email,
	}, got, "should return correct details")
}

func Test_mapEmailChannelDetails(t *testing.T) {
	suite.Run(t, new(mapEmailChannelDetailsSuite))
}

// mapForwardToGroupChannelDetailsSuite tests mapForwardToGroupChannelDetails.
type mapForwardToGroupChannelDetailsSuite struct {
	suite.Suite
	sampleDetails store.ForwardToGroupChannelDetails
	mapper        channelDetailsMapper
}

func (suite *mapForwardToGroupChannelDetailsSuite) SetupTest() {
	suite.sampleDetails = store.ForwardToGroupChannelDetails{
		ForwardToGroup: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	var err error
	suite.mapper, err = mapChannelDetails(store.ChannelTypeForwardToGroup)
	if err != nil {
		suite.FailNow("get channel-details-mapper failed")
	}
}

func (suite *mapForwardToGroupChannelDetailsSuite) TestInvalidDetails() {
	_, err := suite.mapper(store.EmailChannelDetails{})
	suite.Error(err, "should fail")
}

func (suite *mapForwardToGroupChannelDetailsSuite) TestOK() {
	raw, err := suite.mapper(suite.sampleDetails)
	suite.Require().NoError(err, "should not fail")
	var got event.AddressBookEntryForwardToGroupChannelDetails
	suite.Require().NoError(json.Unmarshal(raw, &got), "should return valid details")
	suite.Equal(event.AddressBookEntryForwardToGroupChannelDetails{
		ForwardToGroup: suite.sampleDetails.ForwardToGroup,
	}, got, "should return correct details")
}

func Test_mapForwardToGroupChannelDetails(t *testing.T) {
	suite.Run(t, new(mapForwardToGroupChannelDetailsSuite))
}

// mapForwardToUserChannelDetailsSuite tests mapForwardToUserChannelDetails.
type mapForwardToUserChannelDetailsSuite struct {
	suite.Suite
	sampleDetails store.ForwardToUserChannelDetails
	mapper        channelDetailsMapper
}

func (suite *mapForwardToUserChannelDetailsSuite) SetupTest() {
	suite.sampleDetails = store.ForwardToUserChannelDetails{
		ForwardToUser: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	var err error
	suite.mapper, err = mapChannelDetails(store.ChannelTypeForwardToUser)
	if err != nil {
		suite.FailNow("get channel-details-mapper failed")
	}
}

func (suite *mapForwardToUserChannelDetailsSuite) TestInvalidDetails() {
	_, err := suite.mapper(store.EmailChannelDetails{})
	suite.Error(err, "should fail")
}

func (suite *mapForwardToUserChannelDetailsSuite) TestOK() {
	raw, err := suite.mapper(suite.sampleDetails)
	suite.Require().NoError(err, "should not fail")
	var got event.AddressBookEntryForwardToUserChannelDetails
	suite.Require().NoError(json.Unmarshal(raw, &got), "should return valid details")
	suite.Equal(event.AddressBookEntryForwardToUserChannelDetails{
		ForwardToUser: suite.sampleDetails.ForwardToUser,
	}, got, "should return correct details")
}

func Test_mapForwardToUserChannelDetails(t *testing.T) {
	suite.Run(t, new(mapForwardToUserChannelDetailsSuite))
}

// mapPhoneCallChannelDetailsSuite tests mapPhoneCallChannelDetails.
type mapPhoneCallChannelDetailsSuite struct {
	suite.Suite
	sampleDetails store.PhoneCallChannelDetails
	mapper        channelDetailsMapper
}

func (suite *mapPhoneCallChannelDetailsSuite) SetupTest() {
	suite.sampleDetails = store.PhoneCallChannelDetails{
		Phone: "00491523371522",
	}
	var err error
	suite.mapper, err = mapChannelDetails(store.ChannelTypePhoneCall)
	if err != nil {
		suite.FailNow("get channel-details-mapper failed")
	}
}

func (suite *mapPhoneCallChannelDetailsSuite) TestInvalidDetails() {
	_, err := suite.mapper(store.EmailChannelDetails{})
	suite.Error(err, "should fail")
}

func (suite *mapPhoneCallChannelDetailsSuite) TestOK() {
	raw, err := suite.mapper(suite.sampleDetails)
	suite.Require().NoError(err, "should not fail")
	var got event.AddressBookEntryPhoneCallChannelDetails
	suite.Require().NoError(json.Unmarshal(raw, &got), "should return valid details")
	suite.Equal(event.AddressBookEntryPhoneCallChannelDetails{
		Phone: suite.sampleDetails.Phone,
	}, got, "should return correct details")
}

func Test_mapPhoneCallChannelDetails(t *testing.T) {
	suite.Run(t, new(mapPhoneCallChannelDetailsSuite))
}

// mapRadioChannelDetailsSuite tests mapRadioChannelDetails.
type mapRadioChannelDetailsSuite struct {
	suite.Suite
	sampleDetails store.RadioChannelDetails
	mapper        channelDetailsMapper
}

func (suite *mapRadioChannelDetailsSuite) SetupTest() {
	suite.sampleDetails = store.RadioChannelDetails{
		Info: "article",
	}
	var err error
	suite.mapper, err = mapChannelDetails(store.ChannelTypeRadio)
	if err != nil {
		suite.FailNow("get channel-details-mapper failed")
	}
}

func (suite *mapRadioChannelDetailsSuite) TestInvalidDetails() {
	_, err := suite.mapper(store.EmailChannelDetails{})
	suite.Error(err, "should fail")
}

func (suite *mapRadioChannelDetailsSuite) TestOK() {
	raw, err := suite.mapper(suite.sampleDetails)
	suite.Require().NoError(err, "should not fail")
	var got event.AddressBookEntryRadioChannelDetails
	suite.Require().NoError(json.Unmarshal(raw, &got), "should return valid details")
	suite.Equal(event.AddressBookEntryRadioChannelDetails{
		Info: suite.sampleDetails.Info,
	}, got, "should return correct details")
}

func Test_mapRadioChannelDetails(t *testing.T) {
	suite.Run(t, new(mapRadioChannelDetailsSuite))
}

// PortNotifyAddressBookEntryChannelsUpdatedSuite tests
// Port.NotifyAddressBookEntryChannelsUpdated.
type PortNotifyAddressBookEntryChannelsUpdatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleEntryID   uuid.UUID
	sampleChannels  []store.Channel
	expectedMessage kafkautil.OutboundMessage
}

func (suite *PortNotifyAddressBookEntryChannelsUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleEntryID = testutil.NewUUIDV4()
	suite.sampleChannels = []store.Channel{
		{
			ID:            testutil.NewUUIDV4(),
			Entry:         suite.sampleEntryID,
			Label:         "",
			Type:          store.ChannelTypeDirect,
			Priority:      1,
			MinImportance: 2,
			Details: store.DirectChannelDetails{
				Info: "clean",
			},
			Timeout: 10 * time.Millisecond,
		},
		{
			ID:            testutil.NewUUIDV4(),
			Entry:         suite.sampleEntryID,
			Label:         "give",
			Type:          store.ChannelTypeForwardToUser,
			Priority:      -1,
			MinImportance: 3,
			Details: store.ForwardToUserChannelDetails{
				ForwardToUser: []uuid.UUID{
					testutil.NewUUIDV4(),
					testutil.NewUUIDV4(),
					testutil.NewUUIDV4(),
				},
			},
			Timeout: 11 * time.Minute,
		},
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.AddressBookTopic,
		Key:       suite.sampleEntryID.String(),
		EventType: event.TypeAddressBookEntryChannelsUpdated,
		Value: event.AddressBookEntryChannelsUpdated{
			Entry: suite.sampleEntryID,
			Channels: []event.AddressBookEntryChannelsUpdatedChannel{
				{
					ID:            suite.sampleChannels[0].ID,
					Entry:         suite.sampleEntryID,
					Label:         "",
					Type:          event.AddressBookEntryChannelTypeDirect,
					Priority:      1,
					MinImportance: 2,
					Details: testutil.MarshalJSONMust(event.AddressBookEntryDirectChannelDetails{
						Info: "clean",
					}),
					Timeout: 10 * time.Millisecond,
				},
				{
					ID:            suite.sampleChannels[1].ID,
					Entry:         suite.sampleEntryID,
					Label:         "give",
					Type:          event.AddressBookEntryChannelTypeForwardToUser,
					Priority:      -1,
					MinImportance: 3,
					Details: testutil.MarshalJSONMust(event.AddressBookEntryForwardToUserChannelDetails{
						ForwardToUser: suite.sampleChannels[1].Details.(store.ForwardToUserChannelDetails).ForwardToUser,
					}),
					Timeout: 11 * time.Minute,
				},
			},
		},
	}
}

func (suite *PortNotifyAddressBookEntryChannelsUpdatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryChannelsUpdated(timeout, &testutil.DBTx{}, suite.sampleEntryID, suite.sampleChannels)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyAddressBookEntryChannelsUpdatedSuite) TestUnsupportedChannelType() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryChannelsUpdated(timeout, &testutil.DBTx{}, suite.sampleEntryID, []store.Channel{
			{
				Type: "wf34",
			},
		})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyAddressBookEntryChannelsUpdatedSuite) TestChannelDetailsTypeMismatch() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryChannelsUpdated(timeout, &testutil.DBTx{}, suite.sampleEntryID, []store.Channel{
			{
				Type:    store.ChannelTypePhoneCall,
				Details: store.EmailChannelDetails{},
			},
		})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyAddressBookEntryChannelsUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyAddressBookEntryChannelsUpdated(timeout, &testutil.DBTx{}, suite.sampleEntryID, suite.sampleChannels)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyAddressBookEntryChannelsUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyAddressBookEntryChannelsUpdatedSuite))
}

// assureChannelTypesSupportedSuite tests assureChannelTypesSupported.
type assureChannelTypesSupportedSuite struct {
	suite.Suite
}

func (suite *assureChannelTypesSupportedSuite) TestUnsupportedType() {
	suite.Panics(func() {
		store.ChannelTypeSupplier.ChannelTypes["v2a26"] = struct{}{}
		assureChannelTypesSupported()
	}, "should panic")
}

func (suite *assureChannelTypesSupportedSuite) TestOK() {
	suite.NotPanics(func() {
		assureChannelTypesSupported()
	}, "should not panic")
}

func Test_assureChannelTypesSupported(t *testing.T) {
	suite.Run(t, new(assureChannelTypesSupportedSuite))
}
