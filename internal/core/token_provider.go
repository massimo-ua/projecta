package core

import (
	"errors"
	"github.com/google/uuid"
)

type AuthTokenPayload struct {
	Sub         string   `json:"sub"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
}

type AuthTokenClaims struct {
	AuthTokenPayload
	ID string `json:"jti"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IssuedAt     int64  `json:"issued_at"`
	ExpiresAt    int64  `json:"expires_at"`
}

var AuthTokenIsExpired = errors.New("auth token is expired")
var AuthTokenIsInvalid = errors.New("auth token is invalid")
var RefreshTokenIsInvalid = errors.New("refresh token is invalid")
var AuthTokenGenerationFailed = errors.New("auth token generation failed")

type AuthTokenProvider interface {
	GenerateTokenRing(data AuthTokenPayload) (*AuthResponse, error)
	ValidateToken(token string) (*AuthTokenClaims, error)
	DecodeToken(token string) (*AuthTokenClaims, error)
	ValidateRefreshToken(tokenID uuid.UUID, refreshToken string) bool
}

type TokenRing struct {
	accessToken  string
	refreshToken string
}

func NewTokenRing(accessToken string, refreshToken string) (*TokenRing, error) {
	if accessToken == "" || refreshToken == "" {
		return nil, RefreshTokenIsInvalid
	}

	return &TokenRing{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

func (t *TokenRing) AccessToken() string {
	return t.accessToken
}

func (t *TokenRing) RefreshToken() string {
	return t.refreshToken
}
