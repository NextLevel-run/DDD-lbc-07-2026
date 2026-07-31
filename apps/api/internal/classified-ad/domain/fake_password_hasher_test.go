package domain_test

import "errors"

// fakePasswordHasher is a simple in-memory implementation of domain.PasswordHasher
// for tests: it "hashes" by prefixing the plaintext, and compares accordingly.
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
