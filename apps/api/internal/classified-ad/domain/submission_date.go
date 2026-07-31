package domain

import "time"

// SubmissionDate is a value object representing when a classified ad was submitted.
type SubmissionDate struct {
	value time.Time
}

// NewSubmissionDate builds a SubmissionDate from a time.Time.
func NewSubmissionDate(t time.Time) SubmissionDate {
	return SubmissionDate{value: t}
}

// Time returns the underlying time.Time value.
func (s SubmissionDate) Time() time.Time {
	return s.value
}
