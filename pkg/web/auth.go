package web

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	ht "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"net/http"
	"strings"
)

const authHeaderPrefix = "Bearer "

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

		if !strings.HasPrefix(aHeader, authHeaderPrefix) {
			return ctx
		}

		token := strings.TrimPrefix(aHeader, authHeaderPrefix)

		if token == "" {
			return ctx
		}

		claims, err := tokenProvider.ValidateToken(token)
		if err != nil {
			return ctx
		}

		personUUID, err := uuid.Parse(claims.AuthTokenPayload.Sub)

		if err != nil {
			return ctx
		}

		return context.WithValue(ctx, core.RequesterIDContextKey, personUUID)
	}
}
