package eventport

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

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
	suite.sampleCreated.Assignments = []store.IntelAssignment{
		{
			ID:    testutil.NewUUIDV4(),
			Intel: suite.sampleCreated.ID,
			To:    testutil.NewUUIDV4(),
		},
		{
			ID:    testutil.NewUUIDV4(),
			Intel: suite.sampleCreated.ID,
			To:    testutil.NewUUIDV4(),
		},
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
				Assignments: []event.IntelAssignment{
					{
						ID: suite.sampleCreated.Assignments[0].ID,
						To: suite.sampleCreated.Assignments[0].To,
					},
					{
						ID: suite.sampleCreated.Assignments[1].ID,
						To: suite.sampleCreated.Assignments[1].To,
					},
				},
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
