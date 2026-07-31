package bcrypt

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher implements domain.PasswordHasher using bcrypt.
type BcryptPasswordHasher struct{}

// NewBcryptPasswordHasher creates a new BcryptPasswordHasher.
func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{}
}

// Hash hashes the given plain-text password using bcrypt's default cost.
func (h *BcryptPasswordHasher) Hash(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Compare returns nil if hash is the bcrypt hash of plain, an error otherwise.
func (h *BcryptPasswordHasher) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
