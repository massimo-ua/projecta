package asset

import (
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"time"
)

type CreateAssetCommand struct {
	Name        string
	Description string
	ProjectID   uuid.UUID
	TypeID      uuid.UUID
	Price       *money.Money
	AcquiredAt  time.Time
	WithPayment bool
}

type RemoveAssetCommand struct {
	AssetID   uuid.UUID
	ProjectID uuid.UUID
}
