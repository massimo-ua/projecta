package dal_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/project"
	"gitlab.com/massimo-ua/projecta/pkg/dal"
)

var (
	migrateOnce sync.Once
	migrateErr  error
)

func openIntegrationDB(t *testing.T) *dal.PgDbConnection {
	t.Helper()

	dbURI := os.Getenv("TEST_DB_URI")
	if dbURI == "" {
		t.Skip("TEST_DB_URI is not set; skipping PostgreSQL integration test")
	}

	migrateOnce.Do(func() {
		migrationsDir, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
		if err != nil {
			migrateErr = err
			return
		}
		m, err := migrate.New("file://"+migrationsDir, dbURI)
		if err != nil {
			migrateErr = err
			return
		}
		defer func() { _, _ = m.Close() }()
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			migrateErr = err
		}
	})
	if migrateErr != nil {
		t.Fatalf("failed to apply migrations: %v", migrateErr)
	}

	db, err := dal.NewPgDbConnection(dbURI)
	if err != nil {
		t.Fatalf("failed to connect to integration db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("failed to ping integration db: %v", err)
	}
	return db
}

func seedIntegrationPerson(t *testing.T, db *dal.PgDbConnection, first, last string) *people.Person {
	t.Helper()
	repo := dal.NewPgPeopleRepository(db)
	id := core.NewPersonID()
	cred, err := people.NewCredentials(people.LOCAL, fmt.Sprintf("%s@integration.test", id.String()), "hashed_secret_value")
	if err != nil {
		t.Fatalf("NewCredentials failed: %v", err)
	}
	p, err := people.NewPerson(id, first, last, first+" "+last, []people.Credentials{cred}, people.RoleUser)
	if err != nil {
		t.Fatalf("NewPerson failed: %v", err)
	}
	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("PgPeopleRepository.Save failed: %v", err)
	}
	return p
}

func TestPgProjectRepository_RoundTrip_Access_Rollback_And_Concurrency(t *testing.T) {
	t.Parallel()
	db := openIntegrationDB(t)
	ctx := context.Background()

	ownerPerson := seedIntegrationPerson(t, db, "Owner", "One")
	sharedPerson := seedIntegrationPerson(t, db, "Shared", "Two")
	outsiderPerson := seedIntegrationPerson(t, db, "Outsider", "Three")

	projRepo := dal.NewPgProjectRepository(db)
	owner := &project.Owner{
		PersonID:        ownerPerson.ID(),
		DisplayName:     ownerPerson.DisplayName(),
		CanHaveProjects: true,
	}

	projID := core.NewProjectID()
	now := time.Now().UTC().Truncate(time.Second)
	proj, err := project.NewProject(projID, "PG Project "+projID.String()[:8], "Initial desc", owner, now, now.AddDate(1, 0, 0), "USD")
	if err != nil {
		t.Fatalf("NewProject failed: %v", err)
	}
	if err := proj.SetTagBudgets([]project.TagBudgetInput{
		{Tag: "home", Frequency: "MONTHLY", TargetMonth: 1, ResourceType: "MONEY", Amount: 50000, Currency: "USD", Description: "monthly"},
	}); err != nil {
		t.Fatalf("SetTagBudgets failed: %v", err)
	}
	if err := projRepo.Create(ctx, proj); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Owner can read
	loaded, err := projRepo.GetProject(ctx, projID, ownerPerson.ID())
	if err != nil {
		t.Fatalf("GetProject by owner failed: %v", err)
	}
	if loaded.MainCurrency() != "USD" || len(loaded.TagBudgets()) != 1 {
		t.Fatalf("unexpected loaded project: currency=%s budgets=%d", loaded.MainCurrency(), len(loaded.TagBudgets()))
	}

	// Outsider is blocked
	if _, err := projRepo.GetProject(ctx, projID, outsiderPerson.ID()); err == nil {
		t.Fatalf("expected error when outsider reads private project")
	}

	// Grant share and verify shared user can read
	if err := projRepo.CreateShareRecord(ctx, projID, sharedPerson.ID()); err != nil {
		t.Fatalf("CreateShareRecord failed: %v", err)
	}
	sharedLoaded, err := projRepo.GetProject(ctx, projID, sharedPerson.ID())
	if err != nil {
		t.Fatalf("GetProject by shared user failed: %v", err)
	}
	if !sharedLoaded.IsShared() || !sharedLoaded.HasSharedAccess(sharedPerson.ID()) {
		t.Errorf("expected IsShared=true and HasSharedAccess=true for shared user")
	}

	// Rollback verification: returning an error from UpdateProject closure must leave DB state unchanged
	rollbackErr := errors.New("abort project update")
	err = projRepo.UpdateProject(ctx, projID, ownerPerson.ID(), func(_ context.Context, p *project.Project) (*project.Project, error) {
		_ = p.UpdateDetails("SHOULD_ROLLBACK", "SHOULD_ROLLBACK", "EUR")
		return nil, rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected rollbackErr, got %v", err)
	}
	afterRollback, err := projRepo.GetProject(ctx, projID, ownerPerson.ID())
	if err != nil {
		t.Fatalf("GetProject after rollback failed: %v", err)
	}
	if afterRollback.MainCurrency() != "USD" || afterRollback.Description() != "Initial desc" {
		t.Errorf("expected DB state preserved after rollback, got currency=%s desc=%q", afterRollback.MainCurrency(), afterRollback.Description())
	}

	// Concurrency verification: 20 parallel goroutines updating under SELECT ... FOR UPDATE
	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	startWorkers := make(chan struct{})
	var errCount atomic.Int32
	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			<-startWorkers
			err := projRepo.UpdateProject(ctx, projID, ownerPerson.ID(), func(_ context.Context, p *project.Project) (*project.Project, error) {
				nextDesc := p.Description() + "."
				if uErr := p.UpdateDetails(p.Name(), nextDesc, p.MainCurrency()); uErr != nil {
					return nil, uErr
				}
				return p, nil
			})
			if err != nil {
				errCount.Add(1)
			}
		}(i)
	}
	close(startWorkers)
	wg.Wait()

	if errCount.Load() != 0 {
		t.Fatalf("expected 0 errors in concurrent UpdateProject, got %d", errCount.Load())
	}
	finalProj, err := projRepo.GetProject(ctx, projID, ownerPerson.ID())
	if err != nil {
		t.Fatalf("GetProject final failed: %v", err)
	}
	expectedLen := len("Initial desc") + workers
	if len(finalProj.Description()) != expectedLen {
		t.Errorf("lost update detected under FOR UPDATE: expected description len %d, got %d (%q)", expectedLen, len(finalProj.Description()), finalProj.Description())
	}
}

func TestPgAssetRepository_RoundTrip_DAG_Rollback_And_Concurrency(t *testing.T) {
	t.Parallel()
	db := openIntegrationDB(t)
	ctx := context.Background()

	ownerPerson := seedIntegrationPerson(t, db, "Asset", "Owner")
	projRepo := dal.NewPgProjectRepository(db)
	assetRepo := dal.NewPgAssetRepository(db)

	projID := core.NewProjectID()
	now := time.Now().UTC().Truncate(time.Second)
	proj, err := project.NewProject(projID, "Asset Proj "+projID.String()[:8], "", &project.Owner{PersonID: ownerPerson.ID(), DisplayName: ownerPerson.DisplayName(), CanHaveProjects: true}, now, now.AddDate(1, 0, 0), "UAH")
	if err != nil {
		t.Fatalf("NewProject failed: %v", err)
	}
	if err := projRepo.Create(ctx, proj); err != nil {
		t.Fatalf("Create project failed: %v", err)
	}

	ownerRef := asset.OwnerRef{PersonID: ownerPerson.ID(), DisplayName: ownerPerson.DisplayName()}
	parentID := core.NewAssetID()
	childID := core.NewAssetID()

	parent, err := asset.CreateAsset(parentID, "Solar System", "Parent", projID, money.New(100000, "UAH"), now, ownerRef, asset.AssetStatusActive, nil, []string{"energy"})
	if err != nil {
		t.Fatalf("CreateAsset parent failed: %v", err)
	}
	child, err := asset.CreateAsset(childID, "Battery", "Child", projID, money.New(40000, "UAH"), now, ownerRef, asset.AssetStatusActive, nil, []string{"energy"})
	if err != nil {
		t.Fatalf("CreateAsset child failed: %v", err)
	}
	if err := assetRepo.Save(ctx, parent); err != nil {
		t.Fatalf("Save parent failed: %v", err)
	}
	if err := assetRepo.Save(ctx, child); err != nil {
		t.Fatalf("Save child failed: %v", err)
	}
	if err := assetRepo.AddChild(ctx, parentID, childID, 100.0); err != nil {
		t.Fatalf("AddChild failed: %v", err)
	}

	loadedParent, err := assetRepo.FindOne(ctx, asset.Filter{ID: parentID, ProjectID: projID, OwnerID: ownerPerson.ID()})
	if err != nil {
		t.Fatalf("FindOne parent failed: %v", err)
	}
	if len(loadedParent.Children()) != 1 || loadedParent.Children()[0].ChildID != childID {
		t.Fatalf("expected 1 child link, got %+v", loadedParent.Children())
	}

	// Rollback verification on UpdateAsset
	abortErr := errors.New("abort asset update")
	err = assetRepo.UpdateAsset(ctx, parentID, projID, ownerPerson.ID(), func(_ context.Context, a *asset.Asset) (*asset.Asset, error) {
		_ = a.UpdateDetails("SHOULD_NOT_PERSIST", a.Description(), a.TargetPrice(), a.Tags())
		return nil, abortErr
	})
	if !errors.Is(err, abortErr) {
		t.Fatalf("expected abortErr, got %v", err)
	}
	afterRollback, err := assetRepo.FindOne(ctx, asset.Filter{ID: parentID, ProjectID: projID, OwnerID: ownerPerson.ID()})
	if err != nil || afterRollback.Name() != "Solar System" {
		t.Fatalf("expected original name after rollback, got %v (err=%v)", afterRollback, err)
	}

	// Concurrency verification: 20 goroutines race to mark ACTIVE asset COMPLETED (only 1 winner allowed)
	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	startWorkers := make(chan struct{})
	var winners atomic.Int32
	var alreadyDone atomic.Int32
	var unexpected atomic.Int32
	doneAt := now.Add(time.Hour)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-startWorkers
			err := assetRepo.UpdateAsset(ctx, childID, projID, ownerPerson.ID(), func(_ context.Context, a *asset.Asset) (*asset.Asset, error) {
				if a.Status() == asset.AssetStatusCompleted {
					return nil, exceptions.NewValidationException("already completed", nil)
				}
				a.Complete(doneAt)
				return a, nil
			})
			if err == nil {
				winners.Add(1)
			} else {
				var exc exceptions.Exception
				if errors.As(err, &exc) && exc.Code == exceptions.ValidationFailed {
					alreadyDone.Add(1)
				} else {
					unexpected.Add(1)
				}
			}
		}()
	}
	close(startWorkers)
	wg.Wait()

	if winners.Load() != 1 || alreadyDone.Load() != workers-1 || unexpected.Load() != 0 {
		t.Errorf("FOR UPDATE race check failed: winners=%d alreadyDone=%d unexpected=%d", winners.Load(), alreadyDone.Load(), unexpected.Load())
	}
}

func TestPgInvestmentRepository_RoundTrip_Rollback_And_Concurrency(t *testing.T) {
	t.Parallel()
	db := openIntegrationDB(t)
	ctx := context.Background()

	ownerPerson := seedIntegrationPerson(t, db, "Inv", "Owner")
	projRepo := dal.NewPgProjectRepository(db)
	assetRepo := dal.NewPgAssetRepository(db)
	invRepo := dal.NewPgInvestmentRepository(db)

	projID := core.NewProjectID()
	now := time.Now().UTC().Truncate(time.Second)
	proj, err := project.NewProject(projID, "Inv Proj "+projID.String()[:8], "", &project.Owner{PersonID: ownerPerson.ID(), DisplayName: ownerPerson.DisplayName(), CanHaveProjects: true}, now, now.AddDate(1, 0, 0), "UAH")
	if err != nil {
		t.Fatalf("NewProject failed: %v", err)
	}
	if err := projRepo.Create(ctx, proj); err != nil {
		t.Fatalf("Create project failed: %v", err)
	}

	ownerRef := asset.OwnerRef{PersonID: ownerPerson.ID(), DisplayName: ownerPerson.DisplayName()}
	astID := core.NewAssetID()
	ast, err := asset.CreateAsset(astID, "Equipment", "", projID, money.New(50000, "UAH"), now, ownerRef, asset.AssetStatusActive, nil, nil)
	if err != nil {
		t.Fatalf("CreateAsset failed: %v", err)
	}
	if err := assetRepo.Save(ctx, ast); err != nil {
		t.Fatalf("Save asset failed: %v", err)
	}

	invID := core.NewInvestmentID()
	contrib := investment.ContributorRef{PersonID: ownerPerson.ID(), DisplayName: ownerPerson.DisplayName()}
	inv, err := investment.NewInvestment(
		invID, projID, "UAH", astID, "Equipment", contrib,
		investment.ResourceTypeMoney, money.New(1000, "UAH"),
		nil, nil, nil, "", "", "Initial payment", now, []string{"capex"},
	)
	if err != nil {
		t.Fatalf("NewInvestment failed: %v", err)
	}
	if err := invRepo.Save(ctx, inv); err != nil {
		t.Fatalf("Save investment failed: %v", err)
	}

	// Round-trip + tags read model
	loaded, err := invRepo.FindOne(ctx, investment.Filter{ID: invID, ProjectID: projID})
	if err != nil || loaded.Amount().Amount() != 1000 {
		t.Fatalf("FindOne investment failed: loaded=%v err=%v", loaded, err)
	}
	tags, err := invRepo.FindTags(ctx, projID)
	if err != nil || len(tags) != 1 || tags[0] != "capex" {
		t.Fatalf("FindTags unexpected: tags=%v err=%v", tags, err)
	}

	// Rollback verification
	abortErr := errors.New("abort investment update")
	err = invRepo.UpdateInvestment(ctx, invID, projID, func(_ context.Context, cur *investment.Investment) (*investment.Investment, error) {
		_ = cur.UpdateValuation(cur.ResourceType(), money.New(999999, "UAH"), nil, nil, nil, "", "")
		return nil, abortErr
	})
	if !errors.Is(err, abortErr) {
		t.Fatalf("expected abortErr, got %v", err)
	}
	afterRollback, err := invRepo.FindOne(ctx, investment.Filter{ID: invID, ProjectID: projID})
	if err != nil || afterRollback.Amount().Amount() != 1000 {
		t.Fatalf("expected amount 1000 after rollback, got %v (err=%v)", afterRollback, err)
	}

	// Concurrency verification: 20 goroutines each add +100 under SELECT ... FOR UPDATE
	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	startWorkers := make(chan struct{})
	var errCount atomic.Int32
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-startWorkers
			err := invRepo.UpdateInvestment(ctx, invID, projID, func(_ context.Context, cur *investment.Investment) (*investment.Investment, error) {
				nextAmount := money.New(cur.Amount().Amount()+100, cur.Amount().Currency().Code)
				if uErr := cur.UpdateValuation(cur.ResourceType(), nextAmount, nil, nil, nil, "", ""); uErr != nil {
					return nil, uErr
				}
				return cur, nil
			})
			if err != nil {
				errCount.Add(1)
			}
		}()
	}
	close(startWorkers)
	wg.Wait()

	if errCount.Load() != 0 {
		t.Fatalf("expected 0 errors in concurrent UpdateInvestment, got %d", errCount.Load())
	}
	finalInv, err := invRepo.FindOne(ctx, investment.Filter{ID: invID, ProjectID: projID})
	if err != nil {
		t.Fatalf("FindOne final failed: %v", err)
	}
	expectedAmount := int64(1000 + workers*100)
	if finalInv.Amount().Amount() != expectedAmount {
		t.Errorf("lost update detected in UpdateInvestment: expected %d, got %d", expectedAmount, finalInv.Amount().Amount())
	}
}
