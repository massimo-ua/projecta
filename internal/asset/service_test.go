package asset_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type mockDb struct{}

func (m *mockDb) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	return fn(ctx)
}
func (m *mockDb) Close()                        {}
func (m *mockDb) Ping(ctx context.Context) error { return nil }

type mockAssetRepo struct {
	saveErr    error
	removeErr  error
	findErr    error
	findOneErr error
	asset      *asset.Asset
	col        *asset.Collection
}

func (m *mockAssetRepo) Save(ctx context.Context, a *asset.Asset) error { return m.saveErr }
func (m *mockAssetRepo) Remove(ctx context.Context, a *asset.Asset) error {
	return m.removeErr
}
func (m *mockAssetRepo) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.col != nil {
		return m.col, nil
	}
	return asset.NewCollection(0), nil
}
func (m *mockAssetRepo) FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error) {
	if m.findOneErr != nil {
		return nil, m.findOneErr
	}
	return m.asset, nil
}

type mockPeopleService struct {
	owner *projecta.Owner
	err   error
}

func (m *mockPeopleService) FindOwner(ctx context.Context, id uuid.UUID) (*projecta.Owner, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.owner, nil
}

type mockTypeRepo struct {
	costType *projecta.CostType
	err      error
}

func (m *mockTypeRepo) Save(ctx context.Context, t *projecta.CostType) error   { return nil }
func (m *mockTypeRepo) Remove(ctx context.Context, t *projecta.CostType) error { return nil }
func (m *mockTypeRepo) Find(ctx context.Context, filter projecta.TypeCollectionFilter) (*projecta.CostTypeCollection, error) {
	return nil, nil
}
func (m *mockTypeRepo) FindOne(ctx context.Context, filter projecta.TypeFilter) (*projecta.CostType, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.costType, nil
}

type mockProjectRepo struct {
	project *projecta.Project
	err     error
}

func (m *mockProjectRepo) Create(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Update(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Remove(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Find(ctx context.Context, filter projecta.ProjectCollectionFilter) ([]*projecta.Project, error) {
	return nil, nil
}
func (m *mockProjectRepo) FindOne(ctx context.Context, filter projecta.ProjectFilter) (*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
}

type mockPaymentRepo struct {
	saveErr error
}

func (m *mockPaymentRepo) Save(ctx context.Context, p *projecta.Payment) error { return m.saveErr }
func (m *mockPaymentRepo) Remove(ctx context.Context, p *projecta.Payment) error {
	return nil
}
func (m *mockPaymentRepo) Find(ctx context.Context, filter projecta.PaymentCollectionFilter) (*projecta.PaymentCollection, error) {
	return nil, nil
}
func (m *mockPaymentRepo) FindOne(ctx context.Context, filter projecta.PaymentFilter) (*projecta.Payment, error) {
	return nil, nil
}

func TestAssetService(t *testing.T) {
	requesterID := uuid.New()
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, requesterID)

	now := time.Now()
	owner := &projecta.Owner{PersonID: requesterID, DisplayName: "John Doe"}
	project, _ := projecta.NewProject(uuid.New(), "Project 1", "Desc", owner, now, now)
	costType, _ := projecta.NewCostType(project.ProjectID, nil, "Type 1", "Desc")
	existingAsset := asset.NewAsset(uuid.New(), "Laptop", "Work Laptop", project, costType, money.New(1000, money.USD), now, owner)

	t.Run("Find unauthorized and success", func(t *testing.T) {
		svc := asset.NewService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})

		_, err := svc.Find(context.Background(), asset.CollectionFilter{})
		if err == nil {
			t.Errorf("expected unauthorized error")
		}

		res, err := svc.Find(authedCtx, asset.CollectionFilter{})
		if err != nil || res == nil {
			t.Errorf("expected find success, got err: %v", err)
		}

		svcErr := asset.NewService(&mockDb{}, &mockAssetRepo{findErr: errors.New("db error")}, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})
		_, err = svcErr.Find(authedCtx, asset.CollectionFilter{})
		if err == nil {
			t.Errorf("expected find error")
		}
	})

	t.Run("FindOne unauthorized and success", func(t *testing.T) {
		repo := &mockAssetRepo{asset: existingAsset}
		svc := asset.NewService(&mockDb{}, repo, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})

		_, err := svc.FindOne(context.Background(), asset.Filter{})
		if err == nil {
			t.Errorf("expected unauthorized error")
		}

		res, err := svc.FindOne(authedCtx, asset.Filter{})
		if err != nil || res != existingAsset {
			t.Errorf("expected findone success")
		}

		repoErr := &mockAssetRepo{findOneErr: errors.New("not found")}
		svcErr := asset.NewService(&mockDb{}, repoErr, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})
		_, err = svcErr.FindOne(authedCtx, asset.Filter{})
		if err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("Create success without and with payment", func(t *testing.T) {
		assetRepo := &mockAssetRepo{}
		peopleSvc := &mockPeopleService{owner: owner}
		typeRepo := &mockTypeRepo{costType: costType}
		projRepo := &mockProjectRepo{project: project}
		payRepo := &mockPaymentRepo{}

		svc := asset.NewService(&mockDb{}, assetRepo, peopleSvc, typeRepo, projRepo, payRepo)

		cmd := asset.CreateAssetCommand{
			Name:        "Server",
			Description: "Rack Server",
			ProjectID:   project.ProjectID,
			TypeID:      costType.ID,
			Price:       money.New(5000, money.USD),
			WithPayment: false,
		}
		a, err := svc.Create(authedCtx, cmd)
		if err != nil || a == nil {
			t.Fatalf("unexpected error creating asset: %v", err)
		}

		cmdWithPayment := cmd
		cmdWithPayment.WithPayment = true
		cmdWithPayment.Description = ""
		aPay, err := svc.Create(authedCtx, cmdWithPayment)
		if err != nil || aPay == nil {
			t.Fatalf("unexpected error creating asset with payment: %v", err)
		}
	})

	t.Run("Create error branches", func(t *testing.T) {
		peopleSvc := &mockPeopleService{owner: owner}
		typeRepo := &mockTypeRepo{costType: costType}
		projRepo := &mockProjectRepo{project: project}
		cmd := asset.CreateAssetCommand{}

		svc := asset.NewService(&mockDb{}, &mockAssetRepo{}, peopleSvc, typeRepo, projRepo, &mockPaymentRepo{})
		_, err := svc.Create(context.Background(), cmd)
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{err: errors.New("err")}, typeRepo, projRepo, &mockPaymentRepo{})
		_, err = svc.Create(authedCtx, cmd)
		if err == nil {
			t.Errorf("expected people service error")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{}, peopleSvc, typeRepo, &mockProjectRepo{err: errors.New("err")}, &mockPaymentRepo{})
		_, err = svc.Create(authedCtx, cmd)
		if err == nil {
			t.Errorf("expected project repo error")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{}, peopleSvc, &mockTypeRepo{err: errors.New("err")}, projRepo, &mockPaymentRepo{})
		_, err = svc.Create(authedCtx, cmd)
		if err == nil {
			t.Errorf("expected type repo error")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{}, peopleSvc, typeRepo, projRepo, &mockPaymentRepo{saveErr: errors.New("pay save err")})
		_, err = svc.Create(authedCtx, asset.CreateAssetCommand{WithPayment: true, ProjectID: project.ProjectID, TypeID: costType.ID})
		if err == nil {
			t.Errorf("expected payment save error")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{saveErr: errors.New("asset save err")}, peopleSvc, typeRepo, projRepo, &mockPaymentRepo{})
		_, err = svc.Create(authedCtx, asset.CreateAssetCommand{WithPayment: true, ProjectID: project.ProjectID, TypeID: costType.ID})
		if err == nil {
			t.Errorf("expected asset save error")
		}
	})

	t.Run("Remove success and errors", func(t *testing.T) {
		assetRepo := &mockAssetRepo{asset: existingAsset}
		svc := asset.NewService(&mockDb{}, assetRepo, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})

		err := svc.Remove(context.Background(), asset.RemoveAssetCommand{})
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svcErr := asset.NewService(&mockDb{}, &mockAssetRepo{findOneErr: errors.New("not found")}, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})
		err = svcErr.Remove(authedCtx, asset.RemoveAssetCommand{})
		if err == nil {
			t.Errorf("expected findone error")
		}

		err = svc.Remove(authedCtx, asset.RemoveAssetCommand{AssetID: existingAsset.ID()})
		if err != nil {
			t.Errorf("unexpected remove error: %v", err)
		}
	})

	t.Run("Update success and errors", func(t *testing.T) {
		assetRepo := &mockAssetRepo{asset: existingAsset}
		typeRepo := &mockTypeRepo{costType: costType}
		projRepo := &mockProjectRepo{project: project}

		svc := asset.NewService(&mockDb{}, assetRepo, &mockPeopleService{}, typeRepo, projRepo, &mockPaymentRepo{})

		updCmd := asset.UpdateAssetCommand{
			AssetID:     existingAsset.ID(),
			ProjectID:   project.ProjectID,
			TypeID:      costType.ID,
			Name:        "Updated Name",
			Description: "Updated Desc",
			Price:       money.New(3000, money.USD),
			AcquiredAt:  time.Now(),
		}

		err := svc.Update(context.Background(), updCmd)
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svcProjErr := asset.NewService(&mockDb{}, assetRepo, &mockPeopleService{}, typeRepo, &mockProjectRepo{err: errors.New("proj err")}, &mockPaymentRepo{})
		err = svcProjErr.Update(authedCtx, updCmd)
		if err == nil {
			t.Errorf("expected project error")
		}

		svcAssetErr := asset.NewService(&mockDb{}, &mockAssetRepo{findOneErr: errors.New("asset err")}, &mockPeopleService{}, typeRepo, projRepo, &mockPaymentRepo{})
		err = svcAssetErr.Update(authedCtx, updCmd)
		if err == nil {
			t.Errorf("expected asset err")
		}

		svcTypeErr := asset.NewService(&mockDb{}, assetRepo, &mockPeopleService{}, &mockTypeRepo{err: errors.New("type err")}, projRepo, &mockPaymentRepo{})
		err = svcTypeErr.Update(authedCtx, updCmd)
		if err == nil {
			t.Errorf("expected type err")
		}

		err = svc.Update(authedCtx, updCmd)
		if err != nil {
			t.Errorf("unexpected update error: %v", err)
		}
	})
}
