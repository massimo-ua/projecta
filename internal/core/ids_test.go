package core_test

import (
	"testing"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

func TestPersonID(t *testing.T) {
	raw := uuid.New()
	id := core.PersonID(raw)

	if id.String() != raw.String() {
		t.Errorf("expected string %s, got %s", raw.String(), id.String())
	}
	if id.UUID() != raw {
		t.Errorf("expected UUID %s, got %s", raw, id.UUID())
	}
	if id.IsNil() {
		t.Errorf("expected non-nil ID")
	}
	if !core.NilPersonID.IsNil() {
		t.Errorf("expected NilPersonID to be nil")
	}

	newID := core.NewPersonID()
	if newID.IsNil() {
		t.Errorf("expected generated ID to be non-nil")
	}

	parsed, err := core.ParsePersonID(raw.String())
	if err != nil || parsed != id {
		t.Errorf("ParsePersonID failed: %v, %v", parsed, err)
	}

	if _, err := core.ParsePersonID("invalid"); err == nil {
		t.Errorf("expected error on invalid UUID string")
	}

	mustParsed := core.MustParsePersonID(raw.String())
	if mustParsed != id {
		t.Errorf("MustParsePersonID failed")
	}

	// Text marshaling
	text, err := id.MarshalText()
	if err != nil || string(text) != raw.String() {
		t.Errorf("MarshalText failed: %s, %v", string(text), err)
	}
	var unmarshaledText core.PersonID
	if err := unmarshaledText.UnmarshalText(text); err != nil || unmarshaledText != id {
		t.Errorf("UnmarshalText failed: %v, %v", unmarshaledText, err)
	}
}

func TestProjectID(t *testing.T) {
	raw := uuid.New()
	id := core.ProjectID(raw)

	if id.String() != raw.String() || id.UUID() != raw || id.IsNil() || !core.NilProjectID.IsNil() {
		t.Errorf("basic ProjectID methods failed")
	}

	newID := core.NewProjectID()
	if newID.IsNil() {
		t.Errorf("expected non-nil NewProjectID")
	}

	parsed, err := core.ParseProjectID(raw.String())
	if err != nil || parsed != id {
		t.Errorf("ParseProjectID failed")
	}
	if _, err := core.ParseProjectID("invalid"); err == nil {
		t.Errorf("expected error on invalid string")
	}
	if core.MustParseProjectID(raw.String()) != id {
		t.Errorf("MustParseProjectID failed")
	}

	// Text marshaling
	text, err := id.MarshalText()
	if err != nil || string(text) != raw.String() {
		t.Errorf("MarshalText failed")
	}
	var unmarshaledText core.ProjectID
	if err := unmarshaledText.UnmarshalText(text); err != nil || unmarshaledText != id {
		t.Errorf("UnmarshalText failed")
	}
}

func TestAssetID(t *testing.T) {
	raw := uuid.New()
	id := core.AssetID(raw)

	if id.String() != raw.String() || id.UUID() != raw || id.IsNil() || !core.NilAssetID.IsNil() {
		t.Errorf("basic AssetID methods failed")
	}

	newID := core.NewAssetID()
	if newID.IsNil() {
		t.Errorf("expected non-nil NewAssetID")
	}

	parsed, err := core.ParseAssetID(raw.String())
	if err != nil || parsed != id {
		t.Errorf("ParseAssetID failed")
	}
	if _, err := core.ParseAssetID("invalid"); err == nil {
		t.Errorf("expected error on invalid string")
	}
	if core.MustParseAssetID(raw.String()) != id {
		t.Errorf("MustParseAssetID failed")
	}

	// Text marshaling
	text, err := id.MarshalText()
	if err != nil || string(text) != raw.String() {
		t.Errorf("MarshalText failed")
	}
	var unmarshaledText core.AssetID
	if err := unmarshaledText.UnmarshalText(text); err != nil || unmarshaledText != id {
		t.Errorf("UnmarshalText failed")
	}
}

func TestInvestmentID(t *testing.T) {
	raw := uuid.New()
	id := core.InvestmentID(raw)

	if id.String() != raw.String() || id.UUID() != raw || id.IsNil() || !core.NilInvestmentID.IsNil() {
		t.Errorf("basic InvestmentID methods failed")
	}

	newID := core.NewInvestmentID()
	if newID.IsNil() {
		t.Errorf("expected non-nil NewInvestmentID")
	}

	parsed, err := core.ParseInvestmentID(raw.String())
	if err != nil || parsed != id {
		t.Errorf("ParseInvestmentID failed")
	}
	if _, err := core.ParseInvestmentID("invalid"); err == nil {
		t.Errorf("expected error on invalid string")
	}
	if core.MustParseInvestmentID(raw.String()) != id {
		t.Errorf("MustParseInvestmentID failed")
	}

	// Text marshaling
	text, err := id.MarshalText()
	if err != nil || string(text) != raw.String() {
		t.Errorf("MarshalText failed")
	}
	var unmarshaledText core.InvestmentID
	if err := unmarshaledText.UnmarshalText(text); err != nil || unmarshaledText != id {
		t.Errorf("UnmarshalText failed")
	}
}

func TestInvitationID(t *testing.T) {
	raw := uuid.New()
	id := core.InvitationID(raw)

	if id.String() != raw.String() || id.UUID() != raw || id.IsNil() || !core.NilInvitationID.IsNil() {
		t.Errorf("basic InvitationID methods failed")
	}

	newID := core.NewInvitationID()
	if newID.IsNil() {
		t.Errorf("expected non-nil NewInvitationID")
	}

	parsed, err := core.ParseInvitationID(raw.String())
	if err != nil || parsed != id {
		t.Errorf("ParseInvitationID failed")
	}
	if _, err := core.ParseInvitationID("invalid"); err == nil {
		t.Errorf("expected error on invalid string")
	}
	if core.MustParseInvitationID(raw.String()) != id {
		t.Errorf("MustParseInvitationID failed")
	}

	// Text marshaling
	text, err := id.MarshalText()
	if err != nil || string(text) != raw.String() {
		t.Errorf("MarshalText failed")
	}
	var unmarshaledText core.InvitationID
	if err := unmarshaledText.UnmarshalText(text); err != nil || unmarshaledText != id {
		t.Errorf("UnmarshalText failed")
	}
}
