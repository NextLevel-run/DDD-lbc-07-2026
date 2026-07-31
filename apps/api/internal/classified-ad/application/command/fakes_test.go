package command

import (
	"errors"
	"sync"
	"testing"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeClassifiedAdRepository is a minimal in-memory implementation of
// domain.ClassifiedAdRepository, used only because the real "inmemory" driven
// adapter does not exist yet in this parallel implementation step.
type fakeClassifiedAdRepository struct {
	mu  sync.Mutex
	ads map[uuid.UUID]*domain.ClassifiedAd
}

func newFakeClassifiedAdRepository() *fakeClassifiedAdRepository {
	return &fakeClassifiedAdRepository{ads: make(map[uuid.UUID]*domain.ClassifiedAd)}
}

func (r *fakeClassifiedAdRepository) Save(ad *domain.ClassifiedAd) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ads[ad.ID()] = ad
	return nil
}

func (r *fakeClassifiedAdRepository) FindByID(id uuid.UUID) (*domain.ClassifiedAd, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ad, ok := r.ads[id]
	if !ok {
		return nil, domain.ErrClassifiedAdNotFound
	}
	return ad, nil
}

func (r *fakeClassifiedAdRepository) FindExpirable(now time.Time) ([]*domain.ClassifiedAd, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.ClassifiedAd, 0)
	for _, ad := range r.ads {
		if ad.IsExpirable(now) {
			result = append(result, ad)
		}
	}
	return result, nil
}

func (r *fakeClassifiedAdRepository) Search(criteria domain.SearchCriteria) ([]*domain.ClassifiedAd, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.ClassifiedAd, 0)
	for _, ad := range r.ads {
		if ad.IsOnline() {
			result = append(result, ad)
		}
	}
	return result, nil
}

// count returns the number of ads currently stored (test helper only).
func (r *fakeClassifiedAdRepository) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.ads)
}

// newSubmittedAd builds a valid ClassifiedAd in StatusSubmitted, owned by
// seller@example.com with plaintext password "supersecret" (hashed with
// fakePasswordHasher), without saving it.
func newSubmittedAd(t *testing.T, submittedAt time.Time) *domain.ClassifiedAd {
	t.Helper()

	email, err := domain.NewEmail("seller@example.com")
	require.NoError(t, err)
	password, err := domain.NewPassword("supersecret", fakePasswordHasher{})
	require.NoError(t, err)
	seller, err := domain.NewSeller(email, "seller-pseudo", password)
	require.NoError(t, err)
	category, err := domain.NewCategory(string(domain.CategoryConsumerGoods))
	require.NoError(t, err)
	location, err := domain.NewLocation("75001", "Paris")
	require.NoError(t, err)

	ad, err := domain.NewClassifiedAd(
		"Vélo hollandais",
		"Très bon état, peu servi.",
		15000,
		seller,
		[]string{"http://img/1.jpg"},
		category,
		location,
		domain.NewSubmissionDate(submittedAt),
	)
	require.NoError(t, err)

	return ad
}

// seedSubmittedAd creates and persists a valid classified ad still awaiting
// moderation (StatusSubmitted), returning it.
func seedSubmittedAd(t *testing.T, repo *fakeClassifiedAdRepository, submittedAt time.Time) *domain.ClassifiedAd {
	t.Helper()

	ad := newSubmittedAd(t, submittedAt)
	require.NoError(t, repo.Save(ad))
	return ad
}

// fakePasswordHasher is a plaintext-passthrough implementation of
// domain.PasswordHasher for tests: it "hashes" by prefixing the plaintext.
type fakePasswordHasher struct{}

var errFakeHashMismatch = errors.New("hash mismatch")

func (fakePasswordHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (fakePasswordHasher) Compare(hash, plain string) error {
	if hash == "hashed:"+plain {
		return nil
	}
	return errFakeHashMismatch
}

// fakeClock is a settable implementation of domain.Clock for deterministic tests.
type fakeClock struct {
	now time.Time
}

func newFakeClock(t time.Time) *fakeClock {
	return &fakeClock{now: t}
}

func (c *fakeClock) Now() time.Time {
	return c.now
}
