package project

import (
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type ProjectFilter struct {
	Name      string
	ProjectID uuid.UUID
}

type ProjectCollectionFilter struct {
	core.Pagination
	Name string
}
