package projecta_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

func TestParticipantValueObject(t *testing.T) {
	t.Run("valid participant creation", func(t *testing.T) {
		p, err := projecta.NewParticipant("  John Doe  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.DisplayName() != "John Doe" {
			t.Errorf("expected 'John Doe', got '%s'", p.DisplayName())
		}
		if p.String() != "John Doe" {
			t.Errorf("expected 'John Doe', got '%s'", p.String())
		}
	})

	t.Run("empty display name", func(t *testing.T) {
		_, err := projecta.NewParticipant("")
		if err == nil {
			t.Error("expected error for empty display name")
		}

		_, err = projecta.NewParticipant("   ")
		if err == nil {
			t.Error("expected error for whitespace display name")
		}
	})

	t.Run("display name exceeds max length", func(t *testing.T) {
		longName := strings.Repeat("a", 256)
		_, err := projecta.NewParticipant(longName)
		if err == nil {
			t.Error("expected error for display name exceeding 255 chars")
		}

		exactName := strings.Repeat("a", 255)
		p, err := projecta.NewParticipant(exactName)
		if err != nil {
			t.Fatalf("unexpected error for 255 char name: %v", err)
		}
		if p.DisplayName() != exactName {
			t.Errorf("expected %s, got %s", exactName, p.DisplayName())
		}
	})

	t.Run("equality", func(t *testing.T) {
		p1, _ := projecta.NewParticipant("Alice")
		p2, _ := projecta.NewParticipant("Alice")
		p3, _ := projecta.NewParticipant("Bob")

		if !p1.Equals(p2) {
			t.Error("expected p1 to equal p2")
		}
		if p1.Equals(p3) {
			t.Error("expected p1 not to equal p3")
		}
	})

	t.Run("json marshaler and unmarshaler", func(t *testing.T) {
		p, err := projecta.NewParticipant("Alice Smith")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Marshal
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("failed to marshal participant: %v", err)
		}
		expectedJSON := `"Alice Smith"`
		if string(data) != expectedJSON {
			t.Errorf("expected %s, got %s", expectedJSON, string(data))
		}

		// Unmarshal
		var unmarshaled projecta.Participant
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal participant: %v", err)
		}
		if !unmarshaled.Equals(p) {
			t.Errorf("expected %v, got %v", p, unmarshaled)
		}

		// Unmarshal invalid
		var invalid projecta.Participant
		if err := json.Unmarshal([]byte(`""`), &invalid); err == nil {
			t.Error("expected error unmarshaling empty string into participant")
		}
		if err := json.Unmarshal([]byte(`123`), &invalid); err == nil {
			t.Error("expected error unmarshaling non-string into participant")
		}
	})
}

func TestProjectParticipants(t *testing.T) {
	ownerID := uuid.New()
	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "Owner User"}
	now := time.Now()

	t.Run("new project initializes participants with owner", func(t *testing.T) {
		proj, err := projecta.NewProject(uuid.New(), "Project With Owner", "Desc", owner, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(proj.Participants) != 1 {
			t.Fatalf("expected 1 participant, got %d", len(proj.Participants))
		}
		if proj.Participants[0].DisplayName() != "Owner User" {
			t.Errorf("expected 'Owner User', got '%s'", proj.Participants[0].DisplayName())
		}
	})

	t.Run("new project without owner or empty display name", func(t *testing.T) {
		proj, err := projecta.NewProject(uuid.New(), "Project No Owner", "Desc", nil, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(proj.Participants) != 0 {
			t.Errorf("expected 0 participants, got %d", len(proj.Participants))
		}

		emptyOwner := &projecta.Owner{PersonID: ownerID, DisplayName: ""}
		proj2, err := projecta.NewProject(uuid.New(), "Project Empty Owner", "Desc", emptyOwner, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(proj2.Participants) != 0 {
			t.Errorf("expected 0 participants, got %d", len(proj2.Participants))
		}
	})

	t.Run("add participant and prevent duplicates", func(t *testing.T) {
		proj, err := projecta.NewProject(uuid.New(), "Project Multi", "Desc", owner, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		p1, _ := projecta.NewParticipant("Team Member 1")
		p2, _ := projecta.NewParticipant("Team Member 2")

		proj.AddParticipant(p1)
		proj.AddParticipant(p2)

		if len(proj.Participants) != 3 {
			t.Fatalf("expected 3 participants, got %d", len(proj.Participants))
		}

		// Try to add duplicate
		duplicateP1, _ := projecta.NewParticipant("Team Member 1")
		proj.AddParticipant(duplicateP1)

		if len(proj.Participants) != 3 {
			t.Errorf("expected 3 participants after duplicate add, got %d", len(proj.Participants))
		}
	})
}
