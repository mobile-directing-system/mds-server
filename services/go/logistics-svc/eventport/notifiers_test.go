package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"
)

func Test_mapChannelType(t *testing.T) {
	testutil.TestMapperWithConstExtractionFromDir(t, mapChannelType, "../store", nulls.String{})
}

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

// mapInAppNotificationChannelDetailsSuite tests
// mapInAppNotificationChannelDetails.
type mapInAppNotificationChannelDetailsSuite struct {
	suite.Suite
	sampleDetails store.InAppNotificationChannelDetails
	mapper        channelDetailsMapper
}

func (suite *mapInAppNotificationChannelDetailsSuite) SetupTest() {
	suite.sampleDetails = store.InAppNotificationChannelDetails{}
	var err error
	suite.mapper, err = mapChannelDetails(store.ChannelTypeInAppNotification)
	if err != nil {
		suite.FailNow("get channel-details-mapper failed")
	}
}

func (suite *mapInAppNotificationChannelDetailsSuite) TestInvalidDetails() {
	_, err := suite.mapper(store.EmailChannelDetails{})
	suite.Error(err, "should fail")
}

func (suite *mapInAppNotificationChannelDetailsSuite) TestOK() {
	raw, err := suite.mapper(suite.sampleDetails)
	suite.Require().NoError(err, "should not fail")
	var got event.AddressBookEntryInAppNotificationChannelDetails
	suite.Require().NoError(json.Unmarshal(raw, &got), "should return valid details")
	suite.Equal(event.AddressBookEntryInAppNotificationChannelDetails{}, got, "should return correct details")
}

func Test_mapInAppNotificationChannelDetails(t *testing.T) {
	suite.Run(t, new(mapInAppNotificationChannelDetailsSuite))
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
			IsActive:      true,
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
			IsActive:      false,
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
					IsActive:      suite.sampleChannels[0].IsActive,
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
					IsActive:      suite.sampleChannels[1].IsActive,
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

func Test_eventIntelDeliveryStatusFromStore(t *testing.T) {
	testutil.TestMapperWithConstExtractionFromDir(t, eventIntelDeliveryStatusFromStore, "../../shared/event", nulls.String{})
}

func Test_mapIntelTypeFromStore(t *testing.T) {
	testutil.TestMapperWithConstExtraction(t, mapIntelTypeFromStore, "../store/intel_content.go", nulls.String{})
}

// PortNotifyIntelCreatedSuite tests Port.NotifyIntelCreated.
type PortNotifyIntelCreatedSuite struct {
	suite.Suite
	port             *PortMock
	tx               *testutil.DBTx
	sampleCreated    store.Intel
	expectedMessages []kafkautil.OutboundMessage
}

func (suite *PortNotifyIntelCreatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.tx = &testutil.DBTx{}
	suite.sampleCreated = store.Intel{
		ID:        testutil.NewUUIDV4(),
		CreatedBy: testutil.NewUUIDV4(),
		Operation: testutil.NewUUIDV4(),
		Type:      store.IntelTypePlaintextMessage,
		Content: testutil.MarshalJSONMust(store.IntelTypePlaintextMessageContent{
			Text: "Hello World!",
		}),
		SearchText: nulls.NewString("gold"),
	}
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.IntelTopic,
			Key:       suite.sampleCreated.ID.String(),
			EventType: event.TypeIntelCreated,
			Value: event.IntelCreated{
				ID:        suite.sampleCreated.ID,
				CreatedAt: suite.sampleCreated.CreatedAt,
				CreatedBy: suite.sampleCreated.CreatedBy,
				Operation: suite.sampleCreated.Operation,
				Type:      event.IntelTypePlaintextMessage,
				Content: testutil.MarshalJSONMust(event.IntelTypePlaintextMessageContent{
					Text: "Hello World!",
				}),
				SearchText: suite.sampleCreated.SearchText,
				IsValid:    suite.sampleCreated.IsValid,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyIntelCreatedSuite) TestUnsupportedIntelType() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleCreated.Type = store.IntelType(testutil.NewUUIDV4().String())
	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelCreated(timeout, suite.tx, suite.sampleCreated)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelCreatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelCreated(timeout, suite.tx, suite.sampleCreated)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelCreated(timeout, suite.tx, suite.sampleCreated)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyIntelCreated(t *testing.T) {
	suite.Run(t, new(PortNotifyIntelCreatedSuite))
}

// PortNotifyIntelInvalidatedSuite tests Port.NotifyIntelInvalidated.
type PortNotifyIntelInvalidatedSuite struct {
	suite.Suite
	port             *PortMock
	tx               *testutil.DBTx
	sampleID         uuid.UUID
	sampleBy         uuid.UUID
	expectedMessages []kafkautil.OutboundMessage
}

func (suite *PortNotifyIntelInvalidatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.tx = &testutil.DBTx{}
	suite.sampleID = testutil.NewUUIDV4()
	suite.sampleBy = testutil.NewUUIDV4()
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.IntelTopic,
			Key:       suite.sampleID.String(),
			EventType: event.TypeIntelInvalidated,
			Value: event.IntelInvalidated{
				ID: suite.sampleID,
				By: suite.sampleBy,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyIntelInvalidatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelInvalidated(timeout, suite.tx, suite.sampleID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelInvalidatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelInvalidated(timeout, suite.tx, suite.sampleID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyIntelInvalidated(t *testing.T) {
	suite.Run(t, new(PortNotifyIntelInvalidatedSuite))
}

func Test_mapIntelContentFromStore(t *testing.T) {
	testutil.TestMapperWithConstExtraction(t, func(from store.IntelType) (string, error) {
		// Assure that the type is known.
		_, err := mapIntelContentFromStore(from, json.RawMessage(`{}`))
		if err != nil {
			if strings.Contains(err.Error(), "no intel-content-mapper") {
				return "", err
			}
			if strings.Contains(err.Error(), "mapper fn") {
				return "", nil
			}
		}
		return "", nil
	}, "../store/intel_content.go", nulls.String{})
}

func Test_mapIntelTypeAnalogRadioMessageContent(t *testing.T) {
	s := store.IntelTypeAnalogRadioMessageContent{
		Channel:  "except",
		Callsign: "passage",
		Head:     "redden",
		Content:  "hope",
	}
	e, err := mapIntelTypeAnalogRadioMessageContent(s)
	require.NoError(t, err, "should not fail")
	assert.Equal(t, event.IntelTypeAnalogRadioMessageContent{
		Channel:  s.Channel,
		Callsign: s.Callsign,
		Head:     s.Head,
		Content:  s.Content,
	}, e, "should return correct value")
}

func Test_mapIntelTypePlaintextMessageContent(t *testing.T) {
	s := store.IntelTypePlaintextMessageContent{
		Text: "learn",
	}
	e, err := mapIntelTypePlaintextMessageContent(s)
	require.NoError(t, err, "should not fail")
	assert.Equal(t, event.IntelTypePlaintextMessageContent{
		Text: s.Text,
	}, e, "should return correct value")
}

// PortNotifyIntelDeliveryAttemptCreatedSuite tests
// Port.NotifyIntelDeliveryAttemptCreated.
type PortNotifyIntelDeliveryAttemptCreatedSuite struct {
	suite.Suite
	port                *PortMock
	tx                  *testutil.DBTx
	sampleCreated       store.IntelDeliveryAttempt
	sampleDelivery      store.IntelDelivery
	sampleAssignedEntry store.AddressBookEntryDetailed
	sampleIntel         store.Intel
	expectedMessages    []kafkautil.OutboundMessage
}

func (suite *PortNotifyIntelDeliveryAttemptCreatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.tx = &testutil.DBTx{}
	suite.sampleIntel = store.Intel{
		ID:         testutil.NewUUIDV4(),
		CreatedAt:  time.Date(2022, 9, 6, 9, 41, 14, 0, time.UTC),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       "until",
		Content:    []byte(`{"hello":"world"}`),
		SearchText: nulls.NewString("true"),
		Importance: 260,
		IsValid:    true,
	}
	userID := testutil.NewUUIDV4()
	suite.sampleAssignedEntry = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          testutil.NewUUIDV4(),
			Label:       "bay",
			Description: "cage",
			Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
			User:        nulls.NewUUID(userID),
		},
		UserDetails: nulls.NewJSONNullable(store.User{
			ID:        userID,
			Username:  "that",
			FirstName: "tribe",
			LastName:  "late",
			IsActive:  true,
		}),
	}
	suite.sampleDelivery = store.IntelDelivery{
		ID:       testutil.NewUUIDV4(),
		Intel:    suite.sampleIntel.ID,
		To:       suite.sampleAssignedEntry.ID,
		IsActive: true,
		Success:  false,
		Note:     nulls.NewString("inch"),
	}
	suite.sampleCreated = store.IntelDeliveryAttempt{
		ID:        testutil.NewUUIDV4(),
		Delivery:  suite.sampleDelivery.ID,
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: time.Date(2022, 9, 1, 12, 18, 37, 0, time.UTC),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusOpen,
		StatusTS:  time.Date(2022, 9, 1, 12, 18, 55, 0, time.UTC),
		Note:      nulls.NewString("variety"),
	}
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.IntelDeliveriesTopic,
			Key:       suite.sampleCreated.Delivery.String(),
			EventType: event.TypeIntelDeliveryAttemptCreated,
			Value: event.IntelDeliveryAttemptCreated{
				ID: suite.sampleCreated.ID,
				Delivery: event.IntelDeliveryAttemptCreatedDelivery{
					ID:       suite.sampleDelivery.ID,
					Intel:    suite.sampleDelivery.Intel,
					To:       suite.sampleDelivery.To,
					IsActive: suite.sampleDelivery.IsActive,
					Success:  suite.sampleDelivery.Success,
					Note:     suite.sampleDelivery.Note,
				},
				AssignedEntry: event.IntelDeliveryAttemptCreatedAssignedEntry{
					ID:          suite.sampleAssignedEntry.ID,
					Label:       suite.sampleAssignedEntry.Label,
					Description: suite.sampleAssignedEntry.Description,
					Operation:   suite.sampleAssignedEntry.Operation,
					User:        suite.sampleAssignedEntry.User,
					UserDetails: nulls.NewJSONNullable(event.IntelDeliveryAttemptCreatedAssignedEntryUserDetails{
						ID:        suite.sampleAssignedEntry.UserDetails.V.ID,
						Username:  suite.sampleAssignedEntry.UserDetails.V.Username,
						FirstName: suite.sampleAssignedEntry.UserDetails.V.FirstName,
						LastName:  suite.sampleAssignedEntry.UserDetails.V.LastName,
						IsActive:  suite.sampleAssignedEntry.UserDetails.V.IsActive,
					}),
				},
				Intel: event.IntelDeliveryAttemptCreatedIntel{
					ID:         suite.sampleIntel.ID,
					CreatedAt:  suite.sampleIntel.CreatedAt,
					CreatedBy:  suite.sampleIntel.CreatedBy,
					Operation:  suite.sampleIntel.Operation,
					Type:       event.IntelType(suite.sampleIntel.Type),
					Content:    suite.sampleIntel.Content,
					SearchText: suite.sampleIntel.SearchText,
					Importance: suite.sampleIntel.Importance,
					IsValid:    suite.sampleIntel.IsValid,
				},
				Channel:   suite.sampleCreated.Channel,
				CreatedAt: suite.sampleCreated.CreatedAt,
				IsActive:  suite.sampleCreated.IsActive,
				Status:    event.IntelDeliveryStatusOpen,
				StatusTS:  suite.sampleCreated.StatusTS,
				Note:      suite.sampleCreated.Note,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyIntelDeliveryAttemptCreatedSuite) TestUnsupportedStatus() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleCreated.Status = "3fD0ZRD"

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryAttemptCreated(timeout, suite.tx, suite.sampleCreated,
			suite.sampleDelivery, suite.sampleAssignedEntry, suite.sampleIntel)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelDeliveryAttemptCreatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryAttemptCreated(timeout, suite.tx, suite.sampleCreated,
			suite.sampleDelivery, suite.sampleAssignedEntry, suite.sampleIntel)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelDeliveryAttemptCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryAttemptCreated(timeout, suite.tx, suite.sampleCreated,
			suite.sampleDelivery, suite.sampleAssignedEntry, suite.sampleIntel)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyIntelDeliveryAttemptCreated(t *testing.T) {
	suite.Run(t, new(PortNotifyIntelDeliveryAttemptCreatedSuite))
}

// PortNotifyIntelDeliveryAttemptStatusUpdatedSuite tests
// Port.NotifyIntelDeliveryAttemptStatusUpdated.
type PortNotifyIntelDeliveryAttemptStatusUpdatedSuite struct {
	suite.Suite
	port             *PortMock
	tx               *testutil.DBTx
	sampleUpdated    store.IntelDeliveryAttempt
	expectedMessages []kafkautil.OutboundMessage
}

func (suite *PortNotifyIntelDeliveryAttemptStatusUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.tx = &testutil.DBTx{}
	suite.sampleUpdated = store.IntelDeliveryAttempt{
		ID:        testutil.NewUUIDV4(),
		Delivery:  testutil.NewUUIDV4(),
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: time.Date(2022, 9, 1, 12, 18, 37, 0, time.UTC),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusAwaitingAck,
		StatusTS:  time.Date(2022, 9, 1, 12, 18, 55, 0, time.UTC),
		Note:      nulls.NewString("variety"),
	}
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.IntelDeliveriesTopic,
			Key:       suite.sampleUpdated.Delivery.String(),
			EventType: event.TypeIntelDeliveryAttemptStatusUpdated,
			Value: event.IntelDeliveryAttemptStatusUpdated{
				ID:       suite.sampleUpdated.ID,
				IsActive: suite.sampleUpdated.IsActive,
				Status:   event.IntelDeliveryStatusAwaitingAck,
				StatusTS: suite.sampleUpdated.StatusTS,
				Note:     suite.sampleUpdated.Note,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyIntelDeliveryAttemptStatusUpdatedSuite) TestUnsupportedStatus() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleUpdated.Status = "9FCqD"

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryAttemptStatusUpdated(timeout, suite.tx, suite.sampleUpdated)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelDeliveryAttemptStatusUpdatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryAttemptStatusUpdated(timeout, suite.tx, suite.sampleUpdated)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelDeliveryAttemptStatusUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryAttemptStatusUpdated(timeout, suite.tx, suite.sampleUpdated)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyIntelDeliveryAttemptStatusUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyIntelDeliveryAttemptStatusUpdatedSuite))
}

// PortNotifyIntelDeliveryCreatedSuite tests
// Port.NotifyIntelDeliveryCreated.
type PortNotifyIntelDeliveryCreatedSuite struct {
	suite.Suite
	port             *PortMock
	tx               *testutil.DBTx
	sampleCreated    store.IntelDelivery
	expectedMessages []kafkautil.OutboundMessage
}

func (suite *PortNotifyIntelDeliveryCreatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.tx = &testutil.DBTx{}
	suite.sampleCreated = store.IntelDelivery{
		ID:       testutil.NewUUIDV4(),
		Intel:    testutil.NewUUIDV4(),
		To:       testutil.NewUUIDV4(),
		IsActive: false,
		Success:  true,
		Note:     nulls.NewString("variety"),
	}
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.IntelDeliveriesTopic,
			Key:       suite.sampleCreated.ID.String(),
			EventType: event.TypeIntelDeliveryCreated,
			Value: event.IntelDeliveryCreated{
				ID:       suite.sampleCreated.ID,
				Intel:    suite.sampleCreated.Intel,
				To:       suite.sampleCreated.To,
				IsActive: suite.sampleCreated.IsActive,
				Success:  suite.sampleCreated.Success,
				Note:     suite.sampleCreated.Note,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyIntelDeliveryCreatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryCreated(timeout, suite.tx, suite.sampleCreated)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelDeliveryCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryCreated(timeout, suite.tx, suite.sampleCreated)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyIntelDeliveryCreated(t *testing.T) {
	suite.Run(t, new(PortNotifyIntelDeliveryCreatedSuite))
}

// PortNotifyIntelDeliveryStatusUpdatedSuite tests
// Port.NotifyIntelDeliveryStatusUpdated.
type PortNotifyIntelDeliveryStatusUpdatedSuite struct {
	suite.Suite
	port              *PortMock
	tx                *testutil.DBTx
	sampleID          uuid.UUID
	sampleNewIsActive bool
	sampleNewSuccess  bool
	sampleNewNote     nulls.String
	expectedMessages  []kafkautil.OutboundMessage
}

func (suite *PortNotifyIntelDeliveryStatusUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.tx = &testutil.DBTx{}
	suite.sampleID = testutil.NewUUIDV4()
	suite.sampleNewIsActive = true
	suite.sampleNewSuccess = true
	suite.sampleNewNote = nulls.NewString("kingdom")
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.IntelDeliveriesTopic,
			Key:       suite.sampleID.String(),
			EventType: event.TypeIntelDeliveryStatusUpdated,
			Value: event.IntelDeliveryStatusUpdated{
				ID:       suite.sampleID,
				IsActive: suite.sampleNewIsActive,
				Success:  suite.sampleNewSuccess,
				Note:     suite.sampleNewNote,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyIntelDeliveryStatusUpdatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryStatusUpdated(timeout, suite.tx, suite.sampleID, suite.sampleNewIsActive,
			suite.sampleNewSuccess, suite.sampleNewNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelDeliveryStatusUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryStatusUpdated(timeout, suite.tx, suite.sampleID, suite.sampleNewIsActive,
			suite.sampleNewSuccess, suite.sampleNewNote)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyIntelDeliveryStatusUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyIntelDeliveryStatusUpdatedSuite))
}

// PortNotifyAddressBookEntryAutoDeliveryUpdatedSuite tests
// Port.NotifyAddressBookEntryAutoDeliveryUpdated.
type PortNotifyAddressBookEntryAutoDeliveryUpdatedSuite struct {
	suite.Suite
	port             *PortMock
	entryID          uuid.UUID
	isEnabled        bool
	expectedMessages []kafkautil.OutboundMessage
}

func (suite *PortNotifyAddressBookEntryAutoDeliveryUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.entryID = testutil.NewUUIDV4()
	suite.isEnabled = true
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.IntelDeliveriesTopic,
			Key:       suite.entryID.String(),
			EventType: event.TypeAddressBookEntryAutoDeliveryUpdated,
			Value: event.AddressBookEntryAutoDeliveryUpdated{
				ID:                    suite.entryID,
				IsAutoDeliveryEnabled: suite.isEnabled,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyAddressBookEntryAutoDeliveryUpdatedSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyAddressBookEntryAutoDeliveryUpdated(context.Background(), &testutil.DBTx{}, suite.entryID, suite.isEnabled)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyAddressBookEntryAutoDeliveryUpdatedSuite) TestOK() {
	err := suite.port.Port.NotifyAddressBookEntryAutoDeliveryUpdated(context.Background(), &testutil.DBTx{}, suite.entryID, suite.isEnabled)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
}

func TestPort_NotifyAddressBookEntryAutoDeliveryUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyAddressBookEntryAutoDeliveryUpdatedSuite))
}
