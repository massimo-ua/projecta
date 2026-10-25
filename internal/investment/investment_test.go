package investment_test

import (
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

func TestInvestment_NewAndGetters(t *testing.T) {
	projID := core.NewProjectID()
	ownerID := core.NewPersonID()
	assetID := core.NewAssetID()
	now := time.Now()

	contributor := investment.ContributorRef{PersonID: ownerID, DisplayName: "John Doe"}

	invID := core.NewInvestmentID()
	amount := money.New(15000, "UAH")
	tags := []string{"rent", "monthly", " "}

	inv, err := investment.NewInvestment(
		invID,
		projID,
		"UAH",
		assetID,
		"Rent Jan",
		contributor,
		investment.ResourceTypeMoney,
		amount,
		nil,
		nil,
		nil,
		"",
		"",
		"January Rent Payment",
		now,
		tags,
	)
	if err != nil {
		t.Fatalf("unexpected error creating investment: %v", err)
	}

	if inv.ID() != invID {
		t.Errorf("expected ID %v, got %v", invID, inv.ID())
	}
	if inv.ProjectID() != projID {
		t.Errorf("expected project ID %v, got %v", projID, inv.ProjectID())
	}
	if inv.PrimaryAssetID() != assetID {
		t.Errorf("expected asset ID %v, got %v", assetID, inv.PrimaryAssetID())
	}
	if inv.PrimaryAssetName() != "Rent Jan" {
		t.Errorf("expected asset name 'Rent Jan', got '%s'", inv.PrimaryAssetName())
	}
	if len(inv.AssetAllocations()) != 1 {
		t.Fatalf("expected 1 default allocation, got %d", len(inv.AssetAllocations()))
	}
	if inv.AssetAllocations()[0].SharePercentage != 100.0 {
		t.Errorf("expected 100.0 share percentage, got %f", inv.AssetAllocations()[0].SharePercentage)
	}
	if inv.Amount().Amount() != 15000 {
		t.Errorf("expected amount 15000, got %d", inv.Amount().Amount())
	}
	if len(inv.Tags()) != 2 {
		t.Errorf("expected 2 clean tags, got %d: %+v", len(inv.Tags()), inv.Tags())
	}
}

func TestInvestment_SharedAllocations(t *testing.T) {
	projID := core.NewProjectID()
	ownerID := core.NewPersonID()
	asset1 := core.NewAssetID()
	asset2 := core.NewAssetID()
	now := time.Now()

	contributor := investment.ContributorRef{PersonID: ownerID, DisplayName: "John Doe"}

	allocations := []investment.InvestmentAssetLink{
		{AssetID: asset1, AssetName: "Asset 1", SharePercentage: 60.0},
		{AssetID: asset2, AssetName: "Asset 2", SharePercentage: 40.0},
	}

	inv, err := investment.NewInvestment(
		core.NewInvestmentID(),
		projID,
		"USD",
		core.NilAssetID,
		"",
		contributor,
		investment.ResourceTypeMoney,
		money.New(5000, "USD"),
		nil,
		nil,
		nil,
		"",
		"",
		"Shared Cloud Hosting",
		now,
		nil,
		allocations...,
	)
	if err != nil {
		t.Fatalf("unexpected error creating shared investment: %v", err)
	}

	if len(inv.AssetAllocations()) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(inv.AssetAllocations()))
	}
	if inv.PrimaryAssetID() != asset1 {
		t.Errorf("expected primary asset to be asset1, got %v", inv.PrimaryAssetID())
	}
}

func TestInvestment_AllocationInvariants(t *testing.T) {
	projID := core.NewProjectID()
	ownerID := core.NewPersonID()
	asset1 := core.NewAssetID()
	asset2 := core.NewAssetID()
	now := time.Now()

	contributor := investment.ContributorRef{PersonID: ownerID, DisplayName: "John Doe"}

	// 1. Sum != 100%
	badSum := []investment.InvestmentAssetLink{
		{AssetID: asset1, SharePercentage: 50.0},
		{AssetID: asset2, SharePercentage: 40.0},
	}
	_, err := investment.NewInvestment(
		core.NewInvestmentID(), projID, "USD", core.NilAssetID, "", contributor,
		investment.ResourceTypeMoney, money.New(100, "USD"),
		nil, nil, nil, "", "", "Invalid Sum", now, nil,
		badSum...,
	)
	if err == nil {
		t.Errorf("expected error when allocation sum is not 100%%")
	}

	// 2. Duplicate assets in allocations
	duplicates := []investment.InvestmentAssetLink{
		{AssetID: asset1, SharePercentage: 50.0},
		{AssetID: asset1, SharePercentage: 50.0},
	}
	_, err = investment.NewInvestment(
		core.NewInvestmentID(), projID, "USD", core.NilAssetID, "", contributor,
		investment.ResourceTypeMoney, money.New(100, "USD"),
		nil, nil, nil, "", "", "Duplicate Assets", now, nil,
		duplicates...,
	)
	if err == nil {
		t.Errorf("expected error when duplicate assets are allocated")
	}

	// 3. Share percentage out of bounds
	outOfBounds := []investment.InvestmentAssetLink{
		{AssetID: asset1, SharePercentage: 110.0},
		{AssetID: asset2, SharePercentage: -10.0},
	}
	_, err = investment.NewInvestment(
		core.NewInvestmentID(), projID, "USD", core.NilAssetID, "", contributor,
		investment.ResourceTypeMoney, money.New(100, "USD"),
		nil, nil, nil, "", "", "Out of Bounds", now, nil,
		outOfBounds...,
	)
	if err == nil {
		t.Errorf("expected error when share percentage is out of bounds")
	}
}

func TestInvestment_TimeResourceCalculation(t *testing.T) {
	projID := core.NewProjectID()
	ownerID := core.NewPersonID()
	assetID := core.NewAssetID()
	now := time.Now()

	contributor := investment.ContributorRef{PersonID: ownerID, DisplayName: "John Doe"}

	hours := 40.0
	rate := money.New(50, "USD")

	inv, err := investment.NewInvestment(
		core.NewInvestmentID(),
		projID,
		"USD",
		assetID,
		"Dev Sprint",
		contributor,
		investment.ResourceTypeTime,
		nil, // nil amount, should be calculated: 40 * 50 = 2000 USD
		&hours,
		rate,
		nil,
		"",
		"",
		"Sprint 1 Engineering",
		now,
		[]string{"labor"},
	)
	if err != nil {
		t.Fatalf("unexpected error creating time investment: %v", err)
	}

	if inv.Amount() == nil || inv.Amount().Amount() != 2000 {
		t.Errorf("expected calculated amount 2000, got %v", inv.Amount())
	}
}

func TestInvestment_ReallocateAndUpdate(t *testing.T) {
	projID := core.NewProjectID()
	ownerID := core.NewPersonID()
	asset1 := core.NewAssetID()
	asset2 := core.NewAssetID()
	now := time.Now()

	contributor := investment.ContributorRef{PersonID: ownerID, DisplayName: "John Doe"}

	inv, err := investment.NewInvestment(
		core.NewInvestmentID(),
		projID,
		"UAH",
		asset1,
		"Asset 1",
		contributor,
		investment.ResourceTypeMoney,
		money.New(1000, "UAH"),
		nil, nil, nil, "", "", "Original", now, nil,
	)
	if err != nil {
		t.Fatalf("failed to create investment: %v", err)
	}

	// Reallocate
	newAllocs := []investment.InvestmentAssetLink{
		{AssetID: asset1, AssetName: "Asset 1", SharePercentage: 30.0},
		{AssetID: asset2, AssetName: "Asset 2", SharePercentage: 70.0},
	}
	if err := inv.Allocate(newAllocs); err != nil {
		t.Fatalf("reallocation failed: %v", err)
	}
	if len(inv.AssetAllocations()) != 2 {
		t.Errorf("expected 2 allocations after reallocating")
	}

	// UpdateDetails
	inv.UpdateDetails("Updated Description", now.Add(24*time.Hour), []string{"new-tag"})
	if inv.Description() != "Updated Description" {
		t.Errorf("expected updated description")
	}
	if len(inv.Tags()) != 1 || inv.Tags()[0] != "new-tag" {
		t.Errorf("expected updated tag")
	}

	// UpdateValuation
	newAmt := money.New(2000, "UAH")
	if err := inv.UpdateValuation(investment.ResourceTypeMoney, newAmt, nil, nil, nil, "", ""); err != nil {
		t.Fatalf("valuation update failed: %v", err)
	}
	if inv.Amount().Amount() != 2000 {
		t.Errorf("expected amount 2000, got %d", inv.Amount().Amount())
	}
}

func TestInvestment_Reconstitute(t *testing.T) {
	invID := core.NewInvestmentID()
	projID := core.NewProjectID()
	assetID := core.NewAssetID()
	now := time.Now()

	contributor := investment.ContributorRef{PersonID: core.NewPersonID(), DisplayName: "Jane Doe"}

	inv := investment.ReconstituteInvestment(
		invID,
		projID,
		"USD",
		assetID,
		"Asset Name",
		nil,
		contributor,
		investment.ResourceTypeMoney,
		money.New(500, "USD"),
		nil, nil, nil, "", "",
		"Reconstituted Investment",
		now,
		[]string{"tag1"},
	)

	if inv.ID() != invID {
		t.Errorf("expected ID %v, got %v", invID, inv.ID())
	}
	if len(inv.AssetAllocations()) != 1 {
		t.Errorf("expected 1 default allocation reconstituted from primary asset")
	}

	t.Run("defensive copying on getters and mutators", func(t *testing.T) {
		// AssetAllocations getter defensive copy
		allocs := inv.AssetAllocations()
		allocs[0].SharePercentage = 50.0
		if inv.AssetAllocations()[0].SharePercentage == 50.0 {
			t.Errorf("AssetAllocations getter leaked internal slice pointer")
		}

		// Tags getter defensive copy
		tags := inv.Tags()
		tags[0] = "mutated-tag"
		if inv.Tags()[0] == "mutated-tag" {
			t.Errorf("Tags getter leaked internal slice pointer")
		}

		// Allocate defensive copy
		inputAllocs := []investment.InvestmentAssetLink{
			{AssetID: assetID, AssetName: "Asset Name", SharePercentage: 100.0},
		}
		_ = inv.Allocate(inputAllocs)
		inputAllocs[0].SharePercentage = 25.0
		if inv.AssetAllocations()[0].SharePercentage == 25.0 {
			t.Errorf("Allocate leaked input slice pointer")
		}
	})

	t.Run("UpdateValuation validation errors", func(t *testing.T) {
		negHours := -5.0
		negRate := money.New(-10, "USD")
		negQty := -2.0
		negAmt := money.New(-100, "USD")
		zeroAmt := money.New(0, "USD")

		if err := inv.UpdateValuation("INVALID_TYPE", nil, nil, nil, nil, "", ""); err == nil {
			t.Errorf("expected error for invalid resource type")
		}
		if err := inv.UpdateValuation(investment.ResourceTypeTime, nil, &negHours, nil, nil, "", ""); err == nil {
			t.Errorf("expected error for negative time hours")
		}
		if err := inv.UpdateValuation(investment.ResourceTypeTime, nil, nil, negRate, nil, "", ""); err == nil {
			t.Errorf("expected error for negative time hourly rate")
		}
		if err := inv.UpdateValuation(investment.ResourceTypeGoods, nil, nil, nil, &negQty, "", ""); err == nil {
			t.Errorf("expected error for negative goods quantity")
		}
		if err := inv.UpdateValuation(investment.ResourceTypeMoney, negAmt, nil, nil, nil, "", ""); err == nil {
			t.Errorf("expected error for negative investment amount")
		}
		if err := inv.UpdateValuation(investment.ResourceTypeMoney, zeroAmt, nil, nil, nil, "", ""); err == nil {
			t.Errorf("expected error for zero investment amount")
		}
	})
}

func TestInvestment_PointerImmutability(t *testing.T) {
	projID := core.NewProjectID()
	contributor := investment.ContributorRef{PersonID: core.NewPersonID(), DisplayName: "Dev"}
	assetID := core.NewAssetID()
	hours := 40.0
	qty := 10.0
	rate := money.New(50, "USD")

	inv, err := investment.NewInvestment(
		core.NewInvestmentID(),
		projID,
		"USD",
		assetID,
		"Asset",
		contributor,
		investment.ResourceTypeTime,
		money.New(2000, "USD"),
		&hours,
		rate,
		&qty,
		"pcs",
		"Widget",
		"Original",
		time.Now(),
		nil,
	)
	if err != nil {
		t.Fatalf("NewInvestment failed: %v", err)
	}

	// Mutate input pointer
	hours = 999.0
	if *inv.TimeHours() == 999.0 {
		t.Errorf("NewInvestment leaked input timeHours pointer")
	}

	// Mutate getter pointer
	th := inv.TimeHours()
	*th = 888.0
	if *inv.TimeHours() == 888.0 {
		t.Errorf("TimeHours getter leaked internal pointer")
	}

	// Mutate goods quantity getter pointer
	gq := inv.GoodsQuantity()
	*gq = 777.0
	if *inv.GoodsQuantity() == 777.0 {
		t.Errorf("GoodsQuantity getter leaked internal pointer")
	}
}
