package people_test

import (
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
}
