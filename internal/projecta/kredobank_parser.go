package projecta

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dslipak/pdf"
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

var (
	currencyRegex = regexp.MustCompile(`Валюта\s*рахунку\s*([A-Z]{3})`)
	accountRegex  = regexp.MustCompile(`Рахунок\s*([A-Z0-9]+)`)
	periodRegex   = regexp.MustCompile(`Період\s*(.*?)(?:Сформовано|\n|\r|$)`)
	dateRegex     = regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)
)

// ParseKredobankStatement parses a Kredobank account statement PDF file and returns
// statement metadata along with all expense (negative) transactions converted to positive amounts.
func ParseKredobankStatement(data []byte) (*ParsedStatement, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}

	var fullTextBuf bytes.Buffer
	b, err := reader.GetPlainText()
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF plain text: %w", err)
	}
	_, _ = fullTextBuf.ReadFrom(b)
	fullText := fullTextBuf.String()

	if !strings.Contains(fullText, "Кредобанк") && !strings.Contains(fullText, "Рух коштів за рахунком") {
		return nil, fmt.Errorf("not a recognized Kredobank statement")
	}

	currMatches := currencyRegex.FindStringSubmatch(fullText)
	currency := "PLN"
	if len(currMatches) > 1 {
		currency = currMatches[1]
	}

	accMatches := accountRegex.FindStringSubmatch(fullText)
	account := ""
	if len(accMatches) > 1 {
		account = accMatches[1]
	}

	periodMatches := periodRegex.FindStringSubmatch(fullText)
	period := ""
	if len(periodMatches) > 1 {
		period = strings.TrimSpace(periodMatches[1])
	}

	var allTransactions []ParsedStatementTransaction

	type txAnchor struct {
		date      time.Time
		docNum    string
		rawAmount float64
		y         float64
	}

	numPages := reader.NumPage()
	for p := 1; p <= numPages; p++ {
		page := reader.Page(p)
		rows, err := page.GetTextByRow()
		if err != nil {
			continue
		}

		var tableHeaderY float64 = 1000.0
		var tableFooterY float64 = 0.0

		for _, row := range rows {
			var line strings.Builder
			for _, w := range row.Content {
				line.WriteString(w.S)
				line.WriteString(" ")
			}
			rowStr := line.String()
			if (strings.Contains(rowStr, "Дата") && strings.Contains(rowStr, "Деталі")) ||
				(strings.Contains(rowStr, "Номер") && strings.Contains(rowStr, "Сума")) ||
				(strings.Contains(rowStr, "документа") && strings.Contains(rowStr, "рахунку")) {
				y := float64(row.Position)
				if y < tableHeaderY {
					tableHeaderY = y
				}
			}
			if strings.Contains(rowStr, "Вклади гарантуються") || strings.Contains(rowStr, "Керівник") {
				y := float64(row.Position)
				if y > tableFooterY {
					tableFooterY = y
				}
			}
		}

		var anchors []txAnchor
		for _, row := range rows {
			rowY := float64(row.Position)
			if rowY >= tableHeaderY || rowY <= tableFooterY {
				continue
			}

			var dateStr, docStr, amtStr string
			for _, w := range row.Content {
				trimmed := strings.TrimSpace(w.S)
				if w.X < 90 && dateRegex.MatchString(trimmed) {
					dateStr = trimmed
				} else if w.X >= 90 && w.X < 160 && trimmed != "" {
					docStr = trimmed
				} else if w.X >= 450 && trimmed != "" {
					amtStr = trimmed
				}
			}

			if dateStr != "" && docStr != "" && amtStr != "" {
				parsedDate, errDate := time.Parse("02.01.2006", dateStr)
				cleanAmt := strings.ReplaceAll(amtStr, ",", "")
				cleanAmt = strings.ReplaceAll(cleanAmt, " ", "")
				parsedAmt, errAmt := strconv.ParseFloat(cleanAmt, 64)
				if errDate == nil && errAmt == nil {
					anchors = append(anchors, txAnchor{
						date:      parsedDate,
						docNum:    docStr,
						rawAmount: parsedAmt,
						y:         rowY,
					})
				}
			}
		}

		for i, anchor := range anchors {
			if anchor.rawAmount >= 0 {
				// Skip positive (credit / income) transactions
				continue
			}

			upperY := tableHeaderY
			if i > 0 {
				upperY = (anchor.y + anchors[i-1].y) / 2.0
			}

			lowerY := tableFooterY
			if i+1 < len(anchors) {
				lowerY = (anchor.y + anchors[i+1].y) / 2.0
			}

			var descWords []string
			for _, row := range rows {
				rowPos := float64(row.Position)
				if rowPos < upperY && rowPos >= lowerY {
					for _, w := range row.Content {
						if w.X >= 160 && w.X < 450 {
							descWords = append(descWords, w.S)
						}
					}
				}
			}

			fullDesc := strings.TrimSpace(strings.Join(descWords, " "))
			fullDesc = strings.Join(strings.Fields(fullDesc), " ")

			descWithDoc := fmt.Sprintf("[Doc #%s] %s", anchor.docNum, fullDesc)
			positiveAmt := -anchor.rawAmount
			minorUnits := int64(positiveAmt*100 + 0.5)

			allTransactions = append(allTransactions, ParsedStatementTransaction{
				Date:        anchor.date,
				DocNum:      anchor.docNum,
				Description: descWithDoc,
				AmountMinor: minorUnits,
				AmountFloat: positiveAmt,
				Currency:    currency,
			})
		}
	}

	return &ParsedStatement{
		Account:      account,
		Currency:     currency,
		Period:       period,
		Transactions: allTransactions,
	}, nil
}
