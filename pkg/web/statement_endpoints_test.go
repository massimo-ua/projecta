package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

func TestDecodeParseStatementRequest(t *testing.T) {
	projID := uuid.New()

	t.Run("success", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "statement.pdf")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		_, _ = part.Write([]byte("dummy pdf content"))
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projID.String()+"/payments/parse-statement", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req = mux.SetURLVars(req, map[string]string{"project_id": projID.String()})

		result, err := decodeParseStatementRequest(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parseReq, ok := result.(ParseStatementRequest)
		if !ok {
			t.Fatalf("expected ParseStatementRequest, got %T", result)
		}
		if parseReq.ProjectID != projID {
			t.Errorf("expected project ID %v, got %v", projID, parseReq.ProjectID)
		}
		if string(parseReq.FileData) != "dummy pdf content" {
			t.Errorf("unexpected file content: %s", string(parseReq.FileData))
		}
	})

	t.Run("missing project_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/projects//payments/parse-statement", nil)
		_, err := decodeParseStatementRequest(context.Background(), req)
		if err == nil {
			t.Errorf("expected error for missing project_id")
		}
	})

	t.Run("invalid project_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/projects/invalid/payments/parse-statement", nil)
		req = mux.SetURLVars(req, map[string]string{"project_id": "invalid"})
		_, err := decodeParseStatementRequest(context.Background(), req)
		if err == nil {
			t.Errorf("expected error for invalid project_id")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projID.String()+"/payments/parse-statement", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req = mux.SetURLVars(req, map[string]string{"project_id": projID.String()})

		_, err := decodeParseStatementRequest(context.Background(), req)
		if err == nil {
			t.Errorf("expected error for missing file")
		}
	})
}

func TestMakeParseStatementEndpoint(t *testing.T) {
	projID := uuid.New()

	t.Run("success", func(t *testing.T) {
		svc := &mockPaymentService{}
		ep := makeParseStatementEndpoint(svc)

		res, err := ep(context.Background(), ParseStatementRequest{
			ProjectID: projID,
			FileData:  []byte("pdf"),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		stmtRes, ok := res.(*projecta.StatementParseResult)
		if !ok || stmtRes.Account != "ACC123" {
			t.Errorf("unexpected result: %v", res)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockPaymentService{err: errors.New("svc error")}
		ep := makeParseStatementEndpoint(svc)

		_, err := ep(context.Background(), ParseStatementRequest{
			ProjectID: projID,
			FileData:  []byte("pdf"),
		})
		if err == nil {
			t.Errorf("expected error")
		}
	})
}

func TestDecodeCreateBatchPaymentRequest(t *testing.T) {
	projID := uuid.New()
	typeID := uuid.New()
	now := time.Now().Format(time.RFC3339)

	t.Run("success array format", func(t *testing.T) {
		items := []CreatePaymentDTO{
			{
				TypeID:      typeID.String(),
				Description: "Item 1",
				Amount:      1000,
				Currency:    "PLN",
				PaymentDate: now,
			},
		}
		data, _ := json.Marshal(items)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projID.String()+"/payments/batch", bytes.NewReader(data))
		req = mux.SetURLVars(req, map[string]string{"project_id": projID.String()})

		res, err := decodeCreateBatchPaymentRequest(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cmds, ok := res.([]projecta.CreatePaymentCommand)
		if !ok || len(cmds) != 1 {
			t.Fatalf("expected 1 command, got %v", res)
		}
		if cmds[0].Description != "Item 1" || cmds[0].Amount.Amount() != 1000 {
			t.Errorf("mismatch in command fields: %+v", cmds[0])
		}
	})

	t.Run("success wrapped format", func(t *testing.T) {
		wrapper := BatchCreatePaymentDTO{
			Payments: []CreatePaymentDTO{
				{
					TypeID:      typeID.String(),
					Description: "Wrapped Item",
					Amount:      2500,
					Currency:    "PLN",
					PaymentDate: now,
					Kind:        "DOWN_PAYMENT",
				},
			},
		}
		data, _ := json.Marshal(wrapper)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projID.String()+"/payments/batch", bytes.NewReader(data))
		req = mux.SetURLVars(req, map[string]string{"project_id": projID.String()})

		res, err := decodeCreateBatchPaymentRequest(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cmds, ok := res.([]projecta.CreatePaymentCommand)
		if !ok || len(cmds) != 1 {
			t.Fatalf("expected 1 command, got %v", res)
		}
		if cmds[0].Kind != projecta.DownPayment {
			t.Errorf("expected DOWN_PAYMENT, got %v", cmds[0].Kind)
		}
	})

	t.Run("invalid project_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/projects/invalid/payments/batch", bytes.NewReader([]byte("[]")))
		req = mux.SetURLVars(req, map[string]string{"project_id": "invalid"})

		_, err := decodeCreateBatchPaymentRequest(context.Background(), req)
		if err == nil {
			t.Errorf("expected error for invalid project_id")
		}
	})

	t.Run("invalid type_id", func(t *testing.T) {
		items := []CreatePaymentDTO{
			{
				TypeID:      "invalid-type",
				Description: "Item",
				Amount:      1000,
				Currency:    "PLN",
				PaymentDate: now,
			},
		}
		data, _ := json.Marshal(items)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projID.String()+"/payments/batch", bytes.NewReader(data))
		req = mux.SetURLVars(req, map[string]string{"project_id": projID.String()})

		_, err := decodeCreateBatchPaymentRequest(context.Background(), req)
		if err == nil {
			t.Errorf("expected error for invalid type_id")
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		items := []CreatePaymentDTO{
			{
				TypeID:      typeID.String(),
				Description: "Item",
				Amount:      1000,
				Currency:    "PLN",
				PaymentDate: "not-a-date",
			},
		}
		data, _ := json.Marshal(items)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projID.String()+"/payments/batch", bytes.NewReader(data))
		req = mux.SetURLVars(req, map[string]string{"project_id": projID.String()})

		_, err := decodeCreateBatchPaymentRequest(context.Background(), req)
		if err == nil {
			t.Errorf("expected error for invalid date")
		}
	})
}

func TestMakeCreateBatchPaymentsEndpoint(t *testing.T) {
	projID := uuid.New()
	cat, _ := projecta.NewCostCategory(uuid.New(), projID, "Cat", "")
	costType, _ := projecta.NewCostType(projID, cat, "Type", "")
	owner := &projecta.Owner{PersonID: uuid.New(), FirstName: "A", DisplayName: "B"}
	proj, _ := projecta.NewProject(projID, "Proj", "", owner, time.Now(), time.Now())
	pay := projecta.NewPayment(uuid.New(), proj, owner, costType, "Pay", money.New(500, money.USD), time.Now(), projecta.UponCompletionPayment)

	rateProvider := &mockRateProvider{}

	t.Run("success", func(t *testing.T) {
		svc := &mockPaymentService{pay: pay}
		ep := makeCreateBatchPaymentsEndpoint(svc, rateProvider)

		res, err := ep(context.Background(), []projecta.CreatePaymentCommand{{
			ProjectID: projID,
		}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dtos, ok := res.([]PaymentDTO)
		if !ok || len(dtos) != 1 {
			t.Fatalf("expected 1 DTO, got %v", res)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockPaymentService{err: errors.New("batch failed")}
		ep := makeCreateBatchPaymentsEndpoint(svc, rateProvider)

		_, err := ep(context.Background(), []projecta.CreatePaymentCommand{{}})
		if err == nil {
			t.Errorf("expected error")
		}
	})
}
