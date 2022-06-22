package testutil

import (
	"context"
	"time"
)

// TestFailer is an abstraction for failing tests instantly.
type TestFailer interface {
	FailNow(string, ...interface{}) bool
}

// NewTimeout wraps timeout-detection-functionality. It creates a
// timeout-context, a cancel-function and a wait function. The intended usage is
// creating everything using NewTimeout and then either deferring or, if post
// checks are required, calling the wait-function. The wait-function will also
// definitely call the cancel-function.
func NewTimeout(failer TestFailer, timeout time.Duration) (context.Context, context.CancelFunc, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return ctx, cancel, func() {
		defer cancel()
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			failer.FailNow("timeout", "should not time out")
		}
	}
}
