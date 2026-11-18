package crypto

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"time"
)

type JwtTokenProvider struct {
	secret string
	ttl    int
	hasher core.Hasher
}

func NewJwtTokenProvider(secret string, ttl int, hasher core.Hasher) *JwtTokenProvider {
	return &JwtTokenProvider{
		secret: secret,
		ttl:    ttl,
		hasher: hasher,
	}
}

func (p *JwtTokenProvider) GenerateTokenRing(data core.AuthTokenPayload) (*core.AuthResponse, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(time.Second * time.Duration(p.ttl))
	token := jwt.New(jwt.SigningMethodHS256)
	tokenID := uuid.New().String()
	claims := token.Claims.(jwt.MapClaims)
	claims["jti"] = tokenID
	claims["sub"] = data.Sub
	claims["display_name"] = data.DisplayName
	claims["iat"] = now.Unix()
	claims["exp"] = expiresAt.Unix()

	if data.Roles != nil {
		claims["roles"] = data.Roles
	}

	accessToken, err := token.SignedString([]byte(p.secret))

	if err != nil {
		return nil, errors.Join(core.AuthTokenGenerationFailed, err)
	}

	refreshToken, err := p.hasher.Hash(tokenID)

	if err != nil {
		return nil, errors.Join(core.AuthTokenGenerationFailed, err)
	}

	return &core.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IssuedAt:     now.Unix(),
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func (p *JwtTokenProvider) ValidateToken(token string) (*core.AuthTokenClaims, error) {
	claims, err := p.readClaims(token)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.Join(core.AuthTokenIsInvalid, core.AuthTokenIsExpired, err)
		}
		return nil, errors.Join(core.AuthTokenIsInvalid, err)
	}

	return p.toDomain(claims)
}

func (p *JwtTokenProvider) DecodeToken(token string) (*core.AuthTokenClaims, error) {
	claims, err := p.readClaims(token)

	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return nil, errors.Join(core.AuthTokenIsInvalid, err)
	}

	return p.toDomain(claims)
}

func (p *JwtTokenProvider) toDomain(claims jwt.MapClaims) (*core.AuthTokenClaims, error) {
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return nil, core.AuthTokenIsInvalid
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return nil, core.AuthTokenIsInvalid
	}

	var displayName string
	if rawDN, exists := claims["display_name"]; exists && rawDN != nil {
		dn, ok := rawDN.(string)
		if !ok {
			return nil, core.AuthTokenIsInvalid
		}
		displayName = dn
	}

	var roles []string
	if rawRoles, exists := claims["roles"]; exists && rawRoles != nil {
		roles = []string{}
		switch r := rawRoles.(type) {
		case []string:
			roles = append(roles, r...)
		case []interface{}:
			for _, item := range r {
				s, ok := item.(string)
				if !ok {
					return nil, core.AuthTokenIsInvalid
				}
				roles = append(roles, s)
			}
		default:
			return nil, core.AuthTokenIsInvalid
		}
	}

	var issuedAt int64
	if iat, err := claims.GetIssuedAt(); err == nil && iat != nil {
		issuedAt = iat.Unix()
	}

	var expiresAt int64
	if exp, err := claims.GetExpirationTime(); err == nil && exp != nil {
		expiresAt = exp.Unix()
	}

	return &core.AuthTokenClaims{
		ID:        jti,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		AuthTokenPayload: core.AuthTokenPayload{
			Sub:         sub,
			DisplayName: displayName,
			Roles:       roles,
		},
	}, nil
}

func (p *JwtTokenProvider) readClaims(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(p.secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	return claims, err
}

func (p *JwtTokenProvider) ValidateRefreshToken(tokenID uuid.UUID, refreshToken string) bool {
	return p.hasher.Compare(tokenID.String(), refreshToken)
}
