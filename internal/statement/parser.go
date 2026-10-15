package statement

import (
	"time"
)

type ParsedStatementTransaction struct {
	Date        time.Time
	DocNum      string
	Description string
	AmountMinor int64
	AmountFloat float64
	Currency    string
}

type ParsedStatement struct {
	Account      string
	Currency     string
	Period       string
	Transactions []ParsedStatementTransaction
}

// StatementParser defines a polymorphic contract for extracting structured transaction data from banking statements.
type StatementParser interface {
	Parse(data []byte) (*ParsedStatement, error)
}
