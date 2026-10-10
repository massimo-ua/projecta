package investment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

const (
	failedToCreateInvestment = "failed to create investment"
	failedToFindInvestment   = "failed to find investment"
	failedToUpdateInvestment = "failed to update investment"
)

type ServiceImpl struct {
	db          core.DbConnection
	investments Repository
	assets      asset.Repository
	projects    projecta.ProjectRepository
	people      projecta.PeopleService
}

func NewService(
	db core.DbConnection,
	investments Repository,
	assets asset.Repository,
	projects projecta.ProjectRepository,
	people projecta.PeopleService,
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

	if command.AssetID == uuid.Nil {
		return nil, exceptions.NewValidationException("asset id is required", nil)
	}

	if strings.TrimSpace(command.Description) == "" {
		return nil, exceptions.NewValidationException("description is required", nil)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	project, err := s.projects.FindOne(ctx, projecta.ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	anAsset, err := s.assets.FindOne(ctx, asset.Filter{ID: command.AssetID, ProjectID: command.ProjectID})
	if err != nil {
		return nil, exceptions.NewValidationException("asset not found in project", err)
	}

	amount := command.Amount
	if amount == nil || amount.Amount() == 0 {
		if command.ResourceType == ResourceTypeTime && command.TimeHours != nil && command.TimeHourlyRate != nil {
			calculated := int64(*command.TimeHours * float64(command.TimeHourlyRate.Amount()))
			amount = money.New(calculated, command.TimeHourlyRate.Currency().Code)
		}
	}

	if amount == nil || amount.Amount() <= 0 {
		return nil, exceptions.NewValidationException("investment amount must be greater than 0", nil)
	}

	resType := command.ResourceType
	if resType == "" {
		resType = ResourceTypeMoney
	}

	inv := NewInvestment(
		uuid.New(),
		project,
		anAsset,
		owner,
		resType,
		amount,
		command.TimeHours,
		command.TimeHourlyRate,
		command.GoodsQuantity,
		command.GoodsUnit,
		command.GoodsItemName,
		command.Description,
		core.DateOrNow(command.Date),
		command.Tags,
	)

	if err = s.investments.Save(ctx, inv); err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	return inv, nil
}

func (s *ServiceImpl) CreateBatch(ctx context.Context, command CreateBatchInvestmentsCommand) ([]*Investment, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToCreateInvestment, err)
	}

	if len(command.Items) == 0 {
		return nil, exceptions.NewValidationException("at least one investment item is required", nil)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	project, err := s.projects.FindOne(ctx, projecta.ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	var created []*Investment
	_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
		for i, item := range command.Items {
			targetAssetID := item.AssetID
			if targetAssetID == uuid.Nil {
				targetAssetID = command.DefaultAsset
			}
			if targetAssetID == uuid.Nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("item %d requires an asset", i), nil)
			}

			anAsset, err := s.assets.FindOne(ctx, asset.Filter{ID: targetAssetID, ProjectID: command.ProjectID})
			if err != nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("asset %s not found in project", targetAssetID), err)
			}

			amt := item.Amount
			if (amt == nil || amt.Amount() == 0) && item.ResourceType == ResourceTypeTime && item.TimeHours != nil && item.TimeHourlyRate != nil {
				calc := int64(*item.TimeHours * float64(item.TimeHourlyRate.Amount()))
				amt = money.New(calc, item.TimeHourlyRate.Currency().Code)
			}
			if amt == nil || amt.Amount() <= 0 {
				return nil, exceptions.NewValidationException(fmt.Sprintf("item %d amount must be greater than 0", i), nil)
			}

			resType := item.ResourceType
			if resType == "" {
				resType = ResourceTypeMoney
			}

			inv := NewInvestment(
				uuid.New(),
				project,
				anAsset,
				owner,
				resType,
				amt,
				item.TimeHours,
				item.TimeHourlyRate,
				item.GoodsQuantity,
				item.GoodsUnit,
				item.GoodsItemName,
				item.Description,
				core.DateOrNow(item.Date),
				item.Tags,
			)

			if err := s.investments.Save(ctx, inv); err != nil {
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

	if command.AssetID != uuid.Nil && command.AssetID != inv.Asset.ID() {
		anAsset, err := s.assets.FindOne(ctx, asset.Filter{ID: command.AssetID, ProjectID: command.ProjectID})
		if err != nil {
			return exceptions.NewValidationException("asset not found in project", err)
		}
		inv.Asset = anAsset
	}

	if command.ResourceType != "" {
		inv.ResourceType = command.ResourceType
	}
	if command.Amount != nil && command.Amount.Amount() > 0 {
		inv.Amount = command.Amount
	}
	if command.TimeHours != nil {
		inv.TimeHours = command.TimeHours
	}
	if command.TimeHourlyRate != nil {
		inv.TimeHourlyRate = command.TimeHourlyRate
	}
	if command.GoodsQuantity != nil {
		inv.GoodsQuantity = command.GoodsQuantity
	}
	if command.GoodsUnit != "" {
		inv.GoodsUnit = command.GoodsUnit
	}
	if command.GoodsItemName != "" {
		inv.GoodsItemName = command.GoodsItemName
	}
	if strings.TrimSpace(command.Description) != "" {
		inv.Description = command.Description
	}
	if !command.Date.IsZero() {
		inv.Date = command.Date
	}
	if command.Tags != nil {
		cleanTags := make([]string, 0)
		for _, t := range command.Tags {
			if t != "" {
				cleanTags = append(cleanTags, t)
			}
		}
		inv.Tags = cleanTags
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
