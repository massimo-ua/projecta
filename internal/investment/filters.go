package investment

import (
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
)

type Filter struct {
	ID        core.InvestmentID
	ProjectID core.ProjectID
	AssetID   core.AssetID
}

type CollectionFilter struct {
	core.Pagination
	core.Sorting
	IDs          []core.InvestmentID
	ProjectID    core.ProjectID
	AssetID      core.AssetID
	OwnerID      core.PersonID
	ResourceType ResourceType
	Tag          string
	StartDate    *time.Time
	EndDate      *time.Time
}
