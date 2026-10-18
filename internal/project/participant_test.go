package project_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

func TestParticipantValueObject(t *testing.T) {
	t.Run("valid participant creation", func(t *testing.T) {
		p, err := project.NewParticipant("  John Doe  ")
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
		_, err := project.NewParticipant("")
		if err == nil {
			t.Error("expected error for empty display name")
		}

		_, err = project.NewParticipant("   ")
		if err == nil {
			t.Error("expected error for whitespace display name")
		}
	})

	t.Run("display name exceeds max length", func(t *testing.T) {
		longName := strings.Repeat("a", 256)
		_, err := project.NewParticipant(longName)
		if err == nil {
			t.Error("expected error for display name exceeding 255 chars")
		}

		exactName := strings.Repeat("a", 255)
		p, err := project.NewParticipant(exactName)
		if err != nil {
			t.Fatalf("unexpected error for 255 char name: %v", err)
		}
		if p.DisplayName() != exactName {
			t.Errorf("expected %s, got %s", exactName, p.DisplayName())
		}
	})

	t.Run("equality", func(t *testing.T) {
		p1, _ := project.NewParticipant("Alice")
		p2, _ := project.NewParticipant("Alice")
		p3, _ := project.NewParticipant("Bob")

		if !p1.Equals(p2) {
			t.Error("expected p1 to equal p2")
		}
		if p1.Equals(p3) {
			t.Error("expected p1 not to equal p3")
		}
	})
}

func TestProjectParticipants(t *testing.T) {
	ownerID := uuid.New()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "Owner User"}
	now := time.Now()

	t.Run("new project initializes participants with owner", func(t *testing.T) {
		proj, err := project.NewProject(uuid.New(), "Project With Owner", "Desc", owner, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(proj.Participants()) != 1 {
			t.Fatalf("expected 1 participant, got %d", len(proj.Participants()))
		}
		if proj.Participants()[0].DisplayName() != "Owner User" {
			t.Errorf("expected 'Owner User', got '%s'", proj.Participants()[0].DisplayName())
		}
	})

	t.Run("new project without owner or empty display name", func(t *testing.T) {
		proj, err := project.NewProject(uuid.New(), "Project No Owner", "Desc", nil, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(proj.Participants()) != 0 {
			t.Errorf("expected 0 participants, got %d", len(proj.Participants()))
		}

		emptyOwner := &project.Owner{PersonID: ownerID, DisplayName: ""}
		proj2, err := project.NewProject(uuid.New(), "Project Empty Owner", "Desc", emptyOwner, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(proj2.Participants()) != 0 {
			t.Errorf("expected 0 participants, got %d", len(proj2.Participants()))
		}
	})

	t.Run("add participant and prevent duplicates", func(t *testing.T) {
		proj, err := project.NewProject(uuid.New(), "Project Multi", "Desc", owner, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		p1, _ := project.NewParticipant("Team Member 1")
		p2, _ := project.NewParticipant("Team Member 2")

		proj.AddParticipant(p1)
		proj.AddParticipant(p2)

		if len(proj.Participants()) != 3 {
			t.Fatalf("expected 3 participants, got %d", len(proj.Participants()))
		}

		// Try to add duplicate
		duplicateP1, _ := project.NewParticipant("Team Member 1")
		proj.AddParticipant(duplicateP1)

		if len(proj.Participants()) != 3 {
			t.Errorf("expected 3 participants after duplicate add, got %d", len(proj.Participants()))
		}
	})
}
