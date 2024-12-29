package people

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type Person struct {
	id          uuid.UUID
	firstName   string
	lastName    string
	displayName string
	identities  []Credentials
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

	return false, fmt.Errorf("invalid credentials for a person %s", p.ID)
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

func NewPerson(personID uuid.UUID, firstName string, lastName string, displayName string, identities []Credentials) (*Person, error) {
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

	if err != nil {
		return nil, err
	}

	return &Person{
		id:          id,
		firstName:   firstName,
		lastName:    lastName,
		displayName: displayName,
		identities:  identities,
	}, nil
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
