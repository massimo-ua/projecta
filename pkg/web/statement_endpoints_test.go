package web

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type mockStatementService struct {
	res *projecta.StatementParseResult
	err error
}

func (m *mockStatementService) ParseStatement(_ context.Context, _ uuid.UUID, _ []byte) (*projecta.StatementParseResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.res != nil {
		return m.res, nil
	}
	return &projecta.StatementParseResult{
		Account:      "ACC123",
		Currency:     "PLN",
		Period:       "01.09.2026 - 30.09.2026",
		Transactions: []projecta.StatementParseItem{},
	}, nil
}

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
		svc := &mockStatementService{}
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
		svc := &mockStatementService{err: errors.New("svc error")}
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
