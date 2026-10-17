package statement

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ExistingTransaction struct {
	Date        time.Time
	Amount      int64
	Description string
}

func (t ExistingTransaction) Fingerprint() string {
	return transactionFingerprint(t.Date, t.Amount, t.Description)
}

type ExistingTransactionFinder interface {
	FindExisting(ctx context.Context, projectID uuid.UUID) ([]ExistingTransaction, error)
}

type ProjectValidator interface {
	ValidateProject(ctx context.Context, projectID uuid.UUID) error
}

type StatementParseItem struct {
	DocNum      string    `json:"doc_num"`
	Date        time.Time `json:"date"`
	Amount      int64     `json:"amount"` // in minor units
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	IsDuplicate bool      `json:"is_duplicate"`
}

type StatementParseResult struct {
	Account      string               `json:"account"`
	Currency     string               `json:"currency"`
	Period       string               `json:"period"`
	Transactions []StatementParseItem `json:"transactions"`
}

type Service interface {
	ParseStatement(ctx context.Context, projectID uuid.UUID, fileData []byte) (*StatementParseResult, error)
}

type ServiceImpl struct {
	projects ProjectValidator
	finder   ExistingTransactionFinder
	parser   StatementParser
}

func NewService(projects ProjectValidator, finder ExistingTransactionFinder, parser ...StatementParser) *ServiceImpl {
	var p StatementParser = NewKredobankParser()
	if len(parser) > 0 && parser[0] != nil {
		p = parser[0]
	}
	return &ServiceImpl{
		projects: projects,
		finder:   finder,
		parser:   p,
	}
}

func (s *ServiceImpl) ParseStatement(ctx context.Context, projectID uuid.UUID, fileData []byte) (*StatementParseResult, error) {
	if projectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	if s.projects != nil {
		if err := s.projects.ValidateProject(ctx, projectID); err != nil {
			return nil, exceptions.NewValidationException("project not found", err)
		}
	}

	statement, err := s.parser.Parse(fileData)
	if err != nil {
		return nil, exceptions.NewValidationException("failed to parse statement: "+err.Error(), err)
	}

	existingMap := make(map[string]bool)
	if s.finder != nil {
		existing, err := s.finder.FindExisting(ctx, projectID)
		if err == nil && existing != nil {
			for _, it := range existing {
				existingMap[it.Fingerprint()] = true
			}
		}
	}

	items := make([]StatementParseItem, 0, len(statement.Transactions))
	for _, tx := range statement.Transactions {
		isDup := existingMap[tx.Fingerprint()]
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
