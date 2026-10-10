package investment

import (
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type Filter struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	AssetID   uuid.UUID
}

type CollectionFilter struct {
	core.Pagination
	core.Sorting
	ProjectID    uuid.UUID
	AssetID      uuid.UUID
	OwnerID      uuid.UUID
	ResourceType ResourceType
	Tag          string
	StartDate    *time.Time
	EndDate      *time.Time
}
