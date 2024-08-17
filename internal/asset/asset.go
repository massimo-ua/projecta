package asset

import (
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"time"
)

type Asset struct {
	ID          uuid.UUID
	Name        string
	Description string
	Project     *projecta.Project
	Type        *projecta.CostType
	Price       *money.Money
	AcquiredAt  time.Time
	Owner       *projecta.Owner
}

func NewAsset(
	id uuid.UUID,
	name string,
	description string,
	project *projecta.Project,
	costType *projecta.CostType,
	price *money.Money,
	acquiredAt time.Time,
	owner *projecta.Owner) *Asset {
	return &Asset{
		ID:          id,
		Name:        name,
		Description: description,
		Project:     project,
		Type:        costType,
		Price:       price,
		AcquiredAt:  acquiredAt,
		Owner:       owner,
	}
}

type Collection = core.PaginatedCollection[*Asset]

func NewCollection(total int) *Collection {
	return core.NewPaginatedCollection[*Asset](total)
}
