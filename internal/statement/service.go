package statement

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
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
	FindExisting(ctx context.Context, projectID core.ProjectID) ([]ExistingTransaction, error)
}

type ProjectValidator interface {
	ValidateProject(ctx context.Context, projectID core.ProjectID) error
}

type StatementParseItem struct {
	DocNum      string
	Date        time.Time
	Amount      int64 // in minor units
	Currency    string
	Description string
	IsDuplicate bool
}

type StatementParseResult struct {
	Account      string
	Currency     string
	Period       string
	Transactions []StatementParseItem
}

type Service interface {
	ParseStatement(ctx context.Context, projectID core.ProjectID, fileData []byte) (*StatementParseResult, error)
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

func (s *ServiceImpl) ParseStatement(ctx context.Context, projectID core.ProjectID, fileData []byte) (*StatementParseResult, error) {
	if projectID.IsNil() {
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
