package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
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

func makeGetPaymentEndpoint(svc projecta.PaymentService, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.PaymentFilter)

		p, err := svc.FindOne(ctx, filter)

		if err != nil {
			return nil, err
		}

		return toPaymentDTO(p, rateProvider), nil
	}
}

type ParseStatementRequest struct {
	ProjectID uuid.UUID
	FileData  []byte
}

func decodeParseStatementRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	r.Body = http.MaxBytesReader(nil, r.Body, 10<<20)

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, exceptions.NewValidationException("file is required in multipart form field 'file'", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, exceptions.NewValidationException("failed to read uploaded file", err)
	}

	return ParseStatementRequest{
		ProjectID: projectUUID,
		FileData:  fileBytes,
	}, nil
}

func makeParseStatementEndpoint(svc projecta.PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ParseStatementRequest)
		return svc.ParseStatement(ctx, req.ProjectID, req.FileData)
	}
}

type BatchCreatePaymentDTO struct {
	Payments []CreatePaymentDTO `json:"payments"`
}

func decodeCreateBatchPaymentRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, exceptions.NewValidationException("failed to read body", err)
	}

	var dtos []CreatePaymentDTO
	if err := json.Unmarshal(bodyBytes, &dtos); err != nil {
		var wrapped BatchCreatePaymentDTO
		if errWrap := json.Unmarshal(bodyBytes, &wrapped); errWrap != nil {
			return nil, exceptions.NewValidationException("invalid batch request format", errWrap)
		}
		dtos = wrapped.Payments
	}

	commands := make([]projecta.CreatePaymentCommand, 0, len(dtos))
	for _, req := range dtos {
		typeUUID, err := uuid.Parse(req.TypeID)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid type id", err)
		}

		pID := projectUUID
		if req.ProjectID != "" {
			if parsedPID, err := uuid.Parse(req.ProjectID); err == nil {
				pID = parsedPID
			}
		}

		amount := money.New(req.Amount, req.Currency)
		date, err := time.Parse(time.RFC3339, req.PaymentDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid date", err)
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

		commands = append(commands, projecta.CreatePaymentCommand{
			ProjectID:   pID,
			TypeID:      typeUUID,
			Description: req.Description,
			Amount:      amount,
			PaymentDate: date,
			Kind:        paymentKind,
		})
	}

	return commands, nil
}

func makeCreateBatchPaymentsEndpoint(svc projecta.PaymentService, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		commands := request.([]projecta.CreatePaymentCommand)
		payments, err := svc.CreateBatch(ctx, commands)
		if err != nil {
			return nil, err
		}

		dtos := make([]PaymentDTO, 0, len(payments))
		for _, p := range payments {
			dtos = append(dtos, toPaymentDTO(p, rateProvider))
		}
		return dtos, nil
	}
}
