package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
)

type InvitationEndpoints struct {
	CreateInvitation   endpoint.Endpoint
	ListInvitations    endpoint.Endpoint
	DeleteInvitation   endpoint.Endpoint
	ValidateInvitation endpoint.Endpoint
}

func toInvitationDTO(inv *people.Invitation) InvitationDTO {
	if inv == nil {
		return InvitationDTO{}
	}
	return InvitationDTO{
		ID:          inv.ID().String(),
		Email:       inv.Email(),
		Status:      string(inv.Status()),
		ExpiresAt:   inv.ExpiresAt(),
		CreatedAt:   inv.CreatedAt(),
		CompletedAt: inv.CompletedAt(),
	}
}

func decodeCreateInvitationRequest(ctx context.Context, r *http.Request) (any, error) {
	adminID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("unauthorized", err)
	}

	var req CreateInvitationDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, exceptions.NewValidationException("invalid request body", err)
	}

	return people.CreateInvitationCommand{
		AdminID: adminID,
		Email:   req.Email,
	}, nil
}

func decodeListInvitationsRequest(ctx context.Context, _ *http.Request) (any, error) {
	adminID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("unauthorized", err)
	}
	return adminID, nil
}

func decodeDeleteInvitationRequest(ctx context.Context, r *http.Request) (any, error) {
	adminID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("unauthorized", err)
	}

	vars := mux.Vars(r)
	invitationIDStr := vars["invitation_id"]
	invitationID, err := core.ParseInvitationID(invitationIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid invitation id", err)
	}

	return people.DeleteInvitationCommand{
		AdminID:      adminID,
		InvitationID: invitationID,
	}, nil
}

func decodeValidateInvitationRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	invCode, err := people.NewInvitationCode(vars["code"])
	if err != nil {
		return nil, err
	}
	return invCode.String(), nil
}

func makeCreateInvitationEndpoint(svc people.InvitationService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(people.CreateInvitationCommand)

		invWithCode, err := svc.Create(ctx, req)
		if err != nil {
			return nil, err
		}

		return InvitationCreatedResponseDTO{
			ID:        invWithCode.Invitation.ID().String(),
			Email:     invWithCode.Invitation.Email(),
			Code:      invWithCode.Code,
			ExpiresAt: invWithCode.Invitation.ExpiresAt(),
			CreatedAt: invWithCode.Invitation.CreatedAt(),
			Status:    string(invWithCode.Invitation.Status()),
		}, nil
	}
}

func makeListInvitationsEndpoint(svc people.InvitationService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		adminID := request.(core.PersonID)

		invitations, err := svc.FindAllByCreator(ctx, adminID)
		if err != nil {
			return nil, err
		}

		dtos := make([]InvitationDTO, 0, len(invitations))
		for _, inv := range invitations {
			dtos = append(dtos, toInvitationDTO(inv))
		}

		return ListInvitationsResponse{
			Invitations: dtos,
		}, nil
	}
}

func makeDeleteInvitationEndpoint(svc people.InvitationService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(people.DeleteInvitationCommand)

		if err := svc.Delete(ctx, req); err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func makeValidateInvitationEndpoint(svc people.InvitationService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		code := request.(string)

		inv, err := svc.FindByCode(ctx, code)
		if err != nil {
			return nil, err
		}

		return ValidateInvitationResponseDTO{
			Email:     inv.Email(),
			ExpiresAt: inv.ExpiresAt(),
		}, nil
	}
}

func MakeInvitationEndpoints(svc people.InvitationService) InvitationEndpoints {
	return InvitationEndpoints{
		CreateInvitation:   makeCreateInvitationEndpoint(svc),
		ListInvitations:    makeListInvitationsEndpoint(svc),
		DeleteInvitation:   makeDeleteInvitationEndpoint(svc),
		ValidateInvitation: makeValidateInvitationEndpoint(svc),
	}
}
