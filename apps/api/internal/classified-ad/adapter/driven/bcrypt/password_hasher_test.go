package bcrypt_test

import (
	"testing"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/bcrypt"
)

func TestBcryptPasswordHasher_HashThenCompare_Succeeds(t *testing.T) {
	hasher := bcrypt.NewBcryptPasswordHasher()

	hash, err := hasher.Hash("s3cr3tPassword")
	if err != nil {
		t.Fatalf("Hash returned an error: %v", err)
	}

	if err := hasher.Compare(hash, "s3cr3tPassword"); err != nil {
		t.Errorf("Compare with correct password should succeed, got error: %v", err)
	}
}

func TestBcryptPasswordHasher_CompareWithWrongPassword_Fails(t *testing.T) {
	hasher := bcrypt.NewBcryptPasswordHasher()

	hash, err := hasher.Hash("s3cr3tPassword")
	if err != nil {
		t.Fatalf("Hash returned an error: %v", err)
	}

	if err := hasher.Compare(hash, "wrongPassword"); err == nil {
		t.Error("Compare with wrong password should fail, got nil error")
	}
}
