package domain

// Password is a value object wrapping a hashed password. The hash is never
// exposed directly; use Matches to verify a plaintext candidate.
type Password struct {
	hash string
}

// NewPassword validates the plaintext password and hashes it via the given PasswordHasher.
func NewPassword(plain string, hasher PasswordHasher) (Password, error) {
	if len(plain) < 8 {
		return Password{}, ErrPasswordTooShort
	}
	hash, err := hasher.Hash(plain)
	if err != nil {
		return Password{}, err
	}
	return Password{hash: hash}, nil
}

// Matches returns true if the given plaintext password matches this Password's hash.
func (p Password) Matches(plain string, hasher PasswordHasher) bool {
	return hasher.Compare(p.hash, plain) == nil
}
