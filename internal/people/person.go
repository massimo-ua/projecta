package people

import (
	"errors"
	"fmt"
	"strings"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
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
	id          core.PersonID
	firstName   string
	lastName    string
	displayName DisplayName
	identities  []Credentials
	roles       []Role
}

func (p *Person) ID() core.PersonID { return p.id }
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
	if !p.displayName.IsEmpty() {
		return p.displayName.String()
	}

	return p.FullName()
}

func (p *Person) DisplayNameVO() DisplayName {
	return p.displayName
}

func (p *Person) Identities() []Credentials {
	res := make([]Credentials, len(p.identities))
	copy(res, p.identities)
	return res
}

func (p *Person) Roles() []Role {
	res := make([]Role, len(p.roles))
	copy(res, p.roles)
	return res
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
	for _, r := range p.roles {
		if r == role {
			return nil
		}
	}
	p.roles = append(p.roles, role)
	return nil
}

func NewPerson(personID core.PersonID, firstName string, lastName string, displayName string, identities []Credentials, roles ...Role) (*Person, error) {
	var err error
	var id core.PersonID
	if personID.IsNil() {
		id = core.NewPersonID()
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

	dn, dnErr := NewDisplayName(displayName)
	if dnErr != nil {
		err = errors.Join(err, dnErr)
	}

	if err != nil {
		return nil, err
	}

	p := &Person{
		id:          id,
		firstName:   firstName,
		lastName:    lastName,
		displayName: dn,
		identities:  identities,
	}

	if len(roles) > 0 {
		_ = p.AssignRoles(roles)
	}

	return p, nil
}

func (p *Person) AddOrReplaceIdentity(credentials Credentials) error {
	for i, c := range p.identities {
		if c.Provider() == credentials.Provider() {
			p.identities[i] = credentials
			return nil
		}
	}

	p.identities = append(p.identities, credentials)
	return nil
}

func (p *Person) UpdateDisplayName(displayName DisplayName) {
	p.displayName = displayName
}

func (p *Person) UpdateProfile(firstName string, lastName string, displayName DisplayName) error {
	var err error
	if l := len(firstName); l < 2 || l > 255 {
		err = errors.Join(err, errors.New("invalid person first name"))
	}
	if l := len(lastName); l < 2 || l > 255 {
		err = errors.Join(err, errors.New("invalid person last name"))
	}
	if err != nil {
		return err
	}

	p.firstName = firstName
	p.lastName = lastName
	p.displayName = displayName
	return nil
}

func (p *Person) AcceptInvitation(claims *core.AuthTokenClaims, provider IdentityProvider) error {
	if claims == nil {
		return exceptions.NewValidationException("invalid user profile data", errors.New("claims are required"))
	}

	firstName := claims.FirstName
	lastName := claims.LastName
	displayName := claims.DisplayName

	if displayName == "" {
		if claims.FirstName != "" && claims.LastName != "" {
			displayName = claims.FirstName + " " + claims.LastName
		} else if claims.FirstName != "" {
			displayName = claims.FirstName
		} else if claims.Email != "" {
			displayName = strings.Split(claims.Email, "@")[0]
		}
	}
	if firstName == "" {
		parts := strings.Fields(displayName)
		if len(parts) >= 2 {
			firstName = parts[0]
			lastName = strings.Join(parts[1:], " ")
		} else if len(parts) == 1 {
			firstName = parts[0]
			lastName = "User"
		} else {
			firstName = "Google"
			lastName = "User"
		}
	}
	if lastName == "" {
		lastName = "User"
	}
	if len(firstName) < 2 {
		firstName = firstName + " "
	}
	if len(lastName) < 2 {
		lastName = lastName + " "
	}

	dn, err := NewDisplayName(displayName)
	if err != nil {
		return exceptions.NewValidationException("invalid user profile data", err)
	}

	if err := p.UpdateProfile(firstName, lastName, dn); err != nil {
		return exceptions.NewValidationException("invalid user profile data", err)
	}

	cred, err := NewCredentials(string(provider), claims.Sub, claims.Sub)
	if err != nil {
		return exceptions.NewInternalException("failed to prepare credentials", err)
	}

	if err := p.AddOrReplaceIdentity(cred); err != nil {
		return exceptions.NewInternalException("failed to attach credentials", err)
	}

	if err := p.AddRole(RoleUser); err != nil {
		return exceptions.NewInternalException("failed to assign role", err)
	}

	return nil
}

func ReconstitutePerson(
	id core.PersonID,
	firstName string,
	lastName string,
	displayName DisplayName,
	identities []Credentials,
	roles []Role,
) *Person {
	ids := make([]Credentials, len(identities))
	copy(ids, identities)

	r := make([]Role, len(roles))
	copy(r, roles)

	return &Person{
		id:          id,
		firstName:   firstName,
		lastName:    lastName,
		displayName: displayName,
		identities:  ids,
		roles:       r,
	}
}
