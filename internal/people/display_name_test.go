package people_test

import (
	"encoding/json"
	"strings"
	"testing"

	"gitlab.com/massimo-ua/projecta/internal/people"
)

func TestDisplayName(t *testing.T) {
	t.Run("valid display name", func(t *testing.T) {
		dn, err := people.NewDisplayName("  Alice Smith  ")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dn.String() != "Alice Smith" {
			t.Errorf("expected trimmed name 'Alice Smith', got '%s'", dn.String())
		}
		if dn.IsEmpty() {
			t.Errorf("expected non-empty display name")
		}
	})

	t.Run("empty display name is valid and empty", func(t *testing.T) {
		dn, err := people.NewDisplayName("   ")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !dn.IsEmpty() {
			t.Errorf("expected empty display name")
		}
		if dn.String() != "" {
			t.Errorf("expected empty string, got '%s'", dn.String())
		}
	})

	t.Run("exceeds max length", func(t *testing.T) {
		longName := strings.Repeat("a", 256)
		_, err := people.NewDisplayName(longName)
		if err == nil {
			t.Errorf("expected error for display name exceeding 255 characters")
		}
	})

	t.Run("equals", func(t *testing.T) {
		dn1, _ := people.NewDisplayName("Bob")
		dn2, _ := people.NewDisplayName("Bob")
		dn3, _ := people.NewDisplayName("Alice")

		if !dn1.Equals(dn2) {
			t.Errorf("expected dn1 to equal dn2")
		}
		if dn1.Equals(dn3) {
			t.Errorf("expected dn1 to not equal dn3")
		}
	})

	t.Run("json unmarshal valid", func(t *testing.T) {
		jsonData := []byte(`"Charlie"`)
		var dn people.DisplayName
		err := json.Unmarshal(jsonData, &dn)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dn.String() != "Charlie" {
			t.Errorf("expected 'Charlie', got '%s'", dn.String())
		}
	})

	t.Run("json unmarshal invalid length", func(t *testing.T) {
		longName := strings.Repeat("x", 256)
		jsonData, _ := json.Marshal(longName)
		var dn people.DisplayName
		err := json.Unmarshal(jsonData, &dn)
		if err == nil {
			t.Errorf("expected error unmarshaling too long display name")
		}
	})

	t.Run("json marshal", func(t *testing.T) {
		dn, _ := people.NewDisplayName("David")
		b, err := json.Marshal(dn)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if string(b) != `"David"` {
			t.Errorf(`expected '"David"', got '%s'`, string(b))
		}
	})
}
