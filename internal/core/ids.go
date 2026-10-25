package core

import (
	"github.com/google/uuid"
)

// ============================================================================
// PersonID
// ============================================================================

// PersonID uniquely identifies a person aggregate across contexts.
type PersonID uuid.UUID

var NilPersonID = PersonID(uuid.Nil)

func NewPersonID() PersonID {
	return PersonID(uuid.New())
}

func ParsePersonID(s string) (PersonID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return NilPersonID, err
	}
	return PersonID(u), nil
}

func MustParsePersonID(s string) PersonID {
	return PersonID(uuid.MustParse(s))
}

func (id PersonID) String() string {
	return uuid.UUID(id).String()
}

func (id PersonID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id PersonID) IsNil() bool {
	return id == NilPersonID
}

func (id PersonID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

func (id *PersonID) UnmarshalText(data []byte) error {
	var u uuid.UUID
	if err := u.UnmarshalText(data); err != nil {
		return err
	}
	*id = PersonID(u)
	return nil
}

// ============================================================================
// ProjectID
// ============================================================================

// ProjectID uniquely identifies a project aggregate across contexts.
type ProjectID uuid.UUID

var NilProjectID = ProjectID(uuid.Nil)

func NewProjectID() ProjectID {
	return ProjectID(uuid.New())
}

func ParseProjectID(s string) (ProjectID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return NilProjectID, err
	}
	return ProjectID(u), nil
}

func MustParseProjectID(s string) ProjectID {
	return ProjectID(uuid.MustParse(s))
}

func (id ProjectID) String() string {
	return uuid.UUID(id).String()
}

func (id ProjectID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id ProjectID) IsNil() bool {
	return id == NilProjectID
}

func (id ProjectID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

func (id *ProjectID) UnmarshalText(data []byte) error {
	var u uuid.UUID
	if err := u.UnmarshalText(data); err != nil {
		return err
	}
	*id = ProjectID(u)
	return nil
}

// ============================================================================
// AssetID
// ============================================================================

// AssetID uniquely identifies an asset aggregate across contexts.
type AssetID uuid.UUID

var NilAssetID = AssetID(uuid.Nil)

func NewAssetID() AssetID {
	return AssetID(uuid.New())
}

func ParseAssetID(s string) (AssetID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return NilAssetID, err
	}
	return AssetID(u), nil
}

func MustParseAssetID(s string) AssetID {
	return AssetID(uuid.MustParse(s))
}

func (id AssetID) String() string {
	return uuid.UUID(id).String()
}

func (id AssetID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id AssetID) IsNil() bool {
	return id == NilAssetID
}

func (id AssetID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

func (id *AssetID) UnmarshalText(data []byte) error {
	var u uuid.UUID
	if err := u.UnmarshalText(data); err != nil {
		return err
	}
	*id = AssetID(u)
	return nil
}

// ============================================================================
// InvestmentID
// ============================================================================

// InvestmentID uniquely identifies an investment aggregate across contexts.
type InvestmentID uuid.UUID

var NilInvestmentID = InvestmentID(uuid.Nil)

func NewInvestmentID() InvestmentID {
	return InvestmentID(uuid.New())
}

func ParseInvestmentID(s string) (InvestmentID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return NilInvestmentID, err
	}
	return InvestmentID(u), nil
}

func MustParseInvestmentID(s string) InvestmentID {
	return InvestmentID(uuid.MustParse(s))
}

func (id InvestmentID) String() string {
	return uuid.UUID(id).String()
}

func (id InvestmentID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id InvestmentID) IsNil() bool {
	return id == NilInvestmentID
}

func (id InvestmentID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

func (id *InvestmentID) UnmarshalText(data []byte) error {
	var u uuid.UUID
	if err := u.UnmarshalText(data); err != nil {
		return err
	}
	*id = InvestmentID(u)
	return nil
}

// ============================================================================
// InvitationID
// ============================================================================

// InvitationID uniquely identifies an invitation aggregate across contexts.
type InvitationID uuid.UUID

var NilInvitationID = InvitationID(uuid.Nil)

func NewInvitationID() InvitationID {
	return InvitationID(uuid.New())
}

func ParseInvitationID(s string) (InvitationID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return NilInvitationID, err
	}
	return InvitationID(u), nil
}

func MustParseInvitationID(s string) InvitationID {
	return InvitationID(uuid.MustParse(s))
}

func (id InvitationID) String() string {
	return uuid.UUID(id).String()
}

func (id InvitationID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id InvitationID) IsNil() bool {
	return id == NilInvitationID
}

func (id InvitationID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

func (id *InvitationID) UnmarshalText(data []byte) error {
	var u uuid.UUID
	if err := u.UnmarshalText(data); err != nil {
		return err
	}
	*id = InvitationID(u)
	return nil
}
