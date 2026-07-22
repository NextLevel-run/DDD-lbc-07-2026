package domain

// PasswordHasher is the port through which the seller's plaintext password is
// turned into a hash. Hashing (bcrypt, argon2, ...) is an infrastructure
// concern, so the algorithm lives in a driven adapter, never in the domain.
// The application hashes the plaintext via this port, then feeds the resulting
// hash to the aggregate through NewClassifiedAd / NewHashedPassword.
type PasswordHasher interface {
	// Hash turns a plaintext password into a hash suitable for storage.
	Hash(plainPassword string) (string, error)
}
