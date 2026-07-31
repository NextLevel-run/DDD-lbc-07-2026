package clock

import "time"

// FixedClock is a domain.Clock implementation that always returns a fixed
// point in time. It is intended for use in tests (from any package, hence
// this file is NOT a _test.go file).
type FixedClock struct {
	t time.Time
}

// NewFixedClock creates a new FixedClock fixed at t.
func NewFixedClock(t time.Time) *FixedClock {
	return &FixedClock{t: t}
}

// Now returns the fixed time the clock was created with.
func (c *FixedClock) Now() time.Time {
	return c.t
}
