package web

import (
	"context"
	"encoding/json"
	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"net/http"
	"time"
)

type UpdatePaymentDTO struct {
	ProjectID   string `json:"project_id"`
	TypeID      string `json:"type_id"`
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	PaymentDate string `json:"payment_date"`
	Kind        string `json:"kind,omitempty"`
}

func decodeUpdatePaymentRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	paymentID, ok := vars["payment_id"]

	if !ok {
		return nil, exceptions.NewValidationException("invalid payment id", nil)
	}

	paymentUUID, err := uuid.Parse(paymentID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid payment id", err)
	}

	var req UpdatePaymentDTO
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	amount := money.New(req.Amount, req.Currency)

	date, err := time.Parse(time.RFC3339, req.PaymentDate)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid date", err)
	}

	typeUUID, err := uuid.Parse(req.TypeID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid type id", err)
	}

	var paymentKind projecta.PaymentKind

	if req.Kind == "" {
		paymentKind = projecta.UponCompletionPayment
	} else {
		paymentKind, err = projecta.ToPaymentKind(req.Kind)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid payment kind", err)
		}
	}

	return projecta.UpdatePaymentCommand{
		ID:          paymentUUID,
		ProjectID:   projectUUID,
		TypeID:      typeUUID,
		Description: req.Description,
		Amount:      amount,
		PaymentDate: date,
		Kind:        paymentKind,
	}, err
}

func makeUpdatePaymentEndpoint(svc projecta.PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.UpdatePaymentCommand)

		err := svc.Update(ctx, command)

		return nil, err
	}
}

func decodeGetPaymentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	paymentID, ok := vars["payment_id"]

	if !ok {
		return nil, exceptions.NewValidationException("invalid payment id", nil)
	}

	paymentUUID, err := uuid.Parse(paymentID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid payment id", err)
	}

	return projecta.PaymentFilter{
		ProjectID: projectUUID,
		PaymentID: paymentUUID,
	}, nil
}

func makeGetPaymentEndpoint(svc projecta.PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.PaymentFilter)

		p, err := svc.FindOne(ctx, filter)

		if err != nil {
			return nil, err
		}

		return PaymentDTO{
			PaymentID: p.ID.String(),
			Project: ProjectDTO{
				ProjectID:   p.Project.ProjectID.String(),
				Name:        p.Project.Name,
				Description: p.Project.Description,
				Owner: OwnerDTO{
					PersonID:    p.Owner.PersonID.String(),
					DisplayName: p.Owner.DisplayName,
				},
			},
			Owner: OwnerDTO{
				PersonID:    p.Owner.PersonID.String(),
				DisplayName: p.Owner.DisplayName,
			},
			Type: TypeDTO{
				TypeID:      p.Type.ID.String(),
				Name:        p.Type.Name,
				Description: p.Type.Description,
			},
			Category: CategoryDTO{
				CategoryID:  p.Type.Category.ID.String(),
				Name:        p.Type.Category.Name,
				Description: p.Type.Category.Description,
			},
			Description: p.Description,
			Amount:      p.Amount.Amount(),
			Currency:    p.Amount.Currency().Code,
			PaymentDate: p.Date.Format(time.RFC3339),
		}, nil
	}
}
