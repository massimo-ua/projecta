package investment

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ResourceType string

const (
	ResourceTypeMoney ResourceType = "MONEY"
	ResourceTypeTime  ResourceType = "TIME"
	ResourceTypeGoods ResourceType = "GOODS"
)

func (r ResourceType) String() string {
	return string(r)
}

func ToResourceType(val string) (ResourceType, error) {
	switch strings.ToUpper(strings.TrimSpace(val)) {
	case string(ResourceTypeMoney):
		return ResourceTypeMoney, nil
	case string(ResourceTypeTime):
		return ResourceTypeTime, nil
	case string(ResourceTypeGoods):
		return ResourceTypeGoods, nil
	default:
		return ResourceTypeMoney, exceptions.NewValidationException("invalid resource type", nil)
	}
}

type InvestmentAssetLink struct {
	AssetID         uuid.UUID `json:"asset_id"`
	AssetName       string    `json:"asset_name,omitempty"`
	SharePercentage float64   `json:"share_percentage"`
}

type ContributorRef struct {
	PersonID    uuid.UUID `json:"person_id"`
	DisplayName string    `json:"display_name"`
}

type Investment struct {
	id               uuid.UUID
	projectID        uuid.UUID
	homeCurrency     string
	primaryAssetID   uuid.UUID
	primaryAssetName string
	assetAllocations []InvestmentAssetLink
	contributor      ContributorRef
	resourceType     ResourceType
	amount           *money.Money
	timeHours        *float64
	timeHourlyRate   *money.Money
	goodsQuantity    *float64
	goodsUnit        string
	goodsItemName    string
	description      string
	date             time.Time
	tags             []string
}

func NewInvestment(
	id uuid.UUID,
	projectID uuid.UUID,
	homeCurrency string,
	primaryAssetID uuid.UUID,
	primaryAssetName string,
	contributor ContributorRef,
	resourceType ResourceType,
	amount *money.Money,
	timeHours *float64,
	timeHourlyRate *money.Money,
	goodsQuantity *float64,
	goodsUnit string,
	goodsItemName string,
	description string,
	date time.Time,
	tags []string,
	allocations ...InvestmentAssetLink,
) (*Investment, error) {
	if id == uuid.Nil {
		return nil, exceptions.NewValidationException("investment id is required", nil)
	}
	if projectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project is required", nil)
	}
	if strings.TrimSpace(description) == "" {
		return nil, exceptions.NewValidationException("description is required", nil)
	}

	cleanTags := cleanTagsSlice(tags)
	if date.IsZero() {
		date = time.Now()
	}

	finalAmount := amount
	if (finalAmount == nil || finalAmount.Amount() <= 0) && resourceType == ResourceTypeTime && timeHours != nil && timeHourlyRate != nil {
		if *timeHours > 0 && timeHourlyRate.Amount() > 0 {
			calc := int64(*timeHours * float64(timeHourlyRate.Amount()))
			finalAmount = money.New(calc, timeHourlyRate.Currency().Code)
		}
	}

	if finalAmount == nil || finalAmount.Amount() <= 0 {
		return nil, exceptions.NewValidationException("investment amount must be greater than 0", nil)
	}

	inv := &Investment{
		id:             id,
		projectID:      projectID,
		homeCurrency:   homeCurrency,
		contributor:    contributor,
		resourceType:   resourceType,
		amount:         finalAmount,
		timeHours:      timeHours,
		timeHourlyRate: timeHourlyRate,
		goodsQuantity:  goodsQuantity,
		goodsUnit:      goodsUnit,
		goodsItemName:  goodsItemName,
		description:    description,
		date:           date,
		tags:           cleanTags,
	}

	if len(allocations) > 0 {
		if err := inv.Allocate(allocations); err != nil {
			return nil, err
		}
	} else if primaryAssetID != uuid.Nil {
		inv.primaryAssetID = primaryAssetID
		inv.primaryAssetName = primaryAssetName
		inv.assetAllocations = []InvestmentAssetLink{
			{
				AssetID:         primaryAssetID,
				AssetName:       primaryAssetName,
				SharePercentage: 100.0,
			},
		}
	} else {
		return nil, exceptions.NewValidationException("asset id or asset allocations are required", nil)
	}

	return inv, nil
}

// Allocate sets and validates asset allocations, ensuring the sum of shares equals 100% and no duplicates exist.
func (i *Investment) Allocate(allocations []InvestmentAssetLink) error {
	if len(allocations) == 0 {
		return exceptions.NewValidationException("at least one asset allocation is required", nil)
	}

	if len(allocations) == 1 {
		a := allocations[0]
		if a.AssetID == uuid.Nil {
			return exceptions.NewValidationException("asset id is required for allocation", nil)
		}
		i.primaryAssetID = a.AssetID
		i.primaryAssetName = a.AssetName
		i.assetAllocations = []InvestmentAssetLink{
			{
				AssetID:         a.AssetID,
				AssetName:       a.AssetName,
				SharePercentage: 100.0,
			},
		}
		return nil
	}

	seen := make(map[uuid.UUID]bool)
	var totalShare float64
	for _, a := range allocations {
		if a.AssetID == uuid.Nil {
			return exceptions.NewValidationException("asset id is required for each allocation", nil)
		}
		if seen[a.AssetID] {
			return exceptions.NewValidationException("duplicate asset in allocations", nil)
		}
		seen[a.AssetID] = true

		if a.SharePercentage <= 0 || a.SharePercentage > 100 {
			return exceptions.NewValidationException("share percentage must be between 0 and 100 for shared investments", nil)
		}
		totalShare += a.SharePercentage
	}

	if math.Abs(totalShare-100.0) > 0.01 {
		return exceptions.NewValidationException(fmt.Sprintf("total share percentages must equal 100%%, got %.2f%%", totalShare), nil)
	}

	i.primaryAssetID = allocations[0].AssetID
	i.primaryAssetName = allocations[0].AssetName
	i.assetAllocations = allocations
	return nil
}

// UpdateDetails updates the descriptive information, date, and tags of the investment.
func (i *Investment) UpdateDetails(description string, date time.Time, tags []string) {
	if strings.TrimSpace(description) != "" {
		i.description = description
	}
	if !date.IsZero() {
		i.date = date
	}
	if tags != nil {
		i.tags = cleanTagsSlice(tags)
	}
}

// UpdateValuation updates the resource type and financial valuation metrics.
func (i *Investment) UpdateValuation(
	resType ResourceType,
	amount *money.Money,
	timeHours *float64,
	timeHourlyRate *money.Money,
	goodsQuantity *float64,
	goodsUnit string,
	goodsItemName string,
) error {
	if resType != "" {
		i.resourceType = resType
	}
	if timeHours != nil {
		i.timeHours = timeHours
	}
	if timeHourlyRate != nil {
		i.timeHourlyRate = timeHourlyRate
	}
	if goodsQuantity != nil {
		i.goodsQuantity = goodsQuantity
	}
	if goodsUnit != "" {
		i.goodsUnit = goodsUnit
	}
	if goodsItemName != "" {
		i.goodsItemName = goodsItemName
	}

	finalAmount := amount
	if (finalAmount == nil || finalAmount.Amount() <= 0) && i.resourceType == ResourceTypeTime && i.timeHours != nil && i.timeHourlyRate != nil {
		if *i.timeHours > 0 && i.timeHourlyRate.Amount() > 0 {
			calc := int64(*i.timeHours * float64(i.timeHourlyRate.Amount()))
			finalAmount = money.New(calc, i.timeHourlyRate.Currency().Code)
		}
	}

	if finalAmount != nil && finalAmount.Amount() > 0 {
		i.amount = finalAmount
	}
	return nil
}

// ReconstituteInvestment reconstitutes an Investment entity from persistence data without executing business invariants.
func ReconstituteInvestment(
	id uuid.UUID,
	projectID uuid.UUID,
	homeCurrency string,
	primaryAssetID uuid.UUID,
	primaryAssetName string,
	allocations []InvestmentAssetLink,
	contributor ContributorRef,
	resourceType ResourceType,
	amount *money.Money,
	timeHours *float64,
	timeHourlyRate *money.Money,
	goodsQuantity *float64,
	goodsUnit string,
	goodsItemName string,
	description string,
	date time.Time,
	tags []string,
) *Investment {
	cleanTags := cleanTagsSlice(tags)
	if date.IsZero() {
		date = time.Now()
	}

	finalAllocations := allocations
	if len(finalAllocations) == 0 && primaryAssetID != uuid.Nil {
		finalAllocations = []InvestmentAssetLink{
			{
				AssetID:         primaryAssetID,
				AssetName:       primaryAssetName,
				SharePercentage: 100.0,
			},
		}
	} else if finalAllocations == nil {
		finalAllocations = make([]InvestmentAssetLink, 0)
	}

	return &Investment{
		id:               id,
		projectID:        projectID,
		homeCurrency:     homeCurrency,
		primaryAssetID:   primaryAssetID,
		primaryAssetName: primaryAssetName,
		assetAllocations: finalAllocations,
		contributor:      contributor,
		resourceType:     resourceType,
		amount:           amount,
		timeHours:        timeHours,
		timeHourlyRate:   timeHourlyRate,
		goodsQuantity:    goodsQuantity,
		goodsUnit:        goodsUnit,
		goodsItemName:    goodsItemName,
		description:      description,
		date:             date,
		tags:             cleanTags,
	}
}

// Getters
func (i *Investment) ID() uuid.UUID                               { return i.id }
func (i *Investment) ProjectID() uuid.UUID                        { return i.projectID }
func (i *Investment) HomeCurrency() string                        { return i.homeCurrency }
func (i *Investment) PrimaryAssetID() uuid.UUID                  { return i.primaryAssetID }
func (i *Investment) PrimaryAssetName() string                   { return i.primaryAssetName }
func (i *Investment) AssetAllocations() []InvestmentAssetLink     { return i.assetAllocations }
func (i *Investment) Contributor() ContributorRef                 { return i.contributor }
func (i *Investment) ResourceType() ResourceType                 { return i.resourceType }
func (i *Investment) Amount() *money.Money                       { return i.amount }
func (i *Investment) TimeHours() *float64                        { return i.timeHours }
func (i *Investment) TimeHourlyRate() *money.Money               { return i.timeHourlyRate }
func (i *Investment) GoodsQuantity() *float64                    { return i.goodsQuantity }
func (i *Investment) GoodsUnit() string                          { return i.goodsUnit }
func (i *Investment) GoodsItemName() string                      { return i.goodsItemName }
func (i *Investment) Description() string                        { return i.description }
func (i *Investment) Date() time.Time                            { return i.date }
func (i *Investment) Tags() []string                             { return i.tags }

func cleanTagsSlice(tags []string) []string {
	if tags == nil {
		return make([]string, 0)
	}
	clean := make([]string, 0, len(tags))
	for _, t := range tags {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	return clean
}

type Collection = core.PaginatedCollection[*Investment]

func NewCollection(total int) *Collection {
	return core.NewPaginatedCollection[*Investment](total)
}
