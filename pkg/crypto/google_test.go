package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGoogleAuthProvider(t *testing.T) {
	t.Run("invalid config missing client id", func(t *testing.T) {
		_, err := NewGoogleAuthProvider(GoogleAuthConfig{
			ClientID:     "",
			ClientSecret: "secret",
		})
		if err == nil {
			t.Errorf("expected error when ClientID is empty")
		}
	})

	t.Run("invalid config missing client secret", func(t *testing.T) {
		_, err := NewGoogleAuthProvider(GoogleAuthConfig{
			ClientID:     "client-id",
			ClientSecret: "",
		})
		if err == nil {
			t.Errorf("expected error when ClientSecret is empty")
		}
	})

	t.Run("validateClaims branches", func(t *testing.T) {
		provider, _ := NewGoogleAuthProvider(GoogleAuthConfig{
			ClientID:     "valid-client-id",
			ClientSecret: "secret",
		})

		// Expired claim
		expiredClaims := jwt.MapClaims{
			"exp": float64(time.Now().Add(-10 * time.Second).Unix()),
			"aud": "valid-client-id",
			"iss": "accounts.google.com",
		}
		if err := provider.validateClaims(expiredClaims); err == nil {
			t.Errorf("expected error for expired claims")
		}

		// Invalid aud claim
		invalidAudClaims := jwt.MapClaims{
			"exp": float64(time.Now().Add(10 * time.Minute).Unix()),
			"aud": "wrong-client-id",
			"iss": "accounts.google.com",
		}
		if err := provider.validateClaims(invalidAudClaims); err == nil {
			t.Errorf("expected error for invalid audience")
		}

		// Invalid iss claim
		invalidIssClaims := jwt.MapClaims{
			"exp": float64(time.Now().Add(10 * time.Minute).Unix()),
			"aud": "valid-client-id",
			"iss": "malicious.com",
		}
		if err := provider.validateClaims(invalidIssClaims); err == nil {
			t.Errorf("expected error for invalid issuer")
		}

		// Valid claims (with https://accounts.google.com)
		validClaimsHttps := jwt.MapClaims{
			"exp": float64(time.Now().Add(10 * time.Minute).Unix()),
			"aud": "valid-client-id",
			"iss": "https://accounts.google.com",
		}
		if err := provider.validateClaims(validClaimsHttps); err != nil {
			t.Errorf("expected valid claims for https issuer, got %v", err)
		}

		// Valid claims (with accounts.google.com)
		validClaimsPlain := jwt.MapClaims{
			"exp": float64(time.Now().Add(10 * time.Minute).Unix()),
			"aud": "valid-client-id",
			"iss": "accounts.google.com",
		}
		if err := provider.validateClaims(validClaimsPlain); err != nil {
			t.Errorf("expected valid claims for plain issuer, got %v", err)
		}
	})

	t.Run("full ValidateToken integration with httptest server", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("failed to generate RSA key: %v", err)
		}

		kid := "test-kid-123"
		nBytes := privateKey.N.Bytes()
		eBytes := big.NewInt(int64(privateKey.E)).Bytes()

		nBase64 := base64.RawURLEncoding.EncodeToString(nBytes)
		eBase64 := base64.RawURLEncoding.EncodeToString(eBytes)

		certsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(GoogleCerts{
				Keys: []struct {
					Kid string `json:"kid"`
					E   string `json:"e"`
					N   string `json:"n"`
				}{
					{Kid: kid, E: eBase64, N: nBase64},
				},
			})
		}))
		defer certsServer.Close()

		oldCertsURL := googleCertsURL
		googleCertsURL = certsServer.URL
		defer func() { googleCertsURL = oldCertsURL }()

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub":  "google-user-id-999",
			"name": "Jane Smith",
			"exp":  float64(time.Now().Add(time.Hour).Unix()),
			"aud":  "test-client-id",
			"iss":  "https://accounts.google.com",
		})
		token.Header["kid"] = kid
		idTokenStr, err := token.SignedString(privateKey)
		if err != nil {
			t.Fatalf("failed to sign ID token: %v", err)
		}

		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(googleTokenResponse{
				AccessToken: "google_access_token",
				ExpiresIn:   3600,
				TokenType:   "Bearer",
				IDToken:     idTokenStr,
			})
		}))
		defer tokenServer.Close()

		oldTokenURL := tokenURL
		tokenURL = tokenServer.URL
		defer func() { tokenURL = oldTokenURL }()

		provider, err := NewGoogleAuthProvider(GoogleAuthConfig{
			ClientID:      "test-client-id",
			ClientSecret:  "test-client-secret",
			CertsCacheTTL: 3600,
		})
		if err != nil {
			t.Fatalf("unexpected error creating provider: %v", err)
		}

		claims, err := provider.ValidateToken("valid_auth_code")
		if err != nil {
			t.Fatalf("expected ValidateToken success, got error: %v", err)
		}
		if claims.Sub != "google-user-id-999" || claims.DisplayName != "Jane Smith" {
			t.Errorf("unexpected claims payload: %+v", claims)
		}
	})

	t.Run("http server errors in token exchange and certs fetch", func(t *testing.T) {
		// Non-200 token exchange
		errTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer errTokenServer.Close()

		oldTokenURL := tokenURL
		tokenURL = errTokenServer.URL

		provider, _ := NewGoogleAuthProvider(GoogleAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		})

		_, err := provider.ValidateToken("code")
		if err == nil {
			t.Errorf("expected error for non-200 token response")
		}

		// Invalid JSON token response
		badJsonTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer badJsonTokenServer.Close()
		tokenURL = badJsonTokenServer.URL
		_, err = provider.ValidateToken("code")
		if err == nil {
			t.Errorf("expected error for invalid json token response")
		}

		// Empty id_token in token response
		emptyIdTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(googleTokenResponse{IDToken: ""})
		}))
		defer emptyIdTokenServer.Close()
		tokenURL = emptyIdTokenServer.URL
		_, err = provider.ValidateToken("code")
		if err == nil {
			t.Errorf("expected error for empty id_token")
		}

		// Non-200 certs fetch
		errCertsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errCertsServer.Close()

		oldCertsURL := googleCertsURL
		googleCertsURL = errCertsServer.URL

		err = provider.fetchGoogleCerts()
		if err == nil {
			t.Errorf("expected error for non-200 certs response")
		}

		// Bad JSON certs response
		badJsonCertsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not json"))
		}))
		defer badJsonCertsServer.Close()
		googleCertsURL = badJsonCertsServer.URL
		err = provider.fetchGoogleCerts()
		if err == nil {
			t.Errorf("expected error for non-json certs response")
		}

		// Invalid Base64 in N parameter
		badNBase64CertsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(GoogleCerts{
				Keys: []struct {
					Kid string `json:"kid"`
					E   string `json:"e"`
					N   string `json:"n"`
				}{
					{Kid: "k1", E: "AQAB", N: "invalid base64!"},
				},
			})
		}))
		defer badNBase64CertsServer.Close()
		googleCertsURL = badNBase64CertsServer.URL
		err = provider.fetchGoogleCerts()
		if err == nil {
			t.Errorf("expected error for invalid base64 N parameter")
		}

		// Invalid Base64 in E parameter
		badEBase64CertsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(GoogleCerts{
				Keys: []struct {
					Kid string `json:"kid"`
					E   string `json:"e"`
					N   string `json:"n"`
				}{
					{Kid: "k1", E: "invalid base64!", N: "AQAB"},
				},
			})
		}))
		defer badEBase64CertsServer.Close()
		googleCertsURL = badEBase64CertsServer.URL
		err = provider.fetchGoogleCerts()
		if err == nil {
			t.Errorf("expected error for invalid base64 E parameter")
		}

		googleCertsURL = oldCertsURL
		tokenURL = oldTokenURL
	})

	t.Run("getGoogleCert kid missing from fetched certs", func(t *testing.T) {
		emptyCertsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(GoogleCerts{Keys: nil})
		}))
		defer emptyCertsServer.Close()

		oldCertsURL := googleCertsURL
		googleCertsURL = emptyCertsServer.URL
		defer func() { googleCertsURL = oldCertsURL }()

		provider, _ := NewGoogleAuthProvider(GoogleAuthConfig{
			ClientID:      "id",
			ClientSecret:  "secret",
			CertsCacheTTL: 3600,
		})

		_, err := provider.getGoogleCert("non-existent-kid")
		if err == nil {
			t.Errorf("expected error for non-existent kid")
		}
	})

	t.Run("ValidateToken with invalid id_token from token exchange", func(t *testing.T) {
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(googleTokenResponse{IDToken: "invalid-id-token"})
		}))
		defer tokenServer.Close()

		oldTokenURL := tokenURL
		tokenURL = tokenServer.URL
		defer func() { tokenURL = oldTokenURL }()

		provider, _ := NewGoogleAuthProvider(GoogleAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		})

		_, err := provider.ValidateToken("code")
		if err == nil {
			t.Errorf("expected error when readClaims fails")
		}
	})

	t.Run("ValidateToken with aud claim mismatch", func(t *testing.T) {
		privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		kid := "kid-aud-mismatch"
		nBytes := privateKey.N.Bytes()
		eBytes := big.NewInt(int64(privateKey.E)).Bytes()

		certsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(GoogleCerts{
				Keys: []struct {
					Kid string `json:"kid"`
					E   string `json:"e"`
					N   string `json:"n"`
				}{
					{Kid: kid, E: base64.RawURLEncoding.EncodeToString(eBytes), N: base64.RawURLEncoding.EncodeToString(nBytes)},
				},
			})
		}))
		defer certsServer.Close()

		oldCertsURL := googleCertsURL
		googleCertsURL = certsServer.URL
		defer func() { googleCertsURL = oldCertsURL }()

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub":  "user-1",
			"name": "Jane",
			"exp":  float64(time.Now().Add(time.Hour).Unix()),
			"aud":  "WRONG-CLIENT-ID",
			"iss":  "https://accounts.google.com",
		})
		token.Header["kid"] = kid
		idTokenStr, _ := token.SignedString(privateKey)

		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(googleTokenResponse{IDToken: idTokenStr})
		}))
		defer tokenServer.Close()

		oldTokenURL := tokenURL
		tokenURL = tokenServer.URL
		defer func() { tokenURL = oldTokenURL }()

		provider, _ := NewGoogleAuthProvider(GoogleAuthConfig{
			ClientID:     "correct-client-id",
			ClientSecret: "secret",
		})

		_, err := provider.ValidateToken("code")
		if err == nil {
			t.Errorf("expected error when validateClaims fails")
		}
	})

	t.Run("readClaims direct error branches", func(t *testing.T) {
		provider, _ := NewGoogleAuthProvider(GoogleAuthConfig{ClientID: "id", ClientSecret: "secret"})

		// HMAC signing method
		tokenHS := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "123"})
		tokenHSStr, _ := tokenHS.SignedString([]byte("secret"))
		_, err := provider.readClaims(tokenHSStr)
		if err == nil {
			t.Errorf("expected error for non-RSA signing method")
		}

		// Missing kid header
		tokenRSNoKid := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "123"})
		privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		tokenRSNoKidStr, _ := tokenRSNoKid.SignedString(privateKey)
		_, err = provider.readClaims(tokenRSNoKidStr)
		if err == nil {
			t.Errorf("expected error for missing kid header")
		}
	})
}
