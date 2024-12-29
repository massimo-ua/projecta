package web

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"net/http"
)

type RegisterUserDTO struct {
	Login            string `json:"login"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	IdentityProvider string `json:"identity_provider"`
	Token            string `json:"token"`
}

type UserDTO struct {
	CustomerID  string `json:"customer_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
}

type UserEndpoints struct {
	Register     endpoint.Endpoint
	Login        endpoint.Endpoint
	RefreshToken endpoint.Endpoint
	Profile      endpoint.Endpoint
}

func decodeProfileRequest(ctx context.Context, _ *http.Request) (any, error) {
	requesterID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if !ok {
		return nil, exceptions.NewUnauthorizedException("failed to authorize profile request", nil)
	}

	return requesterID, nil
}

func makeRegisterEndpoint(svc people.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		dto := request.(RegisterUserDTO)

		identityProvider, err := people.ToIdentityProvider(dto.IdentityProvider)

		if err != nil {
			return nil, exceptions.NewValidationException("unknown identity provider", err)
		}

		err = svc.Register(ctx, people.RegisterCommand{
			Login:            dto.Login,
			FirstName:        dto.FirstName,
			LastName:         dto.LastName,
			IdentityProvider: identityProvider,
			Token:            dto.Token,
		})

		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func makeLoginEndpoint(svc people.AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(LoginDTO)
		c, err := people.NewCredentials(req.IdentityProvider, req.ID, req.Token)

		if err != nil {
			return nil, exceptions.NewValidationException("failed to identify customer", err)
		}

		response, err := svc.Login(ctx, c)

		if err != nil {
			return nil, err
		}

		return response, nil
	}
}

func makeProfileEndpoint(svc people.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		personID := request.(uuid.UUID)

		person, err := svc.FindByID(ctx, personID)

		if err != nil {
			return nil, err
		}

		return UserDTO{
			CustomerID:  person.ID().String(),
			FirstName:   person.FirstName(),
			LastName:    person.LastName(),
			DisplayName: person.DisplayName(),
		}, nil
	}
}

func makeRefreshTokenEndpoint(svc people.AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(RefreshTokenDTO)
		tokenRing, err := core.NewTokenRing(req.AccessToken, req.RefreshToken)

		if err != nil {
			return nil, exceptions.NewValidationException("failed to refresh token", err)
		}

		response, err := svc.Refresh(ctx, tokenRing)

		if err != nil {
			return nil, err
		}

		return response, nil
	}
}

func MakeCustomerEndpoints(s people.UserService, a people.AuthService) (UserEndpoints, error) {
	return UserEndpoints{
		Register:     makeRegisterEndpoint(s),
		Login:        makeLoginEndpoint(a),
		RefreshToken: makeRefreshTokenEndpoint(a),
		Profile:      makeProfileEndpoint(s),
	}, nil
}
