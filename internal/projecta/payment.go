package projecta

import (
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"time"
)

type PaymentKind string

const (
	DownPayment           PaymentKind = "DOWN_PAYMENT"
	UponCompletionPayment PaymentKind = "UPON_COMPLETION"
	CreditPayment         PaymentKind = "CREDIT_PAYMENT"
)

type Payment struct {
	ID          uuid.UUID
	Project     *Project
	Owner       *Owner
	Type        *CostType
	Description string
	Amount      *money.Money
	Date        time.Time
	Kind        PaymentKind
}

func ToPaymentKind(kind string) (PaymentKind, error) {
	switch true {
	case kind == DownPayment.String():
		return DownPayment, nil
	case kind == UponCompletionPayment.String():
		return UponCompletionPayment, nil
	case kind == CreditPayment.String():
		return CreditPayment, nil
	default:
		return "", exceptions.NewValidationException("invalid payment kind", nil)
	}
}

func (e PaymentKind) String() string {
	return string(e)
}

func NewPayment(
	id uuid.UUID,
	project *Project,
	person *Owner,
	costType *CostType,
	description string,
	amount *money.Money,
	date time.Time,
	kind PaymentKind) *Payment {
	return &Payment{
		ID:          id,
		Project:     project,
		Owner:       person,
		Type:        costType,
		Description: description,
		Amount:      amount,
		Date:        date,
		Kind:        kind,
	}
}

type PaymentCollection = core.PaginatedCollection[*Payment]

func NewPaymentCollection(total int) *PaymentCollection {
	return core.NewPaginatedCollection[*Payment](total)
}
