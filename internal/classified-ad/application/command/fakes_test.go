package command

import (
	"errors"
	"sync"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"github.com/google/uuid"
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
