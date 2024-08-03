package crypto

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"math/big"
	"net/http"
	"sync"
	"time"
)

const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

type CertsCache struct {
	sync.RWMutex
	Certs map[string]*rsa.PublicKey
	Exp   time.Time
}

type GoogleCerts struct {
	Keys []struct {
		Kid string `json:"kid"`
		E   string `json:"e"`
		N   string `json:"n"`
	} `json:"keys"`
}

type GoogleAuthProvider struct {
	clientID string
	cache    CertsCache
	cacheTTL int
}

func NewGoogleAuthProvider(clientID string, certsCacheTTL int) *GoogleAuthProvider {
	if clientID == "" {
		panic("google auth provider requires client id and secret")
	}

	if certsCacheTTL == 0 {
		certsCacheTTL = 24 * 60 * 60
	}

	return &GoogleAuthProvider{
		clientID: clientID,
		cache: CertsCache{
			Certs: make(map[string]*rsa.PublicKey),
		},
		cacheTTL: certsCacheTTL,
	}
}

func (p *GoogleAuthProvider) fetchGoogleCerts() error {
	resp, err := http.Get(googleCertsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to fetch Google certs")
	}

	var certs GoogleCerts
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return err
	}

	p.cache.Lock()
	defer p.cache.Unlock()

	for _, key := range certs.Keys {
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return err
		}

		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return err
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

	if ok {
		return pubKey, nil
	}

	return nil, errors.New("public key not found")
}

func (p *GoogleAuthProvider) ValidateToken(token string) (*core.AuthTokenClaims, error) {
	claims, err := p.readClaims(token)

	if err != nil {
		return nil, errors.Join(errors.New("failed to read auth claims"), err)
	}

	now := time.Now().UTC()
	if claims["exp"] != nil && now.Unix() > int64(claims["exp"].(float64)) {
		return nil, core.AuthTokenIsExpired
	}

	if claims["aud"] != p.clientID {
		return nil, errors.Join(errors.New("failed to verify audience"), err)
	}

	return &core.AuthTokenClaims{
		ID: claims["jti"].(string),
		AuthTokenPayload: core.AuthTokenPayload{
			Sub:         claims["sub"].(string),
			DisplayName: claims["name"].(string),
			Roles:       []string{},
		},
	}, nil
}

func (p *GoogleAuthProvider) readClaims(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}

		return p.getGoogleCert(kid)
	})
	return claims, err
}
