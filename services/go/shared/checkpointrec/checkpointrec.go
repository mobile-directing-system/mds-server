// Package checkpointrec is a testutil which provides Recorder which is used for
// recording checkpoints and providing functions for assuring order and other
// properties in tests.
package checkpointrec

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"sync"
	"testing"
)

// Recorder allows recording checkpoints via Checkpoint and many methods for
// testing and checking the recorded checkpoints. Create a new Recorder using
// NewRecorder.
type Recorder struct {
	RecorderAsserter
	checkpoints         []string
	waitersByCheckpoint map[string][]chan struct{}
	m                   sync.Mutex
}

// NewRecorder creates a new Recorder ready to use.
func NewRecorder() *Recorder {
	rec := &Recorder{
		checkpoints:         make([]string, 0),
		waitersByCheckpoint: make(map[string][]chan struct{}),
	}
	rec.RecorderAsserter = RecorderAsserter{
		rec: rec,
		fail: func(t *testing.T, failureMessage string, msgAndArgs ...any) {
			assert.Fail(t, failureMessage, msgAndArgs...)
		},
	}
	return rec
}

// Checkpoint records the given checkpoint.
func (rec *Recorder) Checkpoint(name string) {
	rec.m.Lock()
	defer rec.m.Unlock()
	rec.checkpoints = append(rec.checkpoints, name)
	// Unblock waiters.
	if waiters, ok := rec.waitersByCheckpoint[name]; ok {
		for _, waiter := range waiters {
			close(waiter)
		}
		delete(rec.waitersByCheckpoint, name)
	}
}

// WaitForCheckpoint returns a channel that closes once the checkpoint with the
// given name was logged.
func (rec *Recorder) WaitForCheckpoint(name string) <-chan struct{} {
	rec.m.Lock()
	defer rec.m.Unlock()
	c := make(chan struct{})
	for _, checkpoint := range rec.checkpoints {
		if checkpoint == name {
			close(c)
			return c
		}
	}
	rec.waitersByCheckpoint[name] = append(rec.waitersByCheckpoint[name], c)
	return c
}

// WaitForNextCheckpoint returns a channel that closes the next checkpoint with
// the given name was logged.
func (rec *Recorder) WaitForNextCheckpoint(name string) <-chan struct{} {
	rec.m.Lock()
	defer rec.m.Unlock()
	c := make(chan struct{})
	rec.waitersByCheckpoint[name] = append(rec.waitersByCheckpoint[name], c)
	return c
}

// Require returns a RecorderAsserter that uses the failer from require package.
func (rec *Recorder) Require() *RecorderAsserter {
	return &RecorderAsserter{
		rec: rec,
		fail: func(t *testing.T, failureMessage string, msgAndArgs ...any) {
			require.Fail(t, failureMessage, msgAndArgs...)
		},
	}
}

// RecorderAsserter provides assertions for Recorder.
type RecorderAsserter struct {
	rec  *Recorder
	fail func(t *testing.T, failureMessage string, msgAndArgs ...any)
}

// Includes asserts that the given checkpoint is included.
func (ra *RecorderAsserter) Includes(t *testing.T, name string, msgAndArgs ...any) {
	ra.rec.m.Lock()
	defer ra.rec.m.Unlock()
	for _, checkpoint := range ra.rec.checkpoints {
		if checkpoint == name {
			return
		}
	}
	errMessage := fmt.Sprintf(`checkpoint not included
want: %s
logged: %s`, name, strings.Join(ra.rec.checkpoints, ", "))
	ra.fail(t, errMessage, msgAndArgs...)
}

// Before asserts that the first appearance of first is before the first
// appearance of follower.
func (ra *RecorderAsserter) Before(t *testing.T, first string, follower string, msgAndArgs ...any) {
	ra.rec.m.Lock()
	defer ra.rec.m.Unlock()
	firstPos := -1
	followerPos := -1
	for pos, checkpoint := range ra.rec.checkpoints {
		if firstPos == -1 && checkpoint == first {
			firstPos = pos
		}
		if followerPos == -1 && checkpoint == follower {
			followerPos = pos
		}
	}
	// Assure found.
	errDetails := fmt.Sprintf(`want first: %s
want follower: %s
logged: %s`, first, follower, strings.Join(ra.rec.checkpoints, ", "))
	if firstPos == -1 && followerPos == -1 {
		ra.fail(t, "neither first nor follower found\n"+errDetails, msgAndArgs...)
		return
	}
	if firstPos == -1 {
		ra.fail(t, "first not found\n"+errDetails, msgAndArgs...)
		return
	}
	if followerPos == -1 {
		ra.fail(t, "follower not found\n"+errDetails, msgAndArgs...)
		return
	}
	if !(firstPos < followerPos) {
		ra.fail(t, fmt.Sprintf("first at pos %d not before follower at pos %d\n"+errDetails, firstPos, followerPos))
		return
	}
}

// AllBefore asserts that all appearances of first are before any appearances of
// follower.
func (ra *RecorderAsserter) AllBefore(t *testing.T, first string, follower string, msgAndArgs ...any) {
	ra.rec.m.Lock()
	defer ra.rec.m.Unlock()
	firstLastPos := -1
	followerFirstPos := -1
	for pos, checkpoint := range ra.rec.checkpoints {
		if checkpoint == first {
			firstLastPos = pos
		}
		if followerFirstPos == -1 && checkpoint == follower {
			followerFirstPos = pos
		}
	}
	// Assure found.
	errDetails := fmt.Sprintf(`want first: %s
want follower: %s
logged: %s`, first, follower, strings.Join(ra.rec.checkpoints, ", "))
	if firstLastPos == -1 && followerFirstPos == -1 {
		ra.fail(t, "neither first nor follower found\n"+errDetails, msgAndArgs...)
		return
	}
	if firstLastPos == -1 {
		ra.fail(t, "first not found\n"+errDetails, msgAndArgs...)
		return
	}
	if followerFirstPos == -1 {
		ra.fail(t, "follower not found\n"+errDetails, msgAndArgs...)
		return
	}
	if !(firstLastPos < followerFirstPos) {
		ra.fail(t, fmt.Sprintf("first at pos %d not before follower at pos %d\n"+errDetails, firstLastPos, followerFirstPos))
		return
	}
}

// IncludesOrdered assures that the given elements occurr in the specified order.
func (ra *RecorderAsserter) IncludesOrdered(t *testing.T, expect []string) {
	ra.rec.m.Lock()
	defer ra.rec.m.Unlock()
	if len(expect) == 0 {
		return
	}

	currentExpect := 0
	genErrDetails := func() string {
		return fmt.Sprintf(`expect order: %s
current expect: %s (%d of %d)
checkpoints: %s`, strings.Join(expect, ", "),
			expect[currentExpect], currentExpect+1, len(expect), strings.Join(ra.rec.checkpoints, ", "))
	}
checkAllCheckpoints:
	for pos, checkpoint := range ra.rec.checkpoints {
		for _, e := range expect {
			if e != checkpoint {
				continue
			}
			if e != expect[currentExpect] {
				ra.fail(t, fmt.Sprintf("order mismatch (got %s at pos %d)\n%s", e, pos, genErrDetails()))
				return
			}
			// Correct one.
			currentExpect++
			if currentExpect == len(expect) {
				break checkAllCheckpoints
			}
		}
	}
	if currentExpect != len(expect) {
		ra.fail(t, "checkpoint "+expect[currentExpect]+"not included\n"+genErrDetails())
		return
	}
}

// Fail fails and logs all logged checkpoints.
func (ra *RecorderAsserter) Fail(t *testing.T, failureMessage string, msgAndArgs ...any) {
	ra.rec.m.Lock()
	defer ra.rec.m.Unlock()
	ra.fail(t, fmt.Sprintf("%s\ncheckpoints: %s", failureMessage, strings.Join(ra.rec.checkpoints, ", ")), msgAndArgs...)
}
