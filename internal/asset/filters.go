package asset

import (
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CollectionFilter struct {
	core.Pagination
	core.Sorting
	ProjectID uuid.UUID
	TypeID    uuid.UUID
	OwnerID   uuid.UUID
	Name      string
}

type Filter struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	OwnerID   uuid.UUID
	Name      string
}
