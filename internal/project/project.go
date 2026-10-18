package project

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
	MinProjectNameLength = 3
	MaxProjectNameLength = 100
)

var supportedCurrencies = map[string]struct{}{
	"UAH": {},
	"USD": {},
	"EUR": {},
	"PLN": {},
}

func ValidateCurrency(currency string) (string, error) {
	norm := strings.ToUpper(strings.TrimSpace(currency))
	if norm == "" {
		return "UAH", nil
	}
	if _, ok := supportedCurrencies[norm]; !ok {
		return "", exceptions.NewValidationException(fmt.Sprintf("unsupported currency: %s", currency), nil)
	}
	return norm, nil
}

type Project struct {
	id           uuid.UUID
	name         string
	description  string
	owner        *Owner
	startDate    time.Time
	endDate      time.Time
	shareToken   uuid.UUID
	isShared     bool
	mainCurrency string
	participants []Participant
}

func (p *Project) IsOwnedBy(owner *Owner) bool {
	if p.owner == nil || owner == nil {
		return false
	}
	return p.owner.PersonID == owner.PersonID
}

func (p *Project) IsOwnedByPerson(personID uuid.UUID) bool {
	if p.owner == nil || personID == uuid.Nil {
		return false
	}
	return p.owner.PersonID == personID
}

func (p *Project) AddParticipant(participant Participant) {
	for _, existing := range p.participants {
		if existing.Equals(participant) {
			return
		}
	}
	p.participants = append(p.participants, participant)
}

func (p *Project) UpdateDetails(name, description, mainCurrency string) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName != "" {
		if len(trimmedName) < MinProjectNameLength || len(trimmedName) > MaxProjectNameLength {
			return exceptions.NewValidationException("project name must be between 3 and 100 characters", nil)
		}
		p.name = trimmedName
	}

	if description != "" {
		p.description = strings.TrimSpace(description)
	}

	trimmedCurr := strings.TrimSpace(mainCurrency)
	if trimmedCurr != "" {
		validCurr, err := ValidateCurrency(trimmedCurr)
		if err != nil {
			return err
		}
		p.mainCurrency = validCurr
	}

	return nil
}

func (p *Project) SetDates(startDate, endDate time.Time) error {
	if !startDate.IsZero() && !endDate.IsZero() && endDate.Before(startDate) {
		return exceptions.NewValidationException("end date cannot be before start date", nil)
	}
	p.startDate = startDate
	p.endDate = endDate
	return nil
}

func (p *Project) RegenerateShareToken() {
	p.shareToken = uuid.New()
}

func (p *Project) MarkShared() {
	p.isShared = true
}

func NewProject(id uuid.UUID, name string, description string, owner *Owner, startDate time.Time, endDate time.Time, mainCurrency ...string) (*Project, error) {
	if id == uuid.Nil {
		id = uuid.New()
	}

	trimmedName := strings.TrimSpace(name)
	if len(trimmedName) < MinProjectNameLength || len(trimmedName) > MaxProjectNameLength {
		return nil, exceptions.NewValidationException("project name must be between 3 and 100 characters", nil)
	}

	curr := "UAH"
	if len(mainCurrency) > 0 && mainCurrency[0] != "" {
		validCurr, err := ValidateCurrency(mainCurrency[0])
		if err != nil {
			return nil, err
		}
		curr = validCurr
	}

	if !startDate.IsZero() && !endDate.IsZero() && endDate.Before(startDate) {
		return nil, exceptions.NewValidationException("end date cannot be before start date", nil)
	}

	proj := &Project{
		id:           id,
		name:         trimmedName,
		description:  strings.TrimSpace(description),
		owner:        owner,
		startDate:    startDate,
		endDate:      endDate,
		shareToken:   uuid.New(),
		isShared:     false,
		mainCurrency: curr,
		participants: make([]Participant, 0),
	}

	if owner != nil && owner.DisplayName != "" {
		if participant, err := NewParticipant(owner.DisplayName); err == nil {
			proj.AddParticipant(participant)
		}
	}

	return proj, nil
}

// ReconstituteProject hydrates a Project entity from persistence storage without executing domain validation.
func ReconstituteProject(
	id uuid.UUID,
	name string,
	description string,
	owner *Owner,
	startDate time.Time,
	endDate time.Time,
	shareToken uuid.UUID,
	isShared bool,
	mainCurrency string,
	participants []Participant,
) *Project {
	curr := strings.ToUpper(strings.TrimSpace(mainCurrency))
	if curr == "" {
		curr = "UAH"
	}

	var parts []Participant
	if participants != nil {
		parts = make([]Participant, len(participants))
		copy(parts, participants)
	} else {
		parts = make([]Participant, 0)
	}

	return &Project{
		id:           id,
		name:         name,
		description:  description,
		owner:        owner,
		startDate:    startDate,
		endDate:      endDate,
		shareToken:   shareToken,
		isShared:     isShared,
		mainCurrency: curr,
		participants: parts,
	}
}

// Getters

func (p *Project) ID() uuid.UUID                        { return p.id }
func (p *Project) ProjectID() uuid.UUID                 { return p.id }
func (p *Project) Name() string                         { return p.name }
func (p *Project) Description() string                  { return p.description }
func (p *Project) Owner() *Owner {
	if p.owner == nil {
		return nil
	}
	cp := *p.owner
	return &cp
}
func (p *Project) OwnerID() uuid.UUID {
	if p.owner != nil {
		return p.owner.PersonID
	}
	return uuid.Nil
}
func (p *Project) StartDate() time.Time                 { return p.startDate }
func (p *Project) EndDate() time.Time                   { return p.endDate }
func (p *Project) ShareToken() uuid.UUID                { return p.shareToken }
func (p *Project) IsShared() bool                       { return p.isShared }
func (p *Project) MainCurrency() string                 { return p.mainCurrency }

func (p *Project) Participants() []Participant {
	if p.participants == nil {
		return make([]Participant, 0)
	}
	cp := make([]Participant, len(p.participants))
	copy(cp, p.participants)
	return cp
}
