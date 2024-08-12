package asset

import (
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"time"
)

type CreateAssetCommand struct {
	Name        string
	Description string
	ProjectID   string
	TypeID      string
	Price       *money.Money
	AcquiredAt  time.Time
}

type RemoveAssetCommand struct {
	AssetID   uuid.UUID
	ProjectID uuid.UUID
}
