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
    claims := jwt.MapClaims{}
    _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(p.secret), nil
    })

    if err != nil {
        return nil, errors.Join(core.AuthTokenIsInvalid, err)
    }

    now := time.Now().UTC()
    if claims["exp"] != nil && now.Unix() > int64(claims["exp"].(float64)) {
        return nil, core.AuthTokenIsExpired
    }

    roles := []string{}
    if claims["roles"] != nil {
        roles = claims["roles"].([]string)
    }

    return &core.AuthTokenClaims{
        ID: claims["jti"].(string),
        AuthTokenPayload: core.AuthTokenPayload{
            Sub:         claims["sub"].(string),
            DisplayName: claims["display_name"].(string),
            Roles:       roles,
        },
    }, nil
}

func (p *JwtTokenProvider) ValidateRefreshToken(tokenID uuid.UUID, refreshToken string) bool {
    return p.hasher.Compare(tokenID.String(), refreshToken)
}
