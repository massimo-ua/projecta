package websocket

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
	requestAuthFailedError = "request authorization failed"
)

type authorizer = func(aToken string) (uuid.UUID, error)

func createJwtAuthorizer(tokenProvider core.AuthTokenProvider) authorizer {
	return func(aToken string) (uuid.UUID, error) {
		if aToken == "" {
			return uuid.Nil, exceptions.NewUnauthorizedException(requestAuthFailedError, nil)
		}

		claims, err := tokenProvider.ValidateToken(aToken)
		if err != nil {
			return uuid.Nil, exceptions.NewUnauthorizedException(requestAuthFailedError, err)
		}

		personUUID, err := uuid.Parse(claims.AuthTokenPayload.Sub)

		if err != nil {
			return uuid.Nil, exceptions.NewUnauthorizedException(requestAuthFailedError, err)
		}

		return personUUID, nil
	}
}

func createAuthorizedContext(requesterID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), core.RequesterIDContextKey, requesterID)
}
