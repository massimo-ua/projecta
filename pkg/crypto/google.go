package crypto

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"
	tokenURL       = "https://oauth2.googleapis.com/token"
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

	return &core.AuthTokenClaims{
		ID: uuid.New().String(),
		AuthTokenPayload: core.AuthTokenPayload{
			Sub:         claims["sub"].(string),
			DisplayName: claims["name"].(string),
			Roles:       []string{},
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

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", exceptions.NewInternalException("failed to exchange code", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", exceptions.NewInternalException(
			fmt.Sprintf("token exchange failed with status: %d", resp.StatusCode),
			nil,
		)
	}

	var tokenResp googleTokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", exceptions.NewInternalException("failed to decode token response", err)
	}

	if tokenResp.IDToken == "" {
		return "", exceptions.NewInternalException("no id_token in response", nil)
	}

	return tokenResp.IDToken, nil
}

func (p *GoogleAuthProvider) fetchGoogleCerts() error {
	resp, err := http.Get(googleCertsURL)
	if err != nil {
		return exceptions.NewInternalException("failed to fetch Google certs", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return exceptions.NewInternalException("failed to fetch Google certs: non-200 status", nil)
	}

	var certs GoogleCerts
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return exceptions.NewInternalException("failed to decode Google certs", err)
	}

	p.cache.Lock()
	defer p.cache.Unlock()

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

		p.cache.Certs[key.Kid] = pubKey
	}

	p.cache.Exp = time.Now().Add(time.Second * time.Duration(p.cacheTTL))
	return nil
}

func (p *GoogleAuthProvider) getGoogleCert(kid string) (*rsa.PublicKey, error) {
	if p.cache.Exp == (time.Time{}) || time.Now().After(p.cache.Exp) {
		if err := p.fetchGoogleCerts(); err != nil {
			return nil, err
		}
	}

	p.cache.RLock()
	pubKey, ok := p.cache.Certs[kid]
	p.cache.RUnlock()

	if !ok {
		return nil, exceptions.NewInternalException("public key not found", nil)
	}

	return pubKey, nil
}

func (p *GoogleAuthProvider) readClaims(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, exceptions.NewInternalException("unexpected signing method", nil)
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, exceptions.NewInternalException("kid header not found", nil)
		}

		return p.getGoogleCert(kid)
	})
	return claims, err
}

func (p *GoogleAuthProvider) validateClaims(claims jwt.MapClaims) error {
	now := time.Now().UTC()

	if claims["exp"] != nil && now.Unix() > int64(claims["exp"].(float64)) {
		return core.AuthTokenIsExpired
	}

	if aud, ok := claims["aud"].(string); !ok || aud != p.clientID {
		return exceptions.NewInternalException("invalid audience claim", nil)
	}

	if iss, ok := claims["iss"].(string); !ok ||
		(iss != fmt.Sprintf("https://%s", acceptedIssuer) && iss != acceptedIssuer) {
		return exceptions.NewInternalException("invalid issuer claim", nil)
	}

	return nil
}
