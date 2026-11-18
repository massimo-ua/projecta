package crypto

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

var (
	googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"
	tokenURL       = "https://oauth2.googleapis.com/token"
)

const (
	acceptedIssuer = "accounts.google.com"
	// redirectURI is magic value for Google OAuth2 token exchange
	// https://github.com/MomenSherif/react-oauth/issues/12#issuecomment-1231202620
	redirectURI = "postmessage"
)

type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type GoogleCerts struct {
	Keys []struct {
		Kid string `json:"kid"`
		E   string `json:"e"`
		N   string `json:"n"`
	} `json:"keys"`
}

type CertsCache struct {
	sync.RWMutex
	Certs map[string]*rsa.PublicKey
	Exp   time.Time
}

type GoogleAuthConfig struct {
	ClientID      string
	ClientSecret  string
	CertsCacheTTL int
}

type GoogleAuthProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
	cache        CertsCache
	cacheTTL     int
}

func NewGoogleAuthProvider(config GoogleAuthConfig) (*GoogleAuthProvider, error) {
	if config.ClientID == "" {
		return nil, exceptions.NewInternalException("google auth provider requires client id", nil)
	}
	if config.ClientSecret == "" {
		return nil, exceptions.NewInternalException("google auth provider requires client secret", nil)
	}

	return &GoogleAuthProvider{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		redirectURI:  redirectURI,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: CertsCache{
			Certs: make(map[string]*rsa.PublicKey),
		},
		cacheTTL: config.CertsCacheTTL,
	}, nil
}

func (p *GoogleAuthProvider) ValidateToken(code string) (*core.AuthTokenClaims, error) {
	idToken, err := p.exchangeCodeForToken(code)
	if err != nil {
		return nil, err
	}

	claims, err := p.readClaims(idToken)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to read auth claims", err)
	}

	if err = p.validateClaims(claims); err != nil {
		return nil, err
	}

	var sub string
	if s, ok := claims["sub"].(string); ok {
		sub = s
	}
	var displayName string
	if n, ok := claims["name"].(string); ok {
		displayName = n
	}
	var email string
	if verified, ok := claims["email_verified"].(bool); ok && verified {
		if e, ok := claims["email"].(string); ok {
			email = e
		}
	}
	var firstName string
	if fn, ok := claims["given_name"].(string); ok {
		firstName = fn
	}
	var lastName string
	if ln, ok := claims["family_name"].(string); ok {
		lastName = ln
	}

	return &core.AuthTokenClaims{
		ID: uuid.New().String(),
		AuthTokenPayload: core.AuthTokenPayload{
			Sub:         sub,
			DisplayName: displayName,
			Roles:       []string{},
			Email:       email,
			FirstName:   firstName,
			LastName:    lastName,
		},
	}, nil
}

func (p *GoogleAuthProvider) exchangeCodeForToken(code string) (string, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {p.redirectURI},
	}

	resp, err := p.httpClient.PostForm(tokenURL, data)
	if err != nil {
		return "", exceptions.NewInternalException("failed to exchange code", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", exceptions.NewInternalException(
			fmt.Sprintf("token exchange failed with status %d", resp.StatusCode),
			nil,
		)
	}

	var tokenResp googleTokenResponse
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tokenResp); err != nil {
		return "", exceptions.NewInternalException("failed to decode token response", err)
	}

	if tokenResp.IDToken == "" {
		return "", exceptions.NewInternalException("no id_token in response", nil)
	}

	return tokenResp.IDToken, nil
}

func (p *GoogleAuthProvider) fetchGoogleCerts() error {
	resp, err := p.httpClient.Get(googleCertsURL)
	if err != nil {
		return exceptions.NewInternalException("failed to fetch Google certs", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return exceptions.NewInternalException("failed to fetch Google certs: non-200 status", nil)
	}

	var certs GoogleCerts
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&certs); err != nil {
		return exceptions.NewInternalException("failed to decode Google certs", err)
	}

	newCerts := make(map[string]*rsa.PublicKey, len(certs.Keys))
	for _, key := range certs.Keys {
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return exceptions.NewInternalException("failed to decode cert N parameter", err)
		}

		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return exceptions.NewInternalException("failed to decode cert E parameter", err)
		}

		e := big.NewInt(0).SetBytes(eBytes).Int64()
		pubKey := &rsa.PublicKey{
			N: big.NewInt(0).SetBytes(nBytes),
			E: int(e),
		}

		newCerts[key.Kid] = pubKey
	}

	p.cache.Lock()
	p.cache.Certs = newCerts
	p.cache.Exp = time.Now().Add(time.Second * time.Duration(p.cacheTTL))
	p.cache.Unlock()

	return nil
}

func (p *GoogleAuthProvider) getGoogleCert(kid string) (*rsa.PublicKey, error) {
	p.cache.RLock()
	expired := p.cache.Exp.IsZero() || time.Now().After(p.cache.Exp)
	pubKey, ok := p.cache.Certs[kid]
	p.cache.RUnlock()

	if expired {
		if err := p.fetchGoogleCerts(); err != nil {
			return nil, err
		}
		p.cache.RLock()
		pubKey, ok = p.cache.Certs[kid]
		p.cache.RUnlock()
	}

	if !ok {
		return nil, exceptions.NewInternalException("public key not found", nil)
	}

	return pubKey, nil
}

func (p *GoogleAuthProvider) readClaims(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, exceptions.NewInternalException("unexpected signing method", nil)
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, exceptions.NewInternalException("kid header not found", nil)
			}

			return p.getGoogleCert(kid)
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodRS256.Alg(),
			jwt.SigningMethodRS384.Alg(),
			jwt.SigningMethodRS512.Alg(),
		}),
		jwt.WithExpirationRequired(),
	)
	return claims, err
}

func (p *GoogleAuthProvider) validateClaims(claims jwt.MapClaims) error {
	now := time.Now().UTC()

	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return exceptions.NewInternalException("invalid expiration claim", err)
	}
	if now.After(exp.Time) {
		return core.AuthTokenIsExpired
	}

	if aud, ok := claims["aud"].(string); !ok || aud != p.clientID {
		return exceptions.NewInternalException("invalid audience claim", nil)
	}

	if iss, ok := claims["iss"].(string); !ok ||
		(iss != fmt.Sprintf("https://%s", acceptedIssuer) && iss != acceptedIssuer) {
		return exceptions.NewInternalException("invalid issuer claim", nil)
	}

	if ev, exists := claims["email_verified"]; exists {
		if verified, ok := ev.(bool); !ok || !verified {
			return exceptions.NewInternalException("email is not verified", nil)
		}
	}

	return nil
}
