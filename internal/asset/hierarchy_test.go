package asset_test

import (
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
)

func makeTestAsset(name string) *asset.Asset {
	id := uuid.New()
	owner := asset.OwnerRef{PersonID: uuid.New(), DisplayName: "User"}
	projectID := uuid.New()
	return asset.NewAsset(id, name, "", projectID, money.New(0, "UAH"), time.Now(), owner)
}

func TestHierarchy_SelfReferenceAndInvalidShare(t *testing.T) {
	h := asset.NewAssetHierarchy()
	id := uuid.New()

	if err := h.AddEdge(id, id, 100.0); err != asset.ErrSelfReference {
		t.Fatalf("expected ErrSelfReference, got %v", err)
	}

	otherID := uuid.New()
	if err := h.AddEdge(id, otherID, 0.0); err != asset.ErrInvalidShare {
		t.Fatalf("expected ErrInvalidShare for 0, got %v", err)
	}
	if err := h.AddEdge(id, otherID, 101.0); err != asset.ErrInvalidShare {
		t.Fatalf("expected ErrInvalidShare for 101, got %v", err)
	}
}

func TestHierarchy_CycleDetection(t *testing.T) {
	h := asset.NewAssetHierarchy()
	a := makeTestAsset("A")
	b := makeTestAsset("B")
	c := makeTestAsset("C")

	h.AddAsset(a)
	h.AddAsset(b)
	h.AddAsset(c)

	// A -> B -> C
	if err := h.AddEdge(a.ID(), b.ID(), 100.0); err != nil {
		t.Fatalf("unexpected error adding A -> B: %v", err)
	}
	if err := h.AddEdge(b.ID(), c.ID(), 100.0); err != nil {
		t.Fatalf("unexpected error adding B -> C: %v", err)
	}

	// Linking C -> A must be rejected due to cycle
	if err := h.AddEdge(c.ID(), a.ID(), 100.0); err != asset.ErrCircularDependency {
		t.Fatalf("expected ErrCircularDependency, got %v", err)
	}
}

func TestHierarchy_CalculateRollupsAndApply(t *testing.T) {
	h := asset.NewAssetHierarchy()
	a := makeTestAsset("A")
	b := makeTestAsset("B")
	c := makeTestAsset("C")

	h.AddAsset(a)
	h.AddAsset(b)
	h.AddAsset(c)

	// A -> B (100%), B -> C (50%)
	if err := h.AddEdge(a.ID(), b.ID(), 100.0); err != nil {
		t.Fatal(err)
	}
	if err := h.AddEdge(b.ID(), c.ID(), 50.0); err != nil {
		t.Fatal(err)
	}

	directCosts := map[uuid.UUID]*money.Money{
		c.ID(): money.New(1000, "UAH"),
		b.ID(): money.New(200, "UAH"),
		a.ID(): money.New(300, "UAH"),
	}

	h.ApplyToAssets(directCosts, "UAH")

	// Verify C
	if c.DirectCost().Amount() != 1000 || c.TotalCost().Amount() != 1000 {
		t.Errorf("C: direct %d, total %d (expected 1000, 1000)", c.DirectCost().Amount(), c.TotalCost().Amount())
	}

	// Verify B: 200 + 50% of 1000 = 700
	if b.DirectCost().Amount() != 200 || b.TotalCost().Amount() != 700 {
		t.Errorf("B: direct %d, total %d (expected 200, 700)", b.DirectCost().Amount(), b.TotalCost().Amount())
	}
	if len(b.Children()) != 1 || b.Children()[0].TotalCost.Amount() != 1000 {
		t.Errorf("B children mismatch: %+v", b.Children())
	}

	// Verify A: 300 + 100% of 700 = 1000
	if a.DirectCost().Amount() != 300 || a.TotalCost().Amount() != 1000 {
		t.Errorf("A: direct %d, total %d (expected 300, 1000)", a.DirectCost().Amount(), a.TotalCost().Amount())
	}
	if len(a.Children()) != 1 || a.Children()[0].TotalCost.Amount() != 700 {
		t.Errorf("A children mismatch: %+v", a.Children())
	}

	// Verify Root Assets (only A is a root)
	roots := h.RootAssets()
	if len(roots) != 1 || roots[0].ID() != a.ID() {
		t.Errorf("expected only A as root, got %d roots", len(roots))
	}
}
