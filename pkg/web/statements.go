package web

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/statement"
)

type StatementParseItemDTO struct {
	DocNum      string    `json:"doc_num"`
	Date        time.Time `json:"date"`
	Amount      int64     `json:"amount"` // in minor units
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	IsDuplicate bool      `json:"is_duplicate"`
}

type StatementParseResultDTO struct {
	Account      string                  `json:"account"`
	Currency     string                  `json:"currency"`
	Period       string                  `json:"period"`
	Transactions []StatementParseItemDTO `json:"transactions"`
}

func toStatementParseResultDTO(res *statement.StatementParseResult) StatementParseResultDTO {
	if res == nil {
		return StatementParseResultDTO{}
	}
	items := make([]StatementParseItemDTO, 0, len(res.Transactions))
	for _, item := range res.Transactions {
		items = append(items, StatementParseItemDTO{
			DocNum:      item.DocNum,
			Date:        item.Date,
			Amount:      item.Amount,
			Currency:    item.Currency,
			Description: item.Description,
			IsDuplicate: item.IsDuplicate,
		})
	}
	return StatementParseResultDTO{
		Account:      res.Account,
		Currency:     res.Currency,
		Period:       res.Period,
		Transactions: items,
	}
}

type StatementEndpoints struct {
	ParseStatement endpoint.Endpoint
}

func MakeStatementEndpoints(svc statement.Service) StatementEndpoints {
	if svc == nil {
		return StatementEndpoints{}
	}
	return StatementEndpoints{
		ParseStatement: makeParseStatementEndpoint(svc),
	}
}

type ParseStatementRequest struct {
	ProjectID core.ProjectID
	FileData  []byte
}

func decodeParseStatementRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
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

func makeParseStatementEndpoint(svc statement.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ParseStatementRequest)
		res, err := svc.ParseStatement(ctx, req.ProjectID, req.FileData)
		if err != nil {
			return nil, err
		}
		return toStatementParseResultDTO(res), nil
	}
}
