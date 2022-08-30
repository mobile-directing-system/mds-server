package eventport

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

func Test_intelTypeFromStore(t *testing.T) {
	testutil.TestMapperWithConstExtraction(t, intelTypeFromStore, "../store/intel.go", nulls.String{})
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
		ID:         testutil.NewUUIDV4(),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       store.IntelTypePlainTextMessage,
		Content:    json.RawMessage(`null`),
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
				ID:         suite.sampleCreated.ID,
				CreatedAt:  suite.sampleCreated.CreatedAt,
				CreatedBy:  suite.sampleCreated.CreatedBy,
				Operation:  suite.sampleCreated.Operation,
				Type:       event.IntelTypePlainTextMessage,
				Content:    suite.sampleCreated.Content,
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
	suite.sampleCreated.Type = "3fD0ZRD"

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
