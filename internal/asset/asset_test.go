package asset_test

import (
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

func TestAssetGettersAndSetters(t *testing.T) {
	id := core.NewAssetID()
	ownerID := core.NewPersonID()
	owner := asset.OwnerRef{PersonID: ownerID, DisplayName: "John Doe"}
	projectID := core.NewProjectID()
	now := time.Now()
	price := money.New(1000, money.USD)

	a, err := asset.CreateAsset(id, "Laptop", "Work Laptop", projectID, price, now, owner, asset.AssetStatusActive, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error creating asset: %v", err)
	}

	if a.ID() != id {
		t.Errorf("ID mismatch: got %v", a.ID())
	}
	if a.Name() != "Laptop" {
		t.Errorf("Name mismatch: got %s", a.Name())
	}
	if a.Description() != "Work Laptop" {
		t.Errorf("Description mismatch: got %s", a.Description())
	}
	if a.ProjectID() != projectID {
		t.Errorf("ProjectID mismatch")
	}
	if a.Price() != price {
		t.Errorf("Price mismatch")
	}
	if !a.AcquiredAt().Equal(now) {
		t.Errorf("AcquiredAt mismatch")
	}
	if a.Owner() != owner {
		t.Errorf("Owner mismatch")
	}

	// Domain behavior methods
	newPrice := money.New(2000, money.USD)
	newDate := now.Add(time.Hour)

	if err := a.UpdateDetails("Desktop", "Gaming Desktop", newPrice, []string{"hardware"}); err != nil {
		t.Fatalf("unexpected error updating details: %v", err)
	}
	a.Reschedule(newDate)
	a.TransitionStatus(asset.AssetStatusCompleted, &newDate)

	if a.Name() != "Desktop" || a.Description() != "Gaming Desktop" {
		t.Errorf("Updated fields mismatch")
	}
	if a.TargetPrice() != newPrice || !a.AcquiredAt().Equal(newDate) || a.Status() != asset.AssetStatusCompleted {
		t.Errorf("Updated references mismatch")
	}
	if len(a.Tags()) != 1 || a.Tags()[0] != "hardware" {
		t.Errorf("Updated tags mismatch")
	}

	a.Reopen()
	if a.Status() != asset.AssetStatusActive || a.CompletedDate() != nil {
		t.Errorf("Expected asset to reopen")
	}

	col := asset.NewCollection(5)
	if col.Total() != 5 {
		t.Errorf("Collection total mismatch")
	}

	t.Run("defensive copying on getters and mutators", func(t *testing.T) {
		// Tags defensive copy
		tags := a.Tags()
		tags[0] = "mutated-tag"
		if a.Tags()[0] == "mutated-tag" {
			t.Errorf("Tags getter leaked internal slice pointer")
		}

		// UpdateDetails tags input defensive copy
		inputTags := []string{"original"}
		_ = a.UpdateDetails("Desktop", "Desc", newPrice, inputTags)
		inputTags[0] = "mutated-input"
		if a.Tags()[0] == "mutated-input" {
			t.Errorf("UpdateDetails leaked input tags slice")
		}

		// Children and Parents defensive copy
		childID := core.NewAssetID()
		parentID := core.NewAssetID()
		a.AssignValuation(newPrice, newPrice, []asset.ChildAssetLink{{ChildID: childID, SharePercentage: 100}}, []asset.ParentAssetLink{{ParentID: parentID, SharePercentage: 100}})
		children := a.Children()
		children[0].SharePercentage = 50
		if a.Children()[0].SharePercentage == 50 {
			t.Errorf("Children getter leaked internal slice pointer")
		}
		parents := a.Parents()
		parents[0].SharePercentage = 50
		if a.Parents()[0].SharePercentage == 50 {
			t.Errorf("Parents getter leaked internal slice pointer")
		}
	})

	t.Run("UpdateDetails validation errors", func(t *testing.T) {
		if err := a.UpdateDetails("", "Desc", newPrice, nil); err == nil {
			t.Errorf("expected error for empty name")
		}
		if err := a.UpdateDetails("  ", "Desc", newPrice, nil); err == nil {
			t.Errorf("expected error for whitespace name")
		}
		if err := a.UpdateDetails("Valid", "Desc", money.New(-100, "USD"), nil); err == nil {
			t.Errorf("expected error for negative price")
		}
		if err := a.UpdateDetails("Valid", "Desc", money.New(0, "USD"), nil); err == nil {
			t.Errorf("expected error for zero price")
		}
	})
}

func TestAsset_PointerImmutability(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(24 * time.Hour)
	ast, err := asset.CreateAsset(
		core.NewAssetID(),
		"Test Asset",
		"Desc",
		core.NewProjectID(),
		money.New(10000, "UAH"),
		now,
		asset.OwnerRef{PersonID: core.NewPersonID()},
		asset.AssetStatusCompleted,
		&completedAt,
		[]string{"tag1"},
	)
	if err != nil {
		t.Fatalf("CreateAsset failed: %v", err)
	}

	// Mutate completed date
	cd := ast.CompletedDate()
	if cd == nil {
		t.Fatalf("expected non-nil completed date")
	}
	*cd = cd.Add(100 * time.Hour)
	if ast.CompletedDate().Equal(*cd) {
		t.Errorf("CompletedDate pointer mutation leaked into asset")
	}

	// Mutate target home amount
	tha := ast.TargetHomeAmount()
	if tha != nil {
		orig := *tha
		*tha = orig + 99999
		if *ast.TargetHomeAmount() == *tha {
			t.Errorf("TargetHomeAmount pointer mutation leaked into asset")
		}
	}

	// Mutate progress percentage
	pp := ast.ProgressPercentage()
	if pp != nil {
		orig := *pp
		*pp = orig + 50.0
		if *ast.ProgressPercentage() == *pp {
			t.Errorf("ProgressPercentage pointer mutation leaked into asset")
		}
	}
}
