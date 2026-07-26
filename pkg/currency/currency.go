package currency

import "fmt"

type Currency struct {
	Amount int64  `json:"amount"` // Amount in smallest monetary units (e.g. cents, kopiyok)
	Code   string `json:"code"`   // Currency ISO code e.g. "UAH", "USD", "EUR", "PLN"
}

func NewCurrency(amount int64, code string) Currency {
	return Currency{
		Amount: amount,
		Code:   code,
	}
}

func (c Currency) String() string {
	return fmt.Sprintf("%d %s", c.Amount, c.Code)
}

type CurrencyRateProvider interface {
	Convert(currencyA Currency, currencyB Currency) (Currency, error)
}
