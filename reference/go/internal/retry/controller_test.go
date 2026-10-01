package retry

import (
	"testing"
	"time"
)

func TestNoRetryAfterDownstreamStarted(t *testing.T) {
	s := State{Attempt: 1, MaxAttempts: 3, DownstreamStarted: true}
	if CanRetry(s, RetryAnother, time.Now()) {
		t.Fatal("must not retry after downstream has started")
	}
}

func TestRetryWithinBudget(t *testing.T) {
	s := State{Attempt: 1, MaxAttempts: 3, Deadline: time.Now().Add(time.Minute)}
	if !CanRetry(s, RetryAnother, time.Now()) {
		t.Fatal("expected retry to be permitted")
	}
}
