package project

import "gitlab.com/massimo-ua/projecta/internal/core"

type Owner struct {
	PersonID        core.PersonID
	DisplayName     string
	FirstName       string
	CanHaveProjects bool
}

