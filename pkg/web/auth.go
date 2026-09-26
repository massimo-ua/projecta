package web

import (
    "context"
    "github.com/go-kit/kit/endpoint"
    ht "github.com/go-kit/kit/transport/http"
    "github.com/google/uuid"
    "gitlab.com/massimo-ua/projecta/internal/core"
    "gitlab.com/massimo-ua/projecta/internal/exceptions"
    "gitlab.com/massimo-ua/projecta/internal/people"
    "net/http"
    "strings"
)

func loggedInOnly(next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, request any) (any, error) {
        if _, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID); !ok {
            return nil, exceptions.NewUnauthorizedException("Request authorization failed", nil)
        }

        return next(ctx, request)
    }
}

func jwtMiddleware(tokenProvider core.AuthTokenProvider) ht.RequestFunc {
    return func(ctx context.Context, r *http.Request) context.Context {
        aHeader := r.Header.Get("Authorization")
        if aHeader == "" {
            return ctx
        }

        headerParts := strings.Split(aHeader, " ")
        if len(headerParts) != 2 {
            return ctx
        }

        token := headerParts[1]
        claims, err := tokenProvider.ValidateToken(token)
        if err != nil {
            return ctx
        }

        personUUID, err := uuid.Parse(claims.AuthTokenPayload.Sub)

        if err != nil {
            return ctx
        }

        roles := claims.AuthTokenPayload.Roles
        if roles == nil {
            // Backward compatibility: legacy tokens issued before RBAC introduction are granted RoleUser
            roles = []string{string(people.RoleUser)}
        }

        ctx = context.WithValue(ctx, core.RequesterIDContextKey, personUUID)
        ctx = context.WithValue(ctx, core.RequesterRolesContextKey, roles)
        return ctx
    }
}

func requireRole(role string, next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, request any) (any, error) {
        if _, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID); !ok {
            return nil, exceptions.NewUnauthorizedException("Request authorization failed", nil)
        }

        if !core.HasRole(ctx, role) {
            return nil, exceptions.NewForbiddenException("access forbidden: insufficient permissions", nil)
        }

        return next(ctx, request)
    }
}
