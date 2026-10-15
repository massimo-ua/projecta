package asset

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type AssetStatus string

const (
	AssetStatusActive    AssetStatus = "ACTIVE"
	AssetStatusCompleted AssetStatus = "COMPLETED"
)

func (s AssetStatus) String() string {
	return string(s)
}

func ToAssetStatus(status string) AssetStatus {
	if status == string(AssetStatusCompleted) {
		return AssetStatusCompleted
	}
	return AssetStatusActive
}

type ChildAssetLink struct {
	ChildID         uuid.UUID    `json:"child_id"`
	ChildName       string       `json:"child_name"`
	SharePercentage float64      `json:"share_percentage"`
	TotalCost       *money.Money `json:"total_cost,omitempty"`
}

type ParentAssetLink struct {
	ParentID        uuid.UUID `json:"parent_id"`
	ParentName      string    `json:"parent_name"`
	SharePercentage float64   `json:"share_percentage"`
}

type OwnerRef struct {
	PersonID    uuid.UUID `json:"person_id"`
	DisplayName string    `json:"display_name"`
}

type Asset struct {
	id            uuid.UUID
	name          string
	description   string
	projectID     uuid.UUID
	price         *money.Money
	acquiredAt    time.Time
	owner         OwnerRef
	status        AssetStatus
	startDate     time.Time
	completedDate *time.Time
	targetPrice   *money.Money
	directCost    *money.Money
	totalCost     *money.Money
	children      []ChildAssetLink
	parents       []ParentAssetLink
	tags          []string
}

func NewAsset(
	id uuid.UUID,
	name string,
	description string,
	projectID uuid.UUID,
	price *money.Money,
	acquiredAt time.Time,
	owner OwnerRef,
) *Asset {
	start := acquiredAt
	if start.IsZero() {
		start = time.Now()
	}
	return &Asset{
		id:          id,
		name:        name,
		description: description,
		projectID:   projectID,
		price:       price,
		acquiredAt:  start,
		owner:       owner,
		status:      AssetStatusActive,
		startDate:   start,
		targetPrice: price,
		children:    make([]ChildAssetLink, 0),
		parents:     make([]ParentAssetLink, 0),
		tags:        make([]string, 0),
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

func (a *Asset) ProjectID() uuid.UUID {
	return a.projectID
}

func (a *Asset) Price() *money.Money {
	if a.totalCost != nil {
		return a.totalCost
	}
	if a.targetPrice != nil {
		return a.targetPrice
	}
	return a.price
}

func (a *Asset) AcquiredAt() time.Time {
	if !a.startDate.IsZero() {
		return a.startDate
	}
	return a.acquiredAt
}

func (a *Asset) Owner() OwnerRef {
	return a.owner
}

func (a *Asset) Status() AssetStatus {
	if a.status == "" {
		return AssetStatusActive
	}
	return a.status
}

func (a *Asset) StartDate() time.Time {
	if !a.startDate.IsZero() {
		return a.startDate
	}
	return a.acquiredAt
}

func (a *Asset) CompletedDate() *time.Time {
	return a.completedDate
}

func (a *Asset) TargetPrice() *money.Money {
	return a.targetPrice
}

func (a *Asset) DirectCost() *money.Money {
	return a.directCost
}

func (a *Asset) TotalCost() *money.Money {
	if a.totalCost != nil {
		return a.totalCost
	}
	if a.directCost != nil {
		return a.directCost
	}
	return a.price
}

func (a *Asset) Children() []ChildAssetLink {
	return a.children
}

func (a *Asset) Parents() []ParentAssetLink {
	return a.parents
}

func (a *Asset) Tags() []string {
	if a.tags == nil {
		return make([]string, 0)
	}
	return a.tags
}

func (a *Asset) SetName(name string) {
	a.name = name
}

func (a *Asset) SetDescription(description string) {
	a.description = description
}

func (a *Asset) SetProjectID(projectID uuid.UUID) {
	a.projectID = projectID
}

func (a *Asset) SetPrice(price *money.Money) {
	a.price = price
	a.targetPrice = price
}

func (a *Asset) SetAcquiredAt(acquiredAt time.Time) {
	a.acquiredAt = acquiredAt
	a.startDate = acquiredAt
}

func (a *Asset) SetOwner(owner OwnerRef) {
	a.owner = owner
}

func (a *Asset) SetStatus(status AssetStatus) {
	a.status = status
}

func (a *Asset) SetStartDate(startDate time.Time) {
	a.startDate = startDate
}

func (a *Asset) SetCompletedDate(completedDate *time.Time) {
	a.completedDate = completedDate
}

func (a *Asset) SetTargetPrice(targetPrice *money.Money) {
	a.targetPrice = targetPrice
}

func (a *Asset) SetDirectCost(directCost *money.Money) {
	a.directCost = directCost
}

func (a *Asset) SetTotalCost(totalCost *money.Money) {
	a.totalCost = totalCost
}

func (a *Asset) SetChildren(children []ChildAssetLink) {
	a.children = children
}

func (a *Asset) SetParents(parents []ParentAssetLink) {
	a.parents = parents
}

func (a *Asset) SetTags(tags []string) {
	if tags == nil {
		a.tags = make([]string, 0)
	} else {
		a.tags = tags
	}
}

// AssignValuation assigns computed financial metrics and composition links to the asset.
func (a *Asset) AssignValuation(directCost, totalCost *money.Money, children []ChildAssetLink, parents []ParentAssetLink) {
	a.directCost = directCost
	a.totalCost = totalCost
	if children == nil {
		a.children = make([]ChildAssetLink, 0)
	} else {
		a.children = children
	}
	if parents == nil {
		a.parents = make([]ParentAssetLink, 0)
	} else {
		a.parents = parents
	}
}

// UpdateDetails updates the descriptive information, target budget, and tags of the asset.
func (a *Asset) UpdateDetails(name, description string, targetPrice *money.Money, tags []string) {
	if name != "" {
		a.name = name
	}
	a.description = description
	a.targetPrice = targetPrice
	if tags != nil {
		a.tags = tags
	}
}

// Complete transitions the asset to COMPLETED status with a completion timestamp.
func (a *Asset) Complete(completedDate time.Time) {
	a.status = AssetStatusCompleted
	a.completedDate = &completedDate
}

// Reopen transitions a completed asset back to ACTIVE status.
func (a *Asset) Reopen() {
	a.status = AssetStatusActive
	a.completedDate = nil
}

// ReconstituteAsset reconstructs an Asset entity from persistence data without executing business invariants.
func ReconstituteAsset(
	id uuid.UUID,
	name string,
	description string,
	projectID uuid.UUID,
	price *money.Money,
	acquiredAt time.Time,
	owner OwnerRef,
	status AssetStatus,
	startDate time.Time,
	completedDate *time.Time,
	targetPrice *money.Money,
	tags []string,
) *Asset {
	start := startDate
	if start.IsZero() {
		start = acquiredAt
	}
	if start.IsZero() {
		start = time.Now()
	}
	if tags == nil {
		tags = make([]string, 0)
	}
	if status == "" {
		status = AssetStatusActive
	}
	return &Asset{
		id:            id,
		name:          name,
		description:   description,
		projectID:     projectID,
		price:         price,
		acquiredAt:    start,
		owner:         owner,
		status:        status,
		startDate:     start,
		completedDate: completedDate,
		targetPrice:   targetPrice,
		children:      make([]ChildAssetLink, 0),
		parents:       make([]ParentAssetLink, 0),
		tags:          tags,
	}
}

type Collection = core.PaginatedCollection[*Asset]

func NewCollection(total int) *Collection {
	return core.NewPaginatedCollection[*Asset](total)
}
