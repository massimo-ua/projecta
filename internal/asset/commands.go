package asset

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CreateAssetCommand struct {
	Name          string
	Description   string
	ProjectID     core.ProjectID
	TypeID        uuid.UUID
	Status        AssetStatus
	StartDate     time.Time
	CompletedDate *time.Time
	TargetPrice   *money.Money
	Price         *money.Money
	AcquiredAt    time.Time
	WithPayment   bool
	Tags          []string
}

type CreateAssetFromInvestmentsCommand struct {
	ProjectID      core.ProjectID
	InvestmentIDs  []core.InvestmentID
	Name           string
	Description    string
	TypeID         uuid.UUID
	AcquiredAt     time.Time
	TargetCurrency string
}

// Deprecated: Use CreateAssetFromInvestmentsCommand instead.
type CreateAssetFromPaymentsCommand struct {
	ProjectID      core.ProjectID
	PaymentIDs     []core.InvestmentID
	Name           string
	Description    string
	TypeID         uuid.UUID
	AcquiredAt     time.Time
	TargetCurrency string
}

type UpdateAssetCommand struct {
	AssetID       core.AssetID
	Name          string
	Description   string
	ProjectID     core.ProjectID
	TypeID        uuid.UUID
	Status        AssetStatus
	StartDate     time.Time
	CompletedDate *time.Time
	TargetPrice   *money.Money
	Price         *money.Money
	AcquiredAt    time.Time
	Tags          []string
}

type RemoveAssetCommand struct {
	AssetID   core.AssetID
	ProjectID core.ProjectID
}

type LinkChildCommand struct {
	ParentID        core.AssetID
	ChildID         core.AssetID
	ProjectID       core.ProjectID
	SharePercentage float64
}

type UnlinkChildCommand struct {
	ParentID  core.AssetID
	ChildID   core.AssetID
	ProjectID core.ProjectID
}

type AssignInvestmentsCommand struct {
	AssetID       core.AssetID
	ProjectID     core.ProjectID
	InvestmentIDs []core.InvestmentID
}

type GroupAssetsCommand struct {
	ProjectID     core.ProjectID
	Name          string
	Description   string
	ChildAssetIDs []core.AssetID
	Tags          []string
}

type LinkChildrenCommand struct {
	ProjectID     core.ProjectID
	ParentID      core.AssetID
	ChildAssetIDs []core.AssetID
}


