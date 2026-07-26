package crypto_test

import (
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/pkg/crypto"
)

type mockHasher struct {
	hashFunc    func(value string) (string, error)
	compareFunc func(value, hash string) bool
}

func (m *mockHasher) Hash(value string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(value)
	}
	return "hashed_" + value, nil
}

func (m *mockHasher) Compare(value string, hash string) bool {
	if m.compareFunc != nil {
		return m.compareFunc(value, hash)
	}
	return hash == "hashed_"+value
}

func TestJwtTokenProvider(t *testing.T) {
	secret := "super-secret-key-1234567890"
	hasher := &mockHasher{}
	provider := crypto.NewJwtTokenProvider(secret, 3600, hasher)

	payload := core.AuthTokenPayload{
		Sub:         "user-123",
		DisplayName: "John Doe",
		Roles:       []string{"admin", "user"},
	}

	t.Run("GenerateTokenRing success", func(t *testing.T) {
		res, err := provider.GenerateTokenRing(payload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.AccessToken == "" || res.RefreshToken == "" {
			t.Fatalf("expected non-empty tokens")
		}

		claims, err := provider.ValidateToken(res.AccessToken)
		if err != nil {
			t.Fatalf("failed to validate token: %v", err)
		}
		if claims.Sub != payload.Sub || claims.DisplayName != payload.DisplayName {
			t.Errorf("claims mismatch: %+v", claims)
		}
	})

	t.Run("GenerateTokenRing without roles", func(t *testing.T) {
		pNoRoles := core.AuthTokenPayload{
			Sub:         "user-456",
			DisplayName: "Jane Doe",
			Roles:       nil,
		}
		res, err := provider.GenerateTokenRing(pNoRoles)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		claims, err := provider.ValidateToken(res.AccessToken)
		if err != nil {
			t.Fatalf("failed to validate token: %v", err)
		}
		if len(claims.Roles) != 0 {
			t.Errorf("expected empty roles, got %v", claims.Roles)
		}
	})

	t.Run("GenerateTokenRing hasher error", func(t *testing.T) {
		errHasher := &mockHasher{
			hashFunc: func(value string) (string, error) {
				return "", errors.New("hasher failure")
			},
		}
		pErr := crypto.NewJwtTokenProvider(secret, 3600, errHasher)
		_, err := pErr.GenerateTokenRing(payload)
		if err == nil {
			t.Errorf("expected error when hasher fails, got nil")
		}
	})

	t.Run("ValidateToken with expired token", func(t *testing.T) {
		expiredProvider := crypto.NewJwtTokenProvider(secret, -10, hasher)
		res, err := expiredProvider.GenerateTokenRing(payload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = provider.ValidateToken(res.AccessToken)
		if !errors.Is(err, core.AuthTokenIsInvalid) {
			t.Errorf("expected AuthTokenIsInvalid error for expired token, got %v", err)
		}
	})

	t.Run("ValidateToken with invalid signature", func(t *testing.T) {
		res, _ := provider.GenerateTokenRing(payload)
		invalidProvider := crypto.NewJwtTokenProvider("wrong-secret", 3600, hasher)
		_, err := invalidProvider.ValidateToken(res.AccessToken)
		if err == nil {
			t.Errorf("expected error for invalid signature, got nil")
		}
	})

	t.Run("DecodeToken with valid and expired tokens", func(t *testing.T) {
		expiredProvider := crypto.NewJwtTokenProvider(secret, -10, hasher)
		res, _ := expiredProvider.GenerateTokenRing(payload)

		claims, err := provider.DecodeToken(res.AccessToken)
		if err != nil {
			t.Fatalf("expected DecodeToken to succeed for expired token, got error: %v", err)
		}
		if claims.Sub != payload.Sub {
			t.Errorf("expected sub %s, got %s", payload.Sub, claims.Sub)
		}

		_, err = provider.DecodeToken("malformed.token.string")
		if err == nil {
			t.Errorf("expected error for malformed token, got nil")
		}
	})

	t.Run("ValidateToken with no exp claim", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"jti":          uuid.New().String(),
			"sub":          "user-no-exp",
			"display_name": "No Exp User",
		})
		tokenStr, _ := token.SignedString([]byte(secret))

		claims, err := provider.ValidateToken(tokenStr)
		if err != nil {
			t.Fatalf("expected ValidateToken success when no exp claim, got %v", err)
		}
		if claims.Sub != "user-no-exp" {
			t.Errorf("unexpected sub: %s", claims.Sub)
		}
	})

	t.Run("toDomain with []string roles type directly", func(t *testing.T) {
		claims := jwt.MapClaims{
			"jti":          uuid.New().String(),
			"sub":          "sub1",
			"display_name": "disp1",
			"roles":        []string{"admin", "editor"},
		}
		domain, err := provider.DecodeToken("")
		_ = err
		// Call toDomain directly via provider validation of manually created token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString([]byte(secret))

		domain, err = provider.ValidateToken(tokenStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(domain.Roles) != 2 || domain.Roles[0] != "admin" {
			t.Errorf("unexpected roles: %v", domain.Roles)
		}
	})

	t.Run("ValidateRefreshToken", func(t *testing.T) {
		tokenID := uuid.New()
		validRef := "hashed_" + tokenID.String()

		if !provider.ValidateRefreshToken(tokenID, validRef) {
			t.Errorf("expected ValidateRefreshToken to return true")
		}
		if provider.ValidateRefreshToken(tokenID, "invalid_ref") {
			t.Errorf("expected ValidateRefreshToken to return false")
		}
	})
}
