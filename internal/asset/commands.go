package asset

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
)

type CreateAssetCommand struct {
	Name          string
	Description   string
	ProjectID     uuid.UUID
	TypeID        uuid.UUID
	Status        AssetStatus
	StartDate     time.Time
	CompletedDate *time.Time
	TargetPrice   *money.Money
	Price         *money.Money
	AcquiredAt    time.Time
	WithPayment   bool
}

type CreateAssetFromPaymentsCommand struct {
	ProjectID      uuid.UUID
	PaymentIDs     []uuid.UUID
	Name           string
	Description    string
	TypeID         uuid.UUID
	AcquiredAt     time.Time
	TargetCurrency string
}

type UpdateAssetCommand struct {
	AssetID       uuid.UUID
	Name          string
	Description   string
	ProjectID     uuid.UUID
	TypeID        uuid.UUID
	Status        AssetStatus
	StartDate     time.Time
	CompletedDate *time.Time
	TargetPrice   *money.Money
	Price         *money.Money
	AcquiredAt    time.Time
}

type RemoveAssetCommand struct {
	AssetID   uuid.UUID
	ProjectID uuid.UUID
}

type LinkChildCommand struct {
	ParentID        uuid.UUID
	ChildID         uuid.UUID
	ProjectID       uuid.UUID
	SharePercentage float64
}

type UnlinkChildCommand struct {
	ParentID  uuid.UUID
	ChildID   uuid.UUID
	ProjectID uuid.UUID
}
