package investment

import (
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type InvestmentAssetInput struct {
	AssetID         core.AssetID
	SharePercentage float64
}

type CreateInvestmentCommand struct {
	ProjectID        core.ProjectID
	AssetID          core.AssetID
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
	ID               core.InvestmentID
	ProjectID        core.ProjectID
	AssetID          core.AssetID
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
	ID        core.InvestmentID
	ProjectID core.ProjectID
}

type BatchCreateInvestmentItem struct {
	AssetID        core.AssetID
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
	ProjectID    core.ProjectID
	DefaultAsset core.AssetID
	Items        []BatchCreateInvestmentItem
}
