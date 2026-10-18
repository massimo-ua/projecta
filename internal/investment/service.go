package investment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
	failedToCreateInvestment = "failed to create investment"
	failedToFindInvestment   = "failed to find investment"
	failedToUpdateInvestment = "failed to update investment"
)

type ServiceImpl struct {
	db           core.DbConnection
	investments  Repository
	assets       AssetReader
	projects     ProjectReader
	people       ContributorReader
	rateProvider CurrencyConverter
}

func NewService(
	db core.DbConnection,
	investments Repository,
	assets AssetReader,
	projects ProjectReader,
	people ContributorReader,
	rateProvider ...CurrencyConverter,
) *ServiceImpl {
	var rp CurrencyConverter
	if len(rateProvider) > 0 {
		rp = rateProvider[0]
	}
	return &ServiceImpl{
		db:           db,
		investments:  investments,
		assets:       assets,
		projects:     projects,
		people:       people,
		rateProvider: rp,
	}
}

func (s *ServiceImpl) verifyProjectAccess(ctx context.Context, projectID uuid.UUID) error {
	if projectID == uuid.Nil {
		return exceptions.NewValidationException("project id is required", nil)
	}
	_, err := s.projects.FindProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException("project not found", err)
		}
		return exceptions.NewInternalException("failed to verify project access", err)
	}
	return nil
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

	contributor, err := s.people.FindContributor(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	proj, err := s.projects.FindProject(ctx, command.ProjectID)
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}
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
		proj.ID,
		proj.MainCurrency,
		primaryID,
		primaryName,
		contributor,
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

	inv.EvaluateHomeAmount(s.rateProvider)

	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := s.investments.Save(txCtx, inv); err != nil {
			return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
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

	contributor, err := s.people.FindContributor(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	proj, err := s.projects.FindProject(ctx, command.ProjectID)
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}
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

			ast, err := s.assets.FindAsset(txCtx, command.ProjectID, targetAssetID)
			if err != nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("asset %s not found in project", targetAssetID), err)
			}

			resType := item.ResourceType
			if resType == "" {
				resType = ResourceTypeMoney
			}

			inv, err := NewInvestment(
				uuid.New(),
				proj.ID,
				proj.MainCurrency,
				ast.ID,
				ast.Name,
				contributor,
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
			inv.EvaluateHomeAmount(s.rateProvider)

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

	if command.ProjectID != uuid.Nil {
		if err := s.verifyProjectAccess(ctx, command.ProjectID); err != nil {
			return err
		}
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

	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := s.investments.Save(txCtx, inv); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (s *ServiceImpl) AssignToAsset(ctx context.Context, assetID uuid.UUID, investmentIDs []uuid.UUID) error {
	if len(investmentIDs) == 0 {
		return nil
	}

	ast, err := s.assets.FindAsset(ctx, uuid.Nil, assetID)
	if err != nil {
		return err
	}

	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		for _, invID := range investmentIDs {
			inv, err := s.investments.FindOne(txCtx, Filter{ID: invID})
			if err != nil {
				return nil, err
			}
			if err := inv.Allocate([]InvestmentAssetLink{
				{
					AssetID:         ast.ID,
					AssetName:       ast.Name,
					SharePercentage: 100.0,
				},
			}); err != nil {
				return nil, err
			}
			if err := s.investments.Save(txCtx, inv); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (s *ServiceImpl) Remove(ctx context.Context, command RemoveInvestmentCommand) error {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException(failedToFindInvestment, err)
	}

	if command.ProjectID != uuid.Nil {
		if err := s.verifyProjectAccess(ctx, command.ProjectID); err != nil {
			return err
		}
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

	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := s.investments.Remove(txCtx, inv); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (s *ServiceImpl) FindOne(ctx context.Context, filter Filter) (*Investment, error) {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToFindInvestment, err)
	}

	if filter.ProjectID != uuid.Nil {
		if err := s.verifyProjectAccess(ctx, filter.ProjectID); err != nil {
			return nil, err
		}
	}

	inv, err := s.investments.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if inv != nil {
		inv.EvaluateHomeAmount(s.rateProvider)
	}
	return inv, nil
}

func (s *ServiceImpl) Find(ctx context.Context, filter CollectionFilter) (*Collection, error) {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToFindInvestment, err)
	}

	if filter.ProjectID != uuid.Nil {
		if err := s.verifyProjectAccess(ctx, filter.ProjectID); err != nil {
			return nil, err
		}
	}

	col, err := s.investments.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if col != nil {
		for _, inv := range col.Elements() {
			if inv != nil {
				inv.EvaluateHomeAmount(s.rateProvider)
			}
		}
	}
	return col, nil
}

func (s *ServiceImpl) FindTags(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	_, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to find tags", err)
	}

	if err := s.verifyProjectAccess(ctx, projectID); err != nil {
		return nil, err
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
			ast, err := s.assets.FindAsset(ctx, projectID, aInput.AssetID)
			if err != nil {
				return nil, uuid.Nil, "", exceptions.NewValidationException(fmt.Sprintf("asset %s not found in project", aInput.AssetID), err)
			}
			if primaryName == "" {
				primaryName = ast.Name
			}
			allocs = append(allocs, InvestmentAssetLink{
				AssetID:         ast.ID,
				AssetName:       ast.Name,
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

	ast, err := s.assets.FindAsset(ctx, projectID, targetID)
	if err != nil {
		return nil, uuid.Nil, "", exceptions.NewValidationException("asset not found in project", err)
	}
	allocs := []InvestmentAssetLink{
		{
			AssetID:         ast.ID,
			AssetName:       ast.Name,
			SharePercentage: 100.0,
		},
	}
	return allocs, ast.ID, ast.Name, nil
}
