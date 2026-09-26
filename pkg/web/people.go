package web

import (
	"context"
	"encoding/json"
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
	CustomerID  string   `json:"customer_id"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
}

func toUserDTO(p *people.Person) UserDTO {
	if p == nil {
		return UserDTO{}
	}
	return UserDTO{
		CustomerID:  p.ID().String(),
		FirstName:   p.FirstName(),
		LastName:    p.LastName(),
		DisplayName: p.DisplayName(),
		Roles:       p.RoleStrings(),
	}
}

type UpdateProfileDTO struct {
	DisplayName people.DisplayName `json:"display_name"`
}

type UpdateProfileRequest struct {
	PersonID    uuid.UUID
	DisplayName people.DisplayName
}

type UserEndpoints struct {
	Register      endpoint.Endpoint
	Login         endpoint.Endpoint
	RefreshToken  endpoint.Endpoint
	Profile       endpoint.Endpoint
	UpdateProfile endpoint.Endpoint
	ListUsers     endpoint.Endpoint
	AssignRoles   endpoint.Endpoint
}

func decodeProfileRequest(ctx context.Context, _ *http.Request) (any, error) {
	requesterID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if !ok {
		return nil, exceptions.NewUnauthorizedException("failed to authorize profile request", nil)
	}

	return requesterID, nil
}

func decodeUpdateProfileRequest(ctx context.Context, r *http.Request) (any, error) {
	requesterID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if !ok {
		return nil, exceptions.NewUnauthorizedException("failed to authorize profile request", nil)
	}

	var dto UpdateProfileDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return nil, exceptions.NewValidationException("invalid request body", err)
	}

	return UpdateProfileRequest{
		PersonID:    requesterID,
		DisplayName: dto.DisplayName,
	}, nil
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
		if req.InvitationCode != "" {
			provider, err := people.ToIdentityProvider(req.IdentityProvider)
			if err != nil {
				return nil, exceptions.NewValidationException("unknown identity provider", err)
			}
			return svc.LoginWithInvitation(ctx, req.Token, provider, req.InvitationCode)
		}

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

		return toUserDTO(person), nil
	}
}

func makeUpdateProfileEndpoint(svc people.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(UpdateProfileRequest)

		person, err := svc.UpdateDisplayName(ctx, people.UpdateDisplayNameCommand{
			PersonID:    req.PersonID,
			DisplayName: req.DisplayName,
		})

		if err != nil {
			return nil, err
		}

		return toUserDTO(person), nil
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

func makeListUsersEndpoint(svc people.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		pagination := request.(core.Pagination)

		users, total, err := svc.FindAll(ctx, pagination)
		if err != nil {
			return nil, err
		}

		dtos := make([]UserDTO, 0, len(users))
		for _, u := range users {
			dtos = append(dtos, toUserDTO(u))
		}

		return ListUsersResponse{
			Users: dtos,
			PaginationDTO: PaginationDTO{
				Limit:  pagination.Limit,
				Offset: pagination.Offset,
				Total:  total,
			},
		}, nil
	}
}

func makeAssignRolesEndpoint(svc people.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(AssignRolesRequest)

		roles := make([]people.Role, 0, len(req.Roles))
		for _, rStr := range req.Roles {
			r, err := people.ToRole(rStr)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid role", err)
			}
			roles = append(roles, r)
		}

		err := svc.AssignRoles(ctx, people.AssignRolesCommand{
			PersonID: req.UserID,
			Roles:    roles,
		})
		if err != nil {
			return nil, err
		}

		updatedPerson, err := svc.FindByID(ctx, req.UserID)
		if err != nil {
			return nil, err
		}

		return toUserDTO(updatedPerson), nil
	}
}

func MakeCustomerEndpoints(s people.UserService, a people.AuthService) (UserEndpoints, error) {
	return UserEndpoints{
		Register:      makeRegisterEndpoint(s),
		Login:         makeLoginEndpoint(a),
		RefreshToken:  makeRefreshTokenEndpoint(a),
		Profile:       makeProfileEndpoint(s),
		UpdateProfile: makeUpdateProfileEndpoint(s),
		ListUsers:     makeListUsersEndpoint(s),
		AssignRoles:   makeAssignRolesEndpoint(s),
	}, nil
}
