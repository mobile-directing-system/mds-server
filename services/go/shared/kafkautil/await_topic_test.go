package kafkautil

import (
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/stretchr/testify/suite"
	"testing"
)

// topicAwaiterSuite tests topicAwaiter.
type topicAwaiterSuite struct {
	suite.Suite
}

func (suite *topicAwaiterSuite) TestNoneExpected() {
	awaiter := newTopicAwaiter()
	suite.Equal(0, awaiter.remainingExpected(), "should return correct amount of remaining expected")
}

func (suite *topicAwaiterSuite) TestMarkUnknownAsAvailable() {
	awaiter := newTopicAwaiter("hello")
	awaiter.markTopicAsAvailable("world")
	suite.Equal(1, awaiter.remainingExpected(), "should return correct amount of remaining expected")
}

func (suite *topicAwaiterSuite) TestOK() {
	expectedTopics := []event.Topic{"hello", "world", "!", "i", "love", "cookies", "."}
	awaiter := newTopicAwaiter(expectedTopics...)
	for i, topic := range expectedTopics {
		suite.Equal(len(expectedTopics)-i, awaiter.remainingExpected(), "should return correct amount of remaining expected")
		awaiter.markTopicAsAvailable(topic)
		suite.Equal(len(expectedTopics)-i-1, awaiter.remainingExpected(), "should return correct amount of remaining expected")
	}
}

func Test_topicAwaiter(t *testing.T) {
	suite.Run(t, new(topicAwaiterSuite))
}
