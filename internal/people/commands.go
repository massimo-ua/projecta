package people

import "gitlab.com/massimo-ua/projecta/internal/core"

type RegisterCommand struct {
    Login            string
    FirstName        string
    LastName         string
    IdentityProvider IdentityProvider
    Token            string
}

type AssignRolesCommand struct {
    PersonID core.PersonID
    Roles    []Role
}

type UpdateDisplayNameCommand struct {
	PersonID    core.PersonID
	DisplayName DisplayName
}

type CreateInvitationCommand struct {
	AdminID core.PersonID
	Email   string
}

type DeleteInvitationCommand struct {
	AdminID      core.PersonID
	InvitationID core.InvitationID
}

