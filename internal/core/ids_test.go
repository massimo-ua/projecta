package core_test

import (
	"bytes"
	"encoding/json"
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

	// SQL driver Value and Scan
	val, err := id.Value()
	if err != nil || val != raw.String() {
		t.Errorf("Value() failed: %v, %v", val, err)
	}

	var scanned core.PersonID
	if err := scanned.Scan(raw.String()); err != nil || scanned != id {
		t.Errorf("Scan() from string failed: %v, %v", scanned, err)
	}
	if err := scanned.Scan(raw[:]); err != nil || scanned != id {
		t.Errorf("Scan() from bytes failed: %v, %v", scanned, err)
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

	// JSON marshaling
	data, err := json.Marshal(id)
	if err != nil || !bytes.Contains(data, []byte(raw.String())) {
		t.Errorf("json.Marshal failed: %s, %v", string(data), err)
	}
	var unmarshaledJSON core.PersonID
	if err := json.Unmarshal(data, &unmarshaledJSON); err != nil || unmarshaledJSON != id {
		t.Errorf("json.Unmarshal failed: %v, %v", unmarshaledJSON, err)
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

	// SQL Value & Scan
	val, err := id.Value()
	if err != nil || val != raw.String() {
		t.Errorf("Value() failed")
	}
	var scanned core.ProjectID
	if err := scanned.Scan(raw.String()); err != nil || scanned != id {
		t.Errorf("Scan() failed")
	}

	// Text & JSON
	text, err := id.MarshalText()
	if err != nil || string(text) != raw.String() {
		t.Errorf("MarshalText failed")
	}
	var unmarshaledText core.ProjectID
	if err := unmarshaledText.UnmarshalText(text); err != nil || unmarshaledText != id {
		t.Errorf("UnmarshalText failed")
	}

	data, err := json.Marshal(id)
	if err != nil {
		t.Errorf("Marshal failed")
	}
	var unmarshaledJSON core.ProjectID
	if err := json.Unmarshal(data, &unmarshaledJSON); err != nil || unmarshaledJSON != id {
		t.Errorf("Unmarshal failed")
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

	// SQL Value & Scan
	val, err := id.Value()
	if err != nil || val != raw.String() {
		t.Errorf("Value() failed")
	}
	var scanned core.AssetID
	if err := scanned.Scan(raw.String()); err != nil || scanned != id {
		t.Errorf("Scan() failed")
	}

	// Text & JSON
	text, err := id.MarshalText()
	if err != nil || string(text) != raw.String() {
		t.Errorf("MarshalText failed")
	}
	var unmarshaledText core.AssetID
	if err := unmarshaledText.UnmarshalText(text); err != nil || unmarshaledText != id {
		t.Errorf("UnmarshalText failed")
	}

	data, err := json.Marshal(id)
	if err != nil {
		t.Errorf("Marshal failed")
	}
	var unmarshaledJSON core.AssetID
	if err := json.Unmarshal(data, &unmarshaledJSON); err != nil || unmarshaledJSON != id {
		t.Errorf("Unmarshal failed")
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

	// SQL Value & Scan
	val, err := id.Value()
	if err != nil || val != raw.String() {
		t.Errorf("Value() failed")
	}
	var scanned core.InvestmentID
	if err := scanned.Scan(raw.String()); err != nil || scanned != id {
		t.Errorf("Scan() failed")
	}

	// Text & JSON
	text, err := id.MarshalText()
	if err != nil || string(text) != raw.String() {
		t.Errorf("MarshalText failed")
	}
	var unmarshaledText core.InvestmentID
	if err := unmarshaledText.UnmarshalText(text); err != nil || unmarshaledText != id {
		t.Errorf("UnmarshalText failed")
	}

	data, err := json.Marshal(id)
	if err != nil {
		t.Errorf("Marshal failed")
	}
	var unmarshaledJSON core.InvestmentID
	if err := json.Unmarshal(data, &unmarshaledJSON); err != nil || unmarshaledJSON != id {
		t.Errorf("Unmarshal failed")
	}
}
