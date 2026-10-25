package project_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

func TestProject_NewProject(t *testing.T) {
	ownerID := core.NewPersonID()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "Alice"}
	now := time.Now()

	t.Run("creates project successfully with valid parameters", func(t *testing.T) {
		id := core.NewProjectID()
		proj, err := project.NewProject(id, "Alpha Project", "Alpha Description", owner, now, now.Add(24*time.Hour), "EUR")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proj.ID() != id || proj.ProjectID() != id {
			t.Errorf("expected ID %v, got %v", id, proj.ID())
		}
		if proj.Name() != "Alpha Project" {
			t.Errorf("expected name 'Alpha Project', got %s", proj.Name())
		}
		if proj.Description() != "Alpha Description" {
			t.Errorf("expected description 'Alpha Description', got %s", proj.Description())
		}
		if proj.MainCurrency() != "EUR" {
			t.Errorf("expected currency 'EUR', got %s", proj.MainCurrency())
		}
		if proj.OwnerID() != ownerID {
			t.Errorf("expected owner ID %v, got %v", ownerID, proj.OwnerID())
		}
		if proj.ShareToken() == uuid.Nil {
			t.Errorf("expected non-nil share token")
		}
		if proj.IsShared() {
			t.Errorf("expected isShared to be false initially")
		}
		if len(proj.Participants()) != 1 {
			t.Errorf("expected 1 participant (owner), got %d", len(proj.Participants()))
		}
	})

	t.Run("defaults currency to UAH when omitted", func(t *testing.T) {
		proj, err := project.NewProject(core.NewProjectID(), "UAH Project", "Desc", owner, now, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proj.MainCurrency() != "UAH" {
			t.Errorf("expected default currency 'UAH', got %s", proj.MainCurrency())
		}
	})

	t.Run("rejects project name shorter than 3 characters", func(t *testing.T) {
		_, err := project.NewProject(core.NewProjectID(), "ab", "Desc", owner, now, now)
		if err == nil {
			t.Errorf("expected error for short name, got nil")
		}
	})

	t.Run("rejects project name longer than 100 characters", func(t *testing.T) {
		longName := string(make([]byte, 101))
		_, err := project.NewProject(core.NewProjectID(), longName, "Desc", owner, now, now)
		if err == nil {
			t.Errorf("expected error for long name, got nil")
		}
	})

	t.Run("rejects unsupported currency", func(t *testing.T) {
		_, err := project.NewProject(core.NewProjectID(), "Valid Name", "Desc", owner, now, now, "XYZ")
		if err == nil {
			t.Errorf("expected error for unsupported currency, got nil")
		}
	})

	t.Run("rejects end date before start date", func(t *testing.T) {
		_, err := project.NewProject(core.NewProjectID(), "Valid Name", "Desc", owner, now, now.Add(-24*time.Hour))
		if err == nil {
			t.Errorf("expected error when end date is before start date, got nil")
		}
	})
}

func TestProject_UpdateDetails(t *testing.T) {
	owner := &project.Owner{PersonID: core.NewPersonID(), DisplayName: "Alice"}
	proj, err := project.NewProject(core.NewProjectID(), "Original Name", "Original Desc", owner, time.Now(), time.Now(), "UAH")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("updates details successfully", func(t *testing.T) {
		err := proj.UpdateDetails("Updated Name", "Updated Desc", "usd")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proj.Name() != "Updated Name" {
			t.Errorf("expected name 'Updated Name', got %s", proj.Name())
		}
		if proj.Description() != "Updated Desc" {
			t.Errorf("expected description 'Updated Desc', got %s", proj.Description())
		}
		if proj.MainCurrency() != "USD" {
			t.Errorf("expected normalized currency 'USD', got %s", proj.MainCurrency())
		}
	})

	t.Run("rejects invalid updated name", func(t *testing.T) {
		err := proj.UpdateDetails("x", "", "")
		if err == nil {
			t.Errorf("expected error for name too short, got nil")
		}
	})

	t.Run("rejects invalid updated currency", func(t *testing.T) {
		err := proj.UpdateDetails("", "", "BITCOIN")
		if err == nil {
			t.Errorf("expected error for unsupported currency, got nil")
		}
	})
}

func TestProject_SetDates(t *testing.T) {
	owner := &project.Owner{PersonID: core.NewPersonID(), DisplayName: "Alice"}
	proj, _ := project.NewProject(core.NewProjectID(), "Test Proj", "Desc", owner, time.Now(), time.Now())

	start := time.Now()
	end := start.Add(48 * time.Hour)

	if err := proj.SetDates(start, end); err != nil {
		t.Fatalf("unexpected error setting valid dates: %v", err)
	}
	if !proj.StartDate().Equal(start) || !proj.EndDate().Equal(end) {
		t.Errorf("dates not updated properly")
	}

	if err := proj.SetDates(end, start); err == nil {
		t.Errorf("expected error when end date is before start date, got nil")
	}
}

func TestProject_ParticipantsAndDefensiveCopy(t *testing.T) {
	owner := &project.Owner{PersonID: core.NewPersonID(), DisplayName: "Alice"}
	proj, _ := project.NewProject(core.NewProjectID(), "Test Proj", "Desc", owner, time.Now(), time.Now())

	p1, _ := project.NewParticipant("Bob")
	proj.AddParticipant(p1)

	// Add duplicate Bob
	p1Dup, _ := project.NewParticipant("Bob")
	proj.AddParticipant(p1Dup)

	participants := proj.Participants()
	if len(participants) != 2 { // Alice (owner) + Bob
		t.Fatalf("expected 2 participants, got %d", len(participants))
	}

	// Verify slice copy defensiveness (mutating returned slice does not alter project state)
	p2, _ := project.NewParticipant("Charlie")
	participants[0] = p2

	fresh := proj.Participants()
	if fresh[0].DisplayName() == "Charlie" {
		t.Errorf("internal participants mutated through external slice modification")
	}
}

func TestProject_ShareTokenAndOwnership(t *testing.T) {
	ownerID := core.NewPersonID()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "Alice"}
	proj, _ := project.NewProject(core.NewProjectID(), "Test Proj", "Desc", owner, time.Now(), time.Now())

	initialToken := proj.ShareToken()
	proj.RegenerateShareToken()
	if proj.ShareToken() == initialToken || proj.ShareToken() == uuid.Nil {
		t.Errorf("share token was not regenerated properly")
	}

	if !proj.IsOwnedBy(owner) {
		t.Errorf("expected IsOwnedBy true for owner")
	}
	if !proj.IsOwnedByPerson(ownerID) {
		t.Errorf("expected IsOwnedByPerson true for ownerID")
	}
	if proj.IsOwnedByPerson(core.NewPersonID()) {
		t.Errorf("expected IsOwnedByPerson false for other person")
	}

	proj.MarkShared()
	if !proj.IsShared() {
		t.Errorf("expected IsShared true after MarkShared")
	}
}

func TestProject_ReconstituteProject(t *testing.T) {
	id := core.NewProjectID()
	ownerID := core.NewPersonID()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "Alice"}
	shareToken := uuid.New()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	p1, _ := project.NewParticipant("Alice")
	p2, _ := project.NewParticipant("Bob")

	proj := project.ReconstituteProject(
		id,
		"Hydrated Project",
		"Hydrated Description",
		owner,
		start,
		end,
		shareToken,
		true,
		"EUR",
		[]project.Participant{p1, p2},
	)

	if proj.ID() != id || proj.ProjectID() != id {
		t.Errorf("expected ID %v, got %v", id, proj.ID())
	}
	if proj.Name() != "Hydrated Project" || proj.Description() != "Hydrated Description" {
		t.Errorf("unexpected name or description")
	}
	if proj.OwnerID() != ownerID {
		t.Errorf("expected owner ID %v, got %v", ownerID, proj.OwnerID())
	}
	if proj.ShareToken() != shareToken {
		t.Errorf("expected shareToken %v, got %v", shareToken, proj.ShareToken())
	}
	if !proj.IsShared() {
		t.Errorf("expected isShared to be true")
	}
	if proj.MainCurrency() != "EUR" {
		t.Errorf("expected currency EUR, got %s", proj.MainCurrency())
	}
	if len(proj.Participants()) != 2 {
		t.Errorf("expected 2 participants, got %d", len(proj.Participants()))
	}

	t.Run("Owner getter returns defensive copy", func(t *testing.T) {
		gotOwner := proj.Owner()
		if gotOwner == nil {
			t.Fatalf("expected non-nil owner")
		}
		gotOwner.DisplayName = "Mutated DisplayName"
		if proj.Owner().DisplayName == "Mutated DisplayName" {
			t.Errorf("Project.Owner() leaked internal pointer")
		}

		// Mutate original input owner
		owner.DisplayName = "Input Mutated"
		if proj.Owner().DisplayName == "Input Mutated" {
			t.Errorf("ReconstituteProject leaked input owner pointer")
		}
	})
}
