package asset

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
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
	ChildID         uuid.UUID
	ChildName       string
	SharePercentage float64
	TotalCost       *money.Money
}

type ParentAssetLink struct {
	ParentID        uuid.UUID
	ParentName      string
	SharePercentage float64
}

type OwnerRef struct {
	PersonID    uuid.UUID
	DisplayName string
}

type Asset struct {
	id                 uuid.UUID
	name               string
	description        string
	projectID          uuid.UUID
	price              *money.Money
	acquiredAt         time.Time
	owner              OwnerRef
	status             AssetStatus
	startDate          time.Time
	completedDate      *time.Time
	targetPrice        *money.Money
	directCost         *money.Money
	totalCost          *money.Money
	homeCurrency       string
	homeAmount         int64
	targetHomeAmount   *int64
	progressPercentage *float64
	children           []ChildAssetLink
	parents            []ParentAssetLink
	tags               []string
}

// CreateAsset validates business invariants and constructs a new Asset entity.
func CreateAsset(
	id uuid.UUID,
	name string,
	description string,
	projectID uuid.UUID,
	price *money.Money,
	startDate time.Time,
	owner OwnerRef,
	status AssetStatus,
	completedDate *time.Time,
	tags []string,
) (*Asset, error) {
	return createAsset(id, name, description, projectID, price, startDate, owner, status, completedDate, tags, false)
}

func createAsset(
	id uuid.UUID,
	name string,
	description string,
	projectID uuid.UUID,
	price *money.Money,
	startDate time.Time,
	owner OwnerRef,
	status AssetStatus,
	completedDate *time.Time,
	tags []string,
	isGroup bool,
) (*Asset, error) {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if projectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}
	if !isGroup {
		if price == nil || price.Amount() <= 0 {
			return nil, exceptions.NewValidationException("price must be greater than 0", nil)
		}
		if price.Currency() == nil || strings.TrimSpace(price.Currency().Code) == "" {
			return nil, exceptions.NewValidationException("currency is required", nil)
		}
	} else if price == nil {
		price = money.New(0, "UAH")
	}
	if startDate.IsZero() {
		startDate = time.Now()
	}
	if status == "" {
		status = AssetStatusActive
	}
	tagsCopy := make([]string, len(tags))
	copy(tagsCopy, tags)
	a := &Asset{
		id:            id,
		name:          trimmedName,
		description:   description,
		projectID:     projectID,
		price:         price,
		acquiredAt:    startDate,
		owner:         owner,
		status:        status,
		startDate:     startDate,
		completedDate: completedDate,
		targetPrice:   price,
		children:      make([]ChildAssetLink, 0),
		parents:       make([]ParentAssetLink, 0),
		tags:          tagsCopy,
	}
	a.EvaluateValuation("UAH", nil)
	return a, nil
}

// CreateAssetFromInvestments creates a new Asset by aggregating and converting existing investments.
func CreateAssetFromInvestments(
	id uuid.UUID,
	name string,
	description string,
	projectID uuid.UUID,
	mainCurrency string,
	requestedTargetCurrency string,
	acquiredAt time.Time,
	owner OwnerRef,
	invItems []InvestmentItem,
	converter CurrencyConverter,
) (*Asset, error) {
	if projectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}
	if len(invItems) == 0 {
		return nil, exceptions.NewValidationException("at least one investment must be selected", nil)
	}
	if strings.TrimSpace(name) == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	if acquiredAt.IsZero() {
		for _, inv := range invItems {
			if inv.Date.After(acquiredAt) {
				acquiredAt = inv.Date
			}
		}
		if acquiredAt.IsZero() {
			acquiredAt = time.Now()
		}
	}

	targetCurrency := strings.ToUpper(strings.TrimSpace(requestedTargetCurrency))
	if targetCurrency == "" {
		firstCurr := invItems[0].Amount.Currency().Code
		allSameCurr := true
		for _, inv := range invItems[1:] {
			if inv.Amount.Currency().Code != firstCurr {
				allSameCurr = false
				break
			}
		}
		if allSameCurr {
			targetCurrency = firstCurr
		} else {
			targetCurrency = mainCurrency
			if targetCurrency == "" {
				targetCurrency = "UAH"
			}
		}
	}

	var totalAmount int64
	tagSet := make(map[string]struct{})
	for _, inv := range invItems {
		currCode := inv.Amount.Currency().Code
		if currCode == targetCurrency {
			totalAmount += inv.Amount.Amount()
		} else {
			if converter == nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("currency rate provider not available to convert %s to %s", currCode, targetCurrency), nil)
			}
			converted, err := converter.ConvertMoney(inv.Amount, targetCurrency)
			if err != nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("failed to convert currency %s to %s: %s", currCode, targetCurrency, err.Error()), err)
			}
			totalAmount += converted.Amount()
		}

		for _, t := range inv.Tags {
			if t != "" {
				tagSet[t] = struct{}{}
			}
		}
	}

	var tags []string
	for t := range tagSet {
		tags = append(tags, t)
	}

	if totalAmount <= 0 {
		return nil, exceptions.NewValidationException("total asset price must be greater than 0", nil)
	}

	price := money.New(totalAmount, targetCurrency)
	anAsset, err := CreateAsset(
		id,
		name,
		description,
		projectID,
		price,
		acquiredAt,
		owner,
		AssetStatusActive,
		nil,
		tags,
	)
	if err != nil {
		return nil, err
	}
	anAsset.AssignValuation(price, price, nil, nil)
	anAsset.EvaluateValuation(mainCurrency, converter)
	return anAsset, nil
}

// CreateAssetGroup constructs a parent group Asset to aggregate child assets.
func CreateAssetGroup(
	id uuid.UUID,
	name string,
	description string,
	projectID uuid.UUID,
	owner OwnerRef,
	tags []string,
) (*Asset, error) {
	now := time.Now()
	zeroPrice := money.New(0, "UAH")
	return createAsset(
		id,
		name,
		description,
		projectID,
		zeroPrice,
		now,
		owner,
		AssetStatusActive,
		nil,
		tags,
		true,
	)
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

func (a *Asset) HomeCurrency() string {
	if a.homeCurrency == "" {
		return "UAH"
	}
	return a.homeCurrency
}

func (a *Asset) HomeAmount() int64 {
	if a.homeAmount != 0 {
		return a.homeAmount
	}
	if a.Price() != nil {
		return a.Price().Amount()
	}
	return 0
}

func (a *Asset) TargetHomeAmount() *int64 {
	return a.targetHomeAmount
}

func (a *Asset) ProgressPercentage() *float64 {
	return a.progressPercentage
}

func (a *Asset) Children() []ChildAssetLink {
	if a.children == nil {
		return make([]ChildAssetLink, 0)
	}
	cp := make([]ChildAssetLink, len(a.children))
	copy(cp, a.children)
	return cp
}

func (a *Asset) Parents() []ParentAssetLink {
	if a.parents == nil {
		return make([]ParentAssetLink, 0)
	}
	cp := make([]ParentAssetLink, len(a.parents))
	copy(cp, a.parents)
	return cp
}

func (a *Asset) Tags() []string {
	if a.tags == nil {
		return make([]string, 0)
	}
	cp := make([]string, len(a.tags))
	copy(cp, a.tags)
	return cp
}

// EvaluateValuation computes and stores the home-currency valuation and budget progress on the asset.
func (a *Asset) EvaluateValuation(homeCurrency string, converter CurrencyConverter) {
	if homeCurrency == "" {
		if a.homeCurrency != "" {
			homeCurrency = a.homeCurrency
		} else {
			homeCurrency = "UAH"
		}
	}
	a.homeCurrency = homeCurrency

	priceAmount := int64(0)
	priceCurrency := homeCurrency
	if a.Price() != nil {
		priceAmount = a.Price().Amount()
		priceCurrency = a.Price().Currency().Code
	}

	homeAmount := priceAmount
	if converter != nil && priceCurrency != homeCurrency && priceAmount > 0 {
		if converted, err := converter.ConvertMoney(a.Price(), homeCurrency); err == nil && converted != nil {
			homeAmount = converted.Amount()
		}
	}
	a.homeAmount = homeAmount

	var totalCostVal int64
	var totalCostCurr string
	if a.TotalCost() != nil {
		totalCostVal = a.TotalCost().Amount()
		totalCostCurr = a.TotalCost().Currency().Code
	} else {
		totalCostVal = priceAmount
		totalCostCurr = priceCurrency
	}

	bp := CalculateBudgetProgress(
		money.New(totalCostVal, totalCostCurr),
		a.TargetPrice(),
		homeCurrency,
		converter,
	)
	a.targetHomeAmount = bp.TargetHomeAmount
	a.progressPercentage = bp.ProgressPercentage
}

// AssignValuation assigns computed financial metrics and composition links to the asset.
func (a *Asset) AssignValuation(directCost, totalCost *money.Money, children []ChildAssetLink, parents []ParentAssetLink) {
	a.directCost = directCost
	a.totalCost = totalCost
	if children == nil {
		a.children = make([]ChildAssetLink, 0)
	} else {
		cp := make([]ChildAssetLink, len(children))
		copy(cp, children)
		a.children = cp
	}
	if parents == nil {
		a.parents = make([]ParentAssetLink, 0)
	} else {
		cp := make([]ParentAssetLink, len(parents))
		copy(cp, parents)
		a.parents = cp
	}
	a.EvaluateValuation(a.homeCurrency, nil)
}

// UpdateDetails updates the descriptive information, target budget, and tags of the asset.
func (a *Asset) UpdateDetails(name, description string, targetPrice *money.Money, tags []string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return exceptions.NewValidationException("name is required", nil)
	}
	if targetPrice != nil && targetPrice.Amount() <= 0 {
		return exceptions.NewValidationException("price must be greater than 0", nil)
	}
	a.name = trimmed
	a.description = description
	a.targetPrice = targetPrice
	if tags != nil {
		cp := make([]string, len(tags))
		copy(cp, tags)
		a.tags = cp
	}
	a.EvaluateValuation(a.homeCurrency, nil)
	return nil
}

// Reschedule updates the start/acquisition date of the asset.
func (a *Asset) Reschedule(startDate time.Time) {
	if !startDate.IsZero() {
		a.startDate = startDate
		a.acquiredAt = startDate
	}
}

// TransitionStatus applies lifecycle status transitions (Complete, Reopen, or explicit status).
func (a *Asset) TransitionStatus(status AssetStatus, completedDate *time.Time) {
	if completedDate != nil {
		a.Complete(*completedDate)
	} else if status == AssetStatusActive {
		a.Reopen()
	} else if status != "" {
		a.status = status
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
	tagsCopy := make([]string, len(tags))
	copy(tagsCopy, tags)
	if status == "" {
		status = AssetStatusActive
	}
	a := &Asset{
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
		tags:          tagsCopy,
	}
	a.EvaluateValuation("UAH", nil)
	return a
}

type Collection = core.PaginatedCollection[*Asset]

func NewCollection(total int) *Collection {
	return core.NewPaginatedCollection[*Asset](total)
}
