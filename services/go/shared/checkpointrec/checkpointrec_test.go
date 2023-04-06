package checkpointrec

import (
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"
	"testing"
)

func TestRecorder_Includes(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		checkpointCount := rapid.IntRange(0, 512).Draw(t, "checkpoint_count")
		included := rapid.Bool().Draw(t, "included")
		includedPos := int(rapid.Float64Range(0, 1).Draw(t, "included_pos") * float64(checkpointCount-1))

		rec := NewRecorder()
		failed := false
		rec.fail = func(_ *testing.T, _ string, _ ...any) {
			failed = true
		}
		for checkpointNum := 0; checkpointNum < checkpointCount; checkpointNum++ {
			if included && includedPos == checkpointNum {
				rec.Checkpoint("hello")
			} else {
				rec.Checkpoint(testutil.NewUUIDV4().String())
			}
		}

		rec.Includes(nil, "hello")
		assert.Equal(t, included && checkpointCount > 0, !failed, "should report correct result")
	})
}

// RecorderBeforeSuite tests RecorderAsserter.Before.
type RecorderBeforeSuite struct {
	suite.Suite
	rec      *Recorder
	failed   bool
	first    string
	follower string
}

func (suite *RecorderBeforeSuite) SetupTest() {
	suite.rec = NewRecorder()
	suite.failed = false
	suite.rec.fail = func(_ *testing.T, _ string, _ ...any) {
		suite.failed = true
	}
	suite.first = "x_first"
	suite.follower = "x_follower"
}

func (suite *RecorderBeforeSuite) TestEmpty() {
	suite.rec.Before(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderBeforeSuite) TestBothNotIncluded() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderBeforeSuite) TestFirstNotIncluded() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderBeforeSuite) TestFollowerNotIncluded() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderBeforeSuite) TestNotBefore1() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderBeforeSuite) TestNotBefore2() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderBeforeSuite) TestNotBefore3() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderBeforeSuite) TestOKBefore1() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.follower)

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.False(suite.failed, "should not fail")
}

func (suite *RecorderBeforeSuite) TestOKBefore2() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.False(suite.failed, "should not fail")
}

func (suite *RecorderBeforeSuite) TestOKBefore3() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)

	suite.rec.Before(nil, suite.first, suite.follower)
	suite.False(suite.failed, "should not fail")
}

func TestRecorder_Before(t *testing.T) {
	suite.Run(t, new(RecorderBeforeSuite))
}

// RecorderAllBeforeSuite tests RecorderAsserter.AllBefore.
type RecorderAllBeforeSuite struct {
	suite.Suite
	rec      *Recorder
	failed   bool
	first    string
	follower string
}

func (suite *RecorderAllBeforeSuite) SetupTest() {
	suite.rec = NewRecorder()
	suite.failed = false
	suite.rec.fail = func(_ *testing.T, _ string, _ ...any) {
		suite.failed = true
	}
	suite.first = "x_first"
	suite.follower = "x_follower"
}

func (suite *RecorderAllBeforeSuite) TestEmpty() {
	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderAllBeforeSuite) TestBothNotIncluded() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderAllBeforeSuite) TestFirstNotIncluded() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderAllBeforeSuite) TestFollowerNotIncluded() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderAllBeforeSuite) TestNotAllBefore1() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderAllBeforeSuite) TestNotAllBefore2() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderAllBeforeSuite) TestNotAllBefore3() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderAllBeforeSuite) TestNotAllBefore4() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderAllBeforeSuite) TestOKAllBefore1() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.follower)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.False(suite.failed, "should not fail")
}

func (suite *RecorderAllBeforeSuite) TestOKAllBefore2() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.follower)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.False(suite.failed, "should not fail")
}

func (suite *RecorderAllBeforeSuite) TestOKAllBefore3() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(suite.first)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.follower)
	suite.rec.Checkpoint(suite.follower)

	suite.rec.AllBefore(nil, suite.first, suite.follower)
	suite.False(suite.failed, "should not fail")
}

func TestRecorder_AllBefore(t *testing.T) {
	suite.Run(t, new(RecorderAllBeforeSuite))
}

// RecorderIncludesOrderedSuite tests RecorderAsserter.IncludesOrdered.
type RecorderIncludesOrderedSuite struct {
	suite.Suite
	rec      *Recorder
	failed   bool
	c1       string
	c2       string
	c3       string
	includes []string
}

func (suite *RecorderIncludesOrderedSuite) SetupTest() {
	suite.rec = NewRecorder()
	suite.failed = false
	suite.rec.fail = func(_ *testing.T, _ string, _ ...any) {
		suite.failed = true
	}
	suite.c1 = "x_c1"
	suite.c2 = "x_c2"
	suite.c3 = "x_c3"
	suite.includes = []string{suite.c1, suite.c2, suite.c3}
}

func (suite *RecorderIncludesOrderedSuite) TestEmpty() {
	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderIncludesOrderedSuite) TestNotIncluded() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderIncludesOrderedSuite) TestPartiallyIncludedInOrder() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderIncludesOrderedSuite) TestPartiallyIncludedOutOfOrder() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderIncludesOrderedSuite) TestOutOfOrder1() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c3)
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderIncludesOrderedSuite) TestOutOfOrder2() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c3)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderIncludesOrderedSuite) TestOutOfOrder3() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c3)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.True(suite.failed, "should fail")
}

func (suite *RecorderIncludesOrderedSuite) TestOKInOrder1() {
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())
	suite.rec.Checkpoint(suite.c3)
	suite.rec.Checkpoint(testutil.NewUUIDV4().String())

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.False(suite.failed, "should not fail")
}

func (suite *RecorderIncludesOrderedSuite) TestOKInOrder2() {
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(suite.c3)

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.False(suite.failed, "should not fail")
}

func (suite *RecorderIncludesOrderedSuite) TestOKInOrder4() {
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(suite.c2)
	suite.rec.Checkpoint(suite.c3)
	suite.rec.Checkpoint(suite.c1)
	suite.rec.Checkpoint(suite.c3)

	suite.rec.IncludesOrdered(nil, suite.includes)
	suite.False(suite.failed, "should not fail")
}

func TestRecorder_IncludesOrderded(t *testing.T) {
	suite.Run(t, new(RecorderIncludesOrderedSuite))
}
