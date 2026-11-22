package project

import (
	"fmt"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

// CanPersonViewProject enforces the domain rule for reading a project:
// only the project owner or an explicit share participant may view it.
func CanPersonViewProject(personID core.PersonID, p Project) error {
	if personID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	if p.IsOwnedByPerson(personID) || p.HasSharedAccess(personID) || p.IsShared() {
		return nil
	}
	return exceptions.NewForbiddenException(
		fmt.Sprintf("person %s cannot access project %s", personID, p.ID()),
		nil,
	)
}

// CanPersonManageProject enforces the domain rule for destructive or administrative
// mutations (updating details, updating budgets, removing a project): owner only.
func CanPersonManageProject(personID core.PersonID, p Project) error {
	if personID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	if p.IsOwnedByPerson(personID) {
		return nil
	}
	return exceptions.NewForbiddenException(
		fmt.Sprintf("only project owner can modify project %s", p.ID()),
		nil,
	)
}
