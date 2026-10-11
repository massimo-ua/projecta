package web

import (
	"context"
	"io"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

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

func makeParseStatementEndpoint(svc projecta.StatementService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ParseStatementRequest)
		return svc.ParseStatement(ctx, req.ProjectID, req.FileData)
	}
}
