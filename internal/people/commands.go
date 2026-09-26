package people

import "github.com/google/uuid"

type RegisterCommand struct {
    Login            string
    FirstName        string
    LastName         string
    IdentityProvider IdentityProvider
    Token            string
}

type AssignRolesCommand struct {
    PersonID uuid.UUID
    Roles    []Role
}

