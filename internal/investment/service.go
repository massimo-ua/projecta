package investment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

const (
	failedToCreateInvestment = "failed to create investment"
	failedToFindInvestment   = "failed to find investment"
	failedToUpdateInvestment = "failed to update investment"
)

// AssetReader is a consumer-owned interface providing read access to assets.
type AssetReader interface {
	FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error)
}

// ProjectReader is a consumer-owned interface providing read access to projects.
type ProjectReader interface {
	FindOne(ctx context.Context, filter project.ProjectFilter) (*project.Project, error)
}

type ServiceImpl struct {
	db          core.DbConnection
	investments Repository
	assets      AssetReader
	projects    ProjectReader
	people      project.PeopleService
}

func NewService(
	db core.DbConnection,
	investments Repository,
	assets AssetReader,
	projects ProjectReader,
	people project.PeopleService,
) *ServiceImpl {
	return &ServiceImpl{
		db:          db,
		investments: investments,
		assets:      assets,
		projects:    projects,
		people:      people,
	}
}

func (s *ServiceImpl) Create(ctx context.Context, command CreateInvestmentCommand) (*Investment, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToCreateInvestment, err)
	}

	if command.ProjectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	if command.AssetID == uuid.Nil && len(command.AssetAllocations) == 0 {
		return nil, exceptions.NewValidationException("asset id is required", nil)
	}

	if strings.TrimSpace(command.Description) == "" {
		return nil, exceptions.NewValidationException("description is required", nil)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	project, err := s.projects.FindOne(ctx, project.ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	allocations, primaryID, primaryName, err := s.resolveAllocations(ctx, command.ProjectID, command.AssetID, command.AssetAllocations)
	if err != nil {
		return nil, err
	}

	resType := command.ResourceType
	if resType == "" {
		resType = ResourceTypeMoney
	}

	inv, err := NewInvestment(
		uuid.New(),
		project.ID(),
		project.MainCurrency(),
		primaryID,
		primaryName,
		ContributorRef{PersonID: owner.PersonID, DisplayName: owner.DisplayName},
		resType,
		command.Amount,
		command.TimeHours,
		command.TimeHourlyRate,
		command.GoodsQuantity,
		command.GoodsUnit,
		command.GoodsItemName,
		command.Description,
		command.Date,
		command.Tags,
		allocations...,
	)
	if err != nil {
		return nil, err
	}

	if err := s.investments.Save(ctx, inv); err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	return inv, nil
}

func (s *ServiceImpl) CreateBatch(ctx context.Context, command CreateBatchInvestmentsCommand) ([]*Investment, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToCreateInvestment, err)
	}

	if command.ProjectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	if len(command.Items) == 0 {
		return nil, exceptions.NewValidationException("at least one investment item is required", nil)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	project, err := s.projects.FindOne(ctx, project.ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	var created []*Investment
	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		for i, item := range command.Items {
			targetAssetID := item.AssetID
			if targetAssetID == uuid.Nil {
				targetAssetID = command.DefaultAsset
			}
			if targetAssetID == uuid.Nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("item %d requires an asset", i), nil)
			}

			ast, err := s.assets.FindOne(txCtx, asset.Filter{ID: targetAssetID, ProjectID: command.ProjectID})
			if err != nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("asset %s not found in project", targetAssetID), err)
			}

			resType := item.ResourceType
			if resType == "" {
				resType = ResourceTypeMoney
			}

			inv, err := NewInvestment(
				uuid.New(),
				project.ID(),
				project.MainCurrency(),
				ast.ID(),
				ast.Name(),
				ContributorRef{PersonID: owner.PersonID, DisplayName: owner.DisplayName},
				resType,
				item.Amount,
				item.TimeHours,
				item.TimeHourlyRate,
				item.GoodsQuantity,
				item.GoodsUnit,
				item.GoodsItemName,
				item.Description,
				core.DateOrNow(item.Date),
				item.Tags,
			)
			if err != nil {
				return nil, err
			}

			if err := s.investments.Save(txCtx, inv); err != nil {
				return nil, err
			}
			created = append(created, inv)
		}
		return nil, nil
	})

	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	return created, nil
}

func (s *ServiceImpl) Update(ctx context.Context, command UpdateInvestmentCommand) error {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException(failedToUpdateInvestment, err)
	}

	inv, err := s.investments.FindOne(ctx, Filter{
		ID:        command.ID,
		ProjectID: command.ProjectID,
	})
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException(failedToFindInvestment, err)
		}
		return exceptions.NewInternalException(failedToFindInvestment, err)
	}

	if len(command.AssetAllocations) > 0 || (command.AssetID != uuid.Nil && command.AssetID != inv.PrimaryAssetID()) {
		allocations, _, _, err := s.resolveAllocations(ctx, command.ProjectID, command.AssetID, command.AssetAllocations)
		if err != nil {
			return err
		}
		if err := inv.Allocate(allocations); err != nil {
			return err
		}
	}

	inv.UpdateDetails(command.Description, command.Date, command.Tags)
	if err := inv.UpdateValuation(command.ResourceType, command.Amount, command.TimeHours, command.TimeHourlyRate, command.GoodsQuantity, command.GoodsUnit, command.GoodsItemName); err != nil {
		return err
	}

	return s.investments.Save(ctx, inv)
}

func (s *ServiceImpl) Remove(ctx context.Context, command RemoveInvestmentCommand) error {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException(failedToFindInvestment, err)
	}

	inv, err := s.investments.FindOne(ctx, Filter{
		ID:        command.ID,
		ProjectID: command.ProjectID,
	})
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException(failedToFindInvestment, err)
		}
		return exceptions.NewInternalException(failedToFindInvestment, err)
	}

	return s.investments.Remove(ctx, inv)
}

func (s *ServiceImpl) FindOne(ctx context.Context, filter Filter) (*Investment, error) {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToFindInvestment, err)
	}

	return s.investments.FindOne(ctx, filter)
}

func (s *ServiceImpl) Find(ctx context.Context, filter CollectionFilter) (*Collection, error) {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToFindInvestment, err)
	}

	return s.investments.Find(ctx, filter)
}

func (s *ServiceImpl) FindTags(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to find tags", err)
	}

	return s.investments.FindTags(ctx, projectID)
}

func (s *ServiceImpl) resolveAllocations(
	ctx context.Context,
	projectID uuid.UUID,
	primaryID uuid.UUID,
	inputs []InvestmentAssetInput,
) ([]InvestmentAssetLink, uuid.UUID, string, error) {
	if len(inputs) > 1 {
		allocs := make([]InvestmentAssetLink, 0, len(inputs))
		var primaryName string
		for _, aInput := range inputs {
			if aInput.AssetID == uuid.Nil {
				return nil, uuid.Nil, "", exceptions.NewValidationException("asset id is required for each allocation", nil)
			}
			ast, err := s.assets.FindOne(ctx, asset.Filter{ID: aInput.AssetID, ProjectID: projectID})
			if err != nil {
				return nil, uuid.Nil, "", exceptions.NewValidationException(fmt.Sprintf("asset %s not found in project", aInput.AssetID), err)
			}
			if primaryName == "" {
				primaryName = ast.Name()
			}
			allocs = append(allocs, InvestmentAssetLink{
				AssetID:         ast.ID(),
				AssetName:       ast.Name(),
				SharePercentage: aInput.SharePercentage,
			})
		}
		return allocs, inputs[0].AssetID, primaryName, nil
	}

	targetID := primaryID
	if len(inputs) == 1 {
		targetID = inputs[0].AssetID
	}
	if targetID == uuid.Nil {
		return nil, uuid.Nil, "", exceptions.NewValidationException("asset id or asset allocations are required", nil)
	}

	ast, err := s.assets.FindOne(ctx, asset.Filter{ID: targetID, ProjectID: projectID})
	if err != nil {
		return nil, uuid.Nil, "", exceptions.NewValidationException("asset not found in project", err)
	}
	allocs := []InvestmentAssetLink{
		{
			AssetID:         ast.ID(),
			AssetName:       ast.Name(),
			SharePercentage: 100.0,
		},
	}
	return allocs, ast.ID(), ast.Name(), nil
}
