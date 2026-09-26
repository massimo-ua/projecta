package people

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type Role string

const (
	RoleUser          Role = "User"
	RoleAdministrator Role = "Administrator"
)

var validRoles = map[Role]struct{}{
	RoleUser:          {},
	RoleAdministrator: {},
}

func (r Role) IsValid() bool {
	_, ok := validRoles[r]
	return ok
}

func (r Role) String() string {
	return string(r)
}

func ToRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid role: %s", s)
	}
	return r, nil
}

type Person struct {
	id          uuid.UUID
	firstName   string
	lastName    string
	displayName string
	identities  []Credentials
	roles       []Role
}

var lock sync.Mutex

func (p *Person) ID() uuid.UUID     { return p.id }
func (p *Person) FirstName() string { return p.firstName }
func (p *Person) LastName() string  { return p.lastName }
func (p *Person) FullName() string {
	return p.firstName + " " + p.lastName
}

func (p *Person) Identify(credentials Credentials) (bool, error) {
	for _, c := range p.identities {
		if c.Equals(credentials) {
			return true, nil
		}
	}

	return false, fmt.Errorf("invalid credentials for a person %s", p.ID())
}

func (p *Person) DisplayName() string {
	if p.displayName != "" {
		return p.displayName
	}

	return p.FullName()
}

func (p *Person) Identities() []Credentials {
	return p.identities
}

func (p *Person) Roles() []Role {
	return p.roles
}

func (p *Person) RoleStrings() []string {
	strs := make([]string, len(p.roles))
	for i, r := range p.roles {
		strs[i] = string(r)
	}
	return strs
}

func (p *Person) HasRole(role Role) bool {
	for _, r := range p.roles {
		if r == role {
			return true
		}
	}
	return false
}

func (p *Person) IsAdministrator() bool {
	return p.HasRole(RoleAdministrator)
}

func (p *Person) IsUser() bool {
	return p.HasRole(RoleUser)
}

func (p *Person) CanHaveProjects() bool {
	return p.IsUser()
}

func (p *Person) CanManageUsers() bool {
	return p.IsAdministrator()
}

func (p *Person) AssignRoles(roles []Role) error {
	lock.Lock()
	defer lock.Unlock()

	seen := make(map[Role]bool)
	unique := make([]Role, 0, len(roles))
	for _, r := range roles {
		if !r.IsValid() {
			return fmt.Errorf("invalid role: %s", r)
		}
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}
	p.roles = unique
	return nil
}

func (p *Person) AddRole(role Role) error {
	if !role.IsValid() {
		return fmt.Errorf("invalid role: %s", role)
	}
	lock.Lock()
	defer lock.Unlock()
	for _, r := range p.roles {
		if r == role {
			return nil
		}
	}
	p.roles = append(p.roles, role)
	return nil
}

func NewPerson(personID uuid.UUID, firstName string, lastName string, displayName string, identities []Credentials, roles ...Role) (*Person, error) {
	var err error
	var id uuid.UUID
	if personID == uuid.Nil {
		id = uuid.New()
	} else {
		id = personID
	}

	if l := len(firstName); l < 2 || l > 255 {
		err = errors.Join(err, errors.New("invalid person first name"))
	}

	if l := len(lastName); l < 2 || l > 255 {
		err = errors.Join(err, errors.New("invalid person last name"))
	}

	if identities != nil && len(identities) == 0 {
		err = errors.Join(err, errors.New("no identities provided"))
	}

	for _, r := range roles {
		if !r.IsValid() {
			err = errors.Join(err, fmt.Errorf("invalid role: %s", r))
		}
	}

	if err != nil {
		return nil, err
	}

	p := &Person{
		id:          id,
		firstName:   firstName,
		lastName:    lastName,
		displayName: displayName,
		identities:  identities,
	}

	if len(roles) > 0 {
		_ = p.AssignRoles(roles)
	}

	return p, nil
}

func (p *Person) AddOrReplaceIdentity(credentials Credentials) error {
	lock.Lock()
	defer lock.Unlock()

	for i, c := range p.identities {
		if c.Provider() == credentials.Provider() {
			p.identities[i] = credentials
			return nil
		}
	}

	p.identities = append(p.identities, credentials)
	return nil
}
