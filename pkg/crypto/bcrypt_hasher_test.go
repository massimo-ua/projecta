package crypto_test

import (
	"testing"

	"gitlab.com/massimo-ua/projecta/pkg/crypto"
)

func TestBcryptHasher(t *testing.T) {
	t.Run("Default cost initialization and hashing", func(t *testing.T) {
		hasher := crypto.NewBcryptHasher(0)
		password := "my_secure_password"

		hash, err := hasher.Hash(password)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if hash == "" {
			t.Fatalf("expected non-empty hash")
		}

		if !hasher.Compare(password, hash) {
			t.Errorf("expected Compare to return true for matching password")
		}

		if hasher.Compare("wrong_password", hash) {
			t.Errorf("expected Compare to return false for mismatched password")
		}
	})

	t.Run("Invalid cost error", func(t *testing.T) {
		hasher := crypto.NewBcryptHasher(100) // invalid cost > bcrypt.MaxCost (31)
		_, err := hasher.Hash("test")
		if err == nil {
			t.Errorf("expected error for invalid cost, got nil")
		}
	})
}
