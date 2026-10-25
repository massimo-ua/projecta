package core

import (
	"errors"
	"github.com/google/uuid"
)

type AuthTokenPayload struct {
	Sub         string
	DisplayName string
	Roles       []string
	Email       string
	FirstName   string
	LastName    string
}

type AuthTokenClaims struct {
	AuthTokenPayload
	ID string
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	IssuedAt     int64
	ExpiresAt    int64
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

type ThirdPartyAuth interface {
	ValidateToken(token string) (*AuthTokenClaims, error)
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
