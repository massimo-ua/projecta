package investment

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type CurrencyConverter interface {
	ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error)
}

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
	AssetID         core.AssetID
	AssetName       string
	SharePercentage float64
}

func (l InvestmentAssetLink) AllocatedAmount(total *money.Money) int64 {
	if total == nil {
		return 0
	}
	return int64(float64(total.Amount()) * (l.SharePercentage / 100.0))
}

type ContributorRef struct {
	PersonID    core.PersonID
	DisplayName string
}

type Investment struct {
	id               core.InvestmentID
	projectID        core.ProjectID
	homeCurrency     string
	homeAmount       int64
	primaryAssetID   core.AssetID
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
	id core.InvestmentID,
	projectID core.ProjectID,
	homeCurrency string,
	primaryAssetID core.AssetID,
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
	if id.IsNil() {
		return nil, exceptions.NewValidationException("investment id is required", nil)
	}
	if projectID.IsNil() {
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
		timeHours:      copyFloat64Ptr(timeHours),
		timeHourlyRate: timeHourlyRate,
		goodsQuantity:  copyFloat64Ptr(goodsQuantity),
		goodsUnit:      goodsUnit,
		goodsItemName:  goodsItemName,
		description:    strings.TrimSpace(description),
		date:           date,
		tags:           cleanTags,
	}

	if len(allocations) > 0 {
		if err := inv.Allocate(allocations); err != nil {
			return nil, err
		}
	} else if !primaryAssetID.IsNil() {
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
		if a.AssetID.IsNil() {
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

	seen := make(map[core.AssetID]bool)
	var totalShare float64
	for _, a := range allocations {
		if a.AssetID.IsNil() {
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

	allocCopy := make([]InvestmentAssetLink, len(allocations))
	copy(allocCopy, allocations)
	i.primaryAssetID = allocCopy[0].AssetID
	i.primaryAssetName = allocCopy[0].AssetName
	i.assetAllocations = allocCopy
	return nil
}

// UpdateDetails updates the descriptive information, date, and tags of the investment.
func (i *Investment) UpdateDetails(description string, date time.Time, tags []string) {
	if strings.TrimSpace(description) != "" {
		i.description = strings.TrimSpace(description)
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
	targetType := i.resourceType
	if resType != "" {
		if _, err := ToResourceType(string(resType)); err != nil {
			return err
		}
		targetType = resType
	}

	if timeHours != nil && *timeHours <= 0 {
		return exceptions.NewValidationException("time hours must be greater than 0", nil)
	}
	if timeHourlyRate != nil && timeHourlyRate.Amount() <= 0 {
		return exceptions.NewValidationException("time hourly rate must be greater than 0", nil)
	}
	if goodsQuantity != nil && *goodsQuantity <= 0 {
		return exceptions.NewValidationException("goods quantity must be greater than 0", nil)
	}

	finalAmount := amount
	effHours := timeHours
	if effHours == nil {
		effHours = i.timeHours
	}
	effRate := timeHourlyRate
	if effRate == nil {
		effRate = i.timeHourlyRate
	}
	if (finalAmount == nil || finalAmount.Amount() <= 0) && targetType == ResourceTypeTime && effHours != nil && effRate != nil {
		if *effHours > 0 && effRate.Amount() > 0 {
			calc := int64(*effHours * float64(effRate.Amount()))
			finalAmount = money.New(calc, effRate.Currency().Code)
		}
	}

	if amount != nil && (finalAmount == nil || finalAmount.Amount() <= 0) {
		return exceptions.NewValidationException("investment amount must be greater than 0", nil)
	}

	i.resourceType = targetType
	if timeHours != nil {
		i.timeHours = copyFloat64Ptr(timeHours)
	}
	if timeHourlyRate != nil {
		i.timeHourlyRate = timeHourlyRate
	}
	if goodsQuantity != nil {
		i.goodsQuantity = copyFloat64Ptr(goodsQuantity)
	}
	if goodsUnit != "" {
		i.goodsUnit = goodsUnit
	}
	if goodsItemName != "" {
		i.goodsItemName = goodsItemName
	}
	if finalAmount != nil && finalAmount.Amount() > 0 {
		i.amount = finalAmount
	}
	return nil
}

// ReconstituteInvestment reconstitutes an Investment entity from persistence data without executing business invariants.
func ReconstituteInvestment(
	id core.InvestmentID,
	projectID core.ProjectID,
	homeCurrency string,
	primaryAssetID core.AssetID,
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

	var finalAllocations []InvestmentAssetLink
	if len(allocations) == 0 && !primaryAssetID.IsNil() {
		finalAllocations = []InvestmentAssetLink{
			{
				AssetID:         primaryAssetID,
				AssetName:       primaryAssetName,
				SharePercentage: 100.0,
			},
		}
	} else if allocations == nil {
		finalAllocations = make([]InvestmentAssetLink, 0)
	} else {
		finalAllocations = make([]InvestmentAssetLink, len(allocations))
		copy(finalAllocations, allocations)
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
		timeHours:        copyFloat64Ptr(timeHours),
		timeHourlyRate:   timeHourlyRate,
		goodsQuantity:    copyFloat64Ptr(goodsQuantity),
		goodsUnit:        goodsUnit,
		goodsItemName:    goodsItemName,
		description:      strings.TrimSpace(description),
		date:             date,
		tags:             cleanTags,
	}
}

// EvaluateHomeAmount computes and stores the home-currency valuation for this investment.
func (i *Investment) EvaluateHomeAmount(converter CurrencyConverter, homeCurrency ...string) {
	if len(homeCurrency) > 0 && strings.TrimSpace(homeCurrency[0]) != "" {
		i.homeCurrency = strings.ToUpper(strings.TrimSpace(homeCurrency[0]))
	}
	homeCurr := i.HomeCurrency()
	if i.amount == nil {
		i.homeAmount = 0
		return
	}
	i.homeAmount = i.amount.Amount()
	if converter != nil && i.amount.Currency() != nil && i.amount.Currency().Code != homeCurr {
		if converted, err := converter.ConvertMoney(i.amount, homeCurr); err == nil && converted != nil {
			i.homeAmount = converted.Amount()
		}
	}
}

// Getters
func (i *Investment) ID() core.InvestmentID           { return i.id }
func (i *Investment) ProjectID() core.ProjectID       { return i.projectID }
func (i *Investment) HomeCurrency() string {
	if i.homeCurrency == "" {
		return "UAH"
	}
	return i.homeCurrency
}
func (i *Investment) HomeAmount() int64 {
	if i.homeAmount != 0 {
		return i.homeAmount
	}
	if i.amount != nil {
		return i.amount.Amount()
	}
	return 0
}
func (i *Investment) PrimaryAssetID() core.AssetID { return i.primaryAssetID }
func (i *Investment) PrimaryAssetName() string    { return i.primaryAssetName }
func (i *Investment) AssetAllocations() []InvestmentAssetLink {
	if i.assetAllocations == nil {
		return make([]InvestmentAssetLink, 0)
	}
	cp := make([]InvestmentAssetLink, len(i.assetAllocations))
	copy(cp, i.assetAllocations)
	return cp
}
func (i *Investment) Contributor() ContributorRef { return i.contributor }
func (i *Investment) ResourceType() ResourceType { return i.resourceType }
func (i *Investment) Amount() *money.Money       { return i.amount }
func (i *Investment) TimeHours() *float64        { return copyFloat64Ptr(i.timeHours) }
func (i *Investment) TimeHourlyRate() *money.Money { return i.timeHourlyRate }
func (i *Investment) GoodsQuantity() *float64    { return copyFloat64Ptr(i.goodsQuantity) }
func (i *Investment) GoodsUnit() string          { return i.goodsUnit }
func (i *Investment) GoodsItemName() string      { return i.goodsItemName }
func (i *Investment) Description() string        { return i.description }
func (i *Investment) Date() time.Time            { return i.date }
func (i *Investment) Tags() []string {
	if i.tags == nil {
		return make([]string, 0)
	}
	cp := make([]string, len(i.tags))
	copy(cp, i.tags)
	return cp
}

func copyFloat64Ptr(p *float64) *float64 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

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
