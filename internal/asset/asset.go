package asset

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type Asset struct {
	id          uuid.UUID
	name        string
	description string
	project     *projecta.Project
	costType    *projecta.CostType
	price       *money.Money
	acquiredAt  time.Time
	owner       *projecta.Owner
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
		id:          id,
		name:        name,
		description: description,
		project:     project,
		costType:    costType,
		price:       price,
		acquiredAt:  acquiredAt,
		owner:       owner,
	}
}

func (a *Asset) ID() uuid.UUID {
	return a.id
}

func (a *Asset) Name() string {
	return a.name
}

func (a *Asset) Description() string {
	return a.description
}

func (a *Asset) Project() *projecta.Project {
	return a.project
}

func (a *Asset) Type() *projecta.CostType {
	return a.costType
}

func (a *Asset) Price() *money.Money {
	return a.price
}

func (a *Asset) AcquiredAt() time.Time {
	return a.acquiredAt
}

func (a *Asset) Owner() *projecta.Owner {
	return a.owner
}

func (a *Asset) SetName(name string) {
	a.name = name
}

func (a *Asset) SetDescription(description string) {
	a.description = description
}

func (a *Asset) SetProject(project *projecta.Project) {
	a.project = project
}

func (a *Asset) SetType(costType *projecta.CostType) {
	a.costType = costType
}

func (a *Asset) SetPrice(price *money.Money) {
	a.price = price
}

func (a *Asset) SetAcquiredAt(acquiredAt time.Time) {
	a.acquiredAt = acquiredAt
}

func (a *Asset) SetOwner(owner *projecta.Owner) {
	a.owner = owner
}

type Collection = core.PaginatedCollection[*Asset]

func NewCollection(total int) *Collection {
	return core.NewPaginatedCollection[*Asset](total)
}
