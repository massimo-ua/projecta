package investment

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
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
	switch val {
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

type Investment struct {
	ID               uuid.UUID
	Project          *projecta.Project
	Asset            *asset.Asset
	AssetAllocations []InvestmentAssetLink
	Contributor      *projecta.Owner
	ResourceType     ResourceType
	Amount           *money.Money
	TimeHours        *float64
	TimeHourlyRate   *money.Money
	GoodsQuantity    *float64
	GoodsUnit        string
	GoodsItemName    string
	Description      string
	Date             time.Time
	Tags             []string
}

func NewInvestment(
	id uuid.UUID,
	project *projecta.Project,
	anAsset *asset.Asset,
	contributor *projecta.Owner,
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
	cleanTags := make([]string, 0)
	for _, t := range tags {
		if t != "" {
			cleanTags = append(cleanTags, t)
		}
	}
	if date.IsZero() {
		date = time.Now()
	}
	var defaultAllocations []InvestmentAssetLink
	if anAsset != nil {
		defaultAllocations = []InvestmentAssetLink{
			{
				AssetID:         anAsset.ID(),
				AssetName:       anAsset.Name(),
				SharePercentage: 100.0,
			},
		}
	} else {
		defaultAllocations = make([]InvestmentAssetLink, 0)
	}
	return &Investment{
		ID:               id,
		Project:          project,
		Asset:            anAsset,
		AssetAllocations: defaultAllocations,
		Contributor:      contributor,
		ResourceType:     resourceType,
		Amount:           amount,
		TimeHours:        timeHours,
		TimeHourlyRate:   timeHourlyRate,
		GoodsQuantity:    goodsQuantity,
		GoodsUnit:        goodsUnit,
		GoodsItemName:    goodsItemName,
		Description:      description,
		Date:             date,
		Tags:             cleanTags,
	}
}

func (i *Investment) SetAssetAllocations(allocations []InvestmentAssetLink) {
	if allocations == nil {
		i.AssetAllocations = make([]InvestmentAssetLink, 0)
	} else {
		i.AssetAllocations = allocations
	}
}

type Collection = core.PaginatedCollection[*Investment]

func NewCollection(total int) *Collection {
	return core.NewPaginatedCollection[*Investment](total)
}
