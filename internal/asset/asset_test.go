package asset_test

import (
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

func TestAssetGettersAndSetters(t *testing.T) {
	id := uuid.New()
	ownerID := uuid.New()
	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "John Doe"}
	projectID := uuid.New()
	now := time.Now()
	project, _ := projecta.NewProject(projectID, "Test Project", "Desc", owner, now, now)
	costType, _ := projecta.NewCostType(projectID, nil, "Equipment", "Desc")
	price := money.New(1000, money.USD)

	a := asset.NewAsset(id, "Laptop", "Work Laptop", project, costType, price, now, owner)

	if a.ID() != id {
		t.Errorf("ID mismatch: got %v", a.ID())
	}
	if a.Name() != "Laptop" {
		t.Errorf("Name mismatch: got %s", a.Name())
	}
	if a.Description() != "Work Laptop" {
		t.Errorf("Description mismatch: got %s", a.Description())
	}
	if a.Project() != project {
		t.Errorf("Project mismatch")
	}
	if a.Type() != costType {
		t.Errorf("Type mismatch")
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
	newProject, _ := projecta.NewProject(uuid.New(), "New Project", "Desc", owner, now, now)
	newType, _ := projecta.NewCostType(newProject.ProjectID, nil, "Software", "Desc")
	newPrice := money.New(2000, money.USD)
	newDate := now.Add(time.Hour)
	newOwner := &projecta.Owner{PersonID: uuid.New(), DisplayName: "Jane Doe"}

	a.SetName("Desktop")
	a.SetDescription("Gaming Desktop")
	a.SetProject(newProject)
	a.SetType(newType)
	a.SetPrice(newPrice)
	a.SetAcquiredAt(newDate)
	a.SetOwner(newOwner)

	if a.Name() != "Desktop" || a.Description() != "Gaming Desktop" {
		t.Errorf("Updated fields mismatch")
	}
	if a.Project() != newProject || a.Type() != newType || a.Price() != newPrice || !a.AcquiredAt().Equal(newDate) || a.Owner() != newOwner {
		t.Errorf("Updated references mismatch")
	}

	col := asset.NewCollection(5)
	if col.Total() != 5 {
		t.Errorf("Collection total mismatch")
	}
}
