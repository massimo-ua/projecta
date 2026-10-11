package projecta

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ExistingTransaction struct {
	Date        time.Time
	Amount      int64
	Description string
}

type ExistingTransactionFinder interface {
	FindExisting(ctx context.Context, projectID uuid.UUID) ([]ExistingTransaction, error)
}

type StatementService interface {
	ParseStatement(ctx context.Context, projectID uuid.UUID, fileData []byte) (*StatementParseResult, error)
}

type StatementServiceImpl struct {
	projects ProjectService
	finder   ExistingTransactionFinder
}

func NewStatementService(projects ProjectService, finder ...ExistingTransactionFinder) *StatementServiceImpl {
	var f ExistingTransactionFinder
	if len(finder) > 0 {
		f = finder[0]
	}
	return &StatementServiceImpl{
		projects: projects,
		finder:   f,
	}
}

func (s *StatementServiceImpl) ParseStatement(ctx context.Context, projectID uuid.UUID, fileData []byte) (*StatementParseResult, error) {
	if projectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	_, err := s.projects.FindOne(ctx, ProjectFilter{ProjectID: projectID})
	if err != nil {
		return nil, exceptions.NewValidationException("project not found", err)
	}

	statement, err := ParseKredobankStatement(fileData)
	if err != nil {
		return nil, exceptions.NewValidationException("failed to parse statement: "+err.Error(), err)
	}

	existingMap := make(map[string]bool)
	if s.finder != nil {
		existing, err := s.finder.FindExisting(ctx, projectID)
		if err == nil && existing != nil {
			for _, it := range existing {
				key := fmt.Sprintf("%s_%d_%s", it.Date.Format("2006-01-02"), it.Amount, strings.TrimSpace(it.Description))
				existingMap[key] = true
			}
		}
	}

	items := make([]StatementParseItem, 0, len(statement.Transactions))
	for _, tx := range statement.Transactions {
		key := fmt.Sprintf("%s_%d_%s", tx.Date.Format("2006-01-02"), tx.AmountMinor, strings.TrimSpace(tx.Description))
		isDup := existingMap[key]
		items = append(items, StatementParseItem{
			DocNum:      tx.DocNum,
			Date:        tx.Date,
			Amount:      tx.AmountMinor,
			Currency:    tx.Currency,
			Description: tx.Description,
			IsDuplicate: isDup,
		})
	}

	return &StatementParseResult{
		Account:      statement.Account,
		Currency:     statement.Currency,
		Period:       statement.Period,
		Transactions: items,
	}, nil
}
