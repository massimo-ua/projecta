package statement_test

import (
	_ "embed"
	"testing"

	"gitlab.com/massimo-ua/projecta/internal/statement"
)

//go:embed testdata/fake_kredobank_report.pdf
var fakeReportPDF []byte

func TestParseKredobankStatement_Success(t *testing.T) {
	if len(fakeReportPDF) == 0 {
		t.Fatalf("embedded fake report PDF is empty")
	}

	parser := statement.NewKredobankParser()
	parsed, err := parser.Parse(fakeReportPDF)
	if err != nil {
		t.Fatalf("unexpected error parsing statement: %v", err)
	}

	if parsed.Currency != "PLN" {
		t.Errorf("expected currency PLN, got %s", parsed.Currency)
	}

	if parsed.Account != "UA000000000000000000000000000" {
		t.Errorf("expected account UA000000000000000000000000000, got %s", parsed.Account)
	}

	if len(parsed.Transactions) != 26 {
		t.Fatalf("expected 26 expense transactions, got %d", len(parsed.Transactions))
	}

	// First transaction verification
	firstTx := parsed.Transactions[0]
	if firstTx.DocNum != "43539945" {
		t.Errorf("expected first doc num 43539945, got %s", firstTx.DocNum)
	}
	if firstTx.AmountMinor != 920 {
		t.Errorf("expected first amount 920, got %d", firstTx.AmountMinor)
	}
	if firstTx.AmountFloat != 9.20 {
		t.Errorf("expected first float amount 9.20, got %f", firstTx.AmountFloat)
	}
	if firstTx.Currency != "PLN" {
		t.Errorf("expected first currency PLN, got %s", firstTx.Currency)
	}
	if firstTx.Date.Format("2006-01-02") != "2026-09-01" {
		t.Errorf("expected first date 2026-09-01, got %s", firstTx.Date.Format("2006-01-02"))
	}

	// Verify all amounts are positive
	for i, tx := range parsed.Transactions {
		if tx.AmountMinor <= 0 || tx.AmountFloat <= 0 {
			t.Errorf("transaction %d has non-positive amount: %f (%d)", i, tx.AmountFloat, tx.AmountMinor)
		}
		if tx.Currency != "PLN" {
			t.Errorf("transaction %d currency mismatch: %s", i, tx.Currency)
		}
		if tx.DocNum == "" {
			t.Errorf("transaction %d has empty doc num", i)
		}
	}
}

func TestParseKredobankStatement_InvalidPDF(t *testing.T) {
	parser := statement.NewKredobankParser()
	_, err := parser.Parse([]byte("invalid pdf content"))
	if err == nil {
		t.Errorf("expected error for invalid PDF content, got nil")
	}
}

func TestParseKredobankStatement_NonKredobank(t *testing.T) {
	header := "%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Count 1/Kids[3 0 R]>>endobj\n3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]>>endobj\nxref\n0 4\n0000000000 65535 f\n0000000009 00000 n\n0000000052 00000 n\n0000000108 00000 n\ntrailer<</Size 4/Root 1 0 R>>\nstartxref\n180\n%%EOF"
	parser := statement.NewKredobankParser()
	_, err := parser.Parse([]byte(header))
	if err == nil {
		t.Errorf("expected error for non-Kredobank PDF, got nil")
	}
}
