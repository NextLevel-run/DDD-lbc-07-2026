package domain_test

import (
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewSubmissionDate(t *testing.T) {
	now := time.Now()
	s := domain.NewSubmissionDate(now)
	assert.Equal(t, now, s.Time())
}
