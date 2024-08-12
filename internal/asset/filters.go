package asset

import (
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CollectionFilter struct {
	core.Pagination
	ProjectID uuid.UUID
	TypeID    uuid.UUID
	Name      string
}

type Filter struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Name      string
}
