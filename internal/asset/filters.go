package asset

import (
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CollectionFilter struct {
	core.Pagination
	core.Sorting
	ProjectID core.ProjectID
	TypeID    uuid.UUID
	OwnerID   core.PersonID
	Name      string
}

type Filter struct {
	ID        core.AssetID
	ProjectID core.ProjectID
	OwnerID   core.PersonID
	Name      string
}
