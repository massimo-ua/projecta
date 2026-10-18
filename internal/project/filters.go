package project

import (
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type ProjectFilter struct {
	Name        string
	ProjectID   core.ProjectID
	RequesterID core.PersonID
}

type ProjectCollectionFilter struct {
	core.Pagination
	Name        string
	RequesterID core.PersonID
}
