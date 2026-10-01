package retry

import "time"

type Action string

const (
	Stop         Action = "stop"
	RetrySame    Action = "retry_same"
	RetryAnother Action = "retry_another"
)

type State struct {
	Attempt           int
	MaxAttempts       int
	DownstreamStarted bool
	Deadline          time.Time
}

// CanRetry captures the core safety gate. Provider-specific policy still comes
// from the template; this function only enforces request-level limits.
func CanRetry(s State, action Action, now time.Time) bool {
	if action == Stop {
		return false
	}
	if s.DownstreamStarted {
		return false
	}
	if s.Attempt >= s.MaxAttempts {
		return false
	}
	if !s.Deadline.IsZero() && !now.Before(s.Deadline) {
		return false
	}
	return true
}
