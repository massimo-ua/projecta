package statement

import (
	"fmt"
	"strings"
	"time"
)

func transactionFingerprint(date time.Time, amount int64, description string) string {
	return fmt.Sprintf("%s_%d_%s", date.Format("2006-01-02"), amount, strings.TrimSpace(description))
}

type ParsedStatementTransaction struct {
	Date        time.Time
	DocNum      string
	Description string
	AmountMinor int64
	AmountFloat float64
	Currency    string
}

func (t ParsedStatementTransaction) Fingerprint() string {
	return transactionFingerprint(t.Date, t.AmountMinor, t.Description)
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
