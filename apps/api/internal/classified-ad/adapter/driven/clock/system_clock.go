package clock

import "time"

// SystemClock is a domain.Clock implementation backed by the real system time.
type SystemClock struct{}

// NewSystemClock creates a new SystemClock.
func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

// Now returns the current system time.
func (c *SystemClock) Now() time.Time {
	return time.Now()
}
