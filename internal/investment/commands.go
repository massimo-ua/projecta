package investment

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
)

type InvestmentAssetInput struct {
	AssetID         uuid.UUID
	SharePercentage float64
}

type CreateInvestmentCommand struct {
	ProjectID        uuid.UUID
	AssetID          uuid.UUID
	AssetAllocations []InvestmentAssetInput
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

type UpdateInvestmentCommand struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	AssetID          uuid.UUID
	AssetAllocations []InvestmentAssetInput
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

type RemoveInvestmentCommand struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
}

type BatchCreateInvestmentItem struct {
	AssetID        uuid.UUID
	ResourceType   ResourceType
	Amount         *money.Money
	TimeHours      *float64
	TimeHourlyRate *money.Money
	GoodsQuantity  *float64
	GoodsUnit      string
	GoodsItemName  string
	Description    string
	Date           time.Time
	Tags           []string
}

type CreateBatchInvestmentsCommand struct {
	ProjectID   uuid.UUID
	DefaultAsset uuid.UUID
	Items       []BatchCreateInvestmentItem
}
