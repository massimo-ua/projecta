package asset_test

import (
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
)

func TestAssetGettersAndSetters(t *testing.T) {
	id := uuid.New()
	ownerID := uuid.New()
	owner := asset.OwnerRef{PersonID: ownerID, DisplayName: "John Doe"}
	projectID := uuid.New()
	now := time.Now()
	price := money.New(1000, money.USD)

	a := asset.NewAsset(id, "Laptop", "Work Laptop", projectID, price, now, owner)

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

	// Setters
	newProjectID := uuid.New()
	newPrice := money.New(2000, money.USD)
	newDate := now.Add(time.Hour)
	newOwner := asset.OwnerRef{PersonID: uuid.New(), DisplayName: "Jane Doe"}

	a.SetName("Desktop")
	a.SetDescription("Gaming Desktop")
	a.SetProjectID(newProjectID)
	a.SetPrice(newPrice)
	a.SetAcquiredAt(newDate)
	a.SetOwner(newOwner)

	if a.Name() != "Desktop" || a.Description() != "Gaming Desktop" {
		t.Errorf("Updated fields mismatch")
	}
	if a.ProjectID() != newProjectID || a.Price() != newPrice || !a.AcquiredAt().Equal(newDate) || a.Owner() != newOwner {
		t.Errorf("Updated references mismatch")
	}

	col := asset.NewCollection(5)
	if col.Total() != 5 {
		t.Errorf("Collection total mismatch")
	}
}
