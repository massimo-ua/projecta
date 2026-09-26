package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
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

type DeleteInvitationRequest struct {
	AdminID      uuid.UUID
	InvitationID uuid.UUID
}

func decodeCreateInvitationRequest(ctx context.Context, r *http.Request) (any, error) {
	adminID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)
	if !ok || adminID == uuid.Nil {
		return nil, exceptions.NewUnauthorizedException("unauthorized", nil)
	}

	var req CreateInvitationDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, exceptions.NewValidationException("invalid request body", err)
	}

	return struct {
		AdminID uuid.UUID
		Email   string
	}{
		AdminID: adminID,
		Email:   req.Email,
	}, nil
}

func decodeListInvitationsRequest(ctx context.Context, _ *http.Request) (any, error) {
	adminID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)
	if !ok || adminID == uuid.Nil {
		return nil, exceptions.NewUnauthorizedException("unauthorized", nil)
	}
	return adminID, nil
}

func decodeDeleteInvitationRequest(ctx context.Context, r *http.Request) (any, error) {
	adminID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)
	if !ok || adminID == uuid.Nil {
		return nil, exceptions.NewUnauthorizedException("unauthorized", nil)
	}

	vars := mux.Vars(r)
	invitationIDStr := vars["invitation_id"]
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid invitation id", err)
	}

	return DeleteInvitationRequest{
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
		req := request.(struct {
			AdminID uuid.UUID
			Email   string
		})

		invWithCode, err := svc.Create(ctx, req.AdminID, req.Email)
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
		adminID := request.(uuid.UUID)

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
		req := request.(DeleteInvitationRequest)

		if err := svc.Delete(ctx, req.AdminID, req.InvitationID); err != nil {
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
