package password

import (
	"ddd-second-hand-marketplace/internal/classified-ad/domain"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implements the domain.PasswordHasher port using bcrypt, so the
// seller's password is never stored in clear text.
type BcryptHasher struct {
	cost int
}

// compile-time check that the adapter satisfies the domain port.
var _ domain.PasswordHasher = (*BcryptHasher)(nil)

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptHasher) Hash(plainPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
